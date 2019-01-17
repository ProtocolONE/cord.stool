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
    "go.uber.org/zap"
    "strings"
    "io/ioutil"
	"fmt"

    "github.com/pborman/uuid"
    jwt "github.com/dgrijalva/jwt-go"
    request "github.com/dgrijalva/jwt-go/request"
)

func CreateUser(w http.ResponseWriter, r *http.Request) {
    
    w.Header().Set("Content-Type", "application/json")

    reqUser := new(models.Authorisation)
    decoder := json.NewDecoder(r.Body)
    decoder.Decode(&reqUser)

    dbc := database.Get("users")
    usersWon, err := dbc.Find(bson.M{"username": reqUser.Username}).Count()
    if err != nil {
		utils.ServiceError(w, http.StatusInternalServerError, "Cannot read from database", err)
        return
    } 

    if usersWon != 0 {
		utils.ServiceError(w, http.StatusInternalServerError, fmt.Sprintf("User %s already exists", reqUser.Username), nil)
        return
    } 

    storage, err := getUserStorageName(reqUser.Username)
    if err != nil {
    	utils.ServiceError(w, http.StatusInternalServerError, "Cannot generate user files storage name.", err)
        return
    }

    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(reqUser.Password), 10)
    err = dbc.Insert(models.Authorisation{reqUser.Username, string(hashedPassword), storage})
    if err != nil {
    	utils.ServiceError(w, http.StatusInternalServerError, fmt.Sprintf("Cannot add user %s", reqUser.Username), err)
        return
    } 
    
    zap.S().Infof("Created new user %s.", reqUser.Username)
    if !login(w, reqUser) {
        return
    }
    w.WriteHeader(http.StatusOK)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

    reqUser := new(models.Authorisation)
    decoder := json.NewDecoder(r.Body)
    decoder.Decode(&reqUser)
    
    dbc := database.Get("users")
    err := dbc.Remove(bson.M{"username": reqUser.Username})
    if err != nil {
    	utils.ServiceError(w, http.StatusInternalServerError, fmt.Sprintf("Cannot delete user %s", reqUser.Username), err)
        return
    }
        
    zap.S().Infof("Removed user %s", reqUser.Username)
    w.WriteHeader(http.StatusOK)
}

func Login(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

    reqUser := new(models.Authorisation)
    decoder := json.NewDecoder(r.Body)
    decoder.Decode(&reqUser)

    zap.S().Infof("Login: username: %s; password: %s", reqUser.Username, reqUser.Password)
    if !login(w, reqUser) {
        return
    }
    w.WriteHeader(http.StatusOK)
}

func RefreshToken(w http.ResponseWriter, r *http.Request) {

    w.Header().Set("Content-Type", "application/json")

    requestUser := new(models.Authorisation)
    decoder := json.NewDecoder(r.Body)
    decoder.Decode(&requestUser)

    authBackend := authentication.InitJWTAuthenticationBackend()
    uuid := uuid.New()
    
    token, err := authBackend.GenerateToken(requestUser.Username, uuid)
    if err != nil {
    	utils.ServiceError(w, http.StatusInternalServerError, fmt.Sprintf("Cannot generate token for user %s", requestUser.Username), err)
        return
    }

    response, _ := json.Marshal(parameters.TokenAuthentication{token})
    w.Write(response)
    w.WriteHeader(http.StatusOK)
}

func Logout(w http.ResponseWriter, r *http.Request) {
    
    w.Header().Set("Content-Type", "application/json")

    authBackend := authentication.InitJWTAuthenticationBackend()
    tokenRequest, err := request.ParseFromRequest(r, request.OAuth2Extractor, func(token *jwt.Token) (interface{}, error) {
        return authBackend.PublicKey, nil
    })

    if err != nil {
    	utils.ServiceError(w, http.StatusInternalServerError, "Logout failed", err)
        return
    }

    tokenString := r.Header.Get("Authorization")

    err = authBackend.Logout(tokenString, tokenRequest)
    if err != nil {
    	utils.ServiceError(w, http.StatusInternalServerError, "Logout failed", err)
        return
    }
    
    w.WriteHeader(http.StatusOK)
}

//
func login(w http.ResponseWriter, reqUser *models.Authorisation) bool {

    authBackend := authentication.InitJWTAuthenticationBackend()
    if !authBackend.Authenticate(reqUser) {
        utils.ServiceError(w, http.StatusInternalServerError, fmt.Sprintf("Invalid username %s or password", reqUser.Username), nil)
        return false
    }

    uuid := uuid.New()
    token, err := authBackend.GenerateToken(reqUser.Username, uuid)
    if err != nil {
        utils.ServiceError(w, http.StatusInternalServerError, fmt.Sprintf("Cannot generate token for user %s", reqUser.Username), err)
        return false
    } 
    
    zap.S().Infof("token: \"%s\"", token)
    response, _ := json.Marshal(models.AuthToken{reqUser.Username, token})
    w.Write(response)

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
