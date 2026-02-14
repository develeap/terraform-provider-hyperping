// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package terraflyerrors

import (
	"errors"
	"strings"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestEnhanceClientError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		operation    string
		resource     string
		field        string
		wantContains []string
		wantCategory ErrorCategory
	}{
		{
			name:      "nil error returns nil",
			err:       nil,
			operation: "create",
			resource:  "hyperping_monitor.test",
		},
		{
			name:      "unauthorized error",
			err:       client.ErrUnauthorized,
			operation: "create",
			resource:  "hyperping_monitor.test",
			wantContains: []string{
				"Authentication Failed",
				"create",
				"hyperping_monitor.test",
				"Suggestions:",
				"Documentation:",
			},
			wantCategory: CategoryAuth,
		},
		{
			name:      "rate limit error",
			err:       client.ErrRateLimited,
			operation: "update",
			resource:  "hyperping_monitor.prod",
			wantContains: []string{
				"Rate Limit Exceeded",
				"update",
				"hyperping_monitor.prod",
				"parallelism",
			},
			wantCategory: CategoryRateLimit,
		},
		{
			name:      "validation error",
			err:       client.ErrValidation,
			operation: "create",
			resource:  "hyperping_monitor.test",
			field:     "frequency",
			wantContains: []string{
				"Validation Error",
				"create",
				"frequency",
			},
			wantCategory: CategoryValidation,
		},
		{
			name:      "not found error",
			err:       client.ErrNotFound,
			operation: "read",
			resource:  "hyperping_monitor.test",
			wantContains: []string{
				"Resource Not Found",
				"read",
				"terraform import",
			},
			wantCategory: CategoryNotFound,
		},
		{
			name:      "server error",
			err:       client.ErrServerError,
			operation: "delete",
			resource:  "hyperping_incident.test",
			wantContains: []string{
				"Server Error",
				"delete",
				"status.hyperping.io",
			},
			wantCategory: CategoryServer,
		},
		{
			name:      "network error - connection refused",
			err:       errors.New("dial tcp: connection refused"),
			operation: "create",
			resource:  "hyperping_monitor.test",
			wantContains: []string{
				"Network Error",
				"internet connection",
				"firewall",
			},
			wantCategory: CategoryNetwork,
		},
		{
			name:      "network error - timeout",
			err:       errors.New("request failed: i/o timeout"),
			operation: "read",
			resource:  "hyperping_monitor.test",
			wantContains: []string{
				"Network Error",
			},
			wantCategory: CategoryNetwork,
		},
		{
			name:      "circuit breaker error",
			err:       errors.New("circuit breaker is open"),
			operation: "create",
			resource:  "hyperping_monitor.test",
			wantContains: []string{
				"Circuit Breaker Open",
				"30 seconds",
			},
			wantCategory: CategoryCircuit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enhanced := EnhanceClientError(tt.err, tt.operation, tt.resource, tt.field)

			if tt.err == nil {
				if enhanced != nil {
					t.Errorf("Expected nil error for nil input")
				}
				return
			}

			if enhanced == nil {
				t.Fatal("Expected enhanced error, got nil")
			}

			output := enhanced.Error()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Expected output to contain %q\nGot: %s", want, output)
				}
			}
		})
	}
}

func TestEnhanceAPIError(t *testing.T) {
	tests := []struct {
		name         string
		apiErr       *client.APIError
		operation    string
		resource     string
		field        string
		wantContains []string
	}{
		{
			name: "auth error",
			apiErr: &client.APIError{
				StatusCode: 401,
				Message:    "Invalid API key",
			},
			operation: "create",
			resource:  "hyperping_monitor.test",
			wantContains: []string{
				"Authentication Failed",
				"HYPERPING_API_KEY",
			},
		},
		{
			name: "rate limit with retry-after",
			apiErr: &client.APIError{
				StatusCode: 429,
				Message:    "Rate limit exceeded",
				RetryAfter: 60,
			},
			operation: "update",
			resource:  "hyperping_monitor.test",
			wantContains: []string{
				"Rate Limit Exceeded",
				"1m0s",
			},
		},
		{
			name: "validation with details",
			apiErr: &client.APIError{
				StatusCode: 400,
				Message:    "Validation failed",
				Details: []client.ValidationDetail{
					{Field: "frequency", Message: "Invalid value"},
					{Field: "url", Message: "Required"},
				},
			},
			operation: "create",
			resource:  "hyperping_monitor.test",
			wantContains: []string{
				"Validation Error",
				"frequency: Invalid value",
				"url: Required",
			},
		},
		{
			name: "not found",
			apiErr: &client.APIError{
				StatusCode: 404,
				Message:    "Monitor not found",
			},
			operation: "read",
			resource:  "hyperping_monitor.test",
			wantContains: []string{
				"Resource Not Found",
			},
		},
		{
			name: "server error",
			apiErr: &client.APIError{
				StatusCode: 500,
				Message:    "Internal server error",
			},
			operation: "create",
			resource:  "hyperping_monitor.test",
			wantContains: []string{
				"Server Error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enhanced := enhanceAPIError(tt.apiErr, tt.operation, tt.resource, tt.field)
			output := enhanced.Error()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Expected output to contain %q\nGot: %s", want, output)
				}
			}
		})
	}
}

func TestIsNetworkError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "connection refused",
			err:  errors.New("dial tcp 1.2.3.4:443: connection refused"),
			want: true,
		},
		{
			name: "no such host",
			err:  errors.New("lookup api.example.com: no such host"),
			want: true,
		},
		{
			name: "timeout",
			err:  errors.New("request timeout"),
			want: true,
		},
		{
			name: "i/o timeout",
			err:  errors.New("i/o timeout"),
			want: true,
		},
		{
			name: "network unreachable",
			err:  errors.New("network is unreachable"),
			want: true,
		},
		{
			name: "connection reset",
			err:  errors.New("connection reset by peer"),
			want: true,
		},
		{
			name: "broken pipe",
			err:  errors.New("write: broken pipe"),
			want: true,
		},
		{
			name: "not a network error",
			err:  errors.New("validation failed"),
			want: false,
		},
		{
			name: "API error",
			err:  errors.New("API error (status 400): bad request"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isNetworkError(tt.err)
			if got != tt.want {
				t.Errorf("isNetworkError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsCircuitBreakerError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "circuit breaker open",
			err:  errors.New("circuit breaker is open"),
			want: true,
		},
		{
			name: "too many failures",
			err:  errors.New("too many failures"),
			want: true,
		},
		{
			name: "not a circuit breaker error",
			err:  errors.New("validation failed"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isCircuitBreakerError(tt.err)
			if got != tt.want {
				t.Errorf("isCircuitBreakerError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		details []client.ValidationDetail
		want    string
	}{
		{
			name:    "empty details",
			details: []client.ValidationDetail{},
			want:    "validation failed",
		},
		{
			name: "single detail",
			details: []client.ValidationDetail{
				{Field: "url", Message: "Invalid URL format"},
			},
			want: "url: Invalid URL format",
		},
		{
			name: "multiple details",
			details: []client.ValidationDetail{
				{Field: "url", Message: "Invalid URL format"},
				{Field: "frequency", Message: "Must be positive"},
			},
			want: "url: Invalid URL format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatValidationErrors(tt.details)
			if !strings.Contains(got, tt.want) {
				t.Errorf("FormatValidationErrors() = %q, want to contain %q", got, tt.want)
			}
		})
	}
}
