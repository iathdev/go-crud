package sharederror

import (
	"context"
	"errors"
	"learning-go/internal/shared/logger"

	"go.uber.org/zap"
)

type Code string

const (
	CodeNotFound           Code = "NOT_FOUND"
	CodeInvalidInput       Code = "INVALID_INPUT"
	CodeEmailAlreadyExists Code = "EMAIL_ALREADY_EXISTS"
	CodeUnauthorized       Code = "UNAUTHORIZED"
	CodeInternal           Code = "INTERNAL"
	CodeServiceUnavailable Code = "SERVICE_UNAVAILABLE"
	CodeSSOTokenInvalid    Code = "SSO_TOKEN_INVALID"
	CodeSSOServiceError    Code = "SSO_SERVICE_ERROR"
)

type AppError struct {
	code    Code
	message string
	cause   error
}

func (appErr *AppError) Error() string {
	if appErr.cause != nil {
		return appErr.message + ": " + appErr.cause.Error()
	}
	return appErr.message
}

func (appErr *AppError) Code() Code      { return appErr.code }
func (appErr *AppError) Message() string { return appErr.message }
func (appErr *AppError) Unwrap() error   { return appErr.cause }

func (appErr *AppError) Is(target error) bool {
	var t *AppError
	if errors.As(target, &t) {
		return appErr.code == t.code
	}
	return false
}

var (
	ErrNotFound           = &AppError{code: CodeNotFound, message: "common.not_found"}
	ErrInvalidInput       = &AppError{code: CodeInvalidInput, message: "common.invalid_input"}
	ErrEmailAlreadyExists = &AppError{code: CodeEmailAlreadyExists, message: "common.conflict"}
	ErrUnauthorized       = &AppError{code: CodeUnauthorized, message: "common.unauthorized"}
	ErrInternal           = &AppError{code: CodeInternal, message: "common.internal_server_error"}
	ErrServiceUnavailable = &AppError{code: CodeServiceUnavailable, message: "common.service_unavailable"}
	ErrSSOTokenInvalid    = &AppError{code: CodeSSOTokenInvalid, message: "auth.unauthorized"}
	ErrSSOServiceError    = &AppError{code: CodeSSOServiceError, message: "auth.service_unavailable"}
)

func NewNotFound(message string) *AppError {
	return &AppError{code: CodeNotFound, message: message}
}

func NewInvalidInput(message string) *AppError {
	return &AppError{code: CodeInvalidInput, message: message}
}

func NewUnauthorized(message string) *AppError {
	return &AppError{code: CodeUnauthorized, message: message}
}

func NewInternal(ctx context.Context, message string, cause error) *AppError {
	logger.WithContext(ctx).Error(message, zap.Error(cause))
	return &AppError{code: CodeInternal, message: message, cause: cause}
}

func NewServiceUnavailable(ctx context.Context, message string, cause error) *AppError {
	logger.WithContext(ctx).Error(message, zap.Error(cause))
	return &AppError{code: CodeServiceUnavailable, message: message, cause: cause}
}

func IsAppError(err error) (*AppError, bool) {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}
