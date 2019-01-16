package controllers

import (
    "cord.stool/service/api/parameters"
    "cord.stool/service/core/authentication"
    "cord.stool/service/models"
    "cord.stool/service/database"
    "cord.stool/service/config"

    "golang.org/x/crypto/bcrypt"
    "gopkg.in/mgo.v2/bson"
    "encoding/json"
    "net/http"
    "go.uber.org/zap"
    "strings"
    "io/ioutil"

    "github.com/pborman/uuid"
    jwt "github.com/dgrijalva/jwt-go"
    request "github.com/dgrijalva/jwt-go/request"
)

func CreateUser(w http.ResponseWriter, r *http.Request) {
    
    reqUser := new(models.Authorisation)
    decoder := json.NewDecoder(r.Body)
    decoder.Decode(&reqUser)
    dbc := database.Get("users")
    usersWon, err := dbc.Find(bson.M{"username": reqUser.Username}).Count()
    w.Header().Set("Content-Type", "application/json")
    if err != nil {
        zap.S().Errorf("Cannot find user \"%s\", err: %v", reqUser.Username, err)
        w.WriteHeader(http.StatusInternalServerError)
        response, _ := json.Marshal(models.Error{"Can`t add user."})
        w.Write(response)
        return
    } 

    if usersWon != 0 {
        zap.S().Errorf("Can`t add user \"%s\", is exists", reqUser.Username)
        w.WriteHeader(http.StatusUnauthorized)
        response, _ := json.Marshal(models.Error{"User is exists."})
        w.Write(response)
        return
    } 

    storage, err := getUserStorageName(reqUser.Username)
    if err != nil {
        zap.S().Errorf("Cannot generate user storage name, err: %v", err)
        w.WriteHeader(http.StatusInternalServerError)
        response, _ := json.Marshal(models.Error{"Cannot generate user storage name."})
        w.Write(response)
        return
    }

    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(reqUser.Password), 10)
    err = dbc.Insert(models.Authorisation{reqUser.Username, string(hashedPassword), storage})
    if err != nil {
        zap.S().Errorf("Cannot add user \"%s\" in db, err: %v", reqUser.Username, err)
        w.WriteHeader(http.StatusInternalServerError)
        response, _ := json.Marshal(models.Error{"Cannot add user in db."})
        w.Write(response)
        return
    } 
    
    zap.S().Errorf("Create new user \"%s\" in db, err: %v", reqUser.Username, err)
    responseStatus, token := login(reqUser)
    w.WriteHeader(responseStatus)
    w.Write(token)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
    reqUser := new(models.Authorisation)
    decoder := json.NewDecoder(r.Body)
    decoder.Decode(&reqUser)
    dbc := database.Get("users")
    err := dbc.Remove(bson.M{"username": reqUser.Username})
    if err != nil {
        zap.S().Errorf("Remove fail %v", err)
        w.WriteHeader(http.StatusInternalServerError)
        response, _ := json.Marshal(models.Error{"Cannot add user in db."})
        w.Write(response)
    } else {
        zap.S().Infof("Remove user \"%s\" complete", reqUser.Username)
        w.WriteHeader(http.StatusOK)
    }
}

func Login(w http.ResponseWriter, r *http.Request) {
    reqUser := new(models.Authorisation)
    decoder := json.NewDecoder(r.Body)
    decoder.Decode(&reqUser)
    zap.S().Infof("inpt > User: %s; Pswd: %s", reqUser.Username, reqUser.Password)
    responseStatus, token := login(reqUser)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(responseStatus)
    w.Write(token)
}

func RefreshToken(w http.ResponseWriter, r *http.Request) {
    requestUser := new(models.Authorisation)
    decoder := json.NewDecoder(r.Body)
    decoder.Decode(&requestUser)
    w.Header().Set("Content-Type", "application/json")
    w.Write(refreshToken(requestUser))
}

func Logout(w http.ResponseWriter, r *http.Request) {
    err := logout(r)
    w.Header().Set("Content-Type", "application/json")

    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
    } else {
        w.WriteHeader(http.StatusOK)
    }
}

//

func getUserStorageName(username string) (string, error) {

    storage := strings.Replace(username, "/\\:*?\"<>|", "_", -1)
    storage, err := ioutil.TempDir(config.Get().Service.StorageRootPath, storage)
    if (err != nil) {
        return "", err
    }

    return storage, nil
}

func login(reqUser *models.Authorisation) (int, []byte) {
    authBackend := authentication.InitJWTAuthenticationBackend()
    if authBackend.Authenticate(reqUser) {
        uuid := uuid.New()
        token, err := authBackend.GenerateToken(reqUser.Username, uuid)
        if err != nil {
            zap.S().Infof("Cannot generate token; err %v", err)
            return http.StatusInternalServerError, []byte("")
        } else {
            zap.S().Infof("token: \"%s\"", token)
            response, _ := json.Marshal(models.AuthToken{reqUser.Username, token})
            return http.StatusOK, response
        }
    }
    return http.StatusUnauthorized, []byte("")
}

func refreshToken(requestUser *models.Authorisation) []byte {
    authBackend := authentication.InitJWTAuthenticationBackend()
    uuid := uuid.New()
    token, err := authBackend.GenerateToken(requestUser.Username, uuid)
    if err != nil {
        panic(err)
    }
    response, err := json.Marshal(parameters.TokenAuthentication{token})
    if err != nil {
        panic(err)
    }
    return response
}

func logout(req *http.Request) error {
    authBackend := authentication.InitJWTAuthenticationBackend()
    tokenRequest, err := request.ParseFromRequest(req, request.OAuth2Extractor, func(token *jwt.Token) (interface{}, error) {
        return authBackend.PublicKey, nil
    })
    if err != nil {
        return err
    }
    tokenString := req.Header.Get("Authorization")
    return authBackend.Logout(tokenString, tokenRequest)
}
