# Pingdom to Hyperping Migration Tool

Complete Go CLI tool for migrating Pingdom checks to Hyperping monitors with full Terraform integration.

## Features

- Fetches all checks from Pingdom API
- Converts check types to Hyperping equivalents
- Tag-based naming convention conversion
- Generates Terraform HCL configuration
- Creates monitors in Hyperping
- Generates import scripts for Terraform state
- Comprehensive migration reports (JSON, text, markdown)
- Dry-run mode for validation

## Supported Check Types

| Pingdom Type | Hyperping Equivalent | Notes |
|--------------|---------------------|-------|
| HTTP/HTTPS | `protocol: http` | Direct 1:1 mapping |
| TCP | `protocol: port` | Port monitoring |
| PING | `protocol: icmp` | ICMP ping checks |
| SMTP | `protocol: port` (port 25/587) | Converted to TCP |
| POP3 | `protocol: port` (port 110/995) | Converted to TCP |
| IMAP | `protocol: port` (port 143/993) | Converted to TCP |
| DNS | **Not supported** | See manual steps |
| UDP | **Not supported** | See manual steps |
| Transaction | **Not supported** | Use external script + healthcheck |

## Installation

```bash
# Build
go build -o migrate-pingdom ./cmd/migrate-pingdom

# Or run directly
go run ./cmd/migrate-pingdom [flags]
```

## Usage

### Prerequisites

```bash
# Set API credentials
export PINGDOM_API_KEY="your_pingdom_token"
export HYPERPING_API_KEY="sk_your_hyperping_key"
```

### Basic Usage

```bash
# Dry run (generate configs without creating resources)
./migrate-pingdom --dry-run --output=./migration

# Full migration
./migrate-pingdom --output=./migration

# With resource name prefix
./migrate-pingdom --prefix=pingdom_ --output=./migration

# Verbose output
./migrate-pingdom --verbose --output=./migration
```

### CLI Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--pingdom-api-key` | Pingdom API token | `$PINGDOM_API_KEY` |
| `--hyperping-api-key` | Hyperping API key | `$HYPERPING_API_KEY` |
| `--output` | Output directory | `./pingdom-migration` |
| `--prefix` | Terraform resource name prefix | (none) |
| `--dry-run` | Generate configs without creating resources | `false` |
| `--verbose` | Verbose logging | `false` |
| `--pingdom-base-url` | Custom Pingdom API URL | (default) |
| `--hyperping-base-url` | Custom Hyperping API URL | `https://api.hyperping.io` |

## Tag to Naming Convention

The tool converts Pingdom tags to structured Hyperping names.

### Format

```
[ENVIRONMENT]-Category-ServiceName
[ENVIRONMENT-CUSTOMER]-Category-ServiceName
```

### Examples

| Pingdom Tags | Generated Name |
|--------------|----------------|
| `production`, `api`, `critical` | `[PROD]-API-Health-Critical` |
| `staging`, `web`, `frontend` | `[STAGING]-Web-Frontend` |
| `customer-acme`, `api`, `prod` | `[PROD-ACME]-API-ServiceName` |
| `dev`, `database` | `[DEV]-Database-ServiceName` |

### Tag Mappings

**Environment:**
- `production`, `prod` → `PROD`
- `staging`, `stage` → `STAGING`
- `development`, `dev` → `DEV`
- `qa`, `test` → `TEST` / `QA`

**Category:**
- `api` → `API`
- `web`, `website` → `Web`
- `database`, `db` → `Database`
- `cache`, `redis` → `Cache`
- `frontend` → `Frontend`
- `backend` → `Backend`

**Customer:**
- `customer-acme` → `ACME`
- `tenant-xyz` → `XYZ`

## Output Files

The tool generates the following files in the output directory:

### 1. `monitors.tf`

Terraform HCL configuration for all convertible monitors.

```hcl
# Pingdom Check ID: 12345
# Original Name: Production API Health
# Type: http
# Tags: production, api, critical

resource "hyperping_monitor" "prod_api_health" {
  name                 = "[PROD]-API-Health"
  url                  = "https://api.example.com/health"
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 300
  expected_status_code = "200"

  regions = ["virginia", "london", "frankfurt"]
}
```

### 2. `import.sh`

Executable shell script to import created resources into Terraform state.

```bash
#!/bin/bash
# Generated Terraform import script

terraform import hyperping_monitor.prod_api_health "mon_abc123"
```

### 3. `report.json`

Machine-readable JSON report with migration statistics.

```json
{
  "timestamp": "2026-02-13T10:00:00Z",
  "total_checks": 47,
  "supported_checks": 42,
  "unsupported_checks": 5,
  "checks_by_type": {
    "http": 35,
    "tcp": 7,
    "dns": 3,
    "transaction": 2
  },
  "unsupported_types": {
    "dns": 3,
    "transaction": 2
  }
}
```

### 4. `report.txt`

Human-readable text summary of the migration.

### 5. `manual-steps.md`

Markdown document with detailed instructions for handling unsupported check types.

## Migration Workflow

### 1. Export and Convert (Dry Run)

```bash
./migrate-pingdom --dry-run --output=./migration
```

Review generated files:
- Check `monitors.tf` for accuracy
- Review `report.txt` for warnings
- Read `manual-steps.md` for unsupported checks

### 2. Create Resources

