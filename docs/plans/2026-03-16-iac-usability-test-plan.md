# IaC Usability Improvements -- Test Plan

**Date**: 2026-03-16
**Branch**: feat/hyperping-exporter (worktree: iac-usability-improvements)
**Scope**: Cross-field validators, enum validators, error diagnostics, descriptions, monitoring locations data source, bulk data source count/ids

---

## Strategy Reconciliation

The testing strategy specifies:

- **Primary**: Real Terraform acceptance tests via terraform-plugin-testing harness (mock HTTP server, no real API key needed)
- **Secondary**: Unit tests for validators, error helpers, schema factories
- **Coverage target**: 80%+ new code, 100% validators and error handlers
- **Regression**: Zero breakage to existing 667 test functions

After reconciling with the implementation plan:

1. **ValidateConfig tests do not need a mock HTTP server.** ValidateConfig runs at plan time, before any API call. Tests use `terraform-plugin-testing` with `ExpectError` on the plan step. The provider still needs a base_url for configuration, so a minimal mock server that returns 200 is sufficient for the provider to initialise, but no CRUD routes are exercised.
2. **Monitoring locations data source is fully static.** No mock server needed at all -- the data source makes no API calls. A minimal server for provider initialisation is still required by the test harness.
3. **Bulk data source count/ids tests extend existing acceptance tests.** The existing mock servers already return data; we add assertions for the new `count` and `ids` attributes to the existing test steps.
4. **Error diagnostics are pure functions.** Unit tests only; no Terraform harness needed.
5. **Enum validators on maintenance and statuspage are schema-level.** They can be tested via acceptance tests with `ExpectError` for invalid values, using existing mock servers.

No assumptions in the testing strategy are invalidated.

---

## Test Inventory

### Area 1: Cross-Field Validators (monitor_validate_config.go)

#### T01 -- ICMP protocol rejects http_method
- **Name**: ICMP monitor with explicit http_method produces plan-time error
- **Type**: integration
- **Harness**: terraform-plugin-testing (acceptance test)
- **Preconditions**: Mock server running (minimal, provider init only)
- **Actions**: Apply config with `protocol = "icmp"` and `http_method = "POST"`
- **Expected outcome**: `ExpectError` matches regex `http_method.*only valid.*http` (case-insensitive). No API call made.
- **Interactions**: Provider initialisation via mock server

#### T02 -- ICMP protocol rejects expected_status_code
- **Name**: ICMP monitor with explicit expected_status_code produces plan-time error
- **Type**: integration
- **Harness**: terraform-plugin-testing
- **Preconditions**: Mock server running
- **Actions**: Apply config with `protocol = "icmp"` and `expected_status_code = "200"`
- **Expected outcome**: `ExpectError` matches regex `expected_status_code.*only valid.*http`
- **Interactions**: None beyond provider init

#### T03 -- ICMP protocol rejects follow_redirects
- **Name**: ICMP monitor with explicit follow_redirects produces plan-time error
- **Type**: integration
- **Harness**: terraform-plugin-testing
- **Preconditions**: Mock server running
- **Actions**: Apply config with `protocol = "icmp"` and `follow_redirects = false`
- **Expected outcome**: `ExpectError` matches regex `follow_redirects.*only.*http`
- **Interactions**: None beyond provider init

#### T04 -- ICMP protocol rejects request_headers
- **Name**: ICMP monitor with explicit request_headers produces plan-time error
- **Type**: integration
- **Harness**: terraform-plugin-testing
- **Preconditions**: Mock server running
- **Actions**: Apply config with `protocol = "icmp"` and `request_headers = [{ name = "X-Test", value = "val" }]`
- **Expected outcome**: `ExpectError` matches regex `request_headers.*only valid.*http`
- **Interactions**: None beyond provider init

#### T05 -- ICMP protocol rejects request_body
- **Name**: ICMP monitor with explicit request_body produces plan-time error
- **Type**: integration
- **Harness**: terraform-plugin-testing
- **Preconditions**: Mock server running
- **Actions**: Apply config with `protocol = "icmp"` and `request_body = "test"`
- **Expected outcome**: `ExpectError` matches regex `request_body.*only valid.*http`
- **Interactions**: None beyond provider init

