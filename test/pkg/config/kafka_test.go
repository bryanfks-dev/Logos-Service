package config_test

import (
	"testing"

	"github.com/bryanfks-dev/selah/pkg/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNewKafka_Unset(t *testing.T) {
	v := viper.New()

	cfg := config.NewKafka(v)

	assert.False(t, cfg.Enabled)
	assert.Empty(t, cfg.Server)
	assert.Empty(t, cfg.Broker)
	assert.Empty(t, cfg.GroupID)
	assert.Empty(t, cfg.OffsetReset)
}

func TestNewKafka_EnabledWithSettings(t *testing.T) {
	v := viper.New()
	v.Set("kafka.enabled", true)
	v.Set("kafka.server", "localhost:9092")
	v.Set("kafka.group.id", "selah")
	v.Set("kafka.offset.reset", "earliest")

	cfg := config.NewKafka(v)

	assert.True(t, cfg.Enabled)
	assert.Equal(t, "localhost:9092", cfg.Server)
	assert.Equal(t, "localhost:9092", cfg.Broker)
	assert.Equal(t, "selah", cfg.GroupID)
	assert.Equal(t, "earliest", cfg.OffsetReset)
}

func TestNewKafka_Disabled(t *testing.T) {
	v := viper.New()
	v.Set("kafka.enabled", false)
	v.Set("kafka.server", "kafka-cluster:9092")
	v.Set("kafka.group.id", "my-group")
	v.Set("kafka.offset.reset", "latest")

	cfg := config.NewKafka(v)

	assert.False(t, cfg.Enabled)
	assert.Equal(t, "kafka-cluster:9092", cfg.Server)
	assert.Equal(t, "kafka-cluster:9092", cfg.Broker)
	assert.Equal(t, "my-group", cfg.GroupID)
	assert.Equal(t, "latest", cfg.OffsetReset)
}
