package config

import "time"

const defaultPrepMeEndpoint = "/auth/api/v1.1/auth/me"

type AuthConfig struct {
	PrepUserServiceURL    string        `mapstructure:"PREP_USER_SERVICE_URL"`
	PrepMeEndpoint        string        `mapstructure:"PREP_ME_ENDPOINT"`
	PrepHTTPClientTimeout time.Duration `mapstructure:"PREP_HTTP_CLIENT_TIMEOUT"`
}

func (config *Config) GetPrepMeEndpoint() string {
	if config.PrepMeEndpoint == "" {
		return defaultPrepMeEndpoint
	}
	return config.PrepMeEndpoint
}

func (config *Config) GetPrepHTTPClientTimeout() time.Duration {
	if config.PrepHTTPClientTimeout == 0 {
		return 10 * time.Second
	}
	return config.PrepHTTPClientTimeout
}
