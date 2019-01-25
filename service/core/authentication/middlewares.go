package authentication

import (
	"fmt"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	request "github.com/dgrijalva/jwt-go/request"

	"github.com/labstack/echo"
)

func RequireTokenAuthentication(context echo.Context) bool {
	
	authBackend := InitJWTAuthenticationBackend()
	token, err := request.ParseFromRequest(context.Request(), request.OAuth2Extractor, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		} else {
			return authBackend.PublicKey, nil
		}
	})

	claims := token.Claims.(jwt.MapClaims)
	clientID, _ := claims["client_id"].(string)
	context.Request().Header.Set("ClientID", clientID)

	if err == nil && token.Valid && !authBackend.IsInBlacklist(context.Request().Header.Get("Authorization")) {
		return true
	} else {
		context.Echo().Logger.Error(err.Error())
    	context.NoContent(http.StatusUnauthorized)
	}

	return false
}
