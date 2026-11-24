package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func NewViper() *viper.Viper {
	viper := viper.New()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./")
	viper.AddConfigPath("../")
	viper.AddConfigPath("../../")
	viper.AddConfigPath("../../../")
	viper.AddConfigPath("../../../../")
	viper.AddConfigPath("../../../../../")
	viper.AddConfigPath("../../../../../../")

	if err := viper.ReadInConfig(); err != nil {
		message := fmt.Errorf("fatal error config file: %w", err)
		panic(message)
	}

	return viper
}
