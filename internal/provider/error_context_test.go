// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestDetectErrorContext_NotFound(t *testing.T) {
	t.Parallel()

	err := client.NewAPIError(404, "resource not found")
	ctx := DetectErrorContext("Monitor", "mon_123", "read", err)

	assert.Equal(t, "not_found", ctx.Type)
	assert.Equal(t, 404, ctx.HTTPStatus)
	assert.Equal(t, "Monitor", ctx.ResourceType)
	assert.Equal(t, "mon_123", ctx.ResourceID)
	assert.Equal(t, "read", ctx.Operation)
	assert.Contains(t, ctx.Message, "404")
}

func TestDetectErrorContext_Unauthorized(t *testing.T) {
	t.Parallel()

	err := client.NewAPIError(401, "invalid API key")
	ctx := DetectErrorContext("Monitor", "mon_123", "read", err)

	assert.Equal(t, "auth_error", ctx.Type)
	assert.Equal(t, 401, ctx.HTTPStatus)
	assert.Equal(t, "Monitor", ctx.ResourceType)
	assert.Equal(t, "mon_123", ctx.ResourceID)
	assert.Equal(t, "read", ctx.Operation)
}

func TestDetectErrorContext_Forbidden(t *testing.T) {
	t.Parallel()

	err := client.NewAPIError(403, "access forbidden")
	ctx := DetectErrorContext("Monitor", "mon_123", "update", err)

	assert.Equal(t, "auth_error", ctx.Type)
	assert.Equal(t, 403, ctx.HTTPStatus)
	assert.Equal(t, "Monitor", ctx.ResourceType)
	assert.Equal(t, "mon_123", ctx.ResourceID)
	assert.Equal(t, "update", ctx.Operation)
}

func TestDetectErrorContext_RateLimit(t *testing.T) {
	t.Parallel()

	err := client.NewRateLimitError(120)
	ctx := DetectErrorContext("Monitor", "", "create", err)

	assert.Equal(t, "rate_limit", ctx.Type)
	assert.Equal(t, 429, ctx.HTTPStatus)
	assert.Equal(t, 120, ctx.RetryAfter)
	assert.Equal(t, "Monitor", ctx.ResourceType)
	assert.Equal(t, "create", ctx.Operation)
}

func TestDetectErrorContext_ServerError(t *testing.T) {
	t.Parallel()

	err := client.NewAPIError(500, "internal server error")
	ctx := DetectErrorContext("Incident", "inc_123", "delete", err)

	assert.Equal(t, "server_error", ctx.Type)
	assert.Equal(t, 500, ctx.HTTPStatus)
	assert.Equal(t, "Incident", ctx.ResourceType)
	assert.Equal(t, "inc_123", ctx.ResourceID)
	assert.Equal(t, "delete", ctx.Operation)
}

func TestDetectErrorContext_ValidationError(t *testing.T) {
	t.Parallel()

	details := []client.ValidationDetail{
		{Field: "url", Message: "Invalid URL format"},
	}
	err := client.NewValidationError(400, "validation failed", details)
	ctx := DetectErrorContext("Monitor", "", "create", err)

	assert.Equal(t, "validation", ctx.Type)
	assert.Equal(t, 400, ctx.HTTPStatus)
	assert.Equal(t, "Monitor", ctx.ResourceType)
	assert.Equal(t, "create", ctx.Operation)
}

func TestDetectErrorContext_UnknownError(t *testing.T) {
	t.Parallel()

	err := errors.New("generic error")
	ctx := DetectErrorContext("Monitor", "mon_123", "read", err)

	assert.Equal(t, "unknown", ctx.Type)
	assert.Equal(t, 0, ctx.HTTPStatus)
	assert.Equal(t, "Monitor", ctx.ResourceType)
	assert.Equal(t, "mon_123", ctx.ResourceID)
	assert.Equal(t, "read", ctx.Operation)
	assert.Equal(t, "generic error", ctx.Message)
}

func TestDetectErrorContext_NilError(t *testing.T) {
	t.Parallel()

	ctx := DetectErrorContext("Monitor", "mon_123", "read", nil)

	assert.Equal(t, "unknown", ctx.Type)
	assert.Equal(t, 0, ctx.HTTPStatus)
	assert.Equal(t, "Monitor", ctx.ResourceType)
	assert.Equal(t, "mon_123", ctx.ResourceID)
	assert.Equal(t, "read", ctx.Operation)
	assert.Empty(t, ctx.Message)
}

func TestDetectErrorContext_AllFields(t *testing.T) {
	t.Parallel()

	err := client.NewRateLimitError(90)
	ctx := DetectErrorContext("Maintenance", "maint_456", "update", err)

	require.NotEmpty(t, ctx.Type)
	require.NotZero(t, ctx.HTTPStatus)
	require.NotZero(t, ctx.RetryAfter)
	require.NotEmpty(t, ctx.ResourceType)
	require.NotEmpty(t, ctx.ResourceID)
	require.NotEmpty(t, ctx.Operation)
	require.NotEmpty(t, ctx.Message)

	assert.Equal(t, "rate_limit", ctx.Type)
	assert.Equal(t, 429, ctx.HTTPStatus)
	assert.Equal(t, 90, ctx.RetryAfter)
	assert.Equal(t, "Maintenance", ctx.ResourceType)
	assert.Equal(t, "maint_456", ctx.ResourceID)
	assert.Equal(t, "update", ctx.Operation)
}

func TestDetectErrorContext_DifferentOperations(t *testing.T) {
	t.Parallel()

	operations := []string{"create", "read", "update", "delete"}
	err := client.NewAPIError(404, "not found")

	for _, op := range operations {
		t.Run(op, func(t *testing.T) {
			t.Parallel()
			ctx := DetectErrorContext("Monitor", "mon_123", op, err)
			assert.Equal(t, op, ctx.Operation)
			assert.Equal(t, "not_found", ctx.Type)
		})
	}
}

