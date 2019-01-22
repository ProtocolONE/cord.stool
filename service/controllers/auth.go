package controllers

import (
    "cord.stool/service/api/parameters"
    "cord.stool/service/core/authentication"
    "cord.stool/service/models"
    "cord.stool/service/database"
    "cord.stool/service/config"
    "cord.stool/service/core/utils"

    "golang.org/x/crypto/bcrypt"
    "gopkg.in/mgo.v2/bson"
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

    dbc := database.Get("users")
    usersWon, err := dbc.Find(bson.M{"username": reqUser.Username}).Count()
    if err != nil {
		//utils.ServiceError(context, http.StatusInternalServerError, "Cannot read from database", err)
		return fmt.Errorf("Cannot read from database, error: %s", err.Error())
    } 

    if usersWon != 0 {
		//utils.ServiceError(context, http.StatusInternalServerError, fmt.Sprintf("User %s already exists", reqUser.Username), nil)
		return fmt.Errorf("User %s already exists", reqUser.Username)
    } 

    storage, err := getUserStorageName(reqUser.Username)
    if err != nil {
    	//utils.ServiceError(context, http.StatusInternalServerError, "Cannot generate user files storage name", err)
		return fmt.Errorf("Cannot generate user files storage name, error: %s", err.Error())
    }

    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(reqUser.Password), 10)
    err = dbc.Insert(models.Authorization{reqUser.Username, string(hashedPassword), storage})
    if err != nil {
    	//utils.ServiceError(context, http.StatusInternalServerError, fmt.Sprintf("Cannot add user %s", reqUser.Username), err)
		return fmt.Errorf("Cannot add user %s, error: %s", reqUser.Username, err.Error())
    } 
    
    context.Echo().Logger.Info("Created new user %s.", reqUser.Username)
    if !login(context, reqUser) {
        return nil
    }

    return nil
}

func DeleteUser(context echo.Context) error {

	if !authentication.RequireTokenAuthentication(context) {
		return nil
	}

    reqUser := new(models.Authorization)
    decoder := json.NewDecoder(context.Request().Body)
    decoder.Decode(&reqUser)
    
    dbc := database.Get("users")
    err := dbc.Remove(bson.M{"username": reqUser.Username})
    if err != nil {
    	//utils.ServiceError(context, http.StatusInternalServerError, fmt.Sprintf("Cannot delete user %s", reqUser.Username), err)
		return fmt.Errorf("Cannot delete user %s, error: %s", reqUser.Username, err.Error())
    }
        
    context.Echo().Logger.Info("Removed user %s", reqUser.Username)
    return context.NoContent(http.StatusOK)
}

func Login(context echo.Context) error {

    reqUser := new(models.Authorization)
    decoder := json.NewDecoder(context.Request().Body)
    decoder.Decode(&reqUser)

    context.Echo().Logger.Info("Login: username: %s; password: %s", reqUser.Username, reqUser.Password)
    if !login(context, reqUser) {
        return nil
    }
    return nil
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
    	//utils.ServiceError(context, http.StatusInternalServerError, fmt.Sprintf("Cannot generate token for user %s", reqUser.Username), err)
		return fmt.Errorf("Cannot generate token for user %s, error: %s", reqUser.Username, err.Error())
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
    	//utils.ServiceError(context, http.StatusInternalServerError, "Logout failed", err)
		return fmt.Errorf("Logout failed, error: %s", err.Error())
    }

    tokenString := context.Request().Header.Get("Authorization")

    err = authBackend.Logout(tokenString, tokenRequest)
    if err != nil {
    	//utils.ServiceError(context, http.StatusInternalServerError, "Logout failed", err)
		return fmt.Errorf("Logout failed, error: %s", err.Error())
    }
    
    return context.NoContent(http.StatusOK)
}

func login(context echo.Context, reqUser *models.Authorization) bool {

    authBackend := authentication.InitJWTAuthenticationBackend()
    if !authBackend.Authenticate(reqUser) {
        utils.ServiceError(context, http.StatusUnauthorized, fmt.Sprintf("Invalid username %s or password", reqUser.Username), nil)
        return false
    }

    uuid := uuid.New()
    token, err := authBackend.GenerateToken(reqUser.Username, uuid)
    if err != nil {
        utils.ServiceError(context, http.StatusInternalServerError, fmt.Sprintf("Cannot generate token for user %s", reqUser.Username), err)
        return false
    } 
    
    context.Echo().Logger.Info("token: \"%s\"", token)

    context.JSON(http.StatusOK, models.AuthToken{reqUser.Username, token})
    return true
}

func getUserStorageName(username string) (string, error) {

    storage := strings.Replace(username, "/\\:*?\"<>|", "_", -1)
    storage, err := ioutil.TempDir(config.Get().Service.StorageRootPath, storage)
    if (err != nil) {
        return "", err
    }

    return storage, nil
}
