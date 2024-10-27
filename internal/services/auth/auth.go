package auth

import (
	"context"
	"greenlight/proj/internal/domain/models"
	"html/template"
	"log/slog"
)

//go:generate mockery --name=MailProvider
type MailProvider interface {
	Send(recipient string, tmplName string, tmplData any) error
}

type TokensDTO struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type GetUserParams struct {
	ID       int64
	Email    string
	IsActive bool
}

//go:generate mockery --name=SsoProvider
type SsoProvider interface {
	Register(ctx context.Context, email, username, password string) (*SignupData, error)
	Login(ctx context.Context, email, password string) (*TokensDTO, error)
	GetUser(ctx context.Context, params GetUserParams) (*models.User, error)
	ActivateUser(ctx context.Context, plainToken string) (*models.User, error)
	NewActivationToken(ctx context.Context, email string) (string, error)
	VerifyToken(ctx context.Context, token string) (bool, error)
	CheckPermission(ctx context.Context, permissionCode string, userID int64) (bool, error)
}

//go:generate mockery --name=TaskExecutor
type TaskExecutor interface {
	Add(task func())
}

type AuthService struct {
	log          *slog.Logger
	Mailer       MailProvider
	sso          SsoProvider
	taskExecutor TaskExecutor
}

func New(
	log *slog.Logger,
	mailer MailProvider,
	ssoProvider SsoProvider,
	taskExecutor TaskExecutor,
) *AuthService {
	return &AuthService{
		log:          log,
		Mailer:       mailer,
		sso:          ssoProvider,
		taskExecutor: taskExecutor,
	}
}

type SignupData struct {
	UserID          int64
	ActivationToken string
}

type activationEmailData struct {
	activationURL   string
	username        string
	userID          int64
	activationToken string
}

func (a *AuthService) sendActivationEmail(email string, data activationEmailData) {
	a.log.Info("sending activation email")
	err := a.Mailer.Send(
		email,
		"user_welcome.html",
		map[string]interface{}{
			"activationURL":   template.URL(data.activationURL),
			"username":        data.username,
			"userID":          data.userID,
			"activationToken": data.activationToken,
		})
	if err != nil {
		a.log.Error("Error sending activation email", "errMsg", err.Error())
	}
}

func (a *AuthService) Signup(ctx context.Context, email, username, password, activationURL string) (int64, error) {
	const op = "auth.AuthService.Signup"
	log := a.log.With("op", op, "email", email)
	data, err := a.sso.Register(ctx, email, username, password)
	if err != nil {
		log.Error("Error calling Sso.Register", "errMsg", err.Error())
		return 0, err
	}
	a.taskExecutor.Add(func() {
		a.sendActivationEmail(email, activationEmailData{
			activationURL:   activationURL,
			username:        username,
			userID:          data.UserID,
			activationToken: data.ActivationToken,
		})
	})
	return data.UserID, nil
}

func (a *AuthService) Login(ctx context.Context, email, password string) (*TokensDTO, error) {
	const op = "auth.AuthService.Login"
	log := a.log.With("op", op, "email", email)
	resp, err := a.sso.Login(ctx, email, password)
	if err != nil {
		log.Error("Error calling Sso.Login", "errMsg", err.Error())
		return nil, err
	}
	return resp, nil
}

func (a *AuthService) GetNewActivationToken(ctx context.Context, email string, activationURL string) error {
	user, err := a.sso.GetUser(ctx, GetUserParams{Email: email})
	if err != nil {
		return err
	}
	newToken, err := a.sso.NewActivationToken(ctx, user.Email)
	if err != nil {
		return err
	}
	a.taskExecutor.Add(func() {
		a.sendActivationEmail(user.Email, activationEmailData{
			activationURL:   activationURL,
			username:        user.Username,
			userID:          user.ID,
			activationToken: newToken,
		})
	})
	return nil
}

func (a *AuthService) ActivateUser(ctx context.Context, plainToken string) (*models.User, error) {
	user, err := a.sso.ActivateUser(ctx, plainToken)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (a *AuthService) VerifyToken(ctx context.Context, token string) (bool, error) {
	return a.sso.VerifyToken(ctx, token)
}

func (a *AuthService) GetUser(ctx context.Context, params GetUserParams) (*models.User, error) {
	return a.sso.GetUser(ctx, params)
}

func (a *AuthService) CheckPermission(ctx context.Context, permissionCode string, userID int64) (bool, error) {
	return a.sso.CheckPermission(ctx, permissionCode, userID)
}