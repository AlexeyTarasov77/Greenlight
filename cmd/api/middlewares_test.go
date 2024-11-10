package main

import (
	"context"
	"greenlight/proj/internal/config"
	"greenlight/proj/internal/domain/models"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCors(t *testing.T) {
	const testOrigin = "http://testorigin.com"
	app := NewTestApplication(nil, t)
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	testCases := []struct {
		name                  string
		allowedOrigins        []string
		originHeader          string
		expectedAllowedOrigin string
	}{
		{
			name:                  "no origin header",
			allowedOrigins:        []string{testOrigin},
			originHeader:          "",
			expectedAllowedOrigin: "",
		},
		{
			name:                  "single allowed origin",
			allowedOrigins:        []string{testOrigin},
			originHeader:          testOrigin,
			expectedAllowedOrigin: testOrigin,
		},
		{
			name:                  "multiple allowed origins",
			allowedOrigins:        []string{"http://localhost:3000", testOrigin, "http://localhost:4000"},
			originHeader:          testOrigin,
			expectedAllowedOrigin: testOrigin,
		},
		{
			name:                  "not allowed origin",
			allowedOrigins:        []string{testOrigin},
			originHeader:          "http://unknown-origin.com",
			expectedAllowedOrigin: "",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			request.Header.Set("Origin", testCase.originHeader)
			recorder := httptest.NewRecorder()
			app.enableCORS(testCase.allowedOrigins)(next).ServeHTTP(recorder, request)
			assert.Equal(t, testCase.expectedAllowedOrigin, recorder.Header().Get("Access-Control-Allow-Origin"))
		})
	}
	t.Run("preflight request", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodOptions, "/", nil)
		request.Header.Set("Origin", testOrigin)
		request.Header.Set("Access-Control-Request-Method", "GET")
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		app.enableCORS([]string{testOrigin})(next).ServeHTTP(recorder, request)
		assert.NotEmpty(t, recorder.Header().Get("Access-Control-Allow-Methods"))
		assert.NotEmpty(t, recorder.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, recorder.Code, http.StatusOK)
	})
}

func TestRequiredAuthenticatedUser(t *testing.T) {
	app := NewTestApplication(nil, t)
	t.Run("authenticated", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request = request.WithContext(context.WithValue(request.Context(), CtxKeyUser, &models.User{
			ID:       1,
			Username: "test",
			Email:    "test@gmail.com",
		}))
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		app.requireAuthenticatedUser(next).ServeHTTP(recorder, request)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
	t.Run("anonymous", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request = request.WithContext(context.WithValue(request.Context(), CtxKeyUser, models.AnonymousUser))
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		app.requireAuthenticatedUser(next).ServeHTTP(recorder, request)
		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})
}

func TestRequireActivatedUser(t *testing.T) {
	app := NewTestApplication(nil, t)
	testUser := &models.User{
		ID:       1,
		Username: "test",
		Email:    "test@gmail.com",
		IsActive: true,
	}
	t.Run("Activated", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request = request.WithContext(context.WithValue(request.Context(), CtxKeyUser, testUser))
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		app.requireActivatedUser(next).ServeHTTP(recorder, request)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
	t.Run("NotActivated", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		testUser.IsActive = false
		request = request.WithContext(context.WithValue(request.Context(), CtxKeyUser, testUser))
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		app.requireActivatedUser(next).ServeHTTP(recorder, request)
		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})
}

func TestRateLimiter(t *testing.T) {
	app := NewTestApplication(&config.Config{Limiter: config.Limiter{Enabled: true, Rps: 2, Burst: 4}}, t)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	limiter := app.rateLimiter(next)
	t.Run("success", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		limiter.ServeHTTP(recorder, request)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
	time.Sleep(time.Second)
	t.Run("exceeded", func(t *testing.T) {
		exceeded := int(app.cfg.Limiter.Burst) + 1
		for i := 1; i <= exceeded; i++ {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			limiter.ServeHTTP(recorder, request)
			expectedStatus := http.StatusOK
			if i == exceeded {
				expectedStatus = http.StatusTooManyRequests
			}
			assert.Equal(t, expectedStatus, recorder.Code)
		}
	})
}