#### T06 -- ICMP protocol rejects required_keyword
- **Name**: ICMP monitor with explicit required_keyword produces plan-time error
- **Type**: integration
- **Harness**: terraform-plugin-testing
- **Preconditions**: Mock server running
- **Actions**: Apply config with `protocol = "icmp"` and `required_keyword = "HEALTHY"`
- **Expected outcome**: `ExpectError` matches regex `required_keyword.*only valid.*http`
- **Interactions**: None beyond provider init

#### T07 -- ICMP protocol rejects port
- **Name**: ICMP monitor with explicit port produces plan-time error
- **Type**: integration
- **Harness**: terraform-plugin-testing
- **Preconditions**: Mock server running
- **Actions**: Apply config with `protocol = "icmp"` and `port = 443`
- **Expected outcome**: `ExpectError` matches regex `port.*not valid.*icmp`
- **Interactions**: None beyond provider init

#### T08 -- Port protocol rejects HTTP-only fields
- **Name**: Port monitor with HTTP-only fields produces plan-time error
- **Type**: integration
- **Harness**: terraform-plugin-testing
- **Preconditions**: Mock server running
- **Actions**: Apply config with `protocol = "port"`, `port = 5432`, and `http_method = "POST"`
- **Expected outcome**: `ExpectError` matches regex `http_method.*only valid.*http`
- **Interactions**: None beyond provider init

#### T09 -- Port protocol requires port field
- **Name**: Port monitor without port field produces plan-time error
- **Type**: integration
- **Harness**: terraform-plugin-testing
- **Preconditions**: Mock server running
- **Actions**: Apply config with `protocol = "port"` and no `port` attribute
- **Expected outcome**: `ExpectError` matches regex `port.*required.*protocol.*port`
- **Interactions**: None beyond provider init

#### T10 -- HTTP protocol rejects port field
- **Name**: HTTP monitor with explicit port produces plan-time error
- **Type**: integration
- **Harness**: terraform-plugin-testing
- **Preconditions**: Mock server running
- **Actions**: Apply config with `protocol = "http"` and `port = 443`
- **Expected outcome**: `ExpectError` matches regex `port.*not valid.*http`
- **Interactions**: None beyond provider init

#### T11 -- HTTP protocol accepts all HTTP fields without error
- **Name**: HTTP monitor with all HTTP-specific fields succeeds
- **Type**: scenario
- **Harness**: terraform-plugin-testing
- **Preconditions**: Mock server running with CRUD support
- **Actions**: Apply config with `protocol = "http"`, `http_method = "POST"`, `expected_status_code = "201"`, `follow_redirects = false`, `request_headers`, `request_body`, `required_keyword = "OK"`
- **Expected outcome**: No error. Resource created successfully. All attributes match.
- **Interactions**: Mock server CRUD

#### T12 -- Default values do not trigger cross-field errors for non-HTTP protocols
- **Name**: ICMP monitor with only required fields (defaults are null in raw config) succeeds
- **Type**: scenario
- **Harness**: terraform-plugin-testing
- **Preconditions**: Mock server running with CRUD support
- **Actions**: Apply config with `protocol = "icmp"`, `name`, `url` only (no HTTP fields explicitly set)
- **Expected outcome**: No error. Resource created. `http_method` defaults to "GET" in state (from schema default), but ValidateConfig does not flag it because raw config has null for http_method.
- **Interactions**: Mock server CRUD

#### T13 -- Unknown/null protocol skips validation
- **Name**: ValidateConfig skips validation when protocol is unknown (module composition)
- **Type**: unit
- **Harness**: Go testing (direct ValidateConfig call)
- **Preconditions**: MonitorResource struct instantiated
- **Actions**: Call ValidateConfig with a config where protocol is `types.StringUnknown()` and http_method is set
- **Expected outcome**: No diagnostics errors. Validation skipped gracefully.
- **Interactions**: None

