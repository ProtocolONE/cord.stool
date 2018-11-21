package updateapp

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"uplgen/appargs"
	"uplgen/utils"

	"github.com/jlaffaye/ftp"
)

type enumDirCallbackFtp struct {
	conn            *ftp.ServerConn
	continueOnError bool
}

// UploadToFTP ...
func UploadToFTP() {

	if len(appargs.FTPUrl) == 0 || len(appargs.OutputDir) == 0 {
		return
	}

	fmt.Println("Upload to ftp server ...")

	conn, err := initConnect()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer conn.Logout()
	defer conn.Quit()

	var callback utils.IEnumDirCallback = enumDirCallbackFtp{conn: conn, continueOnError: true}
	_, err = utils.EnumDir(appargs.OutputDir, callback)
	if err != nil {
		fmt.Println(err.Error())
	}

}

func (callback enumDirCallbackFtp) Process(fi os.FileInfo, path string) (bool, error) {

	if fi.IsDir() {
		return utils.EnumDir(path, callback)
	}
	return uploadFileToFtp(callback.conn, fi, path)
}

func (callback enumDirCallbackFtp) Error(err error, fi os.FileInfo, path string) bool {
	fmt.Println(err.Error())
	return callback.continueOnError
}

func uploadFileToFtp(conn *ftp.ServerConn, fi os.FileInfo, path string) (bool, error) {

	destname := path[len(appargs.OutputDir)+1 : len(path)]
	destpath, _ := filepath.Split(destname)

	fmt.Println("Uploading file:", path)

	file, err := os.Open(path)
	if err != nil {
		return true, errors.New("Cannot open file: " + err.Error())
	}

	defer file.Close()

	r := bufio.NewReader(file)
	_, err = ioutil.ReadAll(r)
	if err != nil {
		return true, errors.New("Cannot read file: " + err.Error())
	}

	if len(destpath) > 0 {
		err = conn.MakeDir(destpath)
		if err != nil {
			// ignore error
			//return errors.New("Cannnot create ftp directory: " + err.Error())
		}
	}

	err = conn.Stor(destname, r)
	if err != nil {
		return true, errors.New("Cannot upload file: " + err.Error())
	}

	return true, nil
}

func initConnect() (*ftp.ServerConn, error) {

	u, err := url.Parse(appargs.FTPUrl)
	if err != nil {
		return nil, errors.New("Invalid ftp url: " + err.Error())
	}

	if !strings.EqualFold(u.Scheme, "ftp") {
		return nil, errors.New("Invalid url scheme: " + err.Error())
	}

	conn, err := ftp.Connect(u.Host)
	if err != nil {
		return nil, err
	}

	if u.User != nil {
		pass, _ := u.User.Password()
		err = conn.Login(u.User.Username(), pass)
		if err != nil {
			return nil, errors.New("Cannnot login to ftp server: " + err.Error())
		}
	}

	err = conn.ChangeDir(u.Path)
	if err != nil {
		return nil, errors.New("Cannnot change ftp directory: " + err.Error())
	}

	return conn, nil
}
