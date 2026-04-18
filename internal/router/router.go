package router

import (
	"net/http"

	"github.com/ilyinon/go-musthave-metrics/internal/audit"
	"github.com/ilyinon/go-musthave-metrics/internal/handler"
	appmw "github.com/ilyinon/go-musthave-metrics/internal/middleware"
	"github.com/ilyinon/go-musthave-metrics/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func New(storage repository.Storage, auditor *audit.Auditor) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(appmw.Gzip)
	r.Use(appmw.Logger)
	r.Use(middleware.Recoverer)

	indexHandler := handler.NewIndex(storage)

	r.Get("/", indexHandler.ServeHTTP)
	r.Get("/ping", indexHandler.Ping)

	r.Post("/update/{type}/{name}/{value}", handler.NewUpdate(storage, auditor).ServeHTTP)
	r.Post("/update", handler.NewUpdateJSON(storage, auditor).ServeHTTP)
	r.Post("/update/", handler.NewUpdateJSON(storage, auditor).ServeHTTP)

	r.Post("/value", handler.NewValueJSON(storage).ServeHTTP)
	r.Post("/value/", handler.NewValueJSON(storage).ServeHTTP)
	r.Get("/value/{type}/{name}", handler.NewValue(storage).ServeHTTP)

	r.Post("/updates/", handler.NewUpdates(storage, auditor).ServeHTTP)

	return r
}
