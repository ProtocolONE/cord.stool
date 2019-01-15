package cord

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"net/http"
	"encoding/json"
	"bytes"

    "cord.stool/service/models"
	"cord.stool/utils"

	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"
)

var _bar *uiprogress.Bar

// Upload ...
func Upload(url string, username string, password string, sourceDir string, outputDir string) error {

	fmt.Println("Uploading to cord server ...")

	fullSourceDir, err := filepath.Abs(sourceDir)
	if err != nil {
		return err
	}

	fc, err := utils.FileCount(fullSourceDir)
	if err != nil {
		return err
	}

	uiprogress.Start()
	barTotal := uiprogress.AddBar(fc + 1 ).AppendCompleted().PrependElapsed()
	barTotal.PrependFunc(func(b *uiprogress.Bar) string {
		return strutil.Resize("Total progress", 35)
	})

	auth, err := login(url, username, password)
	if err != nil {
		return err
	}

	stopCh := make(chan struct{})
	defer func() {
		select {
		case stopCh <- struct{}{}:
		default:
		}

		close(stopCh)
	}()

	f, e := utils.EnumFilesRecursive(fullSourceDir, stopCh)

	_bar = uiprogress.AddBar(3).AppendCompleted().PrependElapsed()

	var curTitle string
	var title *string
	title = &curTitle

	_bar.PrependFunc(func(b *uiprogress.Bar) string {
		return strutil.Resize(*title, 35)
	})

	barTotal.Incr();

	for path := range f {

		_, fn := filepath.Split(path)
		curTitle = fmt.Sprint("Uploading file: ", fn)

		barTotal.Incr();
		_bar.Set(0);

		err := uploadFile(url, auth.Token, outputDir, path, fullSourceDir)
		if err != nil {
			return err
		}
	}

	err = <-e
	if err != nil {
		return err
	}

	curTitle = "Finished"
	uiprogress.Stop()

	fmt.Println("Upload completed.")
	
	return nil
}

func login(url string, Username string, password string) (*models.AuthToken, error) {

	authReq := &models.Authorisation{Username: Username, Password: password}
	data, err := json.Marshal(authReq)
    if err != nil {
        return nil, err
	}	
	
	res, err := http.Post(url + "/api/v1/token-auth", "application/json", bytes.NewBuffer(data))
	if err != nil {
        return nil, err
	}

	defer res.Body.Close()

	authRes := new(models.AuthToken)
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&authRes)

	return authRes, nil
}

func uploadFile(url string, token string, root string, path string, source string) error {

	_bar.Incr();

	_, fname := filepath.Split(path)
	relativePath, err := filepath.Rel(source, path)
	if err != nil {
		return err
	}

	fpath := filepath.Join(root, relativePath)
	fpath, _ = filepath.Split(fpath)

	file, err := os.Open(path)
	if err != nil {
		return errors.New("Cannot open file: " + err.Error())
	}
	defer file.Close()

	_bar.Incr();

	client := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf(url + "/api/v1/cmd/upload?storage=%s&name=%s", fpath, fname), file)
	if err != nil {
        return err
	}

 	req.Header.Set("Content-Type", "binary/octet-stream")
 	req.Header.Add("Authorization", token)
 	res, err := client.Do(req)	
	if err != nil {
		return errors.New("Cannot upload file: " + err.Error())
	}
	defer res.Body.Close()

	_bar.Incr();

	return nil
}
