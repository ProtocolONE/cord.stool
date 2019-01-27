package routers

import (
	"github.com/labstack/echo"

	"cord.stool/service/controllers"
)

func InitAuthRoutes(e *echo.Echo) {

	e.POST("/api/v1/auth/user", controllers.CreateUser)
	e.DELETE("/api/v1/auth/user", controllers.DeleteUser)
	e.POST("/api/v1/auth/token", controllers.Login)
	e.GET("/api/v1/auth/refresh-token", controllers.RefreshToken)
	e.GET("/api/v1/auth/logout", controllers.Logout)
}
