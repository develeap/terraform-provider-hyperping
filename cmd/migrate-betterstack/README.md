# Better Stack to Hyperping Migration Tool

A complete CLI tool for migrating monitoring resources from Better Stack to Hyperping with Terraform.

## Features

- **Automatic conversion** of Better Stack monitors to Hyperping monitors
- **Heartbeat to healthcheck** conversion
- **Region mapping** between Better Stack and Hyperping regions
- **Frequency normalization** to supported values
- **Terraform HCL generation** with proper formatting
- **Import script generation** for existing resources
- **Detailed migration report** in JSON format
- **Manual steps documentation** for items requiring attention
- **Validation support** with terraform validate
- **Dry-run mode** for testing without file creation

## Installation

### From Source

```bash
cd cmd/migrate-betterstack
go build -o migrate-betterstack
```

### Run Directly

```bash
go run ./cmd/migrate-betterstack [options]
```

## Prerequisites

- Go 1.24 or later
- Better Stack API token
- Hyperping API key
- Terraform 1.8+ (for validation)

## Usage

### Basic Migration

```bash
export BETTERSTACK_API_TOKEN="your_betterstack_token"
export HYPERPING_API_KEY="sk_your_hyperping_key"

migrate-betterstack
```

### With Custom Options

```bash
migrate-betterstack \
  --betterstack-token="your_token" \
  --hyperping-api-key="sk_your_key" \
  --output=production.tf \
  --import-script=import-prod.sh \
  --report=migration-report.json \
  --verbose
```

### Dry Run (Validation Only)

```bash
migrate-betterstack --dry-run --verbose
```

### With Terraform Validation

```bash
migrate-betterstack --validate
```

## Command-Line Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--betterstack-token` | `$BETTERSTACK_API_TOKEN` | Better Stack API token |
| `--hyperping-api-key` | `$HYPERPING_API_KEY` | Hyperping API key |
| `--output` | `migrated-resources.tf` | Terraform configuration output file |
| `--import-script` | `import.sh` | Import script output file |
| `--report` | `migration-report.json` | Migration report output file |
| `--manual-steps` | `manual-steps.md` | Manual steps documentation file |
| `--dry-run` | `false` | Validate without creating files |
| `--validate` | `false` | Run terraform validate on output |
| `--verbose` | `false` | Enable verbose logging |

## Output Files

The tool generates four files:

### 1. migrated-resources.tf

Complete Terraform configuration with:
- Provider configuration
- Monitor resources
- Healthcheck resources
- Inline comments for conversion notes

Example:
```hcl
resource "hyperping_monitor" "api_health" {
  name                 = "API Health Check"
  url                  = "https://api.example.com/health"
  protocol             = "http"
  check_frequency      = 60
  expected_status_code = "200"

  regions = [
    "london",
    "virginia",
    "singapore",
  ]
}
```

### 2. import.sh

Executable bash script for importing resources:
- Prerequisite checks
- Color-coded output
- Error handling
- Summary statistics

Note: Contains placeholder UUIDs that need to be updated after resource creation.

### 3. migration-report.json

Detailed JSON report containing:
- Summary statistics
- Monitor mappings
- Healthcheck mappings
- Conversion issues with severity levels

Example:
```json
{
  "summary": {
    "total_monitors": 15,
    "converted_monitors": 15,
    "total_heartbeats": 5,
    "converted_healthchecks": 5,
    "total_issues": 3,
    "critical_issues": 0,
    "warnings": 3
  },
  "monitors": [...],
  "healthchecks": [...],
  "conversion_issues": [...]
}
```

### 4. manual-steps.md

Documentation of:
- Critical issues requiring immediate attention
- Warnings for review
- Import script update instructions
- Notification setup steps
- Status page migration steps
- Testing procedures

## Conversion Logic

### Monitor Types

| Better Stack | Hyperping | Notes |
|--------------|-----------|-------|
| `status` | `http` | Direct mapping |
| `tcp` | `port` | Port number preserved |
| `ping` | `icmp` | Direct mapping |
| `keyword` | `http` | Warning issued, manual review needed |
| `heartbeat` | `healthcheck` | Converted to healthcheck resource |

### Region Mapping

| Better Stack | Hyperping |
|--------------|-----------|
| `us`, `us-east-1` | `virginia` |
| `us-west-1` | `oregon` |
| `eu`, `eu-west-1` | `london` |
| `eu-central-1` | `frankfurt` |
| `ap-southeast-1` | `singapore` |
| `ap-northeast-1` | `tokyo` |
| `au-southeast` | `sydney` |
| `sa-east-1` | `saopaulo` |

### Frequency Normalization

Unsupported frequencies are rounded to the nearest supported value:

**Supported values**: 10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400 seconds

Examples:
- 45s → 60s
- 90s → 60s
- 240s → 300s
- 7200s → 3600s

### Request Headers

HTTP request headers are converted to Hyperping format:

```hcl
request_headers = [
  {
    name  = "Authorization"
    value = "Bearer token"
  },
  {
    name  = "X-Custom-Header"
    value = "value"
  }
]
```

### Heartbeats to Healthchecks

Better Stack heartbeats with simple period values are converted to Hyperping healthchecks with cron expressions:

