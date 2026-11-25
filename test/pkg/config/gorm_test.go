package config_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/bryanfks-dev/logos/pkg/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewDatabase(t *testing.T) {
	v := config.NewViper()
	logger := zap.NewNop()

	host := v.GetString("database.host")
	port := v.GetInt("database.port")
	conn, err := net.DialTimeout(
		"tcp",
		fmt.Sprintf("%s:%d", host, port),
		500*time.Millisecond,
	)
	if err != nil {
		t.Skipf("skipping live database test: %v", err)
	}
	_ = conn.Close()

	db := config.NewDatabase(v, logger)
	assert.NotNil(t, db)

	sqlDB, err := db.DB()
	assert.NoError(t, err)
	assert.NotNil(t, sqlDB)
}
