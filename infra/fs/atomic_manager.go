package fs

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Manager interface {
	EnsureDir(dirPath string) error
	WriteFileAtomic(filePath string, data []byte) error
	WriteTempAndHash(tmpPath string, r io.Reader) (string, int64, error)
	MoveAtomic(srcPath string, dstPath string) error
	RemoveFile(filePath string) error
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

func (m *manager) WriteTempAndHash(
	tmpPath string,
	r io.Reader,
) (string, int64, error) {
	if strings.TrimSpace(tmpPath) == "" {
		return "", 0, errors.New("temp file path is required")
	}

	if r == nil {
		return "", 0, errors.New("reader is required")
	}

	dir := filepath.Dir(tmpPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", 0, err
	}

	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return "", 0, err
	}

	cleanup := func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
	}

	hasher := sha256.New()
	tee := io.TeeReader(r, hasher)

	written, err := io.Copy(tmpFile, tee)
	if err != nil {
		cleanup()
		return "", 0, err
	}

	if err := tmpFile.Chmod(0o644); err != nil {
		cleanup()
		return "", 0, err
	}

	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return "", 0, err
	}

	digest := hex.EncodeToString(hasher.Sum(nil))

	return digest, written, nil
}

func (m *manager) MoveAtomic(srcPath string, dstPath string) error {
	if strings.TrimSpace(srcPath) == "" || strings.TrimSpace(dstPath) == "" {
		return errors.New("source and destination paths are required")
	}

	dir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	return os.Rename(srcPath, dstPath)
}

func (m *manager) RemoveFile(filePath string) error {
	if strings.TrimSpace(filePath) == "" {
		return nil
	}

	err := os.Remove(filePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return err
}
