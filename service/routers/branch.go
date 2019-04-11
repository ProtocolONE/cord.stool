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

	e.GET("/api/v1/branch/list", controllers.ListBranchCmd, authentication.RequireTokenAuthentication)
	e.POST("/api/v1/branch/shallow", controllers.ShallowBranchCmd, authentication.RequireTokenAuthentication)
}