package cordapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
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
	if sc == http.StatusUnauthorized{

		err = manager.RefreshToken()
		if err != nil {
			return err
		}
		
		_, err = upload(manager.host, manager.authToken.Token, uploadReq)
		if err != nil {
			return err
		}

	} else if err != nil {

		return err
	}

	return nil
}

func (manager *CordAPIManager) CmpHash(cmpReq *models.CompareHashCmd) (*models.CompareHashCmdResult, error) {

	res, sc, err := cmpHash(manager.host, manager.authToken.Token, cmpReq)
	if sc == http.StatusUnauthorized{

		err = manager.RefreshToken()
		if err != nil {
			return nil, err
		}
		
		res, _, err = cmpHash(manager.host, manager.authToken.Token, cmpReq)
		if err != nil {
			return nil, err
		}

	} else if err != nil {

		return nil, err
	}

	return res, nil
}

func (manager *CordAPIManager) GetSignature(path string) (*models.SignatureCmdResult, error) {

	res, sc, err := getSignature(manager.host, manager.authToken.Token, path)
	if sc == http.StatusUnauthorized{

		err = manager.RefreshToken()
		if err != nil {
			return nil, err
		}
		
		res, _, err = getSignature(manager.host, manager.authToken.Token, path)
		if err != nil {
			return nil, err
		}

	} else if err != nil {

		return nil, err
	}

	return res, nil
}

func (manager *CordAPIManager) AddTorrent(torrentReq *models.TorrentCmd) error {

	return manager.torrent(torrentReq, true)
}

func (manager *CordAPIManager) RemoveTorrent(torrentReq *models.TorrentCmd) error {

	return manager.torrent(torrentReq, false)
}

func (manager *CordAPIManager) torrent(torrentReq *models.TorrentCmd, add bool) error {

	var sc int
	var err error

	if add {
		sc, err = addTorrent(manager.host, manager.authToken.Token, torrentReq)
	} else {
		sc, err = removeTorrent(manager.host, manager.authToken.Token, torrentReq)
	}

	if err != nil {

		if sc == http.StatusUnauthorized {

			refreshToken, err := refreshToken(manager.host, manager.authToken.RefreshToken)
			if err != nil {
				return err
			}

			manager.authToken.Token = refreshToken.Token
			manager.authToken.RefreshToken = refreshToken.RefreshToken

			if add {
				_, err = addTorrent(manager.host, manager.authToken.Token, torrentReq)
			} else {
				_, err = removeTorrent(manager.host, manager.authToken.Token, torrentReq)
			}

			if err != nil {
				return err
			}

		} else {

			return err
		}
	}
	return nil
}

func (manager *CordAPIManager) ApplyPatch(applyReq *models.ApplyPatchCmd) error {

	sc, err := applyPatch(manager.host, manager.authToken.Token, applyReq)
	if sc == http.StatusUnauthorized{

		err = manager.RefreshToken()
		if err != nil {
			return err
		}
		
		_, err = applyPatch(manager.host, manager.authToken.Token, applyReq)
		if err != nil {
			return err
		}

	} else if err != nil {

		return err
	}

	return nil
}

func (manager *CordAPIManager) RefreshToken() error {

	refreshToken, err := refreshToken(manager.host, manager.authToken.RefreshToken)
	if err != nil {
		return err
	}

	manager.authToken.Token = refreshToken.Token
	manager.authToken.RefreshToken = refreshToken.RefreshToken

	return nil
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

	res, err := get(host+"/api/v1/auth/refresh-token", token)
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

func addTorrent(host string, token string, cmdTorrent *models.TorrentCmd) (int, error) {

	res, err := post(host+"/api/v1/tracker/torrent", token, "application/json", cmdTorrent)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return res.StatusCode, buldError(res.Body)
	}

	return res.StatusCode, nil
}

func removeTorrent(host string, token string, cmdTorrent *models.TorrentCmd) (int, error) {

	res, err := delete(host+"/api/v1/tracker/torrent", token, "application/json", cmdTorrent)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return res.StatusCode, buldError(res.Body)
	}

	return res.StatusCode, nil
}

func getSignature(host string, token string, path string) (*models.SignatureCmdResult, int, error) {

	res, err := get(host+"/api/v1/file/signature?path="+path, token)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, res.StatusCode, buldError(res.Body)
	}

	cmpRes := new(models.SignatureCmdResult)
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&cmpRes)

	return cmpRes, res.StatusCode, nil
}

func applyPatch(host string, token string, applyReq *models.ApplyPatchCmd) (int, error) {

	res, err := post(host+"/api/v1/file/patch", token, "application/json", applyReq)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return res.StatusCode, buldError(res.Body)
	}

	return res.StatusCode, nil
}

func get(url string, token string, contentType string, obj interface{}) (resp *http.Response, err error) {

	return httpRequest("GET", url, token, contentType, obj)
}

func post(url string, token string, contentType string, obj interface{}) (resp *http.Response, err error) {

	return httpRequest("POST", url, token, contentType, obj)
}

func delete(url string, token string, contentType string, obj interface{}) (resp *http.Response, err error) {

	return httpRequest("DELETE", url, token, contentType, obj)
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
