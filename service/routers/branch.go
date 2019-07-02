package routers

import (
	"github.com/labstack/echo"

	"cord.stool/service/controllers"
	"cord.stool/service/core/authentication"
)

func InitBranchCmdRoutes(e *echo.Echo) {

	e.POST("/api/v1/branch", controllers.CreateBranchCmd, authentication.RequireTokenAuthentication)
	e.DELETE("/api/v1/branch", controllers.DeleteBranchCmd, authentication.RequireTokenAuthentication)

	e.GET("/api/v1/branch", controllers.GetBranchCmd, authentication.RequireTokenAuthentication)
	e.PUT("/api/v1/branch", controllers.UpdateBranchCmd, authentication.RequireTokenAuthentication)

	e.PUT("/api/v1/branch/live", controllers.SetLiveBranchCmd, authentication.RequireTokenAuthentication)
	e.GET("/api/v1/branch/live", controllers.GetLiveBranchCmd, authentication.RequireTokenAuthentication)

	e.GET("/api/v1/branch/list", controllers.ListBranchCmd, authentication.RequireTokenAuthentication)
	e.POST("/api/v1/branch/shallow", controllers.ShallowBranchCmd, authentication.RequireTokenAuthentication)

	e.POST("/api/v1/branch/build", controllers.CreateBuildCmd, authentication.RequireTokenAuthentication)
	e.DELETE("/api/v1/branch/build", controllers.DeleteBuildCmd, authentication.RequireTokenAuthentication)
	e.GET("/api/v1/branch/build", controllers.GetBuildCmd, authentication.RequireTokenAuthentication)
	e.GET("/api/v1/branch/build/list", controllers.ListBuildCmd, authentication.RequireTokenAuthentication)
	e.GET("/api/v1/branch/build/live", controllers.GetLiveBuildCmd, authentication.RequireTokenAuthentication)
	e.PUT("/api/v1/branch/build/publish", controllers.PublishBuildCmd, authentication.RequireTokenAuthentication)
}
