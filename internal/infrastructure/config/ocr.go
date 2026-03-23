package config

type OCRConfig struct {
	OCRServiceURL                string `mapstructure:"OCR_SERVICE_URL"`
	GoogleApplicationCredentials string `mapstructure:"GOOGLE_APPLICATION_CREDENTIALS"`
}
