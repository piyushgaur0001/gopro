package utils

import (
	"errors"
	"mime/multipart"
	"net/http"
	"strconv"
)

const (
	MinQuality = 1
	MaxQuality = 100
)

func ParseQuality(raw string, fallback int) (int, error) {
	if raw == "" {
		return clampQuality(fallback), nil
	}

	quality, err := strconv.Atoi(raw)
	if err != nil {
		return 0, errors.New("quality must be a number")
	}
	if quality < MinQuality || quality > MaxQuality {
		return 0, errors.New("quality must be between 1 and 100")
	}
	return quality, nil
}

func ParseTargetSizeKB(raw string) (int64, bool, error) {
	if raw == "" {
		return 0, false, nil
	}

	sizeKB, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, false, errors.New("target_size_kb must be a number")
	}
	if sizeKB <= 0 {
		return 0, false, errors.New("target_size_kb must be greater than 0")
	}
	return sizeKB * 1024, true, nil
}

func DetectImageType(data []byte) (contentType, extension string, ok bool) {
	contentType = http.DetectContentType(data)
	switch contentType {
	case "image/jpeg":
		return contentType, ".jpg", true
	case "image/png":
		return contentType, ".png", true
	default:
		return contentType, "", false
	}
}

func CompressionRatio(originalSize, compressedSize int64) float64 {
	if originalSize <= 0 {
		return 0
	}
	return float64(originalSize-compressedSize) / float64(originalSize)
}

func UploadedFileSize(header *multipart.FileHeader) int64 {
	if header == nil {
		return 0
	}
	return header.Size
}

func clampQuality(quality int) int {
	if quality < MinQuality {
		return MinQuality
	}
	if quality > MaxQuality {
		return MaxQuality
	}
	return quality
}
