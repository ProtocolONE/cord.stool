package controllers

import (
	"cord.stool/cordapi"
	"cord.stool/service/config"
	"cord.stool/service/core/utils"
	"cord.stool/service/database"
	"cord.stool/service/models"
	"cord.stool/upload/cord"
	"cord.stool/upload/s3"
	utils2 "cord.stool/utils"

	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/labstack/echo"
	"go.uber.org/zap"
)

func getBranchIDOrName(context echo.Context) string {

	nameOrID := ""
	nameOrID = context.Param("id")
	if nameOrID == "" {
		nameOrID = context.Param("name")
	}

	return nameOrID
}

func CreateBranchCmd(context echo.Context) error {

	reqBranch := &models.Branch{}
	err := context.Bind(reqBranch)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorInvalidJSONFormat, err.Error())
	}

	manager := database.NewBranchManager()
	defer manager.Close()
	result, err := manager.FindByName(reqBranch.Name, reqBranch.GameID)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	if result != nil {
		return utils.BuildBadRequestError(context, models.ErrorAlreadyExists, reqBranch.Name)
	}

	branches, err := manager.List(reqBranch.GameID)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	reqBranch.ID = utils2.GenerateID()
	if branches == nil || len(branches) == 0 {
		reqBranch.Live = true
	} else {
		reqBranch.Live = false
	}
	reqBranch.Created = time.Now()

	err = manager.Insert(reqBranch)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	return context.JSON(http.StatusOK, reqBranch)
}

func findBranch(context echo.Context, bidParam string, nameParam string, gidParam string) (*models.Branch, bool, error) {

	var result *models.Branch
	var err error

	manager := database.NewBranchManager()
	defer manager.Close()

	bid := context.QueryParam(bidParam)
	name := context.QueryParam(nameParam)
	gid := context.QueryParam(gidParam)

	if bid == "" && (name == "" || gid == "") {
		return nil, false, utils.BuildBadRequestError(context, models.ErrorInvalidRequest, "Branch ID or Name and Game ID are required")
	}

	if bid != "" {
		result, err = manager.FindByID(bid)
		if err != nil {
			return nil, false, utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
		}
		if result == nil {
			return nil, false, utils.BuildBadRequestError(context, models.ErrorInvalidRequest, "Invalid branch id: "+bid)
		}
	} else {
		result, err = manager.FindByName(name, gid)
		if err != nil {
			return nil, false, utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
		}
		if result == nil {
			return nil, false, utils.BuildBadRequestError(context, models.ErrorInvalidRequest, "Invalid branch name \""+name+"\" or Game ID \""+gid+"\"")
		}
	}

	if bid != "" && gid != "" && result.GameID != gid {
		return nil, false, utils.BuildBadRequestError(context, models.ErrorInvalidRequest, "Specified Game ID \""+gid+"\""+" is not belonged to the Branch ID \""+bid+"\"")
	}

	return result, true, nil
}

func DeleteBranchCmd(context echo.Context) error {

	result, ok, err := findBranch(context, "id", "name", "gid")
	if !ok {
		return err
	}

	managerB := database.NewBuildManager()
	defer managerB.Close()
	builds, err := managerB.FindBuildByBranchID(result.ID)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	for _, b := range builds {

		managerBD := database.NewBuildDepotManager()
		defer managerBD.Close()
		buildDepots, err := managerBD.FindByBuildID(b.ID)
		if err != nil {
			return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
		}

		for _, bd := range buildDepots {

			managerD := database.NewDepotManager()
			defer managerD.Close()
			depot, err := managerD.FindByID(bd.DepotID)
			if err != nil {
				return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
			}

			if depot != nil {

				err = utils.RemoveDepotFiles(context.Request().Header.Get("ClientID"), bd.DepotID, context)
				if err != nil {
					return err
				}

				err = managerD.RemoveByID(bd.DepotID)
				if err != nil {
					return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
				}
			}

			err = managerBD.RemoveByID(bd.ID)
			if err != nil {
				return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
			}
		}

		err = managerB.RemoveByID(b.ID)
		if err != nil {
			return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
		}
	}

	manager := database.NewBranchManager()
	defer manager.Close()
	err = manager.RemoveByID(result.ID)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	return context.JSON(http.StatusOK, result)
}

