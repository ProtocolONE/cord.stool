package ftp

import (
	"fmt"
	"github.com/pkg/errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"net/textproto"

	"cord.stool/utils"
	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"
	"github.com/jlaffaye/ftp"
)

// UploadToFTP upload all files from sourceDir recursive
// User, password and releative ftp path should be set in ftpUrl.
// ftp://ftpuser:ftppass@ftp.protocol.local:21/cordtest/
func Upload(ftpUrl, sourceDir string) (cerr error) {

	fmt.Println("Uploading to FTP Server ...")

	fullSourceDir, cerr := filepath.Abs(sourceDir)
	if cerr != nil {
		return
	}

	fc, err := utils.FileCount(fullSourceDir)
	if err != nil {
		return err
	}

	uiprogress.Start()
	barTotal := uiprogress.AddBar(fc + 1).AppendCompleted().PrependElapsed()
	barTotal.PrependFunc(func(b *uiprogress.Bar) string {
		return strutil.Resize("Total progress", 35)
	})

	u, err := url.Parse(ftpUrl)

	if err != nil {
		return errors.Wrap(err, "UploadToFTP: Invalid ftp url")
	}

	if !strings.EqualFold(u.Scheme, "ftp") {
		return errors.New("UploadToFTP: Invalid url scheme ")
	}

	conn, err := ftp.DialTimeout(u.Host, time.Second*10) // TODO add timeout to config
	if err != nil {
		return errors.Wrap(err, "UploadToFTP: Failed to coonect ftp")
	}

	if u.User != nil {
		pass, _ := u.User.Password()
		err = conn.Login(u.User.Username(), pass)
		if err != nil {
			return errors.Wrap(err, "UploadToFTP: Cannnot login to ftp server")
		}
	}

	defer conn.Quit()
	defer conn.Logout()

	ftpRoot := u.Path
	conn.RemoveDirRecur(ftpRoot)

	stopCh := make(chan struct{})
	defer func() {
		select {
		case stopCh <- struct{}{}:
		default:
		}

		close(stopCh)
	}()

	f, e := utils.EnumFilesRecursive(fullSourceDir, stopCh)

	bar := uiprogress.AddBar(4).AppendCompleted().PrependElapsed()

	var curTitle string
	var title *string
	title = &curTitle

	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return strutil.Resize(*title, 35)
	})

	barTotal.Incr()

	var relativePath string
	var file *os.File

	for path := range f {

		_, fn := filepath.Split(path)
		curTitle = fmt.Sprint("Uploading file: ", fn)

		barTotal.Incr()
		bar.Set(0)
		bar.Incr()

		relativePath, cerr = filepath.Rel(fullSourceDir, path)
		if cerr != nil {
			return
		}

		ftpPath := filepath.Join(ftpRoot, relativePath)

		ftpDir, _ := filepath.Split(ftpPath)
		cerr = mkdirRecursive(conn, ftpDir)

		bar.Incr()

		if cerr != nil {
			return
		}

		file, cerr = os.Open(path)
		if cerr != nil {
			return
		}
		defer file.Close()

		bar.Incr()

		cerr = conn.Stor(filepath.ToSlash(ftpPath), file)

		if cerr != nil {
			return
		}

		bar.Incr()
	}

	cerr = <-e
	if cerr != nil {
		return
	}

	curTitle = "Finished"

	uiprogress.Stop()
	fmt.Println("Upload completed.")

	return nil
}

func mkdirRecursive(conn *ftp.ServerConn, dir string) (err error) {
	dir = filepath.ToSlash(dir)
	dirs := strings.Split(dir, "/")

	tmpDir := ""

	for _, d := range dirs {
		if d == "" {
			continue
		}

		tmpDir += d + "/"

		terr := conn.MakeDir(tmpDir)

		if terr != nil {
			code := terr.(*textproto.Error).Code
			if code == 550 {
				continue
			}

			return terr
		}
	}

	return nil
}
