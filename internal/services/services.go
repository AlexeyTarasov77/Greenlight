package services

import (
	"greenlight/proj/internal/clients/sso/grpc"
	"greenlight/proj/internal/config"
	"greenlight/proj/internal/mails"
	"greenlight/proj/internal/services/auth"
	authmocks "greenlight/proj/internal/services/auth/mocks"
	"greenlight/proj/internal/services/movies"
	"greenlight/proj/internal/services/reviews"
	"greenlight/proj/internal/storage/postgres"
	"greenlight/proj/internal/storage/postgres/models"
	"log/slog"
	"os"
	"testing"
)

type Services struct {
	Auth    *auth.AuthService
	Movies  *movies.MovieService
	Reviews *reviews.ReviewService
}

func New(log *slog.Logger, cfg *config.Config, storage *postgres.Storage, taskExecutor auth.TaskExecutor) *Services {
	mailer := &mails.ApiMailer{
		ApiToken:     cfg.SMTPServer.ApiToken,
		Sender:       cfg.SMTPServer.Sender,
		RetriesCount: cfg.SMTPServer.RetriesCount,
	}
	models := models.New(storage)
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
		Auth:    auth.New(log, mailer, sso, taskExecutor),
		Movies:  movies.New(log, models.Movie, models.Review),
		Reviews: reviews.New(log, models.Review),
	}
}

func NewTestServices(t *testing.T) *Services {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	return &Services{
		Auth:   auth.New(log, authmocks.NewMailProvider(t), authmocks.NewSsoProvider(t), authmocks.NewTaskExecutor(t)),
		Movies: movies.New(log, nil, nil),
	}
}
