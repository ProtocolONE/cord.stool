package updateapp

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/jlaffaye/ftp"

	"../appargs"
)

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

	err = uploadDir(conn, appargs.OutputDir, "")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	conn.Logout()
	conn.Quit()
}

func uploadDir(conn *ftp.ServerConn, path string, ftppath string) error {

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return errors.New("Cannot read directory: " + err.Error())
	}

	for _, f := range files {

		if f.IsDir() {

			var destfile string
			if len(ftppath) > 0 {
				destfile = ftppath + "\\" + f.Name()
			} else {
				destfile = f.Name()
			}

			err = uploadDir(conn, path+"\\"+f.Name(), destfile)
			if err != nil {
				fmt.Println(err.Error())
				//return err
			}

		} else {

			err = uploadFile(conn, appargs.OutputDir, f.Name(), ftppath)
			if err != nil {
				fmt.Println(err.Error())
				//return err
			}
		}
	}

	return nil
}

func uploadFile(conn *ftp.ServerConn, path string, fname string, ftppath string) error {

	var fpath, destfile string

	fpath = path + "\\" + fname

	if len(ftppath) > 0 {
		destfile = ftppath + "\\" + fname
	} else {
		destfile = fname
	}

	fmt.Println("Uploading file:", fpath)
	//fmt.Println("to:", destfile)

	file, err := os.Open(fpath)
	if err != nil {
		return errors.New("Cannot open file: " + err.Error())
	}

	r := bufio.NewReader(file)
	_, err = ioutil.ReadAll(r)
	if err != nil {
		return errors.New("Cannot read file: " + err.Error())
	}

	if len(ftppath) > 0 {
		err = conn.MakeDir(ftppath)
		if err != nil {
			// ignore error
			//return errors.New("Cannnot create ftp directory: " + err.Error())
		}
	}

	err = conn.Stor(destfile, r)
	if err != nil {
		return errors.New("Cannot upload file: " + err.Error())
	}

	return nil
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
