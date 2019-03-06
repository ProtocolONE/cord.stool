package service

import (
	"fmt"

	"cord.stool/compressor/gzip"
	"cord.stool/service/config"
	"cord.stool/service/database"
	"cord.stool/service/routers"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"go.uber.org/zap"
)

func Start(port uint) error {

	logger, _ := zap.NewProduction()
	zap.ReplaceGlobals(logger)
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

	zap.S().Infow("Create service", zap.String("Scheme", conf.Service.HttpScheme), zap.Int("port", conf.Service.ServicePort))

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	routers.InitRoutes(e)

	gzip.Init()

	// Start server
	zap.S().Fatal(e.Start(fmt.Sprintf(":%d", conf.Service.ServicePort)))

	return nil
}
