package service

import (
	"context"
	"encoding/json"
	"learning-go/internal/auth/application/port"
	"learning-go/internal/auth/domain"
	"learning-go/internal/infrastructure/circuitbreaker"
	sharederror "learning-go/internal/shared/error"
	"learning-go/internal/shared/logger"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type PrepUserService struct {
	baseURL string
	client  *http.Client
	breaker *circuitbreaker.Breaker
}

// prepMeResponse maps the response from GET /auth/api/v1.1/auth/me
type prepMeResponse struct {
	Data    prepMeData `json:"data"`
	Message string     `json:"message"`
}

type prepMeData struct {
	ID           int64  `json:"id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	IsFirstLogin bool   `json:"is_first_login"`
}

func NewPrepUserService(baseURL string, breaker *circuitbreaker.Breaker) port.PrepUserServicePort {
	return &PrepUserService{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
		breaker: breaker,
	}
}

func (service *PrepUserService) ValidateToken(ctx context.Context, token string) (*domain.PrepUser, error) {
	result, err := service.breaker.Execute(func() (any, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, service.baseURL+"/auth/api/v1.1/auth/me", nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := service.client.Do(req)
		if err != nil {
			logger.WithContext(ctx).Error("prep user service request failed", zap.Error(err))
			return nil, sharederror.ErrSSOServiceError
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			return nil, sharederror.ErrSSOTokenInvalid
		}
		if resp.StatusCode != http.StatusOK {
			logger.WithContext(ctx).Error("prep user service unexpected status",
				zap.String("status", resp.Status),
			)
			return nil, sharederror.ErrSSOServiceError
		}

		var body prepMeResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			logger.WithContext(ctx).Error("prep user service decode error", zap.Error(err))
			return nil, sharederror.ErrSSOServiceError
		}

		return domain.NewPrepUser(body.Data.ID, body.Data.Email, body.Data.Name), nil
	})
	if err != nil {
		return nil, err
	}
	return result.(*domain.PrepUser), nil
}
