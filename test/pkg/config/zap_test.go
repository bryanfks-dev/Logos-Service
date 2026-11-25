package config_test

import (
	"testing"

	"github.com/bryanfks-dev/logos/pkg/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	levels := []string{
		"debug",
		"info",
		"warn",
		"error",
		"panic",
		"fatal",
		"unknown_level",
	}

	for _, level := range levels {
		t.Run("level_"+level, func(t *testing.T) {
			v := viper.New()
			v.Set("log.level", level)

			logger := config.NewLogger(v)
			assert.NotNil(t, logger)
		})
	}
}