#### T14 -- Null protocol skips validation
- **Name**: ValidateConfig skips validation when protocol is null
- **Type**: unit
- **Harness**: Go testing (direct ValidateConfig call)
- **Preconditions**: MonitorResource struct instantiated
- **Actions**: Call ValidateConfig with a config where protocol is `types.StringNull()` and http_method is set
- **Expected outcome**: No diagnostics errors. Validation skipped gracefully.
- **Interactions**: None

#### T15 -- Port protocol rejects required_keyword
- **Name**: Port monitor with required_keyword produces plan-time error
- **Type**: integration
- **Harness**: terraform-plugin-testing
- **Preconditions**: Mock server running
- **Actions**: Apply config with `protocol = "port"`, `port = 5432`, `required_keyword = "HEALTHY"`
- **Expected outcome**: `ExpectError` matches regex `required_keyword.*only valid.*http`
- **Interactions**: None beyond provider init

#### T16 -- Multiple invalid fields produce multiple errors
- **Name**: ICMP monitor with multiple HTTP fields produces diagnostics for each
- **Type**: boundary
- **Harness**: terraform-plugin-testing
- **Preconditions**: Mock server running
- **Actions**: Apply config with `protocol = "icmp"`, `http_method = "POST"`, `expected_status_code = "200"`, `port = 443`
- **Expected outcome**: `ExpectError` matches at least one of the invalid field patterns. (Terraform reports first error encountered; the test verifies the plan fails.)
- **Interactions**: None beyond provider init

---

### Area 2: Schema-Level Enum Validators

#### T17 -- Maintenance notification_option rejects invalid value
- **Name**: Maintenance with invalid notification_option produces plan-time error
- **Type**: integration
- **Harness**: terraform-plugin-testing
- **Preconditions**: Mock server running (maintenance CRUD)
- **Actions**: Apply config with `notification_option = "invalid_option"`
- **Expected outcome**: `ExpectError` matches regex `notification_option.*value must be one of`
- **Interactions**: None beyond provider init

#### T18 -- Maintenance notification_option accepts valid values
- **Name**: Maintenance with notification_option "immediate" succeeds
- **Type**: scenario
- **Harness**: terraform-plugin-testing
- **Preconditions**: Mock server running (maintenance CRUD)
- **Actions**: Apply config with `notification_option = "immediate"` (and all other required fields)
- **Expected outcome**: No error. Resource created. `notification_option` = "immediate" in state.
- **Interactions**: Mock server CRUD

#### T19 -- StatusPage languages rejects invalid language code
- **Name**: StatusPage with invalid language code produces plan-time error
- **Type**: integration
- **Harness**: terraform-plugin-testing
- **Preconditions**: Mock statuspage server running
- **Actions**: Apply config with `settings = { languages = ["en", "xx"] }`
- **Expected outcome**: `ExpectError` matches regex `value must be one of`
- **Interactions**: None beyond provider init

#### T20 -- StatusPage default_language rejects invalid value
- **Name**: StatusPage with invalid default_language produces plan-time error
- **Type**: integration
- **Harness**: terraform-plugin-testing
- **Preconditions**: Mock statuspage server running
- **Actions**: Apply config with `settings = { default_language = "xx", languages = ["en"] }`
- **Expected outcome**: `ExpectError` matches regex `value must be one of`
- **Interactions**: None beyond provider init

#### T21 -- StatusPage languages accepts all valid language codes
- **Name**: StatusPage with valid language codes succeeds
- **Type**: scenario
- **Harness**: terraform-plugin-testing
- **Preconditions**: Mock statuspage server running
- **Actions**: Apply config with `settings = { languages = ["en", "fr", "de"] }`
- **Expected outcome**: No error. Resource created with languages in state.
- **Interactions**: Mock server CRUD

---

### Area 3: Error Diagnostics (error_diagnostics.go)

#### T22 -- ValidValueReference returns Monitor reference table
- **Name**: ValidValueReference("Monitor") includes all monitor enum fields
- **Type**: unit
- **Harness**: Go testing
- **Preconditions**: None
- **Actions**: Call `ValidValueReference("Monitor")`
- **Expected outcome**: Non-empty string containing "protocol", "http_method", "check_frequency", "expected_status_code", "regions", "alerts_wait". Each value list matches corresponding `client.Allowed*` constants.
- **Interactions**: Reads from `client.Allowed*` package-level vars

