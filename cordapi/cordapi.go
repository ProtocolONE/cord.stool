package cordapi

import (
	"errors"
	"net/http"
	"encoding/json"
	"bytes"
	"io/ioutil"

    "cord.stool/service/models"
	"cord.stool/utils"
)

func Login(host string, Username string, password string) (*models.AuthToken, error) {

	authReq := &models.Authorization{Username: Username, Password: password}
	data, err := json.Marshal(authReq)
    if err != nil {
        return nil, err
	}	
	
	res, err := http.Post(host + "/api/v1/token-auth", "application/json", bytes.NewBuffer(data))
	if err != nil {
        return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		message, _ := ioutil.ReadAll(res.Body)
		return nil, errors.New(string(message))
	}

	authRes := new(models.AuthToken)
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&authRes)

	return authRes, nil
}

func Upload(host string, token string, uploadReq *models.UploadCmd) error {

	data, err := json.Marshal(uploadReq)
    if err != nil {
        return err
	}	

	res, err := utils.Post(host + "/api/v1/cmd/upload", token, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		message, _ := ioutil.ReadAll(res.Body)
		return errors.New(string(message))
	}

	return nil
}
func CmpHash(host string, token string, cmpReq *models.CompareHashCmd) (*models.CompareHashCmdResult, error) {

	data, err := json.Marshal(cmpReq)
    if err != nil {
        return nil, err
	}	

	res, err := utils.Post(host + "/api/v1/cmd/cmp-hash", token, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		message, _ := ioutil.ReadAll(res.Body)
		return nil, errors.New(string(message))
	}

	cmpRes := new(models.CompareHashCmdResult)
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&cmpRes)

	return cmpRes, nil
}