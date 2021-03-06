package cordapi

import (
	"bytes"
	"encoding/json"
	"net/http"

	"cord.stool/service/models"
	"cord.stool/utils"
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
	if sc == http.StatusUnauthorized {

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
	if sc == http.StatusUnauthorized {

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

func (manager *CordAPIManager) GetSignature(buildid string, platform string) (*models.SignatureCmdResult, error) {

	res, sc, err := getSignature(manager.host, manager.authToken.Token, buildid, platform)
	if sc == http.StatusUnauthorized {

		err = manager.RefreshToken()
		if err != nil {
			return nil, err
		}

		res, _, err = getSignature(manager.host, manager.authToken.Token, buildid, platform)
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
	if sc == http.StatusUnauthorized {

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
		return nil, utils.BuldError(res.Body)
	}

	authRes := new(models.AuthToken)
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&authRes)

	return authRes, nil
}

func refreshToken(host string, token string) (*models.AuthRefresh, error) {

	res, err := utils.Get(host+"/api/v1/auth/refresh-token", token, "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, utils.BuldError(res.Body)
	}

	refreshRes := new(models.AuthRefresh)
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&refreshRes)

	return refreshRes, nil
}

func upload(host string, token string, uploadReq *models.UploadCmd) (int, error) {

	res, err := utils.Post(host+"/api/v1/file/upload", token, "application/json", uploadReq)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return res.StatusCode, utils.BuldError(res.Body)
	}

	return res.StatusCode, nil
}

func cmpHash(host string, token string, cmpReq *models.CompareHashCmd) (*models.CompareHashCmdResult, int, error) {

	res, err := utils.Post(host+"/api/v1/file/cmp-hash", token, "application/json", cmpReq)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, res.StatusCode, utils.BuldError(res.Body)
	}

	cmpRes := new(models.CompareHashCmdResult)
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&cmpRes)

	return cmpRes, res.StatusCode, nil
}

func addTorrent(host string, token string, cmdTorrent *models.TorrentCmd) (int, error) {

	res, err := utils.Post(host+"/api/v1/tracker/torrent", token, "application/json", cmdTorrent)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {

		merr := utils.GetModelError(res.Body)
		if merr == nil || merr.Code != models.ErrorTorrentAlreadyExists {
			return res.StatusCode, utils.BuldError(res.Body)
		}
	}

	return res.StatusCode, nil
}

func removeTorrent(host string, token string, cmdTorrent *models.TorrentCmd) (int, error) {

	res, err := utils.Delete(host+"/api/v1/tracker/torrent", token, "application/json", cmdTorrent)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return res.StatusCode, utils.BuldError(res.Body)
	}

	return res.StatusCode, nil
}

func getSignature(host string, token string, buildid string, platform string) (*models.SignatureCmdResult, int, error) {

	res, err := utils.Get(host+"/api/v1/file/signature?buildId="+buildid+"&platform="+platform, token, "application/json", nil)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, res.StatusCode, utils.BuldError(res.Body)
	}

	cmpRes := new(models.SignatureCmdResult)
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&cmpRes)

	return cmpRes, res.StatusCode, nil
}

func applyPatch(host string, token string, applyReq *models.ApplyPatchCmd) (int, error) {

	res, err := utils.Post(host+"/api/v1/file/patch", token, "application/json", applyReq)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return res.StatusCode, utils.BuldError(res.Body)
	}

	return res.StatusCode, nil
}

func (manager *CordAPIManager) CreateBranch(branchReq *models.Branch) (*models.Branch, error) {

	res, sc, err := createBranch(manager.host, manager.authToken.Token, branchReq)
	if sc == http.StatusUnauthorized {

		err = manager.RefreshToken()
		if err != nil {
			return nil, err
		}

		res, _, err = createBranch(manager.host, manager.authToken.Token, branchReq)
		if err != nil {
			return nil, err
		}

	} else if err != nil {

		return nil, err
	}

	return res, nil
}

func createBranch(host string, token string, branchReq *models.Branch) (*models.Branch, int, error) {

	res, err := utils.Post(host+"/api/v1/branch", token, "application/json", branchReq)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, res.StatusCode, utils.BuldError(res.Body)
	}

	branchRes := new(models.Branch)
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&branchRes)

	return branchRes, res.StatusCode, nil
}

func (manager *CordAPIManager) DeleteBranch(id string, name string, gameID string) (*models.Branch, error) {

	res, sc, err := deleteBranch(manager.host, manager.authToken.Token, id, name, gameID)
	if sc == http.StatusUnauthorized {

		err = manager.RefreshToken()
		if err != nil {
			return nil, err
		}

		res, _, err = deleteBranch(manager.host, manager.authToken.Token, id, name, gameID)
		if err != nil {
			return nil, err
		}

	} else if err != nil {

		return nil, err
	}

	return res, nil
}

