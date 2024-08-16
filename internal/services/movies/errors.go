package movies

import "errors"

var (
	ErrMovieNotFound = errors.New("movie not found")
	ErrMovieAlreadyExists = errors.New("movie with that title, version and year already exists")
)
