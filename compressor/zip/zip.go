package zip

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func CompressFile(filePath, archivePath string) (err error) {
	sfi, err := os.Stat(filePath)
	if err != nil {
		return
	}

	if !sfi.Mode().IsRegular() {
		return fmt.Errorf("CompressFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}

	dfi, err := os.Stat(archivePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CompressFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
	}

	sdfi, err := os.Stat(filepath.Dir(filePath))
	if err != nil {
		return
	}

	err = os.MkdirAll(filepath.Dir(archivePath), sdfi.Mode())
	if err != nil {
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		return
	}

	defer file.Close()

	r := bufio.NewReader(file)

	arc, err := os.Create(archivePath)
	if err != nil {
		return
	}

	defer arc.Close()

	w := zip.NewWriter(arc)

	defer func() {
		cerr := w.Close()
		if err == nil {
			err = cerr
		}
	}()

	_, fname := filepath.Split(filePath)
	fw, err := w.Create(fname)
	if err != nil {
		return
	}

	_, err = io.Copy(fw, r)
	if err != nil {
		return
	}

	return
}
