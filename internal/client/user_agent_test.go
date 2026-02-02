// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestBuildUserAgent(t *testing.T) {
	tests := []struct {
		name            string
		version         string
		appendUserAgent string
		expectedPattern string
		shouldContain   []string
	}{
		{
			name:            "default version",
			version:         "dev",
			appendUserAgent: "",
			expectedPattern: "terraform-provider-hyperping/dev",
			shouldContain: []string{
				"terraform-provider-hyperping/dev",
				runtime.Version(),
				runtime.GOOS,
				runtime.GOARCH,
			},
		},
		{
			name:            "release version",
			version:         "1.0.0",
			appendUserAgent: "",
			expectedPattern: "terraform-provider-hyperping/1.0.0",
			shouldContain: []string{
				"terraform-provider-hyperping/1.0.0",
				runtime.Version(),
				runtime.GOOS,
				runtime.GOARCH,
			},
		},
		{
			name:            "with TF_APPEND_USER_AGENT",
			version:         "1.0.0",
			appendUserAgent: "custom-module/2.0",
			expectedPattern: "terraform-provider-hyperping/1.0.0",
			shouldContain: []string{
				"terraform-provider-hyperping/1.0.0",
				runtime.Version(),
				"custom-module/2.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set TF_APPEND_USER_AGENT if provided
			if tt.appendUserAgent != "" {
				oldEnv := os.Getenv("TF_APPEND_USER_AGENT")
				os.Setenv("TF_APPEND_USER_AGENT", tt.appendUserAgent)
				defer os.Setenv("TF_APPEND_USER_AGENT", oldEnv)
			}

			result := buildUserAgent(tt.version)

			// Check that all expected strings are present
			for _, expected := range tt.shouldContain {
				if !strings.Contains(result, expected) {
					t.Errorf("buildUserAgent() = %q, should contain %q", result, expected)
				}
			}
		})
	}
}

func TestBuildUserAgent_Format(t *testing.T) {
	result := buildUserAgent("1.2.3")

	// Should match pattern: terraform-provider-hyperping/VERSION (goVERSION; OS/ARCH)
	expectedPrefix := "terraform-provider-hyperping/1.2.3 ("
	if !strings.HasPrefix(result, expectedPrefix) {
		t.Errorf("User-Agent should start with %q, got %q", expectedPrefix, result)
	}

	// Should contain parentheses with system info
	if !strings.Contains(result, "(") || !strings.Contains(result, ")") {
		t.Errorf("User-Agent should contain parentheses with system info, got %q", result)
	}

	// Should contain semicolon separating Go version and OS/arch
	if !strings.Contains(result, ";") {
		t.Errorf("User-Agent should contain semicolon, got %q", result)
	}

	// Should contain forward slash for OS/arch
	if !strings.Contains(result, "/") {
		t.Errorf("User-Agent should contain forward slash for OS/arch, got %q", result)
	}
}

func TestClient_UserAgentHeader(t *testing.T) {
	var receivedUserAgent string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUserAgent = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	c := NewClient("test_key",
		WithBaseURL(server.URL),
		WithVersion("1.0.0"),
	)

	var result map[string]string
	err := c.doRequest(context.Background(), "GET", "/test", nil, &result)
	if err != nil {
		t.Fatalf("doRequest() failed: %v", err)
	}

	if receivedUserAgent == "" {
		t.Error("User-Agent header was not sent")
	}

	expectedSubstrings := []string{
		"terraform-provider-hyperping/1.0.0",
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	}

	for _, expected := range expectedSubstrings {
		if !strings.Contains(receivedUserAgent, expected) {
			t.Errorf("User-Agent %q should contain %q", receivedUserAgent, expected)
		}
	}
}

