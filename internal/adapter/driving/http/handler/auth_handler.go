package handler

import (
	"errors"
	"learning-go/internal/adapter/driving/http/response"
	"learning-go/internal/application/dto"
	"learning-go/internal/application/port/input"
	domainerror "learning-go/internal/domain/error"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authUseCase input.AuthUseCasePort
}

func NewAuthHandler(authUseCase input.AuthUseCasePort) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "")
		return
	}

	res, err := h.authUseCase.Register(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, domainerror.ErrEmailAlreadyExists) {
			response.Conflict(c, "auth.email_already_exists")
			return
		}
		if errors.Is(err, domainerror.ErrInvalidInput) {
			response.BadRequest(c, "")
			return
		}
		response.InternalServerError(c, "")
		return
	}

	response.Success(c, http.StatusCreated, res)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "")
		return
	}

	res, err := h.authUseCase.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, domainerror.ErrUnauthorized) {
			response.Unauthorized(c, "auth.unauthorized")
			return
		}
		response.InternalServerError(c, "")
		return
	}

	response.Success(c, http.StatusOK, res)
}
