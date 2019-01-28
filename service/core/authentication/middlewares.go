package authentication

import (
	"fmt"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	request "github.com/dgrijalva/jwt-go/request"

	"github.com/labstack/echo"
)

func RequireTokenAuthentication(next echo.HandlerFunc) echo.HandlerFunc {

	return func(context echo.Context) error {

		authBackend := InitJWTAuthenticationBackend()
		token, err := request.ParseFromRequest(context.Request(), request.OAuth2Extractor, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			} else {
				return authBackend.PublicKey, nil
			}
		})

		if err != nil || !token.Valid || authBackend.IsInBlacklist(context.Request().Header.Get("Authorization")) {

			context.Echo().Logger.Error(err.Error())
			return echo.NewHTTPError(http.StatusUnauthorized, "Authorization failed")
		}

		claims := token.Claims.(jwt.MapClaims)
		clientID, _ := claims["client_id"].(string)
		context.Request().Header.Set("ClientID", clientID)

		return next(context)
	}
}
