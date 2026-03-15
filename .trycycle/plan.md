# IaC Usability Improvements Plan

## Goal

Make the Hyperping Terraform provider follow IaC best practices so that plan-time validation catches configuration errors before any API call is made, schema descriptions guide users without external docs, and sensible defaults reduce boilerplate. Six priority items, delivered as a single cutover.

---

## Inventory of Changes

### 1. Cross-Field Validators on Monitor Resource

**Problem**: `http_method`, `expected_status_code`, `follow_redirects`, `request_body`, and `request_headers` are accepted for ICMP and Port monitors, where they have no effect. `port` is accepted for HTTP monitors. `required_keyword` is accepted for non-HTTP protocols where there is no response body to search. Users discover these mistakes only at API time (or worse, silently).

**Solution**: Implement `resource.ResourceWithValidateConfig` on `MonitorResource`. The `ValidateConfig` method inspects the `protocol` value and adds plan-time `AddAttributeError` diagnostics for field combinations that cannot work.

**Decision -- why ValidateConfig, not schema-level ConflictsWith**: The terraform-plugin-framework does not support conditional ConflictsWith (i.e., "conflicts only when protocol=X"). The `ValidateConfig` interface is the idiomatic way to express cross-field business rules that depend on runtime attribute values. Every major competing provider (Datadog, Checkly, StatusCake) uses this pattern for protocol-specific field validation.

**File**: `internal/provider/monitor_resource.go`

**Changes**:
1. Add `_ resource.ResourceWithValidateConfig = &MonitorResource{}` to the interface assertion block.
2. Implement `ValidateConfig(ctx, req, resp)` method on `MonitorResource`:
   - Read `protocol` from config. If unknown/null, skip (module-composed values cannot be validated at plan time).
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
     - Error if `port` is NOT set (required for port protocol). This replaces the existing `RequiredWhenProtocolPort` validator on the `port` attribute. The schema-level validator on `port` will be kept as belt-and-suspenders, but the primary cross-field check moves into `ValidateConfig`.
   - When `protocol` is `"http"`:
     - Error if `port` is set (not used for HTTP monitors; the URL contains the port).
3. All error messages include the field name, the invalid protocol, and what protocols the field applies to. Example: `"http_method is only valid for HTTP monitors. When protocol is \"icmp\", remove http_method or change protocol to \"http\"."`.

**File**: `internal/provider/monitor_validate_config.go` (new, ~120 lines)

Extracted to its own file because `monitor_resource.go` is already 787 lines.

**Decision -- keep existing default values for HTTP fields**: The current defaults (`http_method = "GET"`, `expected_status_code = "2xx"`, `follow_redirects = true`) apply even when protocol is `icmp` or `port`, because the schema `Default` runs before `ValidateConfig`. The `ValidateConfig` check must therefore treat `Computed` + `Default` values as "not explicitly set by the user." The correct approach: only flag a field as conflicting if `config.GetAttribute` returns a non-null, non-unknown value AND the user actually specified it (i.e., it is present in `req.Config`). The framework's `req.Config` only contains user-supplied values, so default-filled values will appear as `null` or `unknown` in the config object. This means `req.Config.GetAttribute(ctx, path.Root("http_method"), &httpMethod)` will give us the raw config value before defaults. If it is null, the user did not set it, and we should not error.

### 2. Schema-Level Validators for Known Enum Values

**Problem**: Some fields that accept enumerated values lack validators (or have incomplete ones).

**Current state after codebase audit**:
- `protocol` -- already has `stringvalidator.OneOf(client.AllowedProtocols...)` (good)
- `http_method` -- already has `stringvalidator.OneOf(client.AllowedMethods...)` (good)
- `check_frequency` -- already has `int64validator.OneOf(...)` (good)
- `regions` -- already has `listvalidator.ValueStringsAre(stringvalidator.OneOf(...))` (good)
- `expected_status_code` -- already has `StatusCodePattern()` (good)
- `alerts_wait` -- already has `AlertsWait()` validator (good)
- `notification_option` on maintenance -- **MISSING validator**. Currently accepts any string; should be `stringvalidator.OneOf(client.AllowedNotificationOptions...)`.
- `type` on incident -- already has `stringvalidator.OneOf(client.AllowedIncidentTypes...)` (good)
- `type` on incident_update -- needs verification.

**Changes**:

