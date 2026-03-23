package config

type OCRConfig struct {
	OCRServiceURL                string `mapstructure:"OCR_SERVICE_URL"`
	GoogleApplicationCredentials string `mapstructure:"GOOGLE_APPLICATION_CREDENTIALS"`
	BaiduOCRAPIKey               string `mapstructure:"BAIDU_OCR_API_KEY"`
	BaiduOCRSecretKey            string `mapstructure:"BAIDU_OCR_SECRET_KEY"`
}
