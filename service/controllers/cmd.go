package controllers

import (
	"cord.stool/service/models"
	"cord.stool/xdelta"
    "cord.stool/service/core/utils"
    "cord.stool/service/core/authentication"
	utils2 "cord.stool/utils"

	"encoding/json"
	"net/http"
	"path"
	"os"
	"io/ioutil"
	"fmt"

    "github.com/labstack/echo"
)

func UploadCmd(context echo.Context) error {

	if !authentication.RequireTokenAuthentication(context) {
		return nil
	}

	userRoot, err := utils.GetUserStorage(context.Request().Header.Get("ClientID"))
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	
    reqUpload := new(models.UploadCmd)
    decoder := json.NewDecoder(context.Request().Body)
    decoder.Decode(&reqUpload)

	fpath := path.Join(userRoot, reqUpload.FilePath)
	err = os.MkdirAll(fpath, 0777)
	if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Cannot create path %s, error: %s", fpath, err.Error()))
	}

	fpath = path.Join(fpath, reqUpload.FileName)

	if reqUpload.Patch {

		fpath = fpath[0 : (len(fpath) - len(".diff"))]
		patchfile, err := ioutil.TempFile(os.TempDir(), "patch")
		if err != nil {
        	return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Cannot get temp file, error: %s", err.Error()))
		} 
		defer os.Remove(patchfile.Name())
		patchfile.Close()

		err = ioutil.WriteFile(patchfile.Name(), reqUpload.FileData, 0777)
		if err != nil {
        	return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Cannot write to file %s, error: %s", fpath, err.Error()))
		} 

		fpathold := fpath
		if _, err := os.Stat(fpathold); os.IsNotExist(err) { // the file is not exist
			fpathold = "NUL" // fake name
		}
		
		err = xdelta.DecodeDiff(fpathold, fpath, patchfile.Name())
		if err != nil {
        	return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Cannot apply patch to %s, error: %s", fpath, err.Error()))
		} 

	} else {

		err = ioutil.WriteFile(fpath, reqUpload.FileData, 0777)
		if err != nil {
        	return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Cannot write to file %s, error: %s", fpath, err.Error()))
		} 
	}

	return context.NoContent(http.StatusOK)
}

func CompareHashCmd(context echo.Context) error {

	if !authentication.RequireTokenAuthentication(context) {
		return nil
	}

	userRoot, err := utils.GetUserStorage(context.Request().Header.Get("ClientID"))
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

    reqCmp := new(models.CompareHashCmd)
    decoder := json.NewDecoder(context.Request().Body)
    decoder.Decode(&reqCmp)

	fpath := path.Join(userRoot, reqCmp.FilePath)
	fpath = path.Join(fpath, reqCmp.FileName)

	hash, err := utils2.Md5(fpath)
	equal := (err == nil && reqCmp.FileHash == hash)

	return context.JSON(http.StatusOK, models.CompareHashCmdResult{Equal: equal})
}