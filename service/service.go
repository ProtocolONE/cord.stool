package service

import (
	"fmt"

	"cord.stool/service/config"
	"cord.stool/service/database"
	"cord.stool/service/routers"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"go.uber.org/zap"
)

func Start(port uint) error {

	logger, _ := zap.NewProduction()
	defer logger.Sync()

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
