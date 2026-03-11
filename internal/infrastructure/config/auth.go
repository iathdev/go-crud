package config

type AuthConfig struct {
	JWTSecret          string `mapstructure:"JWT_SECRET"`
	AccessTokenExpiry  int    `mapstructure:"ACCESS_TOKEN_EXPIRY"`
	RefreshTokenExpiry int    `mapstructure:"REFRESH_TOKEN_EXPIRY"`
	PrepUserServiceURL string `mapstructure:"PREP_USER_SERVICE_URL"`
}

func (config *Config) GetAccessTokenExpiry() int {
	if config.AccessTokenExpiry == 0 {
		return 1440
	}
	return config.AccessTokenExpiry
}

func (config *Config) GetRefreshTokenExpiry() int {
	if config.RefreshTokenExpiry == 0 {
		return 7
	}
	return config.RefreshTokenExpiry
}
