package main

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func (app *Application) routes() http.Handler {
	router := chi.NewRouter()
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		app.Http.NotFound(w, r, "")
	})
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(app.Recoverer)
	router.Route("/api/v1", func(r chi.Router) {
		r.Get("/healthcheck", app.healthcheck)
		r.Route("/movies", func(r chi.Router) {
			r.Get("/{id}", app.getMovie)
			r.Put("/{id}", app.updateMovie)
			r.Delete("/{id}", app.deleteMovie)
			r.Get("/", app.getMovies)
			r.Post("/", app.createMovie)
		})
	})
	return router
}