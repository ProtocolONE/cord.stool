package routers

import (
    "github.com/labstack/echo"

    "cord.stool/service/controllers"
)

func InitAuthRoutes(e *echo.Echo) {

    e.POST("/api/v1/user", controllers.CreateUser)
    e.DELETE("/api/v1/user", controllers.DeleteUser)
    e.POST("/api/v1/token-auth", controllers.Login)
    e.GET("/api/v1/refresh-token-auth", controllers.RefreshToken)
    e.GET("/api/v1/logout", controllers.Logout)

    /*router.Handle(
        "/api/v1/user",
        http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            controllers.CreateUser(w, r)
	})).Methods("POST")
		
    router.Handle(
        "/api/v1/user",
        http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            authentication.RequireTokenAuthentication(w, r, controllers.DeleteUser)
	})).Methods("DELETE")
		
    router.Handle(
        "/api/v1/token-auth",
        http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            controllers.Login(w, r)
	})).Methods("POST")
		
    router.Handle(
        "/api/v1/refresh-token-auth",
        http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            authentication.RequireTokenAuthentication(w, r, controllers.RefreshToken)
	})).Methods("GET")
		
    router.Handle(
        "/api/v1/logout",
        http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            authentication.RequireTokenAuthentication(w, r, controllers.Logout)
	})).Methods("GET")
		
	return router*/
}
