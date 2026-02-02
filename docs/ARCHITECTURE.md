# Architecture Documentation

## Project Overview

The Terraform Provider for Hyperping is a HashiCorp Plugin Framework-based provider that enables Infrastructure as Code management of Hyperping monitoring resources. The provider follows modern Go and Terraform best practices with emphasis on security, reliability, and maintainability.

## Project Structure

```
terraform-provider-hyperping/
├── internal/
│   ├── client/              # Hyperping API client (independent, reusable)
│   │   ├── client.go               # Core HTTP client with retry/circuit breaker
│   │   ├── monitors.go             # Monitor API endpoints
│   │   ├── incidents.go            # Incident API endpoints
│   │   ├── maintenance.go          # Maintenance window endpoints
│   │   ├── outages.go              # Outage API endpoints
│   │   ├── healthchecks.go         # Health check endpoints
│   │   ├── statuspages.go          # Status page endpoints
│   │   ├── reports.go              # Reporting endpoints
│   │   ├── models_*.go             # API request/response models
│   │   ├── errors.go               # Error handling and sanitization
│   │   ├── transport.go            # Custom HTTP transport (TLS 1.2+)
│   │   └── interface.go            # Client interface for testability
│   └── provider/            # Terraform provider implementation
│       ├── provider.go             # Provider configuration
│       ├── *_resource.go           # Resource implementations (CRUD)
│       ├── *_data_source.go        # Data source implementations (Read)
│       ├── mapping.go              # API ↔ Terraform model mapping
│       ├── error_helpers.go        # Standard error handling
│       ├── validators.go           # Custom validators
│       └── logging.go              # Terraform logging helpers
├── examples/                # Usage examples for documentation
│   ├── provider/                   # Provider configuration examples
│   ├── resources/                  # Resource examples
│   ├── data-sources/               # Data source examples
│   ├── complete/                   # End-to-end example
│   └── multi-tenant/               # Multi-tenant pattern example
├── docs/                    # Generated Terraform documentation
│   ├── resources/                  # Auto-generated resource docs
│   ├── data-sources/               # Auto-generated data source docs
│   ├── adr/                        # Architecture Decision Records
│   └── OPERATIONS.md               # Production operations guide
└── .github/                 # CI/CD workflows and templates
    ├── workflows/                  # GitHub Actions
    └── CONTRIBUTING.md             # Contribution guidelines
```

## Architecture Layers

### 1. API Client Layer (`internal/client/`)

The API client is **intentionally separate** from the Terraform provider to allow:
- Independent unit testing without Terraform dependencies
- Potential reuse in other Go projects
- Clear separation of concerns

#### Key Components

**Core Client (`client.go`)**
- HTTP client with customizable timeout
- Exponential backoff retry logic with jitter
- Circuit breaker pattern for fault tolerance
- Optional logging and metrics interfaces
- User-Agent tracking for API versioning

**Transport Layer (`transport.go`)**
- Enforces TLS 1.2+ for security
- Validates HTTPS endpoints
- Custom round tripper for security controls

**Error Handling (`errors.go`)**
- Sanitizes API keys from all error messages
- Structured error types (APIError)
- HTTP status code mapping
- Retry-After header parsing

**API Endpoints**
- Monitors: `/v1/monitors` - HTTP, Port, ICMP checks
- Incidents: `/v3/incidents` - Status page incidents
- Maintenance: `/v1/maintenance-windows` - Scheduled maintenance
- Outages: `/v1/outages` - Monitor outage management
- Healthchecks: `/v1/healthchecks` - Cron job monitoring
- Status Pages: `/v1/statuspages` - Public status pages
- Reports: `/v2/reporting/monitor-reports` - Uptime/SLA data

**Models (`models_*.go`)**
- Separate model files per resource type
- JSON serialization tags
- Validation methods

### 2. Provider Layer (`internal/provider/`)

Implements the HashiCorp Plugin Framework provider interface.

#### Provider Configuration (`provider.go`)

```go
type HyperpingProvider struct {
    version string  // Provider version for telemetry
}

type HyperpingProviderModel struct {
    APIKey  types.String  // Sensitive, from env or config
    BaseURL types.String  // Optional, defaults to production API
}
```

