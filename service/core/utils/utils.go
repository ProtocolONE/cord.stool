package utils

import (
	"cord.stool/service/models"
    "cord.stool/service/database"

	"fmt"

    "github.com/labstack/echo"
)

func ServiceError(context echo.Context, status int, message string, err error) {

	if err != nil {
		message += fmt.Sprintf(". Error: %s", err.Error())	
	}

	context.Echo().Logger.Error(message)

	context.JSON(status, models.Error{Message: message})
}

func GetUserStorage(clientID string) (string, error) {

	/*dbc := database.Get("users")

    var dbUsers []models.Authorization
    err := dbc.Find(bson.M{"username": clientID}).All(&dbUsers)
    if err != nil {
		return "", fmt.Errorf("Cannot find user %s, error: %s", clientID, err.Error())
	}

	return dbUsers[0].Storage, nil*/

	manager := database.GeUserManager()
	users, err := manager.FindByName(clientID)
    if err != nil {
		return "", fmt.Errorf("Cannot find user %s, error: %s", clientID, err.Error())
	}

	if len(users) > 1 {
		return "", fmt.Errorf("Duplicate users %s", clientID)
	}

	return users[0].Storage, nil
}
