package controllers

import (
	"cord.stool/service/models"
    "cord.stool/service/config"
    "cord.stool/service/database"

	"encoding/json"
	"net/http"
	"path"
	"os"
	"io"
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
	}

	userRoot := dbUsers[0].Storage
	
	fpath := path.Join(config.Get().Service.StorageRootPath, userRoot, r.URL.Query().Get("storage"))
	err = os.MkdirAll(fpath, 0777)
	if err != nil {
        zap.S().Errorf("Cannot create path %s, err: %v", fpath, err)
        w.WriteHeader(http.StatusInternalServerError)
        response, _ := json.Marshal(models.Error{Message: fmt.Sprintf("Cannot create path %s.", fpath)})
		w.Write(response)
		return
	}

	fpath = path.Join(fpath, r.URL.Query().Get("name"))

	file, err := os.Create(fpath)
	if err != nil {
        zap.S().Errorf("Cannot create file %s, err: %v", fpath, err)
        w.WriteHeader(http.StatusInternalServerError)
        response, _ := json.Marshal(models.Error{Message: fmt.Sprintf("Cannot create file %s.", fpath)})
		w.Write(response)
		return
	}
	
	defer file.Close()

	n, err := io.Copy(file, r.Body)
	if err != nil {
        zap.S().Errorf("Cannot upload file to %s, err: %v", fpath, err)
        w.WriteHeader(http.StatusInternalServerError)
        response, _ := json.Marshal(models.Error{Message: fmt.Sprintf("Cannot upload file to %s.", fpath)})
        w.Write(response)
		return
	} 

	w.Write([]byte(fmt.Sprintf("%d bytes are recieved.\n", n)))
}

/*
func CreateCmd(w http.ResponseWriter, r *http.Request) {

	reqCmd := new(models.CreateCmd)
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&reqCmd)

	w.Write([]byte("CreateCmd is running.\n"))
	w.Write([]byte("source: " + reqCmd.Source + "\n"))
	w.Write([]byte("output: " + reqCmd.Output + "\n"))
}

func PushCmd(w http.ResponseWriter, r *http.Request) {

	reqCmd := new(models.PushCmd)
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&reqCmd)

	w.Write([]byte("PushCmd is running.\n"))
	w.Write([]byte("source: " + reqCmd.Source + "\n"))
	w.Write([]byte("output: " + reqCmd.Output + "\n"))
}

func DiffCmd(w http.ResponseWriter, r *http.Request) {

	reqCmd := new(models.DiffCmd)
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&reqCmd)

	w.Write([]byte("DiffCmd is running.\n"))
	w.Write([]byte("old: " + reqCmd.SourceOld + "\n"))
	w.Write([]byte("new: " + reqCmd.SourceNew + "\n"))
	w.Write([]byte("patch: " + reqCmd.OutputDiff + "\n"))
}

func TorrentCmd(w http.ResponseWriter, r *http.Request) {

	reqCmd := new(models.TorrentCmd)
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&reqCmd)

	w.Write([]byte("TorrentCmd API is running.\n"))
	w.Write([]byte("source: " + reqCmd.Source + "\n"))
	w.Write([]byte("target: " + reqCmd.Target + "\n"))
}
*/
