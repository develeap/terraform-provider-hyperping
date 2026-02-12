// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// ErrorContext provides structured error information for enhanced error messages.
// This context is used to generate actionable troubleshooting steps for users.
type ErrorContext struct {
	Type         string // "not_found", "auth_error", "rate_limit", "server_error", "validation", "unknown"
	HTTPStatus   int
	RetryAfter   int    // seconds (for rate limit errors)
	ResourceType string // "Monitor", "Incident", "Maintenance", etc.
	ResourceID   string
	Operation    string // "read", "create", "update", "delete"
	Message      string // Original error message
}

// DetectErrorContext analyzes an error and returns structured context.
// This function identifies the error type and extracts relevant information
// to provide context-specific troubleshooting guidance.
func DetectErrorContext(resourceType, resourceID, operation string, err error) ErrorContext {
	if err == nil {
		return ErrorContext{
			ResourceType: resourceType,
			ResourceID:   resourceID,
			Operation:    operation,
			Type:         "unknown",
		}
	}

	ctx := ErrorContext{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Operation:    operation,
		Message:      err.Error(),
		Type:         "unknown",
	}

	// Detect error type from client package using error checking functions
	switch {
	case client.IsNotFound(err):
		ctx.Type = "not_found"
		ctx.HTTPStatus = 404
	case client.IsUnauthorized(err):
		ctx.Type = "auth_error"
		// Determine if 401 or 403 from message
		if strings.Contains(strings.ToLower(err.Error()), "403") {
			ctx.HTTPStatus = 403
		} else {
			ctx.HTTPStatus = 401
		}
	case client.IsRateLimited(err):
		ctx.Type = "rate_limit"
		ctx.HTTPStatus = 429
		ctx.RetryAfter = extractRetryAfter(err)
	case client.IsServerError(err):
		ctx.Type = "server_error"
		ctx.HTTPStatus = extractStatusCode(err, 500)
	case client.IsValidation(err):
		ctx.Type = "validation"
		ctx.HTTPStatus = extractStatusCode(err, 400)
	}

	return ctx
}

// extractRetryAfter parses Retry-After information from error message.
// It supports multiple formats:
// - "retry after 60 seconds"
// - "retry after 120 second"
// - API error messages with retry-after embedded
//
// Returns the number of seconds to wait, or 60 as a default if not found.
func extractRetryAfter(err error) int {
	if err == nil {
		return 60
	}

	msg := err.Error()

	// Try to find "retry after X seconds" pattern (case-insensitive)
	re := regexp.MustCompile(`retry after (\d+) seconds?`)
	matches := re.FindStringSubmatch(strings.ToLower(msg))
	if len(matches) > 1 {
		if seconds, parseErr := strconv.Atoi(matches[1]); parseErr == nil {
			return seconds
		}
	}

	// Default to 60 seconds if pattern not found
	return 60
}

// extractStatusCode attempts to extract HTTP status code from error message.
// Returns the provided default if extraction fails.
func extractStatusCode(err error, defaultCode int) int {
	if err == nil {
		return defaultCode
	}

	msg := err.Error()

	// Try to find "status XXX" pattern
	re := regexp.MustCompile(`status (\d{3})`)
	matches := re.FindStringSubmatch(msg)
	if len(matches) > 1 {
		if code, parseErr := strconv.Atoi(matches[1]); parseErr == nil {
			return code
		}
	}

	return defaultCode
}

// String returns a human-readable representation of the error context.
func (ctx ErrorContext) String() string {
	return fmt.Sprintf("ErrorContext{Type: %s, Status: %d, Resource: %s, ID: %s, Operation: %s}",
		ctx.Type, ctx.HTTPStatus, ctx.ResourceType, ctx.ResourceID, ctx.Operation)
}