func TestClient_UserAgentWithAppend(t *testing.T) {
	var receivedUserAgent string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUserAgent = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	// Set TF_APPEND_USER_AGENT
	oldEnv := os.Getenv("TF_APPEND_USER_AGENT")
	os.Setenv("TF_APPEND_USER_AGENT", "my-terraform-module/1.5.0")
	defer os.Setenv("TF_APPEND_USER_AGENT", oldEnv)

	c := NewClient("test_key",
		WithBaseURL(server.URL),
		WithVersion("2.0.0"),
	)

	var result map[string]string
	err := c.doRequest(context.Background(), "GET", "/test", nil, &result)
	if err != nil {
		t.Fatalf("doRequest() failed: %v", err)
	}

	expectedSubstrings := []string{
		"terraform-provider-hyperping/2.0.0",
		"my-terraform-module/1.5.0",
	}

	for _, expected := range expectedSubstrings {
		if !strings.Contains(receivedUserAgent, expected) {
			t.Errorf("User-Agent %q should contain %q", receivedUserAgent, expected)
		}
	}
}

func TestClient_DefaultVersion(t *testing.T) {
	var receivedUserAgent string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUserAgent = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	// Create client without WithVersion (should use "dev")
	c := NewClient("test_key", WithBaseURL(server.URL))

	var result map[string]string
	err := c.doRequest(context.Background(), "GET", "/test", nil, &result)
	if err != nil {
		t.Fatalf("doRequest() failed: %v", err)
	}

	if !strings.Contains(receivedUserAgent, "terraform-provider-hyperping/dev") {
		t.Errorf("Default User-Agent should contain 'terraform-provider-hyperping/dev', got %q", receivedUserAgent)
	}
}

func TestClient_UserAgentOnRetry(t *testing.T) {
	attempts := 0
	var receivedUserAgents []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUserAgents = append(receivedUserAgents, r.Header.Get("User-Agent"))
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": "Rate limited"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	c := NewClient("test_key",
		WithBaseURL(server.URL),
		WithVersion("1.0.0"),
		WithMaxRetries(2),
	)

	var result map[string]string
	err := c.doRequest(context.Background(), "GET", "/test", nil, &result)
	if err != nil {
		t.Fatalf("doRequest() failed: %v", err)
	}

	if len(receivedUserAgents) != 2 {
		t.Errorf("Expected 2 User-Agent headers (original + retry), got %d", len(receivedUserAgents))
	}

	// All attempts should have the same User-Agent
	for i, ua := range receivedUserAgents {
		if ua == "" {
			t.Errorf("Attempt %d: User-Agent was empty", i+1)
		}
		if !strings.Contains(ua, "terraform-provider-hyperping/1.0.0") {
			t.Errorf("Attempt %d: User-Agent %q should contain version", i+1, ua)
		}
	}
}

func TestBuildUserAgent_EmptyAppendUserAgent(t *testing.T) {
	// Ensure empty TF_APPEND_USER_AGENT doesn't add extra spaces
	oldEnv := os.Getenv("TF_APPEND_USER_AGENT")
	os.Setenv("TF_APPEND_USER_AGENT", "")
	defer os.Setenv("TF_APPEND_USER_AGENT", oldEnv)

	result := buildUserAgent("1.0.0")

	// Should not end with a space
	if strings.HasSuffix(result, " ") {
		t.Errorf("User-Agent should not end with space when TF_APPEND_USER_AGENT is empty, got %q", result)
	}
}

func TestWithVersion_Option(t *testing.T) {
	c := NewClient("test_key", WithVersion("5.4.3"))

	if c.version != "5.4.3" {
		t.Errorf("WithVersion() should set version to '5.4.3', got %q", c.version)
	}

	expectedUA := fmt.Sprintf("terraform-provider-hyperping/5.4.3 (%s; %s/%s)",
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	)

	if c.userAgent != expectedUA {
		t.Errorf("userAgent = %q, want %q", c.userAgent, expectedUA)
	}
}

// Benchmark User-Agent construction
func BenchmarkBuildUserAgent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = buildUserAgent("1.0.0")
	}
}

func BenchmarkBuildUserAgent_WithAppend(b *testing.B) {
	os.Setenv("TF_APPEND_USER_AGENT", "custom-module/2.0")
	defer os.Unsetenv("TF_APPEND_USER_AGENT")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildUserAgent("1.0.0")
	}
}
