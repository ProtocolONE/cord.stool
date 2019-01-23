package controllers

import (
    "cord.stool/service/api/parameters"
    "cord.stool/service/core/authentication"
    "cord.stool/service/models"
    "cord.stool/service/database"
    "cord.stool/service/config"

    "golang.org/x/crypto/bcrypt"
    "encoding/json"
    "net/http"
    //"go.uber.org/zap"
    "strings"
    "io/ioutil"
	"fmt"

    "github.com/pborman/uuid"
    jwt "github.com/dgrijalva/jwt-go"
    request "github.com/dgrijalva/jwt-go/request"

    "github.com/labstack/echo"
)

func CreateUser(context echo.Context) error {
    
    reqUser := new(models.Authorization)
    decoder := json.NewDecoder(context.Request().Body)
    decoder.Decode(&reqUser)

    manager := database.GeUserManager()
	users, err := manager.FindByName(reqUser.Username)
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Cannot read from database, error: %s", err.Error()))
    } 

    if len(users) != 0 {
        return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("User %s already exists", reqUser.Username))
    }

    storage, err := getUserStorageName(reqUser.Username)
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Cannot generate user files storage name, error: %s", err.Error()))
    }

    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(reqUser.Password), 10)
    
    err = manager.Insert(&models.User{reqUser.Username, string(hashedPassword), storage})
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Cannot add user %s, error: %s", reqUser.Username, err.Error()))
    }
    
    context.Echo().Logger.Info("Created new user %s.", reqUser.Username)
    return context.NoContent(http.StatusOK)
}

func DeleteUser(context echo.Context) error {

	if !authentication.RequireTokenAuthentication(context) {
		return nil
	}

    reqUser := new(models.Authorization)
    decoder := json.NewDecoder(context.Request().Body)
    decoder.Decode(&reqUser)
    
    manager := database.GeUserManager()
    err := manager.RemoveByName(reqUser.Username)
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Cannot delete user %s, error: %s", reqUser.Username, err.Error()))
    }
        
    context.Echo().Logger.Info("Removed user %s", reqUser.Username)
    return context.NoContent(http.StatusOK)
}

func Login(context echo.Context) error {

    reqUser := new(models.Authorization)
    decoder := json.NewDecoder(context.Request().Body)
    decoder.Decode(&reqUser)

    context.Echo().Logger.Info("Login: username: %s; password: %s", reqUser.Username, reqUser.Password)

    authBackend := authentication.InitJWTAuthenticationBackend()
    if !authBackend.Authenticate(reqUser) {
        return echo.NewHTTPError(http.StatusUnauthorized, "Invalid username or password")
    }

    uuid := uuid.New()
    token, err := authBackend.GenerateToken(reqUser.Username, uuid)
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Cannot generate token for user %s", reqUser.Username))
    } 
    
    context.Echo().Logger.Info("token: \"%s\"", token)

    return context.JSON(http.StatusOK, models.AuthToken{reqUser.Username, token})
}

func RefreshToken(context echo.Context) error {

	if !authentication.RequireTokenAuthentication(context) {
		return nil
	}

    reqUser := new(models.Authorization)
    decoder := json.NewDecoder(context.Request().Body)
    decoder.Decode(&reqUser)

    authBackend := authentication.InitJWTAuthenticationBackend()
    uuid := uuid.New()
    
    token, err := authBackend.GenerateToken(reqUser.Username, uuid)
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Cannot generate token for user %s, error: %s", reqUser.Username, err.Error()))
    }

    return context.JSON(http.StatusOK, parameters.TokenAuthentication{token})
}

func Logout(context echo.Context) error {
    
	if !authentication.RequireTokenAuthentication(context) {
		return nil
	}

    authBackend := authentication.InitJWTAuthenticationBackend()
    tokenRequest, err := request.ParseFromRequest(context.Request(), request.OAuth2Extractor, func(token *jwt.Token) (interface{}, error) {
        return authBackend.PublicKey, nil
    })

    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Logout failed, error: %s", err.Error()))
    }

    tokenString := context.Request().Header.Get("Authorization")

    err = authBackend.Logout(tokenString, tokenRequest)
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Logout failed, error: %s", err.Error()))
    }
    
    return context.NoContent(http.StatusOK)
}

func getUserStorageName(username string) (string, error) {

    storage := strings.Replace(username, "/\\:*?\"<>|", "_", -1)
    storage, err := ioutil.TempDir(config.Get().Service.StorageRootPath, storage)
    if (err != nil) {
        return "", err
    }

    return storage, nil
}