**Security Features:**
- API key validation (must start with `sk_`)
- BaseURL validation (HTTPS only, *.hyperping.io domain)
- Credential sanitization in all logs
- Environment variable support

#### Resource Pattern

All resources follow this CRUD pattern:

```go
type MyResource struct {
    client client.HyperpingClient  // Injected at runtime
}

// Metadata - Register resource type
func (r *MyResource) Metadata(ctx, req, resp)

// Schema - Define Terraform schema
func (r *MyResource) Schema(ctx, req, resp)

// Create - POST to API
func (r *MyResource) Create(ctx, req, resp) {
    // 1. Extract plan data
    // 2. Map to API request
    // 3. Call client method
    // 4. Map response to state
    // 5. Set state
}

// Read - GET from API
func (r *MyResource) Read(ctx, req, resp)

// Update - PUT to API
func (r *MyResource) Update(ctx, req, resp)

// Delete - DELETE from API
func (r *MyResource) Delete(ctx, req, resp)

// ImportState - Support terraform import
func (r *MyResource) ImportState(ctx, req, resp)
```

#### Data Source Pattern

Data sources only implement Read operations:

```go
type MyDataSource struct {
    client client.HyperpingClient
}

// Metadata, Schema (similar to resources)

// Read - Fetch data from API
func (d *MyDataSource) Read(ctx, req, resp)
```

#### Mapping Layer (`mapping.go`, `statuspage_mapping.go`)

Separates Terraform schema types from API models:

```go
// API → Terraform
func mapAPIMonitorToTF(api *client.Monitor) *MonitorResourceModel

// Terraform → API
func mapTFMonitorToAPI(tf *MonitorResourceModel) *client.CreateMonitorRequest
```

Benefits:
- Clear data flow
- Testable transformations
- Handles type conversions (types.String → string)
- Deals with nullable fields

#### Error Handling (`error_helpers.go`)

Standard error patterns for consistency:

```go
// Resource creation failures
apiCreateError(resourceType, resourceID, err)

// Resource read failures
apiReadError(resourceType, resourceID, err)

// Resource update failures
apiUpdateError(resourceType, resourceID, err)

// Resource deletion failures
apiDeleteError(resourceType, resourceID, err)

// Validation errors
validationError(field, message)
```

#### Validators (`validators.go`, `validators_conditional.go`)

Custom Terraform validators:
- URL validation
- HTTP method validation
- Status code pattern validation
- Conditional requirements (e.g., port required when protocol=port)

## Key Design Patterns

### 1. Client Interface Pattern

**Decision:** Define client as interface, use concrete implementation

**ADR:** [0001-client-interface-for-testability](adr/0001-client-interface-for-testability.md)

```go
// internal/client/interface.go
type HyperpingClient interface {
    ListMonitors(ctx context.Context) ([]Monitor, error)
    CreateMonitor(ctx context.Context, req CreateMonitorRequest) (*Monitor, error)
    // ... all other methods
}

// internal/provider/monitor_resource.go
type monitorResource struct {
    client client.HyperpingClient  // Interface, not *Client
}
```

Benefits:
- Mock client in tests
- No network calls in unit tests
- Faster test execution
- Test error scenarios easily

### 2. Single Source of Truth for Constants

**Decision:** Define constants once, reference everywhere

**ADR:** [0003-single-source-of-truth-for-constants](adr/0003-single-source-of-truth-for-constants.md)

```go
// internal/client/models_common.go
var (
    ValidHTTPMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
    ValidRegions = []string{"london", "frankfurt", "virginia", ...}
    ValidCheckFrequencies = []int{10, 20, 30, 60, 120, ...}
)

// Used in:
// 1. Client validation
// 2. Terraform schema validators
// 3. Documentation
```

### 3. Separation of Concerns

**Client Layer:** Pure Go, no Terraform dependencies
- Can be tested independently
- Could be published as standalone library
- Clear API contracts

**Provider Layer:** Terraform-specific logic only
- Schema definitions
- State management
- Terraform-specific error handling

### 4. Security-First Design

**Credential Sanitization:**
```go
// errors.go: sanitizeMessage()
// Removes any string matching sk_[a-zA-Z0-9_]+
err.Error() // "failed to create monitor: 401 Unauthorized"
            // NOT: "failed with key sk_abc123: 401"
```