func SetLiveBranchCmd(context echo.Context) error {

	result, ok, err := findBranch(context, "id", "name", "gid")
	if !ok {
		return err
	}
	gid := result.GameID

	if result.Live != true {

		manager := database.NewBranchManager()
		defer manager.Close()
		branches, err := manager.List(gid)
		if err != nil {
			return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
		}

		for _, b := range branches {

			if b.Live == true && b.ID != result.ID {
				b.Live = false
				b.Updated = time.Now()
				err = manager.Update(b)
				if err != nil {
					return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
				}
			}
		}

		result.Live = true
		result.Updated = time.Now()
		err = manager.Update(result)
		if err != nil {
			return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
		}
	}

	return context.JSON(http.StatusOK, result)
}

func GetLiveBranchCmd(context echo.Context) error {

	gid := context.QueryParam("gid")
	if gid == "" {
		return utils.BuildBadRequestError(context, models.ErrorInvalidRequest, "Game ID is required")
	}

	manager := database.NewBranchManager()
	defer manager.Close()
	branches, err := manager.List(gid)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	for _, b := range branches {

		if b.Live == true {
			return context.JSON(http.StatusOK, b)
		}
	}

	return context.NoContent(http.StatusNotFound)
}

func GetBranchCmd(context echo.Context) error {

	result, ok, err := findBranch(context, "id", "name", "gid")
	if !ok {
		return err
	}

	return context.JSON(http.StatusOK, result)
}

func UpdateBranchCmd(context echo.Context) error {

	bid := context.QueryParam("id")
	if bid == "" {
		return utils.BuildBadRequestError(context, models.ErrorInvalidRequest, "Branch ID is required")
	}

	manager := database.NewBranchManager()
	defer manager.Close()
	result, err := manager.FindByID(bid)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	if result == nil {
		return utils.BuildBadRequestError(context, models.ErrorNotFound, bid)
	}

	reqBranch := &models.Branch{}
	err = context.Bind(reqBranch)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorInvalidJSONFormat, err.Error())
	}

	reqBranch.ID = result.ID
	reqBranch.Updated = time.Now()
	err = manager.Update(reqBranch)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	return context.NoContent(http.StatusOK)
}

func ListBranchCmd(context echo.Context) error {

	gid := context.QueryParam("gid")
	if gid == "" {
		return utils.BuildBadRequestError(context, models.ErrorInvalidRequest, "Game ID is required")
	}

	manager := database.NewBranchManager()
	defer manager.Close()
	branches, err := manager.List(gid)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	return context.JSON(http.StatusOK, branches)
}

func ShallowBranchCmd(context echo.Context) error {

	sourceBranch, ok, err := findBranch(context, "sid", "sname", "gid")
	if !ok {
		return err
	}

	targetBranch, ok, err := findBranch(context, "tid", "tname", "gid")
	if !ok {
		return err
	}

	targetBranch.LiveBuild = sourceBranch.LiveBuild
	targetBranch.Updated = time.Now()

	manager := database.NewBranchManager()
	defer manager.Close()
	err = manager.Update(targetBranch)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	return context.JSON(http.StatusOK, models.ShallowBranchCmdResult{sourceBranch.ID, sourceBranch.Name, targetBranch.ID, targetBranch.Name})
}

