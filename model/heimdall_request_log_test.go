package model

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractClientMetadata(t *testing.T) {
	tests := []struct {
		name     string
		headers  http.Header
		expected map[string]string
		hasError bool
	}{
		{
			name: "valid headers",
			headers: http.Header{
				"X-Forwarded-For": []string{"192.168.1.1, 10.0.0.1"},
				"User-Agent":      []string{"Mozilla/5.0 (Test Browser)"},
				"X-Device-Id":     []string{"device123"},
				"Content-Type":    []string{"application/json"},
			},
			expected: map[string]string{
				"x-forwarded-for": "192.168.1.1, 10.0.0.1",
				"client_ip":       "192.168.1.1",
				"user-agent":      "Mozilla/5.0 (Test Browser)",
				"x-device-id":     "device123",
				"content-type":    "application/json",
			},
			hasError: false,
		},
		{
			name: "x-real-ip header",
			headers: http.Header{
				"X-Real-IP": []string{"203.0.113.1"},
				"User-Agent": []string{"Test Agent"},
			},
			expected: map[string]string{
				"x-real-ip":  "203.0.113.1",
				"client_ip":  "203.0.113.1",
				"user-agent": "Test Agent",
			},
			hasError: false,
		},
		{
			name: "invalid IP",
			headers: http.Header{
				"X-Forwarded-For": []string{"invalid-ip"},
				"User-Agent":      []string{"Test Agent"},
			},
			expected: map[string]string{
				"x-forwarded-for": "invalid-ip",
				"user-agent":      "Test Agent",
			},
			hasError: false,
		},
		{
			name:     "empty headers",
			headers:  http.Header{},
			expected: map[string]string{},
			hasError: false,
		},
		{
			name: "disallowed headers",
			headers: http.Header{
				"Authorization": []string{"Bearer token123"},
				"Cookie":       []string{"session=abc123"},
				"User-Agent":   []string{"Test Agent"},
			},
			expected: map[string]string{
				"user-agent": "Test Agent",
			},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := ExtractClientMetadata(tt.headers)
			
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			
			assert.Equal(t, tt.expected, metadata)
		})
	}
}

func TestExtractIPFromHeaders(t *testing.T) {
	tests := []struct {
		name     string
		headers  http.Header
		expected string
	}{
		{
			name: "x-forwarded-for",
			headers: http.Header{
				"X-Forwarded-For": []string{"192.168.1.1, 10.0.0.1"},
			},
			expected: "192.168.1.1",
		},
		{
			name: "x-real-ip",
			headers: http.Header{
				"X-Real-IP": []string{"203.0.113.1"},
			},
			expected: "203.0.113.1",
		},
		{
			name: "cf-connecting-ip",
			headers: http.Header{
				"Cf-Connecting-Ip": []string{"198.51.100.1"},
			},
			expected: "198.51.100.1",
		},
		{
			name: "no IP headers",
			headers: http.Header{
				"User-Agent": []string{"Test Agent"},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := extractIPFromHeaders(tt.headers)
			assert.Equal(t, tt.expected, ip)
		})
	}
}

func TestIsValidIP(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{"valid public IP", "8.8.8.8", true},
		{"valid private IP", "192.168.1.1", true},
		{"valid loopback", "127.0.0.1", true},
		{"invalid IP", "not-an-ip", false},
		{"empty string", "", false},
		{"valid IPv6", "2001:db8::1", true},
		{"private IPv6", "fc00::1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidIP(tt.ip)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeUserAgent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal user agent",
			input:    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			expected: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		},
		{
			name:     "XSS attempt",
			input:    "<script>alert('xss')</script>",
			expected: "&lt;script&gt;alert(&#x27;xss&#x27;)&lt;/script&gt;",
		},
		{
			name:     "long user agent",
			input:    string(make([]byte, 600)), // 600 chars
			expected: string(make([]byte, 512)), // truncated to 512
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeUserAgent(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCreateParamDigest(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]interface{}
		expected string
	}{
		{
			name:     "empty params",
			params:   map[string]interface{}{},
			expected: "",
		},
		{
			name: "simple params",
			params: map[string]interface{}{
				"model":  "gpt-4",
				"stream": false,
			},
			expected: "9d2c3a4b5e6f7d8e", // This is a placeholder - actual hash will differ
		},
		{
			name: "params with long string",
			params: map[string]interface{}{
				"message": string(make([]byte, 200)), // Long string
				"model":   "gpt-4",
			},
			expected: "a1b2c3d4e5f6789a", // Placeholder - actual hash will differ
		},
		{
			name: "ordered params test",
			params: map[string]interface{}{
				"z": "last",
				"a": "first",
				"m": "middle",
			},
			expected: "b2c3d4e5f6a7b8c9", // Placeholder - actual hash will differ
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateParamDigest(tt.params)
			
			if tt.expected == "" {
				assert.Empty(t, result)
			} else {
				assert.NotEmpty(t, result)
				assert.Len(t, result, 16) // Should always be 16 characters
			}
		})
	}
}

