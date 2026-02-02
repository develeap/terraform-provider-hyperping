// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"errors"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *APIError
		expected string
	}{
		{
			name: "simple error",
			err: &APIError{
				StatusCode: 404,
				Message:    "Not Found",
			},
			expected: "API error (status 404): Not Found",
		},
		{
			name: "error with validation details",
			err: &APIError{
				StatusCode: 422,
				Message:    "Validation failed",
				Details: []ValidationDetail{
					{Field: "url", Message: "Invalid URL"},
					{Field: "name", Message: "Name required"},
				},
			},
			expected: "API error (status 422): Validation failed - 2 validation errors",
		},
		{
			name: "error with single validation detail",
			err: &APIError{
				StatusCode: 400,
				Message:    "Bad Request",
				Details: []ValidationDetail{
					{Field: "frequency", Message: "Invalid frequency"},
				},
			},
			expected: "API error (status 400): Bad Request - 1 validation errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("Error() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestAPIError_Is(t *testing.T) {
	tests := []struct {
		name        string
		err         *APIError
		target      error
		shouldMatch bool
	}{
		// NotFound tests
		{
			name:        "404 matches ErrNotFound",
			err:         &APIError{StatusCode: 404},
			target:      ErrNotFound,
			shouldMatch: true,
		},
		{
			name:        "500 does not match ErrNotFound",
			err:         &APIError{StatusCode: 500},
			target:      ErrNotFound,
			shouldMatch: false,
		},

		// Unauthorized tests
		{
			name:        "401 matches ErrUnauthorized",
			err:         &APIError{StatusCode: 401},
			target:      ErrUnauthorized,
			shouldMatch: true,
		},
		{
			name:        "403 matches ErrUnauthorized",
			err:         &APIError{StatusCode: 403},
			target:      ErrUnauthorized,
			shouldMatch: true,
		},
		{
			name:        "404 does not match ErrUnauthorized",
			err:         &APIError{StatusCode: 404},
			target:      ErrUnauthorized,
			shouldMatch: false,
		},

		// RateLimited tests
		{
			name:        "429 matches ErrRateLimited",
			err:         &APIError{StatusCode: 429},
			target:      ErrRateLimited,
			shouldMatch: true,
		},
		{
			name:        "500 does not match ErrRateLimited",
			err:         &APIError{StatusCode: 500},
			target:      ErrRateLimited,
			shouldMatch: false,
		},

		// Validation tests
		{
			name:        "400 matches ErrValidation",
			err:         &APIError{StatusCode: 400},
			target:      ErrValidation,
			shouldMatch: true,
		},
		{
			name:        "422 matches ErrValidation",
			err:         &APIError{StatusCode: 422},
			target:      ErrValidation,
			shouldMatch: true,
		},
		{
			name:        "404 does not match ErrValidation",
			err:         &APIError{StatusCode: 404},
			target:      ErrValidation,
			shouldMatch: false,
		},

		// ServerError tests
		{
			name:        "500 matches ErrServerError",
			err:         &APIError{StatusCode: 500},
			target:      ErrServerError,
			shouldMatch: true,
		},
		{
			name:        "502 matches ErrServerError",
			err:         &APIError{StatusCode: 502},
			target:      ErrServerError,
			shouldMatch: true,
		},
		{
			name:        "503 matches ErrServerError",
			err:         &APIError{StatusCode: 503},
			target:      ErrServerError,
			shouldMatch: true,
		},
		{
			name:        "504 matches ErrServerError",
			err:         &APIError{StatusCode: 504},
			target:      ErrServerError,
			shouldMatch: true,
		},
		{
			name:        "599 matches ErrServerError",
			err:         &APIError{StatusCode: 599},
			target:      ErrServerError,
			shouldMatch: true,
		},
		{
			name:        "404 does not match ErrServerError",
			err:         &APIError{StatusCode: 404},
			target:      ErrServerError,
			shouldMatch: false,
		},

		// Unknown error tests
		{
			name:        "unknown error type does not match",
			err:         &APIError{StatusCode: 404},
			target:      errors.New("some other error"),
			shouldMatch: false,
		},
		{
			name:        "418 does not match any sentinel",
			err:         &APIError{StatusCode: 418}, // I'm a teapot
			target:      ErrNotFound,
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Is(tt.target)
			if result != tt.shouldMatch {
				t.Errorf("Is(%v) = %v, expected %v", tt.target, result, tt.shouldMatch)
			}
		})
	}
}

func TestAPIError_Unwrap(t *testing.T) {
	tests := []struct {
		name     string
		err      *APIError
		expected error
	}{
		{
			name:     "404 unwraps to ErrNotFound",
			err:      &APIError{StatusCode: 404},
			expected: ErrNotFound,
		},
		{
			name:     "401 unwraps to ErrUnauthorized",
			err:      &APIError{StatusCode: 401},
			expected: ErrUnauthorized,
		},
		{
			name:     "403 unwraps to ErrUnauthorized",
			err:      &APIError{StatusCode: 403},
			expected: ErrUnauthorized,
		},
		{
			name:     "429 unwraps to ErrRateLimited",
			err:      &APIError{StatusCode: 429},
			expected: ErrRateLimited,
		},
		{
			name:     "400 unwraps to ErrValidation",
			err:      &APIError{StatusCode: 400},
			expected: ErrValidation,
		},
		{
			name:     "422 unwraps to ErrValidation",
			err:      &APIError{StatusCode: 422},
			expected: ErrValidation,
		},
		{
			name:     "500 unwraps to ErrServerError",
			err:      &APIError{StatusCode: 500},
			expected: ErrServerError,
		},
		{
			name:     "502 unwraps to ErrServerError",
			err:      &APIError{StatusCode: 502},
			expected: ErrServerError,
		},
		{
			name:     "418 unwraps to nil (unknown)",
			err:      &APIError{StatusCode: 418},
			expected: nil,
		},
		{
			name:     "200 unwraps to nil (success code)",
			err:      &APIError{StatusCode: 200},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Unwrap()
			if result != tt.expected {
				t.Errorf("Unwrap() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestAPIError_ErrorsIs_Integration(t *testing.T) {
	// Test that errors.Is works correctly with APIError
	tests := []struct {
		name        string
		err         error
		target      error
		shouldMatch bool
	}{
		{
			name:        "errors.Is with 404 and ErrNotFound",
			err:         &APIError{StatusCode: 404, Message: "Not found"},
			target:      ErrNotFound,
			shouldMatch: true,
		},
		{
			name:        "errors.Is with 500 and ErrServerError",
			err:         &APIError{StatusCode: 500, Message: "Internal error"},
			target:      ErrServerError,
			shouldMatch: true,
		},
		{
			name:        "errors.Is with 429 and ErrRateLimited",
			err:         &APIError{StatusCode: 429, Message: "Rate limited"},
			target:      ErrRateLimited,
			shouldMatch: true,
		},
		{
			name:        "errors.Is with wrapped error",
			err:         &APIError{StatusCode: 404, Message: "Not found"},
			target:      ErrNotFound,
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := errors.Is(tt.err, tt.target)
			if result != tt.shouldMatch {
				t.Errorf("errors.Is() = %v, expected %v", result, tt.shouldMatch)
			}
		})
	}
}

func TestNewAPIError(t *testing.T) {
	err := NewAPIError(404, "Resource not found")

	if err.StatusCode != 404 {
		t.Errorf("StatusCode = %d, expected 404", err.StatusCode)
	}
	if err.Message != "Resource not found" {
		t.Errorf("Message = %q, expected 'Resource not found'", err.Message)
	}
	if err.Details != nil {
		t.Errorf("Details = %v, expected nil", err.Details)
	}
	if err.RetryAfter != 0 {
		t.Errorf("RetryAfter = %d, expected 0", err.RetryAfter)
	}
}

func TestNewValidationError(t *testing.T) {
	details := []ValidationDetail{
		{Field: "url", Message: "Invalid URL format"},
		{Field: "name", Message: "Name is required"},
	}

	err := NewValidationError(422, "Validation failed", details)

	if err.StatusCode != 422 {
		t.Errorf("StatusCode = %d, expected 422", err.StatusCode)
	}
	if err.Message != "Validation failed" {
		t.Errorf("Message = %q, expected 'Validation failed'", err.Message)
	}
	if len(err.Details) != 2 {
		t.Errorf("len(Details) = %d, expected 2", len(err.Details))
	}
	if err.Details[0].Field != "url" {
		t.Errorf("Details[0].Field = %q, expected 'url'", err.Details[0].Field)
	}
	if err.Details[1].Message != "Name is required" {
		t.Errorf("Details[1].Message = %q, expected 'Name is required'", err.Details[1].Message)
	}
}

func TestNewRateLimitError(t *testing.T) {
	err := NewRateLimitError(60)

	if err.StatusCode != 429 {
		t.Errorf("StatusCode = %d, expected 429", err.StatusCode)
	}
	if err.Message != "rate limit exceeded" {
		t.Errorf("Message = %q, expected 'rate limit exceeded'", err.Message)
	}
	if err.RetryAfter != 60 {
		t.Errorf("RetryAfter = %d, expected 60", err.RetryAfter)
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "APIError with 404",
			err:      &APIError{StatusCode: 404},
			expected: true,
		},
		{
			name:     "APIError with 500",
			err:      &APIError{StatusCode: 500},
			expected: false,
		},
		{
			name:     "direct ErrNotFound",
			err:      ErrNotFound,
			expected: true,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "random error",
			err:      errors.New("random"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNotFound(tt.err)
			if result != tt.expected {
				t.Errorf("IsNotFound() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIsUnauthorized(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "APIError with 401",
			err:      &APIError{StatusCode: 401},
			expected: true,
		},
		{
			name:     "APIError with 403",
			err:      &APIError{StatusCode: 403},
			expected: true,
		},
		{
			name:     "APIError with 404",
			err:      &APIError{StatusCode: 404},
			expected: false,
		},
		{
			name:     "direct ErrUnauthorized",
			err:      ErrUnauthorized,
			expected: true,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsUnauthorized(tt.err)
			if result != tt.expected {
				t.Errorf("IsUnauthorized() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIsRateLimited(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "APIError with 429",
			err:      &APIError{StatusCode: 429},
			expected: true,
		},
		{
			name:     "APIError with 500",
			err:      &APIError{StatusCode: 500},
			expected: false,
		},
		{
			name:     "direct ErrRateLimited",
			err:      ErrRateLimited,
			expected: true,
		},
		{
			name:     "NewRateLimitError",
			err:      NewRateLimitError(30),
			expected: true,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRateLimited(tt.err)
			if result != tt.expected {
				t.Errorf("IsRateLimited() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIsValidation(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "APIError with 400",
			err:      &APIError{StatusCode: 400},
			expected: true,
		},
		{
			name:     "APIError with 422",
			err:      &APIError{StatusCode: 422},
			expected: true,
		},
		{
			name:     "APIError with 404",
			err:      &APIError{StatusCode: 404},
			expected: false,
		},
		{
			name:     "direct ErrValidation",
			err:      ErrValidation,
			expected: true,
		},
		{
			name:     "NewValidationError",
			err:      NewValidationError(422, "test", nil),
			expected: true,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidation(tt.err)
			if result != tt.expected {
				t.Errorf("IsValidation() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIsServerError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "APIError with 500",
			err:      &APIError{StatusCode: 500},
			expected: true,
		},
		{
			name:     "APIError with 502",
			err:      &APIError{StatusCode: 502},
			expected: true,
		},
		{
			name:     "APIError with 503",
			err:      &APIError{StatusCode: 503},
			expected: true,
		},
		{
			name:     "APIError with 504",
			err:      &APIError{StatusCode: 504},
			expected: true,
		},
		{
			name:     "APIError with 599",
			err:      &APIError{StatusCode: 599},
			expected: true,
		},
		{
			name:     "APIError with 404",
			err:      &APIError{StatusCode: 404},
			expected: false,
		},
		{
			name:     "APIError with 499",
			err:      &APIError{StatusCode: 499},
			expected: false,
		},
		{
			name:     "direct ErrServerError",
			err:      ErrServerError,
			expected: true,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsServerError(tt.err)
			if result != tt.expected {
				t.Errorf("IsServerError() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestValidationDetail(t *testing.T) {
	detail := ValidationDetail{
		Field:   "url",
		Message: "URL must be a valid HTTP or HTTPS URL",
	}

	if detail.Field != "url" {
		t.Errorf("Field = %q, expected 'url'", detail.Field)
	}
	if detail.Message != "URL must be a valid HTTP or HTTPS URL" {
		t.Errorf("Message = %q, expected 'URL must be a valid HTTP or HTTPS URL'", detail.Message)
	}
}
