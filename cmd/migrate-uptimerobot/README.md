# UptimeRobot to Hyperping Migration Tool

Automated CLI tool for migrating monitoring configurations from UptimeRobot to Hyperping using Terraform.

## Features

- **Automatic Conversion:** Converts all UptimeRobot monitor types to Hyperping equivalents
- **Terraform Generation:** Generates ready-to-use Terraform configurations
- **Import Scripts:** Creates shell scripts for importing existing resources
- **Detailed Reports:** JSON reports with warnings and migration statistics
- **Manual Steps Guide:** Documentation for manual configuration requirements
- **Alert Contact Mapping:** Categorizes alert contacts for escalation policy setup

## Supported Monitor Types

| UptimeRobot Type | Hyperping Resource | Notes |
|------------------|-------------------|-------|
| HTTP/HTTPS (Type 1) | `hyperping_monitor` (protocol=http) | Full support |
| Keyword (Type 2) | `hyperping_monitor` (with required_keyword) | "Exists" checks only |
| Ping (Type 3) | `hyperping_monitor` (protocol=icmp) | Full support |
| Port (Type 4) | `hyperping_monitor` (protocol=port) | Full support with port mapping |
| Heartbeat (Type 5) | `hyperping_healthcheck` | Converted to healthcheck with cron/period |

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/develeap/terraform-provider-hyperping.git
cd terraform-provider-hyperping

# Build the migration tool
go build -o migrate-uptimerobot ./cmd/migrate-uptimerobot

# Or run directly
go run ./cmd/migrate-uptimerobot [options]
```

### Using Go Install

```bash
go install github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot@latest
```

## Quick Start

### 1. Set API Keys

```bash
export UPTIMEROBOT_API_KEY="u1234567-abc..."
export HYPERPING_API_KEY="sk_your_key_here"
```

### 2. Validate Your Monitors

```bash
migrate-uptimerobot -validate
```

Output:
```
Validating UptimeRobot monitors...

Monitor Types:
  HTTP/HTTPS: 15
  Keyword: 3
  Ping (ICMP): 2
  Port: 5
  Heartbeat: 2

Total monitors: 27
Alert contacts: 8

Validation complete.
```

### 3. Perform Dry Run

```bash
migrate-uptimerobot -dry-run -verbose
```

### 4. Generate Migration Files

```bash
migrate-uptimerobot \
  -output=hyperping.tf \
  -import-script=import.sh \
  -report=migration-report.json \
  -manual-steps=manual-steps.md
```

Output files:
- `hyperping.tf` - Terraform configuration
- `import.sh` - Import script (executable)
- `migration-report.json` - Detailed JSON report
- `manual-steps.md` - Manual configuration guide

### 5. Review and Apply

```bash
# Initialize Terraform
terraform init

# Review the plan
terraform plan