**TLS Enforcement:**
```go
// transport.go
tlsConfig := &tls.Config{
    MinVersion: tls.VersionTLS12,  // Force TLS 1.2+
}
```

**Domain Validation:**
```go
// provider.go: isAllowedBaseURL()
// Only allow *.hyperping.io domains
// Require HTTPS (except localhost for testing)
```

### 5. Resilience Patterns

**Circuit Breaker:**
```go
// client.go
circuitBreaker := gobreaker.NewCircuitBreaker(gobreaker.Settings{
    Name:        "hyperping-api",
    MaxRequests: 3,
    Interval:    time.Minute,
    Timeout:     30 * time.Second,
    ReadyToTrip: func(counts gobreaker.Counts) bool {
        failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
        return counts.Requests >= 5 && failureRatio >= 0.6
    },
})
```

**Retry with Exponential Backoff:**
```go
// client.go: doRequestWithRetry()
wait := min(c.retryWaitMin * (1 << attempt), c.retryWaitMax)
wait += time.Duration(rand.Int64N(int64(wait) / 10))  // Add jitter
```

**Retry-After Header Support:**
```go
// errors.go: ParseRetryAfter()
// Respects API rate limit guidance
```

### 6. Observable Operations

**Structured Logging:**
```go
// Uses tflog for Terraform integration
tflog.Debug(ctx, "Creating monitor", map[string]interface{}{
    "name": plan.Name.ValueString(),
})
```

**Metrics Interface:**
```go
type Metrics interface {
    RecordAPICall(ctx, method, path string, statusCode int, durationMs int64)
    RecordRetry(ctx, method, path string, attempt int)
    RecordCircuitBreakerState(ctx, state string)
}
```

Allows integration with:
- Prometheus
- CloudWatch
- Datadog
- New Relic

## Testing Strategy

### Unit Tests (95.3% coverage)

**Client Tests:**
- Mock HTTP server responses
- Test error scenarios
- Validate request/response mapping
- No external dependencies

**Provider Tests:**
- Mock client interface
- Test schema validation
- Test mapping functions
- Test error handling

**Example:**
```go
func TestMonitorResource_Create(t *testing.T) {
    mockClient := &MockHyperpingClient{
        CreateMonitorFunc: func(ctx context.Context, req CreateMonitorRequest) (*Monitor, error) {
            return &Monitor{UUID: "mon_123", Name: req.Name}, nil
        },
    }
    // Test resource with mock client
}
```

### Acceptance Tests

**Purpose:** Verify against real API
**Requirement:** HYPERPING_API_KEY environment variable

```bash
TF_ACC=1 go test ./internal/provider/ -v
```

Tests:
- Create resources
- Read/import resources
- Update resources
- Delete resources
- Data source queries

### Coverage Strategy

**ADR:** [0002-coverage-threshold-strategy](adr/0002-coverage-threshold-strategy.md)

- Target: 80% overall coverage
- Client package: 95.3% (exceeds target)
- Provider package: Focus on critical paths
- Exclude generated code

## Configuration Management

### Environment Variables

```bash
# Required
HYPERPING_API_KEY=sk_your_api_key_here

# Optional (for testing)
HYPERPING_BASE_URL=https://api.hyperping.io  # Default
TF_ACC=1                                     # Enable acceptance tests
TF_LOG=DEBUG                                 # Enable debug logging
```

### Provider Configuration

```hcl
provider "hyperping" {
  api_key  = var.hyperping_api_key  # Or use env var
  base_url = "https://api.hyperping.io"  # Optional
}
```

## Release Process

### Version Management

- Semantic versioning (MAJOR.MINOR.PATCH)
- Git tags trigger releases
- GoReleaser handles builds

### Build Process

```bash
# Local build
go build -o terraform-provider-hyperping

# Release build (automated via GitHub Actions)
git tag v0.1.0
git push origin v0.1.0
# GoReleaser builds for all platforms, signs with GPG
```

### Publishing to Terraform Registry

1. Push git tag (e.g., `v0.1.0`)
2. GitHub Actions workflow triggers
3. GoReleaser builds multi-platform binaries
4. Releases signed with GPG key
5. Registry auto-detects new release

## Operational Considerations

See [OPERATIONS.md](OPERATIONS.md) for detailed guidance on:
- Production deployment
- Monitoring and alerting
- Rate limit handling
- Circuit breaker tuning
- Credential rotation
- Troubleshooting

