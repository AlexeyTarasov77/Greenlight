package main

import (
	"context"
	"errors"
	"expvar"
	"greenlight/proj/internal/domain/models"
	"greenlight/proj/internal/services/auth"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
)

func (app *Application) Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil && err != http.ErrAbortHandler {
				app.log.Info("panic recovered", "err", err)
				if _, ok := err.(error); !ok {
					app.log.Error("Invalid error from panic", "err", err)
					app.Http.ServerError(w, r, errors.New("internal server error"), "")
				} else {
					app.Http.ServerError(w, r, err.(error), "")
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *Application) rateLimiter(next http.Handler) http.Handler {
	const op = "middlewares.RateLimiter"
	log := app.log.With("op", op)
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	clients := make(map[string]*client)
	var mu sync.Mutex
	go func() {
		log.Debug("Starting rate limiter cleanup goroutine")
		for {
			mu.Lock()
			for ip, client := range clients {
				if client == nil {
					continue
				}
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
			ip := realip.FromRequest(r)
			if _, ok := clients[ip]; !ok {
				newClient := &client{
					limiter:  rate.NewLimiter(rate.Limit(app.cfg.Limiter.Rps), app.cfg.Limiter.Burst),
					lastSeen: time.Now(),
				}
				mu.Lock()
				clients[ip] = newClient
				mu.Unlock()
				log.Debug("new client", "ip", ip)
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

func (app *Application) enableCORS(allowedOrigins []string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Vary", "Origin") // marks, that response depends on Origin header
			w.Header().Set("Vary", "Access-Control-Request-Method")
			origin := r.Header.Get("Origin")
			if origin != "" {
				// Comparing origin from req with allowed origins, because it's impossible to set multiple origins
				// in Access-Control-Allow-Origin header. So when matched origin is found, duplicating it into header
				for _, allowedOrigin := range allowedOrigins {
					if allowedOrigin == "*" || allowedOrigin == origin {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						if (r.Method == http.MethodOptions) && r.Header.Get("Access-Control-Request-Method") != "" {
							// Identified as a preflight request
							w.Header().Set("Access-Control-Allow-Methods", "PUT, PATCH, DELETE, OPTIONS")
							w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
							w.WriteHeader(http.StatusOK)
							return
						}
						if allowedOrigin == "*" {
							app.log.Warn("Be carefull, your service can be vulnerable to a distributed brute-force attack, if using '*' as allowed origin in conjuction with allowing Authorization header in cors requests")
						}
						break
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (app *Application) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Vary", "Authorization")

		var user *models.User = models.AnonymousUser

		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			app.log.Debug("Has auth header", "header", authHeader)
			const bearerLength = len("Bearer ")
			if !strings.HasPrefix(authHeader, "Bearer ") || len(authHeader) < bearerLength+1 {
				app.log.Warn("Invalid auth header", "header", authHeader)
				app.Http.BadRequest(w, r, "Invalid Authorization header, should have format: 'Bearer <token>'")
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
				app.Http.InvalidAuthToken(w, r)
				return
			}
			parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
				return []byte(app.cfg.AppSecret), nil
			})
			if err != nil {
				app.log.Warn("Failed to parse token", "error", err)
				app.Http.InvalidAuthToken(w, r)
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
							app.Http.InvalidAuthToken(w, r)
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

// Depends on Authenticate middleware
func (app *Application) requireAuthenticatedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.log.Debug("Checking authentication")
		user := app.Http.ContextGetUser(r)
		app.log.Debug("Got user from context", "user", user)
		if user.IsAnonymous() {
			app.Http.Unauthorized(w, r, "you must be authenticated to access this resource")
			return
		}
		app.log.Debug("User is authenticated")
		next.ServeHTTP(w, r)
	})
}

func (app *Application) requireActivatedUser(next http.Handler) http.Handler {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.log.Debug("Checking activated user")
		user := app.Http.ContextGetUser(r)
		app.log.Debug("Got user from context", "user", user)
		if !user.IsActive {
			app.Http.Unauthorized(w, r, "you must activate your account to access this resource")
			return
		}
		app.log.Debug("User is activated")
		next.ServeHTTP(w, r)
	})
	return app.requireAuthenticatedUser(fn)
}

func (app *Application) requirePermission(permissionCode string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			app.log.Debug("Checking permission", "permissionCode", permissionCode)
			user := app.Http.ContextGetUser(r)
			app.log.Debug("Got user from context", "user", user)
			hasPermission, err := app.Services.Auth.CheckPermission(r.Context(), permissionCode, user.ID)
			if err != nil {
				app.Http.ServerError(w, r, err, "")
				return
			}
			app.log.Debug("Has permission", "hasPermission", hasPermission)
			if !hasPermission {
				app.Http.Forbidden(w, r, "you don't have permission to access this resource")
				return
			}
			app.log.Debug("Calling next handler")
			next.ServeHTTP(w, r)
		})
		return app.requireActivatedUser(fn)
	}
}

func (app *Application) metrics(next http.Handler) http.Handler {
	totalRequestsReceived := expvar.NewInt("total_requests_received")
	totalResponsesSent := expvar.NewInt("total_responses_sent")
	totalResponsesSentByStatus := expvar.NewMap("total_responses_sent_by_status")
	totalProcessingTimeμs := expvar.NewInt("total_processing_time_μs")
	avgProcessingTimeμs := expvar.NewInt("avg_processing_time_μs")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		totalRequestsReceived.Add(1)
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		totalResponsesSent.Add(1)
		totalProcessingTimeμs.Add(int64(time.Since(start).Microseconds()))
		avgProcessingTimeμs.Set(totalProcessingTimeμs.Value() / totalResponsesSent.Value())
		totalResponsesSentByStatus.Add(strconv.Itoa(ww.Status()), 1)
	})
}
