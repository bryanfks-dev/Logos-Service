package config

import (
	"github.com/bryanfks-dev/selah/internal/delivery/http/middleware"
	"github.com/labstack/echo/v5"
	"github.com/spf13/viper"
)

func NewServer(config *viper.Viper) *echo.Echo {
	echo := echo.New()

	echo.Use(middleware.GzipMiddleware())
	echo.Use(middleware.CorsMiddleware())
	echo.Use(middleware.LogRequestMiddleware())
	echo.Use(middleware.RequestIDMiddleware())
	echo.Use(middleware.RecoveryMiddleware())

	return echo
}
