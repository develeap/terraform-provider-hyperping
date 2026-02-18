// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

// mockRoundTripper is a mock implementation of http.RoundTripper for testing.
type mockRoundTripper struct {
	response        *http.Response
	err             error
	capturedRequest *http.Request
	callCount       int
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	m.capturedRequest = req
	m.callCount++
	return m.response, m.err
}

// TestAuthTransport_RoundTrip tests the authTransport.RoundTrip function comprehensively.
func TestAuthTransport_RoundTrip(t *testing.T) {
	tests := []struct {
		name          string
		token         []byte
		setupRequest  func() *http.Request
		mockResponse  *http.Response
		mockError     error
		validateAuth  string
		expectError   bool
		errorContains string
	}{
		{
			name:  "adds authorization header to request",
			token: []byte("test_api_key_123"),
			setupRequest: func() *http.Request {
				req, _ := http.NewRequest("GET", "https://api.example.com/v1/monitors", nil)
				return req
			},
			mockResponse: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
			},
			validateAuth: BearerPrefix + "test_api_key_123",
			expectError:  false,
		},
		{
			name:  "preserves existing headers",
			token: []byte("another_key"),
			setupRequest: func() *http.Request {
				req, _ := http.NewRequest("POST", "https://api.example.com/v1/incidents", strings.NewReader(`{"title":"test"}`))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-Custom-Header", "custom-value")
				return req
			},
			mockResponse: &http.Response{
				StatusCode: 201,
				Body:       io.NopCloser(strings.NewReader(`{"id":"inc_123"}`)),
			},
			validateAuth: BearerPrefix + "another_key",
			expectError:  false,
		},
		{
			name:  "handles nil request body",
			token: []byte("key_with_nil_body"),
			setupRequest: func() *http.Request {
				req, _ := http.NewRequest("GET", "https://api.example.com/v1/monitors/mon_123", nil)
				return req
			},
			mockResponse: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`{"id":"mon_123"}`)),
			},
			validateAuth: BearerPrefix + "key_with_nil_body",
			expectError:  false,
		},
		{
			name:  "handles non-nil request body",
			token: []byte("key_with_body"),
			setupRequest: func() *http.Request {
				req, _ := http.NewRequest("POST", "https://api.example.com/v1/monitors", strings.NewReader(`{"name":"test"}`))
				return req
			},
			mockResponse: &http.Response{
				StatusCode: 201,
				Body:       io.NopCloser(strings.NewReader(`{"id":"mon_new"}`)),
			},
			validateAuth: BearerPrefix + "key_with_body",
			expectError:  false,
		},
		{
			name:  "propagates underlying transport error",
			token: []byte("error_key"),
			setupRequest: func() *http.Request {
				req, _ := http.NewRequest("GET", "https://api.example.com/v1/monitors", nil)
				return req
			},
			mockError:     errors.New("network error"),
			expectError:   true,
			errorContains: "network error",
		},
		{
			name:  "handles empty token",
			token: []byte(""),
			setupRequest: func() *http.Request {
				req, _ := http.NewRequest("GET", "https://api.example.com/v1/monitors", nil)
				return req
			},
			mockResponse: &http.Response{
				StatusCode: 401,
				Body:       io.NopCloser(strings.NewReader(`{"error":"unauthorized"}`)),
			},
			validateAuth: BearerPrefix,
			expectError:  false,
		},
		{
			name:  "handles special characters in token",
			token: []byte("sk_test-123_ABC.xyz"),
			setupRequest: func() *http.Request {
				req, _ := http.NewRequest("GET", "https://api.example.com/v1/monitors", nil)
				return req
			},
			mockResponse: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
			},
			validateAuth: BearerPrefix + "sk_test-123_ABC.xyz",
			expectError:  false,
		},
		{
			name:  "does not mutate original request",
			token: []byte("mutation_test_key"),
			setupRequest: func() *http.Request {
				req, _ := http.NewRequest("GET", "https://api.example.com/v1/monitors", nil)
				req.Header.Set("X-Original-Header", "original-value")
				return req
			},
			mockResponse: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
			},
			validateAuth: BearerPrefix + "mutation_test_key",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock transport
			mockTransport := &mockRoundTripper{
				response: tt.mockResponse,
				err:      tt.mockError,
			}

			// Create authTransport
			transport := &authTransport{
				token: tt.token,
				next:  mockTransport,
			}

			// Create request
			req := tt.setupRequest()
			originalAuthHeader := req.Header.Get(HeaderAuthorization)

			// Execute RoundTrip
			resp, err := transport.RoundTrip(req)

			// Verify error handling
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify response
			if resp != tt.mockResponse {
				t.Errorf("expected response to be passed through unchanged")
			}

			// Verify Authorization header was added to the cloned request
			if mockTransport.capturedRequest == nil {
				t.Fatal("expected request to be passed to underlying transport")
			}

			capturedAuth := mockTransport.capturedRequest.Header.Get(HeaderAuthorization)
			if capturedAuth != tt.validateAuth {
				t.Errorf("expected Authorization header %q, got %q", tt.validateAuth, capturedAuth)
			}

			// Verify original request was not mutated
			if req.Header.Get(HeaderAuthorization) != originalAuthHeader {
				t.Errorf("original request was mutated: Authorization header changed from %q to %q",
					originalAuthHeader, req.Header.Get(HeaderAuthorization))
			}

			// Verify underlying transport was called exactly once
			if mockTransport.callCount != 1 {
				t.Errorf("expected underlying transport to be called once, got %d calls", mockTransport.callCount)
			}

			// Close response body to prevent resource leak
			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}
		})
	}
}

