package main

import (
	"flag"
	"fmt"
	"greenlight/proj/internal/config"
	"greenlight/proj/internal/lib/logger/handlers/slogpretty"
	"greenlight/proj/internal/storage/postgres"
	"log/slog"
	"net"
	"net/http"
	"os"
)

const version = "1.0.0"

func main() {
	cfgPath := flag.String("config", "config/local.yml", "path to config file")

	flag.Parse()
	cfg := config.MustLoad(*cfgPath)
	log := setupLogger(cfg.Debug)
	storage, err := postgres.New(cfg.DB.Dsn, cfg.DB.MaxConns, cfg.DB.MaxConnIdleTime)
	if err != nil {
		panic(err)
	}
	defer storage.Conn.Close()
	log.Info("database connection established", "dsn", cfg.DB.Dsn)
	app := NewApplication(cfg, log, storage)
	router := app.routes()
	server := http.Server{
		Addr:    net.JoinHostPort(cfg.Server.Host, cfg.Server.Port),
		Handler: router,
		ReadTimeout: cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout: cfg.Server.IdleTimeout,
	}

	app.log.Info("starting server", "url", fmt.Sprintf("http://%s", server.Addr))
	if err := server.ListenAndServe(); err != nil {
		app.log.Error("shutting down the server", "reason", err.Error())
		os.Exit(1)
	}
}

func setupLogger(debug bool) *slog.Logger {
	if debug {
		return slog.New(slogpretty.NewPrettyHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
}