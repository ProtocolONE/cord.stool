package authentication

import (
    "fmt"
    jwt "github.com/dgrijalva/jwt-go"
    request "github.com/dgrijalva/jwt-go/request"
    "net/http"
    "go.uber.org/zap"
)


func RequireTokenAuthentication(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
    authBackend := InitJWTAuthenticationBackend()
    token, err := request.ParseFromRequest(req, request.OAuth2Extractor, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
            return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
        } else {
            return authBackend.PublicKey, nil
        }
    })
    if err == nil && token.Valid && !authBackend.IsInBlacklist(req.Header.Get("Authorization")) {
        zap.S().Infof("req token %s", req.Header.Get("Authorization"))
        next(rw, req)
    } else {
        zap.S().Info(err.Error())
        rw.WriteHeader(http.StatusUnauthorized)
    }
}
