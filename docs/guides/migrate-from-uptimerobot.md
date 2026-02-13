---
page_title: "Migrating from UptimeRobot to Hyperping"
subcategory: "Guides"
description: |-
  Complete guide to migrating monitoring infrastructure from UptimeRobot to Hyperping using Terraform.
---

# Migrating from UptimeRobot to Hyperping

This guide provides comprehensive instructions for migrating your monitoring infrastructure from UptimeRobot to Hyperping using Terraform. Whether you have a handful of monitors or hundreds, this guide covers everything from planning to execution.

## Table of Contents

- [Why Migrate?](#why-migrate)
- [Prerequisites](#prerequisites)
- [Migration Planning](#migration-planning)
- [Monitor Type Mapping](#monitor-type-mapping)
- [Alert Contact Migration](#alert-contact-migration)
- [Export from UptimeRobot](#export-from-uptimerobot)
- [Import Workflow](#import-workflow)
- [Complete Migration Examples](#complete-migration-examples)
- [Validation and Testing](#validation-and-testing)
- [Troubleshooting](#troubleshooting)
- [Best Practices](#best-practices)

## Why Migrate?

### Benefits of Hyperping over UptimeRobot

**Modern Architecture:**
- Built for API-first monitoring and modern cloud infrastructure
- Better support for dynamic environments (Kubernetes, microservices)
- More flexible alerting with escalation policies

**Developer Experience:**
- Full Terraform provider for Infrastructure as Code
- Better API documentation and reliability
- Webhook-based healthchecks for cron jobs (dead man's switch)

**Monitoring Features:**
- Multiple check regions with granular control
- Advanced HTTP request customization (headers, body, methods)
- Keyword monitoring in response bodies
- Custom check frequencies from 10 seconds to 24 hours

**Status Pages:**
- Integrated status pages with incidents and maintenance windows
- Multi-language support
- Custom domain and branding options

### When to Migrate

Good scenarios for migration:
- ✅ Starting Infrastructure as Code adoption
- ✅ Need more flexible API monitoring
- ✅ Want better integration with modern CI/CD
- ✅ Require advanced status page features
- ✅ Need better webhook-based healthchecks

Consider staying if:
- ❌ Heavily dependent on UptimeRobot-specific features (ping monitoring)
- ❌ Using UptimeRobot's free tier extensively with no budget for alternatives
- ❌ Integration deeply embedded in existing systems

## Prerequisites

Before starting the migration, ensure you have:

### Required Tools

```bash
# 1. Terraform 1.8 or higher
terraform version

# 2. jq for JSON processing (optional but recommended)
jq --version

# 3. curl for API calls
curl --version
```

### Required Access

**UptimeRobot:**
- [ ] UptimeRobot account access
- [ ] API key (Main API Key or Read-Only API Key)
- [ ] Ability to export monitor configurations

**Hyperping:**
- [ ] Hyperping account created
- [ ] API key generated (Settings → API Keys)
- [ ] Sufficient plan limits for your monitors

### Environment Setup

```bash
# UptimeRobot API key
export UPTIMEROBOT_API_KEY="u1234567-abcdefghijklmnopqrstuvwxyz"

# Hyperping API key
export HYPERPING_API_KEY="sk_your_hyperping_api_key"

# Create working directory
mkdir uptimerobot-to-hyperping-migration
cd uptimerobot-to-hyperping-migration
```

## Migration Planning

### Step 1: Inventory Your UptimeRobot Resources

Create a complete inventory of what needs to be migrated:

```bash
# List all monitors
curl -X POST https://api.uptimerobot.com/v2/getMonitors \
  -d "api_key=$UPTIMEROBOT_API_KEY" \
  -d "format=json" \
  | jq -r '.monitors[] | "\(.id)\t\(.friendly_name)\t\(.type)\t\(.url)"' \
  > uptimerobot-monitors.txt

# Count monitors by type
cat uptimerobot-monitors.txt | awk '{print $3}' | sort | uniq -c
```

Example inventory output:

```
Monitor ID  Name                    Type  URL
78901234    Production API          1     https://api.example.com
78901235    Website Homepage        1     https://example.com
78901236    Database Port Check     4     example.com:5432
78901237    Server Ping             3     192.168.1.100
```

### Step 2: Categorize Monitors

Group monitors by migration strategy:

| Category | UptimeRobot Type | Hyperping Support | Migration Strategy |
|----------|------------------|-------------------|-------------------|
| HTTP/HTTPS | Type 1 | ✅ Full | Direct migration |
| Keyword | Type 2 | ✅ Full | Direct migration with keyword field |
| Ping (ICMP) | Type 3 | ✅ Full | Direct migration using protocol="icmp" |
| Port | Type 4 | ✅ Full | Direct migration using protocol="port" |
| Heartbeat | Type 5 | ✅ Full | Migrate to healthchecks |

### Step 3: Create Migration Timeline

**Week 1: Planning and Setup**
- Export UptimeRobot configurations
- Set up Hyperping account
- Create test migrations

**Week 2: Parallel Testing**
- Create monitors in Hyperping
- Run both systems in parallel
- Validate alerting

**Week 3: Cutover**
- Switch primary alerting to Hyperping
- Document new processes
- Deactivate UptimeRobot monitors

**Week 4: Cleanup**
- Cancel UptimeRobot subscription
- Archive migration documentation

## Monitor Type Mapping

### HTTP/HTTPS Monitors (Type 1)

**UptimeRobot Configuration:**

```json
{
  "id": 78901234,
  "friendly_name": "Production API",
  "url": "https://api.example.com/health",
  "type": 1,
  "interval": 300,
  "timeout": 30,
  "http_method": 1,
  "http_auth_type": 0
}
```

**Hyperping Equivalent:**

```hcl
resource "hyperping_monitor" "production_api" {
  name                 = "Production API"
  url                  = "https://api.example.com/health"
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 300
  expected_status_code = "2xx"
  follow_redirects     = true

  regions = ["london", "virginia", "singapore"]
}
```

**Field Mapping:**

| UptimeRobot | Hyperping | Notes |
|-------------|-----------|-------|
| `friendly_name` | `name` | Direct mapping |
| `url` | `url` | Direct mapping |
| `interval` (seconds) | `check_frequency` | Must be allowed value: 10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400 |
| `timeout` | N/A | Hyperping uses region-based defaults |
| `http_method` | `http_method` | 1=GET, 2=POST, 3=PUT, 4=PATCH, 5=DELETE, 6=HEAD |

**HTTP Method Conversion:**

```bash
# Conversion table
1 → "GET"
2 → "POST"
3 → "PUT"
4 → "PATCH"
5 → "DELETE"
6 → "HEAD"
```

### Keyword Monitors (Type 2)

**UptimeRobot Configuration:**

```json
{
  "id": 78901235,
  "friendly_name": "Homepage Status Check",
  "url": "https://example.com",
  "type": 2,
  "keyword_type": 1,
  "keyword_value": "Welcome",
  "interval": 300
}
```

**Hyperping Equivalent:**

```hcl
resource "hyperping_monitor" "homepage_status" {
  name                 = "Homepage Status Check"
  url                  = "https://example.com"
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 300
  expected_status_code = "200"
  required_keyword     = "Welcome"

  regions = ["london", "virginia"]
}
```

**Keyword Type Mapping:**

| UptimeRobot keyword_type | Behavior | Hyperping Equivalent |
|-------------------------|----------|---------------------|
| 1 (exists) | Keyword must exist | `required_keyword = "text"` |
| 2 (not exists) | Keyword must NOT exist | Not directly supported - use different approach |

**Note:** Hyperping's `required_keyword` only supports "must exist" checking. For "must not exist" scenarios, consider using a separate validation endpoint.

### Ping Monitors (Type 3)

**UptimeRobot Configuration:**

```json
{
  "id": 78901236,
  "friendly_name": "Server ICMP Check",
  "url": "192.168.1.100",
  "type": 3,
  "interval": 60
}
```

**Hyperping Equivalent:**

```hcl
resource "hyperping_monitor" "server_icmp" {
  name            = "Server ICMP Check"
  url             = "192.168.1.100"
  protocol        = "icmp"
  check_frequency = 60

  regions = ["london", "virginia", "singapore"]
}
```

**Important Notes:**
- Hyperping ICMP monitors check availability using ping
- No need to specify port number
- URL field contains the IP address or hostname
- Regional availability may vary for ICMP checks

### Port Monitors (Type 4)

**UptimeRobot Configuration:**

```json
{
  "id": 78901237,
  "friendly_name": "PostgreSQL Port Check",
  "url": "db.example.com",
  "type": 4,
  "sub_type": 1,
  "port": 5432,
  "interval": 120
}
```

**Hyperping Equivalent:**

```hcl
resource "hyperping_monitor" "postgresql_port" {
  name            = "PostgreSQL Port Check"
  url             = "db.example.com"
  protocol        = "port"
  port            = 5432
  check_frequency = 120

  regions = ["virginia"]
}
```

**Port Sub-Type Mapping:**

| UptimeRobot sub_type | Service | Default Port | Hyperping |
|---------------------|---------|--------------|-----------|
| 1 | Custom | User-defined | `protocol = "port"`, `port = X` |
| 2 | HTTP (80) | 80 | `protocol = "port"`, `port = 80` |
| 3 | HTTPS (443) | 443 | `protocol = "port"`, `port = 443` |
| 4 | FTP (21) | 21 | `protocol = "port"`, `port = 21` |
| 5 | SMTP (25) | 25 | `protocol = "port"`, `port = 25` |
| 6 | POP3 (110) | 110 | `protocol = "port"`, `port = 110` |
| 7 | IMAP (143) | 143 | `protocol = "port"`, `port = 143` |

### Heartbeat Monitors (Type 5)

UptimeRobot heartbeat monitors are best migrated to Hyperping healthchecks.

**UptimeRobot Configuration:**

```json
{
  "id": 78901238,
  "friendly_name": "Daily Backup Job",
  "type": 5,
  "interval": 86400
}
```

**Hyperping Equivalent:**

```hcl
resource "hyperping_healthcheck" "daily_backup" {
  name               = "Daily Backup Job"
  period_value       = 1
  period_type        = "days"
  grace_period_value = 1
  grace_period_type  = "hours"
}

output "backup_ping_url" {
  value       = hyperping_healthcheck.daily_backup.ping_url
  description = "Update your backup script to ping this URL"
  sensitive   = true
}
```

**Migration Steps for Heartbeat:**

1. Create the healthcheck in Hyperping
2. Get the ping URL from Terraform output
3. Update your cron job/script to ping the new URL
4. Test the healthcheck receives pings correctly
5. Disable the UptimeRobot heartbeat monitor

**Cron Expression Support:**

Hyperping healthchecks support cron expressions for more precise scheduling:

```hcl
resource "hyperping_healthcheck" "backup_cron" {
  name               = "Nightly Backup (2 AM)"
  cron               = "0 2 * * *"
  timezone           = "America/New_York"
  grace_period_value = 30
  grace_period_type  = "minutes"
}
```

### Check Frequency Mapping

UptimeRobot allows intervals in seconds. Hyperping has specific allowed values:

| UptimeRobot Interval | Hyperping check_frequency | Notes |
|---------------------|---------------------------|-------|
| 60 | 60 | Direct match |
| 120 | 120 | Direct match |
| 180 | 180 | Direct match |
| 300 | 300 | Direct match |
| 600 | 600 | Direct match |
| 1800 | 1800 | Direct match |
| 3600 | 3600 | Direct match |
| 43200 | 43200 | Direct match |
| 86400 | 86400 | Direct match |
| Other values | Round to nearest allowed | Use closest allowed value |

**Conversion Helper:**

```bash
# Function to map UptimeRobot interval to Hyperping check_frequency
map_frequency() {
  local interval=$1
  local allowed=(10 20 30 60 120 180 300 600 1800 3600 21600 43200 86400)

  # Find closest allowed value
  local closest=${allowed[0]}
  local min_diff=$((interval > closest ? interval - closest : closest - interval))

  for freq in "${allowed[@]}"; do
    local diff=$((interval > freq ? interval - freq : freq - interval))
    if [ $diff -lt $min_diff ]; then
      min_diff=$diff
      closest=$freq
    fi
  done

  echo $closest
}

# Usage
map_frequency 150  # Returns 120 or 180
```

## Alert Contact Migration

### Understanding Alert Channels

**UptimeRobot Alert Contacts:**
- Email notifications
- SMS
- Webhooks
- Third-party integrations (Slack, PagerDuty, etc.)

**Hyperping Escalation Policies:**
- Centralized alerting configuration
- Multi-step escalation
- Time-based routing
- Integration with notification channels

### Export UptimeRobot Alert Contacts

```bash
# Get all alert contacts
curl -X POST https://api.uptimerobot.com/v2/getAlertContacts \
  -d "api_key=$UPTIMEROBOT_API_KEY" \
  -d "format=json" \
  | jq '.' > uptimerobot-alert-contacts.json

# Extract email contacts
jq -r '.alert_contacts[] | select(.type == 2) | .value' uptimerobot-alert-contacts.json > email-contacts.txt
```

**Alert Contact Types in UptimeRobot:**

| Type | Name | Hyperping Equivalent |
|------|------|---------------------|
| 2 | Email | Escalation Policy → Email |
| 3 | SMS | Escalation Policy → SMS |
| 4 | Webhook | Escalation Policy → Webhook |
| 11 | Slack | Integration + Escalation Policy |
| 14 | PagerDuty | Integration + Escalation Policy |

### Creating Escalation Policies in Hyperping

While escalation policies cannot be created directly via Terraform (must be created in Hyperping dashboard), you can reference them:

**Step 1: Create Escalation Policy in Hyperping Dashboard**

1. Log into Hyperping dashboard
2. Navigate to Settings → Escalation Policies
3. Click "Create Escalation Policy"
4. Configure notification channels:
   - Email addresses
   - Webhooks
   - Third-party integrations
5. Note the escalation policy UUID

**Step 2: Reference in Terraform**

```hcl
variable "primary_escalation_policy" {
  description = "UUID of primary escalation policy"
  type        = string
  default     = "ep_abc123def456"
}

resource "hyperping_monitor" "api_with_alerts" {
  name                 = "API Health"
  url                  = "https://api.example.com/health"
  protocol             = "http"
  check_frequency      = 60
  expected_status_code = "200"

  escalation_policy = var.primary_escalation_policy
}

resource "hyperping_healthcheck" "backup_with_alerts" {
  name               = "Backup Job"
  period_value       = 1
  period_type        = "days"
  grace_period_value = 1
  grace_period_type  = "hours"

  escalation_policy = var.primary_escalation_policy
}
```

### Alert Contact Mapping Strategy

**Create Escalation Policies by Severity:**

```hcl
# Variable definitions
variable "critical_escalation_policy" {
  description = "For production-critical monitors"
  type        = string
  default     = "ep_critical_123"
}

variable "warning_escalation_policy" {
  description = "For non-critical monitors"
  type        = string
  default     = "ep_warning_456"
}

variable "dev_escalation_policy" {
  description = "For development environment"
  type        = string
  default     = "ep_dev_789"
}

# Production API - critical alerts
resource "hyperping_monitor" "prod_api" {
  name                 = "Production API"
  url                  = "https://api.prod.example.com"
  protocol             = "http"
  check_frequency      = 30
  expected_status_code = "200"
  escalation_policy    = var.critical_escalation_policy
}

# Staging API - warning alerts
resource "hyperping_monitor" "staging_api" {
  name                 = "Staging API"
  url                  = "https://api.staging.example.com"
  protocol             = "http"
  check_frequency      = 300
  expected_status_code = "200"
  escalation_policy    = var.warning_escalation_policy
}
```

### Webhook Migration

If you're using UptimeRobot webhooks for custom alerting:

**UptimeRobot Webhook Format:**

```json
{
  "monitorID": "78901234",
  "monitorURL": "https://example.com",
  "monitorFriendlyName": "Example Monitor",
  "alertType": "1",
  "alertTypeFriendlyName": "down",
  "alertDetails": "Connection timeout",
  "alertDateTime": "2026-02-13 10:30:00"
}
```

**Hyperping Webhook Format:**

Hyperping webhooks are configured in escalation policies and send different payload formats. You may need to adapt your webhook receivers.

**Webhook Receiver Adaptation:**

```python
# Example Flask webhook receiver that handles both formats
from flask import Flask, request
import json

app = Flask(__name__)

@app.route('/webhook', methods=['POST'])
def handle_webhook():
    data = request.get_json()

    # Detect source
    if 'monitorID' in data:
        # UptimeRobot format
        monitor_name = data.get('monitorFriendlyName')
        status = data.get('alertTypeFriendlyName')
        details = data.get('alertDetails')
    else:
        # Hyperping format (adjust based on actual format)
        monitor_name = data.get('monitor', {}).get('name')
        status = data.get('status')
        details = data.get('message')

    # Process alert
    send_alert(monitor_name, status, details)

    return '', 200
```

## Export from UptimeRobot

### Complete Monitor Export

Export all monitor configurations:

```bash
#!/bin/bash
# export-uptimerobot.sh

API_KEY="${UPTIMEROBOT_API_KEY}"
OUTPUT_DIR="uptimerobot-export"

mkdir -p "$OUTPUT_DIR"

echo "Exporting UptimeRobot monitors..."

# Get all monitors with full details
curl -s -X POST https://api.uptimerobot.com/v2/getMonitors \
  -d "api_key=$API_KEY" \
  -d "format=json" \
  -d "logs=1" \
  -d "response_times=1" \
  -d "alert_contacts=1" \
  > "$OUTPUT_DIR/monitors-full.json"

# Extract summary
jq -r '.monitors[] | [.id, .friendly_name, .type, .url, .interval] | @tsv' \
  "$OUTPUT_DIR/monitors-full.json" \
  > "$OUTPUT_DIR/monitors-summary.tsv"

echo "Export complete. Files saved to $OUTPUT_DIR/"
echo "Total monitors: $(jq '.monitors | length' "$OUTPUT_DIR/monitors-full.json")"
```

### Parse Monitor Types

```bash
#!/bin/bash
# categorize-monitors.sh

INPUT_FILE="uptimerobot-export/monitors-full.json"

echo "Categorizing monitors..."

# HTTP/HTTPS monitors
jq -r '.monitors[] | select(.type == 1) | "\(.id)\t\(.friendly_name)\t\(.url)"' \
  "$INPUT_FILE" > http-monitors.txt

# Keyword monitors
jq -r '.monitors[] | select(.type == 2) | "\(.id)\t\(.friendly_name)\t\(.url)\t\(.keyword_value)"' \
  "$INPUT_FILE" > keyword-monitors.txt

# Ping monitors
jq -r '.monitors[] | select(.type == 3) | "\(.id)\t\(.friendly_name)\t\(.url)"' \
  "$INPUT_FILE" > ping-monitors.txt

# Port monitors
jq -r '.monitors[] | select(.type == 4) | "\(.id)\t\(.friendly_name)\t\(.url)\t\(.port)"' \
  "$INPUT_FILE" > port-monitors.txt

# Heartbeat monitors
jq -r '.monitors[] | select(.type == 5) | "\(.id)\t\(.friendly_name)\t\(.interval)"' \
  "$INPUT_FILE" > heartbeat-monitors.txt

echo "Categorization complete."
echo "HTTP monitors: $(wc -l < http-monitors.txt)"
echo "Keyword monitors: $(wc -l < keyword-monitors.txt)"
echo "Ping monitors: $(wc -l < ping-monitors.txt)"
echo "Port monitors: $(wc -l < port-monitors.txt)"
echo "Heartbeat monitors: $(wc -l < heartbeat-monitors.txt)"
```

### Generate Terraform Configuration

Automated script to convert UptimeRobot export to Terraform configuration:

```bash
#!/bin/bash
# generate-terraform.sh

INPUT_FILE="uptimerobot-export/monitors-full.json"
OUTPUT_FILE="uptimerobot-migrated.tf"

cat > "$OUTPUT_FILE" <<'EOF'
# Generated from UptimeRobot export
# Review and adjust as needed before applying

terraform {
  required_providers {
    hyperping = {
      source = "develeap/hyperping"
    }
  }
}

provider "hyperping" {}

EOF

# Process each monitor
jq -c '.monitors[]' "$INPUT_FILE" | while read -r monitor; do
  id=$(echo "$monitor" | jq -r '.id')
  name=$(echo "$monitor" | jq -r '.friendly_name')
  url=$(echo "$monitor" | jq -r '.url')
  type=$(echo "$monitor" | jq -r '.type')
  interval=$(echo "$monitor" | jq -r '.interval')

  # Sanitize resource name
  resource_name=$(echo "$name" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/_/g')

  case $type in
    1) # HTTP/HTTPS
      method=$(echo "$monitor" | jq -r '.http_method // 1')
      http_method="GET"
      case $method in
        2) http_method="POST" ;;
        3) http_method="PUT" ;;
        4) http_method="PATCH" ;;
        5) http_method="DELETE" ;;
        6) http_method="HEAD" ;;
      esac

      cat >> "$OUTPUT_FILE" <<RESOURCE

resource "hyperping_monitor" "${resource_name}" {
  name                 = "${name}"
  url                  = "${url}"
  protocol             = "http"
  http_method          = "${http_method}"
  check_frequency      = ${interval}
  expected_status_code = "2xx"

  regions = ["london", "virginia", "singapore"]
}
RESOURCE
      ;;

    2) # Keyword
      keyword=$(echo "$monitor" | jq -r '.keyword_value')
      cat >> "$OUTPUT_FILE" <<RESOURCE

resource "hyperping_monitor" "${resource_name}" {
  name                 = "${name}"
  url                  = "${url}"
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = ${interval}
  expected_status_code = "200"
  required_keyword     = "${keyword}"

  regions = ["london", "virginia"]
}
RESOURCE
      ;;

    3) # Ping
      cat >> "$OUTPUT_FILE" <<RESOURCE

resource "hyperping_monitor" "${resource_name}" {
  name            = "${name}"
  url             = "${url}"
  protocol        = "icmp"
  check_frequency = ${interval}

  regions = ["london", "virginia"]
}
RESOURCE
      ;;

    4) # Port
      port=$(echo "$monitor" | jq -r '.port')
      cat >> "$OUTPUT_FILE" <<RESOURCE

resource "hyperping_monitor" "${resource_name}" {
  name            = "${name}"
  url             = "${url}"
  protocol        = "port"
  port            = ${port}
  check_frequency = ${interval}

  regions = ["virginia"]
}
RESOURCE
      ;;

    5) # Heartbeat
      cat >> "$OUTPUT_FILE" <<RESOURCE

resource "hyperping_healthcheck" "${resource_name}" {
  name               = "${name}"
  period_value       = $((interval / 86400))
  period_type        = "days"
  grace_period_value = 1
  grace_period_type  = "hours"
}
RESOURCE
      ;;
  esac
done

echo "Terraform configuration generated: $OUTPUT_FILE"
echo "Review the file and adjust as needed before running terraform apply"
```

## Import Workflow

### Parallel Operation Strategy

Run both UptimeRobot and Hyperping in parallel during the migration:

**Week 1: Setup**

```bash
# 1. Create Hyperping monitors from generated config
terraform init
terraform plan -out=tfplan
terraform apply tfplan

# 2. Keep UptimeRobot monitors active
# 3. Compare alerting for 7 days
```

**Week 2: Validation**

```bash
# Monitor both platforms
# Verify Hyperping catches same outages as UptimeRobot
# Adjust check frequencies and regions as needed
```

**Week 3: Cutover**

```bash
# 1. Pause UptimeRobot monitors
curl -X POST https://api.uptimerobot.com/v2/editMonitor \
  -d "api_key=$UPTIMEROBOT_API_KEY" \
  -d "id=MONITOR_ID" \
  -d "status=0"  # 0 = paused

# 2. Verify Hyperping is primary alerting source
# 3. Update runbooks and documentation
```

### Step-by-Step Migration

**Step 1: Export and Generate**

```bash
# Export from UptimeRobot
./export-uptimerobot.sh

# Categorize monitors
./categorize-monitors.sh

# Generate Terraform config
./generate-terraform.sh

# Review generated config
less uptimerobot-migrated.tf
```

**Step 2: Create Escalation Policies**

Manual step in Hyperping dashboard:

1. Go to Settings → Escalation Policies
2. Create policies for different severity levels:
   - **Critical**: Immediate email + SMS + PagerDuty
   - **Warning**: Email after 5 minutes
   - **Info**: Email only
3. Note the UUIDs for each policy

**Step 3: Update Terraform Config**

```bash
# Add escalation policies to variables
cat >> variables.tf <<'EOF'
variable "critical_escalation_policy" {
  description = "Escalation policy UUID for critical alerts"
  type        = string
  default     = "ep_your_critical_uuid"
}

variable "warning_escalation_policy" {
  description = "Escalation policy UUID for warning alerts"
  type        = string
  default     = "ep_your_warning_uuid"
}
EOF

# Update monitors to use escalation policies
sed -i 's/regions = \[/escalation_policy = var.critical_escalation_policy\n  regions = [/' uptimerobot-migrated.tf
```

**Step 4: Initialize Terraform**

```bash
# Initialize provider
terraform init

# Validate configuration
terraform validate

# Review plan
terraform plan -out=migration.tfplan
```

**Step 5: Apply in Stages**

```bash
# Apply monitors only (test with small batch first)
terraform apply -target=hyperping_monitor.production_api

# Verify monitor appears in Hyperping dashboard
# Check that alerts are configured correctly

# Apply remaining monitors
terraform apply migration.tfplan
```

**Step 6: Parallel Testing**

```bash
# Create test script to verify monitors
cat > verify-monitors.sh <<'EOF'
#!/bin/bash

HYPERPING_API_KEY="${HYPERPING_API_KEY}"

# Get all monitors
monitors=$(curl -s -H "Authorization: Bearer $HYPERPING_API_KEY" \
  https://api.hyperping.io/v1/monitors)

# Count monitors
total=$(echo "$monitors" | jq 'length')
echo "Total Hyperping monitors: $total"

# Check for any down monitors
down_count=$(echo "$monitors" | jq '[.[] | select(.down == true)] | length')
echo "Monitors currently down: $down_count"

if [ $down_count -gt 0 ]; then
  echo "Down monitors:"
  echo "$monitors" | jq -r '.[] | select(.down == true) | "\(.name) - \(.url)"'
fi
EOF

chmod +x verify-monitors.sh
./verify-monitors.sh
```

**Step 7: Update Heartbeat Scripts**

For migrated heartbeat monitors (now healthchecks):

```bash
# Get ping URLs from Terraform output
terraform output -json | jq -r '.[] | select(.value | type == "object") | select(.value.ping_url) | .value.ping_url' > healthcheck-urls.txt

# Update cron jobs
# Example: Update backup script
cat >> /path/to/backup.sh <<'EOF'

# Ping Hyperping healthcheck
PING_URL="https://ping.hyperping.io/hc_your_uuid"
curl -fsS --retry 3 "$PING_URL" > /dev/null
EOF
```

## Complete Migration Examples

### Example 1: Simple Website Monitoring

**UptimeRobot Configuration:**

```json
{
  "id": 78901234,
  "friendly_name": "Company Website",
  "url": "https://www.example.com",
  "type": 1,
  "interval": 300,
  "alert_contacts": ["123456"]
}
```

**Hyperping Terraform:**

```hcl
resource "hyperping_monitor" "company_website" {
  name                 = "Company Website"
  url                  = "https://www.example.com"
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 300
  expected_status_code = "200"
  follow_redirects     = true

  regions = ["london", "virginia", "singapore", "tokyo"]

  escalation_policy = var.primary_escalation_policy
}

output "website_monitor_id" {
  value = hyperping_monitor.company_website.id
}
```

### Example 2: API with Keyword Check

**UptimeRobot Configuration:**

```json
{
  "id": 78901235,
  "friendly_name": "API Health Endpoint",
  "url": "https://api.example.com/health",
  "type": 2,
  "keyword_type": 1,
  "keyword_value": "\"status\":\"healthy\"",
  "interval": 60,
  "http_method": 1
}
```

**Hyperping Terraform:**

```hcl
resource "hyperping_monitor" "api_health" {
  name                 = "API Health Endpoint"
  url                  = "https://api.example.com/health"
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 60
  expected_status_code = "200"
  required_keyword     = "\"status\":\"healthy\""

  regions = ["london", "virginia", "singapore"]

  request_headers = [
    {
      name  = "Accept"
      value = "application/json"
    }
  ]

  escalation_policy = var.critical_escalation_policy
}
```

### Example 3: Multi-Service Infrastructure

Complete example migrating a full stack:

**UptimeRobot Setup:**
- 1 website monitor
- 2 API monitors
- 1 database port check
- 1 daily backup heartbeat

**Hyperping Terraform Configuration:**

```hcl
# variables.tf
variable "environment" {
  description = "Environment name"
  type        = string
  default     = "production"
}

variable "critical_escalation_policy" {
  description = "Critical alerts escalation policy"
  type        = string
}

# main.tf
terraform {
  required_providers {
    hyperping = {
      source = "develeap/hyperping"
    }
  }
}

provider "hyperping" {}

# Website monitor
resource "hyperping_monitor" "website" {
  name                 = "[${upper(var.environment)}] Company Website"
  url                  = "https://www.example.com"
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 300
  expected_status_code = "200"
  follow_redirects     = true

  regions = ["london", "virginia", "singapore", "tokyo"]

  escalation_policy = var.critical_escalation_policy
}

# Public API monitor
resource "hyperping_monitor" "api_public" {
  name                 = "[${upper(var.environment)}] Public API"
  url                  = "https://api.example.com/v1/health"
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 60
  expected_status_code = "200"
  required_keyword     = "healthy"

  regions = ["london", "virginia", "singapore"]

  request_headers = [
    {
      name  = "Accept"
      value = "application/json"
    }
  ]

  escalation_policy = var.critical_escalation_policy
}

# Internal API monitor with auth
resource "hyperping_monitor" "api_internal" {
  name                 = "[${upper(var.environment)}] Internal API"
  url                  = "https://internal-api.example.com/health"
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 120
  expected_status_code = "200"

  regions = ["virginia"]

  request_headers = [
    {
      name  = "Authorization"
      value = "Bearer ${var.internal_api_token}"
    }
  ]

  escalation_policy = var.critical_escalation_policy
}

# Database port check
resource "hyperping_monitor" "database_port" {
  name            = "[${upper(var.environment)}] Database Port"
  url             = "db.example.com"
  protocol        = "port"
  port            = 5432
  check_frequency = 120

  regions = ["virginia"]

  escalation_policy = var.critical_escalation_policy
}

# Daily backup healthcheck
resource "hyperping_healthcheck" "daily_backup" {
  name               = "[${upper(var.environment)}] Daily Backup"
  cron               = "0 2 * * *"  # 2 AM daily
  timezone           = "America/New_York"
  grace_period_value = 30
  grace_period_type  = "minutes"

  escalation_policy = var.critical_escalation_policy
}

# outputs.tf
output "monitor_ids" {
  description = "All monitor UUIDs"
  value = {
    website       = hyperping_monitor.website.id
    api_public    = hyperping_monitor.api_public.id
    api_internal  = hyperping_monitor.api_internal.id
    database_port = hyperping_monitor.database_port.id
  }
}

output "backup_ping_url" {
  description = "URL for backup script to ping"
  value       = hyperping_healthcheck.daily_backup.ping_url
  sensitive   = true
}

output "migration_summary" {
  description = "Migration summary"
  value = {
    total_monitors    = 4
    total_healthchecks = 1
    environment       = var.environment
  }
}
```

**Apply the configuration:**

```bash
# Set variables
export HYPERPING_API_KEY="sk_your_key"
export TF_VAR_critical_escalation_policy="ep_your_uuid"
export TF_VAR_internal_api_token="your_token"

# Initialize
terraform init

# Plan
terraform plan -out=migration.tfplan

# Apply
terraform apply migration.tfplan

# Get backup ping URL
terraform output -raw backup_ping_url > backup-ping-url.txt

# Update backup script
echo "" >> /path/to/backup.sh
echo "# Ping Hyperping healthcheck" >> /path/to/backup.sh
echo "curl -fsS --retry 3 '$(cat backup-ping-url.txt)' > /dev/null" >> /path/to/backup.sh
```

### Example 4: Multi-Environment Migration

Migrate separate production and staging environments:

```
environments/
├── production/
│   ├── main.tf
│   ├── variables.tf
│   └── terraform.tfvars
├── staging/
│   ├── main.tf
│   ├── variables.tf
│   └── terraform.tfvars
└── modules/
    └── monitoring/
        ├── main.tf
        ├── variables.tf
        └── outputs.tf
```

**Module: modules/monitoring/main.tf**

```hcl
variable "environment" {
  type = string
}

variable "base_url" {
  type = string
}

variable "escalation_policy" {
  type = string
}

resource "hyperping_monitor" "website" {
  name                 = "[${upper(var.environment)}] Website"
  url                  = var.base_url
  protocol             = "http"
  check_frequency      = var.environment == "production" ? 60 : 300
  expected_status_code = "200"

  regions = var.environment == "production" ? [
    "london", "virginia", "singapore", "tokyo"
  ] : [
    "virginia"
  ]

  escalation_policy = var.escalation_policy
}

resource "hyperping_monitor" "api" {
  name                 = "[${upper(var.environment)}] API"
  url                  = "${var.base_url}/api/health"
  protocol             = "http"
  check_frequency      = var.environment == "production" ? 60 : 300
  expected_status_code = "200"
  required_keyword     = "healthy"

  regions = var.environment == "production" ? [
    "london", "virginia", "singapore"
  ] : [
    "virginia"
  ]

  escalation_policy = var.escalation_policy
}

output "monitor_ids" {
  value = {
    website = hyperping_monitor.website.id
    api     = hyperping_monitor.api.id
  }
}
```

**Production: environments/production/main.tf**

```hcl
terraform {
  required_providers {
    hyperping = {
      source = "develeap/hyperping"
    }
  }

  backend "s3" {
    bucket = "terraform-state"
    key    = "hyperping/production/terraform.tfstate"
    region = "us-east-1"
  }
}

provider "hyperping" {}

module "monitoring" {
  source = "../../modules/monitoring"

  environment       = "production"
  base_url          = "https://www.example.com"
  escalation_policy = var.prod_escalation_policy
}

output "production_monitors" {
  value = module.monitoring.monitor_ids
}
```

**Staging: environments/staging/main.tf**

```hcl
terraform {
  required_providers {
    hyperping = {
      source = "develeap/hyperping"
    }
  }

  backend "s3" {
    bucket = "terraform-state"
    key    = "hyperping/staging/terraform.tfstate"
    region = "us-east-1"
  }
}

provider "hyperping" {}

module "monitoring" {
  source = "../../modules/monitoring"

  environment       = "staging"
  base_url          = "https://staging.example.com"
  escalation_policy = var.staging_escalation_policy
}

output "staging_monitors" {
  value = module.monitoring.monitor_ids
}
```

**Apply environments:**

```bash
# Production
cd environments/production
terraform init
terraform plan -out=prod.tfplan
terraform apply prod.tfplan

# Staging
cd ../staging
terraform init
terraform plan -out=staging.tfplan
terraform apply staging.tfplan
```

## Validation and Testing

### Pre-Migration Validation

Before cutting over from UptimeRobot:

```bash
#!/bin/bash
# pre-migration-validation.sh

echo "Pre-Migration Validation Checklist"
echo "==================================="

# 1. Check Hyperping monitors are created
echo -n "1. Hyperping monitors created: "
monitor_count=$(curl -s -H "Authorization: Bearer $HYPERPING_API_KEY" \
  https://api.hyperping.io/v1/monitors | jq 'length')
echo "$monitor_count monitors"

# 2. Verify no monitors are down
echo -n "2. All monitors operational: "
down_count=$(curl -s -H "Authorization: Bearer $HYPERPING_API_KEY" \
  https://api.hyperping.io/v1/monitors | jq '[.[] | select(.down == true)] | length')
if [ $down_count -eq 0 ]; then
  echo "✓ All monitors up"
else
  echo "✗ $down_count monitors down - investigate before migration"
  exit 1
fi

# 3. Check escalation policies are linked
echo "3. Checking escalation policies..."
monitors_without_ep=$(curl -s -H "Authorization: Bearer $HYPERPING_API_KEY" \
  https://api.hyperping.io/v1/monitors | jq '[.[] | select(.escalation_policy == null)] | length')
echo "   Monitors without escalation policy: $monitors_without_ep"

# 4. Verify Terraform state
echo -n "4. Terraform state valid: "
if terraform state list > /dev/null 2>&1; then
  echo "✓ State valid"
  echo "   Resources in state: $(terraform state list | wc -l)"
else
  echo "✗ State invalid"
  exit 1
fi

# 5. Compare counts
echo "5. Comparing monitor counts..."
uptimerobot_count=$(curl -s -X POST https://api.uptimerobot.com/v2/getMonitors \
  -d "api_key=$UPTIMEROBOT_API_KEY" \
  -d "format=json" | jq '.monitors | length')
echo "   UptimeRobot monitors: $uptimerobot_count"
echo "   Hyperping monitors: $monitor_count"

if [ $monitor_count -eq $uptimerobot_count ]; then
  echo "   ✓ Counts match"
else
  echo "   ⚠ Counts differ - verify migration is complete"
fi

echo ""
echo "Validation complete. Review results before proceeding with cutover."
```

### Parallel Monitoring Comparison

Run both platforms and compare results:

```bash
#!/bin/bash
# compare-monitoring.sh

echo "Monitoring Comparison Report"
echo "Generated: $(date)"
echo "============================"

# UptimeRobot status
echo ""
echo "UptimeRobot Status:"
uptimerobot_data=$(curl -s -X POST https://api.uptimerobot.com/v2/getMonitors \
  -d "api_key=$UPTIMEROBOT_API_KEY" \
  -d "format=json")

total_ur=$(echo "$uptimerobot_data" | jq '.monitors | length')
down_ur=$(echo "$uptimerobot_data" | jq '[.monitors[] | select(.status != 2)] | length')

echo "  Total monitors: $total_ur"
echo "  Down monitors: $down_ur"

if [ $down_ur -gt 0 ]; then
  echo "  Down monitors:"
  echo "$uptimerobot_data" | jq -r '.monitors[] | select(.status != 2) | "    - \(.friendly_name) (\(.url))"'
fi

# Hyperping status
echo ""
echo "Hyperping Status:"
hyperping_data=$(curl -s -H "Authorization: Bearer $HYPERPING_API_KEY" \
  https://api.hyperping.io/v1/monitors)

total_hp=$(echo "$hyperping_data" | jq 'length')
down_hp=$(echo "$hyperping_data" | jq '[.[] | select(.down == true)] | length')

echo "  Total monitors: $total_hp"
echo "  Down monitors: $down_hp"

if [ $down_hp -gt 0 ]; then
  echo "  Down monitors:"
  echo "$hyperping_data" | jq -r '.[] | select(.down == true) | "    - \(.name) (\(.url))"'
fi

# Alert if mismatch
echo ""
if [ $down_ur -ne $down_hp ]; then
  echo "⚠ WARNING: Down monitor count mismatch!"
  echo "  UptimeRobot: $down_ur down"
  echo "  Hyperping: $down_hp down"
  echo "  Investigate discrepancies before cutover."
else
  echo "✓ Monitor statuses match between platforms"
fi
```

### Post-Migration Verification

After switching to Hyperping:

```bash
#!/bin/bash
# post-migration-verification.sh

echo "Post-Migration Verification"
echo "==========================="

# 1. Verify all Terraform resources exist
echo "1. Checking Terraform state..."
resource_count=$(terraform state list | wc -l)
echo "   Terraform resources: $resource_count"

# 2. Verify monitors in Hyperping API
echo "2. Checking Hyperping API..."
api_monitor_count=$(curl -s -H "Authorization: Bearer $HYPERPING_API_KEY" \
  https://api.hyperping.io/v1/monitors | jq 'length')
echo "   API monitor count: $api_monitor_count"

# 3. Run terraform plan (should show no changes)
echo "3. Running terraform plan..."
if terraform plan -detailed-exitcode > /dev/null 2>&1; then
  echo "   ✓ No infrastructure drift detected"
else
  echo "   ✗ Infrastructure drift detected - run 'terraform plan' to review"
fi

# 4. Check for recent alerts
echo "4. Checking recent alerts..."
# Note: Adjust based on Hyperping alert API if available
echo "   Manual check: Review Hyperping dashboard for recent alerts"

# 5. Verify healthcheck ping URLs are updated
echo "5. Healthcheck verification..."
terraform output -json | jq -r 'to_entries[] | select(.value.sensitive == false) | select(.value.value | type == "string") | select(.value.value | contains("ping.hyperping.io")) | "   ✓ \(.key): \(.value.value)"'

echo ""
echo "Verification complete."
echo "Next steps:"
echo "  1. Monitor Hyperping dashboard for 24 hours"
echo "  2. Verify all alerts are received"
echo "  3. Update runbooks and documentation"
echo "  4. Pause/delete UptimeRobot monitors after 1 week"
```

## Troubleshooting

### Issue: Check Frequency Not Supported

**Problem:**

```
Error: Invalid check_frequency value
UptimeRobot interval of 45 seconds is not supported
```

**Solution:**

Map to nearest allowed Hyperping value:

```hcl
# UptimeRobot: 45 seconds → Hyperping: 30 or 60
check_frequency = 60  # Round up for safety
```

**Mapping script:**

```bash
# Round to nearest allowed value
round_frequency() {
  local ur_freq=$1
  local allowed=(10 20 30 60 120 180 300 600 1800 3600 21600 43200 86400)

  for freq in "${allowed[@]}"; do
    if [ $freq -ge $ur_freq ]; then
      echo $freq
      return
    fi
  done

  echo 86400  # Default to max if larger
}

# Usage
round_frequency 45   # Returns 60
round_frequency 150  # Returns 180
```

### Issue: Keyword Check Not Working

**Problem:**

```
Monitor passes in Hyperping but failed in UptimeRobot for keyword absence
```

**Solution:**

UptimeRobot supports both "keyword exists" and "keyword does not exist". Hyperping only supports "keyword exists".

For "keyword must not exist" scenarios:

1. **Option 1:** Create a validation endpoint that returns different keywords
2. **Option 2:** Use a proxy service that validates and returns status
3. **Option 3:** Accept this limitation and rely on status code only

```hcl
# Workaround: Use required_keyword for positive check
resource "hyperping_monitor" "with_keyword" {
  name                 = "API Health"
  url                  = "https://api.example.com/health"
  protocol             = "http"
  expected_status_code = "200"
  required_keyword     = "\"status\":\"ok\""  # Must exist
}

# For negative checks, update your endpoint:
# /health should return 200 with "status":"ok" for healthy
# /health should return 503 with "status":"error" for unhealthy
# Then check for 200 status instead of keyword absence
```

### Issue: Too Many Monitors to Migrate

**Problem:**

```
Have 500+ monitors in UptimeRobot, manual migration is impractical
```

**Solution:**

Use the automated generation script and import in batches:

```bash
# Generate all Terraform configs
./generate-terraform.sh

# Split into batches of 50
split -l 50 -d --additional-suffix=.tf uptimerobot-migrated.tf batch_

# Apply in batches
for batch in batch_*.tf; do
  echo "Applying $batch..."
  terraform apply -target="$(grep 'resource "hyperping_' $batch | head -1 | awk '{print $2 "." $3}' | tr -d '"')"
  sleep 5  # Rate limiting
done
```

### Issue: Regional Availability Differences

**Problem:**

```
UptimeRobot checks from locations not available in Hyperping
```

**Solution:**

Map UptimeRobot locations to nearest Hyperping regions:

| UptimeRobot Location | Hyperping Region |
|---------------------|------------------|
| US East | virginia |
| US West | oregon (or virginia) |
| EU | london, frankfurt |
| Asia | singapore, tokyo |
| Australia | sydney |
| South America | saopaulo |

```hcl
# Select regions based on your users' geographic distribution
resource "hyperping_monitor" "global" {
  name = "Global Service"
  url  = "https://example.com"

  # Optimal coverage
  regions = [
    "virginia",    # North America
    "london",      # Europe
    "singapore",   # Asia
    "sydney"       # Australia/Oceania
  ]
}
```

### Issue: Alert Contact Migration

**Problem:**

```
Have 50+ email contacts in UptimeRobot, need to migrate
```

**Solution:**

Create escalation policies in Hyperping dashboard for different contact groups:

1. **Extract UptimeRobot contacts:**

```bash
curl -X POST https://api.uptimerobot.com/v2/getAlertContacts \
  -d "api_key=$UPTIMEROBOT_API_KEY" \
  -d "format=json" \
  | jq -r '.alert_contacts[] | select(.type == 2) | .value' \
  > email-list.txt
```

2. **Group by purpose:**

```bash
# Critical team
grep -E 'oncall|ops|critical' email-list.txt > critical-emails.txt

# General notifications
grep -v -E 'oncall|ops|critical' email-list.txt > general-emails.txt
```

3. **Create escalation policies in Hyperping dashboard using these email lists**

4. **Reference in Terraform:**

```hcl
variable "escalation_policies" {
  type = map(string)
  default = {
    critical = "ep_critical_uuid"
    general  = "ep_general_uuid"
  }
}
```

### Issue: Heartbeat Ping Not Received

**Problem:**

```
Migrated heartbeat to healthcheck but not receiving pings
```

**Solution:**

Verify the ping URL and update scripts:

```bash
# Get ping URL
terraform output -raw healthcheck_ping_url

# Test manually
curl -v "https://ping.hyperping.io/hc_your_uuid"

# Check cron job updated
crontab -l | grep hyperping

# Verify script has correct URL
grep -r "ping.hyperping.io" /path/to/scripts/
```

**Update cron job:**

```bash
# Before (UptimeRobot)
0 2 * * * /usr/local/bin/backup.sh && curl "https://uptimerobot.com/api/heartbeat/12345"

# After (Hyperping)
0 2 * * * /usr/local/bin/backup.sh && curl -fsS --retry 3 "https://ping.hyperping.io/hc_your_uuid"
```

### Issue: HTTP Authentication Not Working

**Problem:**

```
UptimeRobot monitor has HTTP Basic Auth, Hyperping monitor fails
```

**Solution:**

Hyperping supports custom headers for authentication:

```hcl
# Basic Auth (encode username:password in base64)
resource "hyperping_monitor" "with_basic_auth" {
  name     = "Protected Endpoint"
  url      = "https://protected.example.com"
  protocol = "http"

  request_headers = [
    {
      name  = "Authorization"
      value = "Basic ${base64encode("username:password")}"
    }
  ]
}

# Or use variable for security
variable "basic_auth_credentials" {
  type      = string
  sensitive = true
}

resource "hyperping_monitor" "secure" {
  name     = "Protected API"
  url      = "https://api.example.com"
  protocol = "http"

  request_headers = [
    {
      name  = "Authorization"
      value = "Basic ${var.basic_auth_credentials}"
    }
  ]
}
```

## Best Practices

### Pre-Migration

**1. Document Current Setup**

```bash
# Export full UptimeRobot configuration
./export-uptimerobot.sh

# Document all alert contacts
curl -X POST https://api.uptimerobot.com/v2/getAlertContacts \
  -d "api_key=$UPTIMEROBOT_API_KEY" \
  -d "format=json" > alert-contacts-backup.json

# Save monitor response time data (for comparison)
# This helps validate Hyperping is performing similarly
```

**2. Create Test Monitors First**

```hcl
# Create a test monitor to validate setup
resource "hyperping_monitor" "test" {
  name                 = "MIGRATION TEST - Delete After"
  url                  = "https://httpstat.us/200"
  protocol             = "http"
  check_frequency      = 60
  expected_status_code = "200"

  regions = ["virginia"]
}
```

**3. Set Up Terraform Backend**

```hcl
terraform {
  backend "s3" {
    bucket         = "my-terraform-state"
    key            = "hyperping/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-locks"
  }
}
```

### During Migration

**1. Use Modules for Repeated Patterns**

```hcl
# modules/http-monitor/main.tf
variable "name" {
  type = string
}

variable "url" {
  type = string
}

variable "environment" {
  type    = string
  default = "production"
}

resource "hyperping_monitor" "this" {
  name                 = "[${upper(var.environment)}] ${var.name}"
  url                  = var.url
  protocol             = "http"
  check_frequency      = var.environment == "production" ? 60 : 300
  expected_status_code = "2xx"

  regions = var.environment == "production" ? [
    "london", "virginia", "singapore", "tokyo"
  ] : ["virginia"]
}

# Use it
module "api_monitor" {
  source = "./modules/http-monitor"

  name        = "Public API"
  url         = "https://api.example.com/health"
  environment = "production"
}
```

**2. Tag Resources for Organization**

```hcl
# Use naming conventions
resource "hyperping_monitor" "prod_api_public" {
  name = "[PROD] API - Public Endpoints"
  # ...
}

resource "hyperping_monitor" "staging_api_public" {
  name = "[STAGING] API - Public Endpoints"
  # ...
}
```

**3. Run in Parallel for Safety**

Keep UptimeRobot running for at least 1 week while Hyperping runs in parallel. This provides:
- Validation that Hyperping catches the same issues
- Fallback if Hyperping has problems
- Time to tune check frequencies and regions

### Post-Migration

**1. Document New Workflow**

Create a runbook for your team:

```markdown
# Monitoring Runbook

## Adding a New Monitor

1. Edit `monitors.tf`
2. Add monitor resource:
   ```hcl
   resource "hyperping_monitor" "new_service" {
     name = "[PROD] New Service"
     url  = "https://new.example.com"
     # ...
   }
   ```
3. Run `terraform plan`
4. Create PR for review
5. After approval: `terraform apply`

## Viewing Monitor Status

- Dashboard: https://app.hyperping.io/monitors
- CLI: `./scripts/check-monitors.sh`

## Emergency: Pause a Monitor

terraform apply -var="pause_monitor_id=mon_xxx"
```

**2. Set Up Monitoring Alerts**

Monitor your monitoring:

```hcl
# Monitor Hyperping API itself
resource "hyperping_monitor" "hyperping_api" {
  name                 = "Hyperping API Health"
  url                  = "https://api.hyperping.io/health"
  protocol             = "http"
  check_frequency      = 300
  expected_status_code = "200"

  regions = ["virginia"]
}
```

**3. Regular Reviews**

Schedule quarterly reviews:
- Review check frequencies (too often = high cost, too rare = missed issues)
- Validate escalation policies still correct
- Remove deprecated monitors
- Update regions based on user distribution

```bash
# Generate monitor audit report
terraform state list | grep hyperping_monitor | while read resource; do
  echo "Resource: $resource"
  terraform state show "$resource" | grep -E "name|url|check_frequency|regions"
  echo "---"
done > monitor-audit-report.txt
```

**4. Cost Optimization**

Monitor your Hyperping usage:

```bash
# Count monitors
total_monitors=$(terraform state list | grep -c hyperping_monitor)

# Count checks per day
# Assuming average check_frequency of 300 seconds
daily_checks=$((total_monitors * 86400 / 300))

echo "Total monitors: $total_monitors"
echo "Estimated daily checks: $daily_checks"
```

Consider reducing frequency for non-critical monitors:

```hcl
# Production - frequent checks
resource "hyperping_monitor" "prod_api" {
  name            = "[PROD] API"
  url             = "https://api.example.com"
  check_frequency = 60  # Every minute
  # ...
}

# Staging - less frequent
resource "hyperping_monitor" "staging_api" {
  name            = "[STAGING] API"
  url             = "https://staging-api.example.com"
  check_frequency = 600  # Every 10 minutes
  # ...
}

# Development - minimal monitoring
resource "hyperping_monitor" "dev_api" {
  name            = "[DEV] API"
  url             = "https://dev-api.example.com"
  check_frequency = 3600  # Hourly
  # ...
}
```

### Security Considerations

**1. Protect Sensitive Data**

```hcl
# Never hardcode API keys
# Bad:
# api_key = "sk_abc123..."

# Good: Use environment variable
provider "hyperping" {
  # Reads from HYPERPING_API_KEY automatically
}

# For monitor auth tokens
variable "api_auth_token" {
  type      = string
  sensitive = true
}

resource "hyperping_monitor" "secure_api" {
  name     = "Authenticated API"
  url      = "https://api.example.com"
  protocol = "http"

  request_headers = [
    {
      name  = "Authorization"
      value = "Bearer ${var.api_auth_token}"
    }
  ]
}
```

**2. Use Remote State with Encryption**

```hcl
terraform {
  backend "s3" {
    bucket         = "terraform-state-bucket"
    key            = "hyperping/prod/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true  # Encryption at rest
    kms_key_id     = "arn:aws:kms:us-east-1:123456789:key/xxx"
    dynamodb_table = "terraform-locks"
  }
}
```

**3. Limit Terraform Permissions**

Create IAM policy for Terraform:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:PutObject",
        "s3:DeleteObject"
      ],
      "Resource": "arn:aws:s3:::terraform-state-bucket/hyperping/*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "dynamodb:GetItem",
        "dynamodb:PutItem",
        "dynamodb:DeleteItem"
      ],
      "Resource": "arn:aws:dynamodb:*:*:table/terraform-locks"
    }
  ]
}
```

## Summary

You now have a complete migration path from UptimeRobot to Hyperping:

- ✅ **Monitor type mapping** - Convert all UptimeRobot monitor types to Hyperping equivalents
- ✅ **Alert migration** - Map alert contacts to escalation policies
- ✅ **Automated export** - Scripts to export and convert configurations
- ✅ **Complete examples** - Working examples for common scenarios
- ✅ **Validation tools** - Scripts to verify migration success
- ✅ **Troubleshooting** - Solutions to common migration issues
- ✅ **Best practices** - Proven patterns for successful migrations

### Migration Checklist

**Planning Phase:**
- [ ] Export UptimeRobot configurations
- [ ] Inventory all monitors and alert contacts
- [ ] Create Hyperping account and generate API key
- [ ] Create escalation policies in Hyperping dashboard
- [ ] Set up Terraform with remote state

**Migration Phase:**
- [ ] Generate Terraform configurations from UptimeRobot export
- [ ] Review and customize generated configs
- [ ] Apply Terraform to create Hyperping monitors
- [ ] Run both platforms in parallel for 1-2 weeks
- [ ] Validate alert delivery from Hyperping
- [ ] Update healthcheck ping URLs in scripts

**Cutover Phase:**
- [ ] Verify Hyperping monitors are stable
- [ ] Switch primary alerting to Hyperping
- [ ] Pause UptimeRobot monitors
- [ ] Update team documentation and runbooks
- [ ] Monitor Hyperping for 1 week

**Cleanup Phase:**
- [ ] Delete UptimeRobot monitors (after 1 week parallel)
- [ ] Cancel UptimeRobot subscription
- [ ] Archive migration scripts and documentation
- [ ] Train team on new Terraform workflow

### Next Steps

1. **Start small** - Migrate a few non-critical monitors first
2. **Validate** - Run in parallel for at least one week
3. **Scale up** - Migrate remaining monitors in batches
4. **Document** - Update team processes and runbooks
5. **Optimize** - Review and adjust check frequencies and regions

### Additional Resources

- [Hyperping Provider Documentation](https://registry.terraform.io/providers/develeap/hyperping)
- [Importing Existing Resources Guide](importing-resources.md)
- [Validation Guide](validation.md)
- [Error Handling Guide](error-handling.md)
- [Hyperping API Documentation](https://hyperping.io/docs/api)

**Need help?**
- GitHub Issues: [terraform-provider-hyperping/issues](https://github.com/develeap/terraform-provider-hyperping/issues)
- Hyperping Support: support@hyperping.io

Good luck with your migration!
