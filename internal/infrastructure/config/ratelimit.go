package config

type RateLimitConfig struct {
	RateLimitRPS   float64 `mapstructure:"RATE_LIMIT_RPS"`
	RateLimitBurst int     `mapstructure:"RATE_LIMIT_BURST"`
}

func (config *Config) GetRateLimitRPS() float64 {
	if config.RateLimitRPS == 0 {
		return 5
	}
	return config.RateLimitRPS
}

func (config *Config) GetRateLimitBurst() int {
	if config.RateLimitBurst == 0 {
		return 10
	}
	return config.RateLimitBurst
}
