package routers

import (
	"net/http"

	"github.com/gorilla/mux"

	"cord.stool/service/controllers"
    "cord.stool/service/core/authentication"
)

func SetCmdRoutes(router *mux.Router) *mux.Router {

    router.Handle(
        "/api/v1/cmd/create", 
        http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            authentication.RequireTokenAuthentication(w, r, controllers.CreateCmd)
    })).Methods("POST")
		
    router.Handle(
        "/api/v1/cmd/push", 
        http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            authentication.RequireTokenAuthentication(w, r, controllers.PushCmd)
	})).Methods("POST")
		
    router.Handle(
        "/api/v1/cmd/diff", 
        http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            authentication.RequireTokenAuthentication(w, r, controllers.DiffCmd)
	})).Methods("POST")

    router.Handle(
        "/api/v1/cmd/torrent", 
        http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            authentication.RequireTokenAuthentication(w, r, controllers.TorrentCmd)
	})).Methods("POST")

	return router
}
