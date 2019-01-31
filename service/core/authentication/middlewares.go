package authentication

import (
	"fmt"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	request "github.com/dgrijalva/jwt-go/request"

	"github.com/labstack/echo"
	"go.uber.org/zap"

	"cord.stool/service/models"
)

func RequireTokenAuthentication(next echo.HandlerFunc) echo.HandlerFunc {

	return requireTokenAuthentication(next, false)
}

func RequireRefreshTokenAuthentication(next echo.HandlerFunc) echo.HandlerFunc {

	return requireTokenAuthentication(next, true)
}

func requireTokenAuthentication(next echo.HandlerFunc, refreshToken bool) echo.HandlerFunc {

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

			if err != nil {
				zap.S().Error(err.Error())
				return context.JSON(http.StatusUnauthorized, models.Error{models.ErrorUnauthorized, err.Error()})

			} else {
				zap.S().Error("Authorization failed")
				return context.JSON(http.StatusUnauthorized, models.Error{models.ErrorUnauthorized, "Authorization failed"})
			}
		}

		rem := authBackend.GetTokenRemainingValidity(token)
		if rem <= 0 {

			zap.S().Error("Token is expired")
			return context.JSON(http.StatusUnauthorized, models.Error{models.ErrorTokenExpired, "Token is expired"})
		}

		claims := token.Claims.(jwt.MapClaims)

		if refreshToken {

			refresh, ok := claims["refresh"].(bool)
			if !ok || !refresh {

				zap.S().Error("Invalid refresh token")
				return context.JSON(http.StatusBadRequest, models.Error{models.ErrorInvalidToken, "Invalid refresh token"})
			}

		} else {

			access, ok := claims["access"].(bool)
			if !ok || !access {

				zap.S().Error("Invalid access token")
				return context.JSON(http.StatusBadRequest, models.Error{models.ErrorInvalidToken, "Invalid access token"})
			}
		}

		clientID, ok := claims["client_id"].(string)
		if !ok || clientID == "" {

			zap.S().Error("Invalid token")
			return context.JSON(http.StatusBadRequest, models.Error{models.ErrorInvalidToken, "Invalid token"})
		}

		context.Request().Header.Set("ClientID", clientID)
		return next(context)
	}
}
