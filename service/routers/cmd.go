package routers

import (
	"github.com/labstack/echo"

	"cord.stool/service/controllers"
	"cord.stool/service/core/authentication"
)

func InitCmdRoutes(e *echo.Echo) {

	e.POST("/api/v1/file/upload", controllers.UploadCmd, authentication.RequireTokenAuthentication)
	e.POST("/api/v1/file/cmp-hash", controllers.CompareHashCmd, authentication.RequireTokenAuthentication)
}
