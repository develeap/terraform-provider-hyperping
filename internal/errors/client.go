// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package terraflyerrors

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// EnhanceClientError enhances errors from the Hyperping API client with
// context-specific suggestions and documentation links.
func EnhanceClientError(err error, operation, resource, field string) error {
	if err == nil {
		return nil
	}

	// Check if already enhanced
	var enhanced *EnhancedError
	if errors.As(err, &enhanced) {
		return err
	}

	// Extract API error information
	var apiErr *client.APIError
	if errors.As(err, &apiErr) {
		return enhanceAPIError(apiErr, operation, resource, field)
	}

	// Handle standard client errors
	switch {
	case errors.Is(err, client.ErrUnauthorized):
		return enhanceAuthError(err, operation, resource)
	case errors.Is(err, client.ErrRateLimited):
		return enhanceRateLimitError(err, operation, resource)
	case errors.Is(err, client.ErrValidation):
		return enhanceValidationError(err, operation, resource, field)
	case errors.Is(err, client.ErrNotFound):
		return enhanceNotFoundError(err, operation, resource)
	case errors.Is(err, client.ErrServerError):
		return enhanceServerError(err, operation, resource)
	}

	// Check for network errors
	if isNetworkError(err) {
		return enhanceNetworkError(err, operation, resource)
	}

	// Check for circuit breaker errors
	if isCircuitBreakerError(err) {
		return enhanceCircuitBreakerError(err, operation, resource)
	}

	// Generic enhancement for unknown errors
	return EnhanceError(err, CategoryUnknown,
		WithTitle("Unexpected Error"),
		WithDescription(err.Error()),
		WithOperation(operation),
		WithResource(resource),
	)
}

// enhanceAPIError enhances an APIError with detailed context.
func enhanceAPIError(apiErr *client.APIError, operation, resource, field string) error {
	// Determine category
	category := CategoryUnknown
	switch {
	case errors.Is(apiErr, client.ErrUnauthorized):
		category = CategoryAuth
	case errors.Is(apiErr, client.ErrRateLimited):
		category = CategoryRateLimit
	case errors.Is(apiErr, client.ErrValidation):
		category = CategoryValidation
	case errors.Is(apiErr, client.ErrNotFound):
		category = CategoryNotFound
	case errors.Is(apiErr, client.ErrServerError):
		category = CategoryServer
	}

	opts := []EnhancementOption{
		WithOperation(operation),
		WithResource(resource),
	}

	if field != "" {
		opts = append(opts, WithField(field))
	}

	// Add retry information for rate limit errors
	if apiErr.RetryAfter > 0 {
		retryAfter := time.Duration(apiErr.RetryAfter) * time.Second
		opts = append(opts, WithRetryAfter(retryAfter))
	}

	// Add validation details if present
	if len(apiErr.Details) > 0 {
		suggestions := make([]string, 0, len(apiErr.Details))
		for _, detail := range apiErr.Details {
			suggestions = append(suggestions,
				fmt.Sprintf("%s: %s", detail.Field, detail.Message))
		}
		opts = append(opts, WithSuggestions(suggestions...))
	}

	return EnhanceError(apiErr, category, opts...)
}

// enhanceAuthError enhances authentication errors.
func enhanceAuthError(err error, operation, resource string) error {
	return EnhanceError(err, CategoryAuth,
		WithOperation(operation),
		WithResource(resource),
	)
}

// enhanceRateLimitError enhances rate limit errors.
func enhanceRateLimitError(err error, operation, resource string) error {
	opts := []EnhancementOption{
		WithOperation(operation),
		WithResource(resource),
		WithRetryable(true),
	}

	// Try to extract retry-after from API error
	var apiErr *client.APIError
	if errors.As(err, &apiErr) && apiErr.RetryAfter > 0 {
		retryAfter := time.Duration(apiErr.RetryAfter) * time.Second
		opts = append(opts, WithRetryAfter(retryAfter))
	}

	return EnhanceError(err, CategoryRateLimit, opts...)
}

// enhanceValidationError enhances validation errors.
func enhanceValidationError(err error, operation, resource, field string) error {
	opts := []EnhancementOption{
		WithOperation(operation),
		WithResource(resource),
	}

	if field != "" {
		opts = append(opts, WithField(field))
	}

	// Try to extract validation details from API error
	var apiErr *client.APIError
	if errors.As(err, &apiErr) && len(apiErr.Details) > 0 {
		suggestions := make([]string, 0, len(apiErr.Details))
		for _, detail := range apiErr.Details {
			suggestions = append(suggestions,
				fmt.Sprintf("Field '%s': %s", detail.Field, detail.Message))
		}
		opts = append(opts, WithSuggestions(suggestions...))
	}

	return EnhanceError(err, CategoryValidation, opts...)
}

// enhanceNotFoundError enhances not found errors.
func enhanceNotFoundError(err error, operation, resource string) error {
	opts := []EnhancementOption{
		WithOperation(operation),
		WithResource(resource),
	}

	// Add import suggestion for read operations
	if operation == "read" && resource != "" {
		opts = append(opts,
			WithCommands(
				fmt.Sprintf("terraform import %s <resource_id>  # Sync state with existing resource", resource),
			),
		)
	}

	return EnhanceError(err, CategoryNotFound, opts...)
}

// enhanceServerError enhances server errors.
func enhanceServerError(err error, operation, resource string) error {
	return EnhanceError(err, CategoryServer,
		WithOperation(operation),
		WithResource(resource),
		WithRetryable(true),
	)
}

// enhanceNetworkError enhances network connectivity errors.
func enhanceNetworkError(err error, operation, resource string) error {
	return EnhanceError(err, CategoryNetwork,
		WithOperation(operation),
		WithResource(resource),
	)
}

// enhanceCircuitBreakerError enhances circuit breaker errors.
func enhanceCircuitBreakerError(err error, operation, resource string) error {
	return EnhanceError(err, CategoryCircuit,
		WithOperation(operation),
		WithResource(resource),
	)
}

// isNetworkError checks if an error is a network connectivity error.
func isNetworkError(err error) bool {
	errStr := strings.ToLower(err.Error())
	networkIndicators := []string{
		"connection refused",
		"no such host",
		"network is unreachable",
		"timeout",
		"dial tcp",
		"i/o timeout",
		"connection reset",
		"broken pipe",
	}

	for _, indicator := range networkIndicators {
		if strings.Contains(errStr, indicator) {
			return true
		}
	}

	return false
}

// isCircuitBreakerError checks if an error is from a circuit breaker.
func isCircuitBreakerError(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "circuit breaker") ||
		strings.Contains(errStr, "too many failures")
}

// FormatValidationErrors formats multiple validation errors from the API.
func FormatValidationErrors(details []client.ValidationDetail) string {
	if len(details) == 0 {
		return "validation failed"
	}

	parts := []string{}
	for _, detail := range details {
		parts = append(parts, fmt.Sprintf("  • %s: %s", detail.Field, detail.Message))
	}

	return fmt.Sprintf("Validation errors:\n%s", strings.Join(parts, "\n"))
}
