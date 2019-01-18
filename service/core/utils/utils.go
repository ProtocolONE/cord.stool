package utils

import (
	"cord.stool/service/models"
    "cord.stool/service/database"

	"encoding/json"
	"net/http"
	"fmt"

	"go.uber.org/zap"
    "gopkg.in/mgo.v2/bson"
)

func ServiceError(w http.ResponseWriter, status int, message string, err error) {

	if err != nil {
		message += fmt.Sprintf(". Error: %s", err.Error())	
	}

	zap.S().Errorf(message)

    w.WriteHeader(status)
    response, _ := json.Marshal(models.Error{Message: message})
	w.Write(response)
}

func GetUserStorage(clientID string) (string, error) {

	dbc := database.Get("users")

    var dbUsers []models.Authorization
    err := dbc.Find(bson.M{"username": clientID}).All(&dbUsers)
    if err != nil {
		return "", fmt.Errorf("Cannot find user %s, error: %s", clientID, err.Error())
	}

	return dbUsers[0].Storage, nil
}
