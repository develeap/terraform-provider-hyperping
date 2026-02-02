# Error Handling Guide

This guide documents the standard error handling patterns for the terraform-provider-hyperping codebase.

## Overview

All error handling follows consistent patterns using standardized functions from `internal/provider/error_helpers.go`. This ensures:

- **Consistent user experience** - All error messages follow the same format
- **Helpful troubleshooting** - Errors include actionable guidance
- **Easier maintenance** - Centralized error logic
- **Better testing** - Standard patterns are easier to test

## Standard Error Functions

### CRUD Operation Errors

#### `newCreateError(resourceType string, err error) diag.Diagnostic`

Use when a resource creation API call fails.

```go
// Before
resp.Diagnostics.AddError(
    "Error creating monitor",
    fmt.Sprintf("Could not create monitor: %s", err),
)

// After
resp.Diagnostics.Append(newCreateError("Monitor", err))
```

**Includes troubleshooting for:**
- API key permissions
- Required fields
- Dashboard link

#### `newReadError(resourceType, resourceID string, err error) diag.Diagnostic`

Use when reading a resource from the API fails.

```go
// Before
resp.Diagnostics.AddError(
    "Error reading monitor",
    fmt.Sprintf("Could not read monitor (ID: %s): %s", state.ID.ValueString(), err),
)

// After
resp.Diagnostics.Append(newReadError("Monitor", state.ID.ValueString(), err))
```

**Includes troubleshooting for:**
- Resource existence
- API key permissions
- Network connectivity
- Dashboard link

#### `newUpdateError(resourceType, resourceID string, err error) diag.Diagnostic`

Use when updating a resource fails.

```go
// Before
resp.Diagnostics.AddError(
    "Error updating monitor",
    fmt.Sprintf("Could not update monitor (ID: %s): %s", state.ID.ValueString(), err),
)

// After
resp.Diagnostics.Append(newUpdateError("Monitor", state.ID.ValueString(), err))
```

**Includes troubleshooting for:**
- Resource existence
- API key permissions
- Value validation
- Dashboard link

#### `newDeleteError(resourceType, resourceID string, err error) diag.Diagnostic`

Use when deleting a resource fails.

```go
// Before
resp.Diagnostics.AddError(
    "Error deleting monitor",
    fmt.Sprintf("Could not delete monitor (ID: %s): %s", state.ID.ValueString(), err),
)

// After
resp.Diagnostics.Append(newDeleteError("Monitor", state.ID.ValueString(), err))
```

**Includes troubleshooting for:**
- Resource existence
- API key permissions
- Dependencies
- Dashboard link

#### `newListError(resourceType string, err error) diag.Diagnostic`

Use when listing resources fails (data sources).

```go
// Before
resp.Diagnostics.AddError(
    "Error listing monitors",
    fmt.Sprintf("Could not list monitors: %s", err),
)

// After
resp.Diagnostics.Append(newListError("Monitors", err))
```

**Includes troubleshooting for:**
- API key permissions
- Network connectivity
- Service status

### Secondary Operation Errors

#### `newReadAfterCreateError(resourceType, resourceID string, err error) diag.Diagnostic`

Use when a resource is created successfully but reading it back fails.

```go
// Before
resp.Diagnostics.AddError(
    "Error reading created monitor",
    fmt.Sprintf("Monitor created (ID: %s) but failed to read: %s", createResp.UUID, err),
)

// After
resp.Diagnostics.Append(newReadAfterCreateError("Monitor", createResp.UUID, err))
```

**Special features:**
- Explains the resource exists but isn't in state
- Provides terraform import command example

#### `newBuildRequestError(operation string, err error) diag.Diagnostic`

Use when building an API request fails due to invalid configuration.

```go
resp.Diagnostics.Append(newBuildRequestError("create", err))
```

#### `newMapResponseError(resourceType string, err error) diag.Diagnostic`

Use when mapping API response to Terraform state fails.

```go
resp.Diagnostics.Append(newMapResponseError("Monitor", err))
```

**Note:** May indicate API schema changes

### Configuration and Validation Errors

#### `newConfigError(message string) diag.Diagnostic`

Use for general configuration errors.

```go
// Before
resp.Diagnostics.AddError("Invalid Configuration", err.Error())

// After
resp.Diagnostics.Append(newConfigError(err.Error()))
```

#### `newValidationError(field, message string) diag.Diagnostic`

Use for input validation failures.

