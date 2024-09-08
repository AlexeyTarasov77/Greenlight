package main

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)


func (app *Application) Recoverer(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil && err != http.ErrAbortHandler {
				app.Http.ServerError(w, r, err.(error), "")
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func (app *Application) RateLimiter(next http.Handler) http.Handler {
	const op = "middlewares.RateLimiter"
	log := app.log.With("op", op)
	type client struct {
		limiter *rate.Limiter
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
					limiter: rate.NewLimiter(rate.Limit(app.cfg.Limiter.Rps), app.cfg.Limiter.Burst),
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