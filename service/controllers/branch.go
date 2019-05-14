package controllers

import (
	"cord.stool/cordapi"
	"cord.stool/service/config"
	"cord.stool/service/core/utils"
	"cord.stool/service/database"
	"cord.stool/service/models"
	"cord.stool/upload/cord"
	utils2 "cord.stool/utils"

	"net/http"
	"path"
	"time"

	"github.com/labstack/echo"
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
			return nil, false, utils.BuildBadRequestError(context, models.ErrorInvalidRequest, "Invalid branch name: "+name)
		}
	}

	return result, true, nil
}

func DeleteBranchCmd(context echo.Context) error {

	result, ok, err := findBranch(context, "id", "name", "gid")
	if !ok {
		return err
	}

	manager := database.NewBranchManager()
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
	err = manager.Update(targetBranch)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	return context.JSON(http.StatusOK, models.ShallowBranchCmdResult{sourceBranch.ID, sourceBranch.Name, targetBranch.ID, targetBranch.Name})
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
	err = manager.Insert(reqBuild)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	return context.JSON(http.StatusOK, reqBuild)
}

func GetBuildCmd(context echo.Context) error {

	bid := context.QueryParam("id")
	if bid == "" {
		return utils.BuildBadRequestError(context, models.ErrorInvalidRequest, "Build ID is required")
	}

	manager := database.NewBuildManager()
	build, err := manager.FindByID(bid)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	if build == nil {
		return utils.BuildBadRequestError(context, models.ErrorInvalidRequest, "Invalid build id: "+bid)
	}

	return context.JSON(http.StatusOK, build[0])
}

func ListBuildCmd(context echo.Context) error {

	result, ok, err := findBranch(context, "id", "name", "gid")
	if !ok {
		return err
	}

	manager := database.NewBuildManager()
	builds, err := manager.FindBuildByBranch(result.ID)
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
	build, err := manager.FindByID(result.LiveBuild)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	if build == nil || len(build) == 0 {
		return utils.BuildBadRequestError(context, models.ErrorInvalidRequest, "There are no one Live Build")
	}

	return context.JSON(http.StatusOK, build[0])
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
	build, err := manager.FindByID(buildID)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	if build == nil || len(build) == 0 {
		return utils.BuildBadRequestError(context, models.ErrorInvalidRequest, "Invalid build ID or Branch has no Live Build")
	}

	fpath, err := utils.GetUserBuildPath(context.Request().Header.Get("ClientID"), build[0].ID)
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorGetUserStorage, err.Error())
	}

	builtinAnnounceList := []string{
		"udp://tracker.openbittorrent.com:80",
		"udp://tracker.publicbt.com:80",
		"udp://tracker.istole.it:6969",
	}

	targetFile := path.Join(fpath, "torrent.torrent")

	err = cord.CreateTorrent(
		fpath,
		targetFile,
		builtinAnnounceList,
		nil,
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

	return context.JSON(http.StatusOK, build[0])
}

/*
func CloneLiveBuildCmd(context echo.Context) error {

	result, ok, err := findBranch(context, "id", "name", "gid")
	if !ok {
		return err
	}

	manager := database.NewBuildManager()
	build, err := manager.FindByID(result.LiveBuild)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	fpath, err := utils.GetUserBuildPath(context.Request().Header.Get("ClientID"), BuildID)
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorGetUserStorage, err.Error())
	}


	return context.JSON(http.StatusOK, build)
}
*/
