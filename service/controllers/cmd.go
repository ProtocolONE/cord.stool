package controllers

import (
	"cord.stool/service/core/utils"
	"cord.stool/service/models"
	utils2 "cord.stool/utils"
	"cord.stool/xdelta"

	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/labstack/echo"
	"go.uber.org/zap"
)

func UploadCmd(context echo.Context) error {

	reqUpload := &models.UploadCmd{}
	err := context.Bind(reqUpload)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorInvalidJSONFormat, err.Error())
	}

	fpath, err := utils.GetUserBuildPath(context.Request().Header.Get("ClientID"), reqUpload.BuildID)
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorGetUserStorage, err.Error())
	}

	if !reqUpload.Config {
		fpath = path.Join(fpath, "content")
		fpath = path.Join(fpath, reqUpload.FilePath)
	} else {
		reqUpload.FileName = "config.json"
	}

	zap.S().Infow("Uploading", zap.String("path", fpath))

	err = os.MkdirAll(fpath, 0777)
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorFileIOFailure, err.Error())
	}

	fpath = path.Join(fpath, reqUpload.FileName)
	if reqUpload.Config {
		// Checking file format
		_, err := utils.ReadConfigFile(fpath, &context)
		if err != nil {
			return err
		}
	}

	if reqUpload.Patch {

		fpath = fpath[0:(len(fpath) - len(".diff"))]
		patchfile, err := ioutil.TempFile(os.TempDir(), "patch")
		if err != nil {
			return utils.BuildInternalServerError(context, models.ErrorFileIOFailure, err.Error())
		}
		defer os.Remove(patchfile.Name())
		patchfile.Close()

		err = ioutil.WriteFile(patchfile.Name(), reqUpload.FileData, 0777)
		if err != nil {
			return utils.BuildInternalServerError(context, models.ErrorFileIOFailure, err.Error())
		}

		fpathold := fpath
		if _, err := os.Stat(fpathold); os.IsNotExist(err) { // the file is not exist
			fpathold = "NUL" // fake name
		}

		err = xdelta.DecodeDiff(fpathold, fpath, patchfile.Name())
		if err != nil {
			return utils.BuildInternalServerError(context, models.ErrorApplyPatch, err.Error())
		}

	} else {

		err = ioutil.WriteFile(fpath, reqUpload.FileData, 0777)
		if err != nil {
			return utils.BuildInternalServerError(context, models.ErrorFileIOFailure, err.Error())
		}
	}

	return context.NoContent(http.StatusOK)
}

func CompareHashCmd(context echo.Context) error {

	reqCmp := &models.CompareHashCmd{}
	err := context.Bind(reqCmp)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorInvalidJSONFormat, err.Error())
	}

	fpath, err := utils.GetUserBuildPath(context.Request().Header.Get("ClientID"), reqCmp.BuildID)
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorGetUserStorage, err.Error())
	}

	fpath = path.Join(fpath, "content")
	fpath = path.Join(fpath, reqCmp.FilePath)
	fpath = path.Join(fpath, reqCmp.FileName)

	hash, err := utils2.Md5(fpath)
	equal := (err == nil && reqCmp.FileHash == hash)

	return context.JSON(http.StatusOK, models.CompareHashCmdResult{Equal: equal})
}
