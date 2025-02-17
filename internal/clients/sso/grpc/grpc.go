package grpc

import (
	"context"
	"fmt"
	"greenlight/proj/internal/domain/models"
	"greenlight/proj/internal/services/auth"
	"log/slog"
	"time"

	ssov1 "github.com/AlexeySHA256/protos/gen/go/sso"
	grpclogging "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type Client struct {
	api   ssov1.AuthClient
	log   *slog.Logger
	appId int32
}

func checkServiceHealth(conn *grpc.ClientConn, serviceName string) error {
	client := healthpb.NewHealthClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	response, err := client.Check(ctx, &healthpb.HealthCheckRequest{Service: serviceName})
	if err != nil {
		return err
	}

	if response.GetStatus() != healthpb.HealthCheckResponse_SERVING {
		return fmt.Errorf("service %s is not in SERVING state: %v", serviceName, response.GetStatus())
	}
	return nil
}

/*
	New creates a new Client instance.

It takes a context, a logger, an address of the gRPC server, a timeout for retry call,
and a retries count as parameters.
Returns a Client instance and an error.
*/
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
	conn, err := grpc.NewClient(
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
	if err = checkServiceHealth(conn, ""); err != nil {
		log.Error("Failed to check service health", "errMsg", err.Error())
		return nil, err
	}
	return &Client{
		api:   ssov1.NewAuthClient(conn),
		log:   log,
		appId: appId,
	}, nil
}

func (c *Client) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "grpc.Client.GetUser"
	log := c.log.With("op", op)
	resp, err := c.api.IsAdmin(ctx, &ssov1.IsAdminRequest{UserId: userID})
	if err != nil {
		log.Error("Error", "errMsg", err.Error())
		return resp.GetIsAdmin(), err
	}
	return resp.GetIsAdmin(), nil
}

func (c *Client) Login(ctx context.Context, email, password string) (*auth.TokensDTO, error) {
	const op = "grpc.Client.GetUser"
	log := c.log.With("op", op)
	resp, err := c.api.Login(ctx, &ssov1.LoginRequest{Email: email, Password: password, AppId: c.appId})
	if err != nil {
		log.Error("Error", "errMsg", err.Error())
		return nil, err
	}
	return &auth.TokensDTO{AccessToken: resp.GetAccessToken(), RefreshToken: resp.GetRefreshToken()}, nil
}

func (c *Client) Register(ctx context.Context, email, username, password string) (*auth.SignupData, error) {
	const op = "grpc.Client.GetUser"
	log := c.log.With("op", op)
	resp, err := c.api.Register(
		ctx,
		&ssov1.RegisterRequest{Email: email, Password: password, Username: username, AppId: c.appId},
	)
	if err != nil {
		log.Error("Error", "errMsg", err.Error())
		return nil, err
	}
	return &auth.SignupData{UserID: resp.GetUserId(), ActivationToken: resp.GetActivationToken()}, nil
}

func (c *Client) GetUser(ctx context.Context, params auth.GetUserParams) (*models.User, error) {
	const op = "grpc.Client.GetUser"
	log := c.log.With("op", op)
	resp, err := c.api.GetUser(ctx, &ssov1.GetUserRequest{Id: params.ID, Email: params.Email, IsActive: params.IsActive})
	if err != nil {
		grpcErr, ok := status.FromError(err)
		if ok {
			switch grpcErr.Code() {
			case codes.NotFound:
				return nil, auth.ErrUserNotFound
			case codes.InvalidArgument:
				return nil, auth.ErrInvalidData.SetMessage(grpcErr.Message())
			}
		}
		log.Error("Error", "errMsg", err.Error())
		return nil, err
	}
	user := resp.GetUser()
	const timeParseLayout = "2006-01-02 15:04:05.999999 -0700 MST"
	createdAt, err := time.Parse(timeParseLayout, resp.GetUser().GetCreatedAt())
	if err != nil {
		return nil, err
	}
	updatedAt, err := time.Parse(timeParseLayout, resp.GetUser().GetUpdatedAt())
	if err != nil {
		return nil, err
	}
	return &models.User{
		ID: user.GetId(), Email: user.GetEmail(),
		Username: user.GetUsername(), IsActive: user.GetIsActive(),
		CreatedAt: createdAt, UpdatedAt: updatedAt,
	}, nil
}

