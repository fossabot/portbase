package storage

import "errors"

// Errors for storages
var (
	ErrNotFound = errors.New("storage entry could not be found")
)
