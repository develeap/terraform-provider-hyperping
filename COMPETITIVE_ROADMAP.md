# Competitive Roadmap: Terraform Provider Hyperping

Competitive analysis of 7 providers: Better Uptime (32 res), Checkly (22 res), Datadog (126 res), Site24x7 (39 res), Uptime.com (60 res), StatusCake (6 res), UptimeRobot (3 res).

Constraint: No Hyperping server-side API changes. All improvements are provider-side only.

---

## Phase 1: Provider-Level Configuration

**Branch:** `feat/provider-config`
**Source:** Better Uptime, Uptime.com

### Problem

Provider only accepts `api_key` and `base_url`. Retry count (3), backoff (1s-30s), and timeout are hardcoded. Enterprise users behind proxies or on different API tiers cannot tune behavior without forking.

### Changes

- Add optional `max_retries` (int, default 3, range 0-10) to provider schema
- Add optional `retry_wait_min` (string/duration, default "1s") to provider schema
- Add optional `retry_wait_max` (string/duration, default "30s") to provider schema
- Add optional `request_timeout` (string/duration, default "60s") to provider schema
- Add optional `rate_limit` (float, default 0 = disabled) to provider schema
- Pass these values through to `client.NewClient()` via a config struct
- Update client constructor to accept configurable retry/timeout/rate values
- Add acceptance tests for provider config validation
- Update `docs/index.md` with new provider attributes

### Competitive Reference

```hcl
# Better Uptime pattern:
provider "better_uptime" {
  api_token       = var.token
  api_retry_max   = 4
  api_retry_wait_min = 10
  api_retry_wait_max = 300
  api_timeout     = 60
  api_rate_limit  = 8
}

# Uptime.com pattern:
provider "uptime" {
  token      = var.token
  rate_limit = 0.5
}
```

---

## Phase 2: User-Agent Versioning

**Branch:** `feat/user-agent-version`
**Source:** StatusCake, Better Uptime

### Problem

Our HTTP client sends a generic or missing User-Agent. Competitors send `terraform-provider-{name}/{version}` for API analytics and debugging.

### Changes

- Add provider version string to client (via Go build ldflags or provider metadata)
- Set User-Agent header to `terraform-provider-hyperping/{version}` in transport layer
- Ensure User-Agent is not overridable by request_headers validator (already blocked by `ReservedHeaderName`)

### Files

- `internal/client/client.go` or `internal/client/headers.go`
- `main.go` (version injection via ldflags)

---

## Phase 3: Monitoring Locations Data Source -- DONE (v1.5.0)

