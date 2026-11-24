package broker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/bryanfks-dev/selah/pkg/config"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type MessageHandler func(ctx context.Context, msg kafka.Message) error

type MessageFetcher interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type Consumer interface {
	Start(ctx context.Context) error
	SetHandler(handler MessageHandler)
	Close() error
}

type kafkaConsumer struct {
	fetcher MessageFetcher
	cfg     *config.KafkaConfig
	log     *zap.Logger

	mu      sync.RWMutex
	handler MessageHandler
}

func NewKafkaConsumer(cfg *config.KafkaConfig, log *zap.Logger) Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{cfg.Broker},
		GroupID:  cfg.GroupID,
		Topic:    cfg.Topic,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	return NewKafkaConsumerWithFetcher(cfg, log, reader)
}

func NewKafkaConsumerWithFetcher(
	cfg *config.KafkaConfig,
	log *zap.Logger,
	fetcher MessageFetcher,
) Consumer {
	return &kafkaConsumer{
		fetcher: fetcher,
		cfg:     cfg,
		log:     log,
	}
}

func (c *kafkaConsumer) SetHandler(handler MessageHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.handler = handler
}

func (c *kafkaConsumer) Start(ctx context.Context) error {
	if c == nil || c.fetcher == nil {
		return errors.New("kafka consumer fetcher is uninitialized")
	}

	c.log.Info(
		"kafka consumer loop started",
		zap.String("broker", c.cfg.Broker),
		zap.String("topic", c.cfg.Topic),
		zap.String("group_id", c.cfg.GroupID),
	)

	for {
		select {
		case <-ctx.Done():
			c.log.Info("kafka consumer context canceled, shutting down")
			return nil
		default:
		}

		msg, err := c.fetcher.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) {
				c.log.Info("kafka consumer fetch stopped cleanly")
				return nil
			}

			c.log.Error("failed to fetch message from kafka", zap.Error(err))
			return fmt.Errorf("failed to fetch kafka message: %w", err)
		}

		if err := c.handleMessage(ctx, msg); err != nil {
			c.log.Error(
				"failed to handle kafka message",
				zap.Error(err),
				zap.String("key", string(msg.Key)),
			)
			continue
		}

		if err := c.fetcher.CommitMessages(ctx, msg); err != nil {
			c.log.Error(
				"failed to commit kafka message offset",
				zap.Error(err),
			)
		}
	}
}

func (c *kafkaConsumer) handleMessage(
	ctx context.Context,
	msg kafka.Message,
) error {
	c.mu.RLock()
	handler := c.handler
	c.mu.RUnlock()

	if handler != nil {
		return handler(ctx, msg)
	}

	c.log.Info(
		"received kafka message",
		zap.String("topic", msg.Topic),
		zap.Int("partition", msg.Partition),
		zap.Int64("offset", msg.Offset),
		zap.String("key", string(msg.Key)),
		zap.Int("value_bytes", len(msg.Value)),
	)

	return nil
}

func (c *kafkaConsumer) Close() error {
	if c == nil || c.fetcher == nil {
		return nil
	}

	c.log.Info("closing kafka consumer fetcher")
	return c.fetcher.Close()
}
