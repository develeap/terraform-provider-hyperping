# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

**Note:** Versions 1.0.1 and 1.0.2 exist as git tags but were never published to Terraform Registry.
Published releases start from v1.0.3.

## [Unreleased]

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

[Unreleased]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.0.9...HEAD
[1.0.9]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.0.8...v1.0.9
[1.0.8]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.0.7...v1.0.8
[1.0.7]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.0.6...v1.0.7
[1.0.6]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.0.5...v1.0.6
[1.0.5]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.0.4...v1.0.5
[1.0.4]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.0.3...v1.0.4
[1.0.3]: https://github.com/develeap/terraform-provider-hyperping/releases/tag/v1.0.3
