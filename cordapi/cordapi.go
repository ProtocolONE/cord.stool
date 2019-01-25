package cordapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"cord.stool/service/models"
)

func Login(host string, Username string, password string) (*models.AuthToken, error) {

	authReq := &models.Authorization{Username: Username, Password: password}
	data, err := json.Marshal(authReq)
	if err != nil {
		return nil, err
	}

	res, err := http.Post(host+"/api/v1/auth/token", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {

		errorRes := new(models.Error)
		decoder := json.NewDecoder(res.Body)
		if decoder.Decode(&errorRes) == nil {
			return nil, errors.New(errorRes.Message)
		}

		message, _ := ioutil.ReadAll(res.Body)
		return nil, errors.New(string(message))
	}

	authRes := new(models.AuthToken)
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&authRes)

	return authRes, nil
}

func Upload(host string, token string, uploadReq *models.UploadCmd) error {

	res, err := post(host+"/api/v1/file/upload", token, "application/json", uploadReq)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {

		errorRes := new(models.Error)
		decoder := json.NewDecoder(res.Body)
		if decoder.Decode(&errorRes) == nil {
			return errors.New(errorRes.Message)
		}

		message, _ := ioutil.ReadAll(res.Body)
		return errors.New(string(message))
	}

	return nil
}

func CmpHash(host string, token string, cmpReq *models.CompareHashCmd) (*models.CompareHashCmdResult, error) {

	res, err := post(host+"/api/v1/file/cmp-hash", token, "application/json", cmpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {

		errorRes := new(models.Error)
		decoder := json.NewDecoder(res.Body)
		if decoder.Decode(&errorRes) == nil {
			return nil, errors.New(errorRes.Message)
		}

		message, _ := ioutil.ReadAll(res.Body)
		return nil, errors.New(string(message))
	}

	cmpRes := new(models.CompareHashCmdResult)
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&cmpRes)

	return cmpRes, nil
}

func post(url, token string, contentType string, obj interface{}) (resp *http.Response, err error) {

	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Add("Authorization", token)
	return client.Do(req)
}
