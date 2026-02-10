# Contract Testing Framework

This directory contains a comprehensive contract testing framework for validating Hyperping API responses. Contract tests ensure that the API structure matches our expectations and protect against breaking changes.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Framework Components](#framework-components)
- [Usage Examples](#usage-examples)
- [Adding New Validators](#adding-new-validators)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Overview

### What are Contract Tests?

Contract tests verify that API responses match an expected structure. Unlike unit tests that test internal logic, contract tests validate:

- Field names and types
- Required vs optional fields
- Enum value constraints
- Data format (UUIDs, timestamps, URLs)
- Nested object structure
- Array element validation

### Why Use Contract Tests?

1. **API Change Detection**: Immediately detect when the API structure changes
2. **Documentation**: Tests serve as executable documentation of API structure
3. **Confidence**: Refactor client code knowing the API contract is verified
4. **Fast Feedback**: VCR cassettes enable fast test execution without API calls

## Quick Start

### Basic Contract Test

```go
package client

import (
    "testing"
    "github.com/stretchr/testify/require"
)

func TestContract_ListMonitors(t *testing.T) {
    RunContractTest(t, ContractTestConfig{
        CassetteName: "monitor_list",
        ResourceType: "Monitor",
    }, func(t *testing.T, client *Client, ctx context.Context) {
        // Call API
        monitors, err := client.ListMonitors(ctx)
        require.NoError(t, err)

        // Validate contract
        validator := NewContractValidator(t, "Monitor")
        validator.ValidateMonitorList(monitors)
    })
}
```

### CRUD Contract Test

```go
func TestContract_MonitorCRUD(t *testing.T) {
    RunContractTest(t, ContractTestConfig{
        CassetteName: "monitor_crud",
        ResourceType: "Monitor",
    }, func(t *testing.T, client *Client, ctx context.Context) {
        validator := NewContractValidator(t, "Monitor")

        // Create
        createReq := BuildTestMonitor("Contract Test Monitor")
        monitor, err := client.CreateMonitor(ctx, createReq)
        require.NoError(t, err)
        validator.ValidateMonitor(monitor)
        defer client.DeleteMonitor(ctx, monitor.UUID)

        // Read
        retrieved, err := client.GetMonitor(ctx, monitor.UUID)
        require.NoError(t, err)
        validator.ValidateMonitor(retrieved)

        // Update
        updateReq := UpdateMonitorRequest{Name: stringPtr("Updated Name")}
        updated, err := client.UpdateMonitor(ctx, monitor.UUID, updateReq)
        require.NoError(t, err)
        validator.ValidateMonitor(updated)

        // List
        monitors, err := client.ListMonitors(ctx)
        require.NoError(t, err)
        validator.ValidateMonitorList(monitors)
    })
}
```

## Framework Components

### 1. Core Validators (`contract_validators.go`)

#### Basic Field Validators

| Validator | Purpose | Example |
|-----------|---------|---------|
| `ValidateUUID` | Validates UUID format (prefix_id) | `mon_abc123` |
| `ValidateTimestamp` | Validates ISO 8601 timestamps | `2026-01-15T10:30:00Z` |
| `ValidateEnum` | Validates enum values | `protocol: http/https/tcp` |
| `ValidateURL` | Validates URL format | `https://example.com` |
| `ValidateStringField` | Validates string requirements | Required/optional checks |
| `ValidateStringLength` | Validates max length | Name <= 255 chars |
| `ValidateLocalizedText` | Validates multi-language text | EN/FR/DE/ES |
| `ValidateHexColor` | Validates hex color codes | `#ff5733` |

#### Resource Validators

The `ContractValidator` provides fluent validation for all resource types:

```go
validator := NewContractValidator(t, "Monitor")
validator.ValidateMonitor(monitor)
validator.ValidateMonitorList(monitors)
```

Supported resources:
- Monitors
- Incidents
- Maintenance Windows
- Status Pages
- Subscribers
- Healthchecks
- Outages
- Reports

### 2. Test Helpers (`contract_test_helpers.go`)

#### Test Runner

`RunContractTest` handles VCR setup, client creation, and cleanup:

```go
RunContractTest(t, ContractTestConfig{
    CassetteName: "test_name",
    ResourceType: "Monitor",
}, func(t *testing.T, client *Client, ctx context.Context) {
    // Your test logic here
})
```

#### Cleanup Contexts

Structured cleanup for created resources:

```go
testCtx := MonitorTestContext{
    Monitor: monitor,
    Client:  client,
    Context: ctx,
}
defer CleanupMonitor(t, testCtx)
```

#### Test Data Builders

Pre-configured test data with sensible defaults:

```go
monitor := BuildTestMonitor("My Monitor")
incident := BuildTestIncident("Outage", []string{statusPageUUID})
maintenance := BuildTestMaintenance("DB Maintenance", []string{monitorUUID})
statusPage := BuildTestStatusPage("My Status Page", "subdomain")
healthcheck := BuildTestHealthcheck("Cron Monitor")
```

#### Error Validators

Validate expected error conditions:

```go
ExpectValidationError(t, err, "fieldName")
ExpectAPIError(t, err, 404)
ExpectNotFoundError(t, err)
ExpectUnauthorizedError(t, err)
```

## Usage Examples

### Example 1: Monitor List Contract Test

```go
func TestContract_ListMonitors_Structure(t *testing.T) {
    RunContractTest(t, ContractTestConfig{
        CassetteName: "monitor_list_structure",
        ResourceType: "Monitor",
    }, func(t *testing.T, client *Client, ctx context.Context) {
        monitors, err := client.ListMonitors(ctx)
        require.NoError(t, err)
        require.NotEmpty(t, monitors, "Expected at least one monitor")

        validator := NewContractValidator(t, "Monitor")
        validator.ValidateMonitorList(monitors)

        // Additional assertions
        for _, monitor := range monitors {
            assert.NotEmpty(t, monitor.UUID)
            assert.NotEmpty(t, monitor.Name)
            assert.NotEmpty(t, monitor.URL)
        }
    })
}
```

### Example 2: Incident Update Validation

```go
func TestContract_IncidentUpdate_Structure(t *testing.T) {
    RunContractTest(t, ContractTestConfig{
        CassetteName: "incident_update",
        ResourceType: "Incident",
    }, func(t *testing.T, client *Client, ctx context.Context) {
        // Get status page
        pages, _ := client.ListStatusPages(ctx, nil, nil)
        require.NotEmpty(t, pages.StatusPages)

        // Create incident
        createReq := BuildTestIncident("Test Incident", []string{pages.StatusPages[0].UUID})
        incident, err := client.CreateIncident(ctx, createReq)
        require.NoError(t, err)
        defer client.DeleteIncident(ctx, incident.UUID)

        // Add update
        updateReq := AddIncidentUpdateRequest{
            Text: LocalizedText{En: "Update text"},
            Type: "identified",
            Date: time.Now().UTC().Format(time.RFC3339),
        }

        updated, err := client.AddIncidentUpdate(ctx, incident.UUID, updateReq)
        require.NoError(t, err)

        // Validate structure
        validator := NewContractValidator(t, "Incident")
        validator.ValidateIncident(updated)

        require.NotEmpty(t, updated.Updates, "Expected at least one update")
        assert.Equal(t, "identified", updated.Updates[0].Type)
    })
}
```

### Example 3: Pagination Validation

```go
func TestContract_StatusPagePagination(t *testing.T) {
    RunContractTest(t, ContractTestConfig{
        CassetteName: "statuspage_pagination",
        ResourceType: "StatusPage",
    }, func(t *testing.T, client *Client, ctx context.Context) {
        page := 0
        response, err := client.ListStatusPages(ctx, &page, nil)
        require.NoError(t, err)

        validator := NewContractValidator(t, "StatusPage")
        validator.ValidateStatusPageList(response)

        // Validate pagination metadata
        assert.Equal(t, 0, response.Page)
        assert.GreaterOrEqual(t, response.Total, 0)
        assert.Greater(t, response.ResultsPerPage, 0)
    })
}
```

### Example 4: Error Contract Test

```go
func TestContract_Monitor_NotFound(t *testing.T) {
    RunContractTest(t, ContractTestConfig{
        CassetteName: "monitor_not_found",
        ResourceType: "Monitor",
    }, func(t *testing.T, client *Client, ctx context.Context) {
        _, err := client.GetMonitor(ctx, "mon_nonexistent123456")
        ExpectNotFoundError(t, err)
    })
}

func TestContract_Monitor_Unauthorized(t *testing.T) {
    RunContractTest(t, ContractTestConfig{
        CassetteName: "monitor_unauthorized",
        ResourceType: "Monitor",
    }, func(t *testing.T, client *Client, ctx context.Context) {
        badClient := NewClient("invalid_key")
        _, err := badClient.ListMonitors(ctx)
        ExpectUnauthorizedError(t, err)
    })
}
```

### Example 5: Validation Error Test

```go
func TestContract_Monitor_ValidationError(t *testing.T) {
    RunContractTest(t, ContractTestConfig{
        CassetteName: "monitor_validation",
        ResourceType: "Monitor",
    }, func(t *testing.T, client *Client, ctx context.Context) {
        // Empty UUID should cause validation error
        _, err := client.GetMonitor(ctx, "")
        ExpectValidationError(t, err, "UUID")

        // Invalid frequency should cause validation error
        createReq := CreateMonitorRequest{
            Name:           "Test",
            URL:            "https://example.com",
            Protocol:       "http",
            CheckFrequency: 999, // Invalid
        }
        _, err = client.CreateMonitor(ctx, createReq)
        require.Error(t, err)
    })
}
```

## Adding New Validators

### Step 1: Add Basic Validator

Add field-level validators to `contract_validators.go`:

```go
// ValidateCustomField validates a custom field format.
func ValidateCustomField(t *testing.T, fieldName, value string) {
    t.Helper()
    require.NotEmpty(t, value, "%s should not be empty", fieldName)

    // Add your validation logic
    pattern := regexp.MustCompile(`^custom_[a-z]+$`)
    assert.Regexp(t, pattern, value,
        "%s should match custom format", fieldName)
}
```

### Step 2: Add Resource Validator

Add resource-level validator to `ContractValidator`:

```go
// ValidateCustomResource validates a CustomResource response structure.
func (cv *ContractValidator) ValidateCustomResource(resource *CustomResource) {
    cv.t.Helper()
    require.NotNil(cv.t, resource, "CustomResource should not be nil")

    // Validate required fields
    ValidateUUID(cv.t, "UUID", resource.UUID)
    ValidateStringField(cv.t, "Name", resource.Name, true)
    ValidateCustomField(cv.t, "CustomField", resource.CustomField)

    // Validate optional fields
    if resource.OptionalField != nil {
        ValidateStringField(cv.t, "OptionalField", *resource.OptionalField, false)
    }
}

// ValidateCustomResourceList validates a list of CustomResource responses.
func (cv *ContractValidator) ValidateCustomResourceList(resources []CustomResource) {
    cv.t.Helper()
    require.NotNil(cv.t, resources, "CustomResource list should not be nil")

    for i, resource := range resources {
        cv.t.Run(fmt.Sprintf("CustomResource[%d]", i), func(t *testing.T) {
            validator := NewContractValidator(t, cv.resourceName)
            validator.ValidateCustomResource(&resource)
        })
    }
}
```

### Step 3: Add Test Helpers

Add cleanup and builder helpers to `contract_test_helpers.go`:

```go
// CustomResourceTestContext holds resources for cleanup.
type CustomResourceTestContext struct {
    CustomResource *CustomResource
    Client         *Client
    Context        context.Context
}

// CleanupCustomResource is a helper to clean up resources.
func CleanupCustomResource(t *testing.T, ctx CustomResourceTestContext) {
    t.Helper()
    if ctx.CustomResource != nil && ctx.Client != nil {
        if err := ctx.Client.DeleteCustomResource(ctx.Context, ctx.CustomResource.UUID); err != nil {
            t.Logf("Warning: failed to cleanup resource %s: %v",
                ctx.CustomResource.UUID, err)
        }
    }
}

// BuildTestCustomResource creates a resource for testing.
func BuildTestCustomResource(name string) CreateCustomResourceRequest {
    return CreateCustomResourceRequest{
        Name:        name,
        CustomField: "custom_value",
    }
}
```

### Step 4: Write Tests

Create contract tests using the new validators:

```go
func TestContract_CustomResource_CRUD(t *testing.T) {
    RunContractTest(t, ContractTestConfig{
        CassetteName: "custom_resource_crud",
        ResourceType: "CustomResource",
    }, func(t *testing.T, client *Client, ctx context.Context) {
        validator := NewContractValidator(t, "CustomResource")

        // Create
        createReq := BuildTestCustomResource("Test Resource")
        resource, err := client.CreateCustomResource(ctx, createReq)
        require.NoError(t, err)
        validator.ValidateCustomResource(resource)

        testCtx := CustomResourceTestContext{
            CustomResource: resource,
            Client:         client,
            Context:        ctx,
        }
        defer CleanupCustomResource(t, testCtx)

        // Read
        retrieved, err := client.GetCustomResource(ctx, resource.UUID)
        require.NoError(t, err)
        validator.ValidateCustomResource(retrieved)

        // List
        resources, err := client.ListCustomResources(ctx)
        require.NoError(t, err)
        validator.ValidateCustomResourceList(resources)
    })
}
```

## Best Practices

### 1. Test Organization

```
internal/client/
├── contract_validators.go          # Reusable validators
├── contract_test_helpers.go        # Test helpers
├── CONTRACT_TESTING.md            # This file
├── monitors_contract_test.go      # Monitor contract tests
├── incidents_contract_test.go     # Incident contract tests
└── statuspages_contract_test.go   # Status page contract tests
```

### 2. Naming Conventions

- Test files: `{resource}_contract_test.go`
- Test functions: `TestContract_{Resource}_{Operation}`
- Cassettes: `{resource}_{operation}` (e.g., `monitor_crud`, `incident_list`)

### 3. Cassette Management

**Recording New Cassettes:**

```bash
# Record new API interactions
RECORD_MODE=true HYPERPING_API_KEY=your_key go test -v -run TestContract_Monitor_CRUD
```

**Replaying Cassettes:**

```bash
# Replay from cassettes (no API calls)
go test -v -run TestContract_Monitor_CRUD
```

**Updating Cassettes:**

```bash
# Delete old cassette and re-record
rm internal/client/testdata/cassettes/monitor_crud.yaml
RECORD_MODE=true HYPERPING_API_KEY=your_key go test -v -run TestContract_Monitor_CRUD
```

### 4. Validation Coverage

Ensure you validate:

✅ **Required fields** - Must be present and non-empty
✅ **Optional fields** - Validate when present
✅ **Field formats** - UUIDs, timestamps, URLs, enums
✅ **Nested objects** - Validate recursively
✅ **Arrays** - Validate each element
✅ **Error responses** - Validate error structure

### 5. Test Independence

Each test should:

- Create its own test data
- Clean up after itself
- Not depend on other tests
- Work in any order

### 6. Clear Error Messages

Good validators provide context:

```go
// Bad
assert.NotEmpty(t, value)

// Good
require.NotEmpty(t, value, "Monitor UUID should not be empty")
```

## Troubleshooting

### Problem: Cassette Not Found

**Error:**
```
Skipping: no cassette exists at testdata/cassettes/monitor_list.yaml
```

**Solution:**
```bash
RECORD_MODE=true HYPERPING_API_KEY=your_key go test -v -run TestContract_Monitor_List
```

### Problem: API Key Masked in Cassette

**Issue:** API key is visible in committed cassette

**Solution:** The framework automatically masks sensitive headers. Ensure you're using the latest VCR recorder.

### Problem: Validation Failing on Valid Data

**Error:**
```
UUID should match UUID format 'prefix_id', got: mon_abc123xyz
```

**Solution:** Check your regex pattern. Hyperping UUIDs can contain uppercase and lowercase alphanumeric characters after the prefix.

### Problem: Test Passes Locally but Fails in CI

**Issue:** Different API responses between environments

**Solution:**
1. Re-record cassettes with consistent test data
2. Use deterministic test data (avoid random values)
3. Check for timezone differences in timestamps

### Problem: Slow Test Execution

**Issue:** Tests are hitting real API instead of cassettes

**Solution:**
1. Ensure cassettes exist in `testdata/cassettes/`
2. Don't set `RECORD_MODE=true` for regular test runs
3. Check VCR recorder configuration

## Additional Resources

- [VCR Documentation](https://github.com/dnaeon/go-vcr)
- [Testify Assertions](https://github.com/stretchr/testify)
- [Hyperping API Documentation](https://hyperping.io/docs/api)

## Contributing

When adding new contract tests:

1. Follow naming conventions
2. Add comprehensive validators
3. Include error case tests
4. Document new patterns in this file
5. Ensure cassettes are committed
6. Update examples if needed

## License

Copyright (c) 2026 Develeap
SPDX-License-Identifier: MPL-2.0
