package di

import (
	"learning-go/internal/auth"
	"learning-go/internal/infrastructure/circuitbreaker"
	"learning-go/internal/infrastructure/config"
	"learning-go/internal/server"
	sharederror "learning-go/internal/shared/error"
	"learning-go/internal/shared/logger"
	"learning-go/internal/vocabulary"
	vocabservice "learning-go/internal/vocabulary/adapter/service"
	vocabport "learning-go/internal/vocabulary/application/port"
	"strings"

	"go.uber.org/zap"
)

func NewApp() (*server.Server, func(), error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, nil, err
	}

	obs, err := initObservability(cfg)
	if err != nil {
		return nil, nil, err
	}

	pst, err := initPersistence(cfg)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		pst.cleanup()
		obs.cleanup()
	}

	// OCR adapter
	ocrBreaker := circuitbreaker.NewBreaker(circuitbreaker.BreakerConfig{
		Name: "ocr-service",
	}, func(err error) bool {
		if err == nil {
			return true
		}
		if appErr, ok := sharederror.IsAppError(err); ok {
			return appErr.Code() == sharederror.CodeNotFound
		}
		return false
	})
	ocrAdapter := vocabservice.NewOCRService(cfg.OCRServiceURL, ocrBreaker)

	ocrEngines := vocabport.OCREngineRegistry{
		vocabport.OCREnginePaddleOCR: ocrAdapter,
	}

	// Google Vision adapter (chỉ tạo nếu có credentials)
	if cfg.GoogleApplicationCredentials != "" {
		gvBreaker := circuitbreaker.NewBreaker(circuitbreaker.BreakerConfig{
			Name: "google-vision",
		}, func(err error) bool {
			if err == nil {
				return true
			}
			if appErr, ok := sharederror.IsAppError(err); ok {
				return appErr.Code() == sharederror.CodeNotFound
			}
			return false
		})

		gvAdapter, gvCleanup, err := vocabservice.NewGoogleVisionService(
			cfg.GoogleApplicationCredentials, gvBreaker,
		)
		if err != nil {
			logger.Warn("[DI] Google Vision adapter init failed, skipping", zap.Error(err))
		} else {
			ocrEngines[vocabport.OCREngineGoogleVision] = gvAdapter
			prevCleanup := cleanup
			cleanup = func() {
				gvCleanup()
				prevCleanup()
			}
		}
	}

	// Baidu OCR adapter (chỉ tạo nếu có credentials)
	if cfg.BaiduOCRAPIKey != "" && cfg.BaiduOCRSecretKey != "" {
		baiduBreaker := circuitbreaker.NewBreaker(circuitbreaker.BreakerConfig{
			Name: "baidu-ocr",
		}, func(err error) bool {
			if err == nil {
				return true
			}
			if appErr, ok := sharederror.IsAppError(err); ok {
				return appErr.Code() == sharederror.CodeNotFound
			}
			return false
		})

		baiduAdapter := vocabservice.NewBaiduOCRService(
			cfg.BaiduOCRAPIKey, cfg.BaiduOCRSecretKey,
			baiduBreaker, pst.redisClient,
		)
		ocrEngines[vocabport.OCREngineBaiduOCR] = baiduAdapter
	}

	// Modules
	authModule := auth.NewModule(pst.db, cfg)
	vocabularyModule := vocabulary.NewModule(pst.db, ocrEngines)

	// Router & Server
	router := server.NewRouter(authModule, vocabularyModule, pst.db, cfg)
	srv := server.NewServer(cfg, router)

	logger.Info("[SERVER] app initialized successfully",
		zap.String("service", cfg.GetServiceName()),
		zap.String("log_channels", strings.Join(cfg.GetLogChannels(), ",")),
		zap.Bool("tracing_enabled", cfg.OTLPEndpoint != ""),
		zap.Bool("sentry_enabled", cfg.SentryDSN != ""),
	)

	return srv, cleanup, nil
}
