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

## Migration Tools

Automated CLI tools for migrating from other monitoring platforms:

```bash
# Migrate from Better Stack
go install github.com/develeap/terraform-provider-hyperping/cmd/migrate-betterstack@latest
migrate-betterstack migrate --source-api-key $BETTERSTACK_TOKEN --dest-api-key $HYPERPING_KEY

# Migrate from UptimeRobot
go install github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot@latest
migrate-uptimerobot migrate --source-api-key $UPTIMEROBOT_KEY --dest-api-key $HYPERPING_KEY

# Migrate from Pingdom
go install github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom@latest
migrate-pingdom migrate --source-api-key $PINGDOM_KEY --dest-api-key $HYPERPING_KEY
```

**Features:**
- ✅ Automated export from source platform
- ✅ Intelligent resource conversion
- ✅ Production-ready Terraform generation
- ✅ Validation and compatibility checks
- ✅ 85-90% time savings vs. manual migration

**Guides:**
- [Automated Migration Tools](./docs/guides/automated-migration.md) - Complete CLI tool documentation
- [Better Stack Migration](./docs/guides/migrate-from-betterstack.md) - Better Stack-specific guide
- [UptimeRobot Migration](./docs/guides/migrate-from-uptimerobot.md) - UptimeRobot-specific guide
- [Pingdom Migration](./docs/guides/migrate-from-pingdom.md) - Pingdom-specific guide

## Documentation

- **[Terraform Registry](https://registry.terraform.io/providers/develeap/hyperping/latest/docs)** - Complete resource/data source reference
- **[Examples](./examples/)** - Real-world usage examples
- **[Changelog](./CHANGELOG.md)** - Version history and release notes
- **[Hyperping API Docs](https://hyperping.com/docs/api/overview)** - Official API documentation

## Requirements

- Terraform >= 1.0
- Go >= 1.24 (development only)
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

The provider uses **VCR (Video Cassette Recorder)** for hermetic, fast, and deterministic testing.

#### Contract Tests (No API Key Required)

Contract tests validate API response structure using pre-recorded cassettes:

```bash
# Run all contract tests (no API key needed)
export VCR_MODE=replay
go test -v -run "^TestContract" ./internal/client/

# Run specific resource tests
go test -v -run "TestContract_Monitor" ./internal/client/
go test -v -run "TestContract_StatusPage" ./internal/client/

# Check for flakiness
for i in {1..10}; do go test -count=1 -run "^TestContract" ./internal/client/ || exit 1; done
```

**Benefits:**
- No API key required
- Fast execution (~7 seconds for 356 tests)
- Deterministic results (same every time)
- Safe for CI/CD (no rate limiting)
- Detects API breaking changes

#### Recording New Cassettes

When the API changes or new tests are added:

```bash
# Set API key
export HYPERPING_API_KEY=sk_your_key_here

# Record mode
export VCR_MODE=record
go test -v -run TestLiveContract_Monitor_CRUD ./internal/client/

# Verify cassette created
ls -lh internal/client/testdata/cassettes/monitor_crud.yaml

# Test in replay mode
unset HYPERPING_API_KEY
export VCR_MODE=replay
go test -v -run TestContract_Monitor ./internal/client/
```

**Security:** All cassettes automatically mask API keys and sensitive headers.

#### Test Coverage

- **356 contract tests** validating API responses
- **41 VCR cassettes** covering all resources
- **100% passing** core tests (79/79)
- **Zero flaky tests** across multiple runs
- **Race detector clean**

See [internal/client/testdata/cassettes/README.md](internal/client/testdata/cassettes/README.md) for cassette documentation.

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
- **Observability:** Optional metrics integration (Prometheus, CloudWatch, Datadog)
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

## License

Mozilla Public License 2.0 - see [LICENSE](LICENSE) for details.

---

Maintained by [Develeap](https://develeap.com)
