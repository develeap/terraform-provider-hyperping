---
page_title: "Importing Existing Hyperping Resources"
subcategory: "Guides"
description: |-
  Complete guide to importing existing Hyperping resources into Terraform management.
---

# Importing Existing Hyperping Resources

This guide provides comprehensive instructions for importing existing Hyperping resources into Terraform management. Whether you have one resource or hundreds, this guide covers everything from basic imports to bulk operations using the import-generator tool.

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Quick Start (3 minutes)](#quick-start-3-minutes)
- [Resource-Specific Import Guides](#resource-specific-import-guides)
  - [Monitor](#importing-a-monitor)
  - [Healthcheck](#importing-a-healthcheck)
  - [Status Page](#importing-a-status-page)
  - [Status Page Subscriber](#importing-a-status-page-subscriber)
  - [Incident](#importing-an-incident)
  - [Incident Update](#importing-an-incident-update)
  - [Maintenance Window](#importing-a-maintenance-window)
  - [Outage](#importing-an-outage)
- [Bulk Import with Import-Generator](#bulk-import-with-import-generator)
- [Troubleshooting](#troubleshooting)
- [Advanced Scenarios](#advanced-scenarios)
- [Best Practices](#best-practices)

## Overview

### Why Import?

Importing allows you to bring existing Hyperping resources under Terraform management without recreating them. This is useful when:

- Migrating from manual resource management to Infrastructure as Code
- Taking over infrastructure created by another team or tool
- Recovering from state loss or corruption
- Consolidating resources from multiple accounts or environments

### When to Use Import?

Import is the right choice when:

- ✅ Resources already exist in Hyperping
- ✅ You want to manage them with Terraform going forward
- ✅ You need to preserve existing resource IDs and configurations
- ✅ Downtime or recreation is not acceptable

Do NOT import when:

- ❌ Creating new resources from scratch (use normal `terraform apply`)
- ❌ Resources don't exist yet
- ❌ You're okay with recreating resources

### Prerequisites

Before importing resources, ensure you have:

1. **Terraform installed** (version 1.8 or higher)
2. **Provider configured** with valid API key
3. **Resource UUID(s)** from Hyperping dashboard
4. **Write access** to create/modify Terraform configuration files

## Quick Start (3 minutes)

Import an existing monitor in 4 simple steps.

### Step 1: Get the Monitor UUID

Find your monitor's UUID from the Hyperping dashboard URL:

```
https://app.hyperping.io/monitors/mon_abc123def456
                                   └─────┬──────┘
                              This is your monitor UUID
```

The UUID is the last part of the URL after `/monitors/`.

### Step 2: Create Terraform Configuration

Create a minimal resource block in your `.tf` file:

```hcl
resource "hyperping_monitor" "my_api" {
  # Required fields - will be populated after import
  name = "placeholder"
  url  = "https://placeholder.com"
}
```

### Step 3: Import the Resource

Run the import command with your actual UUID:

```bash
terraform import hyperping_monitor.my_api mon_abc123def456
```

Expected output:

```
hyperping_monitor.my_api: Importing from ID "mon_abc123def456"...
hyperping_monitor.my_api: Import prepared!
  Prepared hyperping_monitor for import
hyperping_monitor.my_api: Refreshing state... [id=mon_abc123def456]

Import successful!
```

### Step 4: Update Configuration to Match State

View the imported state:

```bash
terraform show
```

Update your configuration file to match the actual resource values:

```hcl
resource "hyperping_monitor" "my_api" {
  name                 = "Production API Health Check"  # From terraform show
  url                  = "https://api.example.com/health"
  protocol             = "http"
  check_frequency      = 300
  expected_status_code = "200"

  regions = [
    "london",
    "virginia",
    "singapore"
  ]
}
```

### Step 5: Verify Import

Verify there are no differences:

```bash
terraform plan
```

Expected output:

```
No changes. Your infrastructure matches the configuration.
```

✅ **Success!** Your resource is now managed by Terraform.

## Resource-Specific Import Guides

### Importing a Monitor

**What it is:** HTTP, TCP, or ICMP uptime monitoring checks.

**Find the UUID:**
- Navigate to your monitor in the Hyperping dashboard
- URL format: `https://app.hyperping.io/monitors/{uuid}`
- UUID format: `mon_*` (e.g., `mon_abc123def456`)

**Minimal Configuration:**

```hcl
resource "hyperping_monitor" "api" {
  name = "API Health Check"
  url  = "https://api.example.com/health"
}
```

**Import Command:**

```bash
terraform import hyperping_monitor.api mon_abc123def456
```

**Common Configuration Drift Issues:**

⚠️ **Default values** - These fields have defaults that may not match your resource:
- `protocol` defaults to `"http"` - specify if using `"port"` or `"icmp"`
- `http_method` defaults to `"GET"` - specify if using POST/PUT/etc.
- `check_frequency` defaults to `60` - specify if using different interval
- `follow_redirects` defaults to `true` - specify if disabled
- `paused` defaults to `false` - specify if monitor is paused

⚠️ **Regions** - If not specified, uses all regions except Bahrain. Specify explicitly:

```hcl
regions = ["london", "frankfurt", "virginia"]
```

⚠️ **Request headers** - Must match exactly including order:

```hcl
request_headers = [
  {
    name  = "Authorization"
    value = "Bearer ${var.token}"
  }
]
```

**Full Example After Import:**

```hcl
resource "hyperping_monitor" "api" {
  name                 = "Production API Health"
  url                  = "https://api.example.com/health"
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 300
  expected_status_code = "200"
  follow_redirects     = true
  paused               = false

  regions = [
    "london",
    "virginia",
    "singapore"
  ]

  request_headers = [
    {
      name  = "Content-Type"
      value = "application/json"
    }
  ]
}
```

### Importing a Healthcheck

**What it is:** Webhook-based health monitoring for push-based checks.

**Find the UUID:**
- Navigate to your healthcheck in the Hyperping dashboard
- URL format: `https://app.hyperping.io/healthchecks/{uuid}`
- UUID format: `hc_*` (e.g., `hc_xyz789abc123`)

**Minimal Configuration:**

```hcl
resource "hyperping_healthcheck" "worker" {
  name   = "Background Worker"
  period = 60
}
```

**Import Command:**

```bash
terraform import hyperping_healthcheck.worker hc_xyz789abc123
```

**Common Configuration Drift Issues:**

⚠️ **Period** must be one of the allowed values: `30`, `60`, `120`, `300`, `600`, `1800`, `3600`, `86400` (in seconds)

⚠️ **Grace period** - Must specify if non-zero:

```hcl
grace = 300  # 5 minutes grace period
```

⚠️ **Paused state** - Defaults to `false`:

```hcl
paused = true  # If healthcheck is paused
```

**Full Example After Import:**

```hcl
resource "hyperping_healthcheck" "worker" {
  name   = "Background Job Processor"
  period = 300   # Expected every 5 minutes
  grace  = 60    # 1 minute grace period
  paused = false
}
```

### Importing a Status Page

**What it is:** Public or private status pages for communicating service health.

**Find the UUID:**
- Navigate to your status page in the Hyperping dashboard
- URL format: `https://app.hyperping.io/statuspages/{uuid}`
- UUID format: `sp_*` (e.g., `sp_status123`)

**Minimal Configuration:**

```hcl
resource "hyperping_statuspage" "main" {
  name       = "Service Status"
  subdomain  = "status"
  is_private = false
}
```

**Import Command:**

```bash
terraform import hyperping_statuspage.main sp_status123
```

**Common Configuration Drift Issues:**

⚠️ **Subdomain** is unique and immutable - must match exactly

⚠️ **Locale** defaults to `"en"` (English):

```hcl
locale = "en"  # Specify if different
```

⚠️ **Privacy** - Specify if private:

```hcl
is_private = true
```

⚠️ **Branding** - Custom colors, logo, and favicon must match:

```hcl
brand_color = "#2563eb"
logo_url    = "https://example.com/logo.png"
favicon_url = "https://example.com/favicon.ico"
```

⚠️ **Password** - Only set if status page has password protection (cannot be read back from API):

```hcl
password = var.statuspage_password  # Store in variable or secret
```

⚠️ **Monitors** - List of attached monitor UUIDs:

```hcl
monitors = [
  hyperping_monitor.api.id,
  hyperping_monitor.web.id
]
```

**Full Example After Import:**

```hcl
resource "hyperping_statuspage" "main" {
  name        = "Production Status"
  subdomain   = "status"
  locale      = "en"
  is_private  = false
  brand_color = "#2563eb"

  monitors = [
    hyperping_monitor.api.id,
    hyperping_monitor.web.id,
    hyperping_monitor.cdn.id
  ]

  custom_domain      = "status.example.com"
  custom_domain_cert = var.ssl_certificate
}
```

### Importing a Status Page Subscriber

**What it is:** Email subscribers to status page notifications.

**Find the UUID:**
- Navigate to your status page subscribers list
- UUID format: `sub_*` (e.g., `sub_email123`)

**Minimal Configuration:**

```hcl
resource "hyperping_statuspage_subscriber" "user" {
  statuspage_id = hyperping_statuspage.main.id
  email         = "user@example.com"
}
```

**Import Command:**

The import ID format is `statuspage_uuid/subscriber_uuid`:

```bash
terraform import hyperping_statuspage_subscriber.user sp_status123/sub_email123
```

**Common Configuration Drift Issues:**

⚠️ **Email** - Must match exactly (case-insensitive)

⚠️ **Verified status** - Subscribers must be verified to receive notifications (computed field, read-only)

**Full Example After Import:**

```hcl
resource "hyperping_statuspage_subscriber" "user" {
  statuspage_id = hyperping_statuspage.main.id
  email         = "admin@example.com"
}
```

### Importing an Incident

**What it is:** Service incidents reported on status pages.

**Find the UUID:**
- Navigate to your incident in the Hyperping dashboard
- URL format: `https://app.hyperping.io/incidents/{uuid}`
- UUID format: `inc_*` (e.g., `inc_incident123`)

**Minimal Configuration:**

```hcl
resource "hyperping_incident" "outage" {
  statuspage_id = hyperping_statuspage.main.id
  title = {
    en = "Service Disruption"
  }
  description = {
    en = "We are investigating service issues"
  }
  status = "investigating"
}
```

**Import Command:**

```bash
terraform import hyperping_incident.outage inc_incident123
```

**Common Configuration Drift Issues:**

⚠️ **Status** - Must be one of: `"investigating"`, `"identified"`, `"monitoring"`, `"resolved"`

⚠️ **Severity** - Must be one of: `"minor"`, `"major"`, `"critical"`

⚠️ **Multilingual content** - Title and description support multiple languages:

```hcl
title = {
  en = "Service Disruption"
  es = "Interrupción del servicio"
}
```

⚠️ **Affected monitors** - List of monitor UUIDs:

```hcl
affected_monitors = [
  hyperping_monitor.api.id,
  hyperping_monitor.web.id
]
```

**Full Example After Import:**

```hcl
resource "hyperping_incident" "outage" {
  statuspage_id = hyperping_statuspage.main.id

  title = {
    en = "Database Performance Issues"
  }

  description = {
    en = "We are experiencing slow database response times affecting API performance."
  }

  status   = "identified"
  severity = "major"

  affected_monitors = [
    hyperping_monitor.api.id
  ]
}
```

### Importing an Incident Update

**What it is:** Status updates posted to an existing incident.

**Find the UUID:**
- Navigate to the incident in Hyperping dashboard
- Find the specific update in the incident timeline
- UUID format: `upd_*` (e.g., `upd_update123`)

**Minimal Configuration:**

```hcl
resource "hyperping_incident_update" "progress" {
  incident_id = hyperping_incident.outage.id
  message = {
    en = "Update message"
  }
  status = "monitoring"
}
```

**Import Command:**

The import ID format is `incident_uuid/update_uuid`:

```bash
terraform import hyperping_incident_update.progress inc_incident123/upd_update123
```

**Common Configuration Drift Issues:**

⚠️ **Composite ID** - Must use correct format: `incident_id/update_id`

⚠️ **Status progression** - Updates typically progress: `investigating` → `identified` → `monitoring` → `resolved`

⚠️ **Message localization** - Must match incident's supported locales:

```hcl
message = {
  en = "Root cause identified, implementing fix"
  es = "Causa raíz identificada, implementando solución"
}
```

**Full Example After Import:**

```hcl
resource "hyperping_incident_update" "identified" {
  incident_id = hyperping_incident.outage.id

  message = {
    en = "Database query optimization identified as root cause. Deploying fix."
  }

  status = "identified"
}

resource "hyperping_incident_update" "monitoring" {
  incident_id = hyperping_incident.outage.id

  message = {
    en = "Fix deployed. Monitoring system performance."
  }

  status = "monitoring"

  depends_on = [hyperping_incident_update.identified]
}
```

### Importing a Maintenance Window

**What it is:** Scheduled maintenance periods that suppress alerts.

**Find the UUID:**
- Navigate to your maintenance window in the Hyperping dashboard
- URL format: `https://app.hyperping.io/maintenance/{uuid}`
- UUID format: `maint_*` (e.g., `maint_window123`)

**Minimal Configuration:**

```hcl
resource "hyperping_maintenance" "upgrade" {
  statuspage_id  = hyperping_statuspage.main.id
  title = {
    en = "Database Upgrade"
  }
  description = {
    en = "Scheduled maintenance"
  }
  scheduled_start = "2026-03-01T02:00:00Z"
  scheduled_end   = "2026-03-01T04:00:00Z"
}
```

**Import Command:**

```bash
terraform import hyperping_maintenance.upgrade maint_window123
```

**Common Configuration Drift Issues:**

⚠️ **Timestamps** - Must be in RFC3339/ISO8601 format with UTC timezone:

```hcl
scheduled_start = "2026-03-01T02:00:00Z"  # Note the Z suffix
scheduled_end   = "2026-03-01T04:00:00Z"
```

⚠️ **Affected monitors** - List of monitors to silence during maintenance:

```hcl
affected_monitors = [
  hyperping_monitor.api.id,
  hyperping_monitor.db.id
]
```

⚠️ **Multilingual content** - Support multiple locales:

```hcl
title = {
  en = "Database Upgrade"
  es = "Actualización de base de datos"
}
```

**Full Example After Import:**

```hcl
resource "hyperping_maintenance" "monthly" {
  statuspage_id = hyperping_statuspage.main.id

  title = {
    en = "Monthly Infrastructure Maintenance"
  }

  description = {
    en = "Routine system updates and security patches will be applied."
  }

  scheduled_start = "2026-03-15T02:00:00Z"
  scheduled_end   = "2026-03-15T04:00:00Z"

  affected_monitors = [
    hyperping_monitor.api.id,
    hyperping_monitor.web.id,
    hyperping_monitor.db.id
  ]
}
```

### Importing an Outage

**What it is:** Historical record of monitor downtime periods.

**Find the UUID:**
- Navigate to monitor's outage history
- UUID format: `outage_*` (e.g., `outage_down123`)

**Minimal Configuration:**

```hcl
resource "hyperping_outage" "incident" {
  monitor_id = hyperping_monitor.api.id
  start_time = "2026-01-15T10:30:00Z"
}
```

**Import Command:**

```bash
terraform import hyperping_outage.incident outage_down123
```

**Common Configuration Drift Issues:**

⚠️ **Timestamps** - Must be in RFC3339/ISO8601 UTC format:

```hcl
start_time = "2026-01-15T10:30:00Z"
end_time   = "2026-01-15T10:45:00Z"  # Optional, if outage ended
```

⚠️ **Read-only resource** - Outages are typically created automatically by Hyperping when monitors fail. Manual creation is rare.

⚠️ **Duration** - Computed from start and end times (read-only field)

**Full Example After Import:**

```hcl
resource "hyperping_outage" "jan_incident" {
  monitor_id = hyperping_monitor.api.id
  start_time = "2026-01-15T10:30:00Z"
  end_time   = "2026-01-15T10:45:00Z"

  # Note: These are computed/read-only
  # duration = "15m0s"
  # resolved = true
}
```

## Bulk Import with Import-Generator

The `import-generator` tool automates importing multiple resources at once. It fetches all your Hyperping resources and generates both Terraform configurations and import commands.

### Installation

Build the tool from source:

```bash
cd /path/to/terraform-provider-hyperping
cd cmd/import-generator
go build -o import-generator
```

Or use it without installing:

```bash
cd cmd/import-generator
go run . [flags]
```

### Basic Usage

Set your API key and run:

```bash
export HYPERPING_API_KEY="sk_your_api_key"
./import-generator
```

This generates both import commands and HCL configuration for all resources.

### Output Formats

#### Format: `import` (Commands Only)

Generate only Terraform import commands:

```bash
./import-generator -format=import
```

Output:

```bash
terraform import hyperping_monitor.api mon_abc123
terraform import hyperping_monitor.web mon_def456
terraform import hyperping_healthcheck.worker hc_xyz789
```

#### Format: `hcl` (Configuration Only)

Generate only HCL resource configurations:

```bash
./import-generator -format=hcl -output=imported.tf
```

Output file `imported.tf`:

```hcl
resource "hyperping_monitor" "api" {
  name = "API Health Check"
  url  = "https://api.example.com/health"
  # ... full configuration
}

resource "hyperping_monitor" "web" {
  name = "Website Monitor"
  url  = "https://example.com"
  # ... full configuration
}
```

#### Format: `both` (Default)

Generate both commands and configurations:

```bash
./import-generator -format=both -output=import-all.txt
```

Output includes both import commands and HCL blocks.

#### Format: `script` (Executable Bash Script)

Generate an executable bash script:

```bash
./import-generator -format=script -output=import.sh
chmod +x import.sh
./import.sh
```

Output `import.sh`:

```bash
#!/bin/bash
set -e

echo "Importing Hyperping resources..."

terraform import hyperping_monitor.api mon_abc123
terraform import hyperping_monitor.web mon_def456
terraform import hyperping_healthcheck.worker hc_xyz789

echo "Import complete!"
```

### Mode 1: Validation Mode

Validate all resource IDs before generating output:

```bash
./import-generator -validate
```

Output:

```
Validating resources...
✓ Monitors: 42 valid ID(s)
✓ Healthchecks: 12 valid ID(s)
✓ Statuspages: 3 valid ID(s)
✓ Incidents: 1 valid ID(s)
✓ Maintenance: 2 valid ID(s)
✓ Outages: 5 valid ID(s)
All resources valid
```

Or if validation fails:

```
Validating resources...
✓ Monitors: 42 valid ID(s)
✗ Healthchecks: 1 invalid ID (hc_invalid!)
Validation failed
```

**When to use:**
- Before running bulk imports
- To verify API access and resource IDs
- To detect malformed or inaccessible resources

### Mode 2: Progress Mode

Show real-time progress indicators:

```bash
./import-generator -progress -format=both -output=imported.tf
```

Output to stderr:

```
[1/6] Fetching monitors...
  Found 42 monitor(s)
[2/6] Fetching healthchecks...
  Found 12 healthcheck(s)
[3/6] Fetching status pages...
  Found 3 status page(s)
[4/6] Fetching incidents...
  Found 1 incident(s)
[5/6] Fetching maintenance windows...
  Found 2 maintenance window(s)
[6/6] Fetching outages...
  Found 5 outage(s)
✓ Generation complete
Output written to imported.tf
```

**When to use:**
- Large imports that take time
- To monitor progress and detect hangs
- To see counts of resources being imported

### Mode 3: Script Generation Mode

Generate an executable bash script for importing:

```bash
./import-generator -format=script -output=import.sh
```

The script is automatically made executable (`chmod +x`).

Run it:

```bash
./import.sh
```

Output:

```
Importing Hyperping resources...
hyperping_monitor.api: Import prepared!
hyperping_monitor.api: Refreshing state... [id=mon_abc123]
Import successful!

hyperping_monitor.web: Import prepared!
hyperping_monitor.web: Refreshing state... [id=mon_def456]
Import successful!
...
Import complete! Imported 65 resources.
```

**When to use:**
- Automating imports in CI/CD pipelines
- Re-running imports after state loss
- Documenting import process for team members

### Mode 4: Error Recovery Mode

Continue importing even if some resources fail:

```bash
./import-generator -continue-on-error -progress
```

Output:

```
[1/6] Fetching monitors...
  Found 42 monitor(s)
[2/6] Fetching healthchecks...
  Error: 403 Forbidden
[3/6] Fetching status pages...
  Found 3 status page(s)
...
✓ Generation complete with errors
```

**When to use:**
- API permissions are limited to certain resources
- Some resources are temporarily unavailable
- Partial imports are acceptable

### Combining Modes

Use multiple flags together:

```bash
./import-generator -validate -progress -continue-on-error -format=script -output=import.sh
```

This will:
1. Validate all resources first
2. Show progress during generation
3. Continue despite errors
4. Generate executable script
5. Save to `import.sh`

### Filtering Resources

Import only specific resource types:

```bash
# Only monitors
./import-generator -resources=monitors

# Monitors and healthchecks
./import-generator -resources=monitors,healthchecks

# Everything (default)
./import-generator -resources=all
```

Available resource types:
- `monitors`
- `healthchecks`
- `statuspages`
- `incidents`
- `maintenance`
- `outages`

### Adding Name Prefixes

Add prefixes to Terraform resource names:

```bash
./import-generator -prefix=prod_ -resources=monitors
```

Output:

```hcl
resource "hyperping_monitor" "prod_api" {
  name = "API Health Check"
  # ...
}

resource "hyperping_monitor" "prod_web" {
  name = "Website Monitor"
  # ...
}
```

**Use cases:**
- Multi-environment setups (prod, staging, dev)
- Multi-tenant configurations
- Namespace collision avoidance

### Complete Example Workflow

1. **Validate resources:**

```bash
./import-generator -validate
```

2. **Generate with progress:**

```bash
./import-generator -progress -format=both -output=import-package.txt
```

3. **Extract HCL to your Terraform files:**

Copy the HCL configuration section from `import-package.txt` to your `.tf` files.

4. **Run import commands:**

Either run commands manually or generate script:

```bash
./import-generator -format=script -output=import.sh
chmod +x import.sh
./import.sh
```

5. **Verify:**

```bash
terraform plan
```

Should show: `No changes. Your infrastructure matches the configuration.`

## Troubleshooting

### Error: "Invalid Import ID"

**Cause:** Incorrect UUID format or wrong resource type prefix.

**Solution:**

1. Verify UUID format matches resource type:
   - Monitors: `mon_*`
   - Healthchecks: `hc_*`
   - Status pages: `sp_*`
   - Incidents: `inc_*`
   - Maintenance: `maint_*`
   - Outages: `outage_*`

2. Check you're importing to the correct resource type:

```bash
# ❌ Wrong - using healthcheck ID for monitor
terraform import hyperping_monitor.api hc_123

# ✅ Correct
terraform import hyperping_healthcheck.api hc_123
```

3. For composite IDs (incident updates, status page subscribers), verify format:

```bash
# ❌ Wrong - missing slash separator
terraform import hyperping_incident_update.fix inc_123upd_456

# ✅ Correct - proper format
terraform import hyperping_incident_update.fix inc_123/upd_456
```

### Error: "Resource not found"

**Cause:** Resource doesn't exist, was deleted, or API key lacks permissions.

**Solution:**

1. Verify the resource exists in Hyperping dashboard
2. Check the UUID is correct (copy directly from dashboard URL)
3. Verify API key has read permissions:

```bash
export HYPERPING_API_KEY="sk_your_key"
curl -H "Authorization: Bearer $HYPERPING_API_KEY" \
     https://api.hyperping.io/monitors/mon_abc123
```

4. Check for typos in the UUID

### Error: "Plan shows changes after import"

**Cause:** Configuration doesn't match actual resource state.

**Solution:**

1. Run `terraform show` to see imported state:

```bash
terraform show hyperping_monitor.api
```

2. Compare with your configuration file
3. Update configuration to match actual values:

```hcl
# Example: terraform show revealed these values
resource "hyperping_monitor" "api" {
  name                 = "Production API"  # Must match exactly
  url                  = "https://api.example.com/health"
  protocol             = "http"           # Was missing
  check_frequency      = 300              # Was defaulting to 60
  expected_status_code = "200"

  regions = [                             # Was using defaults
    "london",
    "virginia",
    "frankfurt"
  ]
}
```

4. Pay special attention to:
   - Default values that may differ
   - List ordering (must match exactly)
   - Optional fields that are set in actual resource

5. Run `terraform plan` again to verify

### Error: "Unauthorized" or "Forbidden"

**Cause:** API key is invalid, expired, or lacks required permissions.

**Solution:**

1. Verify API key is set correctly:

```bash
echo $HYPERPING_API_KEY
# Should show: sk_...
```

2. Test API key manually:

```bash
curl -H "Authorization: Bearer $HYPERPING_API_KEY" \
     https://api.hyperping.io/monitors
```

3. Generate a new API key if needed:
   - Log into Hyperping dashboard
   - Navigate to Settings → API Keys
   - Create new key with appropriate permissions

4. Update your environment:

```bash
export HYPERPING_API_KEY="sk_new_key"
```

### Error: "Import failed: context deadline exceeded"

**Cause:** Network timeout or API slowness.

**Solution:**

1. Check your internet connection
2. Verify Hyperping API status
3. Retry the import (may be temporary)
4. For bulk imports, use smaller batches:

```bash
# Import specific resource types one at a time
./import-generator -resources=monitors
./import-generator -resources=healthchecks
```

### Warning: "Configuration drift detected"

**Cause:** Resource was modified outside Terraform after import.

**Solution:**

1. Decide whether to:
   - **Update Terraform config** to match current state (recommended)
   - **Apply Terraform config** to revert resource to config

2. To match current state:

```bash
terraform show hyperping_monitor.api > current-state.txt
# Update your .tf file to match current-state.txt
```

3. To revert to config:

```bash
terraform apply
# This will modify the resource to match your configuration
```

### Error: "State already contains imported resource"

**Cause:** Resource was already imported previously.

**Solution:**

1. Check current state:

```bash
terraform state list | grep hyperping_monitor.api
```

2. If resource exists, either:
   - Use existing import (no action needed)
   - Remove from state to re-import:

```bash
terraform state rm hyperping_monitor.api
terraform import hyperping_monitor.api mon_abc123
```

### Error: "Import-generator: API rate limit exceeded"

**Cause:** Too many API requests in short time.

**Solution:**

1. Wait for rate limit to reset (shown in `Retry-After` header)
2. Use smaller resource filters:

```bash
./import-generator -resources=monitors  # Import one type at a time
```

3. The import-generator automatically handles rate limits with exponential backoff, but large accounts may still hit limits

## Advanced Scenarios

### Multi-Tenant Import

Organize resources by tenant or environment using prefixes:

```bash
# Production resources
./import-generator \
  -prefix=prod_ \
  -resources=monitors,healthchecks \
  -output=prod-resources.tf

# Staging resources
./import-generator \
  -prefix=staging_ \
  -resources=monitors,healthchecks \
  -output=staging-resources.tf
```

Result:

```hcl
# prod-resources.tf
resource "hyperping_monitor" "prod_api" {
  name = "Production API"
  url  = "https://api.prod.example.com"
}

# staging-resources.tf
resource "hyperping_monitor" "staging_api" {
  name = "Staging API"
  url  = "https://api.staging.example.com"
}
```

### Cross-Environment Migration

Import from one environment and recreate in another:

**Step 1: Export from production**

```bash
export HYPERPING_API_KEY="sk_prod_key"
./import-generator -format=hcl -output=prod-monitors.tf
```

**Step 2: Modify for staging**

```bash
# Update URLs and names for staging
sed -i 's/prod\.example\.com/staging.example.com/g' prod-monitors.tf
sed -i 's/Production/Staging/g' prod-monitors.tf
mv prod-monitors.tf staging-monitors.tf
```

**Step 3: Apply to staging**

```bash
export HYPERPING_API_KEY="sk_staging_key"
terraform apply
```

### Selective Import with Filtering

Import only resources matching specific patterns:

**Import only production monitors:**

```bash
# Generate all monitors
./import-generator -resources=monitors -format=hcl -output=all-monitors.tf

# Filter for production in name
grep -A 20 'name.*= ".*PROD' all-monitors.tf > prod-monitors.tf
```

**Import monitors from specific regions:**

```bash
# After generating, filter HCL for specific regions
./import-generator -resources=monitors -format=hcl -output=monitors.tf

# Extract only monitors in specific regions
# (Requires manual review or custom scripting)
```

### Import to Module Structure

Organize imports into Terraform modules:

**Directory structure:**

```
terraform/
├── modules/
│   ├── monitors/
│   │   ├── main.tf
│   │   └── variables.tf
│   └── statuspages/
│       ├── main.tf
│       └── variables.tf
└── environments/
    ├── prod/
    │   └── main.tf
    └── staging/
        └── main.tf
```

**Generate for modules:**

```bash
# Generate monitors
./import-generator -resources=monitors -format=hcl \
  -output=modules/monitors/imported.tf

# Generate status pages
./import-generator -resources=statuspages -format=hcl \
  -output=modules/statuspages/imported.tf
```

**Use in environment:**

```hcl
# environments/prod/main.tf
module "monitors" {
  source = "../../modules/monitors"
}

module "statuspages" {
  source = "../../modules/statuspages"
}
```

### Partial Import with State Manipulation

Import resources in stages:

**Stage 1: Import core infrastructure**

```bash
./import-generator -resources=monitors,healthchecks -format=script -output=stage1.sh
./stage1.sh
```

**Stage 2: Import status pages**

```bash
./import-generator -resources=statuspages -format=script -output=stage2.sh
./stage2.sh
```

**Stage 3: Import incidents and maintenance**

```bash
./import-generator -resources=incidents,maintenance -format=script -output=stage3.sh
./stage3.sh
```

### Import with Terragrunt

Using Terragrunt for environment management:

**Generate configs per environment:**

```bash
# Production
cd environments/prod
export HYPERPING_API_KEY="sk_prod_key"
../../cmd/import-generator/import-generator -prefix=prod_ -output=imported.tf

# Staging
cd ../staging
export HYPERPING_API_KEY="sk_staging_key"
../../cmd/import-generator/import-generator -prefix=staging_ -output=imported.tf
```

**Terragrunt configuration:**

```hcl
# environments/prod/terragrunt.hcl
include "root" {
  path = find_in_parent_folders()
}

inputs = {
  environment = "prod"
  api_key     = get_env("HYPERPING_API_KEY")
}
```

## Best Practices

### Pre-Import Checklist

Before importing, verify:

- [ ] **API access** - Key is valid and has necessary permissions
- [ ] **Resource inventory** - List of all resources to import
- [ ] **Backup** - Export current Terraform state (if exists)
- [ ] **Version control** - Commit current state before import
- [ ] **Testing environment** - Consider testing import on staging first
- [ ] **Team coordination** - Ensure no one is making manual changes during import

### Post-Import Verification

After importing, check:

- [ ] **State integrity** - `terraform state list` shows all resources
- [ ] **Configuration accuracy** - All imported resources have matching config
- [ ] **No drift** - `terraform plan` shows no changes
- [ ] **Dependencies** - Resource relationships preserved (e.g., monitors → status pages)
- [ ] **Sensitive data** - No secrets exposed in configuration
- [ ] **Documentation** - Update team docs with new Terraform workflow

### Team Workflow

**For teams adopting Terraform:**

1. **Announce migration** - Notify team to freeze manual changes
2. **Inventory resources** - Document all existing resources
3. **Test import** - Run on non-production first
4. **Generate configs** - Use import-generator for consistency
5. **Peer review** - Have another team member verify configs
6. **Import production** - Run during maintenance window
7. **Verify** - Confirm `terraform plan` shows no changes
8. **Enable protection** - Add branch protection, require PR reviews
9. **Document process** - Create runbook for future imports
10. **Train team** - Ensure everyone knows Terraform workflow

### State Management

**Recommendations:**

1. **Use remote state** - Store state in S3, GCS, or Terraform Cloud
2. **Enable state locking** - Prevent concurrent modifications
3. **Backup state** - Regular backups before major operations
4. **Version state** - Use versioned storage backends

Example S3 backend:

```hcl
terraform {
  backend "s3" {
    bucket         = "my-terraform-state"
    key            = "hyperping/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-lock"
  }
}
```

### Security Considerations

**Protect sensitive data:**

1. **API keys** - Store in environment variables or secrets manager
2. **Passwords** - Use variables with `sensitive = true`
3. **State files** - Encrypt at rest and in transit
4. **Access control** - Limit who can run imports and apply changes

Example sensitive variable:

```hcl
variable "statuspage_password" {
  type      = string
  sensitive = true
}

resource "hyperping_statuspage" "main" {
  name     = "Status Page"
  subdomain = "status"
  password = var.statuspage_password  # Not visible in logs
}
```

### Naming Conventions

**Establish consistent naming:**

```hcl
# Good - descriptive, scoped, consistent
resource "hyperping_monitor" "prod_api_health" { }
resource "hyperping_monitor" "prod_web_homepage" { }
resource "hyperping_monitor" "staging_api_health" { }

# Bad - unclear, inconsistent
resource "hyperping_monitor" "m1" { }
resource "hyperping_monitor" "the_website" { }
resource "hyperping_monitor" "test_monitor_2" { }
```

**Naming pattern:**
- `{environment}_{service}_{check_type}`
- Example: `prod_api_health`, `staging_database_connection`

### Change Management

**Safe workflow for changes:**

1. **Create branch** - `git checkout -b update-monitors`
2. **Make changes** - Edit `.tf` files
3. **Plan** - `terraform plan -out=tfplan`
4. **Review** - Check plan output carefully
5. **Apply** - `terraform apply tfplan`
6. **Verify** - Check Hyperping dashboard
7. **Commit** - `git commit -am "Update monitor configs"`
8. **Pull request** - Open PR for team review
9. **Merge** - After approval

### Documentation

**Document your setup:**

Create `README.md`:

```markdown
# Hyperping Infrastructure

## Setup

1. Install Terraform 1.8+
2. Set API key: `export HYPERPING_API_KEY="sk_..."`
3. Initialize: `terraform init`

## Import Resources

See [IMPORTING.md](IMPORTING.md) for details.

## Make Changes

1. Create branch
2. Edit .tf files
3. Run `terraform plan`
4. Open PR
5. After approval: `terraform apply`

## Resources

- [Provider Docs](https://registry.terraform.io/providers/develeap/hyperping)
- [Hyperping API](https://hyperping.io/docs)
```

---

## Summary

You now have everything needed to import existing Hyperping resources into Terraform:

- ✅ **Basic imports** - Single resource import workflow
- ✅ **Resource-specific guides** - Detailed instructions for all 8 resource types
- ✅ **Bulk imports** - Import-generator tool for automating large imports
- ✅ **Troubleshooting** - Solutions to common issues
- ✅ **Advanced scenarios** - Multi-tenant, cross-environment, and module patterns
- ✅ **Best practices** - Team workflow, security, and change management

**Next steps:**

1. Start with a single resource using the [Quick Start](#quick-start-3-minutes)
2. Verify it works before scaling up
3. Use import-generator for bulk operations
4. Establish team workflow and documentation

**Need help?**

- Check [Troubleshooting](#troubleshooting) section
- Review [resource-specific guides](#resource-specific-import-guides)
- Consult [Hyperping Provider Documentation](https://registry.terraform.io/providers/develeap/hyperping)

Happy importing!
