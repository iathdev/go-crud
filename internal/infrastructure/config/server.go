package config

import "time"

type ServerConfig struct {
	ReadTimeout  time.Duration `mapstructure:"SERVER_READ_TIMEOUT"`
	WriteTimeout time.Duration `mapstructure:"SERVER_WRITE_TIMEOUT"`
	IdleTimeout  time.Duration `mapstructure:"SERVER_IDLE_TIMEOUT"`
}

func (config *Config) GetReadTimeout() time.Duration {
	if config.ReadTimeout == 0 {
		return 15 * time.Second
	}
	return config.ReadTimeout
}

func (config *Config) GetWriteTimeout() time.Duration {
	if config.WriteTimeout == 0 {
		return 30 * time.Second
	}
	return config.WriteTimeout
}

func (config *Config) GetIdleTimeout() time.Duration {
	if config.IdleTimeout == 0 {
		return 120 * time.Second
	}
	return config.IdleTimeout
}
