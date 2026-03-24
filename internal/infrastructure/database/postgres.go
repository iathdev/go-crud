package database

import (
	"database/sql"
	"fmt"
	"learning-go/internal/infrastructure/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgresDB(cfg *config.Config) (*gorm.DB, error) {
	sslmode := cfg.DBSSLMODE
	if sslmode == "" {
		sslmode = "disable"
	}
	timezone := cfg.DBTimezone
	if timezone == "" {
		timezone = "UTC"
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, sslmode, timezone)

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger: NewGormLogger(cfg.GetDBSlowThreshold()),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	configurePool(sqlDB, cfg)

	return db, nil
}

func configurePool(sqlDB *sql.DB, cfg *config.Config) {
	sqlDB.SetMaxOpenConns(cfg.GetDBMaxOpenConns())
	sqlDB.SetMaxIdleConns(cfg.GetDBMaxIdleConns())
	sqlDB.SetConnMaxLifetime(cfg.GetDBConnMaxLifetime())
	sqlDB.SetConnMaxIdleTime(cfg.GetDBConnMaxIdleTime())
}
