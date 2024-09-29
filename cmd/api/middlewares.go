package main

import (
	"context"
	"errors"
	"greenlight/proj/internal/domain/models"
	"greenlight/proj/internal/services/auth"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/time/rate"
)

func (app *Application) Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil && err != http.ErrAbortHandler {
				app.Http.ServerError(w, r, err.(error), "")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *Application) RateLimiter(next http.Handler) http.Handler {
	const op = "middlewares.RateLimiter"
	log := app.log.With("op", op)
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	clients := make(map[string]*client)
	var mu sync.Mutex
	go func() {
		for {
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 5*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
			time.Sleep(5 * time.Minute)
		}
	}()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.cfg.Limiter.Enabled {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.Http.ServerError(w, r, err, "")
				return
			}
			if _, ok := clients[ip]; !ok {
				newClient := &client{
					limiter:  rate.NewLimiter(rate.Limit(app.cfg.Limiter.Rps), app.cfg.Limiter.Burst),
					lastSeen: time.Now(),
				}
				mu.Lock()
				clients[ip] = newClient
				mu.Unlock()
			}
			limiter := clients[ip].limiter
			log.Debug("rate limiting", "ip", ip, "Available requests", limiter.Tokens())
			if !limiter.Allow() {
				log.Warn("rate limit exceeded", "ip", ip)
				app.Http.Response(
					w, r,
					envelop{"error": "rate limit exceeded"},
					"Can't process request see an error below.",
					http.StatusTooManyRequests,
				)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// func (app *Application) LoginRequired(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		httpUnathorized := func() {
// 			app.Http.Response(w, r, nil, "Unauthorized", http.StatusUnauthorized)
// 		}
// 		const bearerLength = len("Bearer ")
// 		authHeader := r.Header.Get("Authorization")
// 		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") || len(authHeader) < bearerLength+1 {
// 			httpUnathorized()
// 			return
// 		}
// 		token := strings.TrimPrefix(authHeader, "Bearer ")

// 		isValidToken, err := app.Services.Auth.VerifyToken(r.Context(), token)
// 		if err != nil || !isValidToken {
// 			httpUnathorized()
// 			return
// 		}
// 		next.ServeHTTP(w, r)
// 	})
// }

type CtxKey string

const CtxKeyUser CtxKey = "user"

func (app *Application) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var user *models.User

		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			app.log.Debug("Has auth header", "header", authHeader)
			const bearerLength = len("Bearer ")
			if !strings.HasPrefix(authHeader, "Bearer ") || len(authHeader) < bearerLength+1 {
				app.log.Warn("Invalid auth header", "header", authHeader)
				app.Http.BadRequest(w, r, "Invalid Authorization header, should be 'Bearer <token>'")
				return
			}
			token := strings.TrimPrefix(authHeader, "Bearer ")
			isValidToken, err := app.Services.Auth.VerifyToken(r.Context(), token)
			if err != nil {
				app.log.Error("Failed to verify token", "error", err)
				app.Http.ServerError(w, r, err, "")
				return
			}
			if !isValidToken {
				app.log.Warn("Invalid or expired token", "token", token)
				app.Http.Unauthorized(w, r, "Invalid or expired token")
				return
			}
			parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
				return []byte(app.cfg.AppSecret), nil
			})
			if err != nil {
				app.log.Error("Failed to parse token", "error", err)
				app.Http.ServerError(w, r, err, "")
				return
			}

			if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
				app.log.Debug("Has claims", "claims", claims)
				userID, exists := claims["uid"].(float64)
				if exists {
					app.log.Debug("Has user id", "user_id", userID)
					user, err = app.Services.Auth.GetUser(r.Context(), auth.GetUserParams{ID: int64(userID), IsActive: true})
					if err != nil {
						switch {
						case errors.Is(err, auth.ErrUserNotFound):
							app.log.Warn("user not found", "user_id", userID)
						default:
							app.log.Error("Failed to get user", "error", err)
							app.Http.ServerError(w, r, err, "")
						}
						return
					}
				}
			}
		}
		r = r.WithContext(context.WithValue(r.Context(), CtxKeyUser, user))
		next.ServeHTTP(w, r)
	})
}