**File**: `internal/provider/maintenance_resource.go`
- Add `stringvalidator.OneOf(client.AllowedNotificationOptions...)` to `notification_option` attribute's Validators list.
- Import `"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"` (add to imports).

**File**: `internal/provider/incident_update_resource.go`
- **Already has** `stringvalidator.OneOf(client.AllowedIncidentUpdateTypes...)` on `type`. No change needed.

**File**: `internal/provider/statuspage_resource_schema.go`
- `theme` -- already has `stringvalidator.OneOf("light", "dark", "system")` (good).
- `font` -- already has `stringvalidator.OneOf(...)` with all 13 font values (good).
- `languages` -- **MISSING validator**. Currently accepts any string list. Add `listvalidator.ValueStringsAre(stringvalidator.OneOf(client.AllowedLanguages...))`.
- `default_language` -- **MISSING validator**. Currently accepts any string. Add `stringvalidator.OneOf(client.AllowedLanguages...)`.
- Import `listvalidator` if not already present (it is not imported in this file).

### 3. Improved Error Diagnostics with Valid Value Lists

**Problem**: When the API returns 422, the raw error message is opaque. The provider already has `BuildTroubleshootingSteps` for error context, but the validation error path only adds generic guidance. We need field-specific error messages that include valid value lists.

**Solution**: Enhance the `buildValidationErrorSteps` function in `error_helpers.go` to parse the API error body for field names and append the known valid values for that field. Additionally, create an `apiErrorEnhancer` that wraps API errors and enriches them with field-specific guidance before they reach the diagnostic layer.

**File**: `internal/provider/error_diagnostics.go` (new, ~150 lines)

This file introduces `EnhanceAPIError(resourceType string, err error) string` which:
1. Checks if the error contains common field names (frequency, regions, protocol, status_code, etc.).
2. For each matched field, appends the valid values from `client.Allowed*` constants.
3. Returns an enhanced error string with specific guidance.

Example output:
```
API error: Unprocessable Entity - field "check_frequency" is invalid.

Valid values for check_frequency: 10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400 (seconds).
```

**File**: `internal/provider/error_helpers.go`

Changes to existing functions:
- `newCreateError`, `newUpdateError`: Call `EnhanceAPIError` on the error before formatting the diagnostic message.
- `buildValidationErrorSteps`: Make the Monitor-specific block more detailed, adding valid values for all enum fields inline.

**Decision -- enhance at diagnostic layer, not client layer**: The user specification says "Do NOT change the client package API surface." Enhancing errors in the provider package respects this boundary. The client returns raw errors; the provider adds IaC-specific guidance.

### 4. Better MarkdownDescriptions with Examples and Constraints

**Problem**: Descriptions like "The name of the monitor" are not actionable. Users need to know what values are valid, what the defaults are, and see examples.

**Solution**: Update MarkdownDescription strings across all resource schemas to include constraints, defaults, and examples. This is a mechanical change to string literals only.

**Files changed** (description-only updates, no logic changes):

**`internal/provider/monitor_resource.go`** -- Schema section:
- `name`: `"The display name of the monitor. Must be 1-255 characters. Example: \"Production API\""`
- `url`: `"The URL to monitor. Must include protocol (http:// or https://). Example: \"https://api.example.com/health\""`
- `protocol`: Already good, keep as-is.
- `http_method`: Already good, keep as-is.
- `check_frequency`: Already good, keep as-is.
- `regions`: `"List of monitoring regions. Use the ` + "`hyperping_monitoring_locations`" + ` data source to discover available locations. Valid values: ` + "`london`" + `, ` + "`frankfurt`" + `, ` + "`singapore`" + `, ` + "`sydney`" + `, ` + "`tokyo`" + `, ` + "`virginia`" + `, ` + "`saopaulo`" + `, ` + "`bahrain`" + `."`
- `request_headers`: Add note "Only valid when protocol is `http`."
- `request_body`: `"HTTP request body for POST/PUT/PATCH requests. Only valid when protocol is ` + "`http`" + ` and http_method is POST, PUT, or PATCH."`
- `expected_status_code`: Already good.
- `follow_redirects`: Add "Only applies to HTTP monitors."
- `port`: `"TCP port number (1-65535). Required when protocol is ` + "`port`" + `. Example: ` + "`443`" + ` for HTTPS, ` + "`5432`" + ` for PostgreSQL."`
- `alerts_wait`: Already good.
- `required_keyword`: `"A string that must appear in the HTTP response body for the check to pass. Only valid when protocol is ` + "`http`" + `. Example: ` + "`\"OK\"`" + `."`
- `escalation_policy`: Already good.
- `project_uuid`: Already good.