// TestAuthTransport_RoundTrip_RequestCloning tests that the request is properly cloned.
func TestAuthTransport_RoundTrip_RequestCloning(t *testing.T) {
	mockTransport := &mockRoundTripper{
		response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
		},
	}

	transport := &authTransport{
		token: []byte("clone_test_key"),
		next:  mockTransport,
	}

	// Create request with context
	type contextKey string
	const testKey contextKey = "test_key"
	ctx := context.WithValue(context.Background(), testKey, "test_value")
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.example.com/test", nil)
	req.Header.Set("X-Original", "value1")

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	// Verify the captured request is different from original
	if mockTransport.capturedRequest == req {
		t.Error("authTransport did not clone the request (same pointer)")
	}

	// Verify context is preserved
	if mockTransport.capturedRequest.Context() != ctx {
		t.Error("request context was not preserved in clone")
	}

	// Verify original headers are preserved in clone
	if mockTransport.capturedRequest.Header.Get("X-Original") != "value1" {
		t.Error("original headers were not preserved in clone")
	}

	// Verify Authorization header was added to clone
	if mockTransport.capturedRequest.Header.Get(HeaderAuthorization) != BearerPrefix+"clone_test_key" {
		t.Error("Authorization header was not added to cloned request")
	}

	// Verify original request does not have Authorization header
	if req.Header.Get(HeaderAuthorization) != "" {
		t.Error("Authorization header was added to original request (mutation)")
	}
}

// TestTLSEnforcedTransport_RoundTrip tests the tlsEnforcedTransport.RoundTrip function.
func TestTLSEnforcedTransport_RoundTrip(t *testing.T) {
	tests := []struct {
		name          string
		requestURL    string
		mockResponse  *http.Response
		mockError     error
		expectError   bool
		errorContains string
	}{
		{
			name:       "allows HTTPS requests",
			requestURL: "https://api.example.com/v1/monitors",
			mockResponse: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
			},
			expectError: false,
		},
		{
			name:          "blocks HTTP requests",
			requestURL:    "http://api.example.com/v1/monitors",
			expectError:   true,
			errorContains: "HTTPS required",
		},
		{
			name:          "blocks HTTP with different host",
			requestURL:    "http://malicious.example.com/steal-credentials",
			expectError:   true,
			errorContains: "HTTPS required",
		},
		{
			name:       "allows HTTPS with different paths",
			requestURL: "https://api.example.com/v2/incidents/inc_123",
			mockResponse: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`{"id":"inc_123"}`)),
			},
			expectError: false,
		},
		{
			name:       "allows HTTPS with query parameters",
			requestURL: "https://api.example.com/v1/monitors?status=down",
			mockResponse: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`[]`)),
			},
			expectError: false,
		},
		{
			name:          "error message includes scheme and host",
			requestURL:    "http://example.com/test",
			expectError:   true,
			errorContains: "http://",
		},
		{
			name:       "propagates underlying transport error for HTTPS",
			requestURL: "https://api.example.com/v1/monitors",
			mockError:  errors.New("connection refused"),
			mockResponse: &http.Response{
				StatusCode: 500,
			},
			expectError:   true,
			errorContains: "connection refused",
		},
		{ //nolint:gosec // G101: test URL with placeholder credentials, not real secrets
			name:       "allows HTTPS with authentication",
			requestURL: "https://user:pass@api.example.com/v1/monitors",
			mockResponse: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
			},
			expectError: false,
		},
		{
			name:          "blocks FTP scheme",
			requestURL:    "ftp://api.example.com/file.txt",
			expectError:   true,
			errorContains: "HTTPS required",
		},
		{
			name:       "allows HTTPS with non-standard port",
			requestURL: "https://api.example.com:8443/v1/monitors",
			mockResponse: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock transport
			mockTransport := &mockRoundTripper{
				response: tt.mockResponse,
				err:      tt.mockError,
			}

			// Create tlsEnforcedTransport
			transport := &tlsEnforcedTransport{
				next: mockTransport,
			}

			// Create request
			req, err := http.NewRequest("GET", tt.requestURL, nil)
			if err != nil {
				t.Fatalf("failed to create test request: %v", err)
			}

			// Execute RoundTrip
			resp, err := transport.RoundTrip(req)

			// Verify error handling
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
				// Should not call underlying transport for blocked requests
				if req.URL.Scheme != "https" && mockTransport.callCount > 0 {
					t.Error("underlying transport was called for blocked HTTP request")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify response
			if resp != tt.mockResponse {
				t.Errorf("expected response to be passed through unchanged")
			}

			// Verify underlying transport was called for HTTPS
			if mockTransport.callCount != 1 {
				t.Errorf("expected underlying transport to be called once for HTTPS, got %d calls", mockTransport.callCount)
			}

			// Verify request was passed through unchanged
			if mockTransport.capturedRequest != req {
				t.Error("request was not passed through unchanged")
			}

			// Close response body to prevent resource leak
			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}
		})
	}
}

