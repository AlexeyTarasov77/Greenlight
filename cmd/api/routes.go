package main

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

const activationURL = "PUT '/api/v1/accounts/activation/'"

func (app *Application) routes() http.Handler {
	router := chi.NewRouter()
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		app.Http.NotFound(w, r, "Page not found")
	})
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(app.Recoverer)
	router.Use(app.RateLimiter)
	router.Use(app.Authenticate)
	router.Route("/api/v1", func(r chi.Router) {
		r.Get("/healthcheck", app.healthcheck)
		r.Route("/movies", func(r chi.Router) {
			r.Get("/{id}", app.getMovie)
			r.Patch("/{id}", app.updateMovie)
			r.Delete("/{id}", app.deleteMovie)
			r.Get("/", app.getMovies)
			r.Post("/", app.createMovie)
		})
		r.Route("/accounts", func(r chi.Router) {
			r.Post("/activation/new-token", app.getNewActivationToken)
			r.Put("/activation", app.activateAccount)
			r.Post("/login", app.login)
			r.Post("/signup", app.signup)
		})
	})
	return router
}