func TestDetectErrorContext_DifferentResources(t *testing.T) {
	t.Parallel()

	resources := []string{"Monitor", "Incident", "Maintenance", "Healthcheck", "Status Page"}
	err := client.NewAPIError(401, "unauthorized")

	for _, resource := range resources {
		t.Run(resource, func(t *testing.T) {
			t.Parallel()
			ctx := DetectErrorContext(resource, "test_123", "read", err)
			assert.Equal(t, resource, ctx.ResourceType)
			assert.Equal(t, "auth_error", ctx.Type)
		})
	}
}

func TestDetectErrorContext_EmptyResourceID(t *testing.T) {
	t.Parallel()

	err := client.NewAPIError(400, "validation error")
	ctx := DetectErrorContext("Monitor", "", "create", err)

	assert.Equal(t, "validation", ctx.Type)
	assert.Empty(t, ctx.ResourceID)
	assert.Equal(t, "create", ctx.Operation)
}

func TestExtractRetryAfter_WithSeconds(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		message  string
		expected int
	}{
		{
			name:     "60 seconds",
			message:  "API error (status 429): rate limit exceeded - retry after 60 seconds",
			expected: 60,
		},
		{
			name:     "120 seconds",
			message:  "rate limit exceeded - retry after 120 seconds",
			expected: 120,
		},
		{
			name:     "single second",
			message:  "retry after 1 second",
			expected: 1,
		},
		{
			name:     "300 seconds",
			message:  "Please retry after 300 seconds",
			expected: 300,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := errors.New(tc.message)
			result := extractRetryAfter(err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestExtractRetryAfter_NoMatch(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		err     error
		message string
	}{
		{
			name:    "no retry pattern",
			err:     errors.New("rate limit exceeded"),
			message: "should default to 60",
		},
		{
			name:    "different format",
			err:     errors.New("too many requests"),
			message: "should default to 60",
		},
		{
			name:    "nil error",
			err:     nil,
			message: "should default to 60",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := extractRetryAfter(tc.err)
			assert.Equal(t, 60, result, tc.message)
		})
	}
}

func TestExtractRetryAfter_MultipleFormats(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		message  string
		expected int
	}{
		{
			name:     "uppercase RETRY AFTER",
			message:  "RETRY AFTER 90 SECONDS",
			expected: 90,
		},
		{
			name:     "mixed case Retry After",
			message:  "Retry After 45 Seconds",
			expected: 45,
		},
		{
			name:     "lowercase retry after",
			message:  "retry after 30 seconds",
			expected: 30,
		},
		{
			name:     "singular second",
			message:  "retry after 1 second",
			expected: 1,
		},
		{
			name:     "large number",
			message:  "retry after 3600 seconds",
			expected: 3600,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := errors.New(tc.message)
			result := extractRetryAfter(err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestExtractStatusCode_Success(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		message  string
		expected int
	}{
		{
			name:     "500 status",
			message:  "API error (status 500): internal server error",
			expected: 500,
		},
		{
			name:     "502 status",
			message:  "status 502 bad gateway",
			expected: 502,
		},
		{
			name:     "503 status",
			message:  "service unavailable status 503",
			expected: 503,
		},
		{
			name:     "400 status",
			message:  "API error (status 400): bad request",
			expected: 400,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := errors.New(tc.message)
			result := extractStatusCode(err, 999)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestExtractStatusCode_DefaultValue(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		err          error
		defaultCode  int
		expectedCode int
	}{
		{
			name:         "no status code",
			err:          errors.New("generic error"),
			defaultCode:  500,
			expectedCode: 500,
		},
		{
			name:         "nil error",
			err:          nil,
			defaultCode:  400,
			expectedCode: 400,
		},
		{
			name:         "invalid status pattern",
			err:          errors.New("status code not found"),
			defaultCode:  503,
			expectedCode: 503,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := extractStatusCode(tc.err, tc.defaultCode)
			assert.Equal(t, tc.expectedCode, result)
		})
	}
}

func TestErrorContext_String(t *testing.T) {
	t.Parallel()

	ctx := ErrorContext{
		Type:         "not_found",
		HTTPStatus:   404,
		ResourceType: "Monitor",
		ResourceID:   "mon_123",
		Operation:    "read",
	}

	str := ctx.String()
	assert.Contains(t, str, "not_found")
	assert.Contains(t, str, "404")
	assert.Contains(t, str, "Monitor")
	assert.Contains(t, str, "mon_123")
	assert.Contains(t, str, "read")
}

func TestDetectErrorContext_ServerErrorVariants(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		statusCode int
	}{
		{"500 internal server error", 500},
		{"502 bad gateway", 502},
		{"503 service unavailable", 503},
		{"504 gateway timeout", 504},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := client.NewAPIError(tc.statusCode, tc.name)
			ctx := DetectErrorContext("Monitor", "mon_123", "read", err)

			assert.Equal(t, "server_error", ctx.Type)
			assert.Equal(t, tc.statusCode, ctx.HTTPStatus)
		})
	}
}

func TestDetectErrorContext_ValidationErrorVariants(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		statusCode int
	}{
		{"400 bad request", 400},
		{"422 unprocessable entity", 422},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := client.NewAPIError(tc.statusCode, tc.name)
			ctx := DetectErrorContext("Monitor", "", "create", err)

			assert.Equal(t, "validation", ctx.Type)
			assert.Equal(t, tc.statusCode, ctx.HTTPStatus)
		})
	}
}
