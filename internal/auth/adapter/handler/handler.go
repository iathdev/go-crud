package handler

import (
	"errors"
	"learning-go/internal/auth/application/dto"
	"learning-go/internal/auth/application/port"
	sharederror "learning-go/internal/shared/error"
	"learning-go/internal/shared/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authUseCase port.AuthUseCasePort
}

func NewAuthHandler(authUseCase port.AuthUseCasePort) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
	}
}

func (handler *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationBadRequest(c, err)
		return
	}

	res, err := handler.authUseCase.Login(c.Request.Context(), req)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	response.Success(c, http.StatusOK, res)
}

func (handler *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationBadRequest(c, err)
		return
	}

	res, err := handler.authUseCase.RefreshToken(c.Request.Context(), req)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	response.Success(c, http.StatusOK, res)
}

func (handler *AuthHandler) Logout(c *gin.Context) {
	var req dto.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationBadRequest(c, err)
		return
	}

	if err := handler.authUseCase.Logout(c.Request.Context(), req); err != nil {
		handleAuthError(c, err)
		return
	}

	response.Success(c, http.StatusOK, nil, "auth.logged_out")
}

func (handler *AuthHandler) GetProfile(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.Unauthorized(c, "auth.unauthorized")
		return
	}

	res, err := handler.authUseCase.GetProfile(c.Request.Context(), userID)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	response.Success(c, http.StatusOK, res, "auth.profile_fetched")
}

func getUserID(c *gin.Context) (uuid.UUID, error) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, errors.New("user_id not in context")
	}
	return uuid.Parse(userIDStr.(string))
}

func handleAuthError(c *gin.Context, err error) {
	var domErr *sharederror.AppError
	if errors.As(err, &domErr) {
		switch domErr.Code() {
		case sharederror.CodeSSOTokenInvalid:
			response.Unauthorized(c, "auth.token_invalid")
		case sharederror.CodeSSOServiceError:
			response.ServiceUnavailable(c, "auth.service_error")
		case sharederror.CodeInvalidInput:
			response.BadRequest(c, "")
		case sharederror.CodeUnauthorized:
			response.Unauthorized(c, "auth.unauthorized")
		case sharederror.CodeNotFound:
			response.NotFound(c, "")
		case sharederror.CodeServiceUnavailable:
			response.ServiceUnavailable(c, "")
		default:
			response.InternalServerError(c, "")
		}
		return
	}

	response.InternalServerError(c, "")
}
