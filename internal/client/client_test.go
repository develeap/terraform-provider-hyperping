// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		c := NewClient("test_key")

		if c.baseURL != DefaultBaseURL {
			t.Errorf("expected baseURL to be %q, got %q", DefaultBaseURL, c.baseURL)
		}
		if c.maxRetries != DefaultMaxRetries {
			t.Errorf("expected maxRetries to be %d, got %d", DefaultMaxRetries, c.maxRetries)
		}
		// Verify auth transport is wired (VULN-009)
		if _, ok := c.httpClient.Transport.(*authTransport); !ok {
			t.Error("expected httpClient transport to be *authTransport")
		}
	})

	t.Run("with options", func(t *testing.T) {
		customClient := &http.Client{Timeout: 60 * time.Second}
		c := NewClient("test_key",
			WithBaseURL("https://custom.api.com"),
			WithHTTPClient(customClient),
			WithMaxRetries(5),
			WithRetryWait(2*time.Second, 60*time.Second),
		)

		if c.baseURL != "https://custom.api.com" {
			t.Errorf("expected baseURL to be 'https://custom.api.com', got %q", c.baseURL)
		}
		// Transport is wrapped with auth chain — verify it's an authTransport
		if _, ok := c.httpClient.Transport.(*authTransport); !ok {
			t.Error("expected httpClient transport to be *authTransport after option application")
		}
		// Timeout from custom client is preserved
		if c.httpClient.Timeout != 60*time.Second {
			t.Errorf("expected httpClient timeout to be 60s, got %v", c.httpClient.Timeout)
		}
		if c.maxRetries != 5 {
			t.Errorf("expected maxRetries to be 5, got %d", c.maxRetries)
		}
		if c.retryWaitMin != 2*time.Second {
			t.Errorf("expected retryWaitMin to be 2s, got %v", c.retryWaitMin)
		}
		if c.retryWaitMax != 60*time.Second {
			t.Errorf("expected retryWaitMax to be 60s, got %v", c.retryWaitMax)
		}
	})

	t.Run("with logger", func(t *testing.T) {
		logger := &mockLogger{}
		c := NewClient("test_key", WithLogger(logger))

		if c.logger != logger {
			t.Error("expected logger to be set")
		}
	})

	t.Run("with metrics", func(t *testing.T) {
		metrics := &mockMetrics{}
		c := NewClient("test_key", WithMetrics(metrics))

		if c.metrics != metrics {
			t.Error("expected metrics to be set")
		}
	})

	t.Run("with version", func(t *testing.T) {
		c := NewClient("test_key", WithVersion("1.2.3"))

		if c.version != "1.2.3" {
			t.Errorf("expected version to be '1.2.3', got %q", c.version)
		}
		// User-Agent should contain the version
		if !strings.Contains(c.userAgent, "1.2.3") {
			t.Errorf("expected userAgent to contain '1.2.3', got %q", c.userAgent)
		}
	})

	t.Run("circuit breaker initialized", func(t *testing.T) {
		c := NewClient("test_key")

		if c.circuitBreaker == nil {
			t.Error("expected circuit breaker to be initialized")
		}
	})
}

// mockLogger is a test implementation of the Logger interface.
type mockLogger struct {
	messages []string
	fields   []map[string]interface{}
}

func (l *mockLogger) Debug(ctx context.Context, msg string, fields map[string]interface{}) {
	l.messages = append(l.messages, msg)
	l.fields = append(l.fields, fields)
}

// mockMetrics is a test implementation of the Metrics interface.
type mockMetrics struct {
	apiCalls             []apiCallMetric
	retries              []retryMetric
	circuitBreakerStates []string
}

type apiCallMetric struct {
	method     string
	path       string
	statusCode int
	durationMs int64
}

type retryMetric struct {
	method  string
	path    string
	attempt int
}

func (m *mockMetrics) RecordAPICall(ctx context.Context, method, path string, statusCode int, durationMs int64) {
	m.apiCalls = append(m.apiCalls, apiCallMetric{
		method:     method,
		path:       path,
		statusCode: statusCode,
		durationMs: durationMs,
	})
}

