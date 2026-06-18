package repository

import "errors"

var (
	ErrDuplicateUsername = errors.New("duplicate username")
	ErrNotFound          = errors.New("record not found")
)
