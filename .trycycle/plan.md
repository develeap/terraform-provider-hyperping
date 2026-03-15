# IaC Usability Improvements Plan

## Goal

Make the Hyperping Terraform provider follow IaC best practices so that plan-time validation catches configuration errors before any API call is made, schema descriptions guide users without external docs, and sensible defaults reduce boilerplate. Seven areas evaluated; six require changes, one (defaults) is already complete.

---

## Inventory of Changes

### 1. Cross-Field Validators on Monitor Resource

**Problem**: `http_method`, `expected_status_code`, `follow_redirects`, `request_body`, and `request_headers` are accepted for ICMP and Port monitors, where they have no effect. `port` is accepted for HTTP monitors. `required_keyword` is accepted for non-HTTP protocols where there is no response body to search. Users discover these mistakes only at API time (or worse, silently).

**Solution**: Implement `resource.ResourceWithValidateConfig` on `MonitorResource`. The `ValidateConfig` method inspects the `protocol` value and adds plan-time `AddAttributeError` diagnostics for field combinations that cannot work.

**Decision -- why ValidateConfig, not schema-level ConflictsWith**: The terraform-plugin-framework does not support conditional ConflictsWith (i.e., "conflicts only when protocol=X"). The `ValidateConfig` interface is the idiomatic way to express cross-field business rules that depend on runtime attribute values. Every major competing provider (Datadog, Checkly, StatusCake) uses this pattern for protocol-specific field validation.

**File**: `internal/provider/monitor_resource.go`

**Changes**:
1. Add `_ resource.ResourceWithValidateConfig = &MonitorResource{}` to the interface assertion block.
2. The `ValidateConfig` method itself lives in `monitor_validate_config.go` (see below).

**File**: `internal/provider/monitor_validate_config.go` (new, ~120 lines)

Extracted to its own file because `monitor_resource.go` is already 787 lines.

**ValidateConfig implementation**:
- Read `protocol` from `req.Config` (raw user config, before defaults). If unknown/null, skip (module-composed values cannot be validated at plan time).
- When `protocol` is `"icmp"`:
  - Error if `http_method` is set (not null/unknown).
  - Error if `expected_status_code` is set.
  - Error if `follow_redirects` is explicitly set.
  - Error if `request_headers` is set.
  - Error if `request_body` is set.
  - Error if `required_keyword` is set.
  - Error if `port` is set.
- When `protocol` is `"port"`:
  - Error if `http_method` is set.
  - Error if `expected_status_code` is set.
  - Error if `follow_redirects` is explicitly set.
  - Error if `request_headers` is set.
  - Error if `request_body` is set.
  - Error if `required_keyword` is set.
  - Error if `port` is NOT set (required for port protocol). Note: the `port` attribute currently only has a `PortRange()` validator (1-65535 range check); there is no existing `RequiredWhenProtocolPort` validator applied to it in the schema. The `RequiredWhenProtocolPort` function exists in `validators_conditional.go` but is unused. The `ValidateConfig` approach supersedes it and is the sole location for this cross-field check.
- When `protocol` is `"http"`:
  - Error if `port` is set (not used for HTTP monitors; the URL contains the port).

All error messages include the field name, the invalid protocol, and what protocols the field applies to. Example: `"http_method is only valid for HTTP monitors. When protocol is \"icmp\", remove http_method or change protocol to \"http\"."`.

**Decision -- handling default values for HTTP fields**: The current defaults (`http_method = "GET"`, `expected_status_code = "2xx"`, `follow_redirects = true`) are set via schema `Default` which runs before `ValidateConfig`. However, `ValidateConfig` reads from `req.Config`, which contains only user-supplied values (raw config, before defaults are applied). Default-filled values appear as `null` in `req.Config`. Therefore, `req.Config.GetAttribute(ctx, path.Root("http_method"), &httpMethod)` returns null when the user did not explicitly set the field, and we do NOT flag it as conflicting. Only explicit user-provided values are flagged.

### 2. Schema-Level Validators for Known Enum Values

**Problem**: Some fields that accept enumerated values lack validators. Users discover invalid values only when the API rejects them (422 error).

