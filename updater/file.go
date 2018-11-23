package updater

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

// CopyFile copies a file from src to dst.
func CopyFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
	}

	sdfi, err := os.Stat(filepath.Dir(src))
	if err != nil {
		return
	}

	err = os.MkdirAll(filepath.Dir(dst), sdfi.Mode())
	if err != nil {
		return
	}

	err = copyFileContents(src, dst)
	return
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

func Md5(filepath string) (result string, err error) {
	sfi, err := os.Stat(filepath)
	if err != nil {
		return
	}

	if !sfi.Mode().IsRegular() {
		return "", fmt.Errorf("Md5: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}

	f, err := os.Open(filepath)
	if err != nil {
		return
	}

	defer f.Close()

	// TODO should we use sync.Pool here?
	h := md5.New()

	_, err = io.Copy(h, f)

	if err != nil {
		return
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func EnumFilesRecursive(rootDir string, stopCh <-chan struct{}) (result chan string, err chan error) {
	result = make(chan string)
	err = make(chan error, 1)

	fullRootDir, e := filepath.Abs(rootDir)
	if e != nil {
		err <- e
		return
	}

	go func(dir string, res chan<- string, stCh <-chan struct{}, errorCh chan<- error) {
		var errorCancelad = errors.New("canceled")

		defer close(res)
		defer close(errorCh)

		r := filepath.Walk(dir, func(path string, info os.FileInfo, err error) (werr error) {
			runtime.Gosched()

			werr = err

			if err != nil {
				return
			}

			if info.IsDir() || !info.Mode().IsRegular() {
				return
			}

			select {
			case res <- path:
			case <-stCh:
				return errorCancelad
			}

			return
		})

		if r == nil || r == errorCancelad {
			return
		}

		errorCh <- r
	}(fullRootDir, result, stopCh, err)
	return
}