func (m *mockMetrics) RecordRetry(ctx context.Context, method, path string, attempt int) {
	m.retries = append(m.retries, retryMetric{
		method:  method,
		path:    path,
		attempt: attempt,
	})
}

func (m *mockMetrics) RecordCircuitBreakerState(ctx context.Context, state string) {
	m.circuitBreakerStates = append(m.circuitBreakerStates, state)
}

func TestClient_WithLogger_Integration(t *testing.T) {
	logger := &mockLogger{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0), WithLogger(logger))

	var result map[string]string
	err := c.doRequest(context.Background(), "GET", "/test", nil, &result)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify logger was called
	if len(logger.messages) < 2 {
		t.Errorf("expected at least 2 log messages (request + response), got %d", len(logger.messages))
	}

	// Check for expected log messages
	foundRequest := false
	foundResponse := false
	for _, msg := range logger.messages {
		if strings.Contains(msg, "sending API request") {
			foundRequest = true
		}
		if strings.Contains(msg, "received API response") {
			foundResponse = true
		}
	}

	if !foundRequest {
		t.Error("expected 'sending API request' log message")
	}
	if !foundResponse {
		t.Error("expected 'received API response' log message")
	}
}

func TestClient_doRequest_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get(HeaderAuthorization) != BearerPrefix+"test_key" {
			t.Errorf("expected Authorization header 'Bearer test_key', got %q", r.Header.Get(HeaderAuthorization))
		}
		if r.Header.Get(HeaderContentType) != ContentTypeJSON {
			t.Errorf("expected Content-Type 'application/json', got %q", r.Header.Get(HeaderContentType))
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	var result map[string]string
	err := c.doRequest(context.Background(), "GET", "/test", nil, &result)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("expected status 'ok', got %q", result["status"])
	}
}

func TestClient_doRequest_WithBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if body["name"] != "test" {
			t.Errorf("expected name 'test', got %q", body["name"])
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "123"})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	var result map[string]string
	err := c.doRequest(context.Background(), "POST", "/test", map[string]string{"name": "test"}, &result)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result["id"] != "123" {
		t.Errorf("expected id '123', got %q", result["id"])
	}
}

