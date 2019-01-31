package cordapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"io"
	"net/http"

	"cord.stool/service/models"
)

type CordAPIManager struct {
	host      string
	authToken *models.AuthToken
}

func NewCordAPI(host string) *CordAPIManager {
	return &CordAPIManager{host: host, authToken: nil}
}

func (manager *CordAPIManager) Login(username string, password string) error {

	var err error

	manager.authToken, err = login(manager.host, username, password)
	if err != nil {
		return err
	}
	return nil
}

func (manager *CordAPIManager) Upload(uploadReq *models.UploadCmd) error {

	sc, err := upload(manager.host, manager.authToken.Token, uploadReq)
	if err != nil {

		if sc == http.StatusUnauthorized {

			refreshToken, err := refreshToken(manager.host, manager.authToken.RefreshToken)
			if err != nil {
				return err
			}

			manager.authToken.Token = refreshToken.Token
			manager.authToken.RefreshToken = refreshToken.RefreshToken

			_, err = upload(manager.host, manager.authToken.Token, uploadReq)
			if err != nil {
				return err
			}

		} else {

			return err
		}
	}
	return nil
}

func (manager *CordAPIManager) CmpHash(cmpReq *models.CompareHashCmd) (*models.CompareHashCmdResult, error) {

	res, sc, err := cmpHash(manager.host, manager.authToken.Token, cmpReq)
	if err != nil {

		if sc == http.StatusUnauthorized {

			refreshToken, err := refreshToken(manager.host, manager.authToken.RefreshToken)
			if err != nil {
				return nil, err
			}

			manager.authToken.Token = refreshToken.Token
			manager.authToken.RefreshToken = refreshToken.RefreshToken

			res, _, err = cmpHash(manager.host, manager.authToken.Token, cmpReq)
			if err != nil {
				return nil, err
			}

		} else {

			return nil, err
		}
	}
	return res, nil
}

func login(host string, username string, password string) (*models.AuthToken, error) {

	authReq := &models.Authorization{Username: username, Password: password}
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
		return nil, buldError(res.Body)
	}

	authRes := new(models.AuthToken)
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&authRes)

	return authRes, nil
}

func refreshToken(host string, token string) (*models.AuthRefresh, error) {

	res, err := get(host+"/api/v1/auth/refresh-token", token, "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, buldError(res.Body)
	}

	refreshRes := new(models.AuthRefresh)
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&refreshRes)

	return refreshRes, nil
}

func upload(host string, token string, uploadReq *models.UploadCmd) (int, error) {

	res, err := post(host+"/api/v1/file/upload", token, "application/json", uploadReq)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return res.StatusCode, buldError(res.Body)
	}

	return res.StatusCode, nil
}

func cmpHash(host string, token string, cmpReq *models.CompareHashCmd) (*models.CompareHashCmdResult, int, error) {

	res, err := post(host+"/api/v1/file/cmp-hash", token, "application/json", cmpReq)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, res.StatusCode, buldError(res.Body)
	}

	cmpRes := new(models.CompareHashCmdResult)
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&cmpRes)

	return cmpRes, res.StatusCode, nil
}

func get(url string, token string, contentType string, obj interface{}) (resp *http.Response, err error) {

	return httpRequest("GET", url, token, contentType, obj)
}

func post(url string, token string, contentType string, obj interface{}) (resp *http.Response, err error) {

	return httpRequest("POST", url, token, contentType, obj)
}

func httpRequest(method string, url string, token string, contentType string, obj interface{}) (resp *http.Response, err error) {

	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Add("Authorization", token)
	return client.Do(req)
}

func buldError(r io.Reader) error {

	errorRes := new(models.Error)
	decoder := json.NewDecoder(r)
	if decoder.Decode(&errorRes) == nil {
		return errors.New(errorRes.Message)
	}

	message, _ := ioutil.ReadAll(r)
	return errors.New(string(message))
}
