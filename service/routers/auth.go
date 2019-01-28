package routers

import (
	"github.com/labstack/echo"

	"cord.stool/service/controllers"
	"cord.stool/service/core/authentication"
)

func InitAuthRoutes(e *echo.Echo) {

	e.POST("/api/v1/auth/user", controllers.CreateUser)
	e.DELETE("/api/v1/auth/user", controllers.DeleteUser, authentication.RequireTokenAuthentication)
	e.POST("/api/v1/auth/token", controllers.Login)
	e.GET("/api/v1/auth/refresh-token", controllers.RefreshToken, authentication.RequireTokenAuthentication)
	e.GET("/api/v1/auth/logout", controllers.Logout, authentication.RequireTokenAuthentication)
}
