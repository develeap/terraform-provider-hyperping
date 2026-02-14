# Enhanced Error System

This package provides enhanced error handling with auto-remediation suggestions, documentation links, and context-specific troubleshooting steps.

## Features

- **Actionable Error Messages**: Every error includes "Try: <command>" suggestions
- **Auto-Remediation**: Automatic retry with exponential backoff for transient failures
- **Rate Limit Handling**: Calculates exact wait time from `Retry-After` header
- **Field-Specific Validation**: Shows allowed values and closest matches
- **Documentation Links**: Direct links to relevant docs for each error type
- **Context-Aware**: Different messages for create/read/update/delete operations
- **Smart Suggestions**: Detects typos and suggests corrections using Levenshtein distance

## Usage

### Basic Error Enhancement

```go
import "github.com/develeap/terraform-provider-hyperping/internal/errors"

// Enhance any error
enhanced := errors.EnhanceClientError(err, "create", "hyperping_monitor.prod", "frequency")
```

### Resource Operations

```go
// Create operation
if err != nil {
    enhanced := errors.CreateError(err, "hyperping_monitor", "prod_api")
    errors.AddDiagnosticError(&resp.Diagnostics, "", enhanced)
    return
}

// Read operation
if err != nil {
    enhanced := errors.ReadError(err, "hyperping_monitor", "prod_api")
    if errors.IsNotFound(err) {
        resp.State.RemoveResource(ctx)
        return
    }
    errors.AddDiagnosticError(&resp.Diagnostics, "", enhanced)
    return
}

// Update operation
if err != nil {
    enhanced := errors.UpdateError(err, "hyperping_monitor", "prod_api")
    errors.AddDiagnosticError(&resp.Diagnostics, "", enhanced)
    return
}

// Delete operation
if err != nil {
    enhanced := errors.DeleteError(err, "hyperping_monitor", "prod_api")
    errors.AddDiagnosticError(&resp.Diagnostics, "", enhanced)
    return
}

// Import operation
if err != nil {
    enhanced := errors.ImportError(err, "hyperping_monitor", importID)
    errors.AddDiagnosticError(&resp.Diagnostics, "", enhanced)
    return
}
```

### Field Validation

```go
// Validate monitor frequency
if !isValidFrequency(freq) {
    return errors.FrequencySuggestion(freq)
}

// Validate region
if !isValidRegion(region) {
    return errors.RegionSuggestion(region)
}

// Validate HTTP method
if !isValidMethod(method) {
    return errors.HTTPMethodSuggestion(method)
}

// Validate status code
if !isValidStatusCode(code) {
    return errors.StatusCodeSuggestion(code)
}

// Validate incident status
if !isValidIncidentStatus(status) {
    return errors.IncidentStatusSuggestion(status)
}

// Validate incident severity
if !isValidSeverity(severity) {
    return errors.SeveritySuggestion(severity)
}
```

### Custom Error Enhancement

```go
// Create custom enhanced error
err := errors.EnhanceError(
    originalErr,
    errors.CategoryValidation,
    errors.WithTitle("Custom Error"),
    errors.WithDescription("Something went wrong"),
    errors.WithOperation("create"),
    errors.WithResource("hyperping_monitor.test"),
    errors.WithField("url"),
    errors.WithSuggestions(
        "Check the URL format",
        "Ensure the URL is accessible",
    ),
    errors.WithCommands(
        "curl https://example.com/health",
    ),
    errors.WithExamples(
        `url = "https://api.example.com/health"`,
    ),
    errors.WithDocLinks(
        "https://registry.terraform.io/providers/develeap/hyperping/latest/docs",
    ),
)
```

## Error Categories

| Category | Description | Auto-Retry |
|----------|-------------|------------|
| `CategoryAuth` | Authentication errors (401, 403) | No |
| `CategoryRateLimit` | Rate limit errors (429) | Yes |
| `CategoryValidation` | Validation errors (400, 422) | No |
| `CategoryNotFound` | Resource not found (404) | No |
| `CategoryServer` | Server errors (5xx) | Yes |
| `CategoryNetwork` | Network connectivity errors | No |
| `CategoryCircuit` | Circuit breaker open | No |

## Example Error Output

### Authentication Error

```
❌ Authentication Failed

Your Hyperping API key is invalid or has expired.

Resource:  hyperping_monitor.prod_api
Operation: create

💡 Suggestions:
  • Verify your API key is correct (should start with 'sk_')
  • Check if the API key has been revoked in the Hyperping dashboard
  • Ensure HYPERPING_API_KEY environment variable is set correctly
  • Generate a new API key if needed

🔧 Try:
  $ echo $HYPERPING_API_KEY                    # Verify key is set
  $ terraform plan                             # Test with valid credentials

📚 Documentation:
  https://registry.terraform.io/providers/develeap/hyperping/latest/docs#authentication
  https://app.hyperping.io/settings/api        # Generate new API key
```

### Rate Limit Error