#### T23 -- ValidValueReference returns Maintenance Window reference table
- **Name**: ValidValueReference("Maintenance Window") includes notification_option
- **Type**: unit
- **Harness**: Go testing
- **Preconditions**: None
- **Actions**: Call `ValidValueReference("Maintenance Window")`
- **Expected outcome**: Non-empty string containing "notification_option" and both valid values "scheduled", "immediate".
- **Interactions**: Reads from `client.AllowedNotificationOptions`

#### T24 -- ValidValueReference returns Incident reference table
- **Name**: ValidValueReference("Incident") includes incident type field
- **Type**: unit
- **Harness**: Go testing
- **Preconditions**: None
- **Actions**: Call `ValidValueReference("Incident")`
- **Expected outcome**: Non-empty string containing "type" and values "outage", "incident".
- **Interactions**: Reads from `client.AllowedIncidentTypes`

#### T25 -- ValidValueReference returns empty for unknown resource type
- **Name**: ValidValueReference("UnknownType") returns empty string
- **Type**: boundary
- **Harness**: Go testing
- **Preconditions**: None
- **Actions**: Call `ValidValueReference("UnknownType")`
- **Expected outcome**: Empty string returned. No panic.
- **Interactions**: None

#### T26 -- newCreateError includes valid value reference
- **Name**: newCreateError("Monitor", err) output includes Quick Reference section
- **Type**: unit
- **Harness**: Go testing
- **Preconditions**: None
- **Actions**: Call `newCreateError("Monitor", errors.New("validation failed"))`
- **Expected outcome**: Diagnostic detail contains "Quick Reference" and "protocol" entries.
- **Interactions**: None

#### T27 -- newUpdateError includes valid value reference
- **Name**: newUpdateError("Monitor", id, err) output includes Quick Reference section
- **Type**: unit
- **Harness**: Go testing
- **Preconditions**: None
- **Actions**: Call `newUpdateError("Monitor", "mon-123", errors.New("validation failed"))`
- **Expected outcome**: Diagnostic detail contains "Quick Reference" and "protocol" entries.
- **Interactions**: None

#### T28 -- newCreateError for non-reference resource type has no Quick Reference
- **Name**: newCreateError("Outage", err) output does not include Quick Reference
- **Type**: boundary
- **Harness**: Go testing
- **Preconditions**: None
- **Actions**: Call `newCreateError("Outage", errors.New("server error"))`
- **Expected outcome**: Diagnostic detail does NOT contain "Quick Reference". Existing troubleshooting section still present.
- **Interactions**: None

---

### Area 4: Monitoring Locations Data Source

#### T29 -- Data source returns all 8 locations
- **Name**: hyperping_monitoring_locations returns all known regions
- **Type**: integration
- **Harness**: terraform-plugin-testing
- **Preconditions**: Minimal mock server for provider init
- **Actions**: Apply config reading `data "hyperping_monitoring_locations" "all" {}`
- **Expected outcome**: `locations.#` = 8. `ids.#` = 8.
- **Interactions**: Provider init only (no API calls for data source)

#### T30 -- ids list matches location id fields
- **Name**: ids convenience attribute matches location objects
- **Type**: integration
- **Harness**: terraform-plugin-testing
- **Preconditions**: Minimal mock server for provider init
- **Actions**: Apply config reading data source
- **Expected outcome**: `ids` list contains "london", "frankfurt", "singapore", "sydney", "tokyo", "virginia", "saopaulo", "bahrain".
- **Interactions**: Provider init only

#### T31 -- Location metadata correctness
- **Name**: Each location has correct name, continent, and cloud_region
- **Type**: unit
- **Harness**: Go testing
- **Preconditions**: None
- **Actions**: Iterate the package-level `monitoringLocations` variable
- **Expected outcome**: "london" maps to "London, UK" / "Europe" / "eu-west-2". "virginia" maps to "Virginia, US" / "North America" / "us-east-1". All 8 entries present with non-empty name, continent, cloud_region.
- **Interactions**: None

