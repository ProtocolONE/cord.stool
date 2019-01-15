package authentication

import (
    "cord.stool/service/models"
    "cord.stool/service/config"
    "cord.stool/service/database"
    "gopkg.in/mgo.v2/bson"
    "bufio"
    "crypto/rsa"
    "crypto/x509"
    "encoding/pem"
    jwt "github.com/dgrijalva/jwt-go"
    "golang.org/x/crypto/bcrypt"
    "os"
    "time"
    "fmt"
    "go.uber.org/zap"
)


type JWTAuthenticationBackend struct {
    privateKey *rsa.PrivateKey
    PublicKey  *rsa.PublicKey
}


const (
    tokenDuration = 72
    expireOffset  = 3600
)


var authBackendInstance *JWTAuthenticationBackend = nil


func InitJWTAuthenticationBackend() *JWTAuthenticationBackend {
    if authBackendInstance == nil {
        authBackendInstance = &JWTAuthenticationBackend{
            privateKey: getPrivateKey(),
            PublicKey:  getPublicKey(),
        }
    }
    return authBackendInstance
}


func (backend *JWTAuthenticationBackend) GenerateToken(clientID string , userUUID string) (string, error) {
    token := jwt.New(jwt.SigningMethodRS512)
    cfg := config.Get().Service
    token.Claims = jwt.MapClaims{
        "exp": time.Now().Add(time.Hour * time.Duration(cfg.JwtExpDelta)).Unix(),
        "iat": time.Now().Unix(),
        "sub": userUUID,
        "client_id": clientID,
    }
    tokenString, err := token.SignedString(backend.privateKey)
    if err != nil {
        zap.S().Errorf("Can`t generate token, err: %v", err)
        return "", err
    }
    return tokenString, nil
}


func (backend *JWTAuthenticationBackend) Authenticate(user *models.Authorisation) bool {
    dbc := database.Get("users")
    var dbUsers []models.Authorisation
    err := dbc.Find(bson.M{"username": user.Username}).All(&dbUsers)
    if err != nil {
        zap.S().Infof("Can`t find user %s", user.Username)
        return false
    }
    return len(dbUsers) == 1 && user.Username == dbUsers[0].Username && bcrypt.CompareHashAndPassword([]byte(dbUsers[0].Password), []byte(user.Password)) == nil
}


func (backend *JWTAuthenticationBackend) getTokenRemainingValidity(timestamp interface{}) int {
    if validity, ok := timestamp.(float64); ok {
        tm := time.Unix(int64(validity), 0)
        remainer := tm.Sub(time.Now())
        if remainer > 0 {
            return int(remainer.Seconds() + expireOffset)
        }
    }
    return expireOffset
}


func (backend *JWTAuthenticationBackend) Logout(tokenStr string, token *jwt.Token) error {
    zap.S().Infof("Logout token: \"%s\"", tokenStr)
    return nil
}


func (backend *JWTAuthenticationBackend) IsInBlacklist(tokenStr string) bool {
    zap.S().Infof("is blist token: \"%s\"", tokenStr)
    return false
}


func getPrivateKey() *rsa.PrivateKey {
    cfg := config.Get().Service
    privateKeyFile, err := os.Open(cfg.PrivateKeyPath)
    if err != nil {
        panic(fmt.Sprintf("Can`t open file \"%s\"", cfg.PrivateKeyPath))
    }
    pemfileinfo, _ := privateKeyFile.Stat()
    var size int64 = pemfileinfo.Size()
    pembytes := make([]byte, size)

    buffer := bufio.NewReader(privateKeyFile)
    _, err = buffer.Read(pembytes)

    data, _ := pem.Decode([]byte(pembytes))

    privateKeyFile.Close()

    privateKeyImported, err := x509.ParsePKCS1PrivateKey(data.Bytes)

    if err != nil {
        panic(err)
    }

    return privateKeyImported
}


func getPublicKey() *rsa.PublicKey {
    cfg := config.Get().Service
    publicKeyFile, err := os.Open(cfg.PublicKeyPath)
    if err != nil {
        panic(err)
    }

    pemfileinfo, _ := publicKeyFile.Stat()
    var size int64 = pemfileinfo.Size()
    pembytes := make([]byte, size)

    buffer := bufio.NewReader(publicKeyFile)
    _, err = buffer.Read(pembytes)

    data, _ := pem.Decode([]byte(pembytes))

    publicKeyFile.Close()

    publicKeyImported, err := x509.ParsePKIXPublicKey(data.Bytes)

    if err != nil {
        panic(err)
    }

    rsaPub, ok := publicKeyImported.(*rsa.PublicKey)

    if !ok {
        panic(err)
    }

    return rsaPub
}