**Current state after codebase audit**:
- `protocol` -- already has `stringvalidator.OneOf(client.AllowedProtocols...)` (good)
- `http_method` -- already has `stringvalidator.OneOf(client.AllowedMethods...)` (good)
- `check_frequency` -- already has `int64validator.OneOf(...)` (good)
- `regions` -- already has `listvalidator.ValueStringsAre(stringvalidator.OneOf(...))` (good)
- `expected_status_code` -- already has `StatusCodePattern()` (good)
- `alerts_wait` -- already has `AlertsWait()` validator (good)
- `notification_option` on maintenance -- **MISSING validator**. Currently accepts any string; should be `stringvalidator.OneOf(client.AllowedNotificationOptions...)`.
- `type` on incident -- already has `stringvalidator.OneOf(client.AllowedIncidentTypes...)` (good)
- `type` on incident_update -- already has `stringvalidator.OneOf(client.AllowedIncidentUpdateTypes...)` (good). No change needed.
- `languages` on statuspage -- **MISSING validator**. Currently accepts any string list.
- `default_language` on statuspage -- **MISSING validator**. Currently accepts any string.

**Changes**:

**File**: `internal/provider/maintenance_resource.go`
- Add `stringvalidator.OneOf(client.AllowedNotificationOptions...)` to `notification_option` attribute's Validators list.
- Add `"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"` to imports (`stringvalidator` is not currently imported in this file; the file imports `listvalidator` and `int64validator` but not `stringvalidator`).

**File**: `internal/provider/statuspage_resource_schema.go`
- `languages`: Add `listvalidator.ValueStringsAre(stringvalidator.OneOf(client.AllowedLanguages...))` to Validators list.
- `default_language`: Add `stringvalidator.OneOf(client.AllowedLanguages...)` to Validators list.
- Add `"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"` to imports (not currently imported in this file; `stringvalidator` IS already imported).

### 3. Improved Error Diagnostics with Valid Value Lists

**Problem**: When the API returns a 422 validation error, the provider's error messages include generic troubleshooting steps. The existing `buildValidationErrorSteps` in `error_helpers.go` already adds some resource-specific guidance (e.g., valid frequencies for Monitor), but the guidance is incomplete and only appears in the `WithContext` error variants. The simpler `newCreateError`/`newUpdateError` functions (which most resources use) lack this specificity.

**Solution**: Upgrade the simpler `newCreateError` and `newUpdateError` functions to include resource-specific valid value reference tables. This avoids fragile error-body parsing; instead, for each resource type, we unconditionally append a "Quick Reference" section listing all enum/constrained fields and their valid values. This is simple, robust, and directly useful when a user hits a 422.

**Decision -- unconditional reference vs. error-body parsing**: Parsing API error strings for field names is fragile (error format may change, field names in errors may not match schema names). An unconditional reference table is always correct and costs only a few extra lines in the error output. Users scanning the error will immediately see the valid values they need.

**Decision -- enhance at diagnostic layer, not client layer**: The user specification says "Do NOT change the client package API surface." Enhancing errors in the provider package respects this boundary.

**File**: `internal/provider/error_diagnostics.go` (new, ~100 lines)

This file introduces `ValidValueReference(resourceType string) string` which returns a formatted reference table of valid values for the given resource type. It reads from `client.Allowed*` constants to stay in sync.

Example output appended to a Monitor create error:
```
Quick Reference (valid values):
  protocol:             http, port, icmp
  http_method:          GET, POST, PUT, PATCH, DELETE, HEAD
  check_frequency:      10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400
  expected_status_code: Specific code (200), wildcard (2xx), or range (1xx-3xx)
  regions:              london, frankfurt, singapore, sydney, tokyo, virginia, saopaulo, bahrain
  alerts_wait:          -1, 0, 1, 2, 3, 5, 10, 30, 60
```

**File**: `internal/provider/error_helpers.go`

Changes to existing functions:
- `newCreateError`: After the existing troubleshooting block, append `ValidValueReference(resourceType)`.
- `newUpdateError`: Same treatment.

No changes to `NewCreateErrorWithContext`/`NewUpdateErrorWithContext` -- they already call `BuildTroubleshootingSteps` which provides detailed guidance. The quick reference supplements the simpler functions.

### 4. Better MarkdownDescriptions with Examples and Constraints

**Problem**: Some descriptions are terse (e.g., "The name of the monitor") and don't tell users about protocol-specific applicability, value constraints, or provide examples. Users must refer to external docs.

