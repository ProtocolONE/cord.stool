package controllers

import (
	"cord.stool/service/config"
	"cord.stool/service/core/authentication"
	"cord.stool/service/core/utils"
	"cord.stool/service/database"
	"cord.stool/service/models"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	request "github.com/dgrijalva/jwt-go/request"
	"github.com/pborman/uuid"

	"github.com/labstack/echo"
)

func CreateUser(context echo.Context) error {

	reqUser := &models.Authorization{}
	err := context.Bind(reqUser)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorInvalidJSONFormat, err.Error())
	}

	manager := database.NewUserManager()
	users, err := manager.FindByName(reqUser.Username)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	if len(users) != 0 {
		return utils.BuildBadRequestError(context, models.ErrorAlreadyExists, reqUser.Username)
	}

	storage, err := getUserStorageName(reqUser.Username)
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorGenUserStorageName, err.Error())
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(reqUser.Password), 10)

	err = manager.Insert(&models.User{reqUser.Username, string(hashedPassword), storage})
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	zap.S().Infow("Created new user", zap.String("username", reqUser.Username))

	return context.NoContent(http.StatusCreated)
}

func DeleteUser(context echo.Context) error {

	reqUser := &models.Authorization{}
	err := context.Bind(reqUser)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorInvalidJSONFormat, err.Error())
	}

	manager := database.NewUserManager()
	err = manager.RemoveByName(reqUser.Username)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorDatabaseFailure, err.Error())
	}

	zap.S().Infow("Removed user", zap.String("username", reqUser.Username))
	return context.NoContent(http.StatusOK)
}

func Login(context echo.Context) error {

	reqUser := &models.Authorization{}
	err := context.Bind(reqUser)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorInvalidJSONFormat, err.Error())
	}

	zap.S().Infow("Login", zap.String("username", reqUser.Username), zap.String("password", reqUser.Password))

	authBackend := authentication.InitJWTAuthenticationBackend()
	if !authBackend.Authenticate(reqUser) {
		return utils.BuildUnauthorizedError(context, models.ErrorInvalidUsernameOrPassword, "")
	}

	userUUID := uuid.New()
	token, err := authBackend.GenerateToken(reqUser.Username, userUUID, false)
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorGenToken, err.Error())
	}

	userUUID = uuid.New()
	refreshToken, err := authBackend.GenerateToken(reqUser.Username, userUUID, true)
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorGenToken, err.Error())
	}

	zap.S().Infow("Login", zap.String("token", token))
	return context.JSON(http.StatusOK, models.AuthToken{reqUser.Username, token, refreshToken})
}

func RefreshToken(context echo.Context) error {

	username := context.Request().Header.Get("ClientID")
	if username == "" {
		return context.NoContent(http.StatusBadRequest)
	}

	authBackend := authentication.InitJWTAuthenticationBackend()

	userUUID := uuid.New()
	token, err := authBackend.GenerateToken(username, userUUID, false)
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorGenToken, err.Error())
	}

	userUUID = uuid.New()
	refreshToken, err := authBackend.GenerateToken(username, userUUID, true)
	if err != nil {
		return utils.BuildInternalServerError(context, models.ErrorGenToken, err.Error())
	}

	return context.JSON(http.StatusOK, models.AuthRefresh{token, refreshToken})
}

func Logout(context echo.Context) error {

	authBackend := authentication.InitJWTAuthenticationBackend()
	tokenRequest, err := request.ParseFromRequest(context.Request(), request.OAuth2Extractor, func(token *jwt.Token) (interface{}, error) {
		return authBackend.PublicKey, nil
	})

	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorLogout, err.Error())
	}

	tokenString := context.Request().Header.Get("Authorization")

	err = authBackend.Logout(tokenString, tokenRequest)
	if err != nil {
		return utils.BuildBadRequestError(context, models.ErrorLogout, err.Error())
	}

	return context.NoContent(http.StatusOK)
}

func getUserStorageName(username string) (string, error) {

	storage := strings.Replace(username, "/\\:*?\"<>|", "_", -1)
	storage, err := ioutil.TempDir(config.Get().Service.StorageRootPath, storage)
	if err != nil {
		return "", err
	}

	return storage, nil
}