**Branch:** `feat/locations-datasource`
**Source:** StatusCake, Uptime.com
**Shipped:** v1.5.0 (PR #82)

### Delivered

- `hyperping_monitoring_locations` static data source with 8 regions
- Schema: `locations` (list with `id`, `name`, `continent`, `cloud_region`) + `ids` convenience list
- No `filter` block — users can filter with Terraform expressions (`[for loc in ... : loc.id if loc.continent == "Europe"]`)

### Example

```hcl
data "hyperping_monitoring_locations" "all" {}

resource "hyperping_monitor" "web" {
  name    = "Website"
  url     = "https://example.com"
  regions = data.hyperping_monitoring_locations.all.ids
}
```

---

## Phase 4: Schema Attribute Helpers (DRY)

**Branch:** `refactor/schema-helpers`
**Source:** Uptime.com, Checkly

### Problem

Schema definitions are repeated across 8 resources. Same `id`, `name`, validators, plan modifiers copy-pasted. Uptime.com uses `IDSchemaAttribute()`, `NameSchemaAttribute()` factory functions. Checkly uses `makeFrequencyAttributeSchema()`.

### Changes

- Create `internal/provider/schema_helpers.go` with reusable attribute builders:
  - `IDAttribute()` - computed string with UseStateForUnknown
  - `NameAttribute(maxLen int)` - required string with StringLength + NoControlCharacters
  - `UUIDAttribute(description string)` - optional string with UUIDFormat validator
  - `UUIDListAttribute(description string)` - optional list of UUIDs
  - `PausedAttribute()` - optional bool with default false
  - `CreatedAtAttribute()` - computed string with UseStateForUnknown
  - `ISO8601Attribute(name, description string)` - required string with ISO8601 validator
- Refactor existing resources to use helpers (no behavioral changes)
- Verify all existing tests still pass

### Benefit

Adding a new resource goes from ~50 lines of boilerplate schema to ~10.

---

## Phase 5: Cross-Field Schema Validators -- IN PROGRESS (v1.5.0 partial)

**Branch:** `feat/cross-field-validators`
**Source:** Better Uptime, StatusCake, Datadog
**Shipped (v1.5.0):** Monitor `ValidateConfig` — protocol-aware cross-field validation at plan time

### Delivered (v1.5.0, PR #82)

- `ValidateConfig()` on monitor resource via `resource.ResourceWithValidateConfig` interface
- ICMP/port protocols reject HTTP-only fields (`http_method`, `expected_status_code`, `follow_redirects`, `request_headers`, `request_body`, `required_keyword`)
- Port protocol requires `port` field; HTTP/ICMP reject `port` field
- Clear error messages with field name and suggested fix
- 513 lines of unit tests covering all protocol/field combinations

### Remaining

- Add `ConfigValidators()` to healthcheck resource:
  - `cron` conflicts with `period_value` + `period_type`
- Add `ConfigValidators()` to maintenance resource:
  - `end_date` must be after `start_date` (custom validator)
- Monitor: `dns_record_type` requires protocol == "dns" (if applicable)

### Competitive Reference

```go
// StatusCake pattern:
ExactlyOneOf: []string{"dns_check", "http_check", "icmp_check", "tcp_check"}

// Datadog pattern:
resp.Diagnostics.Append(resourcevalidator.ConflictsWith(
    path.MatchRoot("one_time_schedule"),
    path.MatchRoot("recurring_schedule"),
).ValidateResource(ctx, req)...)
```

---

## Phase 6: Enhanced Error Diagnostics -- IN PROGRESS (v1.5.0 partial)

**Branch:** `feat/error-diagnostics`
**Source:** Datadog, StatusCake
**Shipped (v1.5.0):** Valid value reference tables in error messages

### Delivered (v1.5.0, PR #82)

- `ValidValueReference()` function that appends formatted valid value tables to API error messages
- Covers Monitor (protocol, http_method, check_frequency, expected_status_code, regions, alerts_wait), Maintenance Window (notification_option), and Incident (type) resources
- `error_diagnostics.go` + `error_diagnostics_test.go` (119 lines of tests)
- Inline schema descriptions now show valid values

### Remaining

- Add doc URL constants per resource (e.g., `https://docs.hyperping.io/api#monitors`)
- Append doc URL as diagnostic detail in error messages
- Add `DiagnosticWithDocLink()` helper
- Show field path + received value + expected values in validation error formatting

### Example (current)

```
Error: failed to create monitor

  API error: invalid check_frequency

  Quick Reference (valid values):
    protocol:             http, icmp, port
    http_method:          GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS
    check_frequency:      30, 60, 120, 300, 600, 900, 1800
    ...
```

---

## Phase 7: PlanValuePreserver Interface

**Branch:** `refactor/plan-value-preserver`
**Source:** Uptime.com

### Problem

Write-only fields (`required_keyword`, `escalation_policy`, `text` on incidents) require manual state preservation logic scattered across Read functions. Uptime.com has a clean `PlanValuePreserver` interface.

### Changes

- Define `PlanValuePreserver` interface in `internal/provider/`:
  ```go
  type PlanValuePreserver interface {
      PreserveFromPlan(ctx context.Context, plan, state tfsdk.Plan) diag.Diagnostics
  }
  ```
- Implement for monitor resource (preserves `required_keyword`, `escalation_policy`, `request_body`, `request_headers`)
- Implement for incident resource (preserves `text`)
- Implement for healthcheck resource (preserves `escalation_policy`)
- Remove duplicated preservation logic from individual Read functions
- Add unit tests for preservation behavior

### Benefit

Centralizes write-only field handling. Adding new write-only fields requires one line instead of scattered Read/Update logic.

---

## Phase 8: Bulk Data Source Enhancements -- DONE (v1.5.0)

**Branch:** `feat/datasource-enhancements`
**Source:** Datadog, Uptime.com
**Shipped:** v1.5.0 (PR #82)

### Delivered

- `total` (int64) and `ids` (list of strings) on all 7 list data sources: monitors, incidents, maintenance_windows, outages, healthchecks, statuspages, statuspage_subscribers
- Used `total` instead of `count` because `count` is a reserved Terraform meta-argument
- Acceptance and unit tests for all data sources

### Example

```hcl
data "hyperping_monitors" "production" {
  filter {
    name_regex = "^prod-"
  }
}

output "production_count" {
  value = data.hyperping_monitors.production.total
}

# Use IDs in other resources
resource "hyperping_maintenance" "deploy" {
  monitors = data.hyperping_monitors.production.ids
}
```

---

## Phase 9: Maintenance Window Recurrence

**Branch:** `feat/maintenance-recurrence`
**Source:** Checkly, StatusCake, Uptime.com, Datadog

### Problem

Our `hyperping_maintenance` has only `start_date` and `end_date`. Competitors support recurring schedules (daily, weekly, monthly). Deploy windows are typically weekly.

### Changes

- Investigate if Hyperping API accepts recurrence parameters (check API docs/responses)
- If API supports recurrence:
  - Add `repeat_interval` field (string: "never", "daily", "weekly", "monthly")
  - Add `repeat_ends_at` field (optional ISO8601 datetime)
  - Add validators for recurrence fields
- If API does not support recurrence:
  - Document limitation clearly
  - Add example showing how to use `count` or `for_each` with `timeadd()` for manual recurrence
  - Consider client-side computed recurrence as future enhancement
- Add acceptance tests for new fields

---

## Phase 10: Status Page Enhancement

**Branch:** `feat/statuspage-enhancements`
**Source:** Better Uptime, Uptime.com

### Problem

Our status page resource is functional but competitors offer richer features: subscription domain allow/block lists, component ordering, custom metrics display, incident templates.

### Changes

- Add `position` optional field to status page sections for explicit ordering
- Add computed `component_count` to status page for quick visibility
- Investigate API support for:
  - Subscription domain filtering (allow/block email domains)
  - Custom CSS/branding fields beyond current settings
  - History period configuration (Better Uptime: 7-365 days)
- Enhance status page data source with `search` filter support
- Add examples for complex multi-section status pages

---

## Phase 11: Monitor Escalation Config Enhancement

**Branch:** `feat/escalation-enhancements`
**Source:** Uptime.com, Better Uptime, Checkly

### Problem

Our monitor has `alerts_wait` (delay before alerting) and `escalation_policy` (UUID reference). Competitors expose richer escalation configuration:
- Multi-level escalation with different contacts per level (Uptime.com)
- `confirmation_period` / `recovery_period` (Better Uptime)
- `num_repeats` for re-alerting (Uptime.com)
- Run-based vs time-based escalation types (Checkly)

### Changes

- Investigate Hyperping API escalation_policy response to find unexposed fields
- Add computed fields to monitor resource for escalation policy details:
  - `escalation_policy_name` (computed) - human-readable name from API
  - `escalation_levels` (computed, list) - levels within referenced policy
- Add `recovery_period` field to monitor if API supports it (seconds before auto-resolving)
- Add `confirmation_period` field if API supports it (seconds to confirm down status)
- If API returns escalation details on monitor GET, expose as computed nested block
- Document escalation patterns with examples showing policy + monitor relationship
- Add acceptance tests for new escalation fields

### Competitive Reference

```hcl
# Uptime.com escalation pattern:
resource "uptime_check_escalations" "web" {
  check_id = uptime_check_http.web.id
  escalation {
    wait_time      = 300    # 5 minutes
    num_repeats    = 3      # re-alert 3 times
    contact_groups = ["ops-team"]
  }
  escalation {
    wait_time      = 900    # 15 minutes
    num_repeats    = 0      # infinite until resolved
    contact_groups = ["management"]
  }
}

# Better Uptime pattern:
resource "betteruptime_monitor" "web" {
  url                = "https://example.com"
  confirmation_period = 180   # 3 min to confirm down
  recovery_period    = 180    # 3 min up before resolved
  team_wait          = 60     # escalate after 60s
}
```

---

## Phase 12: SLA/Uptime Report Enhancement

**Branch:** `feat/sla-report-enhancement`
**Source:** Uptime.com, Datadog

### Problem

We already have `hyperping_monitor_report` and `hyperping_monitor_reports` data sources with `uptime_percentage`, but this is underdocumented. Competitors (Uptime.com) have dedicated SLA report resources and SLA target attributes on checks. We can better leverage what we have.

### Changes

- Add SLA-focused computed fields to monitor report data source:
  - `sla_met` (computed bool) - true if uptime_percentage >= target (if configurable)
  - `downtime_human` (computed string) - human-readable downtime (e.g., "2h 15m")
  - `availability_nines` (computed string) - "99.9%" style representation
- Add `docs/guides/sla-tracking.md` guide showing:
  - How to use monitor reports for SLA tracking
  - Example: monthly uptime report with thresholds
  - Example: multi-monitor SLA dashboard via outputs
  - Comparison with competitor SLA features
- Add examples showing SLA monitoring patterns:

```hcl
data "hyperping_monitor_report" "api" {
  uuid = hyperping_monitor.api.id
  from = "2024-01-01T00:00:00Z"
  to   = "2024-01-31T23:59:59Z"
}

output "api_sla" {
  value = {
    uptime     = data.hyperping_monitor_report.api.uptime_percentage
    checks     = data.hyperping_monitor_report.api.checks_total
    downtime   = data.hyperping_monitor_report.api.downtime_total
  }
}
```

### Competitive Position

Only Uptime.com has dedicated `sla_report` resources. We can match the use case with better documentation and computed convenience fields on our existing data sources.

---

## Phase 13: Validation Library Documentation

**Branch:** `docs/validation-advantages`
**Source:** Competitive analysis results

### Problem

Our provider has the strongest security posture AND the most comprehensive validation library of all 8 analyzed, but neither is documented or marketed.

### Changes

- Add `docs/guides/security.md` documenting security features:
  - Error message sanitization (API keys redacted from all error output)
  - Circuit breaker (prevents cascading failures during outages)
  - TLS enforcement (non-localhost targets must use HTTPS)
  - Control character injection prevention (CR/LF/NULL blocked in headers)
  - Reserved header blocking (Authorization, Host, Cookie blocked in custom headers)
  - Response body size limit (10MB OOM protection)
  - Connection pool limits (per-host idle/total limits)
  - Resource ID validation (path traversal prevention)
  - API key format validation (sk_ prefix enforcement)
  - Base URL domain whitelist (*.hyperping.io only)
- Add `docs/guides/validation.md` documenting our custom validator library:
  - `StatusCodePattern()` - accepts "200", "2xx", "1xx-3xx" (unique to us)
  - `CronExpression()` - validates cron syntax (only us and Checkly)
  - `NoControlCharacters()` - prevents header injection (nobody else does this)
  - `ReservedHeaderName()` - blocks dangerous headers (nobody else)
  - `AlertsWait()` - domain-specific validation
  - `ISO8601()` - date format validation
  - `Timezone()` - IANA timezone validation
  - `PortRange()` - port number validation
  - `HexColor()` - color format validation
  - `EmailFormat()` - email validation
  - `RequiredWhenValueIs()` - conditional field requirements
- Add security and validation comparison tables vs competitors
- Reference both guides from main `docs/index.md`

### Competitive Position

| Security Feature | Us | Better Uptime | Checkly | Datadog | Others |
|-----------------|-----|---------------|---------|---------|--------|
| Error sanitization | Yes | No | No | No | No |
| Circuit breaker | Yes | No | No | No | No |
| TLS enforcement | Yes | No | No | No | No |
| Header injection prevention | Yes | No | No | No | No |
| Response size limit | Yes | No | No | No | No |
| Connection pool limits | Yes | No | No | No | No |

| Validator | Us | Checkly | StatusCake | Datadog | Others |
|-----------|-----|---------|------------|---------|--------|
| Status code patterns (2xx) | Yes | No | No | No | No |
| Cron expression | Yes | No | No | No | No |
| Control char blocking | Yes | No | No | No | No |
| Reserved header blocking | Yes | No | No | No | No |
| Conditional required fields | Yes | No | No | No | No |
| ISO8601 datetime | Yes | No | Yes (RFC3339) | Yes | Partial |

---

## Phase 14: Import Generator Data Source

**Branch:** `feat/import-generator`
**Source:** Unique competitive advantage

### Problem

We have migration CLI tools (betterstack, uptimerobot, pingdom) which are unique among all competitors. But there is no Terraform-native way to discover existing resources and generate import commands for adopting the provider on an existing Hyperping account.

### Changes

- Add `hyperping_import_commands` data source that generates import commands for all existing resources:
  - `resource_type` (optional, filter by type: "monitor", "healthcheck", "statuspage", etc.)
  - `commands` (computed, list of objects):
    - `resource_address` - suggested Terraform address (e.g., `hyperping_monitor.my_api_check`)
    - `import_id` - the UUID to import
    - `import_command` - full `terraform import` command string
    - `resource_name` - original name from Hyperping
- Generate safe Terraform resource names from Hyperping display names (slugify)
- Add documentation with adoption workflow example

### Example

```hcl
data "hyperping_import_commands" "all" {}

output "import_script" {
  value = join("\n", data.hyperping_import_commands.all.commands[*].import_command)
}

# Output:
# terraform import hyperping_monitor.my_api_check abc-123-def
# terraform import hyperping_monitor.website_prod xyz-456-ghi
# terraform import hyperping_healthcheck.cron_job mno-789-pqr
# terraform import hyperping_statuspage.public_page stu-012-vwx
```

### Competitive Position

No other provider offers this. Combined with our existing migration CLI tools, this makes Hyperping the easiest provider to adopt.

---

## Phase 15: Cassette-Based Testing

**Branch:** `feat/cassette-testing`
**Source:** Datadog

### Problem

Our tests use mock HTTP servers which is good for isolation but can drift from real API behavior over time. Datadog uses cassette recording (record real API responses, replay in CI). This gives the best of both worlds: fast CI without API keys, plus real response validation.

### Changes

- Add `go-vcr` (or `go-cassette`) dependency for HTTP recording/replay
- Create test helper: `NewCassetteClient(t, cassetteName)` that:
  - In RECORD mode (`RECORD=true`): hits real API, saves responses to `testdata/cassettes/`
  - In REPLAY mode (default): replays saved responses (no API key needed)
- Add cassette files for core CRUD operations:
  - `testdata/cassettes/monitor_create.yaml`
  - `testdata/cassettes/monitor_read.yaml`
  - `testdata/cassettes/monitor_update.yaml`
  - `testdata/cassettes/monitor_delete.yaml`
  - Same for healthcheck, statuspage, incident, maintenance, outage
- Add `docker-compose.yaml` for acceptance test environment:
  - Go container with mounted source
  - Pre-configured env vars (TF_ACC=1)
  - Rate-limited test configuration
- Add `make docker-test` and `make record-cassettes` targets
- Add cassette sanitizer to redact API keys from recorded responses
- Document testing approach in CONTRIBUTING.md

### Competitive Reference

```go
// Datadog cassette pattern:
func TestAccMonitor_Basic(t *testing.T) {
    // RECORD=false: replays cassettes (default, fast, CI-friendly)
    // RECORD=true:  hits real API, records responses
    // RECORD=none:  live API only (debugging)
    ctx, providers, accProviders := testAccFrameworkMuxProviders(ctx, t)
    // ... test steps using recorded responses
}
```

### Benefit

- CI runs without API keys (replay mode)
- Catches API drift when cassettes are re-recorded periodically
- Faster test execution (no network calls in replay mode)
- Real response shapes ensure mock accuracy

---

## Execution Order

| Phase | Branch | Priority | Effort | Status |
|-------|--------|----------|--------|--------|
| 1 | `feat/provider-config` | P0 | 2-3 days | TODO |
| 2 | `feat/user-agent-version` | P0 | 30 min | TODO |
| 3 | `feat/locations-datasource` | P0 | -- | **DONE** (v1.5.0) |
| 4 | `refactor/schema-helpers` | P1 | 2-3 days | TODO |
| 5 | `feat/cross-field-validators` | P1 | ~1 day | **IN PROGRESS** — monitor done, healthcheck/maintenance remaining |
| 6 | `feat/error-diagnostics` | P1 | ~0.5 day | **IN PROGRESS** — valid value tables done, doc URLs remaining |
| 7 | `refactor/plan-value-preserver` | P2 | 1-2 days | TODO (depends on Phase 4) |
| 8 | `feat/datasource-enhancements` | P2 | -- | **DONE** (v1.5.0) |
| 9 | `feat/maintenance-recurrence` | P2 | 1-2 days | TODO |
| 10 | `feat/statuspage-enhancements` | P2 | 1-2 days | TODO |
| 11 | `feat/escalation-enhancements` | P2 | 1-2 days | TODO |
| 12 | `feat/sla-report-enhancement` | P2 | 1-2 days | TODO |
| 13 | `docs/validation-advantages` | P3 | 1 day | TODO |
| 14 | `feat/import-generator` | P3 | 2-3 days | TODO |
| 15 | `feat/cassette-testing` | P3 | 2-3 days | TODO |

Phases 1, 2 are the remaining P0 items (Phase 3 done).
Phases 4, 5 (remaining), 6 (remaining) can run in parallel.
Phase 7 depends on Phase 4 (uses schema helpers).
Phases 9, 10, 11, 12 are independent (Phase 8 done).
Phases 13, 14, 15 are independent.

---

## Out of Scope (Requires Server API Changes)

These features were identified in competitors but require Hyperping API additions:
- New monitor types (DNS, TCP content matching, SMTP, IMAP, SSH, FTP, NTP, WHOIS)
- New integration resources (Slack, PagerDuty, Opsgenie, webhook destinations)
- SLO/SLA target resources (Datadog pattern)
- On-call calendar/rotation resources (Better Uptime)
- Contact group management resources
- Severity/priority level resources
- Browser/transaction check support (Checkly, Better Uptime)
- Real User Monitoring (Uptime.com)
- Multi-account/subaccount support (Uptime.com)
- Dashboard/reporting resources (Uptime.com, Datadog)

These should be tracked as server-side feature requests separately.
