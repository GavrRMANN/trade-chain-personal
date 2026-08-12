package service

import "errors"

var (
	ErrNotFound     = errors.New("resource not found")
	ErrConflict     = errors.New("resource conflict")
	ErrInvalidInput = errors.New("invalid input")
	ErrForbidden    = errors.New("operation forbidden")
	ErrInternal     = errors.New("internal error")
)
