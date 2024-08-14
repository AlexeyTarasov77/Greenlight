package main

import "net/http"


func (app *Application) Recoverer(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil && err != http.ErrAbortHandler {
				app.Http.ServerError(w, r, err.(error), "")
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}