# Enhanced Error Integration Guide

This guide shows how to integrate the enhanced error system into Terraform provider resources.

## Quick Start

### 1. Import the Package

```go
import (
    "github.com/develeap/terraform-provider-hyperping/internal/errors"
)
```

### 2. Enhance Client Errors in Resources

**Before:**
```go
func (r *MonitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    // ... get plan ...

    monitor, err := r.client.CreateMonitor(ctx, request)
    if err != nil {
        resp.Diagnostics.AddError("Failed to create monitor", err.Error())
        return
    }
}
```

**After:**
```go
func (r *MonitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    // ... get plan ...

    monitor, err := r.client.CreateMonitor(ctx, request)
    if err != nil {
        enhanced := errors.CreateError(err, "hyperping_monitor", plan.Name.ValueString())
        errors.AddDiagnosticError(&resp.Diagnostics, "", enhanced)
        return
    }
}
```

### 3. Field-Specific Validation Errors

```go
func (r *MonitorResource) validateFrequency(freq int64) error {
    allowedFrequencies := []int64{10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400}

    for _, allowed := range allowedFrequencies {
        if freq == allowed {
            return nil
        }
    }

    // Return enhanced error with suggestions
    return errors.FrequencySuggestion(freq)
}
```

### 4. Import Errors

```go
func (r *MonitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    id := req.ID

    monitor, err := r.client.GetMonitor(ctx, id)
    if err != nil {
        enhanced := errors.ImportError(err, "hyperping_monitor", id)
        errors.AddDiagnosticError(&resp.Diagnostics, "", enhanced)
        return
    }

    // ... set state ...
}
```

## Complete Resource Example

```go
package provider

import (
    "context"

    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/develeap/terraform-provider-hyperping/internal/client"
    "github.com/develeap/terraform-provider-hyperping/internal/errors"
)

type MonitorResource struct {
    client client.MonitorAPI
}

func (r *MonitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan MonitorResourceModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Validate before API call
    if err := r.validateMonitor(&plan); err != nil {
        errors.AddDiagnosticError(&resp.Diagnostics, "", err)
        return
    }

    // Build request
    request := client.CreateMonitorRequest{
        Name:           plan.Name.ValueString(),
        URL:            plan.URL.ValueString(),
        CheckFrequency: int(plan.CheckFrequency.ValueInt64()),
        // ... other fields ...
    }

    // Call API
    monitor, err := r.client.CreateMonitor(ctx, request)
    if err != nil {
        // Enhance error with context
        enhanced := errors.CreateError(err, "hyperping_monitor", plan.Name.ValueString())
        errors.AddDiagnosticError(&resp.Diagnostics, "", enhanced)
        return
    }

    // Map response to state
    plan.ID = types.StringValue(monitor.UUID)
    // ... other fields ...

    diags = resp.State.Set(ctx, plan)
    resp.Diagnostics.Append(diags...)
}

func (r *MonitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state MonitorResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    monitor, err := r.client.GetMonitor(ctx, state.ID.ValueString())
    if err != nil {
        enhanced := errors.ReadError(err, "hyperping_monitor", state.Name.ValueString())

        // Handle not found gracefully
        if errors.IsNotFound(err) {
            resp.State.RemoveResource(ctx)
            return
        }

        errors.AddDiagnosticError(&resp.Diagnostics, "", enhanced)
        return
    }

    // Update state
    // ...
}

func (r *MonitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var plan MonitorResourceModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Validate
    if err := r.validateMonitor(&plan); err != nil {
        errors.AddDiagnosticError(&resp.Diagnostics, "", err)
        return
    }

    // Build request
    request := client.UpdateMonitorRequest{
        Name:           plan.Name.ValueString(),
        CheckFrequency: int(plan.CheckFrequency.ValueInt64()),
        // ... other fields ...
    }

    // Call API
    monitor, err := r.client.UpdateMonitor(ctx, plan.ID.ValueString(), request)
    if err != nil {
        enhanced := errors.UpdateError(err, "hyperping_monitor", plan.Name.ValueString())
        errors.AddDiagnosticError(&resp.Diagnostics, "", enhanced)
        return
    }

    // Update state
    // ...
}

func (r *MonitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state MonitorResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    err := r.client.DeleteMonitor(ctx, state.ID.ValueString())
    if err != nil {
        // Ignore not found errors during delete
        if !errors.IsNotFound(err) {
            enhanced := errors.DeleteError(err, "hyperping_monitor", state.Name.ValueString())
            errors.AddDiagnosticError(&resp.Diagnostics, "", enhanced)
            return
        }
    }
}

func (r *MonitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    id := req.ID

    // Validate ID format
    if err := client.ValidateResourceID(id); err != nil {
        enhanced := errors.ImportError(err, "hyperping_monitor", id)
        errors.AddDiagnosticError(&resp.Diagnostics, "", enhanced)
        return
    }

    // Fetch resource
    monitor, err := r.client.GetMonitor(ctx, id)
    if err != nil {
        enhanced := errors.ImportError(err, "hyperping_monitor", id)
        errors.AddDiagnosticError(&resp.Diagnostics, "", enhanced)
        return
    }

    // Map to state
    var state MonitorResourceModel
    state.ID = types.StringValue(monitor.UUID)
    state.Name = types.StringValue(monitor.Name)
    // ... other fields ...

    diags := resp.State.Set(ctx, state)
    resp.Diagnostics.Append(diags...)
}

func (r *MonitorResource) validateMonitor(plan *MonitorResourceModel) error {
    // Validate frequency
    freq := plan.CheckFrequency.ValueInt64()
    allowedFrequencies := []int64{10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400}

    validFreq := false
    for _, allowed := range allowedFrequencies {
        if freq == allowed {
            validFreq = true
            break
        }
    }

    if !validFreq {
        return errors.FrequencySuggestion(freq)
    }

    return nil
}
```

