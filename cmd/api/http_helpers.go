package main

import (
	"errors"
	"greenlight/proj/internal/config"
	"greenlight/proj/internal/domain/models"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Http struct {
	log *slog.Logger
	cfg *config.Config
}

type envelop map[string]any

type Response struct {
	Success bool    `json:"success"`
	Message string  `json:"message,omitempty"`
	Data    envelop `json:"data,omitempty"`
}

func processMsg(status int, msg string) string {
	if msg == "" {
		msg = http.StatusText(status)
	}
	return msg
}

func (h *Http) setupLogPerReq(r *http.Request) *slog.Logger {
	return h.log.With(
		"request_id",
		middleware.GetReqID(r.Context()),
		"method",
		r.Method,
		"path",
		r.URL.Path,
	)
}

func (h *Http) NewResponse(data envelop, msg string, status int) *Response {
	msg = processMsg(status, msg)
	success := status >= 200 && status < 400
	return &Response{Success: success, Message: msg, Data: data}
}

func (h *Http) Response(w http.ResponseWriter, r *http.Request, data envelop, msg string, status int) {
	render.Status(r, status)
	render.JSON(w, r, h.NewResponse(data, msg, status))
}

func (h *Http) Ok(w http.ResponseWriter, r *http.Request, data envelop, msg string) {
	h.Response(w, r, data, msg, http.StatusOK)
}

func (h *Http) Created(w http.ResponseWriter, r *http.Request, data envelop, msg string) {
	h.Response(w, r, data, msg, http.StatusCreated)
}

func (h *Http) NoContent(w http.ResponseWriter, r *http.Request, msg string) {
	h.Response(w, r, nil, msg, http.StatusNoContent)
}

func (h *Http) BadRequest(w http.ResponseWriter, r *http.Request, msg string) {
	h.Response(w, r, nil, msg, http.StatusBadRequest)
}

func (h *Http) Unauthorized(w http.ResponseWriter, r *http.Request, msg string) {
	h.Response(w, r, nil, msg, http.StatusUnauthorized)
}

func (h *Http) Forbidden(w http.ResponseWriter, r *http.Request, msg string) {
	h.Response(w, r, nil, msg, http.StatusForbidden)
}

func (h *Http) InvalidAuthToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	h.Unauthorized(w, r, "Invalid or expired authentication token in authorization header")
}

func (h *Http) Conflict(w http.ResponseWriter, r *http.Request, msg string) {
	h.Response(w, r, nil, msg, http.StatusConflict)
}

func (h *Http) UnprocessableEntity(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	h.Response(w, r, envelop{"errors": errors}, "", http.StatusUnprocessableEntity)
}

func (h *Http) NotFound(w http.ResponseWriter, r *http.Request, msg string) {
	h.Response(w, r, nil, msg, http.StatusNotFound)
}

func (h *Http) ServerError(w http.ResponseWriter, r *http.Request, err error, msg string) {
	status := http.StatusInternalServerError
	defaultErrMsg := "Sorry! Can't process your request. Please try again later."
	log := h.setupLogPerReq(r)
	if err != nil {
		log.Error(err.Error())
	}
	render.Status(r, status)
	if msg == "" {
		msg = defaultErrMsg
	}
	if h.cfg.Debug {
		msg = err.Error() + "\n" + string(debug.Stack())
		w.WriteHeader(status)
		w.Write([]byte(msg))
		return
	}
	render.JSON(w, r, Response{Success: false, Message: msg})
}

func (h *Http) ContextGetUser(r *http.Request) *models.User {
	user, ok := r.Context().Value(CtxKeyUser).(*models.User)
	if !ok {
		panic(errors.New("invalid user in request context"))
	}
	return user
}

func (h *Http) extractIDParam(w http.ResponseWriter, r *http.Request) (id int, extracted bool) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.BadRequest(w, r, "invalid movie ID")
		return 0, false
	}
	if id < 1 {
		h.BadRequest(w, r, "id must be greater than zero")
		return 0, false
	}
	return id, true
}
