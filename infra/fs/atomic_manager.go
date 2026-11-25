package fs

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type Manager interface {
	EnsureDir(dirPath string) error
	WriteFileAtomic(filePath string, data []byte) error
}

type manager struct{}

func NewManager() Manager {
	return &manager{}
}

func (m *manager) EnsureDir(dirPath string) error {
	if strings.TrimSpace(dirPath) == "" {
		return errors.New("directory path is required")
	}

	return os.MkdirAll(dirPath, 0o755)
}

func (m *manager) WriteFileAtomic(filePath string, data []byte) error {
	if filePath == "" {
		return errors.New("file path is required")
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}

	tmpPath := tmpFile.Name()
	cleanup := func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
	}

	if _, err := tmpFile.Write(data); err != nil {
		cleanup()
		return err
	}

	if err := tmpFile.Chmod(0o644); err != nil {
		cleanup()
		return err
	}

	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	if err := os.Rename(tmpPath, filePath); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	return nil
}