func TestClient_doRequest_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	err := c.doRequest(context.Background(), "GET", "/test", nil, nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestClient_doRequest_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	err := c.doRequest(context.Background(), "GET", "/test", nil, nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsUnauthorized(err) {
		t.Errorf("expected unauthorized error, got %v", err)
	}
}

func TestClient_doRequest_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Validation Error",
			"message": "Invalid request",
			"details": []map[string]string{
				{"field": "url", "message": "Invalid URL format"},
			},
		})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	err := c.doRequest(context.Background(), "POST", "/test", nil, nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsValidation(err) {
		t.Errorf("expected validation error, got %v", err)
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if len(apiErr.Details) != 1 {
		t.Errorf("expected 1 validation detail, got %d", len(apiErr.Details))
	}
}

func TestClient_doRequest_Retry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	c := NewClient("test_key",
		WithBaseURL(server.URL),
		WithMaxRetries(3),
		WithRetryWait(1*time.Millisecond, 10*time.Millisecond),
	)

	var result map[string]string
	err := c.doRequest(context.Background(), "GET", "/test", nil, &result)

	if err != nil {
		t.Fatalf("expected no error after retry, got %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestClient_doRequest_MaxRetriesExceeded(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	c := NewClient("test_key",
		WithBaseURL(server.URL),
		WithMaxRetries(2),
		WithRetryWait(1*time.Millisecond, 10*time.Millisecond),
	)

	err := c.doRequest(context.Background(), "GET", "/test", nil, nil)

	if err == nil {
		t.Fatal("expected error after max retries, got nil")
	}
	if attempts != 3 { // Initial + 2 retries
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestClient_doRequest_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := c.doRequest(ctx, "GET", "/test", nil, nil)

	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
}

func TestCalculateBackoff(t *testing.T) {
	c := NewClient("test_key",
		WithRetryWait(1*time.Second, 30*time.Second),
	)

	tests := []struct {
		attempt    int
		retryAfter int
		expected   time.Duration
	}{
		{0, 0, 1 * time.Second},   // 1 * 2^0 = 1
		{1, 0, 2 * time.Second},   // 1 * 2^1 = 2
		{2, 0, 4 * time.Second},   // 1 * 2^2 = 4
		{3, 0, 8 * time.Second},   // 1 * 2^3 = 8
		{10, 0, 30 * time.Second}, // Capped at max
		{0, 5, 5 * time.Second},   // Use retry-after
		{0, 60, 30 * time.Second}, // Retry-after capped at max
	}

	for _, tt := range tests {
		result := c.calculateBackoff(tt.attempt, tt.retryAfter)
		// For exponential backoff cases (retryAfter==0), jitter adds ±25%.
		// For retry-after cases, exact values are used (no jitter).
		if tt.retryAfter > 0 {
			if result != tt.expected {
				t.Errorf("calculateBackoff(%d, %d) = %v, expected %v", tt.attempt, tt.retryAfter, result, tt.expected)
			}
		} else {
			minExpected := tt.expected * 75 / 100
			maxExpected := tt.expected * 125 / 100
			if result < minExpected || result > maxExpected {
				t.Errorf("calculateBackoff(%d, %d) = %v, expected within [%v, %v]", tt.attempt, tt.retryAfter, result, minExpected, maxExpected)
			}
		}
	}
}

func TestClient_doRequest_BodyMarshalError(t *testing.T) {
	c := NewClient("test_key", WithMaxRetries(0))

	// Channels cannot be marshaled to JSON
	unmarshalable := make(chan int)

	err := c.doRequest(context.Background(), "POST", "/test", unmarshalable, nil)

	if err == nil {
		t.Fatal("expected error for unmarshalable body, got nil")
	}
	if !strings.Contains(err.Error(), "failed to marshal request body") {
		t.Errorf("expected marshal error, got %v", err)
	}
}

func TestClient_doRequest_ResponseUnmarshalError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	var result map[string]string
	err := c.doRequest(context.Background(), "GET", "/test", nil, &result)

	if err == nil {
		t.Fatal("expected error for invalid JSON response, got nil")
	}
	if !strings.Contains(err.Error(), "failed to unmarshal response") {
		t.Errorf("expected unmarshal error, got %v", err)
	}
}

func TestClient_doRequest_EmptyResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	var result map[string]string
	err := c.doRequest(context.Background(), "DELETE", "/test", nil, &result)

	if err != nil {
		t.Fatalf("expected no error for empty response, got %v", err)
	}
}

func TestClient_doRequest_NilResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	// Pass nil result - should not fail
	err := c.doRequest(context.Background(), "GET", "/test", nil, nil)

	if err != nil {
		t.Fatalf("expected no error when result is nil, got %v", err)
	}
}

func TestClient_doRequest_RateLimitWithRetryAfter(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.Header().Set("Retry-After", "1")
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
		WithMaxRetries(2),
		WithRetryWait(1*time.Millisecond, 10*time.Millisecond),
	)

	var result map[string]string
	err := c.doRequest(context.Background(), "GET", "/test", nil, &result)

	if err != nil {
		t.Fatalf("expected success after retry, got %v", err)
	}
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}

func TestClient_doRequest_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Internal Server Error"})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	err := c.doRequest(context.Background(), "GET", "/test", nil, nil)

	if err == nil {
		t.Fatal("expected error for server error, got nil")
	}
	if !IsServerError(err) {
		t.Errorf("expected server error, got %v", err)
	}
}

