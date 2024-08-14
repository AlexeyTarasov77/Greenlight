package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

func (app *Application) decodeJSON(src io.Reader, dst interface{}) error {
	defer io.Copy(io.Discard, src)
	err := json.NewDecoder(src).Decode(dst)
	if err != nil {
		// If there is an error during decoding, start the triage...
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
	return nil
}

// func (app *Application) readJsonOrBadRequest(w http.ResponseWriter, r *http.Request, data any) {
// 	if err := app.decodeJSON(r.Body, data); err != nil {
// 		app.Http.BadRequest(w, r, err.Error())
// 		return
// 	}
// }
