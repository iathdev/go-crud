package config

type CircuitBreakerConfig struct {
	CBMaxRequests  uint32  `mapstructure:"CB_MAX_REQUESTS"`
	CBInterval     int     `mapstructure:"CB_INTERVAL"`
	CBTimeout      int     `mapstructure:"CB_TIMEOUT"`
	CBFailureRatio float64 `mapstructure:"CB_FAILURE_RATIO"`
	CBMinRequests  uint32  `mapstructure:"CB_MIN_REQUESTS"`
}

func (config *Config) GetCBMaxRequests() uint32 {
	if config.CBMaxRequests == 0 {
		return 5
	}
	return config.CBMaxRequests
}

func (config *Config) GetCBInterval() int {
	if config.CBInterval == 0 {
		return 60
	}
	return config.CBInterval
}

func (config *Config) GetCBTimeout() int {
	if config.CBTimeout == 0 {
		return 30
	}
	return config.CBTimeout
}

func (config *Config) GetCBFailureRatio() float64 {
	if config.CBFailureRatio == 0 {
		return 0.6
	}
	return config.CBFailureRatio
}

func (config *Config) GetCBMinRequests() uint32 {
	if config.CBMinRequests == 0 {
		return 10
	}
	return config.CBMinRequests
}
