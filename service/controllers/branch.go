package controllers

import (
	"cord.stool/service/database"
	"cord.stool/service/models"
	"github.com/pborman/uuid"

	"fmt"
	"go.uber.org/zap"
	"net/http"
	"time"

	"github.com/labstack/echo"
)

func CreateBranchCmd(context echo.Context) error {

	reqBranch := &models.BranchInfoCmd{}
	err := context.Bind(reqBranch)
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Error{models.ErrorInvalidJSONFormat, "Invalid JSON format: " + err.Error()})
	}

	manager := database.NewBranchManager()
	result, err := manager.FindByName(reqBranch.Name)
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Error{models.ErrorReadDataBase, fmt.Sprintf("Cannot read from database, error: %s", err.Error())})
	}

	if result != nil {
		return context.JSON(http.StatusBadRequest, models.Error{models.ErrorUserAlreadyExists, fmt.Sprintf("Branch %s already exists", reqBranch.Name)})
	}

	branchID := uuid.New()
	err = manager.Insert(&models.Branch{branchID, reqBranch.Name, reqBranch.GameID, "", time.Now()})
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Error{models.ErrorCreateUser, fmt.Sprintf("Cannot create branch %s, error: %s", reqBranch.Name, err.Error())})
	}

	zap.S().Infow("Created new branch", zap.String("branch id", branchID))
	return context.JSON(http.StatusOK, models.BranchInfoCmd{branchID, reqBranch.Name, reqBranch.GameID})
}

func DeleteBranchCmd(context echo.Context) error {

	reqBranch := &models.BranchInfoCmd{}
	err := context.Bind(reqBranch)
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Error{models.ErrorInvalidJSONFormat, "Invalid JSON format: " + err.Error()})
	}

	manager := database.NewBranchManager()
	result, err := manager.FindByID(reqBranch.ID)
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Error{models.ErrorReadDataBase, fmt.Sprintf("Cannot read from database, error: %s", err.Error())})
	}

	if result == nil {
		result, err = manager.FindByName(reqBranch.Name)
		if err != nil {
			return context.JSON(http.StatusBadRequest, models.Error{models.ErrorReadDataBase, fmt.Sprintf("Cannot read from database, error: %s", err.Error())})
		}
	}

	err = manager.RemoveByID(result.ID)
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Error{models.ErrorDeleteUser, fmt.Sprintf("Cannot delete branch %s, error: %s", reqBranch, err.Error())})
	}

	zap.S().Infow("Removed branch", zap.String("branch id", result.ID))

	return context.JSON(http.StatusOK, models.BranchInfoCmd{result.ID, result.Name, result.GameID})
}

func ListBranchCmd(context echo.Context) error {

	reqBranch := &models.ListBranchCmd{}
	err := context.Bind(reqBranch)
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Error{models.ErrorInvalidJSONFormat, "Invalid JSON format: " + err.Error()})
	}

	manager := database.NewBranchManager()
	branches, err := manager.List(reqBranch.GameID)
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Error{models.ErrorReadDataBase, fmt.Sprintf("Cannot read from database, error: %s", err.Error())})
	}

	return context.JSON(http.StatusOK, models.ListBranchCmdResult{branches})
}

func ShallowBranchCmd(context echo.Context) error {

	reqBranch := &models.ShallowBranchCmd{}
	err := context.Bind(reqBranch)
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Error{models.ErrorInvalidJSONFormat, "Invalid JSON format: " + err.Error()})
	}

	manager := database.NewBranchManager()
	branches, err := manager.Shallow(reqBranch.SourceNameOrID, reqBranch.TargetNameOrID)
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Error{models.ErrorReadDataBase, fmt.Sprintf("Cannot read from database, error: %s", err.Error())})
	}

	return context.JSON(http.StatusOK, models.ShallowBranchCmdResult{branches[0].ID, branches[0].Name, branches[1].ID, branches[1].Name})
}
