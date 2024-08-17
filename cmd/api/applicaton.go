package main

import (
	"greenlight/proj/internal/clients/sso/grpc"
	"greenlight/proj/internal/config"
	"greenlight/proj/internal/services/movies"
	"greenlight/proj/internal/storage/postgres"
	"log/slog"

	govalidator "github.com/go-playground/validator/v10"
)

type Application struct {
	cfg       *config.Config
	log       *slog.Logger
	Http      *Http
	movies    *movies.MovieService
	validator *govalidator.Validate
	Sso       *grpc.Client
}

func NewApplication(cfg *config.Config, log *slog.Logger, storage *postgres.PostgresDB) *Application {
	movies := movies.New(log, storage)
	sso, err := grpc.New(
		log,
		cfg.AppID,
		cfg.Clients.SSO.Addr,
		cfg.Clients.SSO.RetryTimeout,
		cfg.Clients.SSO.RetriesCount,
	)
	if err != nil {
		panic(err)
	}
	app := &Application{
		cfg:       cfg,
		log:       log,
		validator: govalidator.New(govalidator.WithRequiredStructEnabled()),
		movies:    movies,
		Http: &Http{
			log: log,
			cfg: cfg,
		},
		Sso: sso,
	}
	return app
}
