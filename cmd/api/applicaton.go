package main

import (
	"greenlight/proj/internal/config"
	"greenlight/proj/internal/services/movies"
	"greenlight/proj/internal/storage/postgres"
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


func NewApplication(cfg *config.Config, log *slog.Logger, storage *postgres.PostgresDB) *Application {
	movies := movies.New(log, storage)
	app := &Application{
		cfg: cfg,
		log: log,
		validator: govalidator.New(govalidator.WithRequiredStructEnabled()),
		movies: movies,
		Http: &Http{
			log: log,
			cfg: cfg,
		},
	}
	return app
}