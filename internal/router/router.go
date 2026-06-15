package router

import (
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/ilyinon/go-musthave-metrics/internal/audit"
	"github.com/ilyinon/go-musthave-metrics/internal/handler"
	appmw "github.com/ilyinon/go-musthave-metrics/internal/middleware"
	"github.com/ilyinon/go-musthave-metrics/internal/repository"
)

// New initializes and returns an HTTP router with all endpoints configured.
func New(storage repository.Storage, key string, auditor *audit.Auditor, trustedSubnet *net.IPNet) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RealIP)
	r.Use(appmw.Logger)
	r.Use(chimw.Recoverer)

	// request validation
	r.Use(appmw.HashVerifier(key))

	// response processing
	r.Use(appmw.Gzip)
	r.Use(appmw.HashSigner(key))

	indexHandler := handler.NewIndex(storage)

	r.Get("/", indexHandler.ServeHTTP)
	r.Get("/ping", indexHandler.Ping)

	r.Group(func(r chi.Router) {
		if trustedSubnet != nil {
			r.Use(appmw.TrustedSubnet(trustedSubnet))
		}

		r.Post("/update/{type}/{name}/{value}", handler.NewUpdate(storage, auditor).ServeHTTP)
		r.Post("/update", handler.NewUpdateJSON(storage, auditor).ServeHTTP)
		r.Post("/update/", handler.NewUpdateJSON(storage, auditor).ServeHTTP)
		r.Post("/updates/", handler.NewUpdates(storage, auditor).ServeHTTP)
	})

	r.Post("/value", handler.NewValueJSON(storage).ServeHTTP)
	r.Post("/value/", handler.NewValueJSON(storage).ServeHTTP)
	r.Get("/value/{type}/{name}", handler.NewValue(storage).ServeHTTP)

	return r
}
