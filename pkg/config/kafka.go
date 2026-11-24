package config

import "github.com/spf13/viper"

type KafkaConfig struct {
	Enabled     bool
	Server      string
	Broker      string
	GroupID     string
	OffsetReset string
}

func NewKafka(v *viper.Viper) *KafkaConfig {
	enabled := false
	if v.IsSet("kafka.enabled") {
		enabled = v.GetBool("kafka.enabled")
	}

	server := v.GetString("kafka.server")
	groupID := v.GetString("kafka.group.id")
	offsetReset := v.GetString("kafka.offset.reset")

	return &KafkaConfig{
		Enabled:     enabled,
		Server:      server,
		Broker:      server,
		GroupID:     groupID,
		OffsetReset: offsetReset,
	}
}
