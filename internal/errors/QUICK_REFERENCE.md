# Enhanced Error System - Quick Reference

One-page cheat sheet for using the enhanced error system.

## Import

```go
import "github.com/develeap/terraform-provider-hyperping/internal/errors"
```

## Quick Start

### Resource Operations

```go
// Create
monitor, err := r.client.CreateMonitor(ctx, req)
if err != nil {
    enhanced := errors.CreateError(err, "hyperping_monitor", plan.Name.ValueString())
    errors.AddDiagnosticError(&resp.Diagnostics, "", enhanced)
    return
}

// Read
monitor, err := r.client.GetMonitor(ctx, id)
if err != nil {
    enhanced := errors.ReadError(err, "hyperping_monitor", state.Name.ValueString())
    if errors.IsNotFound(err) {
        resp.State.RemoveResource(ctx)
        return
    }
    errors.AddDiagnosticError(&resp.Diagnostics, "", enhanced)
    return
}

// Update
monitor, err := r.client.UpdateMonitor(ctx, id, req)
if err != nil {
    enhanced := errors.UpdateError(err, "hyperping_monitor", plan.Name.ValueString())
    errors.AddDiagnosticError(&resp.Diagnostics, "", enhanced)
    return
}

// Delete
err := r.client.DeleteMonitor(ctx, id)
if err != nil && !errors.IsNotFound(err) {
    enhanced := errors.DeleteError(err, "hyperping_monitor", state.Name.ValueString())
    errors.AddDiagnosticError(&resp.Diagnostics, "", enhanced)
    return
}

// Import
monitor, err := r.client.GetMonitor(ctx, importID)
if err != nil {
    enhanced := errors.ImportError(err, "hyperping_monitor", importID)
    errors.AddDiagnosticError(&resp.Diagnostics, "", enhanced)
    return
}
```

## Field Validation

```go
// Frequency (10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400)
if !isValidFrequency(freq) {
    return errors.FrequencySuggestion(freq)
}

// Region (london, frankfurt, singapore, sydney, virginia, oregon, saopaulo, tokyo, bahrain)
if !isValidRegion(region) {
    return errors.RegionSuggestion(region)
}

// HTTP Method (GET, POST, PUT, PATCH, DELETE, HEAD)
if !isValidMethod(method) {
    return errors.HTTPMethodSuggestion(method)
}

// Status Code ("200", "2xx", "200,201")
if !isValidStatusCode(code) {
    return errors.StatusCodeSuggestion(code)
}

// Incident Status (investigating, identified, monitoring, resolved)
if !isValidStatus(status) {
    return errors.IncidentStatusSuggestion(status)
}

// Severity (minor, major, critical)
if !isValidSeverity(severity) {
    return errors.SeveritySuggestion(severity)
}
```

## Custom Errors

```go
// Full control
err := errors.EnhanceError(
    originalErr,
    errors.CategoryValidation,
    errors.WithTitle("Custom Title"),
    errors.WithDescription("What happened"),
    errors.WithOperation("create"),
    errors.WithResource("hyperping_monitor.test"),
    errors.WithField("url"),
    errors.WithSuggestions("Fix this", "Try that"),
    errors.WithCommands("terraform validate"),
    errors.WithExamples("url = \"https://example.com\""),
    errors.WithDocLinks("https://docs.example.com"),
)
```

## Error Categories

| Category | Use When | Auto-Retry |
|----------|----------|------------|
| `CategoryAuth` | 401, 403 | No |
| `CategoryRateLimit` | 429 | Yes |
| `CategoryValidation` | 400, 422 | No |
| `CategoryNotFound` | 404 | No |
| `CategoryServer` | 500, 502, 503, 504 | Yes |
| `CategoryNetwork` | Connection errors | No |
| `CategoryCircuit` | Circuit breaker | No |

## Helper Functions

```go
// Check error type
if errors.IsNotFound(err) { /* handle */ }
if errors.IsUnauthorized(err) { /* handle */ }
if errors.IsRateLimited(err) { /* handle */ }
if errors.IsValidation(err) { /* handle */ }
if errors.IsServerError(err) { /* handle */ }

// Format validation errors
formatted := errors.FormatValidationErrors(apiErr.Details)
```

## Enhancement Options