#### T32 -- All client.AllowedRegions are covered
- **Name**: Every entry in client.AllowedRegions has a corresponding monitoring location
- **Type**: invariant
- **Harness**: Go testing
- **Preconditions**: None
- **Actions**: Compare `client.AllowedRegions` entries against `monitoringLocations` keys
- **Expected outcome**: 1:1 mapping. No region in AllowedRegions is missing from the data source. No extra entries in the data source.
- **Interactions**: None

#### T33 -- Data source Metadata returns correct type name
- **Name**: MonitoringLocationsDataSource.Metadata returns "hyperping_monitoring_locations"
- **Type**: unit
- **Harness**: Go testing
- **Preconditions**: None
- **Actions**: Call Metadata on MonitoringLocationsDataSource
- **Expected outcome**: TypeName = "hyperping_monitoring_locations"
- **Interactions**: None

#### T34 -- Data source Schema has required attributes
- **Name**: Schema includes locations and ids attributes
- **Type**: unit
- **Harness**: Go testing
- **Preconditions**: None
- **Actions**: Call Schema on MonitoringLocationsDataSource
- **Expected outcome**: Schema.Attributes contains "locations" and "ids" keys
- **Interactions**: None

---

### Area 5: Bulk Data Source count/ids

#### T35 -- Monitors data source returns count and ids
- **Name**: hyperping_monitors includes count and ids attributes
- **Type**: integration
- **Harness**: terraform-plugin-testing (extend existing TestAccMonitorsDataSource_basic)
- **Preconditions**: Mock server with 2 pre-created monitors
- **Actions**: Apply config reading monitors data source
- **Expected outcome**: `count` = "2". `ids.#` = "2". Each `ids.*` value matches a monitor UUID.
- **Interactions**: Mock server list monitors API

#### T36 -- Monitors data source empty returns count=0 and ids empty
- **Name**: hyperping_monitors empty returns count=0
- **Type**: boundary
- **Harness**: terraform-plugin-testing (extend existing TestAccMonitorsDataSource_empty)
- **Preconditions**: Mock server with 0 monitors
- **Actions**: Apply config reading monitors data source
- **Expected outcome**: `count` = "0". `ids.#` = "0".
- **Interactions**: Mock server list monitors API

#### T37 -- Incidents data source returns count and ids
- **Name**: hyperping_incidents includes count and ids attributes
- **Type**: integration
- **Harness**: terraform-plugin-testing (extend existing TestAccIncidentsDataSource_basic)
- **Preconditions**: Mock server with 3 pre-created incidents
- **Actions**: Apply config reading incidents data source
- **Expected outcome**: `count` = "3". `ids.#` = "3".
- **Interactions**: Mock server list incidents API

#### T38 -- Healthchecks data source returns count and ids
- **Name**: hyperping_healthchecks includes count and ids attributes
- **Type**: integration
- **Harness**: terraform-plugin-testing (extend existing TestAccHealthchecksDataSource_basic)
- **Preconditions**: Mock server with 2 pre-created healthchecks
- **Actions**: Apply config reading healthchecks data source
- **Expected outcome**: `count` = "2". `ids.#` = "2".
- **Interactions**: Mock server list healthchecks API

#### T39 -- Maintenance windows data source returns count and ids
- **Name**: hyperping_maintenance_windows includes count and ids attributes
- **Type**: integration
- **Harness**: terraform-plugin-testing (new acceptance test)
- **Preconditions**: Mock server with pre-created maintenance windows
- **Actions**: Apply config reading maintenance_windows data source
- **Expected outcome**: `count` matches number of maintenance windows. `ids.#` matches.
- **Interactions**: Mock server list maintenance API

#### T40 -- Outages data source returns count and ids
- **Name**: hyperping_outages includes count and ids attributes
- **Type**: integration
- **Harness**: terraform-plugin-testing (extend existing TestAccOutagesDataSource_basic)
- **Preconditions**: Mock server with 3 pre-created outages
- **Actions**: Apply config reading outages data source
- **Expected outcome**: `count` = "3". `ids.#` = "3".
- **Interactions**: Mock server list outages API