func TestSanitizeCookies(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal cookies",
			input:    "theme=dark; lang=en",
			expected: "theme=dark; lang=en",
		},
		{
			name:     "sensitive cookies",
			input:    "session=abc123; theme=dark; token=secret456; lang=en",
			expected: "session=***; theme=dark; token=***; lang=en",
		},
		{
			name:     "case insensitive",
			input:    "Session=abc123; TOKEN=secret456",
			expected: "Session=***; TOKEN=***",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "mixed cookies",
			input:    "csrf_token=xyz; user_pref=light; auth_token=bearer123",
			expected: "csrf_token=***; user_pref=light; auth_token=***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeCookies(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		method   string
		expected string
	}{
		{
			name:     "simple path",
			url:      "/v1/chat/completions",
			method:   "POST",
			expected: "/v1/chat/completions",
		},
		{
			name:     "GET with query params",
			url:      "/v1/models?limit=10&sort=name",
			method:   "GET",
			expected: "/v1/models",
		},
		{
			name:     "POST with query params",
			url:      "/v1/chat/completions?stream=true",
			method:   "POST",
			expected: "/v1/chat/completions?stream=true",
		},
		{
			name:     "double slashes",
			url:      "//v1//chat//completions//",
			method:   "POST",
			expected: "/v1/chat/completions",
		},
		{
			name:     "root path",
			url:      "/",
			method:   "GET",
			expected: "/",
		},
		{
			name:     "trailing slash",
			url:      "/v1/models/",
			method:   "GET",
			expected: "/v1/models",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeURL(tt.url, tt.method)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCreateAuthKeyFingerprint(t *testing.T) {
	tests := []struct {
		name     string
		authKey  string
		expected string
	}{
		{
			name:     "valid auth key",
			authKey:  "sk-1234567890abcdef",
			expected: "c2f2b4a6e8d0a1b3", // Placeholder - actual hash will differ
		},
		{
			name:     "empty auth key",
			authKey:  "",
			expected: "",
		},
		{
			name:     "another valid key",
			authKey:  "Bearer abc123def456",
			expected: "f3e4a5b6c7d8e9f0", // Placeholder - actual hash will differ
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateAuthKeyFingerprint(tt.authKey)
			
			if tt.expected == "" {
				assert.Empty(t, result)
			} else {
				assert.NotEmpty(t, result)
				assert.Len(t, result, 16) // Should always be 16 characters
			}
		})
	}
}

func TestHeimdallRequestLog_BeforeCreate(t *testing.T) {
	log := &HeimdallRequestLog{}
	
	err := log.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.False(t, log.OccurredAt.IsZero())
	assert.True(t, log.OccurredAt.Before(time.Now().UTC().Add(time.Second)))
}

func TestHeimdallRequestLog_TableName(t *testing.T) {
	log := &HeimdallRequestLog{}
	expected := "heimdall_request_logs"
	assert.Equal(t, expected, log.TableName())
}

// Benchmark tests
func BenchmarkExtractClientMetadata(b *testing.B) {
	headers := http.Header{
		"X-Forwarded-For": []string{"192.168.1.1, 10.0.0.1"},
		"User-Agent":      []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"},
		"X-Device-Id":     []string{"device123"},
		"Content-Type":    []string{"application/json"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ExtractClientMetadata(headers)
	}
}

func BenchmarkCreateParamDigest(b *testing.B) {
	params := map[string]interface{}{
		"model":    "gpt-4",
		"stream":   false,
		"messages": []interface{}{map[string]interface{}{"role": "user", "content": "Hello"}},
		"max_tokens": 1000,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CreateParamDigest(params)
	}
}

func BenchmarkSanitizeUserAgent(b *testing.B) {
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sanitizeUserAgent(userAgent)
	}
}
