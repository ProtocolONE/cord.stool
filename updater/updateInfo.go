package updater

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"cord.stool/updateapp"

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
	fullPath      string   `xml:"-`
	Hash          string   `xml:"crc,string,attr"`
	RawLength     int64    `xml:"rawLength,string,attr"`
	ArchiveLength int64    `xml:"archiveLength,string,attr"`
	Check         bool     `xml:"check,string,attr"`
}

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

func getAllFilesJob(rootDir string, resultCh chan<- FileInfo, stopCh <-chan struct{}) error {
	var errorCancelad = errors.New("canceled")

	defer close(resultCh)

	r := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relative, rerr := filepath.Rel(rootDir, path)

		if rerr != nil {
			return rerr
		}

		result := FileInfo{
			Path:      relative,
			fullPath:  filepath.Join(rootDir, relative),
			RawLength: info.Size(),
		}

		select {
		case resultCh <- result:
		case <-stopCh:
			return errorCancelad
		}

		return nil
	})

	if r == errorCancelad {
		return nil
	}

	return r
}

func calcMd5(fullPath string) (string, error) {
	f, err := os.Open(fullPath)

	if err != nil {
		return "", err
	}

	defer f.Close()
	h := md5.New()

	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), err
}

func calcHashJob(inputCh <-chan FileInfo, resultCh chan<- FileInfo, stopCh <-chan struct{}, wg *sync.WaitGroup) error {
	var (
		fi   FileInfo
		more bool
	)

	defer wg.Done()

LOOP:
	for {
		select {
		case fi, more = <-inputCh:
			if !more {
				break LOOP
			}

			hash, hashErr := Md5(fi.fullPath)
			if hashErr != nil {
				return hashErr
			}

			fi.Hash = hash

		case <-stopCh:
			break LOOP
		}

		select {
		case resultCh <- fi:
		case <-stopCh:
			break LOOP
		}
	}

	return nil

}

func archiveJob(
	sourceDir, targetDir string,
	useArchive bool,
	inputCh <-chan FileInfo,
	resultCh chan<- FileInfo,
	stopCh <-chan struct{},
	wg *sync.WaitGroup) (err error) {

	var (
		fi   FileInfo
		dfi  os.FileInfo
		more bool
	)

	defer wg.Done()

LOOP:
	for {
		select {
		case fi, more = <-inputCh:
			if !more {
				break LOOP
			}

			fullSrc := filepath.Join(sourceDir, fi.Path)
			fullDst := filepath.Join(targetDir, fi.Path)

			if useArchive {
				fullDst += ".zip"
				err = updateapp.CompressFile(fullSrc, fullDst)
				if err != nil {
					return
				}

			} else {
				err = CopyFile(fullSrc, fullDst)
				if err != nil {
					return
				}
			}

			dfi, err = os.Stat(fullDst)
			if err != nil {
				return
			}

			fi.ArchiveLength = dfi.Size()

		case <-stopCh:
			break LOOP
		}

		select {
		case resultCh <- fi:
		case <-stopCh:
			break LOOP
		}
	}

	return nil
}

func PrepairDistr(inputDir string, outputDir string, useArchive bool) (result UpdateInfo, err error) {
	var t tomb.Tomb

	rootInputDir, err := filepath.Abs(inputDir)

	if err != nil {
		return
	}

	scanPathCh := make(chan FileInfo)

	t.Go(func() error {
		return getAllFilesJob(rootInputDir, scanPathCh, t.Dying())
	})

	hashWg := &sync.WaitGroup{}
	hashInfoCh := make(chan FileInfo)

	hashThreadCount := 10
	hashWg.Add(hashThreadCount)

	for q := 0; q < hashThreadCount; q++ {
		t.Go(func() error {
			return calcHashJob(scanPathCh, hashInfoCh, t.Dying(), hashWg)
		})
	}

	go func() {
		hashWg.Wait()
		close(hashInfoCh)
	}()

	// ---- archive files
	archiveWg := &sync.WaitGroup{}
	archiveCh := make(chan FileInfo)

	archiveThreadCount := 10
	archiveWg.Add(archiveThreadCount)

	for q := 0; q < archiveThreadCount; q++ {
		t.Go(func() error {
			return archiveJob(
				inputDir, outputDir,
				useArchive,
				hashInfoCh, archiveCh, t.Dying(), archiveWg)
		})
	}

	go func() {
		archiveWg.Wait()
		close(archiveCh)
	}()

	//----

	t.Go(func() error {
		res := make([]FileInfo, 0, 100)

	LOOPCONSUME:
		for {
			select {
			case fi, more := <-archiveCh:
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

	crcPath := filepath.Join(outputDir, "update.crc")
	crcPathArc := filepath.Join(outputDir, "update.crc.zip")
	err = result.Save(crcPath)
	if err != nil {
		return
	}

	updateapp.CompressFile(crcPath, crcPathArc)
	err = os.Remove(crcPath)

	return result, err
}
