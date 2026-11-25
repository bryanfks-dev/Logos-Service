package config

import "github.com/spf13/viper"

type KafkaConfig struct {
	Enabled     bool
	Server      string
	Broker      string
	GroupID     string
	OffsetReset string
}

func NewKafka(cfg *viper.Viper) *KafkaConfig {
	enabled := false
	if cfg.IsSet("kafka.enabled") {
		enabled = cfg.GetBool("kafka.enabled")
	}

	server := cfg.GetString("kafka.server")
	groupID := cfg.GetString("kafka.group.id")
	offsetReset := cfg.GetString("kafka.offset.reset")

	return &KafkaConfig{
		Enabled:     enabled,
		Server:      server,
		Broker:      server,
		GroupID:     groupID,
		OffsetReset: offsetReset,
	}
}
