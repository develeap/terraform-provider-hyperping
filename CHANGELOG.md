# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- **hyperping_incident**: Preserve plan value for `text` field (write-only in API) to prevent state drift (ISS-005)
- **hyperping_maintenance**: Preserve plan value for `text` field (write-only in API) to prevent state drift (ISS-006)

### Known Issues

- **hyperping_statuspage**: `settings.name` field is overridden by API with resource-level `name` value (ISS-007.3)
- **hyperping_statuspage**: `show_response_times` field may flip from true to false after apply (ISS-007.4)

## [1.0.4] - 2026-02-09

### Fixed (Partial - See v1.0.5 for complete fixes)

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

[Unreleased]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.0.4...HEAD
[1.0.4]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.0.3...v1.0.4
[1.0.3]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.0.2...v1.0.3
[1.0.2]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/develeap/terraform-provider-hyperping/releases/tag/v1.0.0
