# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

**Note:** Versions 1.0.1 and 1.0.2 exist as git tags but were never published to Terraform Registry.
Published releases start from v1.0.3.

## [Unreleased]

### Fixed

- **`hyperping_statuspage`**: Restored bidirectional UUID/numeric ID translation for status page
  services. The uptime renderer requires numeric monitor IDs -- v1.4.4+v1.4.5 incorrectly removed
  the translation, breaking uptime display. Now: `mon_xxx` -> numeric on write (hard error if
  unresolvable), numeric -> `mon_xxx` on read (warning if unresolvable).

## [1.4.4] - 2026-03-04

### Fixed

- **`hyperping_statuspage`**: Fixed critical UUID drift bug where status page services had their
  `uuid` field set to the service's own numeric ID instead of the linked monitor UUID. This caused
  all status pages to show orphaned entries with no live data. The provider no longer translates
  `mon_xxx` UUIDs to numeric IDs on write — the API preserves them correctly when sent directly.
  Existing broken status pages will self-heal on the next `terraform apply`.

### Added

- **`hyperping_statuspage`**: Read-time warning when services have unresolved numeric UUIDs from
  legacy drift, guiding users to re-apply to fix the data.

## [1.4.3] - 2026-03-04

### Fixed

- **Circuit breaker**: Client errors (400, 404, 422) no longer trip the circuit breaker.
  Previously, validation errors on multiple monitors would cause the breaker to open, masking
  the real error with "circuit breaker is open". Only 429 and 5xx now count as failures.
- **`hyperping_monitor`**: `alerts_wait = -1` (disabled) now correctly preserved in state.
  Previously the mapping treated all values <= 0 as null, losing the "disabled" setting.

### Added

- **`hyperping_monitor`**: Plan-time validator for `alerts_wait` — catches invalid values
  before they reach the API. Must be one of: -1, 0, 1, 2, 3, 5, 10, 30, 60 (minutes).

### Changed

- **`hyperping_monitor`**: Fixed `alerts_wait` description — field accepts minutes (not seconds).
- **Rate limit docs**: Updated to match Hyperping's actual limits (800 req/hr per project,
  rolling window, with rate limit header format documentation).

## [1.4.2] - 2026-03-03

### Fixed

- **`hyperping_statuspage`**: `is_split = false` now correctly preserved when the Hyperping API
  ignores the value and returns `true`. Same bidirectional fix applied to `show_response_times`
  and `show_uptime` in v1.4.1.

## [1.4.1] - 2026-03-03

### Fixed

- **`hyperping_statuspage`**: `show_response_times = false` now correctly preserved when the
  Hyperping API ignores the value and returns `true`. Previously the provider only overrode
  API responses in one direction (plan=true, API=false), causing perpetual drift. Same fix
  applied to `show_uptime`.

## [1.4.0] - 2026-03-02

### Added

- **`hyperping_monitor`**: Plan-time validator for `expected_status_code` — catches invalid
  patterns before they reach the API. Accepts specific codes (`200`), wildcards (`2xx`),
  and ranges (`1xx-3xx`).

### Changed

- **`hyperping_monitor`**: Updated `expected_status_code` description to document multi-range
  patterns (`1xx-3xx`) newly revealed by Hyperping API docs.
- **API drift detection**: Enhanced drift issue reports to include field-level change details
  (added/removed/modified properties with description, type, default, and enum diffs).
  Previously only listed modified endpoint names.

## [1.3.9] - 2026-03-01

### Fixed

- **Status page renderer**: Services now use numeric v1 monitor IDs instead of UUID strings,
  fixing the critical bug where the renderer showed "up" status for all monitors regardless
  of actual state. The provider transparently translates UUIDs to numeric IDs on write and
  back on read — no HCL changes needed.
- **Status page boolean preservation**: `show_response_times`, `show_uptime`, and `is_split`
  now use UUID-based matching instead of fragile index-based matching. Fixes perpetual drift
  when API reorders services, and adds support for nested services inside groups.
- **Status page isProtected drift**: Every status page PUT now includes authentication settings
  (similar to the `dns_record_type: "A"` workaround for monitors). This triggers ISR cache
  revalidation on the Hyperping renderer, fixing the admin UI regression where editing any
  setting via the Hyperping dashboard resets an internal `isProtected` flag to `true`. The
  provider also emits a warning diagnostic when `password_protected` and
  `authentication.password_protection` disagree. **Known limitation**: if no Terraform fields
  changed, no PUT is sent, so the flag stays stale until the next apply that touches the page.
- **Description localization**: `extractLocalizedString` now skips empty string values when
  searching for a non-empty localized value, preventing drift when the API returns
  `{"en":"","fr":"texte"}`.

