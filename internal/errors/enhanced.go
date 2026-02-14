// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

// Package terraflyerrors provides enhanced error types with auto-remediation suggestions,
// documentation links, and context-specific troubleshooting steps.
package terraflyerrors

import (
	"fmt"
	"strings"
	"time"
)

// EnhancedError represents an enhanced error with contextual information,
// remediation suggestions, and documentation links.
type EnhancedError struct {
	// What happened
	Title       string
	Description string

	// Context
	Operation string // "create", "read", "update", "delete", "import"
	Resource  string // Resource address (e.g., "hyperping_monitor.prod_api")
	Field     string // Specific field if applicable

	// Remediation
	Suggestions []string
	Commands    []string
	DocLinks    []string
	Examples    []string

	// Auto-remediation
	Retryable  bool
	RetryAfter *time.Duration

	// Original error
	Underlying error
}

// Error implements the error interface with rich, formatted output.
func (e *EnhancedError) Error() string {
	return e.Format()
}

// Format returns a formatted, user-friendly error message.
func (e *EnhancedError) Format() string {
	var b strings.Builder

	// Title with icon
	b.WriteString("\n")
	b.WriteString(formatTitle(e.Title))
	b.WriteString("\n\n")

	// Description
	if e.Description != "" {
		b.WriteString(e.Description)
		b.WriteString("\n\n")
	}

	// Context
	if e.Resource != "" {
		b.WriteString(fmt.Sprintf("Resource:  %s\n", e.Resource))
	}
	if e.Operation != "" {
		b.WriteString(fmt.Sprintf("Operation: %s\n", e.Operation))
	}
	if e.Field != "" {
		b.WriteString(fmt.Sprintf("Field:     %s\n", e.Field))
	}

	// Add blank line after context
	if e.Resource != "" || e.Operation != "" || e.Field != "" {
		b.WriteString("\n")
	}

	// Suggestions
	if len(e.Suggestions) > 0 {
		b.WriteString("💡 Suggestions:\n")
		for _, s := range e.Suggestions {
			b.WriteString(fmt.Sprintf("  • %s\n", s))
		}
		b.WriteString("\n")
	}

	// Commands to try
	if len(e.Commands) > 0 {
		b.WriteString("🔧 Try:\n")
		for _, c := range e.Commands {
			b.WriteString(fmt.Sprintf("  $ %s\n", c))
		}
		b.WriteString("\n")
	}

	// Examples
	if len(e.Examples) > 0 {
		b.WriteString("📝 Examples:\n")
		for _, ex := range e.Examples {
			b.WriteString(fmt.Sprintf("  %s\n", ex))
		}
		b.WriteString("\n")
	}

	// Documentation links
	if len(e.DocLinks) > 0 {
		b.WriteString("📚 Documentation:\n")
		for _, link := range e.DocLinks {
			b.WriteString(fmt.Sprintf("  %s\n", link))
		}
		b.WriteString("\n")
	}

	// Retry information
	if e.RetryAfter != nil {
		b.WriteString(fmt.Sprintf("⏰ Auto-retry after: %v\n", *e.RetryAfter))
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
}

// Unwrap returns the underlying error for error chain unwrapping.
func (e *EnhancedError) Unwrap() error {
	return e.Underlying
}

// formatTitle adds appropriate icon and formatting to the error title.
func formatTitle(title string) string {
	return fmt.Sprintf("❌ %s", title)
}

// ErrorCategory represents the category of error for suggestion generation.
type ErrorCategory string

const (
	CategoryAuth       ErrorCategory = "auth"
	CategoryRateLimit  ErrorCategory = "rate_limit"
	CategoryValidation ErrorCategory = "validation"
	CategoryNotFound   ErrorCategory = "not_found"
	CategoryServer     ErrorCategory = "server"
	CategoryNetwork    ErrorCategory = "network"
	CategoryCircuit    ErrorCategory = "circuit_breaker"
	CategoryUnknown    ErrorCategory = "unknown"
)

// EnhanceError wraps an error with context and suggestions based on error type.
func EnhanceError(err error, category ErrorCategory, opts ...EnhancementOption) error {
	if err == nil {
		return nil
	}

	enhanced := &EnhancedError{
		Underlying: err,
	}

	// Apply options
	for _, opt := range opts {
		opt(enhanced)
	}

	// Apply category-specific enhancements
	applyCategoryDefaults(enhanced, category)

	return enhanced
}

// EnhancementOption is a functional option for enhancing errors.
type EnhancementOption func(*EnhancedError)

// WithTitle sets the error title.
func WithTitle(title string) EnhancementOption {
	return func(e *EnhancedError) {
		e.Title = title
	}
}

// WithDescription sets the error description.
func WithDescription(desc string) EnhancementOption {
	return func(e *EnhancedError) {
		e.Description = desc
	}
}

// WithOperation sets the operation context.
func WithOperation(op string) EnhancementOption {
	return func(e *EnhancedError) {
		e.Operation = op
	}
}

// WithResource sets the resource context.
func WithResource(resource string) EnhancementOption {
	return func(e *EnhancedError) {
		e.Resource = resource
	}
}

// WithField sets the field context.
func WithField(field string) EnhancementOption {
	return func(e *EnhancedError) {
		e.Field = field
	}
}

// WithSuggestions adds remediation suggestions.
func WithSuggestions(suggestions ...string) EnhancementOption {
	return func(e *EnhancedError) {
		e.Suggestions = append(e.Suggestions, suggestions...)
	}
}

// WithCommands adds command suggestions.
func WithCommands(commands ...string) EnhancementOption {
	return func(e *EnhancedError) {
		e.Commands = append(e.Commands, commands...)
	}
}

// WithDocLinks adds documentation links.
func WithDocLinks(links ...string) EnhancementOption {
	return func(e *EnhancedError) {
		e.DocLinks = append(e.DocLinks, links...)
	}
}

// WithExamples adds usage examples.
func WithExamples(examples ...string) EnhancementOption {
	return func(e *EnhancedError) {
		e.Examples = append(e.Examples, examples...)
	}
}

// WithRetryable marks the error as retryable.
func WithRetryable(retryable bool) EnhancementOption {
	return func(e *EnhancedError) {
		e.Retryable = retryable
	}
}

// WithRetryAfter sets the retry delay.
func WithRetryAfter(duration time.Duration) EnhancementOption {
	return func(e *EnhancedError) {
		e.RetryAfter = &duration
	}
}

// applyCategoryDefaults applies default suggestions and links based on error category.
func applyCategoryDefaults(e *EnhancedError, category ErrorCategory) {
	switch category {
	case CategoryAuth:
		applyAuthDefaults(e)
	case CategoryRateLimit:
		applyRateLimitDefaults(e)
	case CategoryValidation:
		applyValidationDefaults(e)
	case CategoryNotFound:
		applyNotFoundDefaults(e)
	case CategoryServer:
		applyServerDefaults(e)
	case CategoryNetwork:
		applyNetworkDefaults(e)
	case CategoryCircuit:
		applyCircuitBreakerDefaults(e)
	}
}

// applyAuthDefaults applies defaults for authentication errors.
func applyAuthDefaults(e *EnhancedError) {
	if e.Title == "" {
		e.Title = "Authentication Failed"
	}
	if e.Description == "" {
		e.Description = "Your Hyperping API key is invalid or has expired."
	}

	e.Suggestions = append(e.Suggestions,
		"Verify your API key is correct (should start with 'sk_')",
		"Check if the API key has been revoked in the Hyperping dashboard",
		"Ensure HYPERPING_API_KEY environment variable is set correctly",
		"Generate a new API key if needed",
	)

	e.Commands = append(e.Commands,
		"echo $HYPERPING_API_KEY                    # Verify key is set",
		"terraform plan                             # Test with valid credentials",
	)

	e.DocLinks = append(e.DocLinks,
		"https://registry.terraform.io/providers/develeap/hyperping/latest/docs#authentication",
		"https://app.hyperping.io/settings/api        # Generate new API key",
		"https://github.com/develeap/terraform-provider-hyperping/tree/main/docs/guides/troubleshooting.md#authentication-errors",
	)
}

// applyRateLimitDefaults applies defaults for rate limit errors.
func applyRateLimitDefaults(e *EnhancedError) {
	if e.Title == "" {
		e.Title = "Rate Limit Exceeded"
	}
	if e.Description == "" {
		e.Description = "You've hit the Hyperping API rate limit. Your request will automatically retry."
	}

	e.Retryable = true

	e.Suggestions = append(e.Suggestions,
		"Reduce the number of parallel terraform operations",
		"Use terraform apply with -parallelism=1 flag for serial execution",
		"Consider upgrading your Hyperping plan for higher rate limits",
		"Use bulk operations where possible instead of individual creates",
	)

	e.Commands = append(e.Commands,
		"terraform apply -parallelism=1             # Reduce concurrent requests",
		"terraform apply -refresh=false             # Skip refresh to reduce API calls",
	)

	e.DocLinks = append(e.DocLinks,
		"https://github.com/develeap/terraform-provider-hyperping/tree/main/docs/guides/rate-limits.md",
		"https://api.hyperping.io/docs#rate-limits",
	)
}

// applyValidationDefaults applies defaults for validation errors.
func applyValidationDefaults(e *EnhancedError) {
	if e.Title == "" {
		e.Title = "Validation Error"
	}
	if e.Description == "" {
		e.Description = "The provided value does not meet the field's validation requirements."
	}

	e.DocLinks = append(e.DocLinks,
		"https://registry.terraform.io/providers/develeap/hyperping/latest/docs",
		"https://github.com/develeap/terraform-provider-hyperping/tree/main/docs/guides/validation.md",
	)
}

// applyNotFoundDefaults applies defaults for not found errors.
func applyNotFoundDefaults(e *EnhancedError) {
	if e.Title == "" {
		e.Title = "Resource Not Found"
	}
	if e.Description == "" {
		e.Description = "The requested resource does not exist or has been deleted."
	}

	e.Suggestions = append(e.Suggestions,
		"Verify the resource ID is correct",
		"Check if the resource was deleted outside of Terraform",
		"Use 'terraform import' to sync existing resources with your state",
		"Check the Hyperping dashboard to confirm resource existence",
	)

	e.DocLinks = append(e.DocLinks,
		"https://registry.terraform.io/providers/develeap/hyperping/latest/docs/guides/import",
		"https://app.hyperping.io                   # View resources in dashboard",
	)
}

// applyServerDefaults applies defaults for server errors.
func applyServerDefaults(e *EnhancedError) {
	if e.Title == "" {
		e.Title = "Server Error"
	}
	if e.Description == "" {
		e.Description = "The Hyperping API returned a server error. Your request will automatically retry."
	}

	e.Retryable = true

	e.Suggestions = append(e.Suggestions,
		"The error is temporary and will automatically retry",
		"Check the Hyperping status page for any ongoing incidents",
		"If the problem persists, contact Hyperping support",
	)

	e.DocLinks = append(e.DocLinks,
		"https://status.hyperping.io                # Check service status",
		"https://hyperping.io/support               # Contact support",
	)
}

// applyNetworkDefaults applies defaults for network errors.
func applyNetworkDefaults(e *EnhancedError) {
	if e.Title == "" {
		e.Title = "Network Error"
	}
	if e.Description == "" {
		e.Description = "Unable to connect to the Hyperping API."
	}

	e.Suggestions = append(e.Suggestions,
		"Check your internet connection",
		"Verify firewall settings allow connections to api.hyperping.io",
		"Check if a proxy is required and configured correctly",
		"Try accessing the API directly with curl to diagnose",
	)

	e.Commands = append(e.Commands,
		"curl -H \"Authorization: Bearer $HYPERPING_API_KEY\" https://api.hyperping.io/v1/monitors",
		"ping api.hyperping.io                      # Test DNS resolution",
	)

	e.DocLinks = append(e.DocLinks,
		"https://github.com/develeap/terraform-provider-hyperping/tree/main/docs/guides/troubleshooting.md#network-errors",
	)
}

// applyCircuitBreakerDefaults applies defaults for circuit breaker errors.
func applyCircuitBreakerDefaults(e *EnhancedError) {
	if e.Title == "" {
		e.Title = "Circuit Breaker Open"
	}
	if e.Description == "" {
		e.Description = "Too many consecutive failures have occurred. The circuit breaker is temporarily blocking requests to prevent cascading failures."
	}

	e.Suggestions = append(e.Suggestions,
		"Wait 30 seconds for the circuit breaker to reset",
		"Check the Hyperping status page for any ongoing incidents",
		"Verify your API credentials are correct",
		"Review recent errors in terraform output for the root cause",
	)

	e.DocLinks = append(e.DocLinks,
		"https://status.hyperping.io                # Check service status",
		"https://github.com/develeap/terraform-provider-hyperping/tree/main/docs/guides/troubleshooting.md#circuit-breaker",
	)
}
