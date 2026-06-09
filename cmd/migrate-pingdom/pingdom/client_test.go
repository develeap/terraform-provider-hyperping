// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package pingdom

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient_Defaults(t *testing.T) {
	c := NewClient("token")
	if c.apiToken != "token" {
		t.Errorf("apiToken = %q, want token", c.apiToken)
	}
	if c.baseURL != defaultBaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, defaultBaseURL)
	}
	if c.httpClient == nil {
		t.Fatal("httpClient is nil")
	}
	if c.httpClient.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want 30s", c.httpClient.Timeout)
	}
}

func TestNewClient_WithOptions(t *testing.T) {
	custom := &http.Client{Timeout: 5 * time.Second}
	c := NewClient("token",
		WithBaseURL("https://example.test/api"),
		WithHTTPClient(custom),
	)
	if c.baseURL != "https://example.test/api" {
		t.Errorf("baseURL = %q", c.baseURL)
	}
	if c.httpClient != custom {
		t.Error("WithHTTPClient did not set custom client")
	}
}

func TestListChecks_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/checks" {
			t.Errorf("path = %s, want /checks", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer my-token" {
			t.Errorf("Authorization = %q", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q", got)
		}
		// ListChecks is a GET; body must be empty (we pass http.NoBody).
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		if len(body) != 0 {
			t.Errorf("request body = %q, want empty", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"checks":[{"id":1,"name":"a","type":"http","hostname":"a.example.com"},{"id":2,"name":"b","type":"tcp","hostname":"b.example.com","port":5432}]}`))
	}))
	defer srv.Close()

	c := NewClient("my-token", WithBaseURL(srv.URL))
	checks, err := c.ListChecks(context.Background())
	if err != nil {
		t.Fatalf("ListChecks error = %v", err)
	}
	if len(checks) != 2 {
		t.Fatalf("got %d checks, want 2", len(checks))
	}
	if checks[0].Name != "a" || checks[1].Port != 5432 {
		t.Errorf("unexpected checks: %#v", checks)
	}
}

func TestListChecks_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer srv.Close()

	_, err := NewClient("bad", WithBaseURL(srv.URL)).ListChecks(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "API error (status 401)") {
		t.Errorf("error = %v, want status 401", err)
	}
}

func TestListChecks_BadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	}))
	defer srv.Close()

	_, err := NewClient("t", WithBaseURL(srv.URL)).ListChecks(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "parsing response") {
		t.Errorf("error = %v, want parse error", err)
	}
}

func TestListChecks_NetworkError(t *testing.T) {
	// Stand up a server, capture its URL, then close it. The URL is now a
	// well-formed but unreachable address, so http.Do fails at dial time.
	// This is more portable than relying on port 0 semantics.
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	deadURL := srv.URL
	srv.Close()

	c := NewClient("t", WithBaseURL(deadURL))
	_, err := c.ListChecks(context.Background())
	if err == nil {
		t.Fatal("expected network error")
	}
	if !strings.Contains(err.Error(), "executing request") {
		t.Errorf("error = %v, want executing request", err)
	}
}

func TestListChecks_BadBaseURL(t *testing.T) {
	c := NewClient("t", WithBaseURL("://invalid"))
	_, err := c.ListChecks(context.Background())
	if err == nil {
		t.Fatal("expected error from invalid URL")
	}
	if !strings.Contains(err.Error(), "creating request") {
		t.Errorf("error = %v, want creating request", err)
	}
}

func TestGetCheck_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/checks/42" {
			t.Errorf("path = %s, want /checks/42", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"check":{"id":42,"name":"detail","type":"http","hostname":"x.example.com"}}`))
	}))
	defer srv.Close()

	c := NewClient("t", WithBaseURL(srv.URL))
	check, err := c.GetCheck(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetCheck error = %v", err)
	}
	if check.ID != 42 || check.Name != "detail" {
		t.Errorf("unexpected check: %#v", check)
	}
}

func TestGetCheck_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`not found`))
	}))
	defer srv.Close()

	_, err := NewClient("t", WithBaseURL(srv.URL)).GetCheck(context.Background(), 1)
	if err == nil || !strings.Contains(err.Error(), "API error (status 404)") {
		t.Errorf("error = %v, want 404 error", err)
	}
}

func TestGetCheck_BadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{`))
	}))
	defer srv.Close()
	_, err := NewClient("t", WithBaseURL(srv.URL)).GetCheck(context.Background(), 1)
	if err == nil || !strings.Contains(err.Error(), "parsing response") {
		t.Errorf("error = %v, want parsing response error", err)
	}
}

func TestGetCheck_BadBaseURL(t *testing.T) {
	c := NewClient("t", WithBaseURL("://invalid"))
	_, err := c.GetCheck(context.Background(), 1)
	if err == nil || !strings.Contains(err.Error(), "creating request") {
		t.Errorf("error = %v, want creating request", err)
	}
}

func TestGetCheck_NetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	deadURL := srv.URL
	srv.Close()

	c := NewClient("t", WithBaseURL(deadURL))
	_, err := c.GetCheck(context.Background(), 1)
	if err == nil || !strings.Contains(err.Error(), "executing request") {
		t.Errorf("error = %v, want executing request error", err)
	}
}
