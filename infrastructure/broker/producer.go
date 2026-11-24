package broker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bryanfks-dev/selah/pkg/config"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type MessageWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type Producer interface {
	Publish(ctx context.Context, key []byte, value []byte) error
	PublishMessage(ctx context.Context, msg kafka.Message) error
	Close() error
}

type kafkaProducer struct {
	writer MessageWriter
	cfg    *config.KafkaConfig
	log    *zap.Logger
}

func NewKafkaProducer(cfg *config.KafkaConfig, log *zap.Logger) Producer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Broker),
		Topic:    cfg.Topic,
		Balancer: &kafka.LeastBytes{},
	}

	return NewKafkaProducerWithWriter(cfg, log, writer)
}

func NewKafkaProducerWithWriter(
	cfg *config.KafkaConfig,
	log *zap.Logger,
	writer MessageWriter,
) Producer {
	return &kafkaProducer{
		writer: writer,
		cfg:    cfg,
		log:    log,
	}
}

func (p *kafkaProducer) Publish(
	ctx context.Context,
	key []byte,
	value []byte,
) error {
	msg := kafka.Message{
		Key:   key,
		Value: value,
		Time:  time.Now(),
	}

	return p.PublishMessage(ctx, msg)
}

func (p *kafkaProducer) PublishMessage(
	ctx context.Context,
	msg kafka.Message,
) error {
	if p == nil || p.writer == nil {
		return errors.New("kafka producer writer is uninitialized")
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		p.log.Error(
			"failed to write message to kafka",
			zap.Error(err),
			zap.String("key", string(msg.Key)),
		)

		return fmt.Errorf("failed to write kafka message: %w", err)
	}

	p.log.Debug(
		"published message to kafka",
		zap.String("topic", p.cfg.Topic),
		zap.String("key", string(msg.Key)),
	)

	return nil
}

func (p *kafkaProducer) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}

	p.log.Info("closing kafka producer writer")
	return p.writer.Close()
}