func mergeBuilds(branchID string, newBuildID string, context echo.Context) error {

	manager := database.NewBuildManager()
	defer manager.Close()
	builds, err := manager.FindBuildByBranchID(branchID)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	buildID := ""
	if builds != nil && len(builds) != 0 {
		// skip new build
		if builds[0].ID == newBuildID && len(builds) > 1 {
			buildID = builds[1].ID
		} else if builds[0].ID != newBuildID && len(builds) > 0 {
			buildID = builds[0].ID
		}
	}

	if buildID == "" {
		return nil
	}

	managerBD := database.NewBuildDepotManager()
	defer managerBD.Close()
	buildDepots, err := managerBD.FindByBuildID(buildID)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	for _, bd := range buildDepots {

		if bd.ID == newBuildID {
			continue
		}

		bd.LinkID = bd.ID
		bd.ID = utils2.GenerateID()
		bd.BuildID = newBuildID
		bd.Created = time.Now()
		err = managerBD.Insert(bd)
		if err != nil {
			return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
		}
	}

	return nil
}

func CreateBuildCmd(context echo.Context) error {

	reqBuild := &models.Build{}
	err := context.Bind(reqBuild)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorInvalidJSONFormat, err.Error())
	}

	reqBuild.ID = utils2.GenerateID()
	reqBuild.Created = time.Now()

	manager := database.NewBuildManager()
	defer manager.Close()
	err = manager.Insert(reqBuild)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	err = mergeBuilds(reqBuild.BranchID, reqBuild.ID, context)
	if err != nil {
		return err
	}

	return context.JSON(http.StatusOK, reqBuild)
}

func DeleteBuildCmd(context echo.Context) error {

	buildID := context.QueryParam("id")
	if buildID == "" {
		return utils.BuildBadRequestError(context, models.ErrorInvalidRequest, "Build ID is required")
	}

	managerBD := database.NewBuildDepotManager()
	defer managerBD.Close()
	buildDepots, err := managerBD.FindByBuildID(buildID)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	for _, bd := range buildDepots {

		if bd.LinkID == "" {

			err = utils.RemoveDepotFiles(context.Request().Header.Get("ClientID"), bd.DepotID, context)
			if err != nil {
				return err
			}

			managerD := database.NewDepotManager()
			defer managerD.Close()
			err = managerD.RemoveByID(bd.DepotID)
			if err != nil {
				return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
			}
		}

		err = managerBD.RemoveByID(bd.ID)
		if err != nil {
			return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
		}
	}

	manager := database.NewBuildManager()
	defer manager.Close()
	err = manager.RemoveByID(buildID)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	return context.NoContent(http.StatusOK)
}

func GetBuildCmd(context echo.Context) error {

	bid := context.QueryParam("id")
	if bid == "" {
		return utils.BuildBadRequestError(context, models.ErrorInvalidRequest, "Build ID is required")
	}

	manager := database.NewBuildManager()
	defer manager.Close()
	build, err := manager.FindByID(bid)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	if build == nil {
		return utils.BuildBadRequestError(context, models.ErrorInvalidRequest, "Invalid build id: "+bid)
	}

	return context.JSON(http.StatusOK, build)
}

func ListBuildCmd(context echo.Context) error {

	result, ok, err := findBranch(context, "id", "name", "gid")
	if !ok {
		return err
	}

	manager := database.NewBuildManager()
	defer manager.Close()
	builds, err := manager.FindBuildByBranchID(result.ID)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	return context.JSON(http.StatusOK, builds)
}

func GetLiveBuildCmd(context echo.Context) error {

	result, ok, err := findBranch(context, "id", "name", "gid")
	if !ok {
		return err
	}

	manager := database.NewBuildManager()
	defer manager.Close()
	build, err := manager.FindByID(result.LiveBuild)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	if build == nil {
		return utils.BuildBadRequestError(context, models.ErrorInvalidRequest, "There are no one Live Build")
	}

	return context.JSON(http.StatusOK, build)
}