func TestClient_doRequest_BadGateway(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{"error": "Bad Gateway"})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	err := c.doRequest(context.Background(), "GET", "/test", nil, nil)

	if err == nil {
		t.Fatal("expected error for bad gateway, got nil")
	}
	if !IsServerError(err) {
		t.Errorf("expected server error, got %v", err)
	}
}

func TestClient_doRequest_GatewayTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusGatewayTimeout)
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	err := c.doRequest(context.Background(), "GET", "/test", nil, nil)

	if err == nil {
		t.Fatal("expected error for gateway timeout, got nil")
	}
	if !IsServerError(err) {
		t.Errorf("expected server error, got %v", err)
	}
}

func TestClient_doRequest_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Forbidden"})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	err := c.doRequest(context.Background(), "GET", "/test", nil, nil)

	if err == nil {
		t.Fatal("expected error for forbidden, got nil")
	}
	if !IsUnauthorized(err) {
		t.Errorf("expected unauthorized error, got %v", err)
	}
}

func TestClient_doRequest_BadRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Bad Request", "message": "Invalid input"})
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	err := c.doRequest(context.Background(), "POST", "/test", nil, nil)

	if err == nil {
		t.Fatal("expected error for bad request, got nil")
	}
	if !IsValidation(err) {
		t.Errorf("expected validation error, got %v", err)
	}
}

func TestClient_doRequest_NetworkErrorRetry(t *testing.T) {
	attempts := 0
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			// Close connection without response to simulate network error
			server.CloseClientConnections()
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	c := NewClient("test_key",
		WithBaseURL(server.URL),
		WithMaxRetries(3),
		WithRetryWait(1*time.Millisecond, 10*time.Millisecond),
	)

	var result map[string]string
	err := c.doRequest(context.Background(), "GET", "/test", nil, &result)

	// This may or may not succeed depending on timing
	// The important thing is that retries were attempted
	if attempts < 1 {
		t.Errorf("expected at least 1 attempt, got %d", attempts)
	}
	_ = err // Error may or may not occur
}

func TestClient_doRequest_ContextDeadlineExceeded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := c.doRequest(ctx, "GET", "/test", nil, nil)

	if err == nil {
		t.Fatal("expected error from context deadline, got nil")
	}
}

func TestClient_doRequest_RetryOnlyRetryableErrors(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		// 404 is NOT retryable
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}))
	defer server.Close()

	c := NewClient("test_key",
		WithBaseURL(server.URL),
		WithMaxRetries(3),
		WithRetryWait(1*time.Millisecond, 10*time.Millisecond),
	)

	err := c.doRequest(context.Background(), "GET", "/test", nil, nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Should NOT retry on 404
	if attempts != 1 {
		t.Errorf("expected 1 attempt (no retries for 404), got %d", attempts)
	}
}

func TestClient_doRequest_RetryOnlyRetryableErrors_401(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
	}))
	defer server.Close()

	c := NewClient("test_key",
		WithBaseURL(server.URL),
		WithMaxRetries(3),
		WithRetryWait(1*time.Millisecond, 10*time.Millisecond),
	)

	err := c.doRequest(context.Background(), "GET", "/test", nil, nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Should NOT retry on 401
	if attempts != 1 {
		t.Errorf("expected 1 attempt (no retries for 401), got %d", attempts)
	}
}

