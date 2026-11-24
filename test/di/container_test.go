package di_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/bryanfks-dev/selah/di"
	"github.com/bryanfks-dev/selah/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewContainer(t *testing.T) {
	v := config.NewViper()
	host := v.GetString("database.host")
	port := v.GetInt("database.port")
	conn, err := net.DialTimeout(
		"tcp",
		fmt.Sprintf("%s:%d", host, port),
		500*time.Millisecond,
	)
	if err != nil {
		t.Skipf("skipping container test with live db: %v", err)
	}
	_ = conn.Close()

	container, err := di.NewContainer()
	require.NoError(t, err)
	require.NotNil(t, container)

	assert.NotNil(t, container.Config)
	assert.NotNil(t, container.Log)
	assert.NotNil(t, container.App)
	assert.NotNil(t, container.TxManager)
	assert.NotNil(t, container.RouteConfig)
	assert.NotNil(t, container.EtcController)
	assert.NotNil(t, container.ErrorController)
	assert.NotNil(t, container.SyncHandler)
	assert.NotNil(t, container.SyncUsecase)

	// Verify route setup completes without panic
	assert.NotPanics(t, func() {
		container.RouteConfig.Setup()
	})
}
