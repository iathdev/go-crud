package handler

import (
	"errors"
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
	return &AuthHandler{authUseCase: authUseCase}
}

func (handler *AuthHandler) GetMe(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.Unauthorized(c, "auth.unauthorized")
		return
	}

	isFirstLogin, _ := c.Get("is_first_login")
	firstLogin, _ := isFirstLogin.(bool)

	res, err := handler.authUseCase.GetMe(c.Request.Context(), userID, firstLogin)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	response.Success(c, http.StatusOK, res)
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
		msg := domErr.Message()
		switch domErr.Code() {
		case sharederror.CodeNotFound:
			response.NotFound(c, msg)
		case sharederror.CodeUnauthorized:
			response.Unauthorized(c, msg)
		case sharederror.CodeServiceUnavailable:
			response.ServiceUnavailable(c, msg)
		default:
			response.InternalServerError(c, msg)
		}
		return
	}

	response.InternalServerError(c, "")
}