**`internal/provider/healthcheck_resource.go`** -- Schema section:
- `cron`: `"Cron expression defining the expected schedule. Mutually exclusive with period_value/period_type. Example: ` + "`\"0 */6 * * *\"`" + ` (every 6 hours). Requires timezone to be set."`
- `timezone`: `"IANA timezone for the cron expression. Required when cron is set. Example: ` + "`\"America/New_York\"`" + `, ` + "`\"UTC\"`" + `."`
- `grace_period_value`: `"Buffer time before alerting after a missed ping. Combined with grace_period_type. Example: ` + "`5`" + ` with grace_period_type ` + "`\"minutes\"`" + ` means alert after 5 minutes of silence."`
- `grace_period_type`: Already has OneOf. Add example.

**`internal/provider/maintenance_resource.go`** -- Schema section:
- `start_date`: Already good.
- `end_date`: Already good.
- `notification_option`: `"When to notify subscribers about this maintenance. Valid values: ` + "`scheduled`" + ` (notify notification_minutes before start), ` + "`immediate`" + ` (notify right away). Defaults to ` + "`scheduled`" + `."`
- `notification_minutes`: `"Minutes before the maintenance window starts to send notifications. Only used when notification_option is ` + "`scheduled`" + `. Must be at least 1. Defaults to ` + "`60`" + `."`

**`internal/provider/incident_resource.go`** -- Schema section:
- `type`: Already good.
- `affected_components`: `"List of monitor UUIDs representing components affected by this incident. These are shown on the status page."`

**`internal/provider/outage_resource.go`** -- Schema section:
- `monitor_uuid`: `"The UUID of the monitor this outage is associated with. Must reference an existing monitor."`
- `status_code`: Add valid range info.

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
- No API call needed.

Register in `provider.go` DataSources:
- Add `NewMonitoringLocationsDataSource` to the `DataSources()` return list.

**`internal/provider/monitoring_locations_data_source_test.go`** (~80 lines)

Unit test:
- Verify all locations are returned.
- Verify `ids` list matches location `id` fields.
- Verify known regions (london, frankfurt, etc.) are present.

Acceptance test:
- `TestAccMonitoringLocationsDataSource_basic`: Uses `terraform-plugin-testing` to verify the data source reads successfully and returns expected attributes.

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

**Files changed**:

**`internal/provider/monitors_data_source.go`**:
- Add `Count types.Int64` and `IDs types.List` to `MonitorsDataSourceModel`.
- Add schema attributes:
  - `count`: `schema.Int64Attribute{Computed: true, MarkdownDescription: "Total number of monitors returned (after filtering)."}`
  - `ids`: `schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "List of monitor UUIDs. Convenient for for_each patterns."}`
- In `Read()`, after mapping monitors, set `config.Count = types.Int64Value(int64(len(filteredMonitors)))` and build `config.IDs` from the monitor UUIDs.

**`internal/provider/incidents_data_source.go`**:
- Same pattern: add `Count` and `IDs` to model, schema, and Read.

**`internal/provider/healthchecks_data_source.go`**:
- Same pattern.

**`internal/provider/maintenance_windows_data_source.go`**:
- Same pattern.

**`internal/provider/outages_data_source.go`**:
- Same pattern.

**`internal/provider/statuspages_data_source.go`**:
- Already has `Total types.Int64` (from API pagination). Add `IDs types.List` only. The `total` field serves as `count`.
- In `Read()`, build IDs list from status page UUIDs after mapping.

**`internal/provider/statuspage_subscribers_data_source.go`**:
- Already has `Total types.Int64` (from API pagination). Add `IDs types.List` only (subscriber IDs as strings).
- In `Read()`, build IDs list from subscriber IDs after mapping.

**Decision -- naming convention**: Using `count` and `ids` to match the convention from major providers (AWS, Azure, GCP all use these names on list data sources).

---

## File Change Summary

