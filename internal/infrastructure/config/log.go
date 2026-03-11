package config

import "strings"

type LogConfig struct {
	LogLevel    string `mapstructure:"LOG_LEVEL"`
	LogFormat   string `mapstructure:"LOG_FORMAT"`
	LogChannels string `mapstructure:"LOG_CHANNELS"`

	// Daily file rotation
	LogFilePath   string `mapstructure:"LOG_FILE_PATH"`
	LogMaxSize    int    `mapstructure:"LOG_MAX_SIZE"`
	LogMaxBackups int    `mapstructure:"LOG_MAX_BACKUPS"`
	LogMaxAge     int    `mapstructure:"LOG_MAX_AGE"`
	LogCompress   bool   `mapstructure:"LOG_COMPRESS"`
}

func (config *Config) GetLogLevel() string {
	if config.LogLevel == "" {
		return "info"
	}
	return config.LogLevel
}

func (config *Config) GetLogFormat() string {
	if config.LogFormat == "" {
		return "json"
	}
	return config.LogFormat
}

// GetLogChannels returns the list of active log channels.
// Supports: console, daily, otlp. Multiple channels comma-separated.
// Default: "console"
func (config *Config) GetLogChannels() []string {
	if config.LogChannels == "" {
		return []string{"console"}
	}
	parts := strings.Split(config.LogChannels, ",")
	channels := make([]string, 0, len(parts))
	for _, p := range parts {
		ch := strings.TrimSpace(p)
		if ch != "" {
			channels = append(channels, ch)
		}
	}
	if len(channels) == 0 {
		return []string{"console"}
	}
	return channels
}

func (config *Config) GetLogFilePath() string {
	if config.LogFilePath == "" {
		return "logs/app.log"
	}
	return config.LogFilePath
}

func (config *Config) GetLogMaxSize() int {
	if config.LogMaxSize == 0 {
		return 100 // MB
	}
	return config.LogMaxSize
}

func (config *Config) GetLogMaxBackups() int {
	if config.LogMaxBackups == 0 {
		return 7
	}
	return config.LogMaxBackups
}

func (config *Config) GetLogMaxAge() int {
	if config.LogMaxAge == 0 {
		return 30 // days
	}
	return config.LogMaxAge
}
