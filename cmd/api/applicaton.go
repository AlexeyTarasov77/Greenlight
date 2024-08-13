package main

import (
	"greenlight/proj/internal/config"
	"log/slog"
)


type Application struct {
	cfg *config.Config
	log *slog.Logger
}


func NewApplication(cfg *config.Config, log *slog.Logger) *Application {
	return &Application{
		cfg: cfg,
		log: log,
	}
}