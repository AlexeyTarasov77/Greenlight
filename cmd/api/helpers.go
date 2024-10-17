package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"greenlight/proj/internal/lib/validator"
	"io"
	"net/http"
	"reflect"
	"google.golang.org/grpc/status"
)

func (app *Application) readReqBodyAndValidate(w http.ResponseWriter, r *http.Request, dst any) (success bool) {
	dstV := reflect.ValueOf(dst)
	if dstV.Kind() != reflect.Ptr || dstV.Elem().Kind() != reflect.Struct {
		panic("api.helpers.readReqBodyAndValidate: dst must be a pointer to a struct")
	}
	if err := app.readJSON(w, r, dst); err != nil {
		app.Http.BadRequest(w, r, err.Error())
		return
	}
	if validationErrs := validator.ValidateStruct(app.validator, dst); len(validationErrs) > 0 {
		app.Http.UnprocessableEntity(w, r, validationErrs)
		return
	}
	return true
}

func (app *Application) handleGRPCError(w http.ResponseWriter, r *http.Request, grpcErr *status.Status, status int) {
	app.log.Debug("GRPC error", "raw message", grpcErr.Message())
	parsedErrors := make(map[string]string)
	if err := json.Unmarshal([]byte(grpcErr.Message()), &parsedErrors); err != nil {
		app.log.Error("Error decoding grpc error message", "err", err, "msg", grpcErr.Message())
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
		return parseJsonErr(err)
	}
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

func parseJsonErr(err error) error {
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