**Solution**: Update MarkdownDescription strings to include constraints, protocol applicability, and examples. This is a mechanical change to string literals only -- no logic changes.

**Principle**: Only update descriptions that are genuinely insufficient. Descriptions that already include valid values, defaults, and constraints are left as-is. Protocol-specific fields get explicit "Only valid when protocol is X" notes to reinforce the cross-field validators from Item 1.

**Files changed** (description-only updates, no logic changes):

**`internal/provider/monitor_resource.go`** -- Schema section:
- `name`: Add length constraint and example. Current: "The name of the monitor." Proposed: "The display name of the monitor. Must be 1-255 characters."
- `url`: Add scheme note. Current: "The URL to monitor." Proposed: "The URL to monitor. Must include protocol scheme (e.g., `https://api.example.com/health`)."
- `regions`: Add reference to new data source. Current: "List of regions to check from. Valid values: ..." Proposed: "List of monitoring regions. Use the `hyperping_monitoring_locations` data source to discover available locations. Valid values: `london`, `frankfurt`, `singapore`, `sydney`, `tokyo`, `virginia`, `saopaulo`, `bahrain`."
- `request_headers`: Add protocol note. Current: "Custom HTTP headers to send with the request." Proposed: "Custom HTTP headers to send with the request. Only valid when protocol is `http`."
- `request_body`: Add protocol and method note. Current: "Request body for POST/PUT/PATCH requests." Proposed: "HTTP request body. Only valid when protocol is `http` and http_method is `POST`, `PUT`, or `PATCH`."
- `follow_redirects`: Add protocol note. Current: "Whether to follow HTTP redirects. Defaults to `true`." Proposed: "Whether to follow HTTP redirects. Only applies to `http` protocol monitors. Defaults to `true`."
- `port`: Add examples. Current: "Port number to check. Required when `protocol` is `port`." Proposed: "TCP port number (1-65535). Required when protocol is `port`. Examples: `443` (HTTPS), `5432` (PostgreSQL), `6379` (Redis)."
- `required_keyword`: Add protocol note. Current: "A keyword that must appear in the response body for the check to pass." Proposed: "A keyword that must appear in the HTTP response body for the check to pass. Only valid when protocol is `http`."
- `protocol`, `http_method`, `check_frequency`, `expected_status_code`, `alerts_wait`, `escalation_policy`, `project_uuid`: Already good, keep as-is.

**`internal/provider/healthcheck_resource.go`** -- Schema section:
- `cron`: Already good ("Cron expression defining the schedule... Mutually exclusive with `period_value`/`period_type`."). Keep as-is.
- `timezone`: Already good ("Timezone for the cron expression... Required when `cron` is set."). Keep as-is.
- `grace_period_value`: Add practical example. Current: likely terse. Proposed: Include example combining value + type.
- `grace_period_type`: Already has OneOf. Keep as-is.

**`internal/provider/maintenance_resource.go`** -- Schema section:
- `notification_option`: Already good ("When to notify subscribers. Valid values: `scheduled`, `immediate`. Defaults to `scheduled`."). Keep as-is.
- `notification_minutes`: Already good. Keep as-is.

**`internal/provider/incident_resource.go`** -- Schema section:
- `affected_components`: Add context. Current: "List of component UUIDs affected by this incident." Proposed: "List of monitor UUIDs representing components affected by this incident. Displayed on the associated status pages."

**`internal/provider/outage_resource.go`** -- Schema section:
- `monitor_uuid`: Already good. Keep as-is.

### 5. Sensible Defaults Where Safe

**Problem**: Users have to specify `expected_status_code`, `check_frequency`, `http_method`, `follow_redirects` even when the defaults are obvious.

**Current state after audit**: The monitor resource ALREADY has defaults for all these fields:
- `protocol` defaults to `"http"` via `stringdefault.StaticString("http")`
- `http_method` defaults to `"GET"` via `stringdefault.StaticString("GET")`
- `check_frequency` defaults to `60` via `int64default.StaticInt64(client.DefaultMonitorFrequency)`
- `expected_status_code` defaults to `"2xx"` via `stringdefault.StaticString("2xx")`
- `follow_redirects` defaults to `true` via `booldefault.StaticBool(true)`
- `paused` defaults to `false` via `booldefault.StaticBool(false)`
- Incident `type` defaults to `"incident"` via `stringdefault.StaticString("incident")`
- Healthcheck `is_paused` defaults to `false`
- Maintenance `notification_option` defaults to `"scheduled"`
- Maintenance `notification_minutes` defaults to `60`

