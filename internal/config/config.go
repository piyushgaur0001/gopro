package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port              string
	UploadDir         string
	CompressedDir     string
	TmpDir            string
	MaxUploadBytes    int64
	DefaultQuality    int
	MaxImageDimension int
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
}

func Load() Config {
	return Config{
		Port:              getEnv("PORT", "8080"),
		UploadDir:         getEnv("UPLOAD_DIR", "uploads"),
		CompressedDir:     getEnv("COMPRESSED_DIR", "compressed"),
		TmpDir:            getEnv("TMP_DIR", "tmp"),
		MaxUploadBytes:    int64(getEnvInt("MAX_UPLOAD_MB", 25)) * 1024 * 1024,
		DefaultQuality:    getEnvInt("DEFAULT_QUALITY", 80),
		MaxImageDimension: getEnvInt("MAX_IMAGE_DIMENSION", 2500),
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      60 * time.Second,
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
