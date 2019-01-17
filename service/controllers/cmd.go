package controllers

import (
	"cord.stool/service/models"
    "cord.stool/service/database"
	"cord.stool/xdelta"
    "cord.stool/service/core/utils"

	"encoding/json"
	"net/http"
	"path"
	"os"
	"io/ioutil"
	"fmt"

    "gopkg.in/mgo.v2/bson"
)

func UploadCmd(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	
	ClientID := r.Header.Get("ClientID")
	dbc := database.Get("users")

    var dbUsers []models.Authorisation
    err := dbc.Find(bson.M{"username": ClientID}).All(&dbUsers)
    if err != nil {
		utils.ServiceError(w, http.StatusInternalServerError, fmt.Sprintf("Cannot find user %s", ClientID), err)
		return
	}

	userRoot := dbUsers[0].Storage
	
    reqUpload := new(models.UploadCmd)
    decoder := json.NewDecoder(r.Body)
    decoder.Decode(&reqUpload)

	fpath := path.Join(userRoot, reqUpload.FilePath)
	err = os.MkdirAll(fpath, 0777)
	if err != nil {
		utils.ServiceError(w, http.StatusInternalServerError, fmt.Sprintf("Cannot create path %s", fpath), err)
		return
	}

	fpath = path.Join(fpath, reqUpload.FileName)

	if reqUpload.Patch {

		fpath = fpath[0 : (len(fpath) - len(".diff"))]
		patchfile, err := ioutil.TempFile(os.TempDir(), "patch")
		if err != nil {
			utils.ServiceError(w, http.StatusInternalServerError, "Cannot get temp file", err)
			return
		} 
		defer os.Remove(patchfile.Name())
		patchfile.Close()

		err = ioutil.WriteFile(patchfile.Name(), reqUpload.FileData, 0777)
		if err != nil {
			utils.ServiceError(w, http.StatusInternalServerError, fmt.Sprintf("Cannot write to file %s", fpath), err)
			return
		} 

		err = xdelta.DecodeDiff(fpath, fpath, patchfile.Name())
		if err != nil {
			utils.ServiceError(w, http.StatusInternalServerError, fmt.Sprintf("Cannot apply patch to %s", fpath), err)
			return
		} 

	} else {

		err = ioutil.WriteFile(fpath, reqUpload.FileData, 0777)
		if err != nil {
			utils.ServiceError(w, http.StatusInternalServerError, fmt.Sprintf("Cannot write to file %s", fpath), err)
			return
		} 
	}

    w.WriteHeader(http.StatusOK)
}