**Conclusion**: Defaults are already comprehensively implemented. No changes needed for this item.

### 6. Monitoring Locations Data Source

**Problem**: Users have to guess valid region codes. No programmatic way to discover available monitoring locations. Competitors (StatusCake, Checkly, Datadog) all provide a locations/regions data source.

**Solution**: Create a `hyperping_monitoring_locations` data source that returns the known set of monitoring regions as a structured list. Since the Hyperping API does not have a dedicated `/locations` endpoint, this data source returns the statically known regions from `client.AllowedRegions` enriched with metadata (display name, continent, cloud provider region).

**Decision -- static vs. API-backed data source**: The Hyperping API has no `/locations` or `/regions` endpoint. The allowed regions are a fixed set (`client.AllowedRegions`). Making a static data source is the correct approach because: (a) it works offline during plan, (b) it never fails due to API issues, (c) the region list changes infrequently and is tied to provider releases anyway, (d) competitors with API-backed location data sources still hardcode the enrichment metadata (display names, coordinates). If Hyperping adds a locations API in the future, this data source can be upgraded to call it while maintaining backward compatibility.

**Decision -- cloud_region field**: The `cloud_region` values (e.g., `eu-west-2`) are indicative AWS region identifiers based on the geographic mapping. They serve as a convenience for users operating in multi-cloud environments who want to co-locate monitoring probes with their infrastructure. These are approximate mappings, not authoritative infrastructure declarations.

**New Files**:

**`internal/provider/monitoring_locations_data_source.go`** (~180 lines)

```go
// MonitoringLocationsDataSource returns available monitoring regions.
type MonitoringLocationsDataSource struct{}

type MonitoringLocationsDataSourceModel struct {
    Locations []MonitoringLocationModel `tfsdk:"locations"`
    IDs       types.List                `tfsdk:"ids"`
}

type MonitoringLocationModel struct {
    ID          types.String `tfsdk:"id"`
    Name        types.String `tfsdk:"name"`
    Continent   types.String `tfsdk:"continent"`
    CloudRegion types.String `tfsdk:"cloud_region"`
}
```

Schema:
- `locations` (computed, list of objects): Each object has `id` (the region code, e.g. "london"), `name` (display name, e.g. "London, UK"), `continent` (e.g. "Europe"), `cloud_region` (e.g. "eu-west-2").
- `ids` (computed, list of strings): Convenience attribute with just the region codes, usable directly in `for_each` or as input to `hyperping_monitor.regions`.

Read implementation:
- Returns a static list derived from a package-level `monitoringLocations` variable that maps each `client.AllowedRegions` entry to its metadata.
- No API call needed. No client field on the struct.

Register in `provider.go` DataSources:
- Add `NewMonitoringLocationsDataSource` to the `DataSources()` return list.

**`internal/provider/monitoring_locations_data_source_test.go`** (~80 lines)

Unit test:
- Verify all locations are returned.
- Verify `ids` list matches location `id` fields.
- Verify known regions (london, frankfurt, etc.) are present.

Acceptance test:
- `TestAccMonitoringLocationsDataSource_basic`: Uses `terraform-plugin-testing` to verify the data source reads successfully and returns expected attributes. No API key needed (static data).

**Location metadata table** (defined as package-level var in the data source file):

| ID        | Name              | Continent     | Cloud Region   |
|-----------|-------------------|---------------|----------------|
| london    | London, UK        | Europe        | eu-west-2      |
| frankfurt | Frankfurt, DE     | Europe        | eu-central-1   |
| singapore | Singapore         | Asia Pacific  | ap-southeast-1 |
| sydney    | Sydney, AU        | Asia Pacific  | ap-southeast-2 |
| tokyo     | Tokyo, JP         | Asia Pacific  | ap-northeast-1 |
| virginia  | Virginia, US      | North America | us-east-1      |
| saopaulo  | Sao Paulo, BR     | South America | sa-east-1      |
| bahrain   | Bahrain, ME       | Middle East   | me-south-1     |

