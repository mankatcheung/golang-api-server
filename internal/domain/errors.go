// Package domain defines shared types and errors used across the application.
package domain

import "errors"

var (
	ErrNotFound = errors.New("resource not found")
	ErrConflict = errors.New("resource already exists")
)
