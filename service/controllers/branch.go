package controllers

import (
	"cord.stool/service/core/utils"
	"cord.stool/service/database"
	"cord.stool/service/models"
	utils2 "cord.stool/utils"

	"net/http"
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
			return nil, false, utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, bid)
		}
	} else {
		result, err = manager.FindByName(name, gid)
		if err != nil {
			return nil, false, utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
		}
		if result == nil {
			return nil, false, utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, name)
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

	var result *models.Branch
	gid := context.QueryParam("gid")
	if gid == "" {

		result, ok, err := findBranch(context, "id", "name", "gid")
		if !ok {
			return err
		}
		gid = result.GameID
	}

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
		return utils.BuildInternalServerError(context, models.ErrorInternalError, err.Error())
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

	targetBranch.BuildID = sourceBranch.BuildID
	targetBranch.Updated = time.Now()

	manager := database.NewBranchManager()
	err = manager.Update(targetBranch)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	return context.JSON(http.StatusOK, models.ShallowBranchCmdResult{sourceBranch.ID, sourceBranch.Name, targetBranch.ID, targetBranch.Name})
}
