package main

import (
	"context"
	"greenlight/proj/internal/domain/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)


func TestRequiredAuthenticatedUser(t *testing.T) {
	app := NewTestApplication(nil, t)
	t.Run("authenticated", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request = request.WithContext(context.WithValue(request.Context(), CtxKeyUser, &models.User{
			ID: 1,
			Username: "test",
			Email: "test@gmail.com",
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
		ID: 1,
		Username: "test",
		Email: "test@gmail.com",
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