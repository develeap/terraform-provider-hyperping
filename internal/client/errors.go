// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"errors"
	"fmt"
	"regexp"
)

// Common errors returned by the client.
var (
	// ErrNotFound is returned when a resource is not found (404).
	ErrNotFound = errors.New("resource not found")

	// ErrUnauthorized is returned when authentication fails (401/403).
	ErrUnauthorized = errors.New("unauthorized: invalid or missing API key")

	// ErrRateLimited is returned when rate limit is exceeded (429).
	ErrRateLimited = errors.New("rate limit exceeded")

	// ErrValidation is returned when request validation fails (400/422).
	ErrValidation = errors.New("validation error")

	// ErrServerError is returned for server errors (5xx).
	ErrServerError = errors.New("server error")
)

// APIError represents an error returned by the Hyperping API.
type APIError struct {
	StatusCode int
	Message    string
	Details    []ValidationDetail
	RetryAfter int // seconds to wait before retrying (for rate limits)
}

// ValidationDetail represents a field-level validation error.
type ValidationDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements the error interface.
// The message is sanitized to remove sensitive information like API keys.
func (e *APIError) Error() string {
	sanitized := sanitizeMessage(e.Message)

	// Enhanced rate limit error messages
	if e.StatusCode == 429 && e.RetryAfter > 0 {
		return fmt.Sprintf("API error (status %d): %s - retry after %d seconds", e.StatusCode, sanitized, e.RetryAfter)
	}

	if len(e.Details) > 0 {
		return fmt.Sprintf("API error (status %d): %s - %d validation errors", e.StatusCode, sanitized, len(e.Details))
	}
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, sanitized)
}

// Is checks if the error matches a target error.
func (e *APIError) Is(target error) bool {
	switch target {
	case ErrNotFound:
		return e.StatusCode == 404
	case ErrUnauthorized:
		return e.StatusCode == 401 || e.StatusCode == 403
	case ErrRateLimited:
		return e.StatusCode == 429
	case ErrValidation:
		return e.StatusCode == 400 || e.StatusCode == 422
	case ErrServerError:
		return e.StatusCode >= 500
	}
	return false
}

// Unwrap returns the underlying error based on status code.
func (e *APIError) Unwrap() error {
	switch {
	case e.StatusCode == 404:
		return ErrNotFound
	case e.StatusCode == 401 || e.StatusCode == 403:
		return ErrUnauthorized
	case e.StatusCode == 429:
		return ErrRateLimited
	case e.StatusCode == 400 || e.StatusCode == 422:
		return ErrValidation
	case e.StatusCode >= 500:
		return ErrServerError
	}
	return nil
}

// NewAPIError creates a new APIError from an HTTP response.
func NewAPIError(statusCode int, message string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
	}
}

// NewValidationError creates a new APIError with validation details.
func NewValidationError(statusCode int, message string, details []ValidationDetail) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		Details:    details,
	}
}

// NewRateLimitError creates a new APIError for rate limiting with retry-after.
func NewRateLimitError(retryAfter int) *APIError {
	return &APIError{
		StatusCode: 429,
		Message:    "rate limit exceeded",
		RetryAfter: retryAfter,
	}
}

// Compile regexes once for performance.
// apiKeyPattern matches both sk_ and hp_ prefixed keys with alphanumeric, underscore, and hyphen (VULN-019).
// bearerPattern matches any Bearer token of 8+ non-whitespace chars (VULN-019).
var (
	apiKeyPattern     = regexp.MustCompile(`sk_[a-zA-Z0-9_-]+`)
	bearerPattern     = regexp.MustCompile(`Bearer\s+(?:[^\s]*[0-9_-][^\s]*|[^\s]{32,})`)
	urlCredPattern    = regexp.MustCompile(`://[^:]+:[^@]+@`)
	authHeaderPattern = regexp.MustCompile(`Authorization:\s+Bearer\s+[^\s]+`)
)

// sanitizeMessage removes sensitive information from error messages.
// This prevents API keys, tokens, and credentials from being exposed in logs or error output.
func sanitizeMessage(msg string) string {
	// Replace Hyperping API keys (sk_alphanumeric) with redacted placeholder
	msg = apiKeyPattern.ReplaceAllString(msg, "sk_***REDACTED***")

	// Replace Authorization headers first (more specific pattern)
	msg = authHeaderPattern.ReplaceAllString(msg, "Authorization: ***REDACTED***")

	// Replace Bearer tokens
	msg = bearerPattern.ReplaceAllString(msg, "Bearer ***REDACTED***")

	// Replace credentials in URLs (https://user:pass@domain.com)
	msg = urlCredPattern.ReplaceAllString(msg, "://***REDACTED***@")

	return msg
}

// IsNotFound checks if an error is a not found error.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsUnauthorized checks if an error is an unauthorized error.
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsRateLimited checks if an error is a rate limit error.
func IsRateLimited(err error) bool {
	return errors.Is(err, ErrRateLimited)
}

// IsValidation checks if an error is a validation error.
func IsValidation(err error) bool {
	return errors.Is(err, ErrValidation)
}

// IsServerError checks if an error is a server error.
func IsServerError(err error) bool {
	return errors.Is(err, ErrServerError)
}
