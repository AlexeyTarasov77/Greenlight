package main

import (
	"expvar"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

const activationURL = "PUT '/api/v1/accounts/activation'"

func (app *Application) routes() http.Handler {
	router := chi.NewRouter()
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		app.Http.NotFound(w, r, "Page not found")
	})

	router.Use(app.metrics)
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(app.Recoverer)
	allowedOrigins := app.cfg.CORS.AllowedOrigins
	if len(allowedOrigins) > 0 {
		router.Use(app.enableCORS(allowedOrigins))
	}
	router.Use(app.RateLimiter)
	router.Use(app.Authenticate)
    router.Get("/debug/vars", expvar.Handler().(http.HandlerFunc))
	router.Route("/api/v1", func(r chi.Router) {
		r.Get("/healthcheck", app.healthcheck)
		r.Route("/movies", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(app.requirePermission("movies:read"))
				r.Get("/{id}", app.getMovie)
				r.Get("/", app.getMovies)
			})
			r.Group(func(r chi.Router) {
				r.Use(app.requirePermission("movies:write"))
				r.Patch("/{id}", app.updateMovie)
				r.Delete("/{id}", app.deleteMovie)
				r.Post("/", app.createMovie)
				r.Post("/{id}/review", app.addReviewForMovie)
			})
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