### New Files (4 source + 3 test)
| File | Lines (est.) | Purpose |
|------|-------------|---------|
| `internal/provider/monitor_validate_config.go` | ~120 | Cross-field validation for monitor protocol/field combos |
| `internal/provider/monitoring_locations_data_source.go` | ~180 | Static data source for monitoring regions |
| `internal/provider/error_diagnostics.go` | ~150 | API error enhancement with valid value lists |
| `internal/provider/monitor_validate_config_test.go` | ~200 | Unit tests for cross-field validation |
| `internal/provider/monitoring_locations_data_source_test.go` | ~80 | Unit + acceptance tests for locations data source |
| `internal/provider/error_diagnostics_test.go` | ~100 | Unit tests for error enhancement |

### Modified Files (~15)
| File | Nature of Change |
|------|-----------------|
| `internal/provider/monitor_resource.go` | Add ValidateConfig interface assertion, update MarkdownDescriptions |
| `internal/provider/maintenance_resource.go` | Add notification_option validator + stringvalidator import, update descriptions |
| `internal/provider/incident_resource.go` | Update MarkdownDescriptions |
| `internal/provider/healthcheck_resource.go` | Update MarkdownDescriptions |
| `internal/provider/outage_resource.go` | Update MarkdownDescriptions |
| `internal/provider/statuspage_resource_schema.go` | Add languages + default_language validators + listvalidator import |
| `internal/provider/provider.go` | Register MonitoringLocationsDataSource |
| `internal/provider/error_helpers.go` | Integrate EnhanceAPIError in error creation functions |
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
   - Test HTTP protocol accepts all HTTP fields.
   - Test that defaults (computed values) do NOT trigger cross-field errors.

2. **Error diagnostics** (`error_diagnostics_test.go`):
   - Test `EnhanceAPIError` with various API error strings.
   - Verify valid value lists appear in output.
   - Test with unknown field names (no crash, graceful passthrough).

3. **Monitoring locations** (`monitoring_locations_data_source_test.go`):
   - Test all 8 regions returned.
   - Test ids list matches.
   - Test metadata correctness (names, continents).

### Acceptance Tests (TF_ACC=1, mock server)

1. **Monitor cross-field validation**:
   - `TestAccMonitorResource_icmpRejectsHTTPFields`: Config with protocol=icmp and http_method=GET, expect plan error.
   - `TestAccMonitorResource_portRequiresPort`: Config with protocol=port and no port, expect plan error.
   - `TestAccMonitorResource_httpRejectsPort`: Config with protocol=http and port=443, expect plan error.

2. **Monitoring locations data source**:
   - `TestAccMonitoringLocationsDataSource_basic`: Read data source, verify count and structure.

3. **Bulk data source count/ids**:
   - Existing list data source tests already verify Read works. Add assertions for `count` and `ids` attributes in existing test functions.

---

## Execution Order

All changes are independent at the code level and can be implemented in parallel. However, for a clean commit history:

1. Create `error_diagnostics.go` + tests (foundation)
2. Create `monitor_validate_config.go` + update `monitor_resource.go` interface assertion + tests
3. Update all MarkdownDescriptions across resources
4. Add `notification_option` validator to maintenance
5. Create `monitoring_locations_data_source.go` + tests + register in provider.go
6. Add `count`/`ids` to all list data sources
7. Run `make lint` and `make test` to verify
8. Run `go generate ./...` to regenerate docs

---

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Cross-field validation breaks existing configs that pass HTTP fields for non-HTTP monitors | Medium | High | ValidateConfig only checks raw config (pre-default). Fields with schema defaults will be null in config, so they won't trigger errors. Only explicit user-provided values are flagged. |
| Static monitoring locations become stale | Low | Low | Region changes are rare (last change was months ago). Provider releases update the list. MarkdownDescription notes "as of provider version X." |
| count/ids on data sources change state shape | Low | Medium | These are new computed-only attributes. Existing configs that don't reference them are unaffected. No breaking change. |
| Validator on notification_option rejects previously accepted invalid values | Low | Medium | Only affects configs with typos. This is the correct behavior -- catching errors early. |

---

## Non-Goals (Explicitly Out of Scope)

- No changes to `internal/client/` package API surface
- No changes to existing resource CRUD logic
- No new API calls for existing resources
- No breaking changes to existing schemas (all additions are Optional/Computed or new data sources)
- No changes to migration CLI tools
- No changes to CI/CD pipeline
