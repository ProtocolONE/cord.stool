package service

import (
	"fmt"

    "cord.stool/service/routers"
    "cord.stool/service/config"
    "cord.stool/service/database"

    "github.com/labstack/echo"
    "github.com/labstack/echo/middleware"
)

func Start(port uint) error {

    conf, err := config.Init()
    if err != nil {
        return err
    }

    err = database.Init()
    if err != nil {
        return err
    }

    e := echo.New()

    e.Logger.Info("Create service. Scheme: \"%s\", port: \"%d\"", conf.Service.HttpScheme, conf.Service.ServicePort)

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
    routers.InitRoutes(e)

	// Start server
    e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", conf.Service.ServicePort)))
    
    return nil
}
