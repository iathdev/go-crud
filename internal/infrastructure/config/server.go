package config

import "time"

type ServerConfig struct {
	ReadTimeout  int `mapstructure:"SERVER_READ_TIMEOUT"`  // seconds, default 15
	WriteTimeout int `mapstructure:"SERVER_WRITE_TIMEOUT"` // seconds, default 30
	IdleTimeout  int `mapstructure:"SERVER_IDLE_TIMEOUT"`  // seconds, default 120
}

func (c *ServerConfig) GetReadTimeout() time.Duration {
	if c.ReadTimeout <= 0 {
		return 15 * time.Second
	}
	return time.Duration(c.ReadTimeout) * time.Second
}

func (c *ServerConfig) GetWriteTimeout() time.Duration {
	if c.WriteTimeout <= 0 {
		return 30 * time.Second
	}
	return time.Duration(c.WriteTimeout) * time.Second
}

func (c *ServerConfig) GetIdleTimeout() time.Duration {
	if c.IdleTimeout <= 0 {
		return 120 * time.Second
	}
	return time.Duration(c.IdleTimeout) * time.Second
}
