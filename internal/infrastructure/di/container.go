package di

import (
	"learning-go/internal/auth"
	"learning-go/internal/infrastructure/config"
	"learning-go/internal/server"
	"learning-go/internal/shared/logger"
	"learning-go/internal/vocabulary"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})
	}
}

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
	authModule := auth.NewModule(pst.db, pst.redisClient, cfg)
	vocabularyModule := vocabulary.NewModule(pst.db, ocr.engines)

	// Router & Server
	router := server.NewRouter(authModule, vocabularyModule, pst.db, pst.redisClient, cfg)
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
