package updater

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"

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

			hash, hashErr := calcMd5(fi.fullPath)
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

func PrepairDistr(inputDir string, outputDir string) (UpdateInfo, error) {
	var t tomb.Tomb
	var result UpdateInfo

	rootInputDir, pathErr := filepath.Abs(inputDir)

	if pathErr != nil {
		return result, pathErr
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

	t.Go(func() error {
		res := make([]FileInfo, 0, 100)

	LOOPCONSUME:
		for {
			select {
			case fi, more := <-hashInfoCh:
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

	totalError := t.Wait()

	return result, totalError
}
