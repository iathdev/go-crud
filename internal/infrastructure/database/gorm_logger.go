package database

import (
	"context"
	"errors"
	"fmt"
	"learning-go/internal/shared/logger"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// GormLogger bridges GORM's logger to our global *zap.Logger.
type GormLogger struct {
	SlowThreshold time.Duration
}

func NewGormLogger() *GormLogger {
	return &GormLogger{
		SlowThreshold: 200 * time.Millisecond,
	}
}

func (gormLogger *GormLogger) LogMode(gormlogger.LogLevel) gormlogger.Interface {
	return gormLogger
}

func (gormLogger *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	logger.WithContext(ctx).Info(fmt.Sprintf(msg, data...))
}

func (gormLogger *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	logger.WithContext(ctx).Warn(fmt.Sprintf(msg, data...))
}

func (gormLogger *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	logger.WithContext(ctx).Error(fmt.Sprintf(msg, data...))
}

func (gormLogger *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	fields := []zap.Field{
		zap.Duration("elapsed", elapsed),
		zap.String("sql", sql),
		zap.Int64("rows", rows),
	}

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		logger.WithContext(ctx).Error("gorm_query", append(fields, zap.Error(err))...)
	case elapsed > gormLogger.SlowThreshold:
		logger.WithContext(ctx).Warn("gorm_slow_query", fields...)
	default:
		logger.WithContext(ctx).Debug("gorm_query", fields...)
	}
}
