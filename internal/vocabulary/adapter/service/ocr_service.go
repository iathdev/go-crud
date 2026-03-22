package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"learning-go/internal/infrastructure/circuitbreaker"
	sharederror "learning-go/internal/shared/error"
	"learning-go/internal/shared/logger"
	"learning-go/internal/vocabulary/application/port"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type OCRService struct {
	baseURL string
	client  *http.Client
	breaker *circuitbreaker.Breaker
}

type ocrExtractRequest struct {
	Image    string `json:"image"`
	Language string `json:"language"`
}

type ocrServiceResponse struct {
	Characters []ocrServiceCharacter `json:"characters"`
	Engine     string                `json:"engine"`
}

type ocrServiceCharacter struct {
	Text       string   `json:"text"`
	Confidence float64  `json:"confidence"`
	Candidates []string `json:"candidates"`
}

func NewOCRService(baseURL string, breaker *circuitbreaker.Breaker) port.OCRServicePort {
	return &OCRService{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 30 * time.Second},
		breaker: breaker,
	}
}

func (svc *OCRService) Recognize(ctx context.Context, req port.OCRRequest) (*port.OCRResult, error) {
	result, err := svc.breaker.Execute(func() (any, error) {
		language := req.Language
		if language == "" {
			language = "zh"
		}

		payload := ocrExtractRequest{
			Image:    base64.StdEncoding.EncodeToString(req.Image),
			Language: language,
		}

		body, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, svc.baseURL+"/recognize", bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := svc.client.Do(httpReq)
		if err != nil {
			logger.WithContext(ctx).Error("[OCR] service request failed", zap.Error(err))
			return nil, sharederror.ErrServiceUnavailable
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			logger.WithContext(ctx).Error("[OCR] service unexpected status",
				zap.String("status", resp.Status),
				zap.String("body", string(respBody)),
			)
			return nil, sharederror.ErrServiceUnavailable
		}

		var ocrResp ocrServiceResponse
		if err := json.NewDecoder(resp.Body).Decode(&ocrResp); err != nil {
			logger.WithContext(ctx).Error("[OCR] service decode error", zap.Error(err))
			return nil, sharederror.ErrInternal
		}

		characters := make([]port.OCRCharacter, 0, len(ocrResp.Characters))
		for _, ch := range ocrResp.Characters {
			characters = append(characters, port.OCRCharacter{
				Text:       ch.Text,
				Confidence: ch.Confidence,
				Candidates: ch.Candidates,
			})
		}

		return &port.OCRResult{
			Characters: characters,
			Engine:     ocrResp.Engine,
		}, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(*port.OCRResult), nil
}
