package di

import (
	"learning-go/internal/auth"
	"learning-go/internal/infrastructure/config"
	"learning-go/internal/server"
	"learning-go/internal/shared/logger"
	"learning-go/internal/vocabulary"
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

	ocr := initOCR(cfg, pst.redisClient)

	cleanup := func() {
		ocr.cleanup()
		pst.cleanup()
		obs.cleanup()
	}


	// Modules
	authModule := auth.NewModule(pst.db, cfg)
	vocabularyModule := vocabulary.NewModule(pst.db, ocr.engines)

	// Router & Server
	router := server.NewRouter(authModule, vocabularyModule, pst.db, cfg)
	srv := server.NewServer(cfg, router)

	ocrEngineNames := make([]string, 0, len(ocr.engines))
	for key := range ocr.engines {
		ocrEngineNames = append(ocrEngineNames, string(key))
	}

	logger.Info("[SERVER] app initialized successfully",
		zap.String("service", cfg.GetServiceName()),
		zap.String("log_channels", strings.Join(cfg.GetLogChannels(), ",")),
		zap.Bool("tracing_enabled", cfg.OTLPEndpoint != ""),
		zap.Bool("sentry_enabled", cfg.SentryDSN != ""),
		zap.Strings("ocr_engines", ocrEngineNames),
	)

	return srv, cleanup, nil
}