func (c *Client) ActivateUser(ctx context.Context, plainToken string) (*models.User, error) {
	const op = "grpc.Client.ActivateUser"
	log := c.log.With("op", op)
	resp, err := c.api.ActivateUser(ctx, &ssov1.ActivateUserRequest{ActivationToken: plainToken})
	if err != nil {
		grpcErr, ok := status.FromError(err)
		if ok {
			switch grpcErr.Code() {
			case codes.NotFound:
				return nil, auth.ErrUserNotFound
			case codes.InvalidArgument:
				return nil, auth.ErrInvalidData.SetMessage(grpcErr.Message())
			case codes.AlreadyExists:
				return nil, auth.ErrUserAlreadyActivated
			}
		}
		log.Error("Error", "errMsg", err.Error())
		return nil, err
	}
	user := resp.GetUser()
	const timeParseLayout = "2006-01-02 15:04:05.999999 -0700 MST"
	updatedAt, err := time.Parse(timeParseLayout, user.GetUpdatedAt())
	if err != nil {
		return nil, err
	}
	createdAt, err := time.Parse(timeParseLayout, user.GetCreatedAt())
	if err != nil {
		return nil, err
	}
	return &models.User{
		ID:        user.GetId(),
		Email:     user.GetEmail(),
		Username:  user.GetUsername(),
		Role:      user.GetRole(),
		IsActive:  user.GetIsActive(),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func (c *Client) NewActivationToken(ctx context.Context, email string) (string, error) {
	const op = "grpc.Client.NewActivationToken"
	log := c.log.With("op", op)
	resp, err := c.api.NewActivationToken(ctx, &ssov1.NewActivationTokenRequest{Email: email})
	if err != nil {
		log.Error("Error", "errMsg", err.Error())
		return "", err
	}
	return resp.GetActivationToken(), nil
}

func (c *Client) VerifyToken(ctx context.Context, token string) (bool, error) {
	const op = "grpc.Client.VerifyToken"
	log := c.log.With("op", op)
	resp, err := c.api.VerifyToken(ctx, &ssov1.VerifyTokenRequest{Token: token, AppId: c.appId})
	if err != nil {
		log.Error("Error", "errMsg", err.Error())
		return false, err
	}
	return resp.GetIsValid(), nil
}

func (c *Client) CreatePermission(ctx context.Context, code string) error {
	const op = "grpc.Client.CreatePermission"
	log := c.log.With("op", op)
	_, err := c.api.CreatePermission(ctx, &ssov1.CreatePermissionRequest{Code: code})
	if err != nil {
		log.Error("Error", "errMsg", err.Error())
		return err
	}
	return nil
}

func (c *Client) CheckPermission(ctx context.Context, permissionCode string, userID int64) (bool, error) {
	const op = "grpc.Client.CheckPermission"
	log := c.log.With("op", op)
	resp, err := c.api.CheckPermission(ctx, &ssov1.CheckPermissionRequest{UserId: userID, PermissionCode: permissionCode})
	if err != nil {
		log.Error("Error", "errMsg", err.Error())
		return false, err
	}
	return resp.GetHasPermission(), nil
}

func (c *Client) GrantPermissions(ctx context.Context, userID int64, permissions []string) error {
	const op = "grpc.Client.GrantPermissions"
	log := c.log.With("op", op)
	_, err := c.api.GrantPermissions(ctx, &ssov1.GrantPermissionsRequest{UserId: userID, PermissionCodes: permissions})
	if err != nil {
		log.Error("Error", "errMsg", err.Error())
		return err
	}
	return nil
}

// Adapter for grpclogging.Logger used to adapt it to slog.Logger
func InterceptorLogger(log *slog.Logger) grpclogging.Logger {
	return grpclogging.LoggerFunc(
		func(ctx context.Context, level grpclogging.Level, msg string, fields ...any) {
			log.Log(ctx, slog.Level(level), msg, fields...)
		},
	)
}
