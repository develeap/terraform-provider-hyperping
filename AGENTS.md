# AGENTS.md

This file provides context for AI coding agents working on the Terraform Provider for Hyperping.

## Project Overview

**What**: Terraform provider enabling Infrastructure as Code management of Hyperping monitoring resources (monitors, incidents, maintenance windows, status pages).

**Language**: Go 1.24+
**Framework**: Terraform Plugin Framework
**API**: Hyperping REST API (https://api.hyperping.io)

## Quick Start Commands

```bash
# Build provider
go build -o terraform-provider-hyperping

# Run unit tests
go test ./...

# Run acceptance tests (requires test API key in environment)
TF_ACC=1 go test ./internal/provider -v

# Run specific test
TF_ACC=1 go test -run TestAccMonitorResource_basic ./internal/provider -v

# Lint (must pass with 0 issues)
golangci-lint run

# Generate Terraform documentation
go generate ./...

# Build all migration tools
go build ./cmd/migrate-betterstack
go build ./cmd/migrate-uptimerobot
go build ./cmd/migrate-pingdom
go build ./cmd/import-generator
```

## Project Structure

```
terraform-provider-hyperping/
├── internal/
│   ├── provider/          # Terraform resources & data sources
│   │   ├── *_resource.go  # Resource implementations
│   │   ├── *_data_source.go
│   │   └── *_test.go      # Acceptance tests
│   ├── client/            # Hyperping API client
│   │   ├── models_*.go    # API models
│   │   ├── *_contract_test.go  # API contract tests
│   │   └── *_test.go      # Unit tests with VCR
│   └── errors/            # Enhanced error handling
├── cmd/
│   ├── migrate-betterstack/   # Better Stack → Hyperping migration
│   ├── migrate-uptimerobot/   # UptimeRobot → Hyperping migration
│   ├── migrate-pingdom/       # Pingdom → Hyperping migration
│   └── import-generator/      # Bulk Terraform import tool
├── pkg/
│   ├── interactive/       # Interactive CLI utilities
│   ├── dryrun/           # Dry-run preview system
│   ├── checkpoint/       # Migration checkpoint/resume
│   └── recovery/         # Error recovery utilities
├── docs/                 # User documentation
├── examples/             # Terraform examples
└── test/
    ├── e2e/              # End-to-end tests
    ├── integration/      # Integration tests
    └── load/             # Load/performance tests
```

## Code Style & Patterns

### Naming Conventions
- **Resources**: `hyperping_monitor`, `hyperping_incident`, etc.
- **Functions**: camelCase (`mapMonitorToModel`)
- **Files**: snake_case (`monitor_resource.go`)
- **Tests**: `TestAccResourceName_scenario` for acceptance tests

### Common Patterns

**1. Save-Restore Pattern (Write-Only Fields)**
```go
// For fields API accepts but doesn't return (e.g., required_keyword, incident.text)
planField := plan.Field
r.mapAPIToModel(apiResp, &plan)
if !planField.IsNull() {
    plan.Field = planField  // Restore from plan
}
```

**2. Read-After-Create Pattern**
```go
// Create returns incomplete data, read full resource
createResp, err := r.client.Create(ctx, req)
monitor, err := r.client.Get(ctx, createResp.UUID)  // Full read
```

**3. Test Patterns**
- Unit tests: Use VCR fixtures (`internal/client/*_test.go`)
- Acceptance tests: Real API calls with `TF_ACC=1`
- Test helpers: `monitor_resource_test_helpers.go`

## Hyperping API Context

### Authentication
- Method: Bearer token
- Header: `Authorization: Bearer {api_key}`
- Format: Starts with `sk_`
- **Source**: Environment variable only (never hardcoded)

### Base URL
- Production: `https://api.hyperping.io`
- Test override: Set via provider `base_url` attribute

### API Versioning
Different endpoints use different versions:
- Monitors: `/v1/monitors`
- Incidents: `/v3/incidents`
- Maintenance: `/v1/maintenance-windows`
- Reports: `/v2/reporting/monitor-reports`

### Key API Quirks
1. **Write-only fields**: Some fields accepted in POST/PUT but not returned in GET
   - `monitor.required_keyword`
   - `incident.text`
   - `maintenance.text`
   - **Solution**: Use save-restore pattern

2. **Read-after-create**: Create responses often incomplete
   - Always fetch full resource after create
   - Pattern: `client.Create()` → `client.Get(uuid)`

3. **Response wrapping**: List endpoints inconsistent
   ```go
   // Can return:
   [{}]                    // Direct array
   {"monitors": [{}]}      // Wrapped in key
   {"data": [{}]}          // Wrapped in "data"
   ```

4. **Protocol-specific fields**: HTTP fields irrelevant for ICMP/Port monitors
   - Save plan values before API mapping
   - Restore for non-HTTP protocols

### Rate Limits
- Handle 429 responses
- Respect `Retry-After` header
- Exponential backoff implemented in `internal/client/transport.go`

## Testing Requirements

### Environment Variables
**Required for acceptance tests:**
- `TF_ACC=1` - Enable acceptance tests
- `HYPERPING_TEST_API_KEY` - Test account API key (not committed)

**Optional for integration tests:**
- `BETTERSTACK_API_TOKEN` - Better Stack test account
- `UPTIMEROBOT_API_KEY` - UptimeRobot test account
- `PINGDOM_API_KEY` - Pingdom test account (paid only)

**Note**: Actual API keys are NEVER committed. Set in environment or CI secrets only.

### Test Execution
```bash
# Unit tests (no API key needed)
go test ./internal/client
go test ./internal/provider

# Acceptance tests (requires HYPERPING_TEST_API_KEY)
TF_ACC=1 HYPERPING_TEST_API_KEY=sk_xxx go test ./internal/provider -v

# Integration tests (currently skip if keys missing)
TF_ACC=1 go test ./cmd/migrate-betterstack -tags=integration

# E2E tests (full workflow validation)
./scripts/run-e2e-tests.sh
```

### Test Coverage Expectations
- New code: 30-90% coverage depending on complexity
- Critical paths: 80%+ (authentication, CRUD operations)
- Error handling: High coverage (87.7% in `internal/errors/`)

## Migration Tools Context

Three CLI tools migrate monitoring from competitors to Hyperping:

**Features (all tools):**
- Interactive wizard mode (zero-config UX)
- Dry-run with compatibility scoring
- Parallel execution (5-8x speedup)
- Checkpoint/resume (fault tolerance)
- Rollback capability
- Enhanced error messages

**Tool-Specific:**
- `migrate-betterstack`: Heartbeat → cron conversion, 28 unit tests
- `migrate-uptimerobot`: 5 monitor types, contact alerts, 10 unit tests
- `migrate-pingdom`: Tag-based naming, 6 check types, 13 unit tests

**Usage Pattern:**
```bash
# Interactive mode (recommended)
migrate-betterstack

# Dry-run (preview changes)
migrate-betterstack --dry-run

# Advanced (filtering, parallel, resume)
migrate-betterstack \
  --filter-name="PROD-.*" \
  --parallel=10 \
  --resume
```

## Security Guidelines

### Secrets Management
- ✅ Use environment variables for all credentials
- ✅ Mark sensitive fields in Terraform schema
- ✅ Sanitize errors (mask API keys in logs)
- ❌ NEVER commit API keys to repo
- ❌ NEVER hardcode credentials in code
- ❌ NEVER include credentials in tests (use env vars)

### Error Sanitization
Implemented in `internal/client/errors.go`:
- API keys (`sk_*`) → `sk_***REDACTED***`
- Bearer tokens → `Bearer ***REDACTED***`
- URL credentials → `://***REDACTED***@`

## PR & Release Workflow

### Commit Convention
```
<type>: <description>

Types: feat, fix, refactor, docs, test, chore, perf, ci
```

### Pre-commit Checks (lefthook)
- `go fmt` - Code formatting
- `go mod tidy` - Dependency cleanup
- `generate-docs` - Terraform docs generation
- All must pass before commit

### Pre-push Checks (lefthook)
- `go vet` - Static analysis
- `go build` - Compilation
- `golangci-lint` - Linting (MUST be 0 issues)
- `go test` - Unit tests
- All must pass before push

### Release Process
1. Update `CHANGELOG.md` (Keep a Changelog format)
2. Commit: `chore: release vX.Y.Z`
3. Tag: `git tag -a vX.Y.Z -m "Release vX.Y.Z"`
4. Push: `git push origin main && git push origin vX.Y.Z`
5. GitHub Actions builds binaries, signs with GPG, publishes release
6. Terraform Registry auto-detects new version

### Version Semantics
- **Patch** (v1.0.X): Bug fixes, no breaking changes
- **Minor** (v1.X.0): New features, backward compatible
- **Major** (vX.0.0): Breaking changes (avoid if possible)

## Common Tasks

### Adding a New Resource
1. Create `internal/provider/{name}_resource.go`
2. Define model struct with `tfsdk` tags
3. Implement CRUD functions (Create, Read, Update, Delete)
4. Add acceptance tests in `{name}_resource_crud_test.go`
5. Add examples in `examples/resources/{name}/`
6. Run `go generate ./...` to generate docs
7. Update `CHANGELOG.md`

### Adding a New Data Source
1. Create `internal/provider/{name}_data_source.go`
2. Define model struct
3. Implement Read function
4. Add tests in `{name}_data_source_test.go`
5. Add examples
6. Generate docs

### Fixing State Drift Bugs
**Pattern**: Field accepted by API but not returned
**Symptoms**: "Provider produced inconsistent result after apply"
**Solution**: Save-restore pattern (see monitor `required_keyword` fix in v1.2.1)

### Adding API Coverage
1. Check `docs/API_COMPLETENESS_AUDIT.md` for gaps
2. Add client methods in `internal/client/`
3. Add models in `internal/client/models_*.go`
4. Add contract tests to verify API behavior
5. Implement provider resource/data source
6. Update audit doc

## Documentation

### Auto-Generated
- `docs/resources/*.md` - Resource docs (generated from schema)
- `docs/data-sources/*.md` - Data source docs (generated)
- Run `go generate ./...` to regenerate

### Manual
- `docs/guides/*.md` - User guides (migration, best practices, etc.)
- `docs/*.md` - Architecture, testing, error recovery
- `README.md` - Project overview
- `CHANGELOG.md` - Release notes (Keep a Changelog format)

## Known Issues & Gotchas

### Issue #32: Integration Tests Skip in CI
Integration tests for migration tools skip when source platform API keys are missing. Setting up free accounts for Better Stack and UptimeRobot would enable full testing. See issue for details.

### API Field Name Mapping
- API uses camelCase: `monitorUuid`, `createdAt`
- Terraform uses snake_case: `monitor_uuid`, `created_at`
- Client models handle translation via JSON tags

### Monitor Protocol Differences
HTTP monitors have fields (http_method, expected_status_code) that don't apply to ICMP/Port. Use save-restore pattern to prevent drift for non-HTTP protocols.

### VCR Test Fixtures
- Located in `internal/client/fixtures/`
- Record real API responses for repeatable tests
- Update fixtures when API changes
- Run tests with real API occasionally to verify contracts

## Performance Considerations

### Parallel Execution
- Import generator supports parallel imports (5-8x speedup)
- Configurable worker pools (default: 10)
- Benchmarks: 100 monitors: 5m → 45s

### Rate Limiting
- Exponential backoff implemented
- Respects `Retry-After` header
- Circuit breaker pattern in transport layer

### Memory Usage
- Load tests: <500MB for 100 monitor migrations
- No known memory leaks
- Profile with `go test -memprofile` if needed

## Dependencies

### Core
- `terraform-plugin-framework` v1.17.0 - Provider framework
- `terraform-plugin-go` v0.29.0 - Plugin protocol
- `terraform-plugin-testing` v1.14.0 - Acceptance testing

### Interactive Mode (v1.2.0+)
- `survey/v2` v2.3.7 - Interactive prompts
- `spinner` v1.23.2 - Loading spinners
- `progressbar/v3` v3.19.0 - Progress bars

### Utilities
- `go-vcr/v3` - API response recording for tests
- `gobreaker` - Circuit breaker pattern
- `cron/v3` - Cron expression parsing

## Resources

- **Hyperping API**: https://hyperping.io/docs/api (unofficial - no public docs)
- **Terraform Plugin Framework**: https://developer.hashicorp.com/terraform/plugin/framework
- **Provider Development**: https://developer.hashicorp.com/terraform/plugin/best-practices
- **Similar Provider (Reference)**: github.com/BetterStackHQ/terraform-provider-better-uptime

## Project History

- **v1.0.0** (2026-02-02): Initial stable release
- **v1.0.3-v1.0.9**: Bug fixes, import support, filtering, error handling
- **v1.1.0** (2026-02-13): Automated migration tools (Better Stack, UptimeRobot, Pingdom)
- **v1.2.0** (2026-02-14): User experience polish (interactive mode, dry-run, parallel imports, enhanced errors)
- **v1.2.1** (2026-02-15): Bug fix for `required_keyword` state drift

## Getting Help

- **GitHub Issues**: Bug reports, feature requests
- **Discussions**: Questions, ideas
- **Provider Registry**: https://registry.terraform.io/providers/develeap/hyperping
- **Changelog**: See `CHANGELOG.md` for detailed release notes