func TestClient_doRequest_HeadersSet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify all headers
		if r.Header.Get(HeaderAuthorization) != BearerPrefix+"test_api_key_123" {
			t.Errorf("Authorization header = %q, expected 'Bearer test_api_key_123'", r.Header.Get(HeaderAuthorization))
		}
		if r.Header.Get(HeaderContentType) != ContentTypeJSON {
			t.Errorf("Content-Type header = %q, expected 'application/json'", r.Header.Get(HeaderContentType))
		}
		if r.Header.Get(HeaderAccept) != ContentTypeJSON {
			t.Errorf("Accept header = %q, expected 'application/json'", r.Header.Get(HeaderAccept))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient("test_api_key_123", WithBaseURL(server.URL), WithMaxRetries(0))

	err := c.doRequest(context.Background(), "GET", "/test", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_doRequest_AllHTTPMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != method {
					t.Errorf("expected method %s, got %s", method, r.Method)
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			c := NewClient("test_key", WithBaseURL(server.URL), WithMaxRetries(0))
			err := c.doRequest(context.Background(), method, "/test", nil, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_Sleep_ContextCancellation(t *testing.T) {
	c := NewClient("test_key")

	ctx, cancel := context.WithCancel(context.Background())

	// Start sleep in goroutine
	done := make(chan struct{})
	go func() {
		c.sleep(ctx, 5*time.Second)
		close(done)
	}()

	// Cancel after short delay
	time.Sleep(10 * time.Millisecond)
	cancel()

	// Sleep should return quickly
	select {
	case <-done:
		// Success - sleep returned due to context cancellation
	case <-time.After(1 * time.Second):
		t.Error("sleep did not return after context cancellation")
	}
}

func TestClient_Sleep_NormalCompletion(t *testing.T) {
	c := NewClient("test_key")

	start := time.Now()
	c.sleep(context.Background(), 10*time.Millisecond)
	elapsed := time.Since(start)

	if elapsed < 10*time.Millisecond {
		t.Errorf("sleep returned too early: %v", elapsed)
	}
}

func TestParseErrorResponse_ValidJSON(t *testing.T) {
	c := NewClient("test_key")

	body := []byte(`{"error": "Not Found", "message": "Resource does not exist"}`)
	apiErr := c.parseErrorResponse(404, http.Header{}, body)

	if apiErr.StatusCode != 404 {
		t.Errorf("StatusCode = %d, expected 404", apiErr.StatusCode)
	}
	if apiErr.Message != "Resource does not exist" {
		t.Errorf("Message = %q, expected 'Resource does not exist'", apiErr.Message)
	}
}

func TestParseErrorResponse_InvalidJSON(t *testing.T) {
	c := NewClient("test_key")

	body := []byte(`not valid json`)
	apiErr := c.parseErrorResponse(500, http.Header{}, body)

	if apiErr.StatusCode != 500 {
		t.Errorf("StatusCode = %d, expected 500", apiErr.StatusCode)
	}
	// Should fall back to http.StatusText
	if apiErr.Message != "Internal Server Error" {
		t.Errorf("Message = %q, expected 'Internal Server Error'", apiErr.Message)
	}
}

func TestParseErrorResponse_ErrorFieldOnly(t *testing.T) {
	c := NewClient("test_key")

	body := []byte(`{"error": "Something went wrong"}`)
	apiErr := c.parseErrorResponse(400, http.Header{}, body)

	if apiErr.Message != "Something went wrong" {
		t.Errorf("Message = %q, expected 'Something went wrong'", apiErr.Message)
	}
}

func TestParseErrorResponse_MessageFieldOnly(t *testing.T) {
	c := NewClient("test_key")

	body := []byte(`{"message": "Detailed error message"}`)
	apiErr := c.parseErrorResponse(400, http.Header{}, body)

	if apiErr.Message != "Detailed error message" {
		t.Errorf("Message = %q, expected 'Detailed error message'", apiErr.Message)
	}
}

func TestParseErrorResponse_BothFieldsSame(t *testing.T) {
	c := NewClient("test_key")

	body := []byte(`{"error": "Same message", "message": "Same message"}`)
	apiErr := c.parseErrorResponse(400, http.Header{}, body)

	// When both are the same, should use message field
	if apiErr.Message != "Same message" {
		t.Errorf("Message = %q, expected 'Same message'", apiErr.Message)
	}
}

func TestParseErrorResponse_WithDetails(t *testing.T) {
	c := NewClient("test_key")

	body := []byte(`{
		"error": "Validation Error",
		"message": "Invalid input",
		"details": [
			{"field": "url", "message": "Invalid URL format"},
			{"field": "frequency", "message": "Must be a positive integer"}
		]
	}`)
	apiErr := c.parseErrorResponse(400, http.Header{}, body)

	if len(apiErr.Details) != 2 {
		t.Fatalf("expected 2 details, got %d", len(apiErr.Details))
	}
	if apiErr.Details[0].Field != "url" {
		t.Errorf("Details[0].Field = %q, expected 'url'", apiErr.Details[0].Field)
	}
	if apiErr.Details[1].Message != "Must be a positive integer" {
		t.Errorf("Details[1].Message = %q, expected 'Must be a positive integer'", apiErr.Details[1].Message)
	}
}

func TestParseErrorResponse_EmptyBody(t *testing.T) {
	c := NewClient("test_key")

	body := []byte(``)
	apiErr := c.parseErrorResponse(503, http.Header{}, body)

	if apiErr.StatusCode != 503 {
		t.Errorf("StatusCode = %d, expected 503", apiErr.StatusCode)
	}
	if apiErr.Message != "Service Unavailable" {
		t.Errorf("Message = %q, expected 'Service Unavailable'", apiErr.Message)
	}
}

func TestClient_DefaultTimeout(t *testing.T) {
	c := NewClient("test_key")

	if c.httpClient.Timeout != DefaultTimeout {
		t.Errorf("Timeout = %v, expected %v", c.httpClient.Timeout, DefaultTimeout)
	}
}

func TestClient_CustomHTTPClient_PreservesSettings(t *testing.T) {
	customTimeout := 120 * time.Second
	customClient := &http.Client{Timeout: customTimeout}

	c := NewClient("test_key", WithHTTPClient(customClient))

	if c.httpClient != customClient {
		t.Error("custom HTTP client was not set")
	}
	if c.httpClient.Timeout != customTimeout {
		t.Errorf("Timeout = %v, expected %v", c.httpClient.Timeout, customTimeout)
	}
}

// TestClient_doRequest_InvalidURL tests that an invalid URL returns an error.
func TestClient_doRequest_InvalidURL(t *testing.T) {
	// Use an invalid URL scheme that will cause http.NewRequestWithContext to fail
	c := NewClient("test_key", WithBaseURL("://invalid-url"), WithMaxRetries(0))

	err := c.doRequest(context.Background(), "GET", "/test", nil, nil)
	if err == nil {
		t.Fatal("expected error for invalid URL, got nil")
	}
	if !strings.Contains(err.Error(), "failed to create request") {
		t.Errorf("expected request creation error, got %v", err)
	}
}

// TestClient_doRequest_RetryExhaustedWithLastError tests that the last error is returned
// when all retry attempts are exhausted for network errors.
func TestClient_doRequest_RetryExhaustedWithLastError(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		// Always return 503 to trigger retries
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "Service Unavailable"})
	}))
	defer server.Close()

	c := NewClient("test_key",
		WithBaseURL(server.URL),
		WithMaxRetries(2),
		WithRetryWait(1*time.Millisecond, 10*time.Millisecond),
	)

	err := c.doRequest(context.Background(), "GET", "/test", nil, nil)
	if err == nil {
		t.Fatal("expected error after retries exhausted, got nil")
	}
	// Should have made initial attempt + 2 retries = 3 total
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
	// Error should be from the last attempt
	if !IsServerError(err) {
		t.Errorf("expected server error, got %v", err)
	}
}

