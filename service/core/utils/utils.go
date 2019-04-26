package utils

import (
	"cord.stool/service/database"
	"cord.stool/service/models"
	"fmt"
	"github.com/labstack/echo"
	"net/http"
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
	case models.ErrorAddTracker:
		errorText = "Cannot add torrent"
	case models.ErrorDeleteTracker:
		errorText = "Cannot remove torrent"
	case models.ErrorGetUserStorage:
		errorText = "Cannot get user storage"
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
