package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bryanfks-dev/selah/di"
	"go.uber.org/zap"
)

func main() {
	container, err := di.NewContainer()
	if err != nil {
		message := fmt.Sprintf("setup container error: %v\n", err)
		panic(message)
	}

	_, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
	)
	defer stop()

	if !container.Config.GetBool("kafka.enabled") {
		container.Log.Info("kafka broker service is disabled")
		return
	}

	container.Log.Info(
		"starting kafka broker service",
		zap.String("server", container.Config.GetString("kafka.server")),
		zap.String("group_id", container.Config.GetString("kafka.group.id")),
	)

	container.Log.Info("broker service stopped cleanly")
}
