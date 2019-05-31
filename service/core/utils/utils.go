package utils

import (
	"cord.stool/service/database"
	"cord.stool/service/models"
	utils2 "cord.stool/utils"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"time"
)

func BuildError(context echo.Context, status int, code int, message string) error {

	var errorText string

	switch code {
	case models.ErrorInvalidJSONFormat:
		errorText = "Invalid JSON format"
	case models.ErrorDatabaseFailure:
		errorText = "Database failure"
	case models.ErrorAlreadyExists:
		errorText = "Object already exists"
	case models.ErrorNotFound:
		errorText = "Object is not found"
	case models.ErrorInvalidRequest:
		errorText = "Invalid request"
	case models.ErrorInternalError:
		errorText = "Internal server error"
	case models.ErrorGenUserStorageName:
		errorText = "Cannot generate user files storage name"
	case models.ErrorInvalidUsernameOrPassword:
		errorText = "Invalid username or password"
	case models.ErrorGenToken:
		errorText = "Cannot generate access-token"
	case models.ErrorLogout:
		errorText = "Logout failed"
	case models.ErrorLoginTracker:
		errorText = "Login to tracker failed"
	case models.ErrorAddTorrent:
		errorText = "Cannot regist torrent file"
	case models.ErrorDeleteTorrent:
		errorText = "Cannot unregist torrent file"
	case models.ErrorGetUserStorage:
		errorText = "Cannot get user build storage"
	case models.ErrorFileIOFailure:
		errorText = "File IO failure"
	case models.ErrorApplyPatch:
		errorText = "Cannot apply patch"
	case models.ErrorWharfLibrary:
		errorText = "Wharf Library failed"
	case models.ErrorUnauthorized:
		errorText = "Authorization failed"
	case models.ErrorTokenExpired:
		errorText = "Token is expired"
	case models.ErrorInvalidToken:
		errorText = "Invalid token"
	case models.ErrorCreateTorrent:
		errorText = "Cannot create torrent file"
	case models.ErrorBuildIsNotPublished:
		errorText = "The branch has no published build"
	case models.ErrorInvalidPlatformName:
		errorText = "Invalid platform name"
	case models.ErrorInvalidBuildPlatform:
		errorText = "The build platform is not matched the specified platform"
	default:
		errorText = "Unknown error"
	}

	if message != "" {
		errorText += ": " + message
	} else {
		errorText += "."
	}

	return context.JSON(status, models.Error{code, errorText})
}

func BuildBadRequestError(context echo.Context, code int, message string) error {

	return BuildError(context, http.StatusBadRequest, code, message)
}

func BuildInternalServerError(context echo.Context, code int, message string) error {

	return BuildError(context, http.StatusInternalServerError, code, message)
}

func BuildUnauthorizedError(context echo.Context, code int, message string) error {

	return BuildError(context, http.StatusUnauthorized, code, message)
}

func ReadConfigData(data []byte, context *echo.Context) (*models.Config, error) {

	config := &models.Config{}
	err := json.Unmarshal(data, config)
	if err != nil {
		if context != nil {
			return nil, BuildBadRequestError(*context, models.ErrorInvalidJSONFormat, err.Error())
		}
		return nil, err
	}

	return config, nil
}

func ReadConfigFile(path string, context *echo.Context) (*models.Config, error) {

	data, err := ioutil.ReadFile(path)
	if err != nil {
		if context != nil {
			return nil, BuildInternalServerError(*context, models.ErrorFileIOFailure, err.Error())
		}
		return nil, err
	}

	return ReadConfigData(data, context)
}

func GetUserStorage(clientID string) (string, error) {

	manager := database.NewUserManager()
	users, err := manager.FindByName(clientID)
	if err != nil {
		return "", fmt.Errorf("Cannot find user %s, error: %s", clientID, err.Error())
	}

	if len(users) > 1 {
		return "", fmt.Errorf("Duplicate users %s", clientID)
	}

	return users[0].Storage, nil
}

func GetUserBuildDepotPath(clientID string, buildID string, platform string, context echo.Context, createDepot bool) (string, error) {

	manager := database.NewBuildManager()
	build, err := manager.FindByID(buildID)
	if err != nil {
		return "", BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	if build == nil {
		return "", BuildBadRequestError(context, models.ErrorInvalidRequest, fmt.Sprintf("Cannot find the specified build %s", buildID))
	}

	storage, err := GetUserStorage(clientID)
	if err != nil {
		return "", BuildInternalServerError(context, models.ErrorGetUserStorage, err.Error())
	}

	managerBD := database.NewBuildDepotManager()
	buildDepot, err := managerBD.FindByBuildAndPlatformID(buildID, platform)
	if err != nil {
		return "", BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	DepotID := ""

	if (buildDepot == nil || buildDepot.LinkID != "") && createDepot {

		depot := &models.Depot{utils2.GenerateID(), time.Now(), platform}
		managerD := database.NewDepotManager()
		err = managerD.Insert(depot)
		if err != nil {
			return "", BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
		}

		if buildDepot != nil && buildDepot.LinkID != "" {

			linkDuildDepot, err := managerBD.FindByID(buildDepot.LinkID)
			if err != nil {
				return "", BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
			}

			if linkDuildDepot.Platform != platform {
				createDepot = true
			} else {
				createDepot = false
			}
		}

		if createDepot {

			buildDepot := &models.BuildDepot{utils2.GenerateID(), buildID, depot.ID, "", platform, time.Now()}
			err = managerBD.Insert(buildDepot)
			if err != nil {
				return "", BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
			}

		} else {

			buildDepot.DepotID = depot.ID
			buildDepot.LinkID = ""
			err = managerBD.Update(buildDepot)
			if err != nil {
				return "", BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
			}
		}

		DepotID = depot.ID

	} else if buildDepot != nil {

		DepotID = buildDepot.DepotID
	}

	if DepotID == "" {
		return "", BuildBadRequestError(context, models.ErrorInvalidBuildPlatform, platform)
	}

	fpath := filepath.Join(storage, DepotID)
	return fpath, nil

}
