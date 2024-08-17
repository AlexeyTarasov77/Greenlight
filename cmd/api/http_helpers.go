package main

import (
	"greenlight/proj/internal/config"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Http struct {
	log *slog.Logger
	cfg *config.Config
}

type envelop map[string]any

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    envelop   `json:"data,omitempty"`
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
	status := http.StatusOK
	render.Status(r, status)
	render.JSON(w, r, h.NewResponse(data, msg, status))
}

func (h *Http) Created(w http.ResponseWriter, r *http.Request, data envelop, msg string) {
	status := http.StatusCreated
	render.Status(r, status)
	render.JSON(w, r, h.NewResponse(data, msg, status))
}

func (h *Http) NoContent(w http.ResponseWriter, r *http.Request, msg string) {
	status := http.StatusNoContent
	render.Status(r, status)
	render.JSON(w, r, h.NewResponse(nil, msg, status))
}

func (h *Http) BadRequest(w http.ResponseWriter, r *http.Request, msg string) {
	status := http.StatusBadRequest
	render.Status(r, status)
	render.JSON(w, r, h.NewResponse(nil, msg, status))
}

func (h *Http) Conflict(w http.ResponseWriter, r *http.Request, msg string) {
	status := http.StatusConflict
	render.Status(r, status)
	render.JSON(w, r, h.NewResponse(nil, msg, status))
}

func (h *Http) UnprocessableEntity(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	status := http.StatusUnprocessableEntity
	render.Status(r, status)
	render.JSON(w, r, h.NewResponse(envelop{
		"errors": errors,
	}, "", status))
}

func (h *Http) NotFound(w http.ResponseWriter, r *http.Request, msg string) {
	status := http.StatusNotFound
	render.Status(r, status)
	render.JSON(w, r, h.NewResponse(nil, msg, status))
}

func (h *Http) ServerError(w http.ResponseWriter, r *http.Request, err error, msg string) {
	status := http.StatusInternalServerError
	log := h.setupLogPerReq(r)
	if err != nil {
		log.Error(err.Error())
	}
	render.Status(r, status)
	msg = processMsg(status, msg)
	if h.cfg.Debug {
		msg = msg + "\n" + string(debug.Stack())
		w.WriteHeader(status)
		w.Write([]byte(msg))
		return
	}
	render.JSON(w, r, Response{Success: false, Message: msg})
}