# Apply the configuration
terraform apply
```

## Command-Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `-uptimerobot-api-key` | UptimeRobot API key | `$UPTIMEROBOT_API_KEY` |
| `-hyperping-api-key` | Hyperping API key | `$HYPERPING_API_KEY` |
| `-output` | Terraform configuration file | `hyperping.tf` |
| `-import-script` | Import script file | `import.sh` |
| `-report` | Migration report file | `migration-report.json` |
| `-manual-steps` | Manual steps documentation | `manual-steps.md` |
| `-dry-run` | Preview without creating files | `false` |
| `-validate` | Validate monitors only | `false` |
| `-verbose` | Enable verbose output | `false` |

## Migration Workflow

### Phase 1: Planning (Day 1)

1. **Export UptimeRobot Configuration**
   ```bash
   migrate-uptimerobot -validate -verbose
   ```

2. **Review Monitor Types**
   - Check supported vs. unsupported types
   - Note frequency adjustments
   - Identify keyword monitors with "not exists" checks

3. **Plan Escalation Policies**
   - Group alert contacts by severity
   - Design escalation policies in Hyperping

### Phase 2: Setup (Day 2-3)

1. **Create Escalation Policies**
   - Log into Hyperping dashboard
   - Navigate to Settings → Escalation Policies
   - Create policies matching your UptimeRobot alert contacts
   - Note the escalation policy UUIDs

2. **Generate Migration Files**
   ```bash
   migrate-uptimerobot
   ```

3. **Review Generated Configuration**
   - Check `hyperping.tf` for accuracy
   - Adjust check frequencies if needed
   - Update escalation policy UUIDs

### Phase 3: Parallel Testing (Week 1-2)

1. **Apply Terraform Configuration**
   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

2. **Update Healthcheck Scripts**
   ```bash
   # Get ping URLs
   terraform output -json | grep ping_url

   # Update your scripts with new URLs
   ```

3. **Run Both Systems**
   - Keep UptimeRobot active
   - Monitor Hyperping for 1-2 weeks
   - Compare alerting behavior

### Phase 4: Cutover (Week 3)

1. **Enable Escalation Policies**
   - Uncomment `escalation_policy` lines in `hyperping.tf`
   - Apply changes:
     ```bash
     terraform apply
     ```

2. **Update Documentation**
   - Runbooks
   - On-call procedures
   - Team wiki

3. **Pause UptimeRobot Monitors**
   - Via API or dashboard
   - Keep as backup for 1 week

### Phase 5: Cleanup (Week 4)

1. **Verify Hyperping Stability**
   ```bash
   terraform plan  # Should show no changes
   ```

2. **Archive UptimeRobot Data**
   ```bash
   curl -X POST https://api.uptimerobot.com/v2/getMonitors \
     -d "api_key=$UPTIMEROBOT_API_KEY" \
     > uptimerobot-backup.json
   ```

3. **Cancel UptimeRobot Subscription**

## Monitor Type Conversions

### HTTP/HTTPS Monitors

**UptimeRobot:**
```json
{
  "id": 12345,
  "friendly_name": "API Health",
  "url": "https://api.example.com/health",
  "type": 1,
  "interval": 60,
  "http_method": 1
}
```

**Hyperping:**
```hcl
resource "hyperping_monitor" "api_health" {
  name                 = "API Health"
  url                  = "https://api.example.com/health"
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 60
  expected_status_code = "2xx"
  regions              = ["london", "virginia", "singapore"]
}
```

### Keyword Monitors

**UptimeRobot:**
```json
{
  "id": 12346,
  "friendly_name": "Status Check",
  "url": "https://example.com",
  "type": 2,
  "keyword_type": 1,
  "keyword_value": "healthy",
  "interval": 300
}
```

**Hyperping:**
```hcl
resource "hyperping_monitor" "status_check" {
  name                 = "Status Check"
  url                  = "https://example.com"
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 300
  expected_status_code = "200"
  required_keyword     = "healthy"
  regions              = ["london", "virginia"]
}
```

### Ping Monitors

**UptimeRobot:**
```json
{
  "id": 12347,
  "friendly_name": "Server Ping",
  "url": "192.168.1.100",
  "type": 3,
  "interval": 60
}
```

**Hyperping:**
```hcl
resource "hyperping_monitor" "server_ping" {
  name            = "Server Ping"
  url             = "192.168.1.100"
  protocol        = "icmp"
  check_frequency = 60
  regions         = ["london", "virginia"]
}
```

### Port Monitors

**UptimeRobot:**
```json
{
  "id": 12348,
  "friendly_name": "PostgreSQL",
  "url": "db.example.com",
  "type": 4,
  "port": 5432,
  "interval": 120
}
```

**Hyperping:**
```hcl
resource "hyperping_monitor" "postgresql" {
  name            = "PostgreSQL"
  url             = "db.example.com"
  protocol        = "port"
  port            = 5432
  check_frequency = 120
  regions         = ["virginia"]
}
```

### Heartbeat Monitors

**UptimeRobot:**
```json
{
  "id": 12349,
  "friendly_name": "Daily Backup",
  "type": 5,
  "interval": 86400
}
```

**Hyperping:**
```hcl
resource "hyperping_healthcheck" "daily_backup" {
  name               = "Daily Backup"
  period_value       = 1
  period_type        = "days"
  grace_period_value = 1
  grace_period_type  = "hours"
}

output "daily_backup_ping_url" {
  value     = hyperping_healthcheck.daily_backup.ping_url
  sensitive = true
}
```

**Update Script:**
```bash
# Old UptimeRobot heartbeat
curl "https://uptimerobot.com/api/heartbeat/12349"