// TestTLSEnforcedTransport_RoundTrip_ErrorMessage tests error message format.
func TestTLSEnforcedTransport_RoundTrip_ErrorMessage(t *testing.T) {
	mockTransport := &mockRoundTripper{}
	transport := &tlsEnforcedTransport{next: mockTransport}

	req, _ := http.NewRequest("GET", "http://api.hyperping.io/v1/monitors", nil)
	resp, err := transport.RoundTrip(req)

	if err == nil {
		t.Fatal("expected error for HTTP request")
	}
	if resp != nil {
		resp.Body.Close()
	}

	// Error message should contain:
	// 1. Transport name
	// 2. Host
	// 3. Actual scheme
	errorMsg := err.Error()
	if !strings.Contains(errorMsg, "tlsEnforcedTransport") {
		t.Errorf("error message missing transport name: %s", errorMsg)
	}
	if !strings.Contains(errorMsg, "api.hyperping.io") {
		t.Errorf("error message missing host: %s", errorMsg)
	}
	if !strings.Contains(errorMsg, "http://") {
		t.Errorf("error message missing actual scheme: %s", errorMsg)
	}
	if !strings.Contains(errorMsg, "HTTPS required") {
		t.Errorf("error message missing 'HTTPS required': %s", errorMsg)
	}
}

// TestDefaultTLSConfig tests the defaultTLSConfig function.
func TestDefaultTLSConfig(t *testing.T) {
	config := defaultTLSConfig()

	if config == nil {
		t.Fatal("expected non-nil TLS config")
	}

	// Verify minimum TLS version is TLS 1.2
	if config.MinVersion != 0x0303 { // TLS 1.2
		t.Errorf("expected MinVersion to be TLS 1.2 (0x0303), got 0x%04x", config.MinVersion)
	}

	// Verify cipher suites are configured
	if len(config.CipherSuites) == 0 {
		t.Fatal("expected cipher suites to be configured")
	}

	// Verify we have exactly 6 cipher suites (all AEAD)
	expectedCount := 6
	if len(config.CipherSuites) != expectedCount {
		t.Errorf("expected %d cipher suites, got %d", expectedCount, len(config.CipherSuites))
	}

	// Verify the cipher suites use the crypto/tls constants
	// We don't hardcode the hex values, just verify they're non-zero and distinct
	seen := make(map[uint16]bool)
	for i, cipher := range config.CipherSuites {
		if cipher == 0 {
			t.Errorf("cipher suite %d: expected non-zero value", i)
		}
		if seen[cipher] {
			t.Errorf("cipher suite %d: duplicate value 0x%04x", i, cipher)
		}
		seen[cipher] = true
	}

	// Verify that the suites include the expected AEAD suites from tls package
	// by checking they match the constants (order doesn't matter for this check)
	expectedSuites := map[uint16]bool{
		0xc030: true, // TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
		0xc02c: true, // TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
		0xc02f: true, // TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
		0xc02b: true, // TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
		0xcca9: true, // TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256
		0xcca8: true, // TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256
	}

	for _, cipher := range config.CipherSuites {
		if !expectedSuites[cipher] {
			t.Errorf("unexpected cipher suite: 0x%04x", cipher)
		}
	}
}

