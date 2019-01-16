package controllers

import (
	"cord.stool/service/models"
    "cord.stool/service/database"

	"encoding/json"
	"net/http"
	"path"
	"os"
	"io/ioutil"
	"fmt"

    "go.uber.org/zap"
    "gopkg.in/mgo.v2/bson"
)

func UploadCmd(w http.ResponseWriter, r *http.Request) {

	ClientID := r.Header.Get("ClientID")
	dbc := database.Get("users")

    var dbUsers []models.Authorisation
    err := dbc.Find(bson.M{"username": ClientID}).All(&dbUsers)
    if err != nil {
        zap.S().Errorf("Cannot find user %s, err: %v", ClientID, err)
        w.WriteHeader(http.StatusInternalServerError)
        response, _ := json.Marshal(models.Error{Message: fmt.Sprintf("Cannot find user %s.", ClientID)})
		w.Write(response)
		return
	}

	userRoot := dbUsers[0].Storage
	
    reqUpload := new(models.UploadCmd)
    decoder := json.NewDecoder(r.Body)
    decoder.Decode(&reqUpload)

	fpath := path.Join(userRoot, reqUpload.FilePath)
	err = os.MkdirAll(fpath, 0777)
	if err != nil {
        zap.S().Errorf("Cannot create path %s, err: %v", fpath, err)
        w.WriteHeader(http.StatusInternalServerError)
        response, _ := json.Marshal(models.Error{Message: fmt.Sprintf("Cannot create path %s.", fpath)})
		w.Write(response)
		return
	}

	fpath = path.Join(fpath, reqUpload.FileName)

	err = ioutil.WriteFile(fpath, reqUpload.FileData, 0777)
	if err != nil {
        zap.S().Errorf("Cannot write to file %s, err: %v", fpath, err)
        w.WriteHeader(http.StatusInternalServerError)
        response, _ := json.Marshal(models.Error{Message: fmt.Sprintf("Cannot write to file %s.", fpath)})
        w.Write(response)
		return
	} 

    w.WriteHeader(http.StatusOK)
}
