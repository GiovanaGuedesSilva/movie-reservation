package apperrors

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError is a domain error that carries an HTTP status code and a
// machine-readable code for API responses.
type AppError struct {
	Code       string
	Message    string
	StatusCode int
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// Sentinel domain errors — compare with errors.Is.
var (
	ErrNotFound = &AppError{
		Code:       "NOT_FOUND",
		Message:    "resource not found",
		StatusCode: http.StatusNotFound,
	}

	ErrUnauthorized = &AppError{
		Code:       "UNAUTHORIZED",
		Message:    "authentication required",
		StatusCode: http.StatusUnauthorized,
	}

	ErrForbidden = &AppError{
		Code:       "FORBIDDEN",
		Message:    "you do not have permission to perform this action",
		StatusCode: http.StatusForbidden,
	}

	ErrConflict = &AppError{
		Code:       "CONFLICT",
		Message:    "resource already exists",
		StatusCode: http.StatusConflict,
	}

	ErrBadRequest = &AppError{
		Code:       "BAD_REQUEST",
		Message:    "invalid request",
		StatusCode: http.StatusBadRequest,
	}

	ErrSeatAlreadyTaken = &AppError{
		Code:       "SEAT_ALREADY_TAKEN",
		Message:    "one or more seats are already reserved",
		StatusCode: http.StatusConflict,
	}

	ErrCannotCancelPastReservation = &AppError{
		Code:       "CANNOT_CANCEL_PAST_RESERVATION",
		Message:    "cannot cancel a reservation for a past showtime",
		StatusCode: http.StatusUnprocessableEntity,
	}

	ErrInvalidCredentials = &AppError{
		Code:       "INVALID_CREDENTIALS",
		Message:    "email or password is incorrect",
		StatusCode: http.StatusUnauthorized,
	}
)

// New wraps an existing error with an AppError.
func New(base *AppError, err error) *AppError {
	return &AppError{
		Code:       base.Code,
		Message:    base.Message,
		StatusCode: base.StatusCode,
		Err:        err,
	}
}

// Newf creates a new AppError with a formatted message.
func Newf(base *AppError, format string, args ...any) *AppError {
	return &AppError{
		Code:       base.Code,
		Message:    fmt.Sprintf(format, args...),
		StatusCode: base.StatusCode,
	}
}

// As attempts to extract an *AppError from err.
func As(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
