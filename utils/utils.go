package utils

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

// IEnumDirCallback ...
type IEnumDirCallback interface {
	Process(fi os.FileInfo, path string) (bool, error)
	Error(err error, fi os.FileInfo, path string) bool
}

// EnumDir ...
func EnumDir(path string, calback IEnumDirCallback) (bool, error) {

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return true, errors.New("Cannot read directory: " + err.Error())
	}

	for _, f := range files {

		cntn, err := calback.Process(f, filepath.Join(path, f.Name()))
		if err != nil {
			if !calback.Error(err, f, filepath.Join(path, f.Name())) {
				return false, nil
			}
		}

		if !cntn {
			return false, nil
		}

	}

	return true, nil
}