#### T41 -- StatusPages data source returns ids (total already exists)
- **Name**: hyperping_statuspages includes ids attribute
- **Type**: integration
- **Harness**: terraform-plugin-testing (extend existing TestAccStatusPagesDataSource_listAll)
- **Preconditions**: Mock statuspage server with 3 pre-created statuspages
- **Actions**: Apply config reading statuspages data source
- **Expected outcome**: `ids.#` = "3". `total` = "3" (already tested, serves as count).
- **Interactions**: Mock statuspage server

#### T42 -- StatusPage subscribers data source returns ids (total already exists)
- **Name**: hyperping_statuspage_subscribers includes ids attribute
- **Type**: integration
- **Harness**: terraform-plugin-testing (extend existing TestAccStatusPageSubscribersDataSource_listAll)
- **Preconditions**: Mock statuspage server with 4 subscribers
- **Actions**: Apply config reading statuspage_subscribers data source
- **Expected outcome**: `ids.#` = "4". `total` = "4" (already tested).
- **Interactions**: Mock statuspage server

#### T43 -- Monitors data source count/ids with filter
- **Name**: hyperping_monitors with filter returns filtered count and ids
- **Type**: integration
- **Harness**: terraform-plugin-testing (extend existing TestAccMonitorsDataSource_filterByStatus)
- **Preconditions**: Mock server with 3 monitors (2 up, 1 down)
- **Actions**: Apply config with `filter = { status = "up" }`
- **Expected outcome**: `count` = "2". `ids.#` = "2". IDs correspond to the "up" monitors only.
- **Interactions**: Mock server list monitors API with client-side filtering

#### T44 -- Monitors data source Schema includes count and ids
- **Name**: MonitorsDataSource schema has count and ids attributes
- **Type**: unit
- **Harness**: Go testing (extend existing TestMonitorsDataSource_Schema)
- **Preconditions**: None
- **Actions**: Call Schema on MonitorsDataSource
- **Expected outcome**: Schema.Attributes contains "count" and "ids" keys
- **Interactions**: None

---

### Area 6: Provider Registration

#### T45 -- Provider registers monitoring locations data source
- **Name**: Provider DataSources includes monitoring locations
- **Type**: unit
- **Harness**: Go testing (extend existing TestProvider_DataSources)
- **Preconditions**: None
- **Actions**: Call DataSources on HyperpingProvider
- **Expected outcome**: Length = 16 (was 15, now +1 for monitoring_locations). Existing data sources still present.
- **Interactions**: None

---

### Area 7: Regression

#### T46 -- Existing monitor CRUD tests pass
- **Name**: All existing TestAccMonitorResource_* tests pass without modification
- **Type**: regression
- **Harness**: `go test ./internal/provider/ -run TestAccMonitorResource -v`
- **Preconditions**: No changes to CRUD logic
- **Actions**: Run full monitor resource test suite
- **Expected outcome**: 0 failures. All existing assertions hold.
- **Interactions**: All existing mock servers

#### T47 -- Existing data source tests pass
- **Name**: All existing TestAcc*DataSource_* tests pass
- **Type**: regression
- **Harness**: `go test ./internal/provider/ -run "TestAcc.*DataSource" -v`
- **Preconditions**: New count/ids attributes are Computed-only (do not affect existing configs)
- **Actions**: Run full data source test suite
- **Expected outcome**: 0 failures. Existing assertions hold. New computed attributes do not cause "unexpected new value" errors because they are Computed.
- **Interactions**: All existing mock servers

#### T48 -- Existing validator tests pass
- **Name**: All existing TestNoSlackSubscriberType_*, TestRequiredWhenValueIs_*, etc. pass
- **Type**: regression
- **Harness**: `go test ./internal/provider/ -run "Test.*Validator\|Test.*Conditional\|TestNoSlack" -v`
- **Preconditions**: validators_conditional.go and validators.go unchanged
- **Actions**: Run existing validator test suite
- **Expected outcome**: 0 failures
- **Interactions**: None

