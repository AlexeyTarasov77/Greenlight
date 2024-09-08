package auth

import (
	"context"
	"fmt"
	"log/slog"
)

type MailProvider interface {
	Send(recipient string, tmplName string, tmplData any) error
}

type Transaction interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type TransactionManager interface {
	Begin(ctx context.Context) (Transaction, error)
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type SsoProvider interface {
	Register(ctx context.Context, email, username, password string) (int64, error)
	Login(ctx context.Context, email, password string) (*Tokens, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AuthService struct {
	log *slog.Logger
	Mailer MailProvider
	sso SsoProvider
	transactions TransactionManager
}

func New(
	log *slog.Logger,
	mailer MailProvider,
	ssoProvider SsoProvider,
	transactions TransactionManager,
) *AuthService {
	return &AuthService{
		log: log,
		Mailer: mailer,
		sso: ssoProvider,
		transactions: transactions,
	}
}

func (a *AuthService) Signup(ctx context.Context, email, username, password, activationLink string) (int64, error) {
	const op = "auth.AuthService.Signup"
	log := a.log.With("op", op, "email", email)
	userID, err := a.sso.Register(ctx, email, username, password)
	if err != nil {
		log.Error("Error calling Sso.Register", "errMsg", err.Error())
		return 0, err
	}
	activationLink = fmt.Sprintf(activationLink, userID)
	if err := a.Mailer.Send(email, "user_welcome.html", map[string]interface{}{"activationLink": activationLink, "username": username}); err != nil {
		log.Error("Error calling sending mail with activation link", "errMsg", err.Error())
		return 0, err
	}
	return userID, nil
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