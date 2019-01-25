package routers

import (

    "github.com/labstack/echo"

	"cord.stool/service/controllers"
)

func InitCmdRoutes(e *echo.Echo) {

    e.POST("/api/v1/file/upload", controllers.UploadCmd)
    e.POST("/api/v1/file/cmp-hash", controllers.CompareHashCmd)
}
