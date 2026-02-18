// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"testing"
)

func TestSanitizeMessage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "hyperping API key in message",
			input:    "Invalid API key: sk_abc123xyz456",
			expected: "Invalid API key: sk_***REDACTED***",
		},
		{
			name:     "multiple hyperping API keys",
			input:    "Keys sk_key1 and sk_key2 are invalid",
			expected: "Keys sk_***REDACTED*** and sk_***REDACTED*** are invalid",
		},
		{
			name:     "Bearer token in message",
			input:    "Authentication failed with Bearer abcd1234efgh5678",
			expected: "Authentication failed with Bearer ***REDACTED***",
		},
		{ //nolint:gosec // G101: test string with placeholder credentials, not real secrets
			name:     "URL with credentials",
			input:    "Failed to connect to https://user:password@api.hyperping.io",
			expected: "Failed to connect to https://***REDACTED***@api.hyperping.io",
		},
		{
			name:     "Authorization header",
			input:    "Request failed: Authorization: Bearer sk_secret123",
			expected: "Request failed: Authorization: ***REDACTED***",
		},
		{ //nolint:gosec // G101: test string with placeholder credentials, not real secrets
			name:     "combined sensitive data",
			input:    "API call with sk_key123 and Bearer token456 failed at https://user:pass@example.com",
			expected: "API call with sk_***REDACTED*** and Bearer ***REDACTED*** failed at https://***REDACTED***@example.com",
		},
		{
			name:     "no sensitive data",
			input:    "Resource not found",
			expected: "Resource not found",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "hp prefix but not API key format",
			input:    "Hyperping service is down",
			expected: "Hyperping service is down",
		},
		{
			name:     "bearer without token",
			input:    "Using Bearer authentication",
			expected: "Using Bearer authentication",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeMessage(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeMessage() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestAPIError_Error_Sanitization(t *testing.T) {
	tests := []struct {
		name             string
		apiError         *APIError
		shouldContain    string
		shouldNotContain string
	}{
		{
			name: "API key in error message is sanitized",
			apiError: &APIError{
				StatusCode: 401,
				Message:    "Invalid API key: sk_secret123",
			},
			shouldContain:    "sk_***REDACTED***",
			shouldNotContain: "sk_secret123",
		},
		{
			name: "Bearer token in error message is sanitized",
			apiError: &APIError{
				StatusCode: 403,
				Message:    "Authorization failed: Bearer abc123xyz",
			},
			shouldContain:    "Bearer ***REDACTED***",
			shouldNotContain: "Bearer abc123xyz",
		},
		{
			name: "URL credentials in error message are sanitized",
			apiError: &APIError{
				StatusCode: 500,
				Message:    "Connection error to https://admin:pass@api.hyperping.io",
			},
			shouldContain:    "https://***REDACTED***@api.hyperping.io",
			shouldNotContain: "admin:pass",
		},
		{
			name: "validation error with sensitive data",
			apiError: &APIError{
				StatusCode: 422,
				Message:    "Validation failed for sk_key123",
				Details: []ValidationDetail{
					{Field: "api_key", Message: "Invalid format"},
				},
			},
			shouldContain:    "sk_***REDACTED***",
			shouldNotContain: "sk_key123",
		},
		{
			name: "clean error message unchanged",
			apiError: &APIError{
				StatusCode: 404,
				Message:    "Monitor not found",
			},
			shouldContain:    "Monitor not found",
			shouldNotContain: "***REDACTED***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorString := tt.apiError.Error()

			if tt.shouldContain != "" {
				if !contains(errorString, tt.shouldContain) {
					t.Errorf("Error() = %q, should contain %q", errorString, tt.shouldContain)
				}
			}

			if tt.shouldNotContain != "" {
				if contains(errorString, tt.shouldNotContain) {
					t.Errorf("Error() = %q, should NOT contain %q", errorString, tt.shouldNotContain)
				}
			}
		})
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark sanitization performance
func BenchmarkSanitizeMessage(b *testing.B) {
	testMessage := "Authentication failed: Invalid API key sk_abc123xyz Bearer token456 at https://user:pass@api.hyperping.io"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sanitizeMessage(testMessage)
	}
}

func BenchmarkSanitizeMessage_NoSensitiveData(b *testing.B) {
	testMessage := "Resource not found at /api/v1/monitors/mon_123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sanitizeMessage(testMessage)
	}
}
