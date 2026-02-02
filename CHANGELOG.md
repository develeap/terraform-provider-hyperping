# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial resources and data sources for Hyperping API management

## [1.0.0] - TBD

Initial stable release of the Terraform Provider for Hyperping.

This provider is production-ready with comprehensive test coverage (76%), complete documentation, and all major Hyperping API features implemented. Per semantic versioning, v1.0.0 indicates a stable public API ready for production use.

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

[Unreleased]: https://github.com/develeap/terraform-provider-hyperping/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/develeap/terraform-provider-hyperping/releases/tag/v1.0.0