```
❌ Rate Limit Exceeded

You've hit the Hyperping API rate limit. Your request will automatically retry.

Resource:  hyperping_monitor.staging_api
Operation: create

⏰ Auto-retry after: 23 seconds (automatically retrying with exponential backoff)

💡 Suggestions:
  • Reduce the number of parallel terraform operations
  • Use terraform apply with -parallelism=1 flag for serial execution
  • Consider upgrading your Hyperping plan for higher rate limits
  • Use bulk operations where possible instead of individual creates

🔧 Try:
  $ terraform apply -parallelism=1             # Reduce concurrent requests
  $ terraform apply -refresh=false             # Skip refresh to reduce API calls

📚 Documentation:
  https://github.com/develeap/terraform-provider-hyperping/tree/main/docs/guides/rate-limits.md
```

### Validation Error with Suggestions

```
❌ Invalid Monitor Frequency

The 'check_frequency' field must be one of the allowed values.

Resource: hyperping_monitor.prod_api
Field:    check_frequency
Value:    45 (invalid)

💡 Allowed values (in seconds):
  10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400

Closest valid values to your input (45):
  • 30 seconds (15 seconds faster)
  • 60 seconds (15 seconds slower)

🔧 Try:
  frequency = 30  # Check every 30 seconds
  frequency = 60  # Check every minute

📚 Documentation:
  https://registry.terraform.io/providers/develeap/hyperping/latest/docs/resources/monitor#check_frequency
```

## API Reference

### Core Functions

#### `EnhanceClientError(err error, operation, resource, field string) error`

Enhances errors from the Hyperping API client with context-specific suggestions.

**Parameters:**
- `err`: Original error from client
- `operation`: Operation being performed ("create", "read", "update", "delete", "import")
- `resource`: Resource address (e.g., "hyperping_monitor.prod_api")
- `field`: Specific field name (optional, use "" if not applicable)

**Returns:** Enhanced error with suggestions and documentation links

#### `EnhanceError(err error, category ErrorCategory, opts ...EnhancementOption) error`

Creates an enhanced error with custom options.

**Parameters:**
- `err`: Original error
- `category`: Error category (CategoryAuth, CategoryRateLimit, etc.)
- `opts`: Enhancement options (WithTitle, WithDescription, WithSuggestions, etc.)

**Returns:** Enhanced error

### Helper Functions

#### `AddDiagnosticError(diags *diag.Diagnostics, summary string, err error)`

Adds an enhanced error to Terraform diagnostics.

#### `CreateError(err error, resourceType, resourceName string) error`

Creates an enhanced error for create operations.

#### `ReadError(err error, resourceType, resourceName string) error`

Creates an enhanced error for read operations.

#### `UpdateError(err error, resourceType, resourceName string) error`

Creates an enhanced error for update operations.

#### `DeleteError(err error, resourceType, resourceName string) error`

Creates an enhanced error for delete operations.

#### `ImportError(err error, resourceType, importID string) error`

Creates an enhanced error for import operations.

### Validation Suggestion Functions

#### `FrequencySuggestion(currentValue int64) *EnhancedError`

Generates suggestions for invalid monitor frequency values.

#### `RegionSuggestion(invalidRegion string) *EnhancedError`

Generates suggestions for invalid region values.

#### `HTTPMethodSuggestion(invalidMethod string) *EnhancedError`

Generates suggestions for invalid HTTP method values.

#### `StatusCodeSuggestion(currentValue string) *EnhancedError`

Generates suggestions for invalid status code values.

#### `IncidentStatusSuggestion(invalidStatus string) *EnhancedError`

Generates suggestions for invalid incident status values.

#### `SeveritySuggestion(invalidSeverity string) *EnhancedError`

Generates suggestions for invalid incident severity values.

## Enhancement Options

| Option | Description |
|--------|-------------|
| `WithTitle(title string)` | Set error title |
| `WithDescription(desc string)` | Set error description |
| `WithOperation(op string)` | Set operation context |
| `WithResource(resource string)` | Set resource context |
| `WithField(field string)` | Set field context |
| `WithSuggestions(suggestions ...string)` | Add remediation suggestions |
| `WithCommands(commands ...string)` | Add command suggestions |
| `WithExamples(examples ...string)` | Add usage examples |
| `WithDocLinks(links ...string)` | Add documentation links |
| `WithRetryable(retryable bool)` | Mark as retryable |
| `WithRetryAfter(duration time.Duration)` | Set retry delay |

## Testing

```bash
# Run tests
go test ./internal/errors -v

# Run specific test
go test ./internal/errors -run TestFrequencySuggestion -v

# Run with coverage
go test ./internal/errors -cover
```

## Documentation

- [Error Reference](../../docs/ERROR_REFERENCE.md) - Complete catalog of all errors
- [Integration Guide](./INTEGRATION.md) - How to integrate into provider resources
- [Troubleshooting Guide](../../docs/guides/troubleshooting.md) - Common issues and solutions

## Design Principles

1. **User-First**: Errors should help users solve problems, not just report them
2. **Actionable**: Every error should tell users what to do next
3. **Contextual**: Errors should include relevant context (resource, field, operation)
4. **Educational**: Errors should teach users about correct usage
5. **Automated**: Where possible, errors should trigger auto-remediation

## Contributing

When adding new error enhancements:

1. Create suggestion function in `suggestions.go`
2. Add tests in `suggestions_test.go`
3. Document in `ERROR_REFERENCE.md`
4. Update integration examples in `INTEGRATION.md`

## License

Copyright (c) 2026 Develeap
SPDX-License-Identifier: MPL-2.0
