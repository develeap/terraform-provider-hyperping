// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestClient_RequestDurationLogging verifies that request duration is logged
func TestClient_RequestDurationLogging(t *testing.T) {
	logger := &mockLogger{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate some processing time
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	c := NewClient("test_key",
		WithBaseURL(server.URL),
		WithLogger(logger),
	)

	var result map[string]string
	err := c.doRequest(context.Background(), "GET", "/test", nil, &result)
	if err != nil {
		t.Fatalf("doRequest() failed: %v", err)
	}

	// Verify duration_ms was logged (find the "received API response" message)
	var loggedFields map[string]interface{}
	for i, msg := range logger.messages {
		if msg == "received API response" {
			loggedFields = logger.fields[i]
			break
		}
	}

	if loggedFields == nil {
		t.Fatal("No 'received API response' log found")
	}

	durationMS, ok := loggedFields["duration_ms"].(int64)
	if !ok {
		t.Errorf("duration_ms not found in log fields or wrong type: %+v", loggedFields)
	}

	// Duration should be at least 10ms (we slept for 10ms)
	if durationMS < 10 {
		t.Errorf("duration_ms = %d, expected >= 10ms", durationMS)
	}

	// Duration should be reasonable (less than 1 second for local test)
	if durationMS > 1000 {
		t.Errorf("duration_ms = %d, expected < 1000ms", durationMS)
	}
}

// TestClient_RequestDurationLogging_OnError verifies duration is logged even on errors
func TestClient_RequestDurationLogging_OnError(t *testing.T) {
	logger := &mockLogger{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Millisecond)
		// Close connection abruptly to cause error
		hj, _ := w.(http.Hijacker)
		conn, _, _ := hj.Hijack()
		conn.Close()
	}))
	defer server.Close()

	c := NewClient("test_key",
		WithBaseURL(server.URL),
		WithLogger(logger),
		WithMaxRetries(0), // Don't retry to make test faster
	)

	var result map[string]string
	_ = c.doRequest(context.Background(), "GET", "/test", nil, &result)

	// Verify duration_ms was logged even on error
	var loggedFields map[string]interface{}
	for i, msg := range logger.messages {
		if msg == "request failed" {
			loggedFields = logger.fields[i]
			break
		}
	}

	if loggedFields == nil {
		t.Fatal("No 'request failed' log found")
	}

	durationMS, ok := loggedFields["duration_ms"].(int64)
	if !ok {
		t.Errorf("duration_ms not found in error log fields: %+v", loggedFields)
	}

	// Duration should be at least 5ms
	if durationMS < 5 {
		t.Errorf("duration_ms = %d, expected >= 5ms", durationMS)
	}
}

// TestAPIError_EnhancedRateLimitMessage verifies enhanced rate limit error messages
func TestAPIError_EnhancedRateLimitMessage(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		message        string
		retryAfter     int
		expectedSubstr string
	}{
		{
			name:           "rate limit with retry-after",
			statusCode:     429,
			message:        "rate limit exceeded",
			retryAfter:     60,
			expectedSubstr: "retry after 60 seconds",
		},
		{
			name:           "rate limit with longer retry-after",
			statusCode:     429,
			message:        "Too Many Requests",
			retryAfter:     300,
			expectedSubstr: "retry after 300 seconds",
		},
		{
			name:           "rate limit without retry-after",
			statusCode:     429,
			message:        "rate limit exceeded",
			retryAfter:     0,
			expectedSubstr: "rate limit exceeded",
		},
		{
			name:           "non-rate-limit error",
			statusCode:     500,
			message:        "internal server error",
			retryAfter:     60, // Should be ignored for non-429
			expectedSubstr: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiErr := &APIError{
				StatusCode: tt.statusCode,
				Message:    tt.message,
				RetryAfter: tt.retryAfter,
			}

			errorString := apiErr.Error()

			if !strings.Contains(errorString, tt.expectedSubstr) {
				t.Errorf("Error() = %q, should contain %q", errorString, tt.expectedSubstr)
			}

			// For rate limit with retry-after, verify format
			if tt.statusCode == 429 && tt.retryAfter > 0 {
				if !strings.Contains(errorString, "retry after") {
					t.Errorf("Error() = %q, should contain 'retry after'", errorString)
				}
				if !strings.Contains(errorString, "seconds") {
					t.Errorf("Error() = %q, should contain 'seconds'", errorString)
				}
			}
		})
	}
}

