// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"net/http"
	"testing"
	"time"
)

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "integer seconds - valid",
			input:    "120",
			expected: 120,
		},
		{
			name:     "integer seconds - with whitespace",
			input:    "  60  ",
			expected: 60,
		},
		{
			name:     "integer seconds - zero",
			input:    "0",
			expected: 0,
		},
		{
			name:     "integer seconds - negative",
			input:    "-10",
			expected: 0,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "HTTP-date format - future time",
			input:    time.Now().Add(90 * time.Second).UTC().Format(http.TimeFormat),
			expected: 89, // Allow 1 second tolerance for test execution time
		},
		{
			name:     "HTTP-date format - past time",
			input:    time.Now().Add(-30 * time.Second).UTC().Format(http.TimeFormat),
			expected: 0,
		},
		{
			name:     "invalid format",
			input:    "not-a-number",
			expected: 0,
		},
		{
			name:     "invalid date format",
			input:    "2025-01-26 10:30:00",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRetryAfter(tt.input)

			// For HTTP-date tests, allow some tolerance due to execution time
			if tt.name == "HTTP-date format - future time" {
				// Result should be between 85-95 seconds (90 seconds ± 5 second tolerance)
				if result < 85 || result > 95 {
					t.Errorf("parseRetryAfter() = %d, expected ~%d (±5s)", result, tt.expected)
				}
			} else {
				if result != tt.expected {
					t.Errorf("parseRetryAfter() = %d, want %d", result, tt.expected)
				}
			}
		})
	}
}

func TestParseRetryAfter_HTTPDateFormats(t *testing.T) {
	// Test all three HTTP-date formats that http.ParseTime supports
	futureTime := time.Now().Add(60 * time.Second).UTC()

	formats := []struct {
		name    string
		dateStr string
	}{
		{
			name:    "HTTP TimeFormat (RFC1123 with GMT)",
			dateStr: futureTime.Format(http.TimeFormat),
		},
		{
			name:    "RFC850 format",
			dateStr: futureTime.Format(time.RFC850),
		},
		{
			name:    "ANSI C format",
			dateStr: futureTime.Format(time.ANSIC),
		},
	}

	for _, tt := range formats {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRetryAfter(tt.dateStr)

			// Should be approximately 60 seconds (allow ±5 second tolerance)
			if result < 55 || result > 65 {
				t.Errorf("parseRetryAfter(%s) = %d, expected ~60 seconds (±5s)", tt.name, result)
			}
		})
	}
}

func TestParseRetryAfter_Integration_WithAPIError(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		retryAfterHdr string
		expectedRetry int
	}{
		{
			name:          "429 with Retry-After header",
			statusCode:    429,
			retryAfterHdr: "30",
			expectedRetry: 30,
		},
		{
			name:          "429 without Retry-After header",
			statusCode:    429,
			retryAfterHdr: "",
			expectedRetry: 0,
		},
		{
			name:          "500 with Retry-After header (should be ignored)",
			statusCode:    500,
			retryAfterHdr: "60",
			expectedRetry: 0, // Only 429 responses parse Retry-After
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient("test_key")

			// Create mock HTTP headers
			headers := make(map[string][]string)
			if tt.retryAfterHdr != "" {
				headers["Retry-After"] = []string{tt.retryAfterHdr}
			}

			apiErr := c.parseErrorResponse(tt.statusCode, headers, []byte(`{"error":"test"}`))

			if apiErr.RetryAfter != tt.expectedRetry {
				t.Errorf("parseErrorResponse() RetryAfter = %d, want %d", apiErr.RetryAfter, tt.expectedRetry)
			}
		})
	}
}

func TestCalculateBackoff_WithRetryAfter(t *testing.T) {
	tests := []struct {
		name         string
		attempt      int
		retryAfter   int
		retryWaitMin time.Duration
		retryWaitMax time.Duration
		expectedWait time.Duration
	}{
		{
			name:         "Retry-After provided and within max",
			attempt:      0,
			retryAfter:   5,
			retryWaitMin: 1 * time.Second,
			retryWaitMax: 30 * time.Second,
			expectedWait: 5 * time.Second,
		},
		{
			name:         "Retry-After exceeds max wait",
			attempt:      0,
			retryAfter:   60,
			retryWaitMin: 1 * time.Second,
			retryWaitMax: 30 * time.Second,
			expectedWait: 30 * time.Second,
		},
		{
			name:         "Retry-After zero, uses exponential backoff",
			attempt:      2, // 2^2 = 4x multiplier
			retryAfter:   0,
			retryWaitMin: 1 * time.Second,
			retryWaitMax: 30 * time.Second,
			expectedWait: 4 * time.Second,
		},
		{
			name:         "Retry-After negative, uses exponential backoff",
			attempt:      1, // 2^1 = 2x multiplier
			retryAfter:   -10,
			retryWaitMin: 2 * time.Second,
			retryWaitMax: 30 * time.Second,
			expectedWait: 4 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient("test_key",
				WithRetryWait(tt.retryWaitMin, tt.retryWaitMax),
			)

			wait := c.calculateBackoff(tt.attempt, tt.retryAfter)

			// Retry-After > 0 uses exact value; otherwise jitter adds ±25%.
			if tt.retryAfter > 0 {
				if wait != tt.expectedWait {
					t.Errorf("calculateBackoff() = %v, want %v", wait, tt.expectedWait)
				}
			} else {
				minWait := tt.expectedWait * 75 / 100
				maxWait := tt.expectedWait * 125 / 100
				if wait < minWait || wait > maxWait {
					t.Errorf("calculateBackoff() = %v, want within [%v, %v]", wait, minWait, maxWait)
				}
			}
		})
	}
}

// Benchmark parseRetryAfter with different input types
func BenchmarkParseRetryAfter_Integer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = parseRetryAfter("120")
	}
}

func BenchmarkParseRetryAfter_HTTPDate(b *testing.B) {
	dateStr := time.Now().Add(60 * time.Second).UTC().Format(time.RFC1123)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parseRetryAfter(dateStr)
	}
}

func BenchmarkParseRetryAfter_Invalid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = parseRetryAfter("invalid-input")
	}
}
