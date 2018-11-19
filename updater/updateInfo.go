package updater

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
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

func calcHash(root string, fi *FileInfo, wg *sync.WaitGroup, quotaCh chan struct{}) {
	quotaCh <- struct{}{}
	defer wg.Done()
	defer func() {
		<-quotaCh
	}()

	fullPath := filepath.Join(root, fi.Path)

	// UNDONE return error to somewhere
	f, err := os.Open(fullPath)
	if err != nil {
		return
	}

	defer f.Close()
	h := md5.New()

	if _, err := io.Copy(h, f); err != nil {
		// UNDONE return error to somewhere
		return
	}

	fi.Hash = hex.EncodeToString(h.Sum(nil))

	fmt.Println(fi.Hash, fi.Path)
}

func Calculate(dir string) (UpdateInfo, error) {
	var updateInfo UpdateInfo

	rootDir, _ := filepath.Abs(dir)

	filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
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

		fi := FileInfo{
			Path:      relative,
			RawLength: info.Size(),
		}

		updateInfo.FilesInfo.Files = append(updateInfo.FilesInfo.Files, fi)
		return nil
	})

	wg := &sync.WaitGroup{}
	quotaLimit := 10
	quotaCh := make(chan struct{}, quotaLimit) // ratelim.go

	for fi, _ := range updateInfo.FilesInfo.Files {
		wg.Add(1)
		go calcHash(rootDir, &updateInfo.FilesInfo.Files[fi], wg, quotaCh)
	}

	wg.Wait()

	return updateInfo, nil
}
