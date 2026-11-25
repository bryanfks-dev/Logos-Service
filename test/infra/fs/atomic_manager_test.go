package fs_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bryanfks-dev/logos/infra/fs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_WriteFileAtomic(t *testing.T) {
	manager := fs.NewManager()
	require.NotNil(t, manager)

	t.Run("returns error when filePath is empty", func(t *testing.T) {
		err := manager.WriteFileAtomic("", []byte("content"))
		assert.EqualError(t, err, "file path is required")
	})

	t.Run(
		"creates parent directories, writes content, and leaves no temp files",
		func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "nested", "notes")
			target := filepath.Join(dir, "note.json")
			payload := []byte(`{"ops":[{"insert":"hello"}]}`)

			err := manager.WriteFileAtomic(target, payload)
			require.NoError(t, err)

			got, err := os.ReadFile(target)
			require.NoError(t, err)
			assert.Equal(t, payload, got)

			entries, err := os.ReadDir(dir)
			require.NoError(t, err)
			require.Len(t, entries, 1)
			assert.Equal(t, "note.json", entries[0].Name())
		},
	)

	t.Run(
		"overwrites existing file atomically with new content",
		func(t *testing.T) {
			target := filepath.Join(t.TempDir(), "note.json")
			require.NoError(
				t,
				manager.WriteFileAtomic(target, []byte("first")),
			)
			require.NoError(
				t,
				manager.WriteFileAtomic(target, []byte("second")),
			)

			got, err := os.ReadFile(target)
			require.NoError(t, err)
			assert.Equal(t, []byte("second"), got)
		},
	)

	t.Run(
		"returns error when target path is an existing directory and cleans temp file",
		func(t *testing.T) {
			parent := t.TempDir()
			targetDir := filepath.Join(parent, "existing-dir")
			require.NoError(t, os.MkdirAll(targetDir, 0o755))

			err := manager.WriteFileAtomic(targetDir, []byte("data"))
			assert.Error(t, err)

			entries, err := os.ReadDir(parent)
			require.NoError(t, err)
			require.Len(t, entries, 1)
			assert.Equal(t, "existing-dir", entries[0].Name())
		},
	)
}