func deleteBranch(host string, token string, id string, name string, gameID string) (*models.Branch, int, error) {

	res, err := utils.Delete(host+"/api/v1/branch?id="+id+"&name="+name+"&gid="+gameID, token, "application/json", nil)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, res.StatusCode, utils.BuldError(res.Body)
	}

	branchRes := new(models.Branch)
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&branchRes)

	return branchRes, res.StatusCode, nil
}

func (manager *CordAPIManager) SetLiveBranch(id string, name string, gameID string) (*models.Branch, error) {

	res, sc, err := setLiveBranch(manager.host, manager.authToken.Token, id, name, gameID)
	if sc == http.StatusUnauthorized {

		err = manager.RefreshToken()
		if err != nil {
			return nil, err
		}

		res, _, err = setLiveBranch(manager.host, manager.authToken.Token, id, name, gameID)
		if err != nil {
			return nil, err
		}

	} else if err != nil {

		return nil, err
	}

	return res, nil
}

func setLiveBranch(host string, token string, id string, name string, gameID string) (*models.Branch, int, error) {

	res, err := utils.Put(host+"/api/v1/branch/live?id="+id+"&name="+name+"&gid="+gameID, token, "application/json", nil)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, res.StatusCode, utils.BuldError(res.Body)
	}

	branchRes := new(models.Branch)
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&branchRes)

	return branchRes, res.StatusCode, nil
}

func (manager *CordAPIManager) GetLiveBranch(gameID string) (*models.Branch, error) {

	res, sc, err := getLiveBranch(manager.host, manager.authToken.Token, gameID)
	if sc == http.StatusUnauthorized {

		err = manager.RefreshToken()
		if err != nil {
			return nil, err
		}

		res, _, err = getLiveBranch(manager.host, manager.authToken.Token, gameID)
		if err != nil {
			return nil, err
		}

	} else if err != nil {

		return nil, err
	}

	return res, nil
}

func getLiveBranch(host string, token string, gameID string) (*models.Branch, int, error) {

	res, err := utils.Get(host+"/api/v1/branch/live?gid="+gameID, token, "application/json", nil)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, res.StatusCode, utils.BuldError(res.Body)
	}

	branchRes := new(models.Branch)
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&branchRes)

	return branchRes, res.StatusCode, nil
}

func (manager *CordAPIManager) GetBranch(id string, name string, gameID string) (*models.Branch, error) {

	res, sc, err := getBranch(manager.host, manager.authToken.Token, id, name, gameID)
	if sc == http.StatusUnauthorized {

		err = manager.RefreshToken()
		if err != nil {
			return nil, err
		}

		res, _, err = getBranch(manager.host, manager.authToken.Token, id, name, gameID)
		if err != nil {
			return nil, err
		}

	} else if err != nil {

		return nil, err
	}

	return res, nil
}

func getBranch(host string, token string, id string, name string, gameID string) (*models.Branch, int, error) {

	res, err := utils.Get(host+"/api/v1/branch?id="+id+"&name="+name+"&gid="+gameID, token, "application/json", nil)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, res.StatusCode, utils.BuldError(res.Body)
	}

	branchRes := new(models.Branch)
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&branchRes)

	return branchRes, res.StatusCode, nil
}

func (manager *CordAPIManager) UpdateBranch(branchReq *models.Branch) error {

	sc, err := updateBranch(manager.host, manager.authToken.Token, branchReq)
	if sc == http.StatusUnauthorized {

		err = manager.RefreshToken()
		if err != nil {
			return err
		}

		_, err = updateBranch(manager.host, manager.authToken.Token, branchReq)
		if err != nil {
			return err
		}

	} else if err != nil {

		return err
	}

	return nil
}

func updateBranch(host string, token string, branchReq *models.Branch) (int, error) {

	res, err := utils.Put(host+"/api/v1/branch?id="+branchReq.ID, token, "application/json", branchReq)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return res.StatusCode, utils.BuldError(res.Body)
	}

	return res.StatusCode, nil
}

func (manager *CordAPIManager) ListBranch(gameID string) (*[]models.Branch, error) {

	res, sc, err := listBranch(manager.host, manager.authToken.Token, gameID)
	if sc == http.StatusUnauthorized {

		err = manager.RefreshToken()
		if err != nil {
			return nil, err
		}

		res, _, err = listBranch(manager.host, manager.authToken.Token, gameID)
		if err != nil {
			return nil, err
		}

	} else if err != nil {

		return nil, err
	}

	return res, nil
}

