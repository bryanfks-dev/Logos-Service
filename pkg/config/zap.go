package config

import (
	"fmt"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(cfg *viper.Viper) *zap.Logger {
	zapConfig := zap.NewProductionConfig()

	level := loggerLevel[cfg.GetString("log.level")]
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	logger, err := zapConfig.Build()
	if err != nil {
		message := fmt.Sprintf("fatal error logger: %v\n", err)
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
