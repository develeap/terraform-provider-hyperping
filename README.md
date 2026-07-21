# Terraform Provider for Hyperping

[![Tests](https://github.com/develeap/terraform-provider-hyperping/actions/workflows/test.yml/badge.svg)](https://github.com/develeap/terraform-provider-hyperping/actions/workflows/test.yml)
[![GitHub Release](https://img.shields.io/github/v/release/develeap/terraform-provider-hyperping)](https://github.com/develeap/terraform-provider-hyperping/releases)
[![Terraform Registry](https://img.shields.io/badge/terraform-registry-623CE4?logo=terraform)](https://registry.terraform.io/providers/develeap/hyperping)
[![License: MPL-2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/develeap/terraform-provider-hyperping)](https://goreportcard.com/report/github.com/develeap/terraform-provider-hyperping)

Community Terraform provider for [Hyperping](https://hyperping.io/) - manage uptime monitoring, incidents, status pages, and maintenance windows as infrastructure-as-code.

## Features

**Resources:**
- `hyperping_monitor` - HTTP/HTTPS uptime monitors with multi-region checks
- `hyperping_incident` - Incident management with status updates
- `hyperping_maintenance` - Scheduled maintenance windows
- `hyperping_statuspage` - Public status pages with customization
- `hyperping_statuspage_subscriber` - Status page notifications (email, SMS, Teams)
- `hyperping_healthcheck` - Cron job monitoring
- `hyperping_outage` - Outage tracking and management

**Data Sources:**
- `hyperping_monitors` - List/filter all monitors
- `hyperping_statuspages` - List/search status pages with pagination
- `hyperping_statuspage` - Single status page details
- `hyperping_statuspage_subscribers` - List subscribers by type
- `hyperping_monitor_report` - Uptime/SLA reports
- `hyperping_escalation_policies` - List all escalation policies (MCP)
- `hyperping_escalation_policy` - Single escalation policy lookup (MCP)
- `hyperping_on_call_schedules` - List all on-call schedules (MCP)
- `hyperping_on_call_schedule` - Single on-call schedule lookup (MCP)
- `hyperping_integrations` - List all integrations (MCP)

## Quick Start

```hcl
terraform {
  required_providers {
    hyperping = {
      source  = "develeap/hyperping"
      version = "~> 1.0"
    }
  }
}

provider "hyperping" {
  # Set HYPERPING_API_KEY environment variable
  # Or: api_key = "sk_..."
}

# Create a monitor
resource "hyperping_monitor" "api" {
  name                 = "API Health Check"
  url                  = "https://api.example.com/health"
  protocol             = "http"
  check_frequency      = 60
  expected_status_code = "200"
  regions              = ["london", "virginia", "singapore"]
}

# Create a status page
resource "hyperping_statuspage" "public" {
  name      = "Service Status"
  subdomain = "status"  # status.hyperping.app
  theme     = "dark"

  sections = [{
    name     = { en = "API Services" }
    is_split = true
    services = [{
      monitor_uuid        = hyperping_monitor.api.id
      show_uptime         = true
      show_response_times = true
    }]
  }]
}

# Add email subscribers
resource "hyperping_statuspage_subscriber" "team" {
  statuspage_uuid = hyperping_statuspage.public.id
  type            = "email"
  email           = "team@example.com"
}

# Schedule maintenance
resource "hyperping_maintenance" "upgrade" {
  name       = "Database Upgrade"
  start_date = "2026-02-15T02:00:00Z"
  end_date   = "2026-02-15T04:00:00Z"
  monitors   = [hyperping_monitor.api.id]
}
```

## Authentication

Get your API key from the [Hyperping dashboard](https://app.hyperping.io/).

```bash
export HYPERPING_API_KEY="sk_your_api_key"
terraform plan
```

Or configure in the provider block (not recommended for production).

### Credential handling

The `api_key` attribute is marked `Sensitive`, so Terraform redacts it from CLI
output and from `terraform show -json`. For production use, provide the key via the
`HYPERPING_API_KEY` environment variable rather than hardcoding it in a `.tf` file:

- Environment variable values are **never** written to state or to saved plan files
  (`terraform plan -out=plan.tfplan`). Provider configuration is ephemeral: Terraform
  re-evaluates it during apply and only records the provider address, not its values.
- A key hardcoded as a literal in a `provider "hyperping"` block is captured in the
  plan file's configuration snapshot. `Sensitive` does not remove it from that snapshot,
  so treat any saved `.tfplan` as a secret and never commit it to version control.

Terraform's `WriteOnly` attribute feature does not apply here: it is a managed-resource
concept that the `ConfigureProvider` RPC does not process for provider schemas, so it is
not available on (and would not benefit) provider configuration attributes.

## Migration Tools

Automated CLI tools for migrating from other monitoring platforms:

```bash
# Migrate from Better Stack
go install github.com/develeap/terraform-provider-hyperping/cmd/migrate-betterstack@latest
migrate-betterstack --betterstack-token $BETTERSTACK_TOKEN --hyperping-api-key $HYPERPING_KEY

# Migrate from UptimeRobot
go install github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot@latest
migrate-uptimerobot --uptimerobot-api-key $UPTIMEROBOT_KEY --hyperping-api-key $HYPERPING_KEY

# Migrate from Pingdom
go install github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom@latest
migrate-pingdom --pingdom-api-key $PINGDOM_KEY --hyperping-api-key $HYPERPING_KEY
```

**Features:**
- ✅ Automated export from source platform
- ✅ Intelligent resource conversion
- ✅ Production-ready Terraform generation
- ✅ Validation and compatibility checks
- ✅ Shared utilities (`pkg/migrate/`) for frequency mapping, region translation, name sanitization
- ✅ 85-90% time savings vs. manual migration

**Guides:**
- [Automated Migration Tools](./docs/guides/automated-migration.md) - Complete CLI tool documentation
- [Better Stack Migration](./docs/guides/migrate-from-betterstack.md) - Better Stack-specific guide
- [UptimeRobot Migration](./docs/guides/migrate-from-uptimerobot.md) - UptimeRobot-specific guide
- [Pingdom Migration](./docs/guides/migrate-from-pingdom.md) - Pingdom-specific guide

## Documentation

- **[Wiki](https://github.com/develeap/terraform-provider-hyperping/wiki)** - Complete guide with all resources, examples, migration tools, and architecture docs
- **[Terraform Registry](https://registry.terraform.io/providers/develeap/hyperping/latest/docs)** - Complete resource/data source reference
- **[Examples](./examples/)** - Real-world usage examples
- **[Changelog](./CHANGELOG.md)** - Version history and release notes
- **[Hyperping API Docs](https://hyperping.com/docs/api/overview)** - Official API documentation

## Requirements

- Terraform >= 1.0
- Go >= 1.26.2 (development only)
- Hyperping account with API key

## Development

```bash
# Clone and build
git clone https://github.com/develeap/terraform-provider-hyperping.git
cd terraform-provider-hyperping
go build -v

# Run all tests
go test -v ./...

# Run acceptance tests (requires API key)
HYPERPING_API_KEY=sk_xxx TF_ACC=1 go test -v ./internal/provider/

# Lint and security scan
golangci-lint run
gosec ./...
```

### Testing

Provider acceptance tests use mock HTTP servers — no real API key required for unit/acceptance tests. VCR cassettes for the REST client are maintained in the [`hyperping-go`](https://github.com/develeap/hyperping-go) module.

#### Acceptance Tests

```bash
# Run all acceptance tests (mock HTTP servers, no real API key needed)
go test -v ./internal/provider/

# Run specific tests
go test -v -run "TestAccMonitorResource" ./internal/provider/
go test -v -run "TestAccDataSource" ./internal/provider/

# Run with real API (optional, for full contract validation)
TF_ACC=1 HYPERPING_TEST_API_KEY=sk_xxx go test -v ./internal/provider/
```

#### Test Coverage

- **Race detector clean**
- **Zero flaky tests** across multiple runs
- Protocol-specific tests: Port, ICMP, HTTP protocol handling
- State drift detection tests for all CRUD resources
- Unit tests for all MCP data sources (constructor, Metadata, Schema, Configure)

#### Integration Tests for Migration Tools

Integration tests validate end-to-end workflows for all 3 migration tools with real APIs:

```bash
# Setup credentials
export BETTERSTACK_API_TOKEN=your_token
export UPTIMEROBOT_API_KEY=your_key
export PINGDOM_API_KEY=your_key
export HYPERPING_API_KEY=sk_your_key

# Run all integration tests
go test -v -tags=integration -timeout=30m ./cmd/migrate-betterstack/...
go test -v -tags=integration -timeout=30m ./cmd/migrate-uptimerobot/...
go test -v -tags=integration -timeout=30m ./cmd/migrate-pingdom/...

# Run specific scenarios
go test -v -tags=integration -run=".*SmallScenario" ./cmd/migrate-betterstack/...
```

**Each test validates:**
1. API connection to source platform
2. Migration tool execution without errors
3. All 4 output files generated (.tf, import.sh, report.json, manual-steps.md)
4. Terraform validation passes (terraform validate)
5. Terraform plan shows expected resources (0 errors)
6. Import script is executable with valid syntax
7. Report and manual steps files are valid JSON/Markdown
8. Resource count matches expected scenario

See [docs/INTEGRATION_TESTING.md](docs/INTEGRATION_TESTING.md) for complete documentation.

## Production Features

- **Security:** TLS 1.2+, credential sanitization, HTTPS enforcement
- **Reliability:** Circuit breaker, exponential backoff, rate limit handling
- **Testing:** 50.8% code coverage, race condition testing

Enable debug logging:
```bash
export TF_LOG=DEBUG
terraform apply
```

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

1. Fork the repository
2. Create a feature branch
3. Write tests (maintain coverage standards)
4. Submit a pull request

## Support

- [Issue Tracker](https://github.com/develeap/terraform-provider-hyperping/issues)

## Related Projects

- **[hyperping-exporter](https://github.com/develeap/hyperping-exporter)** — Prometheus exporter for Hyperping monitoring metrics. Standalone Go service with Grafana dashboards, alerting rules, and Kubernetes manifests.

## License

Mozilla Public License 2.0 - see [LICENSE](LICENSE) for details.

---

Maintained by [Develeap](https://develeap.com)
