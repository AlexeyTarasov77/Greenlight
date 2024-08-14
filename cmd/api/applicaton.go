package main

import (
	"greenlight/proj/internal/config"
	"greenlight/proj/internal/services/movies"
	"log/slog"

	govalidator "github.com/go-playground/validator/v10"
)

type Application struct {
	cfg *config.Config
	log *slog.Logger
	Http *Http
	movies *movies.MovieService
	validator *govalidator.Validate
}


func NewApplication(cfg *config.Config, log *slog.Logger) *Application {
	app := &Application{
		cfg: cfg,
		log: log,
		validator: govalidator.New(govalidator.WithRequiredStructEnabled()),
		// movies: movies.New(log, ),
		Http: &Http{
			log: log,
			cfg: cfg,
		},
	}
	return app
}