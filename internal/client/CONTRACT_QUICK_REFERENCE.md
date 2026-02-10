# Contract Testing Quick Reference

Quick reference guide for using the contract testing framework.

## Quick Start

### 1. Basic Test Template

```go
func TestContract_YourResource_Operation(t *testing.T) {
    RunContractTest(t, ContractTestConfig{
        CassetteName: "yourresource_operation",
        ResourceType: "YourResource",
    }, func(t *testing.T, client *Client, ctx context.Context) {
        // Your test code here
        validator := NewContractValidator(t, "YourResource")
        // ... perform operations
        // ... validate responses
    })
}
```

### 2. Record a Cassette

```bash
RECORD_MODE=true HYPERPING_API_KEY=your_key go test -v -run TestContract_YourTest
```

### 3. Replay Tests (No API Calls)

```bash
go test -v -run TestContract_YourTest
```

## Common Validators

### Resource Validators

| Resource | Validator |
|----------|-----------|
| Monitor | `validator.ValidateMonitor(monitor)` |
| Monitor List | `validator.ValidateMonitorList(monitors)` |
| Incident | `validator.ValidateIncident(incident)` |
| Incident List | `validator.ValidateIncidentList(incidents)` |
| Maintenance | `validator.ValidateMaintenance(maintenance)` |
| Maintenance List | `validator.ValidateMaintenanceList(maintenances)` |
| Status Page | `validator.ValidateStatusPage(statusPage)` |
| Status Page List | `validator.ValidateStatusPageList(response)` |
| Subscriber | `validator.ValidateSubscriber(subscriber)` |
| Subscriber List | `validator.ValidateSubscriberList(response)` |
| Healthcheck | `validator.ValidateHealthcheck(healthcheck)` |
| Healthcheck List | `validator.ValidateHealthcheckList(healthchecks)` |
| Outage | `validator.ValidateOutage(outage)` |
| Outage List | `validator.ValidateOutageList(outages)` |
| Report | `validator.ValidateMonitorReport(report)` |
| Report List | `validator.ValidateMonitorReportList(response)` |

### Field Validators

| Field Type | Validator |
|------------|-----------|
| UUID | `ValidateUUID(t, "field", value)` |
| Timestamp | `ValidateTimestamp(t, "field", value)` |
| Optional Timestamp | `ValidateOptionalTimestamp(t, "field", &value)` |
| Enum | `ValidateEnum(t, "field", value, allowedValues)` |
| Optional Enum | `ValidateOptionalEnum(t, "field", &value, allowed)` |
| String | `ValidateStringField(t, "field", value, required)` |
| String Length | `ValidateStringLength(t, "field", value, maxLen)` |
| URL | `ValidateURL(t, "field", value)` |
| Optional URL | `ValidateOptionalURL(t, "field", &value)` |
| Integer Range | `ValidateIntegerRange(t, "field", value, min, max)` |
| Positive Integer | `ValidatePositiveInteger(t, "field", value)` |
| Optional Integer | `ValidateOptionalInteger(t, "field", &value, min, max)` |
| Hex Color | `ValidateHexColor(t, "field", value)` |
| Localized Text | `ValidateLocalizedText(t, "field", text, maxLen)` |
| Optional Localized Text | `ValidateOptionalLocalizedText(t, "field", &text, max)` |

## Test Data Builders

| Resource | Builder |
|----------|---------|
| Monitor | `BuildTestMonitor("name")` |
| Incident | `BuildTestIncident("title", statusPageUUIDs)` |
| Maintenance | `BuildTestMaintenance("name", monitorUUIDs)` |
| Status Page | `BuildTestStatusPage("name", "subdomain")` |
| Healthcheck | `BuildTestHealthcheck("name")` |

## Cleanup Helpers

```go
// Monitors
testCtx := MonitorTestContext{Monitor: monitor, Client: client, Context: ctx}
defer CleanupMonitor(t, testCtx)

// Incidents
testCtx := IncidentTestContext{Incident: incident, Client: client, Context: ctx}
defer CleanupIncident(t, testCtx)

// Maintenance
testCtx := MaintenanceTestContext{Maintenance: maintenance, Client: client, Context: ctx}
defer CleanupMaintenance(t, testCtx)

// Status Pages
testCtx := StatusPageTestContext{StatusPage: statusPage, Client: client, Context: ctx}
defer CleanupStatusPage(t, testCtx)

// Healthchecks
testCtx := HealthcheckTestContext{Healthcheck: healthcheck, Client: client, Context: ctx}
defer CleanupHealthcheck(t, testCtx)
```

## Error Validators

```go
// Validation errors
ExpectValidationError(t, err, "fieldName")

// API errors
ExpectAPIError(t, err, 404)
ExpectNotFoundError(t, err)
ExpectUnauthorizedError(t, err)
```

## CRUD Test Pattern