```go
// Before
resp.Diagnostics.AddError("Validation Error", err.Error())

// After
resp.Diagnostics.Append(newValidationError("email", err.Error()))
```

#### `newImportError(resourceType string, err error) diag.Diagnostic`

Use when importing a resource fails.

```go
// Before
resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("Cannot import monitor: %s", err))

// After
resp.Diagnostics.Append(newImportError("Monitor", err))
```

**Includes:**
- Import format guidance
- Link to documentation

### Provider Configuration Errors

#### `newProviderConfigError(message string) diag.Diagnostic`

Use for provider configuration issues.

```go
resp.Diagnostics.Append(newProviderConfigError("API key is required"))
```

**Includes troubleshooting for:**
- Environment variables
- Provider block configuration
- API key format

#### `newClientConfigError(err error) diag.Diagnostic`

Use when configuring the API client fails.

```go
resp.Diagnostics.Append(newClientConfigError(err))
```

#### `newUnexpectedConfigTypeError(expected string, actual interface{}) diag.Diagnostic`

Use for type assertion failures in Configure() methods.

```go
// Before
resp.Diagnostics.AddError(
    "Unexpected Resource Configure Type",
    fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
)

// After
resp.Diagnostics.Append(newUnexpectedConfigTypeError("*client.Client", req.ProviderData))
```

### Warning Helpers

#### `newDeleteWarning(resourceType, message string) diag.Diagnostic`

Use when a resource is not found during deletion (likely already deleted).

```go
// Before
resp.Diagnostics.AddWarning(
    "Monitor Not Found",
    "Resource not found in Hyperping. It may have been deleted outside Terraform.",
)

// After
resp.Diagnostics.Append(newDeleteWarning("Monitor", "Resource not found in Hyperping"))
```

#### `newUpdateWarning(resourceType, message string) diag.Diagnostic`

Use for non-critical update warnings.

```go
resp.Diagnostics.Append(newUpdateWarning("Monitor", "Some fields were not updated"))
```

## Usage Examples by Operation

### Create Operation

```go
func (r *MonitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan MonitorResourceModel

    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Build request
    createReq := r.buildCreateRequest(ctx, &plan, &resp.Diagnostics)
    if resp.Diagnostics.HasError() {
        return
    }

    // Call API
    createResp, err := r.client.CreateMonitor(ctx, createReq)
    if err != nil {
        resp.Diagnostics.Append(newCreateError("Monitor", err))
        return
    }

    // Read back
    monitor, err := r.client.GetMonitor(ctx, createResp.UUID)
    if err != nil {
        resp.Diagnostics.Append(newReadAfterCreateError("Monitor", createResp.UUID, err))
        return
    }

    // Map response
    r.mapMonitorToModel(monitor, &plan, &resp.Diagnostics)
    if resp.Diagnostics.HasError() {
        return
    }

    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
```

### Read Operation

```go
func (r *MonitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state MonitorResourceModel

    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    monitor, err := r.client.GetMonitor(ctx, state.ID.ValueString())
    if err != nil {
        if client.IsNotFound(err) {
            resp.State.RemoveResource(ctx)
            return
        }
        resp.Diagnostics.Append(newReadError("Monitor", state.ID.ValueString(), err))
        return
    }

    r.mapMonitorToModel(monitor, &state, &resp.Diagnostics)
    if resp.Diagnostics.HasError() {
        return
    }

    resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
```

### Update Operation

```go
func (r *MonitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var plan, state MonitorResourceModel

    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    updateReq := r.buildUpdateRequest(ctx, &plan, &resp.Diagnostics)
    if resp.Diagnostics.HasError() {
        return
    }

    _, err := r.client.UpdateMonitor(ctx, state.ID.ValueString(), updateReq)
    if err != nil {
        resp.Diagnostics.Append(newUpdateError("Monitor", state.ID.ValueString(), err))
        return
    }

    // Read back to get updated state
    monitor, err := r.client.GetMonitor(ctx, state.ID.ValueString())
    if err != nil {
        resp.Diagnostics.Append(newReadError("Monitor", state.ID.ValueString(), err))
        return
    }

    r.mapMonitorToModel(monitor, &plan, &resp.Diagnostics)
    if resp.Diagnostics.HasError() {
        return
    }

    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
```

### Delete Operation

