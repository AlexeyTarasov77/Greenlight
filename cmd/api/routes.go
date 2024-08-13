package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *Application) routes() http.Handler {
	router := chi.NewRouter()
	router.Route("/api/v1", func(r chi.Router) {
		r.Get("/healthcheck", app.healthcheck)
	})
	return router
}