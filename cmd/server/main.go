package main

import (
	"log"
	"net/http"

	"media-platform/internal/config"
	"media-platform/internal/handlers"
	"media-platform/internal/router"
	"media-platform/internal/storage"
)

func main() {
	cfg := config.Load()

	store := storage.NewLocalStore(cfg.UploadDir, cfg.CompressedDir, cfg.TmpDir)
	if err := store.EnsureDirs(); err != nil {
		log.Fatalf("ensure storage directories: %v", err)
	}

	app := handlers.NewApp(cfg, store)
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router.New(app),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	log.Printf("media platform API listening on :%s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
