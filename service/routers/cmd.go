package routers

import (

    "github.com/labstack/echo"

	"cord.stool/service/controllers"
)

func InitCmdRoutes(e *echo.Echo) {

    e.POST("/api/v1/cmd/upload", controllers.UploadCmd)
    e.POST("/api/v1/cmd/cmp-hash", controllers.CompareHashCmd)
}
