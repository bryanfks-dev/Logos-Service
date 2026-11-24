package main

import (
	"context"
	"errors"
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

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
	)
	defer stop()

	container.Log.Info(
		"starting selah kafka broker service",
		zap.String("broker", container.KafkaConfig.Broker),
		zap.String("topic", container.KafkaConfig.Topic),
		zap.String("group_id", container.KafkaConfig.GroupID),
	)

	defer func() {
		if err := container.BrokerConsumer.Close(); err != nil {
			container.Log.Warn(
				"failed to close broker consumer",
				zap.Error(err),
			)
		}

		if err := container.BrokerProducer.Close(); err != nil {
			container.Log.Warn(
				"failed to close broker producer",
				zap.Error(err),
			)
		}
	}()

	if err := container.BrokerConsumer.Start(ctx); err != nil &&
		!errors.Is(err, context.Canceled) {
		container.Log.Sugar().Fatalf("broker consumer error: %v", err)
	}

	container.Log.Info("selah kafka broker service stopped cleanly")
}
