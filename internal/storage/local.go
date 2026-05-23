package storage

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type LocalStore struct {
	UploadDir     string
	CompressedDir string
	TmpDir        string
}

func NewLocalStore(uploadDir, compressedDir, tmpDir string) *LocalStore {
	return &LocalStore{
		UploadDir:     uploadDir,
		CompressedDir: compressedDir,
		TmpDir:        tmpDir,
	}
}

func (s *LocalStore) EnsureDirs() error {
	for _, dir := range []string{s.UploadDir, s.CompressedDir, s.TmpDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func (s *LocalStore) NewFilename(extension string) (string, error) {
	if extension == "" || !strings.HasPrefix(extension, ".") {
		return "", errors.New("invalid file extension")
	}

	token := make([]byte, 16)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	return hex.EncodeToString(token) + strings.ToLower(extension), nil
}

func (s *LocalStore) SaveUpload(filename string, data []byte) error {
	return writeFile(filepath.Join(s.UploadDir, filename), data)
}

func (s *LocalStore) SaveCompressed(filename string, data []byte) error {
	return writeFile(filepath.Join(s.CompressedDir, filename), data)
}

func (s *LocalStore) OpenCompressed(filename string) (*os.File, os.FileInfo, error) {
	if !IsSafeFilename(filename) {
		return nil, nil, os.ErrNotExist
	}

	file, err := os.Open(filepath.Join(s.CompressedDir, filename))
	if err != nil {
		return nil, nil, err
	}

	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, nil, err
	}
	return file, info, nil
}

func IsSafeFilename(filename string) bool {
	if filename == "" {
		return false
	}
	if filename != filepath.Base(filename) {
		return false
	}
	if strings.Contains(filename, "..") {
		return false
	}
	return true
}

func writeFile(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, bytes.NewReader(data))
	return err
}