// TestClient_TLSMinVersion verifies TLS 1.2+ is enforced on the base transport
func TestClient_TLSMinVersion(t *testing.T) {
	c := NewClient("test_key")

	// Transport chain: authTransport → http.Transport
	auth, ok := c.httpClient.Transport.(*authTransport)
	if !ok {
		t.Fatal("HTTP client transport is not *authTransport")
	}
	transport, ok := auth.next.(*http.Transport)
	if !ok {
		t.Fatal("authTransport.next is not *http.Transport")
	}

	if transport.TLSClientConfig == nil {
		t.Fatal("TLS config is nil")
	}

	if transport.TLSClientConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("TLS MinVersion = %d, want %d (TLS 1.2)", transport.TLSClientConfig.MinVersion, tls.VersionTLS12)
	}
}

// TestClient_TLSMinVersion_PreservedWithCustomClient verifies custom HTTPS clients
// get TLS enforcement applied (VULN-011: WithHTTPClient must not bypass TLS).
func TestClient_TLSMinVersion_PreservedWithCustomClient(t *testing.T) {
	customTransport := &http.Transport{}
	customClient := &http.Client{
		Transport: customTransport,
		Timeout:   5 * time.Second,
	}

	c := NewClient("test_key", WithHTTPClient(customClient))

	// Transport chain: authTransport → http.Transport (with TLS enforced)
	auth, ok := c.httpClient.Transport.(*authTransport)
	if !ok {
		t.Fatal("HTTP client transport is not *authTransport")
	}
	transport, ok := auth.next.(*http.Transport)
	if !ok {
		t.Fatal("authTransport.next is not *http.Transport")
	}

	// TLS config must be applied even on custom transports (VULN-011)
	if transport.TLSClientConfig == nil {
		t.Fatal("TLS config is nil — custom transport TLS enforcement failed")
	}
	if transport.TLSClientConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("TLS MinVersion = %d, want %d (TLS 1.2)", transport.TLSClientConfig.MinVersion, tls.VersionTLS12)
	}
}

// TestClient_DurationLogging_MultipleRequests verifies duration logging across retries
func TestClient_DurationLogging_MultipleRequests(t *testing.T) {
	attempts := 0
	logger := &mockLogger{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		time.Sleep(5 * time.Millisecond)
		if attempts < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{"error": "Rate limited"})
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	c := NewClient("test_key",
		WithBaseURL(server.URL),
		WithLogger(logger),
		WithMaxRetries(2),
		WithRetryWait(1*time.Millisecond, 10*time.Millisecond),
	)

	var result map[string]string
	err := c.doRequest(context.Background(), "GET", "/test", nil, &result)
	if err != nil {
		t.Fatalf("doRequest() failed: %v", err)
	}

	// Extract logged durations from fields
	var loggedDurations []int64
	for i, msg := range logger.messages {
		if msg == "received API response" || msg == "request failed" {
			if durationMS, ok := logger.fields[i]["duration_ms"].(int64); ok {
				loggedDurations = append(loggedDurations, durationMS)
			}
		}
	}

	// Should have logged duration for both attempts
	if len(loggedDurations) != 2 {
		t.Errorf("Expected 2 duration logs, got %d", len(loggedDurations))
	}

	// All durations should be >= 5ms (our sleep time)
	for i, duration := range loggedDurations {
		if duration < 5 {
			t.Errorf("Attempt %d: duration_ms = %d, expected >= 5ms", i+1, duration)
		}
	}
}

// TestAPIError_EnhancedMessage_Integration verifies enhanced messages in real scenarios
func TestAPIError_EnhancedMessage_Integration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "120")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Rate limit exceeded. Please try again later.",
		})
	}))
	defer server.Close()

	c := NewClient("test_key",
		WithBaseURL(server.URL),
		WithMaxRetries(0), // Don't retry
	)

	var result map[string]string
	err := c.doRequest(context.Background(), "GET", "/test", nil, &result)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	errorString := err.Error()

	// Should contain enhanced rate limit message
	expectedSubstrings := []string{
		"retry after",
		"120 seconds",
		"status 429",
	}

	for _, expected := range expectedSubstrings {
		if !strings.Contains(errorString, expected) {
			t.Errorf("Error message %q should contain %q", errorString, expected)
		}
	}
}
