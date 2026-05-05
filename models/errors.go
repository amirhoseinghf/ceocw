package models

import "errors"

var (
	ErrInvalidCredentials = errors.New("models: invalid credentials")
	ErrDuplicateEmail     = errors.New("models: duplicate email")
	ErrNoRecord           = errors.New("models: no matching record found")
)
