package fs_test

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
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

func TestManager_EnsureDir(t *testing.T) {
	manager := fs.NewManager()
	require.NotNil(t, manager)

	t.Run(
		"returns error when dirPath is empty or whitespace",
		func(t *testing.T) {
			assert.EqualError(
				t,
				manager.EnsureDir(""),
				"directory path is required",
			)
			assert.EqualError(
				t,
				manager.EnsureDir("   "),
				"directory path is required",
			)
		},
	)

	t.Run("creates directory if it does not exist", func(t *testing.T) {
		targetDir := filepath.Join(t.TempDir(), "app", "files")

		err := manager.EnsureDir(targetDir)
		require.NoError(t, err)

		info, err := os.Stat(targetDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("succeeds when directory already exists", func(t *testing.T) {
		targetDir := filepath.Join(t.TempDir(), "existing")
		require.NoError(t, os.MkdirAll(targetDir, 0o755))

		err := manager.EnsureDir(targetDir)
		require.NoError(t, err)

		info, err := os.Stat(targetDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run(
		"returns error when target path is an existing file",
		func(t *testing.T) {
			targetFile := filepath.Join(t.TempDir(), "file.txt")
			require.NoError(t, os.WriteFile(targetFile, []byte("data"), 0o644))

			err := manager.EnsureDir(targetFile)
			assert.Error(t, err)
		},
	)
}

func TestManager_WriteTempAndHashTempFile_MoveAtomic_RemoveFile(t *testing.T) {
	manager := fs.NewManager()
	require.NotNil(t, manager)

	t.Run("WriteTempAndHashTempFile validates inputs", func(t *testing.T) {
		_, _, err := manager.WriteTempAndHashTempFile(
			"",
			strings.NewReader("abc"),
		)
		assert.EqualError(t, err, "temp file path is required")

		_, _, err = manager.WriteTempAndHashTempFile(
			filepath.Join(t.TempDir(), "a.part"),
			nil,
		)
		assert.EqualError(t, err, "reader is required")
	})

	t.Run(
		"WriteTempAndHashTempFile streams, hashes, and MoveAtomic moves atomically",
		func(t *testing.T) {
			baseDir := t.TempDir()
			tmpPath := filepath.Join(baseDir, ".tmp", "att-1.part")
			finalPath := filepath.Join(baseDir, "2026", "att-1.pdf")
			payload := "hello attachment bytes"

			sum := sha256.Sum256([]byte(payload))
			expectedHash := hex.EncodeToString(sum[:])

			gotHash, written, err := manager.WriteTempAndHashTempFile(
				tmpPath,
				strings.NewReader(payload),
			)
			require.NoError(t, err)
			assert.Equal(t, expectedHash, gotHash)
			assert.Equal(t, int64(len(payload)), written)

			require.NoError(t, manager.MoveAtomic(tmpPath, finalPath))

			_, err = os.Stat(tmpPath)
			assert.True(t, os.IsNotExist(err))

			gotBytes, err := os.ReadFile(finalPath)
			require.NoError(t, err)
			assert.Equal(t, payload, string(gotBytes))

			require.NoError(t, manager.RemoveFile(finalPath))
			require.NoError(t, manager.RemoveFile(finalPath))
		},
	)
}