func PublishBuildCmd(context echo.Context) error {

	result, ok, err := findBranch(context, "id", "name", "gid")
	if !ok {
		return err
	}

	buildID := context.QueryParam("build")
	if buildID == "" {
		buildID = result.LiveBuild
	}

	manager := database.NewBuildManager()
	defer manager.Close()
	build, err := manager.FindByID(buildID)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	if build == nil {
		return utils.BuildBadRequestError(context, models.ErrorInvalidRequest, "Invalid build ID or Branch has no Live Build")
	}

	platform := context.QueryParam("platform")

	fpath, err := utils.GetUserBuildDepotPath(context.Request().Header.Get("ClientID"), build.ID, platform, context, false)
	if err != nil {
		return err
	}

	zap.S().Infow("PublishBuildCmd", zap.String("GetUserBuildDepotPath", fpath))

	announceList := strings.Split(config.Get().Tracker.TrackersList, ";")
	urlList := strings.Split(config.Get().Tracker.TrackersUrlList, ";")

	for i, _ := range urlList {
		urlList[i] = urlList[i] + "/" + buildID
	}

	ignoreFiles := map[string]bool{}

	targetFile := path.Join(fpath, "torrent.torrent")
	fpath = path.Join(fpath, "content")

	zap.S().Infow("PublishBuildCmd", zap.String("CreateTorrent", fpath))

	err = cord.CreateTorrent(
		fpath,
		targetFile,
		announceList,
		urlList,
		ignoreFiles,
		512,
		true,
	)

	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorCreateTorrent, err.Error())
	}

	infoHash, err := cord.GetInfoHash(targetFile)
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorCreateTorrent, err.Error())
	}

	api := cordapi.NewCordAPI(config.Get().Tracker.Url)
	err = api.Login(config.Get().Tracker.User, config.Get().Tracker.Password)
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorLoginTracker, err.Error())
	}

	err = api.AddTorrent(&models.TorrentCmd{infoHash})
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorAddTorrent, err.Error())
	}

	var s3Args s3.Args
	s3Args.SourceDir = fpath
	s3Args.OutputDir = buildID + "/content"
	s3Args.ID = config.Get().AwsS3.ID
	s3Args.Key = config.Get().AwsS3.Key
	s3Args.Region = config.Get().AwsS3.Region
	s3Args.S3Bucket = config.Get().AwsS3.Bucket
	err = s3.Upload(s3Args)
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorFileIOFailure, err.Error())
	}

	return context.JSON(http.StatusOK, result)
}

func filterFiles(manifest *models.ConfigManifest, locale string, fpath string, files []string, context echo.Context) ([]string, error) {

	var result []string

	for _, f := range files {

		relpath, err := filepath.Rel(fpath, f)
		if err != nil {
			return nil, utils.BuildInternalServerError(context, models.ErrorFileIOFailure, err.Error())
		}

		useFile := true

		if locale != "" {

			index := len("content/")

			for _, l := range manifest.Locales {

				rpath, _ := filepath.Split(relpath)
				rpath = filepath.ToSlash(rpath)

				match := strings.Index(rpath, l.LocalRoot)
				if match == index || (rpath == "content/" && l.LocalRoot == "./") {

					if locale != l.Locale {
						useFile = false
					}
				}
			}
		}

		if useFile {
			result = append(result, relpath)
		}
	}

	return result, nil
}

func getBuildVersion(cfgFile string, platform string, context echo.Context) (string, error) {

	cfg, err := utils.ReadConfigFile(cfgFile, &context)
	if err != nil {
		return "", err
	}

	for _, m := range cfg.Application.Manifests {

		if m.Platform == platform {
			return m.Version, nil
		}
	}

	return "", nil
}

