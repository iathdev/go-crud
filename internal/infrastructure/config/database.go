package config

type DatabaseConfig struct {
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	DBSSLMODE  string `mapstructure:"DB_SSLMODE"`
	DBTimezone string `mapstructure:"DB_TIMEZONE"`

	DBMaxOpenConns    int `mapstructure:"DB_MAX_OPEN_CONNS"`
	DBMaxIdleConns    int `mapstructure:"DB_MAX_IDLE_CONNS"`
	DBConnMaxLifetime int `mapstructure:"DB_CONN_MAX_LIFETIME"`
	DBConnMaxIdleTime int `mapstructure:"DB_CONN_MAX_IDLE_TIME"`
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

func (config *Config) GetDBConnMaxLifetime() int {
	if config.DBConnMaxLifetime == 0 {
		return 5
	}
	return config.DBConnMaxLifetime
}

func (config *Config) GetDBConnMaxIdleTime() int {
	if config.DBConnMaxIdleTime == 0 {
		return 1
	}
	return config.DBConnMaxIdleTime
}
