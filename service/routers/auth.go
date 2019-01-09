package routers

import (
	"net/http"

	"github.com/gorilla/mux"

	"cord.stool/service/controllers"
    "cord.stool/service/core/authentication"
)

func SetAuthRoutes(router *mux.Router) *mux.Router {

	router.Handle(
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
		
	return router
}
