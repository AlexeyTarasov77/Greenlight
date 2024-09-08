package main

import (
	"greenlight/proj/internal/config"
	"greenlight/proj/internal/services"
	"greenlight/proj/internal/services/movies"
	"greenlight/proj/internal/storage/postgres"
	"log/slog"

	govalidator "github.com/go-playground/validator/v10"
	"github.com/gorilla/schema"
)

type Application struct {
	cfg       *config.Config
	log       *slog.Logger
	Http      *Http
	movies    *movies.MovieService
	validator *govalidator.Validate
	Services  *services.Services
	Decoder   *schema.Decoder
}

func NewApplication(cfg *config.Config, log *slog.Logger, storage *postgres.PostgresDB) *Application {
	services := services.New(log, cfg, storage)
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	app := &Application{
		cfg:       cfg,
		log:       log,
		validator: govalidator.New(govalidator.WithRequiredStructEnabled()),
		Http: &Http{
			log: log,
			cfg: cfg,
		},
		Services: services,
		Decoder: decoder,
	}
	return app
}
