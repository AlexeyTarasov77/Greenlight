package main

import (
	"errors"
	"fmt"
	"greenlight/proj/internal/domain/fields"
	"greenlight/proj/internal/lib/validator"
	"greenlight/proj/internal/services/movies"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func (app *Application) healthcheck(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]any{
		"status":  "available",
		"debug":   app.cfg.Debug,
		"version": version,
	})
}

func (app *Application) getMovie(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		app.Http.BadRequest(w, r, "invalid movie ID")
		return
	}
	movie, err := app.movies.Get(id)
	if err != nil {
		if errors.Is(err, movies.ErrMovieNotFound) {
			http.NotFound(w, r)
		} else {
			app.Http.ServerError(w, r, fmt.Errorf("error during retrieving movie from db: %w", err), "")
		}
		return
	}
	app.Http.Ok(w, r, envelop{"movie": movie}, "")
}

func (app *Application) getMovies(w http.ResponseWriter, r *http.Request) {
	const limitUnset = -1
	limit := limitUnset
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			app.Http.BadRequest(w, r, "invalid limit param")
			return
		}
	}
	movies, err := app.movies.List(limit)
	if err != nil {
		app.Http.ServerError(w, r, err, "")
		return
	}
	app.Http.Ok(w, r, envelop{"movies": movies}, "")
}

func (app *Application) createMovie(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Title   string              `validate:"required,max=255"`
		Year    int32               `validate:"required,min=1888,max=2100"`
		Runtime fields.MovieRuntime `validate:"required,gt=0"`
		Genres  []string            `validate:"required,min=1,max=5,unique"`
	}
	var req request
	if err := app.readJSON(w, r, &req); err != nil {
		app.Http.BadRequest(w, r, err.Error())
		return
	}
	if validationErrs := validator.ValidateStruct(app.validator, req); len(validationErrs) > 0 {
		app.Http.UnprocessableEntity(w, r, validationErrs)
		return
	}
	createdMovie, err := app.movies.Create(req.Title, req.Year, req.Runtime, req.Genres)
	if err != nil {
		if errors.Is(err, movies.ErrMovieAlreadyExists) {
			app.Http.Conflict(w, r, err.Error())
			return
		}
		app.Http.ServerError(w, r, err, "")
		return
	}
	app.Http.Created(w, r, envelop{"movie": createdMovie}, "Movie successfully created")
}


func (app *Application) updateMovie(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		app.Http.BadRequest(w, r, "invalid movie ID")
		return
	}
	type request struct {
		Title   string              `validate:"omitempty,max=255"`
		Year    int32               `validate:"omitempty,min=1888,max=2100"`
		Runtime fields.MovieRuntime `validate:"omitempty,gt=0"`
		Genres  []string            `validate:"omitempty,min=1,max=5,unique"`
	}
	var req request
	if err := app.readJSON(w, r, &req); err != nil {
		app.Http.BadRequest(w, r, err.Error())
		return
	}
	if validationErrs := validator.ValidateStruct(app.validator, req); len(validationErrs) > 0 {
		app.Http.UnprocessableEntity(w, r, validationErrs)
		return
	}
	updatedMovie, err := app.movies.Update(id, req.Title, req.Year, req.Runtime, req.Genres)
	if err != nil {
		if errors.Is(err, movies.ErrMovieNotFound) {
			app.Http.NotFound(w, r, err.Error())
		} else if errors.Is(err, movies.ErrMovieAlreadyExists) {
			app.Http.Conflict(w, r, err.Error())
		} else {
			app.Http.ServerError(w, r, err, "")
		}
		return
	}
	app.Http.Ok(w, r, envelop{"movie": updatedMovie}, "Movie successfully updated")
}

func (app *Application) deleteMovie(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		app.Http.BadRequest(w, r, "invalid movie ID")
		return
	}
	err = app.movies.Delete(id)
	if err != nil {
		if errors.Is(err, movies.ErrMovieNotFound) {
			app.Http.NotFound(w, r, err.Error())
		} else {
			app.Http.ServerError(w, r, err, "")
		}
		return
	}
	app.Http.NoContent(w, r, "Movie successfully deleted")
}