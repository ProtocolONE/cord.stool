package qilinapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"cord.stool/service/models"
	"cord.stool/utils"
)

type QilinAPIManager struct {
	host      string
	authToken string
}

func NewQilinAPI(host string) (*QilinAPIManager, error) {

	token, err := getAccessToken()
	if err != nil {
		return nil, err
	}

	return &QilinAPIManager{host: host, authToken: token}, nil
}

func (manager *QilinAPIManager) ListVendor() (*[]models.Vendor, error) {

	res, _, err := listVendor(manager.host, manager.authToken)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func listVendor(host string, token string) (*[]models.Vendor, int, error) {

	res, err := utils.Get(host+"/api/v1/vendors", " Bearer "+token, "application/json", nil)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, res.StatusCode, utils.BuldError(res.Body)
	}

	listRes := new([]models.Vendor)
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&listRes)

	return listRes, res.StatusCode, nil
}

func (manager *QilinAPIManager) ListGame(vendorID string) (*[]models.Game, error) {

	res, _, err := listGame(manager.host, manager.authToken, vendorID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func listGame(host string, token string, vendorID string) (*[]models.Game, int, error) {

	res, err := utils.Get(host+"/api/v1/vendors/"+vendorID+"/games", " Bearer "+token, "application/json", nil)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, res.StatusCode, utils.BuldError(res.Body)
	}

	listRes := new([]models.Game)
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&listRes)

	return listRes, res.StatusCode, nil
}

func (manager *QilinAPIManager) GetGameInfo(gameID string) (*models.GameInfo, error) {

	res, _, err := getGameInfo(manager.host, manager.authToken, gameID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func getGameInfo(host string, token string, gameID string) (*models.GameInfo, int, error) {

	res, err := utils.Get(host+"/api/v1/games/"+gameID, " Bearer "+token, "application/json", nil)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, res.StatusCode, utils.BuldError(res.Body)
	}

	info := new(models.GameInfo)
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&info)

	return info, res.StatusCode, nil
}

func defaultKeyPath() string {

	dir := ".config/ProtocolONE"
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE")
	}

	if runtime.GOOS == "darwin" {
		home = path.Join(home, "Library", "Application Support")
		dir = "ProtocolONE"
	}

	configPath := filepath.FromSlash(path.Join(home, dir, "cord.stool.creds"))
	return configPath
}

func getAccessToken() (string, error) {

	keyPath := defaultKeyPath()
	stats, err := os.Lstat(keyPath)
	exists := !os.IsNotExist(err)
	if !exists {
		return "", err
	}

	if stats.Mode()&077 > 0 && runtime.GOOS != "windows" {
		err = os.Chmod(keyPath, 0600)
	}

	buf, err := ioutil.ReadFile(keyPath)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	return strings.TrimSpace(string(buf)), nil
}