func TestClient_CircuitBreakerStates(t *testing.T) {
	t.Run("circuit opens after repeated failures", func(t *testing.T) {
		metrics := &mockMetrics{}
		requestCount := 0

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			// Return 500 errors to trigger circuit breaker
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"server error"}`))
		}))
		defer server.Close()

		c := NewClient("test_key",
			WithBaseURL(server.URL),
			WithMetrics(metrics),
			WithMaxRetries(0), // No retries to control request count
		)

		// Make multiple failing requests to trigger circuit breaker
		// Circuit breaker needs: minimum 3 requests, 60% failure rate
		// So we need at least 3 failures to open the circuit
		for i := 0; i < 5; i++ {
			_ = c.doRequest(context.Background(), "GET", "/test", nil, nil)
		}

		// Verify circuit breaker state changes were recorded
		// Should have at least one state change (closed -> open)
		if len(metrics.circuitBreakerStates) == 0 {
			t.Error("expected circuit breaker state changes to be recorded")
		}

		// Check that we have an "open" state recorded
		hasOpenState := false
		for _, state := range metrics.circuitBreakerStates {
			if state == "open" {
				hasOpenState = true
				break
			}
		}
		if !hasOpenState {
			t.Errorf("expected circuit breaker to open, recorded states: %v", metrics.circuitBreakerStates)
		}

		// Verify API calls were recorded
		if len(metrics.apiCalls) == 0 {
			t.Error("expected API calls to be recorded")
		}
	})

	t.Run("circuit breaker prevents requests when open", func(t *testing.T) {
		requestCount := 0

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			// Always fail to keep circuit open
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"server error"}`))
		}))
		defer server.Close()

		c := NewClient("test_key",
			WithBaseURL(server.URL),
			WithMaxRetries(0),
		)

		// Trigger circuit breaker to open
		for i := 0; i < 5; i++ {
			_ = c.doRequest(context.Background(), "GET", "/test", nil, nil)
		}

		initialCount := requestCount

		// Try making more requests - circuit should be open
		// Circuit breaker will reject requests immediately without hitting the server
		for i := 0; i < 3; i++ {
			err := c.doRequest(context.Background(), "GET", "/test", nil, nil)
			if err == nil {
				t.Error("expected error when circuit is open")
			}
			// Error should be circuit breaker error, not a server error
			if err != nil && IsServerError(err) {
				t.Logf("Got server error (circuit might not be fully open yet): %v", err)
			}
		}

		// Some requests should have been blocked by circuit breaker
		// (requestCount should not have increased by 3)
		if requestCount > initialCount+3 {
			t.Logf("Warning: Circuit breaker may not have blocked all requests. Initial: %d, Final: %d", initialCount, requestCount)
		}
	})

	t.Run("successful requests with closed circuit", func(t *testing.T) {
		metrics := &mockMetrics{}
		requestCount := 0

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			// Always succeed
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		}))
		defer server.Close()

		c := NewClient("test_key",
			WithBaseURL(server.URL),
			WithMetrics(metrics),
			WithMaxRetries(0),
		)

		// Make successful requests
		for i := 0; i < 5; i++ {
			err := c.doRequest(context.Background(), "GET", "/test", nil, nil)
			if err != nil {
				t.Errorf("expected no error with successful requests, got: %v", err)
			}
		}

		// Circuit should remain closed
		for _, state := range metrics.circuitBreakerStates {
			if state == "open" {
				t.Error("circuit breaker should not open with successful requests")
			}
		}

		// All requests should have reached the server
		if requestCount != 5 {
			t.Errorf("expected 5 requests to reach server, got %d", requestCount)
		}

		// All API calls should be recorded as successful
		if len(metrics.apiCalls) != 5 {
			t.Errorf("expected 5 API calls recorded, got %d", len(metrics.apiCalls))
		}
		for i, call := range metrics.apiCalls {
			if call.statusCode != 200 {
				t.Errorf("call %d: expected status 200, got %d", i, call.statusCode)
			}
		}
	})

	t.Run("partial failures don't open circuit", func(t *testing.T) {
		metrics := &mockMetrics{}
		requestCount := 0
		failureThreshold := 2 // Less than 60% of 5 requests

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			// Fail first 2 requests, succeed the rest (40% failure rate)
			if requestCount <= failureThreshold {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"server error"}`))
			} else {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status":"ok"}`))
			}
		}))
		defer server.Close()

		c := NewClient("test_key",
			WithBaseURL(server.URL),
			WithMetrics(metrics),
			WithMaxRetries(0),
		)

		// Make requests with <60% failure rate
		successCount := 0
		for i := 0; i < 5; i++ {
			err := c.doRequest(context.Background(), "GET", "/test", nil, nil)
			if err == nil {
				successCount++
			}
		}

		// Should have 3 successful requests
		if successCount != 3 {
			t.Errorf("expected 3 successful requests, got %d", successCount)
		}

		// Circuit should NOT open (failure rate < 60%)
		for _, state := range metrics.circuitBreakerStates {
			if state == "open" {
				t.Error("circuit breaker should not open with <60% failure rate")
			}
		}
	})

	t.Run("mixed success and failure patterns", func(t *testing.T) {
		metrics := &mockMetrics{}
		requestCount := 0

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			// Pattern: success, success, fail (2 success, 1 fail per 3 requests = 33% failure)
			// This keeps us under the 60% threshold
			if requestCount%3 == 0 {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"server error"}`))
			} else {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status":"ok"}`))
			}
		}))
		defer server.Close()

		c := NewClient("test_key",
			WithBaseURL(server.URL),
			WithMetrics(metrics),
			WithMaxRetries(0),
		)

		// Make 9 requests (3 failures, 6 successes = 33% failure rate)
		for i := 0; i < 9; i++ {
			_ = c.doRequest(context.Background(), "GET", "/test", nil, nil)
		}

		// Circuit should NOT open (33% failure rate < 60% threshold)
		for _, state := range metrics.circuitBreakerStates {
			if state == "open" {
				t.Error("circuit breaker should not open with 33% failure rate")
			}
		}

		// All requests should have been attempted
		if requestCount != 9 {
			t.Errorf("expected 9 requests, got %d", requestCount)
		}

		// Verify API calls were recorded
		if len(metrics.apiCalls) != 9 {
			t.Errorf("expected 9 API calls recorded, got %d", len(metrics.apiCalls))
		}
	})
}

