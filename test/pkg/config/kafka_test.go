package config_test

import (
	"testing"

	"github.com/bryanfks-dev/selah/pkg/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNewKafkaConfig_Default(t *testing.T) {
	v := viper.New()

	cfg := config.NewKafkaConfig(v)

	assert.Equal(t, "localhost:9092", cfg.Broker)
	assert.Equal(t, "selah-broker-group", cfg.GroupID)
	assert.Equal(t, "selah-events", cfg.Topic)
}

func TestNewKafkaConfig_Custom(t *testing.T) {
	v := viper.New()
	v.Set("kafka.broker", "kafka-single:9092")
	v.Set("kafka.group_id", "custom-group")
	v.Set("kafka.topic", "custom-topic")

	cfg := config.NewKafkaConfig(v)

	assert.Equal(t, "kafka-single:9092", cfg.Broker)
	assert.Equal(t, "custom-group", cfg.GroupID)
	assert.Equal(t, "custom-topic", cfg.Topic)
}
