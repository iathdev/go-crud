package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	AppPort string `mapstructure:"APP_PORT"`
	GinMode string `mapstructure:"GIN_MODE"`

	DatabaseConfig       `mapstructure:",squash"`
	CircuitBreakerConfig `mapstructure:",squash"`
	LogConfig            `mapstructure:",squash"`
	ObservabilityConfig  `mapstructure:",squash"`
	RedisConfig          `mapstructure:",squash"`
	AuthConfig           `mapstructure:",squash"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	if config.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required and must not be empty")
	}

	return &config, nil
}
