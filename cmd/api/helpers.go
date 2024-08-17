package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc/status"
)

func (app *Application) extractIDParam(w http.ResponseWriter, r *http.Request) (id int, extracted bool) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		app.Http.BadRequest(w, r, "invalid movie ID")
		return 0, false
	}
	if id < 1 {
		app.Http.BadRequest(w, r, "id must be greater than zero")
		return 0, false
	}
	return id, true
}

func (app *Application) handlegRPCError(w http.ResponseWriter, r *http.Request, grpcErr *status.Status, status int) {
	app.log.Info("Sso login response msg not empty", "raw message", grpcErr.Message())
	parsedErrors := make(map[string]string)
	if err := json.Unmarshal([]byte(grpcErr.Message()), &parsedErrors); err != nil {
		app.log.Error("Error decoding grpc error message", "errMsg", err.Error())
		app.Http.ServerError(w, r, err, "")
		return
	}
	app.Http.Response(w, r, envelop{"errors": parsedErrors}, "", status)
}

func (app *Application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	maxBytes := 1_048_576 // 1MB
	src := http.MaxBytesReader(w, r.Body, int64(maxBytes))
	defer io.Copy(io.Discard, src)
	dec := json.NewDecoder(src)
	dec.DisallowUnknownFields()
	err := dec.Decode(dst)
	if err != nil {
		return handleJsonErr(err)
	}
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value") 
	}

	return nil
}

func handleJsonErr(err error) error {
	var syntaxError *json.SyntaxError
	var unmarshalTypeError *json.UnmarshalTypeError
	var invalidUnmarshalError *json.InvalidUnmarshalError
	switch {
	case errors.As(err, &syntaxError):
		return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

	case errors.Is(err, io.ErrUnexpectedEOF):
		return errors.New("body contains badly-formed JSON")

	case errors.As(err, &unmarshalTypeError):
		if unmarshalTypeError.Field != "" {
			return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
		}
		return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

	case errors.Is(err, io.EOF):
		return errors.New("body must not be empty")

	case errors.As(err, &invalidUnmarshalError):
		panic(err)
	default:
		return err
	}
}

// func (app *Application) readJsonOrBadRequest(w http.ResponseWriter, r *http.Request, data any) {
// 	if err := app.decodeJSON(r.Body, data); err != nil {
// 		app.Http.BadRequest(w, r, err.Error())
// 		return
// 	}
// }
