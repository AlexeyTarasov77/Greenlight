package main

import (
	"context"
	"expvar"
	"flag"
	"fmt"
	"greenlight/proj/internal/config"
	"greenlight/proj/internal/lib/logger"
	"greenlight/proj/internal/storage/postgres"
	"os"
	"runtime"
	"time"
)

const version = "1.0.0"

func main() {
	expvar.NewString("version").Set(version)

	expvar.Publish("runningGoroutinesNum", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	expvar.Publish("timestamp", expvar.Func(func() interface{} {
		return time.Now().Unix()
	}))

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
	expvar.Publish("database", expvar.Func(func() any {
		stats := storage.Conn.Stat()
		return map[string]any{
			"dsn": cfg.DB.GetDsn(),
			"AcquireCount": stats.AcquireCount(),
			"AcquireDuration": stats.AcquireDuration().String(),
			"IdleConns": stats.IdleConns(),
			"MaxConns": stats.MaxConns(),
			"TotalConns": stats.TotalConns(),
			"ConstructingConns": stats.ConstructingConns(),
			"AcquiredConns": stats.AcquiredConns(),
			"CanceledAcquireCount": stats.CanceledAcquireCount(),
			"EmptyAcquireCount": stats.EmptyAcquireCount(),
			"NewConnsCount": stats.NewConnsCount(),
            "MaxOpenConns": cfg.DB.MaxConns,
            "MaxIdleConns": cfg.DB.MaxConnIdleTime,
		}
	}))

	defer storage.Conn.Close()
	log.Info("database connection established", "dsn", cfg.DB.GetDsn())
	app := NewApplication(cfg, log, storage)
	if err := app.serve(); err != nil {
		app.log.Error("Error serving server", "reason", err.Error())
		os.Exit(1)
	}
}
