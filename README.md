# Terraform Provider Hyperping

[![Terraform Registry](https://img.shields.io/badge/terraform-registry-623CE4?logo=terraform)](https://registry.terraform.io/providers/develeap/hyperping)
[![License: MPL-2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://go.dev/)
[![Terraform](https://img.shields.io/badge/Terraform-1.0+-844FBA?logo=terraform)](https://www.terraform.io/)

<!-- After first release, add these badges:
[![Tests](https://github.com/develeap/terraform-provider-hyperping/actions/workflows/test.yml/badge.svg)](https://github.com/develeap/terraform-provider-hyperping/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/develeap/terraform-provider-hyperping)](https://goreportcard.com/report/github.com/develeap/terraform-provider-hyperping)
[![GitHub Release](https://img.shields.io/github/v/release/develeap/terraform-provider-hyperping)](https://github.com/develeap/terraform-provider-hyperping/releases)
-->

Terraform provider for [Hyperping](https://hyperping.io/) - uptime monitoring with public status pages.

Manage your Hyperping monitors, incidents, and maintenance windows as infrastructure-as-code.

## Features

### Resources
- **hyperping_monitor** - Create and manage HTTP, port, and ICMP monitors
- **hyperping_incident** - Manage incidents for status page communication
- **hyperping_incident_update** - Add updates to existing incidents
- **hyperping_maintenance** - Schedule maintenance windows

### Data Sources
- **hyperping_monitors** - List all monitors
- **hyperping_monitor** - Get a single monitor by ID
- **hyperping_incidents** - List all incidents
- **hyperping_incident** - Get a single incident by ID
- **hyperping_maintenance_windows** - List all maintenance windows
- **hyperping_maintenance_window** - Get a single maintenance window by ID
- **hyperping_monitor_report** - Get uptime/SLA report for a monitor

## Security & Reliability

### Built-in Security Features
- **🔐 Credential Protection**: API keys are automatically redacted from all logs and error messages
- **🔒 TLS 1.2+ Enforcement**: All HTTPS connections require TLS 1.2 or higher
- **🛡️ BaseURL Validation**: Provider enforces HTTPS and validates API endpoints
- **📊 Request Duration Logging**: Performance monitoring with millisecond precision
- **♻️ Smart Retry Logic**: Automatic retry with exponential backoff for transient failures
- **⏱️ Rate Limit Handling**: Respects `Retry-After` headers from the API
- **🔍 User-Agent Tracking**: Helps Hyperping identify provider version for better support
- **🔌 Circuit Breaker**: Prevents cascading failures during API degradation
- **📈 Metrics Interface**: Optional integration with Prometheus, CloudWatch, Datadog, etc.

### Debug Logging

Enable debug logging to troubleshoot issues (API keys are automatically redacted):

```bash
export TF_LOG=DEBUG
terraform apply
```

Example debug output:
```
[DEBUG] received API response: method=GET path=/v1/monitors status_code=200 duration_ms=145
```

### Production Reliability

The provider includes production-grade reliability features:

- **Circuit Breaker**: Automatically opens the circuit after 60% failure rate to prevent cascading failures. The circuit recovers automatically after 30 seconds.
- **Metrics Collection**: Optional metrics interface for observability. Integrate with your monitoring system:
  - API call latency and status codes
  - Retry attempts
  - Circuit breaker state changes

See [OPERATIONS.md](docs/OPERATIONS.md) for detailed production deployment guidance.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24 (for development only)
- [Hyperping](https://hyperping.io/) account with API key

## Installation

```hcl
terraform {
  required_providers {
    hyperping = {
      source = "develeap/hyperping"
      version = "~> 1.0"
    }
  }
}

provider "hyperping" {
  # API key can be set via HYPERPING_API_KEY environment variable
  # api_key = "sk_..."
}
```

## Authentication

The provider requires a Hyperping API key. You can obtain one from the [Hyperping dashboard](https://app.hyperping.io/).

Set the API key using one of these methods:

1. **Environment variable** (recommended):
   ```bash
   export HYPERPING_API_KEY="sk_your_api_key_here"
   ```

2. **Provider configuration**:
   ```hcl
   provider "hyperping" {
     api_key = "sk_your_api_key_here"
   }
   ```

## Quick Start

```hcl
# Create a monitor
resource "hyperping_monitor" "website" {
  name                 = "My Website"
  url                  = "https://example.com"
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 60
  expected_status_code = "200"
  regions              = ["london", "virginia", "singapore"]
}

# Create an incident
resource "hyperping_incident" "outage" {
  title               = "Website Performance Issues"
  text                = "Investigating reports of slow response times."
  type                = "incident"
  status_pages        = ["sp_your_status_page_id"]
  affected_components = []  # Optional
}

# Add an update to the incident
resource "hyperping_incident_update" "investigating" {
  incident_id = hyperping_incident.outage.id
  text        = "We have identified the issue and are working on a fix."
  type        = "identified"
}

# Schedule a maintenance window
resource "hyperping_maintenance" "upgrade" {
  name       = "infrastructure-upgrade"
  title      = "Scheduled Maintenance"
  text       = "Upgrading infrastructure for improved performance."
  start_date = "2026-02-01T02:00:00.000Z"
  end_date   = "2026-02-01T04:00:00.000Z"
  monitors   = [hyperping_monitor.website.id]
}
```

## Resources

### hyperping_monitor

Manages uptime monitors.

```hcl
resource "hyperping_monitor" "api" {
  name                 = "API Health Check"
  url                  = "https://api.example.com/health"
  protocol             = "http"
  http_method          = "POST"
  check_frequency      = 300
  expected_status_code = "200"
  follow_redirects     = false
  regions              = ["london", "frankfurt", "virginia"]

  request_headers = [
    { name = "Content-Type", value = "application/json" },
    { name = "Authorization", value = "Bearer ${var.api_token}" }
  ]

  request_body = jsonencode({ check = "health" })
}
```

| Argument | Type | Required | Description |
|----------|------|----------|-------------|
| `name` | string | Yes | Monitor display name |
| `url` | string | Yes | URL to monitor |
| `protocol` | string | No | Protocol: `http`, `port`, `icmp`. Default: `http` |
| `http_method` | string | No | HTTP method: `GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `HEAD`, `OPTIONS`. Default: `GET` |
| `check_frequency` | number | No | Check interval in seconds. Valid: 10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400. Default: `30` |
| `expected_status_code` | string | No | Expected status code (`200`) or pattern (`2xx`). Default: `2xx` |
| `follow_redirects` | bool | No | Follow HTTP redirects. Default: `true` |
| `regions` | list(string) | No | Check regions (see [Regions](#available-regions)) |
| `request_headers` | list(object) | No | Custom headers: `[{name, value}]` |
| `request_body` | string | No | Request body for POST/PUT/PATCH |
| `paused` | bool | No | Pause monitoring. Default: `false` |
| `port` | number | No | Port number (required when protocol is `port`) |
| `alerts_wait` | number | No | Seconds to wait before alerting |
| `escalation_policy` | string | No | Escalation policy UUID |
| `required_keyword` | string | No | Keyword that must appear in response |

**Computed Attributes:** `id`

### hyperping_incident

Manages incidents for status page communication.

```hcl
resource "hyperping_incident" "outage" {
  title                = "Service Disruption"
  text                 = "We are investigating connectivity issues."
  type                 = "outage"
  status_pages         = ["sp_your_status_page_id"]
  affected_components  = ["comp_api", "comp_web"]
}
```

| Argument | Type | Required | Description |
|----------|------|----------|-------------|
| `title` | string | Yes | Incident title |
| `text` | string | Yes | Incident description |
| `type` | string | No | Type: `outage` or `incident`. Default: `incident` |
| `status_pages` | list(string) | Yes | Status page UUIDs to display on |
| `affected_components` | list(string) | No | Affected component UUIDs |

**Computed Attributes:** `id`, `date`

### hyperping_incident_update

Adds an update to an existing incident.

```hcl
resource "hyperping_incident_update" "resolved" {
  incident_id = hyperping_incident.outage.id
  text        = "The issue has been resolved."
  type        = "resolved"
}
```

| Argument | Type | Required | Description |
|----------|------|----------|-------------|
| `incident_id` | string | Yes | Parent incident UUID |
| `text` | string | Yes | Update text |
| `type` | string | Yes | Type: `investigating`, `identified`, `update`, `monitoring`, `resolved` |
| `date` | string | No | Update date (ISO 8601). Default: current time |

**Computed Attributes:** `id`

### hyperping_maintenance

Manages scheduled maintenance windows.

```hcl
resource "hyperping_maintenance" "upgrade" {
  name                 = "Database Maintenance"
  title                = "Scheduled Database Maintenance"
  text                 = "We will be performing database maintenance."
  start_date           = "2026-02-01T02:00:00Z"
  end_date             = "2026-02-01T04:00:00Z"
  monitors             = [hyperping_monitor.api.id]
  status_pages         = ["sp_your_status_page_id"]
  notification_option  = "scheduled"
  notification_minutes = 60
}
```

| Argument | Type | Required | Description |
|----------|------|----------|-------------|
| `name` | string | Yes | Internal name |
| `start_date` | string | Yes | Start time (ISO 8601) |
| `end_date` | string | Yes | End time (ISO 8601) |
| `monitors` | list(string) | Yes | Monitor UUIDs affected |
| `title` | string | No | Public title |
| `text` | string | No | Public description |
| `status_pages` | list(string) | No | Status page UUIDs |
| `notification_option` | string | No | Notification timing: `none`, `immediate`, or `scheduled` |
| `notification_minutes` | number | No | Minutes before to notify |

**Computed Attributes:** `id`, `timezone`, `created_by`, `created_at`

## Data Sources

### hyperping_monitors

```hcl
data "hyperping_monitors" "all" {}

output "monitor_names" {
  value = [for m in data.hyperping_monitors.all.monitors : m.name]
}
```

### hyperping_monitor

```hcl
data "hyperping_monitor" "api" {
  id = "mon_abc123"
}

output "api_url" {
  value = data.hyperping_monitor.api.url
}
```

### hyperping_monitor_report

```hcl
data "hyperping_monitor_report" "api" {
  monitor_id = hyperping_monitor.api.id
  from       = "2026-01-01T00:00:00Z"
  to         = "2026-01-31T23:59:59Z"
}

output "uptime_sla" {
  value = "${data.hyperping_monitor_report.api.sla}%"
}

output "outage_count" {
  value = data.hyperping_monitor_report.api.outages_count
}
```

### hyperping_incidents / hyperping_incident

```hcl
data "hyperping_incidents" "all" {}

data "hyperping_incident" "specific" {
  id = "inci_abc123"
}
```

### hyperping_maintenance_windows / hyperping_maintenance_window

```hcl
data "hyperping_maintenance_windows" "all" {}

data "hyperping_maintenance_window" "specific" {
  id = "mw_abc123"
}
```

## Provider Configuration

| Argument | Type | Required | Description |
|----------|------|----------|-------------|
| `api_key` | string | No | Hyperping API key. Can use `HYPERPING_API_KEY` env var |
| `base_url` | string | No | API base URL. Default: `https://api.hyperping.io` |

## Available Regions

Monitors can check from these regions:

| Region Code | Location |
|-------------|----------|
| `sanfrancisco` | San Francisco, USA |
| `california` | California, USA |
| `virginia` | Virginia, USA |
| `nyc` | New York City, USA |
| `oregon` | Oregon, USA |
| `toronto` | Toronto, Canada |
| `saopaulo` | São Paulo, Brazil |
| `london` | London, UK |
| `paris` | Paris, France |
| `frankfurt` | Frankfurt, Germany |
| `amsterdam` | Amsterdam, Netherlands |
| `seoul` | Seoul, South Korea |
| `tokyo` | Tokyo, Japan |
| `singapore` | Singapore |
| `sydney` | Sydney, Australia |
| `mumbai` | Mumbai, India |
| `bangalore` | Bangalore, India |
| `bahrain` | Bahrain |
| `capetown` | Cape Town, South Africa |

## Development

### Building

```bash
git clone https://github.com/develeap/terraform-provider-hyperping.git
cd terraform-provider-hyperping
go build -v ./...
```

### Testing

```bash
# Run unit tests
go test -v ./...

# Run unit tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# Run acceptance tests (requires HYPERPING_API_KEY)
TF_ACC=1 go test -v ./internal/provider/
```

### Linting

```bash
# Static analysis
go vet ./...
staticcheck ./...

# Security scan
gosec ./...
```

### Generate Documentation

```bash
go generate ./...
```

## Troubleshooting

### Enable Debug Logging

```bash
export TF_LOG=DEBUG
terraform apply 2>&1 | tee terraform.log
```

### Common Issues

**Authentication Error (401)**
- Verify your API key: `curl -H "Authorization: Bearer $HYPERPING_API_KEY" https://api.hyperping.io/v1/monitors`

**Invalid Region**
- Use only valid region codes from the [Available Regions](#available-regions) table

**Rate Limiting (429)**
- The provider automatically retries with exponential backoff
- Respects `Retry-After` headers from the API
- Error messages show how long to wait: `retry after 60 seconds`
- For bulk changes: `terraform apply -parallelism=1`

**Resource Not Found (404)**
- Resource may have been deleted outside Terraform
- Remove from state: `terraform state rm hyperping_monitor.example`

## Roadmap

### ✅ Implemented (Terraform Resources)
- ✅ **Monitors** - HTTP, Port, and ICMP monitors
- ✅ **Incidents** - Incident management and updates
- ✅ **Maintenance Windows** - Schedule maintenance windows
- ✅ **Monitor Reports** - Uptime/SLA data via data source
- ✅ **All Data Sources** - List and read all resource types
- ✅ **Security Hardening** - Credential sanitization, TLS 1.2+
- ✅ **User-Agent Telemetry** - Version tracking for better support
- ✅ **Smart Retry Logic** - Exponential backoff with Retry-After support

### ✅ API Client Implemented (No Terraform Resources Yet)
These APIs are fully implemented at the client level (33 total endpoints with 95.3% test coverage):
- ✅ **Outages API** (8 endpoints) - List, Get, Create, Acknowledge, Unacknowledge, Resolve, Escalate, Delete
- ✅ **Healthchecks API** (7 endpoints) - Full CRUD + Pause/Resume for cron job monitoring

**Want Terraform resources for these?** [Open an issue](https://github.com/develeap/terraform-provider-hyperping/issues) to request!

### 🔮 Planned (Waiting for Hyperping API)
These features will be added when Hyperping releases the corresponding APIs:

- ⏳ **Status Pages** - Manage public status pages (referenced but no CRUD endpoints found)
- ⏳ **Components** - Define service components for status pages
- ⏳ **Subscribers** - Manage notification subscribers
- ⏳ **Escalation Policies** - Configure alert escalation workflows (referenced but no CRUD endpoints)
- ⏳ **Integrations** - Connect to Slack, PagerDuty, webhooks

### 💡 Community Requests
Missing something? [Open an issue](https://github.com/develeap/terraform-provider-hyperping/issues) to request new features!

**Update Schedule**: This provider is actively maintained with weekly updates.

## Contributing

This is a **community-maintained** project and contributions are highly encouraged! Whether you're fixing bugs, adding features, or improving documentation, your help is appreciated.

### Ways to Contribute

- 🐛 **Report bugs** - [Open an issue](https://github.com/develeap/terraform-provider-hyperping/issues/new?template=bug_report.yml)
- ✨ **Request features** - [Open a feature request](https://github.com/develeap/terraform-provider-hyperping/issues/new?template=feature_request.yml)
- 📖 **Improve docs** - Fix typos, clarify instructions, add examples
- 🔧 **Submit PRs** - Bug fixes and new features

### Pull Request Process

1. **Fork** the repository
2. **Create** a feature branch: `git checkout -b feature/my-feature`
3. **Write tests** for new functionality (we maintain 95%+ coverage)
4. **Ensure** all tests pass: `go test -v ./...`
5. **Run** linters: `go vet ./...` and `staticcheck ./...`
6. **Commit** with clear messages following [Conventional Commits](https://www.conventionalcommits.org/)
7. **Submit** a pull request with a detailed description

### Development Setup

See the [CONTRIBUTING.md](.github/CONTRIBUTING.md) guide for detailed setup instructions, coding standards, and testing guidelines.

## License

This project is licensed under the Mozilla Public License 2.0 - see the [LICENSE](LICENSE) file for details.

## Links

- [Hyperping Website](https://hyperping.io/)
- [Hyperping API Documentation](https://hyperping.notion.site/Hyperping-API-documentation-a0dc48fb818e4542a8f7fb4163ede2c3)
- [Issue Tracker](https://github.com/develeap/terraform-provider-hyperping/issues)