| Period | Cron Expression |
|--------|----------------|
| 60s | `* * * * *` (every minute) |
| 300s | `*/5 * * * *` (every 5 minutes) |
| 3600s | `0 * * * *` (hourly) |
| 86400s | `0 0 * * *` (daily) |

## Migration Workflow

### 1. Export and Convert

```bash
# Run migration tool
migrate-betterstack --verbose

# Review generated files
cat migrated-resources.tf
cat manual-steps.md
cat migration-report.json
```

### 2. Review Configuration

- Check `migrated-resources.tf` for accuracy
- Review conversion notes in comments
- Adjust check frequencies for cost optimization
- Update environment-specific values

### 3. Initialize Terraform

```bash
cd <directory-with-tf-file>
terraform init
```

### 4. Plan Changes

```bash
terraform plan
```

Review the plan to ensure all resources will be created correctly.

### 5. Apply Configuration

```bash
terraform apply
```

### 6. Configure Notifications

Set up notification channels in Hyperping dashboard:
- Email addresses
- Slack webhooks
- PagerDuty integration

### 7. Parallel Testing

Run both Better Stack and Hyperping in parallel for 1-2 weeks:
- Compare uptime metrics
- Verify alert delivery
- Test status pages
- Validate healthcheck pings

### 8. Cutover

Once confident:
1. Pause Better Stack monitors
2. Monitor Hyperping for 48 hours
3. Cancel Better Stack subscription
4. Update documentation

## Error Handling

### Conversion Issues

The tool categorizes issues by severity:

**Critical (errors)**: Must be fixed before applying
- Unsupported monitor types
- Invalid configurations

**Warnings**: Review recommended but not blocking
- Rounded frequencies
- Simplified configurations
- Feature limitations

### API Errors

Graceful handling of:
- Authentication failures
- Rate limiting
- Network timeouts
- Invalid responses

### Partial Migration

The tool supports partial migration:
- Continues on non-critical errors
- Reports all issues in migration report
- Allows manual fixes before apply

## Testing

### Run Unit Tests

```bash
go test ./cmd/migrate-betterstack/...
```

### Test Conversion Logic

```bash
go test ./cmd/migrate-betterstack/converter -v
```

### Test with Mock Data

Create a test Better Stack export:

```json
{
  "data": [
    {
      "id": "123",
      "type": "monitor",
      "attributes": {
        "pronounceable_name": "Test Monitor",
        "url": "https://example.com",
        "monitor_type": "status",
        "check_frequency": 60,
        "request_method": "GET",
        "expected_status_codes": [200],
        "follow_redirects": true,
        "regions": ["us-east-1"]
      }
    }
  ]
}
```

## Troubleshooting

### Missing API Token

```
Error: Better Stack API token is required
```

**Solution**: Set `BETTERSTACK_API_TOKEN` environment variable or use `--betterstack-token` flag.

### Terraform Validation Failed

```
Error: Terraform validation failed
```

**Solution**:
1. Run `terraform init` first
2. Check generated configuration syntax
3. Review error messages from terraform

### Unsupported Monitor Type

```
Warning: Unknown monitor type 'udp', defaulting to 'http'
```

**Solution**: Review the monitor in `manual-steps.md` and configure manually in Hyperping.

### Region Mapping Issues

```
Warning: No regions specified, using default regions
```

**Solution**: Review region configuration in generated Terraform and adjust as needed.

## Advanced Usage

### Custom Region Mapping

Modify `converter/monitor.go` to add custom region mappings:

```go
regionMap: map[string]string{
    "custom-region": "virginia",
    // ... other mappings
}
```

### Custom Frequency Mapping

Modify `converter/monitor.go` to add custom frequency mappings:

```go
frequencyMap: map[int]int{
    45: 60,  // Round 45s to 60s
    // ... other mappings
}
```

## Best Practices

1. **Test First**: Always run with `--dry-run` first
2. **Review Carefully**: Check all generated files before applying
3. **Parallel Run**: Keep Better Stack active during initial testing
4. **Backup**: Export Better Stack data before decommissioning
5. **Document**: Update team documentation with new Hyperping URLs
6. **Monitor**: Watch for false positives after cutover

## Limitations

### Not Automatically Migrated

- **Status pages**: Must be created manually or with separate Terraform config
- **Notification channels**: Configure in Hyperping dashboard
- **On-call schedules**: Not supported in Hyperping, use PagerDuty
- **Team members**: Add manually in Hyperping dashboard
- **Historical data**: Export from Better Stack for records

### Feature Gaps

- **UDP monitoring**: Not supported in Hyperping
- **Keyword monitoring**: Limited support, uses HTTP with status code check
- **Domain expiration**: Not supported in Hyperping
- **SMS alerts**: Not supported in Hyperping
- **Phone call alerts**: Not supported in Hyperping

## Support

For issues or questions:

1. Check the migration guide: `docs/guides/migrate-from-betterstack.md`
2. Review `manual-steps.md` for common issues
3. Check migration report for specific problems
4. Open an issue on GitHub

## License

Copyright (c) 2026 Develeap
SPDX-License-Identifier: MPL-2.0