## Validation Examples

### HTTP Method Validation

```go
func validateHTTPMethod(method string) error {
    allowedMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"}

    for _, allowed := range allowedMethods {
        if method == allowed {
            return nil
        }
    }

    return errors.HTTPMethodSuggestion(method)
}
```

### Region Validation

```go
func validateRegions(regions []string) error {
    allowedRegions := []string{
        "london", "frankfurt", "singapore", "sydney",
        "virginia", "oregon", "saopaulo", "tokyo", "bahrain",
    }

    for _, region := range regions {
        found := false
        for _, allowed := range allowedRegions {
            if region == allowed {
                found = true
                break
            }
        }
        if !found {
            return errors.RegionSuggestion(region)
        }
    }

    return nil
}
```

### Incident Status Validation

```go
func validateIncidentStatus(status string) error {
    allowedStatuses := []string{"investigating", "identified", "monitoring", "resolved"}

    for _, allowed := range allowedStatuses {
        if status == allowed {
            return nil
        }
    }

    return errors.IncidentStatusSuggestion(status)
}
```

### Incident Severity Validation

```go
func validateIncidentSeverity(severity string) error {
    allowedSeverities := []string{"minor", "major", "critical"}

    for _, allowed := range allowedSeverities {
        if severity == allowed {
            return nil
        }
    }

    return errors.SeveritySuggestion(severity)
}
```

## Custom Error Enhancement

For resource-specific errors:

```go
func (r *MonitorResource) validateCustom(plan *MonitorResourceModel) error {
    // Check if URL requires authentication but no headers provided
    url := plan.URL.ValueString()
    headers := plan.RequestHeaders.Elements()

    if requiresAuth(url) && len(headers) == 0 {
        return errors.EnhanceError(
            fmt.Errorf("URL requires authentication"),
            errors.CategoryValidation,
            errors.WithTitle("Authentication Required"),
            errors.WithDescription(fmt.Sprintf("The URL %s typically requires authentication headers", url)),
            errors.WithField("request_headers"),
            errors.WithSuggestions(
                "Add Authorization header to request_headers",
                "Or add API key as query parameter in URL",
            ),
            errors.WithExamples(
                `request_headers = [{ key = "Authorization", value = "Bearer token" }]`,
            ),
        )
    }

    return nil
}

func requiresAuth(url string) bool {
    // Your logic here
    return strings.Contains(url, "/api/") && !strings.Contains(url, "?api_key=")
}
```

## Migration Tool Integration

For migration tools (cmd/migrate-*):

```go
package main

import (
    "github.com/develeap/terraform-provider-hyperping/internal/errors"
    "github.com/develeap/terraform-provider-hyperping/internal/client"
)

func migrateMonitor(ctx context.Context, sourceMonitor SourceMonitor) error {
    // Convert to Hyperping format
    monitor := convertMonitor(sourceMonitor)

    // Create in Hyperping
    created, err := hyperpingClient.CreateMonitor(ctx, monitor)
    if err != nil {
        // Enhance error for CLI output
        enhanced := errors.EnhanceClientError(
            err,
            "create",
            fmt.Sprintf("monitor-%s", sourceMonitor.Name),
            "",
        )

        // Print enhanced error
        fmt.Fprintln(os.Stderr, enhanced.Error())
        return err
    }

    log.Printf("Migrated monitor: %s -> %s", sourceMonitor.Name, created.UUID)
    return nil
}
```

## Testing Enhanced Errors

```go
package provider_test

import (
    "testing"

    "github.com/develeap/terraform-provider-hyperping/internal/errors"
)

func TestMonitorCreate_InvalidFrequency(t *testing.T) {
    // Simulate validation error
    err := errors.FrequencySuggestion(45)

    output := err.Error()

    // Verify error contains helpful information
    if !strings.Contains(output, "Invalid Monitor Frequency") {
        t.Error("Missing error title")
    }

    if !strings.Contains(output, "30 seconds") {
        t.Error("Missing suggestion for closest valid value")
    }

    if !strings.Contains(output, "Documentation:") {
        t.Error("Missing documentation links")
    }
}
```

## Best Practices

1. **Always enhance client errors** with context (operation, resource, field)
2. **Use field-specific validators** for better error messages
3. **Provide examples** in validation errors
4. **Link to documentation** for complex configurations
5. **Handle not found gracefully** in Read operations
6. **Test error messages** to ensure they're helpful

## Error Message Checklist

- [ ] Includes clear title (what went wrong)
- [ ] Provides description (why it happened)
- [ ] Shows context (resource, operation, field)
- [ ] Offers suggestions (how to fix)
- [ ] Provides commands (what to try)
- [ ] Includes examples (correct usage)
- [ ] Links to documentation
- [ ] Handles auto-retry for transient failures