# New Hyperping healthcheck
PING_URL=$(terraform output -raw daily_backup_ping_url)
curl -fsS --retry 3 "$PING_URL"
```

## Field Mappings

### HTTP Methods

| UptimeRobot | Value | Hyperping |
|-------------|-------|-----------|
| GET | 1 | "GET" |
| POST | 2 | "POST" |
| PUT | 3 | "PUT" |
| PATCH | 4 | "PATCH" |
| DELETE | 5 | "DELETE" |
| HEAD | 6 | "HEAD" |

### Check Frequencies

UptimeRobot intervals are mapped to nearest allowed Hyperping values:

| UptimeRobot | Hyperping | Notes |
|-------------|-----------|-------|
| 60s | 60s | Direct match |
| 120s | 120s | Direct match |
| 300s | 300s | Direct match |
| 45s | 60s | Rounded up |
| 90s | 60s | Rounded down |
| 150s | 180s | Rounded up |

Allowed Hyperping values: `10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400`

### Port Sub-Types

| UptimeRobot Sub-Type | Service | Port |
|---------------------|---------|------|
| 1 | Custom | User-defined |
| 2 | HTTP | 80 |
| 3 | HTTPS | 443 |
| 4 | FTP | 21 |
| 5 | SMTP | 25 |
| 6 | POP3 | 110 |
| 7 | IMAP | 143 |

## Alert Contact Migration

Alert contacts must be manually configured as escalation policies in Hyperping.

### Alert Contact Types

| Type | UptimeRobot | Hyperping Equivalent |
|------|-------------|---------------------|
| 2 | Email | Escalation Policy → Email |
| 3 | SMS | Escalation Policy → SMS |
| 4 | Webhook | Escalation Policy → Webhook |
| 11 | Slack | Integration + Escalation Policy |
| 14 | PagerDuty | Integration + Escalation Policy |

### Setup Process

1. Run migration tool to extract alert contacts
2. Review `manual-steps.md` for contact list
3. Create escalation policies in Hyperping dashboard
4. Update `terraform.tfvars` with escalation policy UUIDs
5. Uncomment `escalation_policy` lines in resources
6. Apply Terraform configuration

## Troubleshooting

### Issue: API Key Invalid

```
Error fetching monitors: API error: invalid_parameter - api_key is invalid
```

**Solution:**
- Verify `UPTIMEROBOT_API_KEY` is set correctly
- Check API key in UptimeRobot dashboard (Integrations & API)
- Ensure key has read permissions

### Issue: Keyword "Not Exists" Check

```
Warning: Keyword check 'must not exist' is not supported by Hyperping
```

**Solution:**
Options for handling:
1. Modify endpoint to return different status codes
2. Use status code checks instead
3. Create validation proxy that transforms response

### Issue: Frequency Adjusted

```
Warning: Check frequency adjusted from 45s to 60s (nearest allowed value)
```

**Action:**
- Review the adjusted frequency in `hyperping.tf`
- Adjust if needed for your SLAs
- Consider cost implications of higher frequencies

### Issue: Heartbeat Not Working

```
Healthcheck showing as down after migration
```

**Solution:**
1. Get the ping URL:
   ```bash
   terraform output -raw healthcheck_name_ping_url
   ```
2. Update your script with new URL
3. Test manually:
   ```bash
   curl -v "https://ping.hyperping.io/hc_xxx"
   ```
4. Check Hyperping dashboard for successful ping

## Migration Report

The JSON report includes:

```json
{
  "timestamp": "2026-02-13T10:30:00Z",
  "summary": {
    "total_monitors": 27,
    "migrated_monitors": 22,
    "migrated_healthchecks": 3,
    "skipped_monitors": 2,
    "monitors_with_warnings": 5
  },
  "monitors": [
    {
      "original_id": 12345,
      "original_name": "API Health",
      "original_type": 1,
      "resource_type": "hyperping_monitor",
      "resource_name": "api_health",
      "migration_status": "migrated",
      "warnings": []
    }
  ],
  "warnings": [],
  "errors": []
}
```

## Best Practices

### Before Migration

- [ ] Export and backup UptimeRobot configuration
- [ ] Document current alert routing
- [ ] Create test Hyperping account
- [ ] Plan escalation policies

### During Migration

- [ ] Use `-validate` flag first
- [ ] Review all generated files
- [ ] Test with a small subset first
- [ ] Run both systems in parallel

### After Migration

- [ ] Verify all monitors are working
- [ ] Test escalation policies
- [ ] Update team documentation
- [ ] Archive UptimeRobot data before canceling

## Examples

See the [migration guide](../../docs/guides/migrate-from-uptimerobot.md) for complete examples.

## Support

- **Documentation:** [Migration Guide](../../docs/guides/migrate-from-uptimerobot.md)
- **Provider Docs:** [Terraform Registry](https://registry.terraform.io/providers/develeap/hyperping)
- **Issues:** [GitHub Issues](https://github.com/develeap/terraform-provider-hyperping/issues)
- **Hyperping Support:** support@hyperping.io

## License

Mozilla Public License 2.0 - See [LICENSE](../../LICENSE) for details.

---

**Sources:**
- [UptimeRobot Official API](https://uptimerobot.com/api/)
- [UptimeRobot API Documentation](https://uptimerobot.com/api/legacy/)
- [GitHub - bitfield/uptimerobot Client Library](https://github.com/bitfield/uptimerobot)
