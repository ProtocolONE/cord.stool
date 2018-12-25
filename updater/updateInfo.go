package updater

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"cord.stool/compressor/zip"
	"cord.stool/utils"

	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"

	tomb "gopkg.in/tomb.v2"
)

type UpdateInfo struct {
	XMLName   xml.Name     `xml:"UpdateFileList"`
	Version1  string       `xml:"version,attr,omitempty"`
	FilesInfo FileInfoList `xml:"files"`
}

type FileInfoList struct {
	XMLName xml.Name   `xml:"files"`
	Files   []FileInfo `xml:"file"`
}

type FileInfo struct {
	XMLName       xml.Name `xml:"file"`
	Path          string   `xml:"path,string,attr"`
	fullPath      string   `xml:"-"`
	Hash          string   `xml:"crc,string,attr"`
	RawLength     int64    `xml:"rawLength,string,attr"`
	ArchiveLength int64    `xml:"archiveLength,string,attr"`
	Check         bool     `xml:"check,string,attr"`
}

var progrssBar *uiprogress.Bar
var curProgressTitle string

func (s UpdateInfo) Pack() (string, error) {
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	writer.Write([]byte(xml.Header))
	enc := xml.NewEncoder(writer)
	enc.Indent("", "  ")

	if err := enc.Encode(s); err != nil {
		return "", err
	}

	writer.Flush()

	return b.String(), nil

}

func (s *UpdateInfo) Unpack(buffer []byte) error {
	return xml.Unmarshal(buffer, s)
}

func UnpackUpdateInfo(buffer []byte) (UpdateInfo, error) {
	u := new(UpdateInfo)
	e := xml.Unmarshal(buffer, u)
	return *u, e
}

func (s *UpdateInfo) Save(filepath string) (err error) {
	sfi, err := os.Stat(filepath)

	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(sfi.Mode().IsRegular()) {
			return fmt.Errorf("UpdateInfo.Save: non-regular destination file %s (%q)", sfi.Name(), sfi.Mode().String())
		}
	}

	f, err := os.Create(filepath)
	if err != nil {
		return
	}

	defer f.Close()

	writer := bufio.NewWriter(f)

	writer.Write([]byte(xml.Header))
	enc := xml.NewEncoder(writer)
	enc.Indent("", "  ")
	err = enc.Encode(s)

	if err != nil {
		return
	}

	writer.Flush()

	return nil
}

func processFile(sourceDir, targetDir string,
	useArchive bool,
	inputCh <-chan string,
	resultCh chan<- FileInfo,
	stopCh <-chan struct{},
	wg *sync.WaitGroup) (err error) {

	var (
		path, relativePath, hash string
		more                     bool
		fi                       os.FileInfo
		ufi                      FileInfo
	)

	defer wg.Done()
LOOP:
	for {
		select {
		case path, more = <-inputCh:
			if !more {
				break LOOP
			}

			_, fn := filepath.Split(path)
			curProgressTitle = fmt.Sprint("Calcing hash for ", fn)
			progrssBar.Incr();
		
			relativePath, err = filepath.Rel(sourceDir, path)

			if err != nil {
				return
			}

			fi, err = os.Stat(path)
			if err != nil {
				return
			}

			ufi = FileInfo{
				Path:      relativePath,
				fullPath:  path,
				RawLength: fi.Size(),
			}

			hash, err = utils.Md5(path)
			if err != nil {
				return
			}

			ufi.Hash = hash

			fullDst := filepath.Join(targetDir, relativePath)

			if useArchive {

				progrssBar.Incr();
				curProgressTitle = fmt.Sprint("Compressing file: ", fn)

				fullDst += ".zip"

				err = zip.CompressFile(path, fullDst)
				if err != nil {
					return
				}

				fi, err = os.Stat(fullDst)
				if err != nil {
					return
				}

				ufi.ArchiveLength = fi.Size()

			} else {
				err = utils.CopyFile(path, fullDst)
				if err != nil {
					return
				}

				ufi.ArchiveLength = ufi.RawLength
			}

			progrssBar.Incr();

		case <-stopCh:
			break LOOP
		}

		select {
		case resultCh <- ufi:
		case <-stopCh:
			break LOOP
		}
	}

	return nil
}

func PrepairDistr(inputDir string, outputDir string, useArchive bool) (result UpdateInfo, err error) {

	fmt.Println("Creating update ...")

	result = UpdateInfo{}
	files, err := utils.GetAllFiles(inputDir)
	if err != nil {
		return
	}

	pbIndex := 2
	if useArchive {
		pbIndex++
	}

	uiprogress.Start()
	progrssBar = uiprogress.AddBar(len(files) * pbIndex + 3).AppendCompleted().PrependElapsed()
 
	var title *string
	title = &curProgressTitle
	curProgressTitle = "Preparing ..."

	progrssBar.PrependFunc(func(b *uiprogress.Bar) string {
		return strutil.Resize(*title, 35)
	})

	var t tomb.Tomb

	pathCh := make(chan string)

	t.Go(func() error {
		defer close(pathCh)

	LOOPPRODUCE:
		for _, p := range files {
			select {
			case pathCh <- p:
				continue
			case <-t.Dying():
				break LOOPPRODUCE
			}
		}

		return nil
	})

	progrssBar.Incr();

	wg := &sync.WaitGroup{}
	fiCh := make(chan FileInfo)

	threadCount := 10
	wg.Add(threadCount)

	for q := 0; q < threadCount; q++ {
		t.Go(func() error {
			return processFile(
				inputDir, outputDir,
				useArchive,
				pathCh, fiCh, t.Dying(), wg)
		})
	}

	go func() {
		wg.Wait()
		close(fiCh)
	}()

	curProgressTitle = "Createing update.crc ..."
	title = &curProgressTitle

	t.Go(func() error {
		res := make([]FileInfo, 0, 100)

	LOOPCONSUME:
		for {
			select {
			case fi, more := <-fiCh:
				if !more {
					break LOOPCONSUME
				}

				res = append(res, fi)
			case <-t.Dying():
				break LOOPCONSUME
			}
		}

		result.FilesInfo.Files = res
		return nil
	})

	err = t.Wait()
	if err != nil {
		return
	}

	progrssBar.Incr();

	crcPath := filepath.Join(outputDir, "update.crc")
	crcPathArc := filepath.Join(outputDir, "update.crc.zip")
	err = result.Save(crcPath)
	if err != nil {
		return
	}

	progrssBar.Incr();
	curProgressTitle = "Compressing update.crc ..."
	title = &curProgressTitle

	zip.CompressFile(crcPath, crcPathArc)
	err = os.Remove(crcPath)

	progrssBar.Incr();

	curProgressTitle = "Finished"
	title = &curProgressTitle
	uiprogress.Stop()

	fmt.Println("Creating is completed.")

	return result, err
}
