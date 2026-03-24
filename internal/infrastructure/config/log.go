package config

import "strings"

type LogConfig struct {
	LogLevel    string `mapstructure:"LOG_LEVEL"`
	LogFormat   string `mapstructure:"LOG_FORMAT"`
	LogChannels string `mapstructure:"LOG_CHANNELS"`
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
// Supports: console, otlp. Multiple channels comma-separated.
// Default: "console"
func (config *Config) GetLogChannels() []string {
	if config.LogChannels == "" {
		return []string{"console"}
	}
	parts := strings.Split(config.LogChannels, ",")
	channels := make([]string, 0, len(parts))
	for _, part := range parts {
		ch := strings.TrimSpace(part)
		if ch != "" {
			channels = append(channels, ch)
		}
	}
	if len(channels) == 0 {
		return []string{"console"}
	}
	return channels
}
