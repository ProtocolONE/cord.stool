package controllers

import (
	"cord.stool/service/core/utils"
	"cord.stool/service/models"
	utils2 "cord.stool/utils"
	"cord.stool/xdelta"

	"fmt"
	"io/ioutil"
	"os"
	"path"
	"net/http"

	"github.com/labstack/echo"
)

func CreateBranchCmd(context echo.Context) error {

	reqBranch := &models.CreateBranchCmd{}
	err := context.Bind(reqBranch)
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Error{models.ErrorInvalidJSONFormat, "Invalid JSON format: " + err.Error()})
	}

	manager := database.NewBranchManager()
	branch, err := manager.FindByNameOrID(reqBranch.NameOrID)
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Error{models.ErrorReadDataBase, fmt.Sprintf("Cannot read from database, error: %s", err.Error())})
	}

	if len(branch) != 0 {
		return context.JSON(http.StatusBadRequest, models.Error{models.ErrorUserAlreadyExists, fmt.Sprintf("Branch %s already exists", reqBranch.NameOrID)})
	}

	err = manager.Insert(&models.Branch{})
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Error{models.ErrorCreateUser, fmt.Sprintf("Cannot create branch %s, error: %s", reqBranch.NameOrID, err.Error())})
	}

	zap.S().Infow("Created new branch", zap.String("branch", reqBranch.NameOrID))

	return context.NoContent(http.StatusCreated)
}

func DeleteBranchCmd(context echo.Context) error {

	reqBranch := &models.DeleteBranchCmd{}
	err := context.Bind(reqBranch)
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Error{models.ErrorInvalidJSONFormat, "Invalid JSON format: " + err.Error()})
	}

	manager := database.NewBranchManager()
	err = manager.RemoveByName(reqUser.reqBranch)
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Error{models.ErrorDeleteUser, fmt.Sprintf("Cannot delete branch %s, error: %s", reqBranch, err.Error())})
	}

	zap.S().Infow("Removed branch", zap.String("username", reqBranch))
	return context.NoContent(http.StatusOK)
}

func ListBranchCmd(context echo.Context) error {

	reqBranch := &models.ListBranchCmd{}
	err := context.Bind(reqBranch)
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Error{models.ErrorInvalidJSONFormat, "Invalid JSON format: " + err.Error()})
	}

	manager := database.NewBranchManager()
	branches, err := manager.ListBranch(reqBranch.GameID)
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Error{models.ErrorReadDataBase, fmt.Sprintf("Cannot read from database, error: %s", err.Error())})
	}

	// to do

	return context.NoContent(http.StatusOK)
}

func ShallowBranchCmd(context echo.Context) error {

	reqBranch := &models.ShallowBranchCmd{}
	err := context.Bind(reqBranch)
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Error{models.ErrorInvalidJSONFormat, "Invalid JSON format: " + err.Error()})
	}

	manager := database.NewBranchManager()
	branches, err := manager.ShallowBranch(reqBranch.SourceNameOrID, reqBranch.SourceNameOrID.TargetNameOrID)
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Error{models.ErrorReadDataBase, fmt.Sprintf("Cannot read from database, error: %s", err.Error())})
	}

	// to do

	return context.NoContent(http.StatusOK)
}

