package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDatabase(cfg *viper.Viper, log *zap.Logger) *gorm.DB {
	username := cfg.GetString("database.username")
	password := cfg.GetString("database.password")
	host := cfg.GetString("database.host")
	port := cfg.GetInt("database.port")
	database := cfg.GetString("database.name")
	idleConnection := cfg.GetInt("database.pool.idle")
	maxConnection := cfg.GetInt("database.pool.max")
	maxLifeTimeConnection := cfg.GetInt("database.pool.lifetime")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=UTC",
		host,
		username,
		password,
		database,
		port,
	)

	logConfig := logger.Config{
		SlowThreshold:             time.Second * 5,
		Colorful:                  true,
		IgnoreRecordNotFoundError: true,
		ParameterizedQueries:      true,
		LogLevel:                  logger.Info,
	}

	gormLogger := logger.New(zap.NewStdLog(log), logConfig)
	gormConfig := &gorm.Config{Logger: gormLogger}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		log.Sugar().Fatalf("failed to connect to database: %v", err)
	}

	conn, err := db.DB()
	if err != nil {
		log.Sugar().Fatalf("failed to connect to database: %v", err)
	}

	conn.SetMaxIdleConns(idleConnection)
	conn.SetMaxOpenConns(maxConnection)
	conn.SetConnMaxLifetime(time.Second * time.Duration(maxLifeTimeConnection))

	return db
}
