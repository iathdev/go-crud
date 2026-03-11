package di

import (
	"learning-go/internal/infrastructure/config"
	"learning-go/internal/infrastructure/database"
	infraredis "learning-go/internal/infrastructure/redis"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type persistence struct {
	db          *gorm.DB
	redisClient *redis.Client
}

func initPersistence(cfg *config.Config) (*persistence, error) {
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		return nil, err
	}

	redisClient, err := infraredis.NewRedisClient(cfg)
	if err != nil {
		return nil, err
	}

	return &persistence{
		db:          db,
		redisClient: redisClient,
	}, nil
}

func (persist *persistence) cleanup() {
	sqlDB, err := persist.db.DB()
	if err == nil {
		sqlDB.Close()
	}
	persist.redisClient.Close()
}