### 7. Bulk Data Source Enhancements (count + ids)

**Problem**: The list data sources (`hyperping_monitors`, `hyperping_incidents`, etc.) do not expose `count` or `ids` fields, making `for_each` and conditional patterns harder.

**Solution**: Add `count` (computed Int64) and `ids` (computed List of String) attributes to all list data sources.

**Decision -- naming convention**: Using `count` and `ids` to match the convention from major providers (AWS, Azure, GCP all use these names on list data sources).

**Files changed**:

**`internal/provider/monitors_data_source.go`**:
- Add `Count types.Int64 \`tfsdk:"count"\`` and `IDs types.List \`tfsdk:"ids"\`` to `MonitorsDataSourceModel`.
- Add schema attributes:
  - `count`: `schema.Int64Attribute{Computed: true, MarkdownDescription: "Total number of monitors returned (after filtering)."}`
  - `ids`: `schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "List of monitor UUIDs. Convenient for for_each patterns."}`
- In `Read()`, after mapping monitors, set `config.Count = types.Int64Value(int64(len(filteredMonitors)))` and build `config.IDs` from the monitor UUIDs using `types.ListValueFrom`.

**`internal/provider/incidents_data_source.go`**:
- Same pattern: add `Count` and `IDs` to model, schema, and Read.
- IDs extracted from `incident.UUID` for each filtered incident.

**`internal/provider/healthchecks_data_source.go`**:
- Same pattern: add `Count` and `IDs`.
- IDs extracted from healthcheck UUIDs.

**`internal/provider/maintenance_windows_data_source.go`**:
- Same pattern: add `Count` and `IDs`.
- IDs extracted from maintenance UUIDs.

**`internal/provider/outages_data_source.go`**:
- Same pattern: add `Count` and `IDs`.
- IDs extracted from outage UUIDs.

**`internal/provider/statuspages_data_source.go`**:
- Already has `Total types.Int64` (from API pagination). Add `IDs types.List \`tfsdk:"ids"\`` only. The `total` field serves as `count`.
- **Implementation note**: This data source uses `types.List` for `StatusPages` (not a Go slice), so IDs must be extracted during the mapping loop (before converting to `types.ListValueFrom`). Collect UUIDs into a `[]string` during the `for i, sp := range filteredStatusPages` loop, then convert to `types.ListValueFrom` after the loop.

**`internal/provider/statuspage_subscribers_data_source.go`**:
- Already has `Total types.Int64` (from API pagination). Add `IDs types.List \`tfsdk:"ids"\`` only (subscriber IDs as strings).
- **Implementation note**: Same approach as statuspages -- collect IDs during the existing mapping loop.

---

## File Change Summary

### New Files (3 source + 3 test)
| File | Lines (est.) | Purpose |
|------|-------------|---------|
| `internal/provider/monitor_validate_config.go` | ~120 | Cross-field validation for monitor protocol/field combos |
| `internal/provider/monitoring_locations_data_source.go` | ~180 | Static data source for monitoring regions |
| `internal/provider/error_diagnostics.go` | ~100 | Resource-specific valid value reference tables |
| `internal/provider/monitor_validate_config_test.go` | ~200 | Unit tests for cross-field validation |
| `internal/provider/monitoring_locations_data_source_test.go` | ~80 | Unit + acceptance tests for locations data source |
| `internal/provider/error_diagnostics_test.go` | ~80 | Unit tests for valid value references |

### Modified Files (~15)
| File | Nature of Change |
|------|-----------------|
| `internal/provider/monitor_resource.go` | Add ValidateConfig interface assertion, update MarkdownDescriptions |
| `internal/provider/maintenance_resource.go` | Add notification_option validator + stringvalidator import |
| `internal/provider/incident_resource.go` | Update MarkdownDescriptions |
| `internal/provider/healthcheck_resource.go` | Update MarkdownDescriptions |
| `internal/provider/statuspage_resource_schema.go` | Add languages + default_language validators + listvalidator import |
| `internal/provider/provider.go` | Register MonitoringLocationsDataSource |
| `internal/provider/error_helpers.go` | Integrate ValidValueReference in newCreateError/newUpdateError |
| `internal/provider/monitors_data_source.go` | Add count + ids attributes |
| `internal/provider/incidents_data_source.go` | Add count + ids attributes |
| `internal/provider/healthchecks_data_source.go` | Add count + ids attributes |
| `internal/provider/maintenance_windows_data_source.go` | Add count + ids attributes |
| `internal/provider/outages_data_source.go` | Add count + ids attributes |
| `internal/provider/statuspages_data_source.go` | Add ids attribute (already has total) |
| `internal/provider/statuspage_subscribers_data_source.go` | Add ids attribute (already has total) |