```go
errors.WithTitle("Title")                      // Set title
errors.WithDescription("Description")          // Set description
errors.WithOperation("create")                 // Set operation
errors.WithResource("hyperping_monitor.test") // Set resource
errors.WithField("frequency")                  // Set field
errors.WithSuggestions("Suggestion 1")         // Add suggestions
errors.WithCommands("terraform plan")          // Add commands
errors.WithExamples("frequency = 60")          // Add examples
errors.WithDocLinks("https://...")             // Add doc links
errors.WithRetryable(true)                     // Mark retryable
errors.WithRetryAfter(30 * time.Second)        // Set retry delay
```

## Common Patterns

### Validation Before API Call

```go
func (r *MonitorResource) Create(...) {
    // Validate first
    if err := r.validateMonitor(&plan); err != nil {
        errors.AddDiagnosticError(&resp.Diagnostics, "", err)
        return
    }

    // Then call API
    monitor, err := r.client.CreateMonitor(ctx, req)
    // ...
}

func (r *MonitorResource) validateMonitor(plan *MonitorResourceModel) error {
    freq := plan.CheckFrequency.ValueInt64()
    if !isValidFrequency(freq) {
        return errors.FrequencySuggestion(freq)
    }
    return nil
}
```

### Graceful Not Found Handling

```go
// Read operation
if err != nil {
    if errors.IsNotFound(err) {
        resp.State.RemoveResource(ctx)  // Remove from state
        return
    }
    enhanced := errors.ReadError(err, "hyperping_monitor", name)
    errors.AddDiagnosticError(&resp.Diagnostics, "", enhanced)
    return
}

// Delete operation
if err != nil && !errors.IsNotFound(err) {
    enhanced := errors.DeleteError(err, "hyperping_monitor", name)
    errors.AddDiagnosticError(&resp.Diagnostics, "", enhanced)
    return
}
```

### Field-Specific Errors

```go
// When you know which field failed
enhanced := errors.FieldError(
    err,
    "create",
    "hyperping_monitor",
    "prod_api",
    "check_frequency",  // Specific field
)
```

## Testing

```go
func TestMonitorCreate_InvalidFrequency(t *testing.T) {
    err := errors.FrequencySuggestion(45)
    output := err.Error()

    // Verify error message
    if !strings.Contains(output, "Invalid Monitor Frequency") {
        t.Error("Missing title")
    }
    if !strings.Contains(output, "30 seconds") {
        t.Error("Missing suggestion")
    }
}
```

## Migration Tools

```go
// In CLI tools
if err != nil {
    enhanced := errors.EnhanceClientError(
        err,
        "migrate",
        fmt.Sprintf("monitor-%s", name),
        "",
    )
    fmt.Fprintln(os.Stderr, enhanced.Error())
    return err
}
```

## Best Practices

1. ✅ **Always enhance client errors** with context
2. ✅ **Validate before API calls** to provide better errors
3. ✅ **Use field-specific validators** for common fields
4. ✅ **Handle not-found gracefully** in Read/Delete
5. ✅ **Test error messages** to ensure they're helpful
6. ✅ **Use consistent resource naming** (e.g., "hyperping_monitor.name")

## Common Gotchas

❌ **Don't**: Nest enhanced errors
```go
err := errors.CreateError(err, "monitor", "test")
err = errors.EnhanceError(err, ...)  // Don't enhance twice!
```

✅ **Do**: Enhance once at the point of use
```go
err := r.client.CreateMonitor(...)
if err != nil {
    enhanced := errors.CreateError(err, "monitor", "test")
    errors.AddDiagnosticError(&resp.Diagnostics, "", enhanced)
    return
}
```

❌ **Don't**: Ignore IsNotFound in Read
```go
if err != nil {
    return err  // Wrong! Resource might be deleted
}
```

✅ **Do**: Remove from state on not found
```go
if err != nil {
    if errors.IsNotFound(err) {
        resp.State.RemoveResource(ctx)
        return
    }
    // Handle other errors
}
```

## Documentation

- [Complete Error Reference](../../docs/ERROR_REFERENCE.md)
- [Integration Guide](./INTEGRATION.md)
- [Package Documentation](./README.md)
- [Real-World Examples](./EXAMPLES.md)

## Support

- GitHub Issues: https://github.com/develeap/terraform-provider-hyperping/issues
- Documentation: https://registry.terraform.io/providers/develeap/hyperping/latest/docs

---

**Quick Tip**: When in doubt, use the resource operation helpers (CreateError, ReadError, etc.). They provide the right context automatically.
