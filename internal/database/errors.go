package database

import "errors"

// ErrNotFound is returned when a database query matches no rows.
var ErrNotFound = errors.New("not found")