#### T49 -- Existing error helper tests pass
- **Name**: All existing TestNewCreateError, TestNewUpdateError, etc. pass
- **Type**: regression
- **Harness**: `go test ./internal/provider/ -run "TestNew.*Error\|TestBuild" -v`
- **Preconditions**: newCreateError and newUpdateError are modified (Quick Reference appended)
- **Actions**: Run existing error helper tests
- **Expected outcome**: TestNewCreateError must be updated to also assert "Quick Reference" is present (for Monitor). TestNewUpdateError assertion for "Troubleshooting" still holds. 0 unexpected failures.
- **Interactions**: None

#### T50 -- Full lint passes
- **Name**: golangci-lint reports 0 issues
- **Type**: invariant
- **Harness**: `make lint`
- **Preconditions**: All new code follows project conventions
- **Actions**: Run linter
- **Expected outcome**: Exit code 0. No warnings.
- **Interactions**: None

#### T51 -- Full test suite passes
- **Name**: `make test` exits cleanly
- **Type**: invariant
- **Harness**: `make test`
- **Preconditions**: All new and modified tests pass
- **Actions**: Run `make test`
- **Expected outcome**: Exit code 0. 0 failures across all 667+ test functions.
- **Interactions**: All test infrastructure

---

## Test File Mapping

| Test IDs | File | Status |
|----------|------|--------|
| T01-T16 | `internal/provider/monitor_validate_config_test.go` (new) | To create |
| T17-T18 | `internal/provider/maintenance_resource_test.go` (extend) or new acceptance test file | To create/extend |
| T19-T21 | `internal/provider/statuspage_resource_test.go` (extend) or new acceptance test file | To create/extend |
| T22-T28 | `internal/provider/error_diagnostics_test.go` (new) | To create |
| T29-T34 | `internal/provider/monitoring_locations_data_source_test.go` (new) | To create |
| T35-T36, T43-T44 | `internal/provider/monitors_data_source_test.go` (extend) | To extend |
| T37 | `internal/provider/incidents_data_source_acceptance_test.go` (extend) | To extend |
| T38 | `internal/provider/healthchecks_data_source_acceptance_test.go` (extend) | To extend |
| T39 | `internal/provider/maintenance_windows_data_source_acceptance_test.go` (extend) | To extend |
| T40 | `internal/provider/outages_data_source_acceptance_test.go` (extend) | To extend |
| T41 | `internal/provider/statuspages_data_source_test.go` (extend) | To extend |
| T42 | `internal/provider/statuspage_subscribers_data_source_test.go` (extend) | To extend |
| T45 | `internal/provider/provider_test.go` (extend) | To extend |
| T46-T51 | Existing test files (run, no modification) | Verify |

---

## Priority Order

Tests are ordered by quality impact (integration/scenario first, then boundary, regression, unit last).

1. **T01-T12, T15-T16** -- Cross-field validator integration tests (highest user-facing impact)
2. **T17-T21** -- Enum validator integration tests (catches invalid values at plan time)
3. **T29-T30** -- Monitoring locations data source integration tests (new user-facing feature)
4. **T35-T43** -- Bulk data source count/ids integration tests (new user-facing attributes)
5. **T11-T12, T18, T21** -- Positive scenario tests (verify happy path still works)
6. **T09, T16, T25, T28, T36** -- Boundary tests (edge cases, empty states)
7. **T46-T49** -- Regression tests (verify no breakage)
8. **T50-T51** -- Invariant tests (lint, full suite)
9. **T13-T14, T22-T27, T31-T34, T44-T45** -- Unit tests (pure function validation)

---

## Coverage Expectations

| Component | New Test Functions | Coverage Target |
|-----------|-------------------|-----------------|
| monitor_validate_config.go | ~16 (T01-T16) | 100% |
| error_diagnostics.go | ~7 (T22-T28) | 100% |
| monitoring_locations_data_source.go | ~6 (T29-T34) | 100% |
| Bulk data source count/ids (across 7 files) | ~10 (T35-T44) | 80%+ (new code paths) |
| Enum validators (maintenance, statuspage) | ~5 (T17-T21) | 100% (validator code paths) |
| error_helpers.go changes | 0 new (T26-T28 cover) | 100% (modified lines) |

**Total new test functions**: ~44
**Total test functions after**: ~711+
