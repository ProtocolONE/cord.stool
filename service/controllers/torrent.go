package controllers

import (
	"cord.stool/service/models"
	"cord.stool/cordapi"

	"fmt"
	"net/http"
	"go.uber.org/zap"

	"github.com/labstack/echo"
)

func AddTorrent(context echo.Context) error {
	
	reqTorrent := &models.TorrentCmd{}
	err := context.Bind(reqTorrent)
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Error{models.ErrorInvalidJSONFormat, "Invalid JSON format: " + err.Error()})
	}

	api := cordapi.NewCordAPI("http://127.0.0.1:5002")
	err = api.Login("admin", "123456")
	if err != nil {
		return context.JSON(http.StatusInternalServerError, models.Error{models.ErrorLoginTracker, fmt.Sprintf("Login to tracker failed, error: %s", err.Error())})
	}

	err = api.AddTorrent(reqTorrent)
	if err != nil {
		return context.JSON(http.StatusInternalServerError, models.Error{models.ErrorAddTracker, fmt.Sprintf("Cannot add torrent %s, error: %s", reqTorrent.InfoHash, err.Error())})
	}

	zap.S().Infow("Added torrent", zap.String("info_hash", reqTorrent.InfoHash))

	return context.NoContent(http.StatusOK)
}

func DeleteTorrent(context echo.Context) error {

	reqTorrent := &models.TorrentCmd{}
	err := context.Bind(reqTorrent)
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Error{models.ErrorInvalidJSONFormat, "Invalid JSON format: " + err.Error()})
	}

	api := cordapi.NewCordAPI("http://127.0.0.1:5002")
	err = api.Login("admin", "123456")
	if err != nil {
		return context.JSON(http.StatusInternalServerError, models.Error{models.ErrorLoginTracker, fmt.Sprintf("Login to tracker failed, error: %s", err.Error())})
	}

	err = api.RemoveTorrent(reqTorrent)
	if err != nil {
		return context.JSON(http.StatusInternalServerError, models.Error{models.ErrorDeleteTracker, fmt.Sprintf("Cannot remove torrent %s, error: %s", reqTorrent.InfoHash, err.Error())})
	}

	zap.S().Infow("Removed torrent", zap.String("info_hash", reqTorrent.InfoHash))

	return context.NoContent(http.StatusOK)
}
