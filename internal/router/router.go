package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"media-platform/internal/handlers"
)

func New(app *handlers.App) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", app.Home)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", app.Health)
		r.Post("/images/compress", app.CompressImage)
		r.Get("/downloads/{filename}", app.Download)
	})

	return r
}