---

## Validation Strategy

### Unit Tests (no API key needed)

1. **Cross-field validators** (`monitor_validate_config_test.go`):
   - Test ICMP protocol rejects http_method, expected_status_code, follow_redirects, request_headers, request_body, required_keyword, port.
   - Test Port protocol rejects http_method, expected_status_code, follow_redirects, request_headers, request_body, required_keyword; requires port.
   - Test HTTP protocol rejects port.
   - Test unknown/null protocol skips validation (module composition support).
   - Test HTTP protocol accepts all HTTP fields without error.
   - Test that default values (null in raw config) do NOT trigger cross-field errors.

2. **Error diagnostics** (`error_diagnostics_test.go`):
   - Test `ValidValueReference("Monitor")` returns all monitor enum fields.
   - Test `ValidValueReference("Maintenance")` returns maintenance-specific fields.
   - Test `ValidValueReference("UnknownType")` returns empty string (graceful fallback).

3. **Monitoring locations** (`monitoring_locations_data_source_test.go`):
   - Test all 8 regions returned.
   - Test ids list matches location id fields.
   - Test metadata correctness (names, continents, cloud regions).

### Acceptance Tests (TF_ACC=1, mock server)

1. **Monitor cross-field validation**:
   - `TestAccMonitorResource_icmpRejectsHTTPFields`: Config with protocol=icmp and http_method=GET, expect plan error.
   - `TestAccMonitorResource_portRequiresPort`: Config with protocol=port and no port, expect plan error.
   - `TestAccMonitorResource_httpRejectsPort`: Config with protocol=http and port=443, expect plan error.

2. **Monitoring locations data source**:
   - `TestAccMonitoringLocationsDataSource_basic`: Read data source, verify count and structure. No API key needed.

3. **Bulk data source count/ids**:
   - Existing list data source tests already verify Read works. Add assertions for `count` and `ids` attributes in existing test functions.

---

## Execution Order

All changes are independent at the code level and can be implemented in parallel. However, for a clean commit history:

1. Create `error_diagnostics.go` + tests (foundation, no dependencies)
2. Create `monitor_validate_config.go` + update `monitor_resource.go` interface assertion + tests
3. Add missing enum validators (maintenance notification_option, statuspage languages/default_language)
4. Update MarkdownDescriptions across resources
5. Create `monitoring_locations_data_source.go` + tests + register in provider.go
6. Add `count`/`ids` to all list data sources
7. Run `make lint` and `make test` to verify
8. Run `go generate ./...` to regenerate docs

---

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Cross-field validation breaks existing configs that pass HTTP fields for non-HTTP monitors | Medium | High | ValidateConfig reads from `req.Config` (raw config, before defaults). Fields with schema defaults appear as `null` in raw config, so they do not trigger errors. Only explicit user-provided values are flagged. |
| Static monitoring locations become stale | Low | Low | Region changes are rare. Provider releases update the list. MarkdownDescription notes the list is as of the current provider version. |
| count/ids on data sources change state shape | Low | Medium | These are new computed-only attributes. Existing configs that do not reference them are unaffected. No breaking change. |
| Validator on notification_option rejects previously accepted invalid values | Low | Medium | Only affects configs with typos or invalid values. This is the correct behavior -- catching errors at plan time instead of API time. |
| Validator on languages/default_language rejects previously accepted values | Low | Medium | Only affects configs with invalid language codes. Same justification as above. |

---

## Non-Goals (Explicitly Out of Scope)

- No changes to `internal/client/` package API surface
- No changes to existing resource CRUD logic (Create, Read, Update, Delete methods)
- No new API calls for existing resources
- No breaking changes to existing schemas (all additions are Optional/Computed or new data sources)
- No changes to migration CLI tools
- No changes to CI/CD pipeline