## Extension Points

### Adding New Resources

1. Add client methods in `internal/client/`
2. Create resource file in `internal/provider/`
3. Implement CRUD methods
4. Add mapping functions
5. Write tests (unit + acceptance)
6. Add examples in `examples/resources/`
7. Generate docs: `make docs`

### Adding New Data Sources

Similar to resources but only implement Read:
1. Add client methods if needed
2. Create data source file
3. Implement Read method
4. Add examples and tests
5. Generate docs

### Custom Validators

```go
// internal/provider/validators.go
type myValidator struct{}

func (v myValidator) Description(ctx context.Context) string {
    return "validates something"
}

func (v myValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
    // Validation logic
}
```

## Security Considerations

### Threat Model

**VULN-001: Credential Exposure**
- Mitigation: Sanitize all error messages
- Mitigation: Mark fields as Sensitive in schema

**VULN-002: Man-in-the-Middle**
- Mitigation: Enforce TLS 1.2+
- Mitigation: Validate HTTPS URLs

**VULN-003: Domain Confusion**
- Mitigation: Allowlist *.hyperping.io domains

**VULN-010: User-Agent Overflow**
- Mitigation: Limit User-Agent length to 256 chars

**VULN-013: Resource ID Injection**
- Mitigation: Validate resource IDs (max 128 chars)

**VULN-014: Response Body DoS**
- Mitigation: Limit response body to 10 MB

See [SECURITY.md](../SECURITY.md) for reporting vulnerabilities.

## Performance Characteristics

### API Rate Limits

Hyperping API enforces rate limits:
- Default: ~100 requests/minute
- Returns 429 with Retry-After header
- Client automatically respects limits

### Terraform Concurrency

```bash
# Reduce parallelism if hitting rate limits
terraform apply -parallelism=1
```

### Circuit Breaker Tuning

Default settings:
- Opens after 60% failure rate
- Requires 5+ requests before tripping
- Half-open after 30 seconds
- Allows 3 requests in half-open state

## Dependencies

### Direct Dependencies

```go
require (
    github.com/hashicorp/terraform-plugin-framework v1.12.0
    github.com/hashicorp/terraform-plugin-go v0.24.0
    github.com/hashicorp/terraform-plugin-log v0.9.0
    github.com/sony/gobreaker v1.0.0
)
```

### Why These Dependencies?

- **terraform-plugin-framework**: Modern plugin SDK (preferred over SDKv2)
- **terraform-plugin-go**: Low-level protocol implementation
- **terraform-plugin-log**: Structured logging for Terraform
- **gobreaker**: Circuit breaker for resilience

## References

### Internal Documentation

- [ADR Index](adr/README.md) - Architecture Decision Records
- [OPERATIONS.md](OPERATIONS.md) - Production operations guide
- [CONTRIBUTING.md](../.github/CONTRIBUTING.md) - Development guide
- [SECURITY.md](../SECURITY.md) - Security policy

### External Resources

- [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework)
- [Terraform Plugin Best Practices](https://developer.hashicorp.com/terraform/plugin/best-practices)
- [Hyperping API Documentation](https://hyperping.notion.site/Hyperping-API-documentation-a0dc48fb818e4542a8f7fb4163ede2c3)
- [Go Project Layout](https://github.com/golang-standards/project-layout)

## Frequently Asked Questions

**Q: Why is the client in a separate package?**
A: Independence, testability, and potential reusability. See [ADR-0001](adr/0001-client-interface-for-testability.md).

**Q: Why use Plugin Framework instead of SDKv2?**
A: Plugin Framework is the modern approach with better type safety, validation, and features.

**Q: How do I test without hitting the real API?**
A: Use unit tests with mock clients. Acceptance tests require a real API key.

**Q: What's the test coverage goal?**
A: 80% overall, with critical paths at 100%. See [ADR-0002](adr/0002-coverage-threshold-strategy.md).

**Q: How do I add a new resource type?**
A: See "Adding New Resources" section above.

**Q: Why is there a circuit breaker?**
A: Prevents cascading failures during API degradation. See "Resilience Patterns" above.

**Q: How are credentials protected?**
A: Multiple layers: sanitization, TLS enforcement, domain validation. See "Security-First Design" above.
