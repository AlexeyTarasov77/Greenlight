package services

import (
	"greenlight/proj/internal/clients/sso/grpc"
	"greenlight/proj/internal/config"
	"greenlight/proj/internal/mails"
	"greenlight/proj/internal/services/auth"
	"greenlight/proj/internal/services/movies"
	"greenlight/proj/internal/storage/postgres"
	"log/slog"
)

type Services struct {
	Auth   *auth.AuthService
	Movies *movies.MovieService
}

func New(log *slog.Logger, cfg *config.Config, storage *postgres.PostgresDB, taskExecutor auth.TaskExecutor) *Services {
	// mailer := mails.New(
	// 	cfg.SMTPServer.Host,
	// 	cfg.SMTPServer.Port,
	// 	cfg.SMTPServer.Timeout,
	// 	cfg.SMTPServer.Username,
	// 	cfg.SMTPServer.Password,
	// 	cfg.SMTPServer.Sender,
	// )
	mailer := &mails.ApiMailer{
		ApiToken: cfg.SMTPServer.ApiToken,
		Sender:   cfg.SMTPServer.Sender,
		RetriesCount: cfg.SMTPServer.RetriesCount,
	}
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
	return &Services{
		Auth:   auth.New(log, mailer, sso, taskExecutor),
		Movies: movies.New(log, storage),
	}
}