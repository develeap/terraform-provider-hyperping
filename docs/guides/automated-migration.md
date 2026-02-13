---
page_title: "Automated Migration Tools"
subcategory: "Guides"
description: |-
  Complete guide to using automated CLI tools for migrating from Better Stack, UptimeRobot, and Pingdom to Hyperping.
---

# Automated Migration Tools

This guide covers the automated CLI tools that streamline migrating monitoring infrastructure from Better Stack, UptimeRobot, and Pingdom to Hyperping. These tools eliminate manual conversion work by automatically exporting, converting, and generating ready-to-use Terraform configurations.

## Table of Contents

- [Overview](#overview)
- [When to Use Automated vs Manual Migration](#when-to-use-automated-vs-manual-migration)
- [Prerequisites](#prerequisites)
- [Common Workflow](#common-workflow)
- [Tool-Specific Guides](#tool-specific-guides)
  - [Better Stack Migration Tool](#better-stack-migration-tool)
  - [UptimeRobot Migration Tool](#uptimerobot-migration-tool)
  - [Pingdom Migration Tool](#pingdom-migration-tool)
- [Output Files Explained](#output-files-explained)
- [Troubleshooting](#troubleshooting)
- [Advanced Usage](#advanced-usage)
- [Best Practices](#best-practices)
- [FAQ](#faq)

## Overview

The Hyperping migration tools provide end-to-end automation for platform migrations:

### Available Tools

| Tool | Source Platform | Status | Command |
|------|----------------|--------|---------|
| **migrate-betterstack** | Better Stack / Better Uptime | Available | `migrate-betterstack` |
| **migrate-uptimerobot** | UptimeRobot | Available | `migrate-uptimerobot` |
| **migrate-pingdom** | Pingdom | Available | `migrate-pingdom` |

### What Gets Automated

All tools provide the following automated capabilities:

| Step | Manual Process | Automated Process |
|------|---------------|-------------------|
| **Export** | Manual API calls, JSON parsing | Single command exports all resources |
| **Conversion** | Write Python/shell scripts | Built-in smart conversion logic |
| **Mapping** | Manual field mapping | Automatic field and type mapping |
| **Generation** | Hand-write Terraform HCL | Generate production-ready .tf files |
| **Import Scripts** | Write bash import scripts | Generate executable import.sh |
| **Validation** | Manual testing | Built-in validation checks |
| **Reporting** | Track changes manually | Detailed migration report |

### Key Benefits

**Time Savings:**
- **Manual migration:** 5-9 hours for 50-100 monitors
- **Automated migration:** 15-30 minutes for same workload
- **Reduction:** ~90% time savings

**Accuracy:**
- Eliminate human error in field mapping
- Consistent conversion logic
- Validation catches issues before apply

**Maintainability:**
- Generated code follows Terraform best practices
- Consistent naming conventions
- Proper resource dependencies

## When to Use Automated vs Manual Migration

### Use Automated Tools When

✅ **You have many monitors** (10+ resources)
- Automation shines with bulk operations
- Time savings compound with scale

✅ **Standard configurations**
- HTTP/HTTPS monitors
- TCP port checks
- Standard alert contacts
- Common check frequencies

✅ **Need quick turnaround**
- Production migrations with tight timelines
- Multiple environment migrations
- Rapid testing and validation

✅ **Want consistency**
- Multiple teams migrating
- Standardized naming conventions
- Repeatable process

### Use Manual Migration When

⚠️ **Complex custom configurations**
- Highly customized alerting logic
- Non-standard integrations
- Complex webhook configurations

⚠️ **Few resources** (<5 monitors)
- Manual may be faster for very small deployments
- More control over exact configuration

⚠️ **Need to learn**
- Understanding migration process
- Learning Terraform patterns
- Educational purposes

⚠️ **Unsupported features**
- Platform-specific features without Hyperping equivalent
- Require custom logic beyond automated mapping

### Hybrid Approach (Recommended)

The best approach often combines both:

1. **Use automated tool** for initial bulk migration
2. **Manually review** and customize generated configurations
3. **Extend automation** for specific needs
4. **Validate thoroughly** before cutover

## Prerequisites

### Required Software

Install these tools before using migration utilities:

```bash
# 1. Terraform (version 1.8 or higher)
terraform version
# Terraform v1.8.0
# on linux_amd64

# 2. Go (version 1.24 or higher) - for building tools
go version
# go version go1.24.0 linux/amd64

# 3. jq (for JSON processing)
jq --version
# jq-1.6

# 4. curl (for API calls)
curl --version
# curl 7.68.0
```

### API Keys

**Source Platform API Keys:**

| Platform | Where to Get API Key |
|----------|---------------------|
| **Better Stack** | [betteruptime.com/users/sign_in](https://betteruptime.com/users/sign_in) → Settings → API Tokens |
| **UptimeRobot** | [uptimerobot.com/dashboard](https://uptimerobot.com/dashboard) → My Settings → API Settings |
| **Pingdom** | [my.pingdom.com](https://my.pingdom.com) → Settings → API Keys |

**Hyperping API Key:**

Generate at [app.hyperping.io/settings/api-keys](https://app.hyperping.io/settings/api-keys)

### Environment Setup

```bash
# Set source platform API key (choose one)
export BETTERSTACK_API_TOKEN="your_betterstack_token"
export UPTIMEROBOT_API_KEY="your_uptimerobot_key"
export PINGDOM_API_KEY="your_pingdom_key"

# Set destination Hyperping API key
export HYPERPING_API_KEY="sk_your_hyperping_key"

# Verify environment
env | grep -E "(BETTERSTACK|UPTIMEROBOT|PINGDOM|HYPERPING)"
```

### Installation

#### Install from Binary (Recommended)

Download pre-built binaries from GitHub releases:

```bash
# Download latest release for Linux
curl -L -o migrate-betterstack \
  https://github.com/develeap/terraform-provider-hyperping/releases/latest/download/migrate-betterstack-linux-amd64
chmod +x migrate-betterstack
sudo mv migrate-betterstack /usr/local/bin/

# Download for macOS (Apple Silicon)
curl -L -o migrate-betterstack \
  https://github.com/develeap/terraform-provider-hyperping/releases/latest/download/migrate-betterstack-darwin-arm64
chmod +x migrate-betterstack
sudo mv migrate-betterstack /usr/local/bin/

# Verify installation
migrate-betterstack --version
# migrate-betterstack version 1.0.0
```

Repeat for other tools (`migrate-uptimerobot`, `migrate-pingdom`).

#### Build from Source

```bash
# Clone repository
git clone https://github.com/develeap/terraform-provider-hyperping.git
cd terraform-provider-hyperping

# Build Better Stack migration tool
go build -o migrate-betterstack ./cmd/migrate-betterstack
sudo mv migrate-betterstack /usr/local/bin/

# Build UptimeRobot migration tool
go build -o migrate-uptimerobot ./cmd/migrate-uptimerobot
sudo mv migrate-uptimerobot /usr/local/bin/

# Build Pingdom migration tool
go build -o migrate-pingdom ./cmd/migrate-pingdom
sudo mv migrate-pingdom /usr/local/bin/

# Verify
migrate-betterstack --version
migrate-uptimerobot --version
migrate-pingdom --version
```

## Common Workflow

All migration tools follow the same workflow pattern:

### Step 1: Export from Source Platform

```bash
# Run migration tool with export flag
migrate-betterstack export

# Output:
# ✓ Fetching monitors from Better Stack...
# ✓ Found 47 monitors
# ✓ Fetching status pages...
# ✓ Found 3 status pages
# ✓ Fetching incidents...
# ✓ Found 12 incidents
# ✓ Exported to betterstack-export.json
```

**What happens:**
- Connects to source platform API
- Fetches all monitors, status pages, incidents, etc.
- Saves raw data to JSON file
- Validates API responses
- Reports summary

**Output files:**
- `{platform}-export.json` - Raw exported data
- `{platform}-export.log` - Export log with timestamps

### Step 2: Convert to Hyperping Format

```bash
# Run conversion
migrate-betterstack convert \
  --input betterstack-export.json \
  --output hyperping-config

# Output:
# ✓ Loading Better Stack export...
# ✓ Converting 47 monitors...
#   - API Gateway Health Check → hyperping_monitor.api_gateway
#   - Database Connection Test → hyperping_monitor.database_connection
#   ...
# ✓ Converting 3 status pages...
# ✓ Generating Terraform configuration...
# ✓ Validating generated configuration...
# ✓ Success! Generated files in hyperping-config/
```

**What happens:**
- Parses exported JSON
- Maps fields to Hyperping schema
- Applies naming conventions
- Handles special cases
- Validates conversions
- Generates Terraform HCL

**Output files:**
- `hyperping-config/migrated-resources.tf` - Terraform resources
- `hyperping-config/variables.tf` - Input variables
- `hyperping-config/outputs.tf` - Output values
- `hyperping-config/terraform.tfvars` - Variable values

### Step 3: Review Generated Files

```bash
# Review generated Terraform
cd hyperping-config
cat migrated-resources.tf

# Validate Terraform syntax
terraform fmt
terraform validate

# Plan (dry-run)
terraform plan
```

**What to review:**
- Resource names are meaningful
- Monitor URLs are correct
- Check frequencies are appropriate
- Regions are properly mapped
- Alert contacts are included

### Step 4: Execute Imports

```bash
# Run generated import script
./import.sh

# Output:
# ✓ Importing hyperping_monitor.api_gateway...
# ✓ Importing hyperping_monitor.database_connection...
# ✓ Importing hyperping_statuspage.main...
# ✓ Successfully imported 50 resources
```

**What happens:**
- Creates resources in Hyperping
- Imports into Terraform state
- Validates each import
- Reports progress
- Handles errors gracefully

### Step 5: Validate (Zero Drift)

```bash
# Verify no configuration drift
terraform plan

# Expected output:
# No changes. Your infrastructure matches the configuration.
```

**If drift detected:**
- Review differences
- Update configuration as needed
- Re-run plan until clean

### Step 6: Decommission Old Platform

```bash
# Generate decommission report
migrate-betterstack decommission-report

# Output: decommission-checklist.md with:
# - Resources to delete
# - Verification steps
# - Rollback plan
```

**Decommission steps:**
1. Pause old monitors (don't delete yet)
2. Monitor for 48-72 hours
3. Verify Hyperping alerts working
4. Export historical data
5. Cancel old platform subscription
6. Delete old resources after 30-day safety period

## Tool-Specific Guides

### Better Stack Migration Tool

#### Overview

Migrates from Better Stack (formerly Better Uptime) to Hyperping.

**Supported Resources:**
- ✅ HTTP/HTTPS monitors
- ✅ TCP port monitors
- ✅ Heartbeat monitors → Healthchecks
- ✅ Status pages
- ✅ Incidents
- ✅ Maintenance windows
- ⚠️ On-call schedules (manual setup required)
- ⚠️ Escalation policies (manual setup required)

#### Quick Start

```bash
# Full automated migration
migrate-betterstack migrate \
  --source-api-key $BETTERSTACK_API_TOKEN \
  --dest-api-key $HYPERPING_API_KEY \
  --output ./hyperping-migration

# Output structure:
# hyperping-migration/
# ├── migrated-resources.tf
# ├── variables.tf
# ├── outputs.tf
# ├── terraform.tfvars
# ├── import.sh
# ├── migration-report.json
# └── manual-steps.md
```

#### Step-by-Step Usage

**1. Export from Better Stack:**

```bash
migrate-betterstack export \
  --api-key $BETTERSTACK_API_TOKEN \
  --output betterstack-export.json \
  --include monitors,statuspages,incidents,maintenance

# Options:
#   --include: comma-separated resource types
#   --exclude: resource types to skip
#   --filter: filter by tags or names
#   --output: output file path
```

**2. Preview conversion:**

```bash
migrate-betterstack convert \
  --input betterstack-export.json \
  --dry-run \
  --show-mapping

# Shows conversion mapping:
# Better Stack → Hyperping
# ============================
# Monitor: API Health
#   pronounceable_name → name
#   url → url
#   check_frequency: 60s → 60s
#   monitor_type: status → protocol: http
#   ...
```

**3. Convert to Hyperping format:**

```bash
migrate-betterstack convert \
  --input betterstack-export.json \
  --output hyperping-config \
  --prefix prod_ \
  --naming-convention kebab-case \
  --validate

# Options:
#   --prefix: prefix for resource names
#   --naming-convention: snake-case, kebab-case, or camelCase
#   --validate: validate after conversion
#   --continue-on-error: continue if some resources fail
```

**4. Review and customize:**

```bash
cd hyperping-config

# Review generated configuration
cat migrated-resources.tf

# Customize as needed
vim migrated-resources.tf

# Validate
terraform init
terraform validate
terraform plan
```

**5. Execute migration:**

```bash
# Import existing resources (if already created)
./import.sh

# Or create new resources
terraform apply

# Verify
terraform plan
# No changes expected
```

#### Feature Mapping

| Better Stack Feature | Hyperping Equivalent | Notes |
|---------------------|---------------------|-------|
| **Monitors** | | |
| HTTP/HTTPS status | `hyperping_monitor` | Direct mapping |
| TCP port | `hyperping_monitor` (protocol: port) | Direct mapping |
| Heartbeat | `hyperping_healthcheck` | Webhook URL changes |
| Keyword matching | `expected_status_code` + body | Limited support |
| SSL monitoring | Auto-monitored | Automatic in Hyperping |
| **Check Settings** | | |
| Check frequency | `check_frequency` | Must be supported value |
| Timeout | Smart timeout | Hyperping manages automatically |
| Regions | `regions` | Map region names |
| Request headers | `request_headers` | Direct mapping |
| Request body | `request_body` | Direct mapping |
| **Status Pages** | | |
| Status page | `hyperping_statuspage` | Direct mapping |
| Components | `sections[].services` | Different structure |
| Subscribers | `hyperping_statuspage_subscriber` | Per-subscriber resource |
| Custom domain | `custom_domain` | Requires DNS setup |
| **Incidents** | | |
| Incident | `hyperping_incident` | Multi-language support |
| Updates | `hyperping_incident.updates` | Nested in resource |
| Affected monitors | `affected_monitors` | UUID array |
| **Maintenance** | | |
| Maintenance window | `hyperping_maintenance` | Direct mapping |
| Scheduled time | `scheduled_start/end` | ISO 8601 format |
| **Alerting** | | |
| Email notifications | Status page subscribers | Different approach |
| Slack integration | Webhook configuration | Manual setup |
| PagerDuty | Webhook configuration | Manual setup |
| On-call schedules | Not supported | Use external tool |

#### Conversion Examples

**Monitor Conversion:**

```hcl
# Better Stack (API response)
{
  "attributes": {
    "pronounceable_name": "API Gateway Health",
    "url": "https://api.example.com/health",
    "monitor_type": "status",
    "check_frequency": 60,
    "request_method": "GET",
    "expected_status_codes": [200, 201],
    "monitor_group_id": ["us", "eu"]
  }
}

# Generated Hyperping Terraform
resource "hyperping_monitor" "api_gateway_health" {
  name                 = "API Gateway Health"
  url                  = "https://api.example.com/health"
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 60
  expected_status_code = "2xx"  # Converted from [200, 201]
  regions              = ["virginia", "london"]  # Mapped from ["us", "eu"]
  follow_redirects     = true
  paused               = false
}
```

**Heartbeat to Healthcheck:**

```hcl
# Better Stack Heartbeat
{
  "attributes": {
    "name": "Daily Backup Job",
    "period": 86400,  # 24 hours
    "grace": 3600     # 1 hour grace
  }
}

# Generated Hyperping Healthcheck
resource "hyperping_healthcheck" "daily_backup_job" {
  name               = "Daily Backup Job"
  cron               = "0 0 * * *"  # Converted from period
  timezone           = "UTC"
  grace_period_value = 60  # Converted to minutes
  grace_period_type  = "minutes"
  paused             = false
}

# Output includes new webhook URL
output "daily_backup_webhook" {
  description = "Update your backup script to use this URL"
  value       = hyperping_healthcheck.daily_backup_job.ping_url
  sensitive   = true
}
```

#### Troubleshooting

**Issue: API Rate Limiting**

```
Error: Rate limit exceeded (HTTP 429)
```

**Solution:**

```bash
# Add delay between API calls
migrate-betterstack export \
  --api-key $BETTERSTACK_API_TOKEN \
  --rate-limit-delay 1000  # 1 second between requests

# Or use batch mode
migrate-betterstack export \
  --api-key $BETTERSTACK_API_TOKEN \
  --batch-size 10  # Process 10 at a time
```

**Issue: Unsupported Monitor Type**

```
Warning: Monitor 'dns-check' uses unsupported type 'dns'
```

**Solution:**

```bash
# Skip unsupported monitors
migrate-betterstack convert \
  --input betterstack-export.json \
  --skip-unsupported

# Or convert to closest equivalent
migrate-betterstack convert \
  --input betterstack-export.json \
  --convert-unsupported http  # Convert DNS to HTTP check
```

**Issue: Check Frequency Not Supported**

```
Warning: Frequency 45s not supported, rounding to 60s
```

**Solution:**

```bash
# See rounding decisions in report
cat migration-report.json | jq '.warnings[] | select(.type == "frequency_rounded")'

# Manually adjust after generation
vim hyperping-config/migrated-resources.tf
# Change check_frequency to desired value (10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600)
```

### UptimeRobot Migration Tool

#### Overview

Migrates from UptimeRobot to Hyperping.

**Supported Resources:**
- ✅ HTTP/HTTPS monitors
- ✅ Port monitors
- ✅ Keyword monitors
- ✅ Ping monitors → HTTP monitors (workaround)
- ✅ Alert contacts
- ⚠️ Status pages (requires Public Status Pages addon)
- ⚠️ Maintenance windows (manual creation)

#### Quick Start

```bash
# Full automated migration
migrate-uptimerobot migrate \
  --source-api-key $UPTIMEROBOT_API_KEY \
  --dest-api-key $HYPERPING_API_KEY \
  --output ./hyperping-migration

# With filtering
migrate-uptimerobot migrate \
  --source-api-key $UPTIMEROBOT_API_KEY \
  --dest-api-key $HYPERPING_API_KEY \
  --filter "type=http,keyword" \
  --output ./hyperping-migration
```

#### Step-by-Step Usage

**1. Export from UptimeRobot:**

```bash
migrate-uptimerobot export \
  --api-key $UPTIMEROBOT_API_KEY \
  --output uptimerobot-export.json

# With alert contacts
migrate-uptimerobot export \
  --api-key $UPTIMEROBOT_API_KEY \
  --include-alert-contacts \
  --output uptimerobot-export.json
```

**2. Analyze export:**

```bash
migrate-uptimerobot analyze \
  --input uptimerobot-export.json

# Output:
# UptimeRobot Export Analysis
# ============================
# Total Monitors: 32
#   - HTTP(S): 24
#   - Port: 5
#   - Keyword: 2
#   - Ping: 1 (will convert to HTTP)
#
# Alert Contacts: 8
#   - Email: 5
#   - Webhook: 2
#   - SMS: 1 (not supported)
#
# Potential Issues:
#   ⚠ 1 Ping monitor will be converted to HTTP
#   ⚠ 1 SMS alert contact not supported
```

**3. Convert to Hyperping:**

```bash
migrate-uptimerobot convert \
  --input uptimerobot-export.json \
  --output hyperping-config \
  --prefix prod_ \
  --handle-unsupported prompt

# Options:
#   --handle-unsupported: skip, convert, or prompt
#   --ping-to-http: convert ping monitors to HTTP
#   --keyword-strategy: body-check or status-code
```

**4. Review conversion report:**

```bash
cat hyperping-config/migration-report.json | jq
```

**5. Execute migration:**

```bash
cd hyperping-config
terraform init
terraform plan
terraform apply
```

#### Feature Mapping

| UptimeRobot Feature | Hyperping Equivalent | Notes |
|---------------------|---------------------|-------|
| **Monitor Types** | | |
| HTTP(S) | `hyperping_monitor` (protocol: http) | Direct mapping |
| Port | `hyperping_monitor` (protocol: port) | Direct mapping |
| Keyword | `hyperping_monitor` + validation | Limited support |
| Ping (ICMP) | `hyperping_monitor` (protocol: icmp) | Direct mapping |
| Heartbeat | `hyperping_healthcheck` | Webhook URL changes |
| **Check Settings** | | |
| Monitoring interval | `check_frequency` | Map to supported values |
| Monitor timeout | Smart timeout | Automatic |
| HTTP method | `http_method` | Direct mapping |
| POST value | `request_body` | Direct mapping |
| Custom HTTP headers | `request_headers` | Direct mapping |
| **Alert Contacts** | | |
| Email | `hyperping_statuspage_subscriber` | Via status page |
| Webhook | Configure in dashboard | Manual setup |
| SMS | Not supported | Use external service |
| Slack | Configure in dashboard | Manual setup |
| **Status Pages** | | |
| Public status page | `hyperping_statuspage` | Requires addon in UR |
| Custom domain | `custom_domain` | Requires DNS |

#### Conversion Examples

**HTTP Monitor:**

```hcl
# UptimeRobot (API response)
{
  "friendly_name": "Website Homepage",
  "url": "https://example.com",
  "type": 1,  # HTTP(S)
  "interval": 300,  # 5 minutes
  "http_username": "",
  "http_password": "",
  "alert_contacts": [{"id": "123"}]
}

# Generated Hyperping Terraform
resource "hyperping_monitor" "website_homepage" {
  name                 = "Website Homepage"
  url                  = "https://example.com"
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 300
  expected_status_code = "200"
  follow_redirects     = true
  paused               = false
  regions              = ["virginia", "london", "singapore"]
}
```

**Keyword Monitor:**

```hcl
# UptimeRobot Keyword Monitor
{
  "friendly_name": "API Status Check",
  "url": "https://api.example.com/status",
  "type": 2,  # Keyword
  "keyword_type": 1,  # Exists
  "keyword_value": "\"status\":\"healthy\""
}

# Generated Hyperping Monitor
resource "hyperping_monitor" "api_status_check" {
  name                 = "API Status Check"
  url                  = "https://api.example.com/status"
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 300
  expected_status_code = "200"
  # Note: Keyword checking limited - verify response body manually
}

# Manual step added to manual-steps.md:
# ⚠ Verify monitor 'api_status_check' response contains: "status":"healthy"
```

**Port Monitor:**

```hcl
# UptimeRobot Port Monitor
{
  "friendly_name": "PostgreSQL Database",
  "url": "db.example.com",
  "type": 4,  # Port
  "sub_type": 5432
}

# Generated Hyperping Monitor
resource "hyperping_monitor" "postgresql_database" {
  name            = "PostgreSQL Database"
  url             = "db.example.com"
  protocol        = "port"
  port            = 5432
  check_frequency = 300
  paused          = false
  regions         = ["virginia"]  # Port checks typically single region
}
```

#### Troubleshooting

**Issue: Ping Monitor Conversion**

```
Warning: Ping monitors not directly supported
```

**Solution:**

```bash
# Convert to HTTP check (common approach)
migrate-uptimerobot convert \
  --input uptimerobot-export.json \
  --ping-strategy http \
  --ping-path /health

# Or convert to ICMP (if supported)
migrate-uptimerobot convert \
  --input uptimerobot-export.json \
  --ping-strategy icmp
```

**Issue: Keyword Monitor Limitations**

```
Warning: Keyword validation not fully supported
```

**Solution:**

```bash
# Generate as basic HTTP monitor
migrate-uptimerobot convert \
  --input uptimerobot-export.json \
  --keyword-strategy basic

# Manual verification added to manual-steps.md
# Review output and add custom validation if needed
```

### Pingdom Migration Tool

#### Overview

Migrates from Pingdom (Solarwinds) to Hyperping.

**Supported Resources:**
- ✅ HTTP/HTTPS checks
- ✅ TCP checks
- ✅ Transaction checks → HTTP monitors (basic)
- ⚠️ DNS checks (manual conversion)
- ⚠️ SMTP checks (manual conversion)

#### Quick Start

```bash
# Full automated migration
migrate-pingdom migrate \
  --source-api-key $PINGDOM_API_KEY \
  --dest-api-key $HYPERPING_API_KEY \
  --output ./hyperping-migration
```

#### Step-by-Step Usage

**1. Export from Pingdom:**

```bash
migrate-pingdom export \
  --api-key $PINGDOM_API_KEY \
  --output pingdom-export.json

# Include check details
migrate-pingdom export \
  --api-key $PINGDOM_API_KEY \
  --include-details \
  --output pingdom-export.json
```

**2. Analyze compatibility:**

```bash
migrate-pingdom analyze \
  --input pingdom-export.json \
  --show-compatibility

# Output:
# Pingdom Export Analysis
# ========================
# Total Checks: 18
#   ✅ HTTP: 12 (fully supported)
#   ✅ TCP: 4 (fully supported)
#   ⚠️  Transaction: 2 (basic conversion)
#
# Compatibility: 89% (16/18 checks)
```

**3. Convert to Hyperping:**

```bash
migrate-pingdom convert \
  --input pingdom-export.json \
  --output hyperping-config \
  --transaction-strategy simplify

# Options:
#   --transaction-strategy: simplify or skip
#   --include-integrations: convert integrations
```

**4. Review and apply:**

```bash
cd hyperping-config
terraform init
terraform plan
terraform apply
```

#### Feature Mapping

| Pingdom Feature | Hyperping Equivalent | Notes |
|-----------------|---------------------|-------|
| **Check Types** | | |
| HTTP/HTTPS | `hyperping_monitor` (protocol: http) | Direct mapping |
| TCP | `hyperping_monitor` (protocol: port) | Direct mapping |
| DNS | Not supported | Manual conversion |
| SMTP | Not supported | Use HTTP/port |
| Transaction (multi-step) | `hyperping_monitor` | Simplified to single check |
| **Check Settings** | | |
| Resolution (interval) | `check_frequency` | Map to supported values |
| Probe locations | `regions` | Map location names |
| Custom headers | `request_headers` | Direct mapping |
| Post data | `request_body` | Direct mapping |
| **Alerting** | | |
| Contact groups | Status page subscribers | Different model |
| Integrations | Webhook configuration | Manual setup |
| SMS | Not supported | External service |

#### Conversion Examples

**HTTP Check:**

```hcl
# Pingdom HTTP Check
{
  "name": "Main Website",
  "hostname": "www.example.com",
  "type": "http",
  "resolution": 1,  # 1 minute
  "probe_filters": ["region:NA", "region:EU"]
}

# Generated Hyperping Monitor
resource "hyperping_monitor" "main_website" {
  name                 = "Main Website"
  url                  = "https://www.example.com"  # HTTPS assumed
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 60
  expected_status_code = "200"
  regions              = ["virginia", "london"]  # Mapped from NA, EU
  follow_redirects     = true
  paused               = false
}
```

**Transaction Check (Simplified):**

```hcl
# Pingdom Transaction Check (multi-step)
{
  "name": "Login Flow Test",
  "type": "transaction",
  "steps": [
    {"url": "https://example.com/login"},
    {"url": "https://example.com/dashboard"}
  ]
}

# Generated Hyperping Monitor (simplified to final step)
resource "hyperping_monitor" "login_flow_test" {
  name                 = "Login Flow Test"
  url                  = "https://example.com/dashboard"  # Final step
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 60
  expected_status_code = "200"
}

# Note added to manual-steps.md:
# ⚠ Original check was multi-step transaction. Simplified to final endpoint.
# Consider implementing full flow test separately.
```

#### Troubleshooting

**Issue: Transaction Check Complexity**

```
Warning: Transaction check 'Login Flow' has 5 steps
```

**Solution:**

```bash
# Simplify to final step
migrate-pingdom convert \
  --input pingdom-export.json \
  --transaction-strategy simplify

# Or skip transactions
migrate-pingdom convert \
  --input pingdom-export.json \
  --transaction-strategy skip
```

**Issue: Unsupported Check Types**

```
Error: DNS check 'Domain Resolver' not supported
```

**Solution:**

```bash
# Skip unsupported checks
migrate-pingdom convert \
  --input pingdom-export.json \
  --skip-unsupported

# Check manual-steps.md for alternative approaches
cat hyperping-config/manual-steps.md
```

## Output Files Explained

All migration tools generate a consistent set of output files:

### Directory Structure

```
hyperping-config/
├── migrated-resources.tf      # Main Terraform configuration
├── variables.tf                # Input variables
├── outputs.tf                  # Output values
├── terraform.tfvars            # Variable values (sensitive)
├── import.sh                   # Import script (executable)
├── migration-report.json       # Detailed migration report
├── manual-steps.md             # Manual post-migration tasks
└── rollback.sh                 # Emergency rollback script
```

### migrated-resources.tf

Main Terraform configuration with all converted resources.

**Structure:**

```hcl
# Header with metadata
# Generated by migrate-betterstack v1.0.0
# Source: Better Stack
# Generated: 2026-02-15T10:30:00Z
# Resources: 47 monitors, 3 status pages, 12 incidents

# Provider configuration
terraform {
  required_version = ">= 1.8"
  required_providers {
    hyperping = {
      source  = "develeap/hyperping"
      version = "~> 1.0"
    }
  }
}

provider "hyperping" {
  # API key from HYPERPING_API_KEY environment variable
}

# Monitors
resource "hyperping_monitor" "api_gateway" {
  # ... configuration
}

# Status pages
resource "hyperping_statuspage" "main" {
  # ... configuration
}

# ... more resources
```

**Usage:**

```bash
# Review configuration
cat migrated-resources.tf

# Format
terraform fmt migrated-resources.tf

# Validate
terraform validate
```

### variables.tf

Input variables for customization.

**Content:**

```hcl
variable "environment" {
  description = "Environment name (production, staging, development)"
  type        = string
  default     = "production"
}

variable "check_frequency_override" {
  description = "Override check frequency for all monitors"
  type        = number
  default     = null
}

variable "regions" {
  description = "Default regions for monitors"
  type        = list(string)
  default     = ["virginia", "london", "singapore"]
}

variable "alert_emails" {
  description = "Email addresses for status page alerts"
  type        = list(string)
  default     = []
}
```

**Usage:**

```bash
# Override variables
terraform apply -var="check_frequency_override=300"

# Or use tfvars file
terraform apply -var-file="custom.tfvars"
```

### terraform.tfvars

Variable values (gitignored for security).

**Content:**

```hcl
environment = "production"

alert_emails = [
  "ops@example.com",
  "sre@example.com",
  "oncall@example.com"
]

regions = ["virginia", "london", "singapore", "tokyo"]
```

**Security:**

```bash
# Ensure tfvars is gitignored
echo "*.tfvars" >> .gitignore
echo "terraform.tfstate*" >> .gitignore

# Encrypt sensitive tfvars for repo storage
gpg --encrypt terraform.tfvars
git add terraform.tfvars.gpg
```

### import.sh

Executable shell script for importing existing resources.

**When to use:**
- Resources already exist in Hyperping
- Need to bring existing infrastructure under Terraform management
- Partial migration where some resources were manually created

**Content:**

```bash
#!/bin/bash
set -e

echo "Starting Terraform import..."

# Check environment
if [ -z "$HYPERPING_API_KEY" ]; then
  echo "Error: HYPERPING_API_KEY not set"
  exit 1
fi

# Initialize Terraform
terraform init

# Import monitors
echo "Importing monitors..."
terraform import 'hyperping_monitor.api_gateway' 'mon_abc123' || echo "Warning: Failed to import api_gateway"
terraform import 'hyperping_monitor.database' 'mon_def456' || echo "Warning: Failed to import database"
# ... more imports

# Import status pages
echo "Importing status pages..."
terraform import 'hyperping_statuspage.main' 'sp_xyz789' || echo "Warning: Failed to import main"

# Verify
echo "Verifying imports..."
terraform plan

echo "Import complete!"
```

**Usage:**

```bash
# Make executable
chmod +x import.sh

# Run import
./import.sh

# Verify zero drift
terraform plan
```

**Common errors:**

```bash
# Resource doesn't exist in Hyperping
Error: Cannot import non-existent resource

# Solution: Remove from configuration or create in Hyperping first

# Resource already in state
Error: Resource already managed by Terraform

# Solution: Skip this resource or remove from state first
terraform state rm 'hyperping_monitor.api_gateway'
```

### migration-report.json

Detailed JSON report of migration process.

**Structure:**

```json
{
  "migration": {
    "tool": "migrate-betterstack",
    "version": "1.0.0",
    "source": "Better Stack",
    "timestamp": "2026-02-15T10:30:00Z",
    "duration_seconds": 45
  },
  "summary": {
    "total_resources": 62,
    "successful": 60,
    "warnings": 2,
    "errors": 0,
    "skipped": 0
  },
  "resources": {
    "monitors": {
      "total": 47,
      "converted": 47,
      "skipped": 0
    },
    "statuspages": {
      "total": 3,
      "converted": 3,
      "skipped": 0
    },
    "incidents": {
      "total": 12,
      "converted": 10,
      "skipped": 2
    }
  },
  "warnings": [
    {
      "type": "frequency_rounded",
      "resource": "monitor:api_gateway",
      "message": "Check frequency rounded from 45s to 60s",
      "impact": "low"
    },
    {
      "type": "feature_unsupported",
      "resource": "incident:resolved_old",
      "message": "Resolved incidents older than 90 days skipped",
      "impact": "low"
    }
  ],
  "manual_steps_required": 2,
  "estimated_monthly_cost": "$49.00"
}
```

**Usage:**

```bash
# View summary
cat migration-report.json | jq '.summary'

# Check warnings
cat migration-report.json | jq '.warnings[]'

# View specific resource type
cat migration-report.json | jq '.resources.monitors'

# Export to readable format
jq -r '.warnings[] | "\(.resource): \(.message)"' migration-report.json
```

### manual-steps.md

Checklist of manual tasks required after automated migration.

**Content:**

```markdown
# Manual Post-Migration Steps

This file lists tasks that require manual intervention after automated migration.

## Status: 🟡 2 tasks pending

## Tasks

### 1. Verify Keyword Monitor Response Bodies

**Priority:** High
**Affected Resources:**
- `hyperping_monitor.api_status_check`
- `hyperping_monitor.health_endpoint`

**Description:**
Keyword monitoring has limited support. Verify that response bodies contain expected keywords.

**Steps:**
1. Open Hyperping dashboard
2. Navigate to each monitor
3. Trigger manual check
4. Verify response body contains keywords:
   - api_status_check: `"status":"healthy"`
   - health_endpoint: `"ok":true`

**Verification:**
```bash
curl https://api.example.com/status | jq '.status'
# Expected: "healthy"
```

### 2. Configure Slack Webhook Integration

**Priority:** Medium
**Affected Resources:**
- Status page notifications

**Description:**
Slack webhooks need manual configuration in Hyperping dashboard.

**Steps:**
1. Go to Hyperping → Settings → Integrations
2. Add Slack integration
3. Enter webhook URL: `https://hooks.slack.com/services/XXX`
4. Test notification
5. Link to status pages

**Documentation:**
https://docs.hyperping.io/integrations/slack

---

## Completed Tasks

None yet. Mark tasks as completed:

```bash
# Edit this file and move completed tasks here
```
```

**Usage:**

```bash
# View manual steps
cat manual-steps.md

# Track progress
vim manual-steps.md
# Update "Status:" count as you complete tasks
```

### rollback.sh

Emergency rollback script (use with caution).

**Content:**

```bash
#!/bin/bash
set -e

echo "⚠️  WARNING: This will destroy all migrated Hyperping resources!"
echo "Press Ctrl+C to cancel, or Enter to continue..."
read

# Backup current state
cp terraform.tfstate terraform.tfstate.backup-$(date +%Y%m%d-%H%M%S)

# Destroy Hyperping resources
terraform destroy -auto-approve

echo "Rollback complete. Hyperping resources destroyed."
echo "State backed up to terraform.tfstate.backup-*"
echo ""
echo "Next steps:"
echo "1. Re-enable monitors in original platform"
echo "2. Review what went wrong"
echo "3. Fix issues and retry migration"
```

**When to use:**
- Critical migration failure
- Need to abort migration
- Emergency situations only

**Safety:**

```bash
# Review before running
cat rollback.sh

# Dry-run (safer)
terraform plan -destroy

# Actual rollback (destructive)
./rollback.sh
```

## Troubleshooting

### Common Issues Across All Tools

#### Issue: Authentication Failed

**Symptoms:**

```
Error: HTTP 401 Unauthorized
Failed to authenticate with source platform API
```

**Solutions:**

1. **Verify API key:**

```bash
# Check environment variable is set
echo $BETTERSTACK_API_TOKEN
echo $HYPERPING_API_KEY

# Test API key manually
curl -H "Authorization: Bearer $HYPERPING_API_KEY" \
  https://api.hyperping.io/v1/monitors
```

2. **Check API key permissions:**
   - Source platform: Read permissions required
   - Hyperping: Write permissions required

3. **Regenerate API key:**
   - Old key may have expired
   - Generate new key and update environment

#### Issue: Network Timeout

**Symptoms:**

```
Error: Request timeout after 30s
Failed to fetch monitors from API
```

**Solutions:**

```bash
# Increase timeout
migrate-betterstack export \
  --api-key $BETTERSTACK_API_TOKEN \
  --timeout 60s

# Use retry logic
migrate-betterstack export \
  --api-key $BETTERSTACK_API_TOKEN \
  --retries 3 \
  --retry-delay 5s
```

#### Issue: Large Export Size

**Symptoms:**

```
Warning: Export file size 50MB
Processing may take several minutes
```

**Solutions:**

```bash
# Export in batches
migrate-betterstack export \
  --api-key $BETTERSTACK_API_TOKEN \
  --batch-size 100 \
  --output batch-{batch}.json

# Filter during export
migrate-betterstack export \
  --api-key $BETTERSTACK_API_TOKEN \
  --filter "tags:production" \
  --output prod-export.json
```

#### Issue: Terraform Validation Failed

**Symptoms:**

```
Error: Invalid resource name
Resource names must contain only alphanumeric characters
```

**Solutions:**

```bash
# Review naming convention
migrate-betterstack convert \
  --input export.json \
  --naming-convention snake-case \
  --sanitize-names

# Manually fix invalid names
vim hyperping-config/migrated-resources.tf
# Replace invalid characters in resource names
```

#### Issue: Import Conflicts

**Symptoms:**

```
Error: Resource already exists
hyperping_monitor.api_gateway already exists in state
```

**Solutions:**

```bash
# Remove from state first
terraform state rm 'hyperping_monitor.api_gateway'

# Then re-import
terraform import 'hyperping_monitor.api_gateway' 'mon_abc123'

# Or skip if already managed
# Edit import.sh and comment out problematic imports
```

#### Issue: Plan Shows Unexpected Changes

**Symptoms:**

```
# terraform plan
Plan: 0 to add, 15 to change, 0 to destroy

Changes to hyperping_monitor.api_gateway:
  - regions = ["london", "virginia"] -> ["virginia", "london"]
```

**Root causes:**
- List ordering differences
- Default value mismatches
- Computed field conflicts

**Solutions:**

```bash
# Fix list ordering
vim migrated-resources.tf
# Match exact order from state: terraform state show hyperping_monitor.api_gateway

# Specify all defaults explicitly
resource "hyperping_monitor" "api_gateway" {
  # ... other fields
  protocol             = "http"           # Explicit even if default
  check_frequency      = 60               # Explicit even if default
  expected_status_code = "200"            # Explicit even if default
  follow_redirects     = true             # Explicit even if default
}

# Verify fix
terraform plan
# Expected: No changes
```

### Platform-Specific Troubleshooting

See tool-specific sections above for detailed troubleshooting:
- [Better Stack Troubleshooting](#better-stack-migration-tool)
- [UptimeRobot Troubleshooting](#uptimerobot-migration-tool)
- [Pingdom Troubleshooting](#pingdom-migration-tool)

## Advanced Usage

### Selective Migration

Migrate only specific resources:

```bash
# Migrate only critical monitors
migrate-betterstack migrate \
  --source-api-key $BETTERSTACK_API_TOKEN \
  --dest-api-key $HYPERPING_API_KEY \
  --filter "tags:critical" \
  --output critical-migration

# Migrate by name pattern
migrate-betterstack migrate \
  --source-api-key $BETTERSTACK_API_TOKEN \
  --dest-api-key $HYPERPING_API_KEY \
  --filter "name:*production*" \
  --output prod-migration

# Migrate specific resource types
migrate-betterstack migrate \
  --source-api-key $BETTERSTACK_API_TOKEN \
  --dest-api-key $HYPERPING_API_KEY \
  --resources monitors,statuspages \
  --output partial-migration
```

### Dry-Run Testing

Test migration without creating resources:

```bash
# Dry-run full migration
migrate-betterstack migrate \
  --source-api-key $BETTERSTACK_API_TOKEN \
  --dest-api-key $HYPERPING_API_KEY \
  --dry-run \
  --output test-migration

# Shows what would be created:
# ✓ Would create 47 monitors
# ✓ Would create 3 status pages
# ✓ Would create 12 incidents
# ⚠ 2 warnings (see report)
# ✓ Dry-run complete. No resources created.

# Review generated files
cd test-migration
terraform plan
```

### Validation Before Import

Validate configuration before applying:

```bash
# Generate configuration only
migrate-betterstack convert \
  --input export.json \
  --output hyperping-config \
  --no-import

cd hyperping-config

# Validate Terraform
terraform init
terraform validate
# Success! The configuration is valid.

# Plan without applying
terraform plan -out=migration.tfplan

# Review plan carefully
terraform show migration.tfplan

# Apply when ready
terraform apply migration.tfplan
```

### Incremental Migration Strategy

Migrate in phases for large deployments:

```bash
# Phase 1: Critical monitors
migrate-betterstack migrate \
  --source-api-key $BETTERSTACK_API_TOKEN \
  --dest-api-key $HYPERPING_API_KEY \
  --filter "tags:critical" \
  --prefix phase1_ \
  --output phase1-migration

cd phase1-migration
terraform apply

# Wait 24-48 hours, validate

# Phase 2: Standard monitors
migrate-betterstack migrate \
  --source-api-key $BETTERSTACK_API_TOKEN \
  --dest-api-key $HYPERPING_API_KEY \
  --filter "tags:standard" \
  --prefix phase2_ \
  --output phase2-migration

cd phase2-migration
terraform apply

# Continue until complete
```

### Customization After Generation

Modify generated configuration:

```bash
# Generate base configuration
migrate-betterstack convert \
  --input export.json \
  --output hyperping-config

cd hyperping-config

# Customize monitors
vim migrated-resources.tf

# Example customizations:

# 1. Add tags/labels via naming convention
resource "hyperping_monitor" "api_gateway" {
  name = "[CRITICAL]-API-Gateway"  # Add prefix
  # ... other config
}

# 2. Group by environment
locals {
  prod_monitors = {
    for k, v in hyperping_monitor.* : k => v
    if can(regex("prod", v.name))
  }
}

# 3. Add custom outputs
output "critical_monitor_count" {
  value = length([
    for m in hyperping_monitor.* : m
    if can(regex("CRITICAL", m.name))
  ])
}

# 4. Create status page sections by environment
resource "hyperping_statuspage" "main" {
  # ...
  sections = [
    {
      name     = { en = "Production Services" }
      services = [for m in local.prod_monitors : {
        monitor_uuid = m.id
      }]
    }
  ]
}

# Validate changes
terraform fmt
terraform validate
terraform plan
```

### Parallel Operation Period

Run both platforms simultaneously:

```bash
# Step 1: Keep Better Stack running
# Don't pause/delete monitors yet

# Step 2: Deploy Hyperping monitors
cd hyperping-migration
terraform apply

# Step 3: Compare for 1-2 weeks
./compare-platforms.sh  # Generated script

# Step 4: Decommission Better Stack when confident
migrate-betterstack decommission \
  --api-key $BETTERSTACK_API_TOKEN \
  --pause-all  # Pause, don't delete

# Step 5: After 30 days, permanently delete
migrate-betterstack decommission \
  --api-key $BETTERSTACK_API_TOKEN \
  --delete-all \
  --confirm
```

### Multi-Environment Migration

Migrate multiple environments with variable files:

```bash
# Directory structure
mkdir -p environments/{dev,staging,prod}

# Generate base configuration
migrate-betterstack convert \
  --input export.json \
  --output base-config

# Copy to each environment
cp -r base-config/* environments/dev/
cp -r base-config/* environments/staging/
cp -r base-config/* environments/prod/

# Customize per environment
cat > environments/dev/terraform.tfvars << EOF
environment = "development"
check_frequency_override = 300  # Less frequent in dev
regions = ["virginia"]           # Single region in dev
EOF

cat > environments/prod/terraform.tfvars << EOF
environment = "production"
check_frequency_override = 60   # More frequent in prod
regions = ["virginia", "london", "singapore", "tokyo"]  # Multi-region
EOF

# Apply to each environment
for env in dev staging prod; do
  echo "Applying to $env..."
  cd environments/$env
  terraform init
  terraform apply -auto-approve
  cd ../..
done
```

## Best Practices

### 1. Pre-Migration Preparation

**Inventory your current setup:**

```bash
# Export and analyze
migrate-betterstack export --output export.json
migrate-betterstack analyze --input export.json

# Document findings
cat analyze-report.md >> migration-planning.md
```

**Plan your approach:**

- [ ] Choose migration strategy (big bang, phased, parallel)
- [ ] Identify critical vs. non-critical monitors
- [ ] Schedule maintenance window (if needed)
- [ ] Prepare rollback plan
- [ ] Brief team on changes

### 2. Test in Non-Production First

**Always test migration process:**

```bash
# Migrate dev environment first
migrate-betterstack migrate \
  --source-api-key $BETTERSTACK_API_TOKEN_DEV \
  --dest-api-key $HYPERPING_API_KEY_DEV \
  --output dev-migration

# Validate thoroughly
cd dev-migration
terraform apply
# ... test for 1-2 weeks ...

# Then migrate staging
# Then migrate production
```

### 3. Backup Everything

**Before migration:**

```bash
# Export from source platform
migrate-betterstack export --output betterstack-backup.json

# Backup Terraform state (if exists)
cp terraform.tfstate terraform.tfstate.pre-migration

# Backup current configurations
tar -czf pre-migration-backup-$(date +%Y%m%d).tar.gz \
  betterstack-backup.json \
  terraform.tfstate* \
  *.tf
```

### 4. Version Control

**Track migration in git:**

```bash
# Initialize git if not already
git init

# Add gitignore
cat > .gitignore << EOF
*.tfstate
*.tfstate.backup
.terraform/
*.tfvars
*.tfplan
EOF

# Commit migration artifacts
git add migrated-resources.tf variables.tf outputs.tf
git add migration-report.json manual-steps.md
git commit -m "feat: migrate from Better Stack to Hyperping"

# Tag migration milestone
git tag migration-$(date +%Y%m%d)
```

### 5. Gradual Cutover

**Minimize risk with phased approach:**

```bash
# Week 1: Deploy Hyperping alongside Better Stack
terraform apply

# Week 2: Validate alerts, uptime data, status pages
./validate-migration.sh

# Week 3: Gradually shift alerting to Hyperping
# Update Slack webhooks, email contacts, etc.

# Week 4: Pause Better Stack monitors
migrate-betterstack decommission --pause-all

# Week 8: Delete Better Stack after 30-day safety period
migrate-betterstack decommission --delete-all --confirm
```

### 6. Validate Thoroughly

**Comprehensive validation checklist:**

```bash
# 1. Terraform plan should show no changes
terraform plan
# Expected: No changes

# 2. All monitors are active
terraform state list | grep hyperping_monitor | wc -l
# Expected: matches export count

# 3. Status pages are accessible
curl -I https://status.example.com
# Expected: HTTP 200

# 4. Test alerts
# Trigger test alert in Hyperping dashboard

# 5. Verify status page subscribers
terraform state show hyperping_statuspage_subscriber.team

# 6. Check for drift over time
# Run daily for 1 week
terraform plan | grep "No changes"
```

### 7. Document the Migration

**Maintain migration log:**

```markdown
# Migration Log - Better Stack to Hyperping

## Date: 2026-02-15

### Pre-Migration
- [x] Exported 47 monitors from Better Stack
- [x] Analyzed compatibility: 100%
- [x] Reviewed generated Terraform
- [x] Backed up all configurations

### Migration Execution
- [x] Applied Terraform configuration
- [x] Imported 47 monitors successfully
- [x] Created 3 status pages
- [x] Configured 10 email subscribers

### Validation
- [x] Zero drift after initial apply
- [x] Test alerts delivered successfully
- [x] Status pages accessible
- [x] Uptime data populating

### Issues Encountered
1. Check frequency rounded from 45s to 60s (monitors: api_gateway)
   - Resolution: Accepted rounding, within acceptable range

### Post-Migration
- [x] Updated team documentation
- [x] Briefed on-call team
- [ ] Monitor for 1 week before decommission
- [ ] Decommission Better Stack (scheduled: 2026-02-22)

### Rollback Plan
- Terraform state backed up: terraform.tfstate.pre-migration
- Better Stack monitors paused (not deleted)
- Can re-enable Better Stack in <5 minutes if needed
```

### 8. Monitor the Migration

**Track migration progress:**

```bash
# Real-time progress
migrate-betterstack migrate \
  --source-api-key $BETTERSTACK_API_TOKEN \
  --dest-api-key $HYPERPING_API_KEY \
  --progress \
  --output migration

# Example output:
# Exporting from Better Stack...
# [████████████████████] 47/47 monitors (100%)
# [███████░░░░░░░░░░░░░] 3/12 incidents (25%)
#
# Converting to Hyperping format...
# [████████████████████] 47/47 monitors (100%)
#
# Generating Terraform...
# [████████████████████] 50/50 resources (100%)
#
# ✓ Migration complete!
```

### 9. Security Best Practices

**Protect sensitive data:**

```bash
# Never commit API keys
echo "*.tfvars" >> .gitignore
echo ".env" >> .gitignore

# Use environment variables
export HYPERPING_API_KEY="sk_your_key"
# Don't: api_key = "sk_your_key" in .tf files

# Encrypt state files
terraform init -backend-config=backend.hcl
# backend.hcl contains encrypted S3 backend config

# Rotate API keys after migration
# Generate new Hyperping API key
# Update in CI/CD and local environments
```

### 10. Cost Optimization

**Optimize costs during migration:**

```bash
# Analyze costs before migration
migrate-betterstack analyze --input export.json --cost-estimate

# Output:
# Cost Estimate
# =============
# Current (Better Stack): $199/month
# Estimated (Hyperping):  $149/month
# Savings:                $50/month (25%)

# Reduce check frequencies for non-critical monitors
vim migrated-resources.tf
# Change check_frequency from 60 to 300 for non-critical

# Reduce monitoring regions
# Critical monitors: multi-region
# Non-critical: single region

# Estimate savings
terraform plan | grep "check_frequency"
```

## FAQ

### General Questions

**Q: How long does a typical migration take?**

A: Time varies by size:
- **Small** (1-20 monitors): 30 minutes - 1 hour
- **Medium** (21-100 monitors): 1-3 hours
- **Large** (100+ monitors): 3-6 hours
- Add 1-2 weeks for parallel validation period

**Q: Will I experience downtime during migration?**

A: No downtime if you follow the parallel run approach:
1. Deploy Hyperping alongside existing platform
2. Validate both are monitoring correctly
3. Gradually shift alerting to Hyperping
4. Decommission old platform

**Q: What happens if migration fails?**

A: Multiple safety nets:
- Original platform remains active
- Terraform state can be destroyed
- Rollback script provided
- No data loss from source platform

**Q: Can I migrate just some resources?**

A: Yes, use filtering:

```bash
migrate-betterstack migrate \
  --filter "tags:production" \
  --resources monitors,statuspages
```

**Q: Are there any costs during migration?**

A: You'll pay for both platforms during parallel run period (typically 1-2 weeks). Plan accordingly.

### Technical Questions

**Q: What Terraform version is required?**

A: Terraform >= 1.8 is required for the Hyperping provider.

**Q: Can I use Terraform Cloud/Enterprise?**

A: Yes, the generated configuration works with:
- Terraform Cloud
- Terraform Enterprise
- Spacelift
- Atlantis
- Any standard Terraform backend

**Q: How do I handle state management?**

A: Configure remote state before migration:

```hcl
terraform {
  backend "s3" {
    bucket = "company-terraform-state"
    key    = "hyperping/terraform.tfstate"
    region = "us-east-1"
  }
}
```

**Q: What if I already have some Hyperping resources?**

A: Use import mode:

```bash
# Export existing Hyperping resources
import-generator --format=hcl --output=existing.tf

# Merge with migrated resources
cat migrated-resources.tf existing.tf > combined.tf

# Remove duplicates manually
vim combined.tf
```

**Q: Can I customize resource names?**

A: Yes, multiple options:

```bash
# Prefix all resources
migrate-betterstack convert \
  --prefix prod_ \
  --input export.json

# Choose naming convention
migrate-betterstack convert \
  --naming-convention kebab-case \
  --input export.json

# Or manually edit generated files
vim migrated-resources.tf
```

### Platform-Specific Questions

**Q: What Better Stack features aren't supported?**

A: Partially supported or unsupported:
- On-call schedules (manual setup)
- Complex escalation policies (simplified)
- Phone call alerts (not available)
- SMS alerts (not available)

**Q: What about UptimeRobot ping monitors?**

A: Converted to HTTP or ICMP checks:

```bash
migrate-uptimerobot convert \
  --ping-strategy http \  # or icmp
  --input export.json
```

**Q: Can I migrate Pingdom transaction checks?**

A: Simplified to final step:

```bash
migrate-pingdom convert \
  --transaction-strategy simplify \
  --input export.json
```

See `manual-steps.md` for complex transaction flows.

### Troubleshooting Questions

**Q: Export fails with rate limit error?**

A: Add delays:

```bash
migrate-betterstack export \
  --rate-limit-delay 1000 \  # 1 second between requests
  --retries 3
```

**Q: Terraform plan shows unexpected changes?**

A: Usually list ordering or default values:

```bash
# Check actual vs. expected
terraform plan -out=plan.tfplan
terraform show plan.tfplan

# Fix ordering
vim migrated-resources.tf
# Match order from: terraform state show resource_name
```

**Q: Import script fails halfway?**

A: Continue from where it stopped:

```bash
# Edit import.sh
vim import.sh
# Comment out successful imports
# Re-run
./import.sh
```

**Q: How do I rollback a failed migration?**

A: Use rollback script or manual:

```bash
# Automated rollback
./rollback.sh

# Or manual
terraform destroy
# Then re-enable old platform monitors
```

### Post-Migration Questions

**Q: How do I add new monitors after migration?**

A: Just use Terraform:

```hcl
# Add to migrated-resources.tf
resource "hyperping_monitor" "new_service" {
  name = "New Service Monitor"
  url  = "https://new.example.com/health"
  # ...
}
```

```bash
terraform apply
```

**Q: Should I keep migration files?**

A: Yes, archive for reference:

```bash
mkdir -p archive/migration-2026-02-15
mv migration-report.json archive/migration-2026-02-15/
mv betterstack-export.json archive/migration-2026-02-15/
tar -czf archive/migration-2026-02-15.tar.gz archive/migration-2026-02-15/
```

**Q: How do I migrate another environment?**

A: Repeat process with different API keys:

```bash
# Staging migration
export BETTERSTACK_API_TOKEN="staging_token"
export HYPERPING_API_KEY="staging_key"
migrate-betterstack migrate --output staging-migration

# Production migration
export BETTERSTACK_API_TOKEN="prod_token"
export HYPERPING_API_KEY="prod_key"
migrate-betterstack migrate --output prod-migration
```

**Q: Can I re-use the migration tool?**

A: Yes, for:
- Migrating additional environments
- Incremental migrations (new monitors)
- Testing configuration changes
- Exporting for backup purposes

---

## Summary

The automated migration tools dramatically simplify migrating to Hyperping:

**Key Advantages:**
- ⏱️ **90% time savings** vs. manual migration
- ✅ **Consistent results** with validated conversions
- 📋 **Complete documentation** of migration process
- 🔄 **Repeatable** for multiple environments
- 🛡️ **Safe** with dry-run, validation, and rollback

**Migration Checklist:**

- [ ] Install migration tool
- [ ] Set API keys (source and destination)
- [ ] Export from source platform
- [ ] Review and customize conversion
- [ ] Test in non-production environment
- [ ] Execute production migration
- [ ] Validate thoroughly (zero drift)
- [ ] Run parallel operation period (1-2 weeks)
- [ ] Decommission old platform
- [ ] Document and archive migration

**Next Steps:**

1. **Choose your tool**: Better Stack, UptimeRobot, or Pingdom
2. **Review tool-specific guide** above
3. **Run test migration** in development
4. **Execute production migration** with confidence
5. **Join the community** for support and tips

**Need Help?**

- 📚 [Hyperping Documentation](https://docs.hyperping.io)
- 💬 [GitHub Discussions](https://github.com/develeap/terraform-provider-hyperping/discussions)
- 🐛 [Report Issues](https://github.com/develeap/terraform-provider-hyperping/issues)
- 📧 [Hyperping Support](https://hyperping.io/support)

**Related Guides:**

- [Manual Migration from Better Stack](./migrate-from-betterstack.md)
- [Manual Migration from UptimeRobot](./migrate-from-uptimerobot.md)
- [Manual Migration from Pingdom](./migrate-from-pingdom.md)
- [Importing Existing Resources](./importing-resources.md)
- [Best Practices](./best-practices.md)

---

Happy migrating! 🚀