// TestBuildTransportChain tests the buildTransportChain function.
func TestBuildTransportChain(t *testing.T) {
	tests := []struct {
		name       string
		apiKey     []byte
		baseURL    string
		expectAuth bool
		expectTLS  bool
	}{
		{
			name:       "HTTPS non-localhost",
			apiKey:     []byte("test_key"),
			baseURL:    "https://api.hyperping.io",
			expectAuth: true,
			expectTLS:  false, // TLS enforcement applied to standard transport
		},
		{
			name:       "HTTP localhost",
			apiKey:     []byte("test_key"),
			baseURL:    "http://localhost:8080",
			expectAuth: true,
			expectTLS:  false, // Localhost exempt from TLS enforcement
		},
		{
			name:       "HTTP non-localhost",
			apiKey:     []byte("test_key"),
			baseURL:    "http://api.example.com",
			expectAuth: true,
			expectTLS:  true, // Non-HTTPS non-localhost wrapped in TLS enforcement
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseTransport := &mockRoundTripper{
				response: &http.Response{StatusCode: 200},
			}

			chain := buildTransportChain(tt.apiKey, baseTransport, tt.baseURL)

			// Top layer should always be authTransport
			authT, ok := chain.(*authTransport)
			if !ok {
				t.Fatal("expected top layer to be authTransport")
			}

			// Verify API key
			if string(authT.token) != string(tt.apiKey) {
				t.Errorf("expected token %q, got %q", string(tt.apiKey), string(authT.token))
			}

			// Check if TLS enforcement is applied
			if tt.expectTLS {
				if _, ok := authT.next.(*tlsEnforcedTransport); !ok {
					t.Error("expected tlsEnforcedTransport in chain")
				}
			}
		})
	}
}

// TestEnforceTLS tests the enforceTLS function.
func TestEnforceTLS(t *testing.T) {
	tests := []struct {
		name            string
		baseURL         string
		expectWrapped   bool
		expectTLSConfig bool
	}{
		{
			name:            "localhost HTTP exempted",
			baseURL:         "http://localhost:8080",
			expectWrapped:   false,
			expectTLSConfig: false,
		},
		{
			name:            "127.0.0.1 HTTP exempted",
			baseURL:         "http://127.0.0.1:3000",
			expectWrapped:   false,
			expectTLSConfig: false,
		},
		{
			name:            "IPv6 localhost exempted",
			baseURL:         "http://[::1]:8080",
			expectWrapped:   false,
			expectTLSConfig: false,
		},
		{
			name:            "HTTPS non-localhost applies TLS config",
			baseURL:         "https://api.hyperping.io",
			expectWrapped:   false,
			expectTLSConfig: true,
		},
		{
			name:            "HTTP non-localhost wrapped",
			baseURL:         "http://api.example.com",
			expectWrapped:   true,
			expectTLSConfig: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseTransport := &http.Transport{}
			result := enforceTLS(baseTransport, tt.baseURL)

			if tt.expectWrapped {
				if _, ok := result.(*tlsEnforcedTransport); !ok {
					t.Error("expected result to be wrapped in tlsEnforcedTransport")
				}
			}

			if tt.expectTLSConfig {
				if httpT, ok := result.(*http.Transport); ok {
					if httpT.TLSClientConfig == nil {
						t.Error("expected TLSClientConfig to be set")
					} else {
						if httpT.TLSClientConfig.MinVersion != 0x0303 {
							t.Error("expected TLS 1.2 minimum version")
						}
					}
				}
			}
		})
	}
}

// TestEnforceTLS_CustomTransport tests enforceTLS with non-standard transport.
func TestEnforceTLS_CustomTransport(t *testing.T) {
	customTransport := &mockRoundTripper{}

	// Non-localhost HTTPS with custom transport should wrap
	result := enforceTLS(customTransport, "https://api.hyperping.io")

	if _, ok := result.(*tlsEnforcedTransport); !ok {
		t.Error("expected custom transport to be wrapped in tlsEnforcedTransport")
	}
}
