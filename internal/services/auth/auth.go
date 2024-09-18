package auth

import (
	"context"
	"html/template"
	"log/slog"
)

type MailProvider interface {
	Send(recipient string, tmplName string, tmplData any) error
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type SsoProvider interface {
	Register(ctx context.Context, email, username, password string) (*SignupData, error)
	Login(ctx context.Context, email, password string) (*Tokens, error)
	// ActivateUser(ctx context.Context, token string) (bool, error)
}

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

func (a *AuthService) Signup(ctx context.Context, email, username, password, activationLink string) (int64, error) {
	const op = "auth.AuthService.Signup"
	log := a.log.With("op", op, "email", email)
	data, err := a.sso.Register(ctx, email, username, password)
	if err != nil {
		log.Error("Error calling Sso.Register", "errMsg", err.Error())
		return 0, err
	}
	a.taskExecutor.Add(func() {
		log.Info("sending mail")
		err = a.Mailer.Send(
			email,
			"user_welcome.html",
			map[string]interface{}{
				"activationLink":  template.URL(activationLink),
				"username":        username,
				"userID":          data.UserID,
				"activationToken": data.ActivationToken,
			})
		if err != nil {
			log.Error("Error sending activation email", "errMsg", err.Error())
		}
	})
	return data.UserID, nil
}

func (a *AuthService) Login(ctx context.Context, email, password string) (*Tokens, error) {
	const op = "auth.AuthService.Login"
	log := a.log.With("op", op, "email", email)
	resp, err := a.sso.Login(ctx, email, password)
	if err != nil {
		log.Error("Error calling Sso.Login", "errMsg", err.Error())
		return nil, err
	}
	return resp, nil
}