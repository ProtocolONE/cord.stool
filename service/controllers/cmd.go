package controllers

import (
	"cord.stool/service/core/utils"
	"cord.stool/service/models"
	utils2 "cord.stool/utils"

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

	if reqUpload.Config {
		// Checking file format
		_, err := utils.ReadConfigData(reqUpload.FileData, &context)
		if err != nil {
			return err
		}
	}

	fpath, err := utils.GetUserBuildDepotPath(context.Request().Header.Get("ClientID"), reqUpload.BuildID, reqUpload.Platform, context, true)
	if err != nil {
		return err
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

	err = ioutil.WriteFile(fpath, reqUpload.FileData, 0777)
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorFileIOFailure, err.Error())
	}

	return context.NoContent(http.StatusOK)
}

func CompareHashCmd(context echo.Context) error {

	reqCmp := &models.CompareHashCmd{}
	err := context.Bind(reqCmp)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorInvalidJSONFormat, err.Error())
	}

	fpath, err := utils.GetUserBuildDepotPath(context.Request().Header.Get("ClientID"), reqCmp.BuildID, reqCmp.Platform, context, false)
	if err != nil {
		return err
	}

	fpath = path.Join(fpath, "content")
	fpath = path.Join(fpath, reqCmp.FilePath)
	fpath = path.Join(fpath, reqCmp.FileName)

	hash, err := utils2.Md5(fpath)
	equal := (err == nil && reqCmp.FileHash == hash)

	return context.JSON(http.StatusOK, models.CompareHashCmdResult{Equal: equal})
}
