package config

import "time"

type DatabaseConfig struct {
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	DBSSLMODE  string `mapstructure:"DB_SSLMODE"`
	DBTimezone string `mapstructure:"DB_TIMEZONE"`

	DBMaxOpenConns    int           `mapstructure:"DB_MAX_OPEN_CONNS"`
	DBMaxIdleConns    int           `mapstructure:"DB_MAX_IDLE_CONNS"`
	DBConnMaxLifetime time.Duration `mapstructure:"DB_CONN_MAX_LIFETIME"`
	DBConnMaxIdleTime time.Duration `mapstructure:"DB_CONN_MAX_IDLE_TIME"`
	DBSlowThreshold   time.Duration `mapstructure:"DB_SLOW_THRESHOLD"`
}

func (config *Config) GetDBMaxOpenConns() int {
	if config.DBMaxOpenConns == 0 {
		return 25
	}
	return config.DBMaxOpenConns
}

func (config *Config) GetDBMaxIdleConns() int {
	if config.DBMaxIdleConns == 0 {
		return 10
	}
	return config.DBMaxIdleConns
}

func (config *Config) GetDBConnMaxLifetime() time.Duration {
	if config.DBConnMaxLifetime == 0 {
		return 5 * time.Minute
	}
	return config.DBConnMaxLifetime
}

func (config *Config) GetDBConnMaxIdleTime() time.Duration {
	if config.DBConnMaxIdleTime == 0 {
		return 1 * time.Minute
	}
	return config.DBConnMaxIdleTime
}

func (config *Config) GetDBSlowThreshold() time.Duration {
	if config.DBSlowThreshold == 0 {
		return 200 * time.Millisecond
	}
	return config.DBSlowThreshold
}
