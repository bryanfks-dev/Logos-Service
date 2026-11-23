package config

import (
	"fmt"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(viper *viper.Viper) *zap.Logger {
	config := zap.NewProductionConfig()

	level := loggerLevel[viper.GetString("log.level")]
	config.Level = zap.NewAtomicLevelAt(level)

	logger, err := config.Build()
	if err != nil {
		message := fmt.Sprintf("Fatal error logger: %v\n", err)
		panic(message)
	}

	return logger
}

var loggerLevel = map[string]zapcore.Level{
	"debug": zap.DebugLevel,
	"info":  zap.InfoLevel,
	"warn":  zap.WarnLevel,
	"error": zap.ErrorLevel,
	"panic": zap.PanicLevel,
	"fatal": zap.FatalLevel,
}