```go
RunContractTest(t, ContractTestConfig{
    CassetteName: "resource_crud",
    ResourceType: "Resource",
}, func(t *testing.T, client *Client, ctx context.Context) {
    validator := NewContractValidator(t, "Resource")

    // CREATE
    createReq := BuildTestResource("name")
    resource, err := client.CreateResource(ctx, createReq)
    require.NoError(t, err)
    validator.ValidateResource(resource)
    defer CleanupResource(t, ResourceTestContext{resource, client, ctx})

    // READ
    retrieved, err := client.GetResource(ctx, resource.UUID)
    require.NoError(t, err)
    validator.ValidateResource(retrieved)

    // UPDATE
    updateReq := UpdateResourceRequest{Name: stringPtr("new name")}
    updated, err := client.UpdateResource(ctx, resource.UUID, updateReq)
    require.NoError(t, err)
    validator.ValidateResource(updated)

    // LIST
    resources, err := client.ListResources(ctx)
    require.NoError(t, err)
    validator.ValidateResourceList(resources)
})
```

## Allowed Values

### Monitor

```go
AllowedProtocols = []string{"http", "port", "icmp"}
AllowedMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
AllowedRegions = []string{"london", "frankfurt", "singapore", "sydney", "tokyo", "virginia", "saopaulo", "bahrain"}
AllowedFrequencies = []int{10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400}
```

### Incident

```go
AllowedIncidentTypes = []string{"outage", "incident"}
AllowedIncidentUpdateTypes = []string{"investigating", "identified", "update", "monitoring", "resolved"}
```

### Maintenance

```go
AllowedNotificationOptions = []string{"scheduled", "immediate"}
```

### Status Page

```go
AllowedStatusPageThemes = []string{"light", "dark", "system"}
AllowedStatusPageFonts = []string{"system-ui", "Lato", "Manrope", "Inter", "Open Sans", "Montserrat", "Poppins", "Roboto", "Raleway", "Nunito", "Merriweather", "DM Sans", "Work Sans"}
AllowedLanguages = []string{"en", "fr", "de", "ru", "nl", "es", "it", "pt", "ja", "zh"}
```

### Subscriber

```go
AllowedSubscriberTypes = []string{"email", "sms", "teams"}
```

### Healthcheck

```go
AllowedPeriodTypes = []string{"seconds", "minutes", "hours", "days"}
```

## Common Assertions

```go
// Not nil
require.NotNil(t, value, "should not be nil")

// Not empty
require.NotEmpty(t, value, "should not be empty")

// Equal
assert.Equal(t, expected, actual, "should match")

// Contains
assert.Contains(t, slice, element, "should contain")

// Error handling
require.NoError(t, err, "should succeed")
require.Error(t, err, "should fail")

// Greater/Less
assert.Greater(t, value, min)
assert.Less(t, value, max)
assert.GreaterOrEqual(t, value, min)
assert.LessOrEqual(t, value, max)
```

## Useful Patterns

### Test with Cleanup

```go
RunContractTestWithCleanup(t, config,
    func(t *testing.T, client *Client, ctx context.Context) {
        // Test logic
    },
    func(t *testing.T, client *Client, ctx context.Context) {
        // Cleanup logic
    })
```

### Subtests

```go
validator := NewContractValidator(t, "Resource")
for i, item := range items {
    t.Run(fmt.Sprintf("Item[%d]", i), func(t *testing.T) {
        validator.ValidateItem(&item)
    })
}
```

### Conditional Validation

```go
if field != nil {
    ValidateOptionalField(t, "field", field)
}

if len(array) > 0 {
    ValidateArrayNotEmpty(t, "array", array)
}
```

## File Organization

```
internal/client/
├── contract_validators.go          # Framework validators
├── contract_test_helpers.go        # Framework helpers
├── contract_example_test.go        # Usage examples
├── monitors_contract_test.go       # Monitor contract tests
├── incidents_contract_test.go      # Incident contract tests
├── maintenance_contract_test.go    # Maintenance contract tests
├── statuspages_contract_test.go    # Status page contract tests
└── testdata/
    └── cassettes/
        ├── monitor_list.yaml
        ├── monitor_crud.yaml
        └── ...
```

## Environment Variables

```bash
# Record new cassettes (hits real API)
RECORD_MODE=true

# API key for recording
HYPERPING_API_KEY=your_key_here
```

## Common Commands

```bash
# Run all contract tests
go test -v -run TestContract ./internal/client/

# Run specific test
go test -v -run TestContract_Monitor_CRUD ./internal/client/

# Record new cassette
RECORD_MODE=true HYPERPING_API_KEY=key go test -v -run TestContract_Monitor_CRUD ./internal/client/

# Run tests with coverage
go test -v -coverprofile=coverage.out -run TestContract ./internal/client/

# View coverage
go tool cover -html=coverage.out
```

## Tips

1. **Always validate** - Use validators instead of manual assertions
2. **Clean up** - Use cleanup contexts to ensure resources are deleted
3. **Clear names** - Use descriptive cassette and test names
4. **Error cases** - Test both success and error scenarios
5. **Documentation** - Tests serve as API documentation
6. **Deterministic** - Use fixed test data, avoid random values
7. **Isolated** - Each test should be independent
8. **Fast** - Use cassettes to avoid slow API calls

## See Also

- [CONTRACT_TESTING.md](CONTRACT_TESTING.md) - Full documentation
- [contract_example_test.go](contract_example_test.go) - Complete examples