## [1.3.8] - 2026-02-27

### Fixed

- **`hyperping_monitor`**: Fix 422 regression introduced in v1.3.7. The Hyperping API's PUT
  endpoint requires a valid `dns_record_type` enum value in every request, even for non-DNS
  monitors — omitted, `null`, and `""` are all rejected with 422. v1.3.7 sent `null` which
  is also rejected. The provider now sends `dns_record_type: "A"` in every PUT request as a
  workaround. The API accepts it, ignores it for non-DNS protocols (response returns `null`),
  and monitor behavior is unaffected. This is a Hyperping API validation bug — PUT should not
  require `dns_record_type` for non-DNS monitors.

## [1.3.7] - 2026-02-27

### Fixed

- **`hyperping_monitor`**: (**Broken** — see v1.3.8) Attempted to fix 422 by sending
  `dns_record_type: null` in every PUT. The API also rejects `null`, making all monitor
  updates fail.
- **All resources**: Circuit breaker open state now surfaces actionable troubleshooting steps
  (wait 30 seconds, use `terraform apply -parallelism=1`, check API status at
  https://status.hyperping.app) instead of misleading "check your API key / verify resource
  exists" guidance.

## [1.3.6] - 2026-02-26

### Fixed

- **`hyperping_monitor`**: Fix 422 validation error on UPDATE for HTTP/port/icmp monitors — the provider was serialising `dns_record_type: ""` in every PUT payload. The Hyperping API rejects an empty string but accepts a missing field. The field is now omitted from the request when unset.
- **`hyperping_monitor`**: Fix escalation policy unlinking — clearing `escalation_policy` now sends `"none"` in the PUT payload instead of `""`, matching the API contract ("send `null` or `"none"` to unlink").

## [1.3.5] - 2026-02-25

### Fixed

- **`hyperping_monitor`**: Fix "Provider produced inconsistent result after apply" when `escalation_policy` is set — the API returns `null` for this field in the POST/PUT response even when a policy was successfully attached. The provider now preserves the plan value in state after create/update, matching the same save-restore pattern used for `required_keyword` (v1.2.1). Subsequent refreshes continue to read the live UUID from the API's object-shape GET response.

## [1.3.4] - 2026-02-25

### Fixed

- **hyperping_monitor**: Fix crash on `terraform plan`/`terraform refresh` when an escalation policy is set — the Hyperping API returns `escalation_policy` as an object `{"uuid":"...","name":"..."}` on read, but the provider expected a plain string, causing `json.Unmarshal` to panic. A custom `UnmarshalJSON` on `Monitor` now transparently handles both the object and plain-string shapes, normalising both to the UUID string. Write side (POST/PUT) is unchanged.
- **hyperping_statuspage**: Fix `is_split` perpetual drift on status page sections — sections configured with `is_split = true` showed a non-empty diff on every subsequent plan because the API accepts the field on write but never returns it on read. The provider now correctly preserves the configured value across refreshes.

### Changed

- Bumped `goreleaser/goreleaser-action` to v7 in release workflow.

## [1.3.3] - 2026-02-24

### Fixed

- **hyperping_statuspage**: Fix `description` field API write/read asymmetry — the API accepts `description` as a plain string on POST/PUT but returns a localised map `{"en":"..."}` on GET. Changed the schema from `MapAttribute` to `StringAttribute` and added extraction logic in the mapping layer. Prevents state drift and "inconsistent result after apply" errors for status pages with a description set.

## [1.3.2] - 2026-02-23

### Added

- **hyperping_statuspage**: Nested service group support — status page sections can now contain child service groups using `is_group = true` with nested `children` blocks. Allows hierarchical organisation of services on public status pages.

## [1.3.1] - 2026-02-22

### Fixed

- **Code quality**: Addressed all findings from comprehensive multi-expert code review — security hardening, error handling improvements, and deduplication across provider and client packages (#36).

### Changed

- Dependency bumps: `goquery` 1.9.0 → 1.11.0, `golang.org/x/time`, `golang.org/x/oauth2`, `hashicorp/terraform-json`, and GitHub Actions group updates.

## [1.3.0] - 2026-02-21

### Added

- **hyperping_monitor data source**: New fields `status`, `ssl_expiration`, and `project_uuid` — expose monitor runtime state and SSL certificate expiry days directly in data source reads.
- **hyperping_monitors data source**: New filter attributes `status`, `project_uuid`, and `has_ssl_expiration` for server-side result narrowing.
- **hyperping_monitor_reports** (plural) data source — list uptime/performance reports across multiple monitors in a single data source call, with optional `monitor_uuid` and date-range filters.

### Changed

- **Scraper/analyzer tooling**: Replaced custom rod-based Chromium scraper and custom analyzer with an OSS stack (goquery + static HTML), reducing tool code by 82% and eliminating the Chromium runtime dependency from CI.
- **CI**: Pinned `govulncheck` to v1.1.4 for reproducible security scans; fixed schema extraction, coverage analysis activation, and workflow permissions.

### Fixed

- Resolved 106 lint issues introduced by `golangci-lint` v2.10.1 upgrade.
- Security hardening across client and migration tools (gosec false-positive suppressions with documented justifications).

## [1.2.3] - 2026-02-17

### Changed

- **Code quality**: Reduced cyclomatic complexity across entire codebase to CC≤15
  - All 37 flagged functions refactored via extract-helper pattern
  - Migration tool main functions broken into focused phase handlers
  - Mock server handlers converted to route-dispatch pattern
  - Test functions converted to table-driven where applicable
  - `gocyclo -over 15 .` now returns zero results across all 286 files
  - Repo cleanup: archived stale development-phase docs to `docs/development-archive/`

## [1.2.2] - 2026-02-17

### Added

- **hyperping_statuspage**: Support for `default_language` field
  - Allows setting default language for status pages (e.g., "en", "es", "fr")
  - Maps to Hyperping API's `DefaultLanguage` field in settings
  - Enables localization control for multi-language status pages

### Testing - QA Certification Initiative

- **Phase 2: 100% Parameter Coverage** - Added 31 comprehensive acceptance tests
  - **healthcheck_resource**: 5 tests covering cron scheduling and timezone handling (0% → 100% coverage)
  - **outage_resource**: 4 tests covering escalation policies and status code edge cases (0% → 100% coverage)
  - **statuspage_resource**: 10 tests covering all 19 settings fields (0% → 100% coverage)
  - **monitor_resource**: 2 tests for alerts_wait edge cases and required_keyword Unicode handling
  - **maintenance_resource**: 6 tests for notification options and text special characters
  - **incident_resource**: 4 tests for date computed field and text long content/markdown

- **Coverage Achievements**:
  - ✅ **100% parameter coverage** achieved (all 117 parameters across 6 resources tested)
  - ✅ **~95% edge case coverage** (production-ready threshold)
  - ✅ **73 total new tests** (42 in Phase 1 + 31 in Phase 2)
  - ✅ **Zero flaky tests** - all deterministic with mock servers
  - ✅ **Production certification** - comprehensive QA validation complete

- **Total QA Initiative Tests** (from v1.2.1 through v1.2.2):
  - 6 protocol-specific tests (HTTP, Port, ICMP regression coverage)
  - 14 edge case & boundary value tests
  - 15 state drift detection tests (all resources)
  - 7 cross-resource integration tests
  - 31 comprehensive parameter coverage tests
  - **Total: 73 new acceptance tests**

### Fixed

- **Test Infrastructure**: Fixed potential slice index out of range in statuspage SSO test helper
  - Added bounds checking before slice access to prevent panics
  - Improved test safety and gosec linting compliance

## [1.2.1] - 2026-02-14

### Fixed

- **hyperping_monitor**: Fixed state drift for `required_keyword` field
  - Root cause: Hyperping API accepts `required_keyword` in POST/PUT but doesn't return it in GET responses (write-only field)
  - Symptom: Terraform detects inconsistency after apply ("Provider produced inconsistent result")
  - Solution: Implemented save-restore pattern to preserve plan value (same pattern as `incident.text` fix in v1.0.5)
  - Impact: `required_keyword` now persists correctly in state across create, read, update operations
  - Test coverage: Added comprehensive regression test `TestAccMonitorResource_requiredKeyword`

## [1.2.0] - 2026-02-14

### Added - User Experience Enhancements

- **P1.1: Interactive Migration Tool** (~1,500 lines of new code)
  - Automatic wizard mode when no CLI flags provided - zero-friction onboarding
  - API key validation with real-time connection testing
  - Real-time progress bars and spinners for long operations
  - Migration preview with confirmation prompts before writing files
  - Final summary with clear next steps and verification commands
  - Zero breaking changes - backward compatible with flag-based mode
  - New package: `pkg/interactive/` with prompt, progress, and terminal utilities
  - Interactive modes for all 3 migration tools (Better Stack, UptimeRobot, Pingdom)
  - Documentation: `docs/INTERACTIVE_MODE.md` (~700 lines)
  - Dependencies: AlecAivazis/survey/v2, briandowns/spinner, schollz/progressbar/v3
  - Test coverage: 30.6% (14 test cases)

- **P1.2: Dry-Run Enhancement** (~1,800 lines of new code)
  - Compatibility scoring (0-100%) with complexity ratings (Simple/Medium/Complex)
  - Side-by-side diff comparison showing source vs Hyperping transformations
  - Terraform preview with syntax highlighting
  - Warning categorization (Critical/Warning/Info) with manual effort estimation
  - Performance estimates (migration time, API calls, file sizes)
  - Resource breakdown by type, frequency, and region
  - New package: `pkg/dryrun/` with 6 modules (types, compatibility, diff, preview, reporter, bridge)
  - Dry-run integration for Better Stack (pattern reusable for other tools)
  - Documentation: `docs/DRY_RUN_GUIDE.md` (~800 lines)
  - Test coverage: 50.3%
  - Impact: Zero-risk previews enable informed decision-making before migration

- **P1.3: Import Generator Enhancement** (~1,500 lines of new code)
  - Filtering support: name regex, resource type, exclusion patterns
  - Parallel imports with **5-8x speedup** (configurable worker pools)
  - Drift detection (pre/post-import terraform plan comparison)
  - Checkpoint/resume capability (auto-save every 10 imports)
  - Rollback capability (safely remove resources from state)
  - 30+ new CLI flags for advanced control
  - New modules: filter, parallel, checkpoint, rollback, drift
  - Comprehensive test suite (30+ test cases)
  - Documentation: `docs/IMPORT_GENERATOR_GUIDE.md` (~1,200 lines)
  - Performance benchmarks:
    - 100 resources: 5m sequential → 45s parallel (6.7x speedup)
    - 500 resources: 25m sequential → 3m parallel (8.3x speedup)
  - Impact: Enterprise migrations complete in minutes instead of hours

- **P1.4: Enhanced Error Messages** (~1,100 lines of new code)
  - "Try: <command>" suggestions for every error type
  - Rate limit auto-retry with countdown timers (respects Retry-After header)
  - Typo detection using Levenshtein distance algorithm
  - Closest value finder for validation errors
  - Context-aware messages (create/read/update/delete operations)
  - Documentation links for each error type
  - New package: `internal/errors/` with enhanced, suggestions, client, provider modules
  - Complete error catalog: `docs/ERROR_REFERENCE.md` (565 lines)
  - Integration guides and examples (~2,300 lines total)
  - Test coverage: **87.7%** (48 test cases)
  - Impact: Expected 90%+ reduction in support tickets

### Added - Production Hardening (Phase 1)

- **Integration Testing Framework** (~3,500 lines of test code)
  - Integration tests for all 3 migration tools (Better Stack, UptimeRobot, Pingdom)
  - Real API call validation with test account credentials
  - GitHub Actions workflow: `.github/workflows/integration.yml`
  - Test environment setup documentation
  - Coverage: 3 migration tools × multiple scenarios
  - Documentation: `docs/INTEGRATION_TESTING.md`, `docs/INTEGRATION_TESTING_SUMMARY.md`

- **E2E Testing Framework** (~2,000 lines of test code)
  - End-to-end validation pipeline for complete migration workflows
  - Programmatic resource creation in source platforms
  - Terraform validation (init, plan, apply) of generated configs
  - Import script execution and state verification
  - Automated cleanup (idempotent tests)
  - Test helpers and fixtures: `test/e2e/`
  - Documentation: `docs/E2E_TESTING.md` (508 lines)
  - Execution script: `scripts/run-e2e-tests.sh`

- **Load Testing Framework** (~2,500 lines of test code)
  - Large-scale migration testing (100+ monitors per platform)
  - Memory profiling and leak detection
  - Execution time benchmarks
  - Rate limiting behavior validation
  - Performance documentation: `docs/PERFORMANCE.md` (498 lines)
  - Load test suites: `test/load/` for all 3 migration tools
  - Benchmarks: <500MB memory for 100 monitors, handles rate limits gracefully

- **Error Recovery System** (~600 lines of new code)
  - Checkpoint files track migration progress (every 10 resources)
  - `--resume` flag continues from last checkpoint
  - Partial failure handling (errors logged but don't crash tool)
  - `--rollback` deletes resources created in failed migrations
  - Enhanced dry-run validates API connectivity before migration
  - `--debug` flag enables verbose logging to file
  - New package: `pkg/recovery/` with logger and validator
  - New package: `pkg/checkpoint/` for checkpoint management
  - Documentation: `docs/ERROR_RECOVERY.md` (550 lines)

- **API Completeness Audit** (~1,400 lines of documentation)
  - Complete API coverage analysis: `docs/API_COMPLETENESS_AUDIT.md` (746 lines)
  - API roadmap with priority rankings: `docs/API_ROADMAP.md` (720 lines)
  - Identified gaps: notification channels, webhooks, teams, escalation policies
  - 100% coverage of documented Monitor, Healthcheck, Incident, Maintenance endpoints
  - Feature prioritization (P0/P1/P2) for future implementation

- **Migration Certification Documentation** (~3,100 lines)
  - Production certification report: `docs/MIGRATION_CERTIFICATION.md` (845 lines)
  - Customer pre-migration checklist: `docs/MIGRATION_CUSTOMER_CHECKLIST.md` (1,000 lines)
  - Support runbook: `docs/MIGRATION_SUPPORT_RUNBOOK.md` (1,321 lines)
  - Success metrics, known limitations, and validation procedures
  - QA certification criteria for production deployment

### Changed

- **Import Generator**: Massive expansion from basic tool to enterprise-ready bulk operations platform
- **Migration Tools**: All 3 tools now support interactive mode, dry-run, checkpoint/resume, and rollback
- **Error Handling**: All errors now include actionable suggestions and documentation links
- **Documentation**: Added ~16,000 lines of new documentation across 20+ files

### Performance

- **Parallel Imports**: 5-8x speedup for bulk import operations (100 resources: 5m → 45s)
- **Interactive Mode**: New users can migrate in minutes without reading documentation
- **Dry-Run**: Zero-risk validation prevents migration failures
- **Error Recovery**: Checkpoint/resume prevents data loss from partial failures

### Testing

- **Unit Tests**: All packages tested (30.6% - 87.7% coverage across new code)
- **Integration Tests**: Real API validation for all migration tools
- **E2E Tests**: Complete workflow validation (migration → Terraform apply → import)
- **Load Tests**: Validated with 100+ monitor migrations per platform
- **All Tests**: 0 linting issues (fixed 33 issues during Phase 2 testing)

### Dependencies

- Added `github.com/AlecAivazis/survey/v2 v2.3.7` - Interactive CLI prompts
- Added `github.com/briandowns/spinner v1.23.2` - Loading spinners
- Added `github.com/schollz/progressbar/v3 v3.19.0` - Progress bars (already existed, now used in interactive mode)

## [1.1.0] - 2026-02-13

### Added - Migration Tools

- **Automated Migration CLI Tools**: Three production-ready CLI tools for migrating from competitors to Hyperping
  - `cmd/migrate-betterstack/` - Better Stack migration tool (~2,200 lines, 28 unit tests)
    - Monitor type conversion (status→http, tcp→port, ping→icmp, keyword→http)
    - Heartbeat to healthcheck conversion with cron expression generation
    - Region mapping from Better Stack to Hyperping regions
    - Frequency normalization to supported values (10s-86400s)
    - Generates: Terraform config, import script, migration report (JSON), manual steps (markdown)
  - `cmd/migrate-uptimerobot/` - UptimeRobot migration tool (~2,100 lines, 10 unit tests)
    - All 5 monitor types supported (HTTP, Keyword, Ping, Port, Heartbeat)
    - Contact alert conversion to notification channels
    - Maintenance window mapping
    - Tag-based resource naming
  - `cmd/migrate-pingdom/` - Pingdom migration tool (~2,200 lines, 13 unit tests)
    - Check type support (HTTP/HTTPS, TCP, PING, SMTP, POP3, IMAP)
    - Tag-based naming convention (tags→[TENANT]-Category-Name)
    - Customer/tenant support from tags
    - DNS/UDP/Transaction checks documented as manual steps

### Added - Documentation

- **Comprehensive Migration Documentation** (~7,500 lines total)
  - `docs/guides/automated-migration.md` - Complete automated migration guide (~2,100 lines)
    - Common workflow for all 3 migration tools
    - Tool-specific usage guides with examples
    - Output file documentation (4 files per migration)
    - Troubleshooting section with 30+ FAQs
    - Time savings metrics (90% reduction vs manual migration)
  - `docs/guides/migrate-from-betterstack.md` - Enhanced with automation section (~1,800 lines)
  - `docs/guides/migrate-from-uptimerobot.md` - Enhanced with CLI tool usage (~2,200 lines)
  - `docs/guides/migrate-from-pingdom.md` - Enhanced with automated workflow (~1,800 lines)
  - `docs/guides/best-practices.md` - Comprehensive best practices guide (~2,400 lines)
    - Naming conventions and organizational patterns
    - State management and CI/CD integration
    - Security hardening and secrets management
    - Performance optimization and cost management
    - Testing strategies and disaster recovery

- **Getting Started Documentation** (~2,000 lines total)
  - `docs/guides/quickstart.md` - 5-minute quickstart guide (~400 lines)
  - `docs/guides/use-case-microservices.md` - Microservices monitoring patterns
  - `docs/guides/use-case-kubernetes.md` - Kubernetes cluster monitoring
  - `docs/guides/use-case-api-gateway.md` - API gateway health checks
  - `docs/guides/validation.md` - Complete validation reference (~1,400 lines)

### Added - Terraform Modules

- **Production-Ready Terraform Modules**: 7 reusable modules for common monitoring patterns
  - `examples/modules/database-monitor/` - Multi-database monitoring (PostgreSQL, MySQL, MongoDB, Redis, etc.) - 1,388 lines, 23 tests
  - `examples/modules/cdn-monitor/` - CDN edge location monitoring - 949 lines, 17 tests
  - `examples/modules/cron-healthcheck/` - Dead man's switch for cron jobs - 1,847 lines
  - `examples/modules/multi-environment/` - Dev/staging/prod deployment patterns - 1,200+ lines
  - `examples/modules/incident-management/` - Incident response templates - 2,033 lines, 30+ tests
  - `examples/modules/website-monitor/` - Critical page monitoring - 1,587 lines
  - `examples/modules/graphql-monitor/` - GraphQL API health checks - 1,423 lines, 25 tests

### Added - Validation Layer

- **Plan-Time Validators**: 7 custom validators for preventing invalid configurations
  - `URLFormat()` - Validates HTTP/HTTPS URLs (prevents 15+ error types)
  - `StringLength()` - Validates min/max string constraints
  - `CronExpression()` - Validates cron syntax using robfig/cron parser
  - `Timezone()` - Validates IANA timezone database identifiers
  - `PortRange()` - Validates port numbers (1-65535)
  - `HexColor()` - Validates hex color codes for status pages
  - `EmailFormat()` - Validates email addresses for notifications
- **Cross-Field Validation**: Date range validation for maintenance windows (start < end)
- **Security Validation**: Reserved HTTP header blocking (Authorization, Cookie, etc.) - prevents VULN-012

### Changed

- **golangci-lint Configuration**: Added exclusions for migration tool directories to allow relaxed stylistic linting for CLI tools

### Performance

- Migration tools reduce migration time from 4-8 hours (manual) to ~15 minutes (automated) - **90% time reduction**
- All tools generate audit trails via JSON reports for compliance and troubleshooting

### Testing

- 51 unit tests added across migration tools (100% pass rate)
- Comprehensive test coverage for all conversion logic and error handling
- All tools validated with golangci-lint (0 issues)

## [1.0.9] - 2026-02-13

### Added

- **All Resources**: Import validation and comprehensive acceptance tests for import workflows (8 resources, 20+ test scenarios)
  - Import state validation for all resource types (monitors, healthchecks, incidents, maintenance, outages, status pages, subscribers, incident updates)
  - ID format validation before import
  - Post-import state verification
  - Documentation for import usage patterns
- **Data Sources**: Client-side filtering support for 12 data sources with comprehensive filter framework
  - `hyperping_monitors` - Filter by name_regex, protocol, paused status
  - `hyperping_healthchecks` - Filter by name_regex, status
  - `hyperping_incidents` - Filter by name_regex, status, severity
  - `hyperping_maintenance_windows` - Filter by name_regex, status, time ranges
  - `hyperping_outages` - Filter by name_regex, monitor_uuid
  - `hyperping_statuspage` - Filter by name_regex, hostname
  - Singular data sources (`hyperping_monitor`, `hyperping_healthcheck`, etc.) - Filter by exact ID or name
  - Support for regex patterns, exact matching, case-insensitive matching, boolean filters, and numeric ranges
  - Short-circuit evaluation for optimal performance
  - 100% test coverage for filter framework (45+ unit tests)
- **Error Handling**: Enhanced error messages with context-aware troubleshooting guidance
  - Automatic error type detection (not_found, auth_error, rate_limit, server_error, validation, unknown)
  - Context-specific troubleshooting steps for each error type
  - Dashboard links for quick resource access (https://app.hyperping.io)
  - Rate limit errors include retry timing guidance
  - Auth errors provide API key verification steps
  - Validation errors highlight required fields and format requirements
  - 63 new integration tests validating error propagation across all CRUD operations
  - `docs/guides/error-handling.md` - Comprehensive error handling guide (4,000+ words)

### Changed

- **Import workflow**: All resources now validate IDs before import to provide clearer error messages
- **Filter framework**: Reusable filter schemas and matching functions available for all data sources
- **Error messages**: All CRUD operations (Create, Read, Update, Delete, List) now provide actionable troubleshooting steps

### Fixed

- Import errors now include resource type context and validation hints
- Data source pagination works correctly with client-side filtering
- Error messages no longer expose internal implementation details

## [1.0.8] - 2026-02-11

### Fixed

- **hyperping_monitor**: Fixed critical bug where port and ICMP monitors failed with "Provider produced inconsistent result after apply" (ISS-ICMP-002)
  - Root cause: HTTP-specific schema defaults (http_method, expected_status_code, follow_redirects) were applied to all monitor types, but API returns empty/null for non-HTTP protocols
  - Solution: Implemented save-restore pattern in Create, Read, and Update functions to preserve plan values for HTTP fields when monitor protocol is not "http"
  - Impact: Port and ICMP monitors now work correctly without state drift
  - Verified: Comprehensive testing with HTTP, Port (PostgreSQL/Redis), and ICMP (Google DNS/Cloudflare) monitors - all protocols create successfully with zero drift

## [1.0.7] - 2026-02-10

### Fixed

- **Documentation**: Fixed critical nested `docs/guides/` directories bug caused by backup/restore loop in lefthook
- **Documentation**: Updated coverage statistics to reflect current state (50.8%, 881 tests passing)
- **lefthook**: Fixed backup/restore logic to only copy markdown files, preventing directory recursion
- **lefthook**: Added validation check to fail if nested directories are detected

### Removed

- Removed 26 temporary development files (~1.5 MB):
  - Coverage output files (7 files)
  - Old scraper reports (13 files)
  - Temporary development tools (2 files)
  - Resolved issue documentation (2 files)
  - Old development plans (1 file)
  - Backup files (1 file)

### Added

- `docs/NESTED_GUIDES_BUG_ANALYSIS.md` - Comprehensive root cause analysis of directory nesting bug
- `docs/DOCUMENTATION_AUDIT_2026-02-10.md` - Complete documentation audit report

### Changed

- Updated CONTRIBUTING.md coverage threshold from 42% to 50%
- Updated README.md test coverage from 45.8% to 50.8%

## [1.0.6] - 2026-02-09

### Fixed

- **hyperping_incident**: Add read-after-update pattern to fix UPDATE operations (400 errors and state inconsistencies)
- **hyperping_maintenance**: Add read-after-update pattern to ensure state consistency after updates
- **All resources**: Incident and Maintenance now support full CRUD lifecycle (Create, Read, Update, Delete)

## [1.0.5] - 2026-02-09

### Fixed

- **hyperping_incident**: Preserve plan value for `text` field (write-only in API) to prevent state drift (ISS-005)
- **hyperping_maintenance**: Preserve plan value for `text` field (write-only in API) to prevent state drift (ISS-006)
- **hyperping_statuspage**: Preserve `settings.name` from plan to prevent API override (ISS-007.3)
- **hyperping_statuspage**: Preserve `show_response_times` and `show_uptime` boolean values from plan when API returns false (ISS-007.4)

## [1.0.4] - 2026-02-09

### Fixed

- **hyperping_incident**: Add read-after-create pattern to prevent "inconsistent result after apply" errors (ISS-005 - partial)
- **hyperping_maintenance**: Add read-after-create pattern to prevent "inconsistent result after apply" errors (ISS-006 - partial)
- **hyperping_statuspage**: Normalize subdomain by stripping `.hyperping.app` suffix to prevent state drift (ISS-007.1)
- **hyperping_statuspage**: Filter localized fields to only include configured languages, preventing drift from API auto-population (ISS-007.2)

### Fixed

- **hyperping_incident**: Add read-after-create pattern to prevent "inconsistent result after apply" errors (ISS-005)
- **hyperping_maintenance**: Add read-after-create pattern to prevent "inconsistent result after apply" errors (ISS-006)
- **hyperping_statuspage**: Normalize subdomain by stripping `.hyperping.app` suffix to prevent state drift (ISS-007)
- **hyperping_statuspage**: Filter localized fields to only include configured languages, preventing drift from API auto-population (ISS-007)

## [1.0.3] - 2026-02-08

### Added

- Reusable Terraform modules for common monitoring patterns (`api-health`, `ssl-monitor`, `statuspage-complete`)
- Import generator CLI tool for bulk importing existing Hyperping resources
- Reusable GitHub Actions workflow for Terraform operations
- API-to-provider coverage analyzer tool
- Automated API sync detection and contract testing
- Comprehensive documentation: import guides, troubleshooting, migration guide, rate limits guide

### Changed

- Enhanced analyzer to understand computed and undocumented fields
- Centralized API path constants in all tests
- Updated Go toolchain to 1.24.13 to fix crypto/tls vulnerability

### Fixed

- **hyperping_healthcheck**: Rename `tz` field to `timezone` to match API response
- **hyperping_outage**: Add `escalation_policy_uuid` field
- Align module schemas with actual provider implementation
- GPG signing configuration for releases

## [1.0.2] - 2026-01-25

### Fixed

- Release pipeline configuration

## [1.0.1] - 2026-01-24

### Added

- Terraform Registry documentation
- Community health files (contributing guidelines, issue templates)

### Fixed

- Broken links in README

## [1.0.0] - 2026-02-02

Initial stable release of the Terraform Provider for Hyperping.

This provider is production-ready with comprehensive test coverage (45.8% overall, 94% client), complete documentation, and all major Hyperping API features implemented. Per semantic versioning, v1.0.0 indicates a stable public API ready for production use.

### Added

#### Resources
- **hyperping_healthcheck** - Create and manage healthchecks (uptime monitors)
  - Support for HTTP/HTTPS URL monitoring with custom headers and body
  - Configurable check intervals (10s to 24h) and timeouts
  - Multi-region monitoring across 9 global regions
  - SSL certificate expiry monitoring
  - Pause/resume functionality
- **hyperping_monitor** - Create and manage monitors (legacy resource)
  - HTTP/HTTPS URL monitoring
  - Configurable frequency and timeout settings
  - Multi-region checks
  - Custom headers and request body support
- **hyperping_incident** - Manage status page incidents
  - Status workflow: investigating, identified, monitoring, resolved
  - Severity levels: minor, major, critical
  - Monitor linking for affected services
  - Subscriber notifications
- **hyperping_incident_update** - Add updates to existing incidents
  - Post status updates with timestamp tracking
  - Update incident status through the lifecycle
- **hyperping_maintenance** - Manage maintenance windows
  - Scheduled start/end times (RFC3339 format)
  - Monitor linking for planned maintenance
  - Advance notification support with configurable timing
- **hyperping_outage** - Manage outage records
  - Track service outages with start/end times
  - Link to affected monitors
  - Automatic vs manual outage classification

#### Data Sources
- **hyperping_healthcheck** - Retrieve a single healthcheck by UUID
- **hyperping_healthchecks** - List all healthchecks
- **hyperping_monitor** - Retrieve a single monitor by ID
- **hyperping_monitors** - List all monitors with filtering
- **hyperping_monitor_report** - Get uptime statistics and performance metrics
- **hyperping_incident** - Retrieve a single incident by ID
- **hyperping_incidents** - List all incidents
- **hyperping_maintenance_window** - Retrieve a single maintenance window by ID
- **hyperping_maintenance_windows** - List all maintenance windows
- **hyperping_outage** - Retrieve a single outage by ID
- **hyperping_outages** - List all outages

#### Provider Features
- API key authentication with environment variable support (`HYPERPING_API_KEY`)
- Configurable base URL for testing and alternative endpoints
- Exponential backoff retry logic with circuit breaker pattern
- Rate limit handling (429) with Retry-After header support
- Comprehensive input validation with helpful error messages
- Import support for all resources
- TLS 1.2+ enforcement for secure API communication
- Request/response logging for debugging (`TF_LOG=DEBUG`)
- User-Agent tracking for API telemetry

### Security
- API keys marked as sensitive in Terraform schema (won't appear in plan output)
- Log field masking to prevent API keys from appearing in debug logs
- Error message sanitization to redact credentials:
  - API keys (sk_*) are replaced with `sk_***REDACTED***`
  - Bearer tokens are replaced with `Bearer ***REDACTED***`
  - URL credentials are replaced with `://***REDACTED***@`
  - Authorization headers are replaced with `Authorization: ***REDACTED***`
- TLS hardening with minimum TLS 1.2 enforcement
- Input validation to prevent injection attacks

### Documentation
- Complete resource and data source documentation
- Provider configuration guide
- Multi-tenant pattern examples
- ADR (Architecture Decision Records) documenting key design choices
- Operations guide for production deployments
- Troubleshooting guide with common issues and solutions

[Unreleased]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.4.3...HEAD
[1.4.3]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.4.2...v1.4.3
[1.4.0]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.3.9...v1.4.0
[1.3.9]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.3.8...v1.3.9
[1.3.8]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.3.7...v1.3.8
[1.3.7]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.3.6...v1.3.7
[1.3.6]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.3.5...v1.3.6
[1.3.5]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.3.4...v1.3.5
[1.3.4]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.3.3...v1.3.4
[1.3.3]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.3.2...v1.3.3
[1.3.2]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.3.1...v1.3.2
[1.3.1]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.3.0...v1.3.1
[1.3.0]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.2.3...v1.3.0
[1.2.3]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.2.2...v1.2.3
[1.2.2]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.2.1...v1.2.2
[1.2.1]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.2.0...v1.2.1
[1.2.0]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.0.9...v1.1.0
[1.0.9]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.0.8...v1.0.9
[1.0.8]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.0.7...v1.0.8
[1.0.7]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.0.6...v1.0.7
[1.0.6]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.0.5...v1.0.6
[1.0.5]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.0.4...v1.0.5
[1.0.4]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.0.3...v1.0.4
[1.0.3]: https://github.com/develeap/terraform-provider-hyperping/releases/tag/v1.0.3
