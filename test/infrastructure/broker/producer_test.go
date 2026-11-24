package broker_test

import (
	"context"
	"errors"
	"testing"

	"github.com/bryanfks-dev/selah/infrastructure/broker"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockWriter struct {
	writeFunc func(ctx context.Context, msgs ...kafka.Message) error
	closeFunc func() error
}

func (m *mockWriter) WriteMessages(
	ctx context.Context,
	msgs ...kafka.Message,
) error {
	if m.writeFunc != nil {
		return m.writeFunc(ctx, msgs...)
	}

	return nil
}

func (m *mockWriter) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}

	return nil
}

func TestNewKafkaProducer(t *testing.T) {
	cfg := testKafkaConfig()
	log := zap.NewNop()

	producer := broker.NewKafkaProducer(cfg, log)
	require.NotNil(t, producer)

	err := producer.Close()
	assert.NoError(t, err)
}

func TestKafkaProducer_Publish_Success(t *testing.T) {
	cfg := testKafkaConfig()
	log := zap.NewNop()

	var writtenMsgs []kafka.Message
	writer := &mockWriter{
		writeFunc: func(ctx context.Context, msgs ...kafka.Message) error {
			writtenMsgs = append(writtenMsgs, msgs...)
			return nil
		},
	}

	producer := broker.NewKafkaProducerWithWriter(cfg, log, writer)

	err := producer.Publish(
		context.Background(),
		[]byte("order-123"),
		[]byte(`{"event":"created"}`),
	)

	assert.NoError(t, err)
	require.Len(t, writtenMsgs, 1)
	assert.Equal(t, []byte("order-123"), writtenMsgs[0].Key)
	assert.Equal(t, []byte(`{"event":"created"}`), writtenMsgs[0].Value)
}

func TestKafkaProducer_PublishMessage_UninitializedWriter(t *testing.T) {
	cfg := testKafkaConfig()
	log := zap.NewNop()

	producer := broker.NewKafkaProducerWithWriter(cfg, log, nil)

	err := producer.PublishMessage(context.Background(), kafka.Message{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "uninitialized")
}

func TestKafkaProducer_PublishMessage_WriteError(t *testing.T) {
	cfg := testKafkaConfig()
	log := zap.NewNop()

	writer := &mockWriter{
		writeFunc: func(ctx context.Context, msgs ...kafka.Message) error {
			return errors.New("kafka broker down")
		},
	}

	producer := broker.NewKafkaProducerWithWriter(cfg, log, writer)

	err := producer.Publish(
		context.Background(),
		[]byte("key"),
		[]byte("value"),
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write kafka message")
}

func TestKafkaProducer_Close(t *testing.T) {
	cfg := testKafkaConfig()
	log := zap.NewNop()

	closed := false
	writer := &mockWriter{
		closeFunc: func() error {
			closed = true
			return nil
		},
	}

	producer := broker.NewKafkaProducerWithWriter(cfg, log, writer)
	err := producer.Close()

	assert.NoError(t, err)
	assert.True(t, closed)
}
