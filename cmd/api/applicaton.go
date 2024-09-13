package main

import (
	"greenlight/proj/internal/api/tasks"
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
	BackgroundTasks *tasks.BackgroudTasks
}

func NewApplication(cfg *config.Config, log *slog.Logger, storage *postgres.PostgresDB) *Application {
	bgTasks := tasks.New(log, 3, 10)
	bgTasks.Run()
	services := services.New(log, cfg, storage, bgTasks)
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
		BackgroundTasks: bgTasks,
	}
	return app
}