func prepareConfig(configFile string, platform string, context echo.Context) ([]byte, error) {

	cfg, err := utils.ReadConfigFile(configFile, &context)
	if err != nil {
		return nil, err
	}

	var manifest *models.ConfigManifest
	manifest = nil

	for _, m := range cfg.Application.Manifests {

		if m.Platform == platform {
			manifest = &m
			break
		}
	}

	manager := database.NewRedistrManager()
	defer manager.Close()

	for i, r := range manifest.Redistributables {

		redistr, err := manager.FindByName(r)
		if err != nil {
			return nil, utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
		}

		if redistr != nil {
			manifest.Redistributables[i] = redistr.Url
		} else {
			manifest.Redistributables[i] = ""
		}
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, utils.BuildBadRequestError(context, models.ErrorInvalidJSONFormat, err.Error())
	}

	return data, nil
}

func GetUpdateInfoCmd(context echo.Context) error {

	result, ok, err := findBranch(context, "id", "name", "gid")
	if !ok {
		return err
	}

	if result.LiveBuild == "" {
		return utils.BuildBadRequestError(context, models.ErrorBuildIsNotPublished, "")
	}

	platform := context.QueryParam("platform")
	fpath, err := utils.GetUserBuildDepotPath(context.Request().Header.Get("ClientID"), result.LiveBuild, platform, context, false)
	if err != nil {
		return err
	}

	configFile := path.Join(fpath, "config.json")
	currVer, err := getBuildVersion(configFile, platform, context)
	if err != nil {
		return err
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) { // the file is not exist
		return utils.BuildBadRequestError(context, models.ErrorConfigFileNotFound, "")
	}

	torrentFile := path.Join(fpath, "torrent.torrent")
	if _, err := os.Stat(torrentFile); os.IsNotExist(err) { // the file is not exist
		return utils.BuildBadRequestError(context, models.ErrorBuildIsNotPublished, "")
	}

	info := &models.UpdateInfo{}
	info.BuildID = result.LiveBuild
	info.Version = currVer

	info.ConfigData, err = prepareConfig(configFile, platform, context)
	if err != nil {
		return err
	}

	info.TorrentData, err = ioutil.ReadFile(torrentFile)
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorFileIOFailure, err.Error())
	}

	return context.JSON(http.StatusOK, info)
}

func GetUpdatePatchCmd(context echo.Context) error {

	result, ok, err := findBranch(context, "id", "name", "gid")
	if !ok {
		return err
	}

	if result.LiveBuild == "" {
		return utils.BuildBadRequestError(context, models.ErrorBuildIsNotPublished, "")
	}

	platform := context.QueryParam("platform")
	fpath, err := utils.GetUserBuildDepotPath(context.Request().Header.Get("ClientID"), result.LiveBuild, platform, context, false)
	if err != nil {
		return err
	}

	configFile := path.Join(fpath, "config.json")
	if _, err := os.Stat(configFile); os.IsNotExist(err) { // the file is not exist
		return utils.BuildBadRequestError(context, models.ErrorConfigFileNotFound, "")
	}

	version := context.QueryParam("ver")
	torrentPatchFile := path.Join(fpath, "patch_for_"+version+"_"+platform+".torrent")
	if _, err := os.Stat(torrentPatchFile); os.IsNotExist(err) { // the file is not exist
		//return utils.BuildBadRequestError(context, models.ErrorNoUpdateAvailable, version)
		torrentPatchFile = ""
	}

	torrentFile := path.Join(fpath, "torrent.torrent")
	if _, err := os.Stat(torrentFile); os.IsNotExist(err) { // the file is not exist
		return utils.BuildBadRequestError(context, models.ErrorBuildIsNotPublished, "")
	}

	info := &models.UpdateInfo{}
	info.BuildID = result.LiveBuild

	info.ConfigData, err = prepareConfig(configFile, platform, context)
	if err != nil {
		return err
	}

	info.TorrentData, err = ioutil.ReadFile(torrentFile)
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorFileIOFailure, err.Error())
	}

	info.TorrentPatchData, err = ioutil.ReadFile(torrentPatchFile)
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorFileIOFailure, err.Error())
	}

	return context.JSON(http.StatusOK, info)
}
