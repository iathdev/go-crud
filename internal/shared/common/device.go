package common

import "strings"

// ParseDeviceInfo extracts device information from User-Agent string.
func ParseDeviceInfo(userAgent string) map[string]string {
	device := map[string]string{
		"type":     "unknown",
		"os":       "unknown",
		"browser":  "unknown",
		"platform": "unknown",
	}

	ua := strings.ToLower(userAgent)

	// Detect device type
	switch {
	case strings.Contains(ua, "mobile") || strings.Contains(ua, "android") && !strings.Contains(ua, "tablet"):
		device["type"] = "mobile"
	case strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad"):
		device["type"] = "tablet"
	default:
		device["type"] = "desktop"
	}

	// Detect OS
	switch {
	case strings.Contains(ua, "windows"):
		device["os"] = "Windows"
	case strings.Contains(ua, "mac os") || strings.Contains(ua, "macintosh"):
		device["os"] = "macOS"
	case strings.Contains(ua, "linux"):
		device["os"] = "Linux"
	case strings.Contains(ua, "android"):
		device["os"] = "Android"
	case strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") || strings.Contains(ua, "ios"):
		device["os"] = "iOS"
	}

	// Detect browser
	switch {
	case strings.Contains(ua, "edg/") || strings.Contains(ua, "edge"):
		device["browser"] = "Edge"
	case strings.Contains(ua, "chrome") && !strings.Contains(ua, "edg"):
		device["browser"] = "Chrome"
	case strings.Contains(ua, "firefox"):
		device["browser"] = "Firefox"
	case strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome"):
		device["browser"] = "Safari"
	case strings.Contains(ua, "opera") || strings.Contains(ua, "opr/"):
		device["browser"] = "Opera"
	}

	// Detect platform
	switch {
	case strings.Contains(ua, "win64") || strings.Contains(ua, "x64"):
		device["platform"] = "x64"
	case strings.Contains(ua, "win32") || strings.Contains(ua, "x86"):
		device["platform"] = "x86"
	case strings.Contains(ua, "arm"):
		device["platform"] = "ARM"
	}

	return device
}
