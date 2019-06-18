package routers

import (
	"github.com/labstack/echo"

	"cord.stool/service/controllers"
	"cord.stool/service/core/authentication"
)

func InitCmdRoutes(e *echo.Echo) {

	e.POST("/api/v1/file/upload", controllers.UploadCmd, authentication.RequireTokenAuthentication)
	e.POST("/api/v1/file/cmp-hash", controllers.CompareHashCmd, authentication.RequireTokenAuthentication)
	e.POST("/api/v1/tracker/torrent", controllers.AddTorrent, authentication.RequireTokenAuthentication)
	e.DELETE("/api/v1/tracker/torrent", controllers.DeleteTorrent, authentication.RequireTokenAuthentication)
	e.GET("/api/v1/file/signature", controllers.SignatureCmd, authentication.RequireTokenAuthentication)
	e.POST("/api/v1/file/patch", controllers.ApplyPatchCmd, authentication.RequireTokenAuthentication)

	e.GET("/api/v1/file/update", controllers.GetUpdateCmd, authentication.RequireTokenAuthentication)
	e.GET("/api/v1/file/update-info", controllers.GetUpdateInfoCmd, authentication.RequireTokenAuthentication)
	e.GET("/api/v1/file/update-patch", controllers.GetUpdatePatchCmd, authentication.RequireTokenAuthentication)
	e.GET("/api/v1/file/download", controllers.DownloadCmd, authentication.RequireTokenAuthentication)
}
