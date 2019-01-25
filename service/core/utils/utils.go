package utils

import (
    "cord.stool/service/database"

	"fmt"
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
