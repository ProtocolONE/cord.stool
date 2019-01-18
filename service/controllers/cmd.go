package controllers

import (
	"cord.stool/service/models"
	"cord.stool/xdelta"
    "cord.stool/service/core/utils"
	utils2 "cord.stool/utils"

	"encoding/json"
	"net/http"
	"path"
	"os"
	"io/ioutil"
	"fmt"

)

func UploadCmd(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	
	userRoot, err := utils.GetUserStorage(r.Header.Get("ClientID"))
    if err != nil {
		utils.ServiceError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	
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

		fpathold := fpath
		if _, err := os.Stat(fpathold); os.IsNotExist(err) { // the file is not exist
			fpathold = "NUL" // fake name
		}
		
		err = xdelta.DecodeDiff(fpathold, fpath, patchfile.Name())
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

func CompareHashCmd(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	userRoot, err := utils.GetUserStorage(r.Header.Get("ClientID"))
    if err != nil {
		utils.ServiceError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

    reqCmp := new(models.CompareHashCmd)
    decoder := json.NewDecoder(r.Body)
    decoder.Decode(&reqCmp)

	fpath := path.Join(userRoot, reqCmp.FilePath)
	fpath = path.Join(fpath, reqCmp.FileName)

	hash, err := utils2.Md5(fpath)
	equal := (err == nil && reqCmp.FileHash == hash)

    response, _ := json.Marshal(models.CompareHashCmdResult{Equal: equal})
    w.Write(response)
    w.WriteHeader(http.StatusOK)
}