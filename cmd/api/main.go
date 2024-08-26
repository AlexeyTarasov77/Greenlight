package main

import (
	"context"
	"flag"
	"greenlight/proj/internal/config"
	"greenlight/proj/internal/lib/logger"
	"greenlight/proj/internal/storage/postgres"
	"os"
	"time"
)

const version = "1.0.0"

func main() {
	cfgPath := flag.String("config", "config/local.yml", "path to config file")

	flag.Parse()
	cfg := config.MustLoad(*cfgPath)
	log := logger.SetupLogger(cfg.Debug)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	storage, err := postgres.New(ctx, cfg.DB.Dsn, cfg.DB.MaxConns, cfg.DB.MaxConnIdleTime)
	if err != nil {
		panic(err)
	}
	defer storage.Conn.Close()
	log.Info("database connection established", "dsn", cfg.DB.Dsn)
	app := NewApplication(cfg, log, storage)
	if err := app.serve(); err != nil {
		app.log.Error("Error serving server", "reason", err.Error())
		os.Exit(1)
	}
}