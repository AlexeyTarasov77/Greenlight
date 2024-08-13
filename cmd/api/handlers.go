package main

import (
	"net/http"

	"github.com/go-chi/render"
)

func (app *Application) healthcheck(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, struct {
		Status  string `json:"status"`
		Debug   bool   `json:"debug"`
		Version string `json:"version"`
	}{
		Status:  "available",
		Debug:   app.cfg.Debug,
		Version: version,
	})
}
