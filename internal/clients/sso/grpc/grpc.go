package grpc

import (
	"context"
	"log/slog"
	"time"

	ssov1 "github.com/AlexeySHA256/protos/gen/go/sso"
	grpclogging "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	api       ssov1.AuthClient
	log       *slog.Logger
	appId     int32
}

// New creates a new Client instance.
//
// It takes a context, a logger, an address of the gRPC server, a timeout for retry call, and a retries count as parameters.
// Returns a Client instance and an error.
func New(
	log *slog.Logger,
	appId int32,
	addr string,
	timeout time.Duration,
	retriesCount int,
) (*Client, error) {
	retryOpts := []grpcretry.CallOption{
		grpcretry.WithPerRetryTimeout(timeout),
		grpcretry.WithMax(uint(retriesCount)),
		grpcretry.WithCodes(codes.Aborted, codes.DeadlineExceeded),
	}
	logOpts := []grpclogging.Option{
		grpclogging.WithLogOnEvents(grpclogging.PayloadReceived, grpclogging.PayloadSent),
	}
	cc, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpcretry.UnaryClientInterceptor(retryOpts...),
			grpclogging.UnaryClientInterceptor(InterceptorLogger(log), logOpts...),
		),
	)
	if err != nil {
		return nil, err
	}
	return &Client{
		api:       ssov1.NewAuthClient(cc),
		log:       log,
		appId:     appId,
	}, nil
}

func (c *Client) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	_, err := c.api.IsAdmin(ctx, &ssov1.IsAdminRequest{UserId: userID})
	if err != nil {
		c.log.Error("Error calling Client.IsAdmin", "errMsg", err.Error())
		return false, err
	}
	return true, nil
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (c *Client) Login(ctx context.Context, email, password string) (*Tokens, error) {
	resp, err := c.api.Login(ctx, &ssov1.LoginRequest{Email: email, Password: password, AppId: c.appId})
	if err != nil {
		c.log.Error("Error calling Client.Login", "errMsg", err.Error())
		return nil, err
	}
	return &Tokens{AccessToken: resp.AccessToken, RefreshToken: resp.RefreshToken}, nil
}

func (c *Client) Register(ctx context.Context, email, password string) (int64, error) {
	resp, err := c.api.Register(
		ctx,
		&ssov1.RegisterRequest{Email: email, Password: password, Username: email},
	)
	if err != nil {
		c.log.Error("Error calling Client.Register", "errMsg", err.Error())
		return 0, err
	}
	return resp.GetUserId(), nil
}

// Adapter for grpclogging.Logger used to adapt it to slog.Logger
func InterceptorLogger(log *slog.Logger) grpclogging.Logger {
	return grpclogging.LoggerFunc(
		func(ctx context.Context, level grpclogging.Level, msg string, fields ...any) {
			log.Log(ctx, slog.Level(level), msg, fields...)
		},
	)
}
