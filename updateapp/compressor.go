package updateapp

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// CompressFile ...
func CompressFile(path string) {

	fmt.Println("Compress file:", path)

	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Cannot open file: " + err.Error())
		return
	}
	defer file.Close()

	r := bufio.NewReader(file)
	b, err := ioutil.ReadAll(r)
	if err != nil {
		fmt.Println("Cannot read file: " + err.Error())
		return
	}

	arc, err := os.Create(path + ".zip")
	if err != nil {
		fmt.Println("Cannot create file: " + err.Error())
		return
	}
	defer arc.Close()

	w := zip.NewWriter(arc)

	_, fname := filepath.Split(path)
	fw, err := w.Create(fname)
	if err != nil {
		fmt.Println("Cannot create archive: " + err.Error())
		return
	}
	_, err = fw.Write(b)
	if err != nil {
		fmt.Println("Cannot write archive: " + err.Error())
		return
	}

	err = w.Close()
	if err != nil {
		fmt.Println("Cannot close archive: " + err.Error())
		return
	}
}