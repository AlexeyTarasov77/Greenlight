package movies

import "errors"

var (
	ErrMovieNotFound      = errors.New("movie not found")
	ErrMovieAlreadyExists = errors.New("movie with that title, version and year already exists")
	ErrNoArgumentsChanged = errors.New("no arguments changed")
	ErrEditConflict       = errors.New("unable to update the record due to an edit conflict, please try again")
)
