package handlers

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"

	"media-platform/internal/compressor"
	"media-platform/internal/config"
	"media-platform/internal/storage"
	"media-platform/internal/utils"
)

type App struct {
	cfg   config.Config
	store *storage.LocalStore
}

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

func NewApp(cfg config.Config, store *storage.LocalStore) *App {
	return &App{cfg: cfg, store: store}
}

func (a *App) Home(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, indexHTMLPath())
}

func (a *App) Health(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (a *App) CompressImage(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, a.cfg.MaxUploadBytes)
	if err := r.ParseMultipartForm(a.cfg.MaxUploadBytes); err != nil {
		utils.WriteError(w, http.StatusRequestEntityTooLarge, "upload is too large or malformed")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "missing multipart file field: file")
		return
	}
	defer file.Close()

	mode := r.URL.Query().Get("mode")
	if mode == "" {
		mode = "quality"
	}
	if mode != "quality" && mode != "size" {
		utils.WriteError(w, http.StatusBadRequest, "mode must be quality or size")
		return
	}

	quality, err := utils.ParseQuality(r.URL.Query().Get("quality"), a.cfg.DefaultQuality)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	targetBytes, targetSet, err := utils.ParseTargetSize(
		r.URL.Query().Get("target_size"),
		r.URL.Query().Get("target_unit"),
		r.URL.Query().Get("target_size_kb"),
	)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if mode == "size" && !targetSet {
		utils.WriteError(w, http.StatusBadRequest, "target_size is required when mode is size")
		return
	}
	if mode == "quality" {
		targetBytes = 0
	}

	data, err := io.ReadAll(file)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "failed to read uploaded file")
		return
	}
	if len(data) == 0 {
		utils.WriteError(w, http.StatusBadRequest, "uploaded file is empty")
		return
	}

	contentType, extension, ok := utils.DetectImageType(data)
	if !ok {
		utils.WriteError(w, http.StatusUnsupportedMediaType, "unsupported image type; only jpg, jpeg, and png are supported")
		return
	}

	originalFilename, err := a.store.NewFilename(extension)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to generate upload filename")
		return
	}
	if err := a.store.SaveUpload(originalFilename, data); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to store uploaded file")
		return
	}

	result, err := compressor.CompressImage(data, compressor.ImageOptions{
		Quality:      quality,
		TargetBytes:  targetBytes,
		MaxDimension: a.cfg.MaxImageDimension,
	})
	if err != nil {
		utils.WriteError(w, http.StatusUnprocessableEntity, "failed to compress image")
		return
	}

	outputFilename, err := a.store.NewFilename(result.Extension)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to generate compressed filename")
		return
	}
	if err := a.store.SaveCompressed(outputFilename, result.Data); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to store compressed file")
		return
	}

	originalSize := int64(len(data))
	if header.Size > 0 {
		originalSize = header.Size
	}
	compressedSize := int64(len(result.Data))
	if result.ContentType != contentType {
		contentType = result.ContentType
	}

	utils.WriteJSON(w, http.StatusCreated, compressResponse{
		Filename:         outputFilename,
		OriginalSize:     originalSize,
		CompressedSize:   compressedSize,
		CompressionRatio: utils.CompressionRatio(originalSize, compressedSize),
		DownloadURL:      "/api/v1/downloads/" + outputFilename,
		ContentType:      contentType,
		Mode:             mode,
		QualityUsed:      result.QualityUsed,
		TargetSize:       result.TargetBytes,
	})
}

func (a *App) Download(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	file, info, err := a.store.OpenCompressed(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			utils.WriteError(w, http.StatusNotFound, "compressed file not found")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "failed to open compressed file")
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", contentTypeForFilename(filename))
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	http.ServeContent(w, r, filename, info.ModTime(), file)
}

func contentTypeForFilename(filename string) string {
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".jpg"), strings.HasSuffix(lower, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(lower, ".png"):
		return "image/png"
	default:
		return "application/octet-stream"
	}
}

func indexHTMLPath() string {
	for _, path := range []string{"web/index.html", "../../web/index.html"} {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return "web/index.html"
}
