package main

import (
	"context"
	"errors"
	"fmt"
	"greenlight/proj/internal/lib/logger"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)


func (app *Application) serve() error {
	server := http.Server{
		Addr:    net.JoinHostPort(app.cfg.Server.Host, app.cfg.Server.Port),
		Handler: app.routes(),
		ReadTimeout: app.cfg.Server.ReadTimeout,
		WriteTimeout: app.cfg.Server.WriteTimeout,
		IdleTimeout: app.cfg.Server.IdleTimeout,
		ErrorLog: logger.LogAdapter(app.log),
	}
	shutdownErr := make(chan error)
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		sig := <-ch
		app.log.Info("shutting down the server gracefully", "signal", sig.String())
		ctx, cancel := context.WithTimeout(context.Background(), app.cfg.Server.ShutdownTimeout)
		defer cancel()
		shutdownErr <- server.Shutdown(ctx)
	}()
	app.log.Info("starting server", "url", fmt.Sprintf("http://%s", server.Addr))
	err := server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	err = <-shutdownErr
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			app.log.Error("graceful shutdown timed out.. forcing exit", "timeout", app.cfg.Server.ShutdownTimeout)
			return fmt.Errorf("graceful shutdown timed out: %w", err)
		}
		return err 
	}
	app.log.Info("Server succesfully stopped")
	return nil
}