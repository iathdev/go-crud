package config

type RateLimitConfig struct {
	RateLimitRPS   float64 `mapstructure:"RATE_LIMIT_RPS"`
	RateLimitBurst int     `mapstructure:"RATE_LIMIT_BURST"`
}

func (config *Config) GetRateLimitRPS() float64 {
	if config.RateLimitRPS == 0 {
		return 20
	}
	return config.RateLimitRPS
}

func (config *Config) GetRateLimitBurst() int {
	if config.RateLimitBurst == 0 {
		return 50
	}
	return config.RateLimitBurst
}
