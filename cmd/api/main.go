package main

import (
	"context"
	"flag"
	"fmt"
	"greenlight/proj/internal/config"
	"greenlight/proj/internal/lib/logger/handlers/slogpretty"
	"greenlight/proj/internal/storage/postgres"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"
)

const version = "1.0.0"

func main() {
	cfgPath := flag.String("config", "config/local.yml", "path to config file")

	flag.Parse()
	cfg := config.MustLoad(*cfgPath)
	log := setupLogger(cfg.Debug)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	storage, err := postgres.New(ctx, cfg.DB.Dsn, cfg.DB.MaxConns, cfg.DB.MaxConnIdleTime)
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
		ErrorLog: logAdapter(log),
	}

	app.log.Info("starting server", "url", fmt.Sprintf("http://%s", server.Addr))
	if err := server.ListenAndServe(); err != nil {
		app.log.Error("shutting down the server", "reason", err.Error())
		os.Exit(1)
	}
}

func setupLogger(debug bool) *slog.Logger {
	var handler slog.Handler
	if debug {
		handler = slogpretty.NewPrettyHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	return slog.New(handler)
}

type out struct {
	stdLog *slog.Logger
}
func (l out) Write(p []byte) (n int, err error) {
	l.stdLog.Info(string(p))
	return len(p), nil
}

func logAdapter(logger *slog.Logger) *log.Logger {
	return log.New(&out{logger}, "", 0)
}