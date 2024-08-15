package storage

import "errors"

const EmptyIntValue = -1

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)