func listBranch(host string, token string, gameID string) (*[]models.Branch, int, error) {

	res, err := utils.Get(host+"/api/v1/branch/list?gid="+gameID, token, "application/json", nil)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, res.StatusCode, utils.BuldError(res.Body)
	}

	branchRes := new([]models.Branch)
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&branchRes)

	return branchRes, res.StatusCode, nil
}

func (manager *CordAPIManager) ShallowBranch(sid string, sname string, tid string, tname string, gameID string) (*models.ShallowBranchCmdResult, error) {

	res, sc, err := shallowBranch(manager.host, manager.authToken.Token, sid, sname, tid, tname, gameID)
	if sc == http.StatusUnauthorized {

		err = manager.RefreshToken()
		if err != nil {
			return nil, err
		}

		res, _, err = shallowBranch(manager.host, manager.authToken.Token, sid, sname, tid, tname, gameID)
		if err != nil {
			return nil, err
		}

	} else if err != nil {

		return nil, err
	}

	return res, nil
}

func shallowBranch(host string, token string, sid string, sname string, tid string, tname string, gameID string) (*models.ShallowBranchCmdResult, int, error) {

	res, err := utils.Post(host+"/api/v1/branch/shallow?sid="+sid+"&sname="+sname+"&tid="+tid+"&tname="+tname+"&gid="+gameID, token, "application/json", nil)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, res.StatusCode, utils.BuldError(res.Body)
	}

	branchRes := new(models.ShallowBranchCmdResult)
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&branchRes)

	return branchRes, res.StatusCode, nil
}

func (manager *CordAPIManager) CreateBuild(buildReq *models.Build) (*models.Build, error) {

	res, sc, err := createBuild(manager.host, manager.authToken.Token, buildReq)
	if sc == http.StatusUnauthorized {

		err = manager.RefreshToken()
		if err != nil {
			return nil, err
		}

		res, _, err = createBuild(manager.host, manager.authToken.Token, buildReq)
		if err != nil {
			return nil, err
		}

	} else if err != nil {

		return nil, err
	}

	return res, nil
}

func createBuild(host string, token string, buildReq *models.Build) (*models.Build, int, error) {

	res, err := utils.Post(host+"/api/v1/branch/build", token, "application/json", buildReq)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, res.StatusCode, utils.BuldError(res.Body)
	}

	buildRes := new(models.Build)
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&buildRes)

	return buildRes, res.StatusCode, nil
}

func (manager *CordAPIManager) DeleteBuild(id string) error {

	sc, err := deleteBuild(manager.host, manager.authToken.Token, id)
	if sc == http.StatusUnauthorized {

		err = manager.RefreshToken()
		if err != nil {
			return err
		}

		_, err = deleteBuild(manager.host, manager.authToken.Token, id)
		if err != nil {
			return err
		}

	} else if err != nil {

		return err
	}

	return nil
}

func deleteBuild(host string, token string, id string) (int, error) {

	res, err := utils.Delete(host+"/api/v1/branch/build?id="+id, token, "application/json", nil)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return res.StatusCode, utils.BuldError(res.Body)
	}

	return res.StatusCode, nil
}

func (manager *CordAPIManager) GetBuild(id string) (*models.Build, error) {

	res, sc, err := getBuild(manager.host, manager.authToken.Token, id)
	if sc == http.StatusUnauthorized {

		err = manager.RefreshToken()
		if err != nil {
			return nil, err
		}

		res, _, err = getBuild(manager.host, manager.authToken.Token, id)
		if err != nil {
			return nil, err
		}

	} else if err != nil {

		return nil, err
	}

	return res, nil
}

func getBuild(host string, token string, id string) (*models.Build, int, error) {

	res, err := utils.Get(host+"/api/v1/branch/build?id="+id, token, "application/json", nil)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, res.StatusCode, utils.BuldError(res.Body)
	}

	buildRes := new(models.Build)
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&buildRes)

	return buildRes, res.StatusCode, nil
}

func (manager *CordAPIManager) ListBuild(gameID string, branch string) (*[]models.Build, error) {

	res, sc, err := listBuild(manager.host, manager.authToken.Token, gameID, branch)
	if sc == http.StatusUnauthorized {

		err = manager.RefreshToken()
		if err != nil {
			return nil, err
		}

		res, _, err = listBuild(manager.host, manager.authToken.Token, gameID, branch)
		if err != nil {
			return nil, err
		}

	} else if err != nil {

		return nil, err
	}

	return res, nil
}

