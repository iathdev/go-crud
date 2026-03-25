package sharederror

import (
	"errors"
)

type Code string

const (
	CodeBadRequest          Code = "BAD_REQUEST"           // 400
	CodeUnauthorized        Code = "UNAUTHORIZED"          // 401
	CodeForbidden           Code = "FORBIDDEN"             // 403
	CodeNotFound            Code = "NOT_FOUND"             // 404
	CodeConflict            Code = "CONFLICT"              // 409
	CodeUnprocessableEntity Code = "UNPROCESSABLE_ENTITY"  // 422
	CodeInternalServerError Code = "INTERNAL_SERVER_ERROR" // 500
	CodeServiceUnavailable  Code = "SERVICE_UNAVAILABLE"   // 503
)

type AppError struct {
	code    Code
	message string
	cause   error
	data    map[string]any
}

func (appErr *AppError) Error() string {
	return string(appErr.code) + ": " + appErr.message
}

func (appErr *AppError) Code() Code           { return appErr.code }
func (appErr *AppError) Message() string      { return appErr.message }
func (appErr *AppError) Unwrap() error        { return appErr.cause }
func (appErr *AppError) Data() map[string]any { return appErr.data }

func (appErr *AppError) WithData(data map[string]any) *AppError {
	appErr.data = data
	return appErr
}

func (appErr *AppError) Is(target error) bool {
	var t *AppError
	if errors.As(target, &t) {
		return appErr.code == t.code
	}
	return false
}

// --- Client errors (4xx): no cause, no logging needed ---

func BadRequest(message string) *AppError {
	return &AppError{code: CodeBadRequest, message: message}
}

func Unauthorized(message string) *AppError {
	return &AppError{code: CodeUnauthorized, message: message}
}

func Forbidden(message string) *AppError {
	return &AppError{code: CodeForbidden, message: message}
}

func NotFound(message string) *AppError {
	return &AppError{code: CodeNotFound, message: message}
}

func Conflict(message string) *AppError {
	return &AppError{code: CodeConflict, message: message}
}

func UnprocessableEntity(message string) *AppError {
	return &AppError{code: CodeUnprocessableEntity, message: message}
}

// --- Server errors (5xx): carry cause for handler-layer logging ---

func InternalServerError(message string, cause error) *AppError {
	return &AppError{code: CodeInternalServerError, message: message, cause: cause}
}

func ServiceUnavailable(message string, cause error) *AppError {
	return &AppError{code: CodeServiceUnavailable, message: message, cause: cause}
}

func IsAppError(err error) (*AppError, bool) {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}
