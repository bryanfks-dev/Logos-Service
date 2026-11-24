package broker_test

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/bryanfks-dev/selah/infrastructure/broker"
	"github.com/bryanfks-dev/selah/pkg/config"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockFetcher struct {
	fetchFunc  func(ctx context.Context) (kafka.Message, error)
	commitFunc func(ctx context.Context, msgs ...kafka.Message) error
	closeFunc  func() error
}

func (m *mockFetcher) FetchMessage(ctx context.Context) (kafka.Message, error) {
	if m.fetchFunc != nil {
		return m.fetchFunc(ctx)
	}

	return kafka.Message{}, io.EOF
}

func (m *mockFetcher) CommitMessages(
	ctx context.Context,
	msgs ...kafka.Message,
) error {
	if m.commitFunc != nil {
		return m.commitFunc(ctx, msgs...)
	}

	return nil
}

func (m *mockFetcher) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}

	return nil
}

func testKafkaConfig() *config.KafkaConfig {
	return &config.KafkaConfig{
		Broker:  "localhost:9092",
		GroupID: "test-group",
		Topic:   "test-topic",
	}
}

func TestNewKafkaConsumer(t *testing.T) {
	cfg := testKafkaConfig()
	log := zap.NewNop()

	consumer := broker.NewKafkaConsumer(cfg, log)
	require.NotNil(t, consumer)

	err := consumer.Close()
	assert.NoError(t, err)
}

func TestKafkaConsumer_Start_UninitializedFetcher(t *testing.T) {
	cfg := testKafkaConfig()
	log := zap.NewNop()

	consumer := broker.NewKafkaConsumerWithFetcher(cfg, log, nil)
	err := consumer.Start(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "uninitialized")
}

func TestKafkaConsumer_Start_ContextCanceledImmediately(t *testing.T) {
	cfg := testKafkaConfig()
	log := zap.NewNop()

	fetcher := &mockFetcher{}
	consumer := broker.NewKafkaConsumerWithFetcher(cfg, log, fetcher)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := consumer.Start(ctx)
	assert.NoError(t, err)
}

func TestKafkaConsumer_Start_WithCustomHandlerAndEOF(t *testing.T) {
	cfg := testKafkaConfig()
	log := zap.NewNop()

	var handledCount int
	var committedCount int

	messages := []kafka.Message{
		{Topic: "test-topic", Key: []byte("k1"), Value: []byte("v1")},
		{Topic: "test-topic", Key: []byte("k2"), Value: []byte("v2")},
	}
	index := 0

	fetcher := &mockFetcher{
		fetchFunc: func(ctx context.Context) (kafka.Message, error) {
			if index < len(messages) {
				msg := messages[index]
				index++
				return msg, nil
			}

			return kafka.Message{}, io.EOF
		},
		commitFunc: func(ctx context.Context, msgs ...kafka.Message) error {
			committedCount += len(msgs)
			return nil
		},
	}

	consumer := broker.NewKafkaConsumerWithFetcher(cfg, log, fetcher)
	consumer.SetHandler(func(ctx context.Context, msg kafka.Message) error {
		handledCount++
		return nil
	})

	err := consumer.Start(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 2, handledCount)
	assert.Equal(t, 2, committedCount)
}

func TestKafkaConsumer_Start_DefaultHandler(t *testing.T) {
	cfg := testKafkaConfig()
	log := zap.NewNop()

	fetched := false
	fetcher := &mockFetcher{
		fetchFunc: func(ctx context.Context) (kafka.Message, error) {
			if !fetched {
				fetched = true
				return kafka.Message{
					Topic: "test-topic",
					Key:   []byte("test-key"),
					Value: []byte("test-value"),
				}, nil
			}

			return kafka.Message{}, io.EOF
		},
	}

	consumer := broker.NewKafkaConsumerWithFetcher(cfg, log, fetcher)

	err := consumer.Start(context.Background())
	assert.NoError(t, err)
	assert.True(t, fetched)
}

func TestKafkaConsumer_Start_HandlerErrorContinuesLoop(t *testing.T) {
	cfg := testKafkaConfig()
	log := zap.NewNop()

	messages := []kafka.Message{
		{Topic: "test-topic", Key: []byte("fail"), Value: []byte("v1")},
		{Topic: "test-topic", Key: []byte("ok"), Value: []byte("v2")},
	}
	index := 0
	var committedCount int

	fetcher := &mockFetcher{
		fetchFunc: func(ctx context.Context) (kafka.Message, error) {
			if index < len(messages) {
				msg := messages[index]
				index++
				return msg, nil
			}

			return kafka.Message{}, io.EOF
		},
		commitFunc: func(ctx context.Context, msgs ...kafka.Message) error {
			committedCount += len(msgs)
			return nil
		},
	}

	consumer := broker.NewKafkaConsumerWithFetcher(cfg, log, fetcher)
	consumer.SetHandler(func(ctx context.Context, msg kafka.Message) error {
		if string(msg.Key) == "fail" {
			return errors.New("handler error")
		}

		return nil
	})

	err := consumer.Start(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, committedCount)
}

func TestKafkaConsumer_Start_FetchError(t *testing.T) {
	cfg := testKafkaConfig()
	log := zap.NewNop()

	fetcher := &mockFetcher{
		fetchFunc: func(ctx context.Context) (kafka.Message, error) {
			return kafka.Message{}, errors.New("network disconnect")
		},
	}

	consumer := broker.NewKafkaConsumerWithFetcher(cfg, log, fetcher)

	err := consumer.Start(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch kafka message")
}

func TestKafkaConsumer_Close(t *testing.T) {
	cfg := testKafkaConfig()
	log := zap.NewNop()

	closed := false
	fetcher := &mockFetcher{
		closeFunc: func() error {
			closed = true
			return nil
		},
	}

	consumer := broker.NewKafkaConsumerWithFetcher(cfg, log, fetcher)
	err := consumer.Close()

	assert.NoError(t, err)
	assert.True(t, closed)
}
