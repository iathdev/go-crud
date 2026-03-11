package sharederror

import "errors"

type Code string

const (
	CodeNotFound           Code = "NOT_FOUND"
	CodeInvalidInput       Code = "INVALID_INPUT"
	CodeEmailAlreadyExists Code = "EMAIL_ALREADY_EXISTS"
	CodeUnauthorized       Code = "UNAUTHORIZED"
	CodeInternal           Code = "INTERNAL"
	CodeServiceUnavailable Code = "SERVICE_UNAVAILABLE"
	CodeSSOTokenInvalid  Code = "SSO_TOKEN_INVALID"
	CodeSSOServiceError  Code = "SSO_SERVICE_ERROR"
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

func (appErr *AppError) Code() Code    { return appErr.code }
func (appErr *AppError) Unwrap() error { return appErr.cause }

func (appErr *AppError) Is(target error) bool {
	var t *AppError
	if errors.As(target, &t) {
		return appErr.code == t.code
	}
	return false
}

var (
	ErrNotFound           = &AppError{code: CodeNotFound, message: "not found"}
	ErrInvalidInput       = &AppError{code: CodeInvalidInput, message: "invalid input"}
	ErrEmailAlreadyExists = &AppError{code: CodeEmailAlreadyExists, message: "email already exists"}
	ErrUnauthorized       = &AppError{code: CodeUnauthorized, message: "unauthorized"}
	ErrInternal           = &AppError{code: CodeInternal, message: "internal error"}
	ErrServiceUnavailable = &AppError{code: CodeServiceUnavailable, message: "service unavailable"}
	ErrSSOTokenInvalid  = &AppError{code: CodeSSOTokenInvalid, message: "SSO token invalid"}
	ErrSSOServiceError  = &AppError{code: CodeSSOServiceError, message: "SSO service error"}
)

func NewInternal(cause error) *AppError {
	return &AppError{code: CodeInternal, message: "internal error", cause: cause}
}

func NewServiceUnavailable(cause error) *AppError {
	return &AppError{code: CodeServiceUnavailable, message: "service unavailable", cause: cause}
}

func IsAppError(err error) (*AppError, bool) {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}