func listBuild(host string, token string, gameID string, branch string) (*[]models.Build, int, error) {

	res, err := utils.Get(host+"/api/v1/branch/build/list?gid="+gameID+"&name="+branch, token, "application/json", nil)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, res.StatusCode, utils.BuldError(res.Body)
	}

	listRes := new([]models.Build)
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&listRes)

	return listRes, res.StatusCode, nil
}

func (manager *CordAPIManager) GetLiveBuild(gameID string, branch string) (*models.Build, error) {

	res, sc, err := getLiveBuild(manager.host, manager.authToken.Token, gameID, branch)
	if sc == http.StatusUnauthorized {

		err = manager.RefreshToken()
		if err != nil {
			return nil, err
		}

		res, _, err = getLiveBuild(manager.host, manager.authToken.Token, gameID, branch)
		if err != nil {
			return nil, err
		}

	} else if err != nil {

		return nil, err
	}

	return res, nil
}

func getLiveBuild(host string, token string, gameID string, branch string) (*models.Build, int, error) {

	res, err := utils.Get(host+"/api/v1/branch/build/live?gid="+gameID+"&name="+branch, token, "application/json", nil)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, res.StatusCode, utils.BuldError(res.Body)
	}

	buildRes := new(models.Build)
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&buildRes)

	return buildRes, res.StatusCode, nil
}

func (manager *CordAPIManager) PublishBuild(gameID string, branch string, build string, platform string) (*models.Branch, error) {

	res, sc, err := publishBuild(manager.host, manager.authToken.Token, gameID, branch, build, platform)
	if sc == http.StatusUnauthorized {

		err = manager.RefreshToken()
		if err != nil {
			return nil, err
		}

		res, _, err = publishBuild(manager.host, manager.authToken.Token, gameID, branch, build, platform)
		if err != nil {
			return nil, err
		}

	} else if err != nil {

		return nil, err
	}

	return res, nil
}

func publishBuild(host string, token string, gameID string, branch string, build string, platform string) (*models.Branch, int, error) {

	res, err := utils.Put(host+"/api/v1/branch/build/publish?gid="+gameID+"&name="+branch+"&build="+build+"&platform="+platform, token, "application/json", nil)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, res.StatusCode, utils.BuldError(res.Body)
	}

	branchdRes := new(models.Branch)
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&branchdRes)

	return branchdRes, res.StatusCode, nil
}

func (manager *CordAPIManager) GetUpdateInfo(gameID string, branch string, locale string, platform string, ver string) (*models.UpdateInfo, error) {

	res, sc, err := getUpdateInfo(manager.host, manager.authToken.Token, gameID, branch, locale, platform, ver)
	if sc == http.StatusUnauthorized {

		err = manager.RefreshToken()
		if err != nil {
			return nil, err
		}

		res, _, err = getUpdateInfo(manager.host, manager.authToken.Token, gameID, branch, locale, platform, ver)
		if err != nil {
			return nil, err
		}

	} else if err != nil {

		return nil, err
	}

	return res, nil
}

func getUpdateInfo(host string, token string, gameID string, branch string, locale string, platform string, ver string) (*models.UpdateInfo, int, error) {

	res, err := utils.Get(host+"/api/v1/file/update-info?gid="+gameID+"&name="+branch+"&locale="+locale+"&platform="+platform+"&ver="+ver, token, "application/json", nil)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, res.StatusCode, utils.BuldError(res.Body)
	}

	info := new(models.UpdateInfo)
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&info)

	return info, res.StatusCode, nil
}

func (manager *CordAPIManager) GetUpdatePatch(gameID string, branch string, locale string, platform string, ver string) (*models.UpdateInfo, error) {

	res, sc, err := getUpdatePatch(manager.host, manager.authToken.Token, gameID, branch, locale, platform, ver)
	if sc == http.StatusUnauthorized {

		err = manager.RefreshToken()
		if err != nil {
			return nil, err
		}

		res, _, err = getUpdatePatch(manager.host, manager.authToken.Token, gameID, branch, locale, platform, ver)
		if err != nil {
			return nil, err
		}

	} else if err != nil {

		return nil, err
	}

	return res, nil
}

func getUpdatePatch(host string, token string, gameID string, branch string, locale string, platform string, ver string) (*models.UpdateInfo, int, error) {

	res, err := utils.Get(host+"/api/v1/file/update-patch?gid="+gameID+"&name="+branch+"&locale="+locale+"&platform="+platform+"&ver="+ver, token, "application/json", nil)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, res.StatusCode, utils.BuldError(res.Body)
	}

	info := new(models.UpdateInfo)
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&info)

	return info, res.StatusCode, nil
}
