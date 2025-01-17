package database

import (
	"errors"
)

// Errors
var (
	ErrNotFound         = errors.New("database entry could not be found")
	ErrPermissionDenied = errors.New("access to database record denied")
	ErrReadOnly         = errors.New("database is read only")
	ErrShuttingDown     = errors.New("database system is shutting down")
)
