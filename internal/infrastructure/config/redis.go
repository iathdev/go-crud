package config

type RedisConfig struct {
	RedisHost     string `mapstructure:"REDIS_HOST"`
	RedisPort     string `mapstructure:"REDIS_PORT"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisDB       int    `mapstructure:"REDIS_DB"`
}

func (config *Config) GetRedisAddr() string {
	host := config.RedisHost
	if host == "" {
		host = "localhost"
	}
	port := config.RedisPort
	if port == "" {
		port = "6379"
	}
	return host + ":" + port
}
