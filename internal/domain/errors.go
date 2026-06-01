package domain

import "errors"

// Ошибки домена (business logic errors)
var (
	ErrNotFound            = errors.New("resource not found")
	ErrInvalidInput        = errors.New("invalid input parameters")
	ErrInternalServer      = errors.New("internal server error")
	ErrInsufficientBalance = errors.New("insufficient balance")
)
