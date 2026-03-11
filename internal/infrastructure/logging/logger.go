package logging

import (
	"fmt"
	"learning-go/internal/infrastructure/config"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// NewLogger creates a *zap.Logger based on config.
// LOG_CHANNELS: comma-separated list of channels (console, daily, otlp).
// Multiple channels are combined via zapcore.NewTee.
func NewLogger(cfg *config.Config) (*zap.Logger, error) {
	level, err := zapcore.ParseLevel(cfg.GetLogLevel())
	if err != nil {
		level = zapcore.InfoLevel
	}

	channels := cfg.GetLogChannels()
	cores := make([]zapcore.Core, 0, len(channels))

	for _, ch := range channels {
		core, err := buildChannelCore(ch, cfg, level)
		if err != nil {
			return nil, fmt.Errorf("log channel %q: %w", ch, err)
		}
		cores = append(cores, core)
	}

	if len(cores) == 0 {
		return nil, fmt.Errorf("no log channels configured")
	}

	combined := zapcore.NewTee(cores...)
	z := zap.New(combined, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	z = z.With(zap.String("service", cfg.GetServiceName()))

	zap.RedirectStdLog(z)

	return z, nil
}

func buildChannelCore(channel string, cfg *config.Config, level zapcore.Level) (zapcore.Core, error) {
	switch channel {
	case "console":
		return buildConsoleCore(cfg, level), nil
	case "daily":
		return buildDailyCore(cfg, level)
	case "otlp":
		encoder := buildEncoder(cfg)
		buffered := &zapcore.BufferedWriteSyncer{
			WS:   zapcore.AddSync(os.Stdout),
			Size: 4096,
		}
		return zapcore.NewCore(encoder, buffered, level), nil
	default:
		return nil, fmt.Errorf("unknown log channel: %s", channel)
	}
}

func buildConsoleCore(cfg *config.Config, level zapcore.Level) zapcore.Core {
	encoder := buildEncoder(cfg)
	return zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
}

func buildDailyCore(cfg *config.Config, level zapcore.Level) (zapcore.Core, error) {
	dir := filepath.Dir(cfg.GetLogFilePath())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create log directory %s: %w", dir, err)
	}

	writer := &lumberjack.Logger{
		Filename:   cfg.GetLogFilePath(),
		MaxSize:    cfg.GetLogMaxSize(),
		MaxBackups: cfg.GetLogMaxBackups(),
		MaxAge:     cfg.GetLogMaxAge(),
		Compress:   cfg.LogCompress,
		LocalTime:  true,
	}

	encoder := buildEncoder(cfg)
	return zapcore.NewCore(encoder, zapcore.AddSync(writer), level), nil
}

func buildEncoder(cfg *config.Config) zapcore.Encoder {
	if cfg.GetLogFormat() == "text" {
		return zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	}
	return zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
}
