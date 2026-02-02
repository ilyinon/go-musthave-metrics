package router

import (
	"net/http"

	"github.com/ilyinon/go-musthave-metrics/internal/handler"
	appmw "github.com/ilyinon/go-musthave-metrics/internal/middleware"
	"github.com/ilyinon/go-musthave-metrics/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func New(storage repository.Storage) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(appmw.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", handler.NewIndex(storage).ServeHTTP)
	r.Post("/update/{type}/{name}/{value}", handler.NewUpdate(storage).ServeHTTP)
	r.Post("/update", handler.NewUpdateJSON(storage).ServeHTTP)
	r.Post("/update/", handler.NewUpdateJSON(storage).ServeHTTP)

	r.Get("/value/{type}/{name}", handler.NewValue(storage).ServeHTTP)

	return r
}
