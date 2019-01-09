package routers

import (
	"github.com/gorilla/mux"
)

func InitRoutes() *mux.Router {
	router := mux.NewRouter()
	router = SetAuthRoutes(router)
	router = SetCmdRoutes(router)
	return router
}
