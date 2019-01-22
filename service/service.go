package service

import (
	"fmt"
	"net/http"
    "encoding/binary"

    "go.uber.org/zap"
    "github.com/gorilla/handlers"

    "cord.stool/service/routers"
    "cord.stool/service/config"
    "cord.stool/service/database"
)

type LogWriter struct {}
func (w *LogWriter) Write(p []byte) (n int, err error) {
    zap.S().Info(fmt.Sprintf("Handle: %s", p))
    return binary.Size(p), nil
}

var NotImplemented = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
    w.Write([]byte("Not Implemented"))
  })

func Start(port uint) error {

    fmt.Println("Service starting...")

    logger, err := zap.NewDevelopment()
    if err != nil {
        return err
    }

	zap.ReplaceGlobals(logger)
    lw := &LogWriter{}
	
    conf, err := config.Init()
    if err != nil {
        return err
    }

    err = database.Init()
    if err != nil {
        return err
    }

    router := routers.InitRoutes()
    zap.S().Infof("Create service. Scheme: \"%s\", port: \"%d\"", conf.Service.HttpScheme, conf.Service.ServicePort)
	err = http.ListenAndServe(fmt.Sprintf(":%d", conf.Service.ServicePort), handlers.LoggingHandler(lw, router))
    if err != nil {
        return err
    }

	return nil
}
