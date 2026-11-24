package config_test

import (
	"os"
	"testing"

	"github.com/bryanfks-dev/selah/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestNewViper(t *testing.T) {
	t.Run("reads config file successfully from workspace", func(t *testing.T) {
		v := config.NewViper()
		assert.NotNil(t, v)
		assert.Equal(t, "skeleton", v.GetString("app.name"))
		assert.Equal(t, 3333, v.GetInt("web.port"))
		assert.Equal(t, "selah_dev", v.GetString("database.name"))
		assert.NotEmpty(t, v.GetString("database.host"))
	})

	t.Run("panics when config file not found", func(t *testing.T) {
		tempDir := t.TempDir()
		origDir, err := os.Getwd()
		assert.NoError(t, err)

		defer func() {
			_ = os.Chdir(origDir)
		}()

		err = os.Chdir(tempDir)
		assert.NoError(t, err)

		assert.Panics(t, func() {
			config.NewViper()
		})
	})
}
