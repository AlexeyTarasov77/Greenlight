package main

import (
	"context"
	"flag"
	"fmt"
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
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	storage, err := postgres.New(ctx, cfg.DB.GetDsn(), cfg.DB.MaxConns, cfg.DB.MaxConnIdleTime)
	if err != nil {
		panic(fmt.Errorf("failed to connect to database: %w", err))
	}
	defer storage.Conn.Close()
	log.Info("database connection established", "dsn", cfg.DB.GetDsn())
	app := NewApplication(cfg, log, storage)
	if err := app.serve(); err != nil {
		app.log.Error("Error serving server", "reason", err.Error())
		os.Exit(1)
	}
}