```bash
./migrate-pingdom --output=./migration
```

This:
- Fetches all Pingdom checks
- Converts to Hyperping format
- Creates monitors in Hyperping
- Generates Terraform configs and import scripts

### 3. Import to Terraform

```bash
cd migration

# Initialize Terraform
terraform init

# Validate configuration
terraform validate

# Import resources
chmod +x import.sh
./import.sh

# Verify state matches
terraform plan
```

### 4. Handle Manual Steps

Review `manual-steps.md` for unsupported checks:

**DNS Checks:**
```hcl
# Option 1: Monitor DNS-over-HTTPS
resource "hyperping_monitor" "dns_check" {
  name = "[PROD]-DNS-Example"
  url  = "https://dns.google/resolve?name=example.com&type=A"
  protocol = "http"
  expected_status_code = "200"
  required_keyword = "example.com"
}

# Option 2: Monitor the service using DNS
resource "hyperping_monitor" "service" {
  name = "[PROD]-Service-UsingDNS"
  url  = "https://example.com/health"
  protocol = "http"
}
```

**Transaction Checks:**
```python
# Create external script (Playwright/Selenium)
# Deploy as Kubernetes CronJob
# Ping Hyperping healthcheck on success
```

## Architecture

```
cmd/migrate-pingdom/
├── main.go                    # CLI entry point
├── pingdom/
│   └── client.go             # Pingdom API client
├── converter/
│   ├── check.go              # Check type conversion
│   └── tags.go               # Tag to name mapping
├── generator/
│   ├── terraform.go          # HCL generation
│   └── import.go             # Import script generation
└── report/
    └── reporter.go           # Report generation
```

## Check Type Conversion Details

### HTTP/HTTPS → HTTP Monitor

```
Pingdom:
  type: http
  hostname: api.example.com
  url: /health
  encryption: true
  resolution: 5 (minutes)

Hyperping:
  protocol: http
  url: https://api.example.com/health
  check_frequency: 300 (seconds)
```

### TCP → Port Monitor

```
Pingdom:
  type: tcp
  hostname: db.example.com
  port: 5432

Hyperping:
  protocol: port
  url: db.example.com
  port: 5432
```

### PING → ICMP Monitor

```
Pingdom:
  type: ping
  hostname: server.example.com

Hyperping:
  protocol: icmp
  url: server.example.com
```

## Frequency Conversion

Pingdom uses minutes, Hyperping uses seconds. The tool rounds to nearest allowed frequency:

| Pingdom (min) | Hyperping (sec) |
|---------------|-----------------|
| 1 | 60 |
| 5 | 300 |
| 10 | 600 |
| 15 | 900 |
| 30 | 1800 |
| 60 | 3600 |

Allowed Hyperping frequencies: `60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400`

## Region Conversion

| Pingdom | Hyperping |
|---------|-----------|
| `region:NA` | `virginia, oregon` |
| `region:EU` | `london, frankfurt` |
| `region:APAC` | `singapore, sydney, tokyo` |
| `region:LATAM` | `saopaulo` |

Default (no filters): `virginia, london, frankfurt, singapore`

## Testing

```bash
# Run unit tests
go test ./cmd/migrate-pingdom/...

# Test specific functionality
go test -v ./cmd/migrate-pingdom -run TestCheckConversion
go test -v ./cmd/migrate-pingdom -run TestTagConversion
go test -v ./cmd/migrate-pingdom -run TestFrequencyConversion
```

## Examples

### Example 1: Basic Migration

```bash
export PINGDOM_API_KEY="abc123"
export HYPERPING_API_KEY="sk_xyz789"

./migrate-pingdom --output=./migration
```

Output:
```
Fetching Pingdom checks...
Fetched 47 checks from Pingdom
Converting checks to Hyperping format...
Converted 42/47 checks (5 unsupported)
Generating Terraform configuration...
Creating monitors in Hyperping...
Created 42 monitors in Hyperping (0 errors)

Migration Complete!
Output: ./migration
Files: monitors.tf, import.sh, report.json, report.txt, manual-steps.md
```

### Example 2: Dry Run with Prefix

```bash
./migrate-pingdom \
  --dry-run \
  --prefix=pingdom_ \
  --output=./test-migration
```

### Example 3: Custom API URLs

```bash
./migrate-pingdom \
  --pingdom-base-url=https://custom-pingdom.example.com/api/3.1 \
  --hyperping-base-url=https://staging-api.hyperping.io \
  --output=./staging-migration
```

## Troubleshooting

### API Authentication Errors

```
Error: Pingdom API key is required
```

**Solution:** Set `PINGDOM_API_KEY` environment variable or use `--pingdom-api-key` flag.

### Unsupported Check Types

The tool generates `manual-steps.md` with detailed instructions for handling:
- DNS checks
- UDP checks
- Transaction/browser checks

### Rate Limiting

If you hit API rate limits:

1. Run in dry-run mode first
2. Review generated configs
3. Create resources in batches (manually split config)

## Related Documentation

- [Pingdom Migration Guide](../../docs/guides/migrate-from-pingdom.md)
- [Hyperping Provider Documentation](https://registry.terraform.io/providers/develeap/hyperping)
- [Terraform Import Guide](https://developer.hashicorp.com/terraform/cli/import)

## License

MPL-2.0 - See LICENSE file