```go
func (r *MonitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state MonitorResourceModel

    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    err := r.client.DeleteMonitor(ctx, state.ID.ValueString())
    if err != nil {
        if client.IsNotFound(err) {
            resp.Diagnostics.Append(newDeleteWarning("Monitor", "Resource not found in Hyperping"))
            return
        }
        resp.Diagnostics.Append(newDeleteError("Monitor", state.ID.ValueString(), err))
        return
    }
}
```

### ImportState Operation

```go
func (r *MonitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    // Validate the import ID
    if err := client.ValidateResourceID(req.ID); err != nil {
        resp.Diagnostics.Append(newImportError("Monitor", err))
        return
    }
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

### Configure Operation

```go
func (r *MonitorResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }

    c, ok := req.ProviderData.(*client.Client)
    if !ok {
        resp.Diagnostics.Append(newUnexpectedConfigTypeError("*client.Client", req.ProviderData))
        return
    }

    r.client = c
}
```

## Migration Checklist

When migrating a file to use standard error patterns:

- [ ] Replace all `AddError` calls with appropriate `newXXXError` functions
- [ ] Replace all `AddWarning` calls with appropriate `newXXXWarning` functions
- [ ] Remove manual `fmt.Sprintf` for error messages
- [ ] Ensure resource type names are capitalized consistently
- [ ] Test that the file still compiles: `go build ./internal/provider`
- [ ] Run unit tests: `go test ./internal/provider/...`
- [ ] Verify error messages are helpful and actionable

## Resource Type Naming

Use these standard resource type names:

| Resource | Type Name |
|----------|-----------|
| Monitor | `"Monitor"` |
| Healthcheck | `"Healthcheck"` |
| Incident | `"Incident"` |
| Incident Update | `"Incident Update"` |
| Outage | `"Outage"` |
| Maintenance Window | `"Maintenance Window"` |
| Status Page | `"Status Page"` |
| Status Page Subscriber | `"Subscriber"` |

## Benefits

### For Users

- Consistent error format across all resources
- Clear, actionable troubleshooting steps
- Links to dashboard and documentation
- Better error context (resource IDs, operations)

### For Developers

- Single source of truth for error messages
- Less code duplication
- Easier to update error messages globally
- Standard patterns reduce cognitive load
- Better test coverage

## Testing Error Handling

Example test for error helpers:

```go
func TestNewCreateError(t *testing.T) {
    err := errors.New("invalid request")
    diag := newCreateError("Monitor", err)

    if !strings.Contains(diag.Summary(), "Failed to Create Monitor") {
        t.Errorf("Expected summary to contain 'Failed to Create Monitor', got: %s", diag.Summary())
    }

    if !strings.Contains(diag.Detail(), "invalid request") {
        t.Errorf("Expected detail to contain error message, got: %s", diag.Detail())
    }

    if !strings.Contains(diag.Detail(), "Troubleshooting") {
        t.Errorf("Expected detail to contain troubleshooting section, got: %s", diag.Detail())
    }
}
```

## Common Patterns

### Check for Not Found Errors

```go
if err != nil {
    if client.IsNotFound(err) {
        // Handle not found case (usually remove resource or return nil)
        resp.State.RemoveResource(ctx)
        return
    }
    resp.Diagnostics.Append(newReadError("Monitor", id, err))
    return
}
```

### Handle Diagnostics from Mapping Functions

```go
r.mapMonitorToModel(monitor, &plan, &resp.Diagnostics)
if resp.Diagnostics.HasError() {
    return
}
```

### Wrap Errors for Context

When you need to add context to an error before passing to a helper:

```go
if err != nil {
    resp.Diagnostics.Append(newUpdateError("Monitor", id, fmt.Errorf("failed to pause: %w", err)))
    return
}
```

## Future Enhancements

Potential improvements to consider:

1. **Error codes** - Add structured error codes for programmatic handling
2. **Retry hints** - Indicate which errors are retryable
3. **Severity levels** - Distinguish between fatal and recoverable errors
4. **Telemetry** - Add optional error tracking/reporting
5. **Localization** - Support multiple languages for error messages

## Related Files

- `internal/provider/error_helpers.go` - Error function definitions
- `internal/provider/error_helpers_test.go` - Error function tests
- `internal/client/errors.go` - API client error types

## Questions?

If you have questions about error handling patterns or need help migrating a file, please:

1. Review existing migrated files (e.g., `monitor_resource.go`)
2. Check the test files for examples
3. Consult this guide for standard patterns
