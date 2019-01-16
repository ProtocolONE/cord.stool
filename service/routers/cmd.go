package routers

import (
	"net/http"

	"github.com/gorilla/mux"

	"cord.stool/service/controllers"
    "cord.stool/service/core/authentication"
)

func SetCmdRoutes(router *mux.Router) *mux.Router {

    router.Handle(
        "/api/v1/cmd/upload", 
        http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            authentication.RequireTokenAuthentication(w, r, controllers.UploadCmd)
    })).Methods("POST")

	return router
}
