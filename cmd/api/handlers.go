package main

import (
	"errors"
	"fmt"
	"greenlight/proj/internal/domain/fields"
	"greenlight/proj/internal/lib/validator"
	"greenlight/proj/internal/services/movies"
	"math"
	"net/http"

	"github.com/go-chi/render"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/status"
)

func (app *Application) healthcheck(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]any{
		"status":  "available",
		"debug":   app.cfg.Debug,
		"version": version,
	})
}

func (app *Application) getMovie(w http.ResponseWriter, r *http.Request) {
	id, extracted := app.extractIDParam(w, r)
	if !extracted {
		return
	}
	movie, err := app.movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, movies.ErrMovieNotFound):
			app.Http.NotFound(w, r, "")
		default:
			app.Http.ServerError(w, r, fmt.Errorf("error during retrieving movie from db: %w", err), "")
		}
		return
	}
	app.Http.Ok(w, r, envelop{"movie": movie}, "")
}

func (app *Application) getMovies(w http.ResponseWriter, r *http.Request) {
	type queryParams struct {
		Sort  string `validate:"omitempty,sortbymoviefield" schema:"sort,default:-id"`
		PageSize int `validate:"omitempty,min=1,max=100" schema:"page_size,default:20"`
		Page int `validate:"omitempty,min=1,max=10000000" schema:"page,default:1"`
		Title string `validate:"omitempty,max=255"`
		Year int32 `validate:"omitempty,min=1888,max=2100"`
		Genres []string `validate:"omitempty,min=1,max=5,unique" schema:"genres"`
	}
	app.validator.RegisterValidation("sortbymoviefield", validator.ValidateSortByMovieField)
	var params queryParams
	qs := r.URL.Query()
	// TODO: написать свой декодер
	// params.Limit = app.readInt(qs, "limit", limitUnset)
	// params.Sort = app.readString(qs, "sort", "")
	// params.Page = app.readInt(qs, "page", 1)
	// params.PageSize = app.readInt(qs, "page_size", 20)
	if err := app.Decoder.Decode(&params, qs); err != nil {
		app.log.Error("Error during decoding query params", "msg", err.Error())
		app.Http.BadRequest(w, r, "Invalid query params provided. Ensure that all query params are valid")
		return
	}
	app.log.Info("query params values", "params", params)
	if validationErrs := validator.ValidateStruct(app.validator, params); len(validationErrs) > 0 {
		app.Http.UnprocessableEntity(w, r, validationErrs)
		return
	}
	if params.Genres == nil {
		params.Genres = []string{}
	}
	movies, totalRecords, err := app.movies.List(
		params.Title,
		params.Genres,
		params.Page,
		params.PageSize,
		params.Sort,
	)
	if err != nil {
		app.Http.ServerError(w, r, err, "")
		return
	}
	app.Http.Ok(
		w, r,
		envelop{
			"total_on_page": len(movies),
			"current_page": params.Page,
			"page_size": params.PageSize,
			"total_records": totalRecords,
			"first_page": 1,
			"last_page": math.Ceil(float64(totalRecords) / float64(params.PageSize)),
			"movies": movies,
		}, "",
	)
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
	w.Header().Set("Location", fmt.Sprintf("/v1/movies/%d", createdMovie.ID))
	app.Http.Created(w, r, envelop{"movie": createdMovie}, "Movie successfully created")
}

func (app *Application) updateMovie(w http.ResponseWriter, r *http.Request) {
	id, extracted := app.extractIDParam(w, r)
	if !extracted {
		return
	}
	type request struct {
		Title   *string              `validate:"omitempty,max=255,min=1"`
		Year    *int32               `validate:"omitempty,min=1888,max=2100"`
		Runtime *fields.MovieRuntime `validate:"omitempty,gt=0"`
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
		switch {
		case errors.Is(err, movies.ErrMovieNotFound):
			app.Http.NotFound(w, r, err.Error())
		case errors.Is(err, movies.ErrNoArgumentsChanged):
			app.Http.BadRequest(w, r, err.Error())
		case errors.Is(err, movies.ErrMovieAlreadyExists) || errors.Is(err, movies.ErrEditConflict):
			app.Http.Conflict(w, r, err.Error())
		default:
			app.Http.ServerError(w, r, err, "")
		}
		return
	}
	app.Http.Ok(w, r, envelop{"movie": updatedMovie}, "Movie successfully updated")
}

func (app *Application) deleteMovie(w http.ResponseWriter, r *http.Request) {
	id, extracted := app.extractIDParam(w, r)
	if !extracted {
		return
	}
	err := app.movies.Delete(id)
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

// AUTH

func (app *Application) login(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Email    string `validate:"required,email"`
		Password string `validate:"required,min=8"`
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
	tokens, err := app.Sso.Login(r.Context(), req.Email, req.Password)
	grpcErr, ok := status.FromError(err)
	httpRespCode := runtime.HTTPStatusFromCode(grpcErr.Code())
	if grpcErr.Message() != "" {
		app.handleGRPCError(w, r, grpcErr, httpRespCode)
		return
	}
	if ok {
		app.Http.Response(w, r, envelop{"tokens": tokens}, "", httpRespCode)
		return
	}
	app.Http.ServerError(w, r, err, "")
}

func (app *Application) signup(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Username string `validate:"required,max=50,alphanum"`
		Email    string `validate:"required,email"`
		Password string `validate:"required,min=8"`
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
	id, err := app.Sso.Register(r.Context(), req.Email, req.Password)
	grpcErr, ok := status.FromError(err)
	httpRespCode := runtime.HTTPStatusFromCode(grpcErr.Code())
	if grpcErr.Message() != "" {
		app.handleGRPCError(w, r, grpcErr, httpRespCode)
		return
	}
	if ok {
		app.Http.Created(w, r, envelop{"id": id}, "User successfully created")
		return
	}
	app.Http.ServerError(w, r, err, "")
}
