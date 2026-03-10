package i18n

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"en", "en"},
		{"vi", "vi"},
		{"vi-VN", "vi"},
		{"th-TH", "th"},
		{"zh-CN", "zh"},
		{"id-ID", "id"},
		{"en-US", "en"},
		{"en-GB", "en"},
		{"", "en"},
		{"fr", "en"},
		{"de-DE", "en"},
		{"  vi  ", "vi"},
		{"VI", "vi"},
	}

	for _, tt := range tests {
		result := Normalize(tt.input)
		if result != tt.expected {
			t.Errorf("Normalize(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFromAcceptLanguage(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"vi-VN,vi;q=0.9,en;q=0.8", "vi"},
		{"en-US,en;q=0.9", "en"},
		{"", "en"},
		{"th", "th"},
	}

	for _, tt := range tests {
		result := FromAcceptLanguage(tt.input)
		if result != tt.expected {
			t.Errorf("FromAcceptLanguage(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestIsKey(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"common.not_found", true},
		{"auth.email_already_exists", true},
		{"product.name_required", true},
		{"just_text", false},
		{"user@email.com", false},
		{"Hello World", false},
		{"", false},
		{"a.b", true},
		{"192.168.1.1", false},
		{"COMMON.NOT_FOUND", false},
	}

	for _, tt := range tests {
		result := IsKey(tt.input)
		if result != tt.expected {
			t.Errorf("IsKey(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}