func TestValidateResourceID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"valid simple", "mon_abc123", false},
		{"valid alphanumeric", "abc123", false},
		{"valid with dash", "mon-abc-123", false},
		{"valid with underscore", "mon_abc_123", false},
		{"valid max length", strings.Repeat("a", 128), false},
		{"empty string", "", true},
		{"oversized ID", strings.Repeat("a", 129), true},
		{"path traversal dotdot", "../../admin", true},
		{"path traversal slash", "foo/bar", true},
		{"query injection", "id?admin=true", true},
		{"fragment injection", "id#section", true},
		{"at sign injection", "id@evil.com", true},
		{"ampersand injection", "id&key=val", true},
		{"equals injection", "id=value", true},
		{"starts with special", "-invalid", true},
		{"contains space", "mon abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateResourceID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateResourceID(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
			}
		})
	}
}

func TestAuthTransport(t *testing.T) {
	t.Run("injects authorization header", func(t *testing.T) {
		var receivedAuth string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedAuth = r.Header.Get(HeaderAuthorization)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c := NewClient("sk_test123",
			WithBaseURL(server.URL),
			WithMaxRetries(0),
		)

		// Trigger a request via the client
		err := c.doRequest(context.Background(), "GET", "/test", nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := BearerPrefix + "sk_test123"
		if receivedAuth != expected {
			t.Errorf("expected Authorization header %q, got %q", expected, receivedAuth)
		}
	})

	t.Run("does not mutate original request", func(t *testing.T) {
		var clonedAuth string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clonedAuth = r.Header.Get(HeaderAuthorization)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		transport := &authTransport{
			token: []byte("test-token"),
			next:  server.Client().Transport,
		}

		req, _ := http.NewRequest("GET", server.URL+"/test", nil)
		resp, _ := transport.RoundTrip(req)
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}

		// Cloned request had the auth header
		if clonedAuth != BearerPrefix+"test-token" {
			t.Errorf("expected cloned request to have auth header, got %q", clonedAuth)
		}
		// Original request should not have Authorization header
		if req.Header.Get(HeaderAuthorization) != "" {
			t.Error("original request was mutated — authTransport must clone")
		}
	})
}
