package config_test

import (
	"testing"

	"github.com/bryanfks-dev/logos/pkg/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	v := viper.New()
	server := config.NewServer(v)
	assert.NotNil(t, server)
}
