package handlers_test

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"media-platform/internal/config"
	"media-platform/internal/handlers"
	"media-platform/internal/router"
	"media-platform/internal/storage"
)

type compressResponse struct {
	Filename         string  `json:"filename"`
	OriginalSize     int64   `json:"original_size"`
	CompressedSize   int64   `json:"compressed_size"`
	CompressionRatio float64 `json:"compression_ratio"`
	DownloadURL      string  `json:"download_url"`
	ContentType      string  `json:"content_type"`
	Mode             string  `json:"mode"`
	QualityUsed      int     `json:"quality_used"`
	TargetSize       int64   `json:"target_size,omitempty"`
}

func TestHealth(t *testing.T) {
	server := newTestServer(t)

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	server.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", res.Code, http.StatusOK)
	}
}

func TestHome(t *testing.T) {
	server := newTestServer(t)

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	server.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", res.Code, http.StatusOK)
	}
	if contentType := res.Header().Get("Content-Type"); contentType == "" {
		t.Fatal("expected content type header")
	}
}

func TestCompressImageRejectsMissingFile(t *testing.T) {
	server := newTestServer(t)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/images/compress", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	server.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", res.Code, http.StatusBadRequest)
	}
}

func TestCompressImageRejectsUnsupportedFile(t *testing.T) {
	server := newTestServer(t)

	res := upload(t, server, "/api/v1/images/compress", "note.txt", []byte("not an image"))
	if res.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("got status %d, want %d", res.Code, http.StatusUnsupportedMediaType)
	}
}

func TestCompressImageAcceptsJPEGAndDownloads(t *testing.T) {
	server := newTestServer(t)

	res := upload(t, server, "/api/v1/images/compress?quality=70", "photo.jpg", testJPEG(t))
	if res.Code != http.StatusCreated {
		t.Fatalf("got status %d, want %d; body: %s", res.Code, http.StatusCreated, res.Body.String())
	}

	var payload compressResponse
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.ContentType != "image/jpeg" {
		t.Fatalf("got content type %q, want image/jpeg", payload.ContentType)
	}
	if payload.Filename == "" || payload.DownloadURL == "" {
		t.Fatalf("response is missing filename/download URL: %+v", payload)
	}

	download := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, payload.DownloadURL, nil)
	server.ServeHTTP(download, req)
	if download.Code != http.StatusOK {
		t.Fatalf("download got status %d, want %d", download.Code, http.StatusOK)
	}
}

func TestCompressImageAcceptsPNG(t *testing.T) {
	server := newTestServer(t)

	res := upload(t, server, "/api/v1/images/compress", "image.png", testPNG(t))
	if res.Code != http.StatusCreated {
		t.Fatalf("got status %d, want %d; body: %s", res.Code, http.StatusCreated, res.Body.String())
	}

	var payload compressResponse
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.ContentType != "image/png" {
		t.Fatalf("got content type %q, want image/png", payload.ContentType)
	}
}

func TestCompressImageRejectsBadQuality(t *testing.T) {
	server := newTestServer(t)

	res := upload(t, server, "/api/v1/images/compress?quality=101", "photo.jpg", testJPEG(t))
	if res.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", res.Code, http.StatusBadRequest)
	}
}

func TestCompressImageAcceptsTargetSize(t *testing.T) {
	server := newTestServer(t)

	res := upload(t, server, "/api/v1/images/compress?mode=size&target_size=10&target_unit=kb", "photo.jpg", testJPEG(t))
	if res.Code != http.StatusCreated {
		t.Fatalf("got status %d, want %d; body: %s", res.Code, http.StatusCreated, res.Body.String())
	}

	var payload compressResponse
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.TargetSize != 10*1024 {
		t.Fatalf("got target size %d, want %d", payload.TargetSize, 10*1024)
	}
	if payload.Mode != "size" {
		t.Fatalf("got mode %q, want size", payload.Mode)
	}
	if payload.QualityUsed < 1 || payload.QualityUsed > 100 {
		t.Fatalf("quality used %d is out of range", payload.QualityUsed)
	}
}

func TestCompressImageRejectsBadTargetSize(t *testing.T) {
	server := newTestServer(t)

	res := upload(t, server, "/api/v1/images/compress?mode=size&target_size=0&target_unit=kb", "photo.jpg", testJPEG(t))
	if res.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", res.Code, http.StatusBadRequest)
	}
}

func TestCompressImageRejectsSizeModeWithoutTarget(t *testing.T) {
	server := newTestServer(t)

	res := upload(t, server, "/api/v1/images/compress?mode=size", "photo.jpg", testJPEG(t))
	if res.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", res.Code, http.StatusBadRequest)
	}
}

func TestDownloadMissingFile(t *testing.T) {
	server := newTestServer(t)

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/downloads/missing.jpg", nil)
	server.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d", res.Code, http.StatusNotFound)
	}
}

func newTestServer(t *testing.T) http.Handler {
	t.Helper()

	root := t.TempDir()
	cfg := config.Config{
		Port:              "0",
		UploadDir:         filepath.Join(root, "uploads"),
		CompressedDir:     filepath.Join(root, "compressed"),
		TmpDir:            filepath.Join(root, "tmp"),
		MaxUploadBytes:    2 * 1024 * 1024,
		DefaultQuality:    80,
		MaxImageDimension: 2500,
	}
	store := storage.NewLocalStore(cfg.UploadDir, cfg.CompressedDir, cfg.TmpDir)
	if err := store.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs returned error: %v", err)
	}
	return router.New(handlers.NewApp(cfg, store))
}

func upload(t *testing.T, server http.Handler, target, filename string, data []byte) *httptest.ResponseRecorder {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("CreateFormFile returned error: %v", err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatalf("write multipart file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, target, &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	server.ServeHTTP(res, req)
	return res
}

func testJPEG(t *testing.T) []byte {
	t.Helper()

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, testImage(), &jpeg.Options{Quality: 95}); err != nil {
		t.Fatalf("encode jpeg fixture: %v", err)
	}
	return buf.Bytes()
}

func testPNG(t *testing.T) []byte {
	t.Helper()

	var buf bytes.Buffer
	if err := png.Encode(&buf, testImage()); err != nil {
		t.Fatalf("encode png fixture: %v", err)
	}
	return buf.Bytes()
}

func testImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 3), G: uint8(y * 3), B: 150, A: 255})
		}
	}
	return img
}
