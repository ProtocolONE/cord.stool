package utils

import (
	"cord.stool/service/database"
	"cord.stool/service/models"
	"fmt"
	"github.com/labstack/echo"
	"net/http"
	"path/filepath"
	"encoding/json"
	"io/ioutil"
)

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

func GetUserBuildPath(clientID string, buildID string) (string, error) {

	storage, err := GetUserStorage(clientID)
	if err != nil {
		return "", err
	}

	manager := database.NewBuildManager()
	build, err := manager.FindByID(buildID)
	if err != nil {
		return "", err
	}

	if build == nil {
		return "", fmt.Errorf("Cannot find the specified build %s", buildID)
	}

	path := filepath.Join(storage, buildID)
	return path, nil
}

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
	default:
		errorText = "Unknown error"
	}

	if message != "" {
		errorText += ". " + message
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

func ReadConfigFile(path string, context *echo.Context) (*models.Config, error) {

	data, err := ioutil.ReadFile(path)
	if err != nil {
		if context != nil {
			return nil, BuildInternalServerError(*context, models.ErrorFileIOFailure, err.Error())
		}
		return nil, err
	}

	config := &models.Config{}
	err = json.Unmarshal(data, config)
	if err != nil {
		if context != nil {
			return nil, BuildBadRequestError(*context, models.ErrorInvalidJSONFormat, err.Error())
		}
		return nil, err
	}

	return config, nil
}