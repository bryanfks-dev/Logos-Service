package main

import (
	"fmt"

	"github.com/bryanfks-dev/logos/di"
	_ "github.com/bryanfks-dev/logos/docs"
)

// @title Logos Service API
// @version 1.0
// @description API documentation for Logos Service
func main() {
	container, err := di.NewContainer()
	if err != nil {
		message := fmt.Sprintf("setup container error: %v\n", err)
		panic(message)
	}

	container.RouteConfig.Setup()

	port := container.Config.GetInt("web.port")
	if err := container.App.Start(fmt.Sprintf(":%d", port)); err != nil {
		container.Log.Sugar().Fatalf("start server error: %v", err)
	}
}
