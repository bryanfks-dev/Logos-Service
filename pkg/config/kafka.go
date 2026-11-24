package config

import "github.com/spf13/viper"

type KafkaConfig struct {
	Broker  string
	GroupID string
	Topic   string
}

func NewKafkaConfig(v *viper.Viper) *KafkaConfig {
	broker := v.GetString("kafka.broker")
	if broker == "" {
		broker = "localhost:9092"
	}

	groupID := v.GetString("kafka.group_id")
	if groupID == "" {
		groupID = "selah-broker-group"
	}

	topic := v.GetString("kafka.topic")
	if topic == "" {
		topic = "selah-events"
	}

	return &KafkaConfig{
		Broker:  broker,
		GroupID: groupID,
		Topic:   topic,
	}
}
