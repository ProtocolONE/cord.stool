package controllers

import (
	"cord.stool/cordapi"
	"cord.stool/service/config"
	"cord.stool/service/core/utils"
	"cord.stool/service/models"

	"go.uber.org/zap"
	"net/http"

	"github.com/labstack/echo"
)

func AddTorrent(context echo.Context) error {

	reqTorrent := &models.TorrentCmd{}
	err := context.Bind(reqTorrent)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorInvalidJSONFormat, err.Error())
	}

	api := cordapi.NewCordAPI(config.Get().Tracker.Url)
	err = api.Login(config.Get().Tracker.User, config.Get().Tracker.Password)
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorLoginTracker, err.Error())
	}

	err = api.AddTorrent(reqTorrent)
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorAddTorrent, err.Error())
	}

	zap.S().Infow("Added torrent", zap.String("info_hash", reqTorrent.InfoHash))

	return context.NoContent(http.StatusOK)
}

func DeleteTorrent(context echo.Context) error {

	reqTorrent := &models.TorrentCmd{}
	err := context.Bind(reqTorrent)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorInvalidJSONFormat, err.Error())
	}

	api := cordapi.NewCordAPI(config.Get().Tracker.Url)
	err = api.Login(config.Get().Tracker.User, config.Get().Tracker.Password)
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorLoginTracker, err.Error())
	}

	err = api.RemoveTorrent(reqTorrent)
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorDeleteTorrent, err.Error())
	}

	zap.S().Infow("Removed torrent", zap.String("info_hash", reqTorrent.InfoHash))

	return context.NoContent(http.StatusOK)
}
