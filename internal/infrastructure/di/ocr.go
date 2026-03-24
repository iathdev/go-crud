package di

import (
	"learning-go/internal/infrastructure/circuitbreaker"
	"learning-go/internal/infrastructure/config"
	sharederror "learning-go/internal/shared/error"
	"learning-go/internal/shared/logger"
	vocabservice "learning-go/internal/vocabulary/adapter/service"
	vocabport "learning-go/internal/vocabulary/application/port"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type ocrResult struct {
	engines  vocabport.OCREngineRegistry
	cleanups []func()
}

func (r *ocrResult) cleanup() {
	for i := len(r.cleanups) - 1; i >= 0; i-- {
		r.cleanups[i]()
	}
}

func (r *ocrResult) register(key vocabport.OCREngineKey, engine vocabport.OCRServicePort, cleanupFn ...func()) {
	r.engines[key] = engine
	if len(cleanupFn) > 0 {
		r.cleanups = append(r.cleanups, cleanupFn[0])
	}
}

func initOCR(cfg *config.Config, redisClient *redis.Client) ocrResult {
	result := ocrResult{engines: vocabport.OCREngineRegistry{}}

	withRetry := func(engine vocabport.OCRServicePort) vocabport.OCRServicePort {
		return vocabservice.NewOCRRetryDecorator(engine, cfg.GetOCRRetryMax(), cfg.GetOCRRetryDelay())
	}

	if cfg.OCRServiceURL != "" {
		adapter := vocabservice.NewOCRService(cfg.OCRServiceURL, newOCRBreaker("paddle-ocr"))
		result.register(vocabport.OCREnginePaddleOCR, withRetry(adapter))
	}

	if cfg.GoogleApplicationCredentials != "" {
		adapter, cleanup, err := vocabservice.NewGoogleVisionService(cfg.GoogleApplicationCredentials, newOCRBreaker("google-vision"))
		if err != nil {
			logger.Warn("[DI] Google Vision init failed, skipping", zap.Error(err))
		} else {
			result.register(vocabport.OCREngineGoogleVision, withRetry(adapter), cleanup)
		}
	}

	if cfg.BaiduOCRAPIKey != "" && cfg.BaiduOCRSecretKey != "" {
		adapter := vocabservice.NewBaiduOCRService(cfg.BaiduOCRAPIKey, cfg.BaiduOCRSecretKey, newOCRBreaker("baidu-ocr"), redisClient)
		result.register(vocabport.OCREngineBaiduOCR, withRetry(adapter))
	}

	return result
}

func newOCRBreaker(name string) *circuitbreaker.Breaker {
	return circuitbreaker.NewBreaker(circuitbreaker.BreakerConfig{
		Name: name,
	}, func(err error) bool {
		if err == nil {
			return true
		}
		if appErr, ok := sharederror.IsAppError(err); ok {
			return appErr.Code() == sharederror.CodeNotFound
		}
		return false
	})
}
