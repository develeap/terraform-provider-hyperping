# Migration Guide: Manual to Terraform

This guide helps teams migrate from manually-managed Hyperping resources to Infrastructure as Code (IaC) using Terraform.

## Overview

Migration involves three phases:

1. **Discovery** - Identify existing resources in Hyperping
2. **Import** - Import resources into Terraform state
3. **Codify** - Write Terraform configurations matching imported state

## Prerequisites

Before starting:

```bash
# 1. Install Terraform 1.8+
terraform version

# 2. Set your API key
export HYPERPING_API_KEY="sk_your_api_key"

# 3. Initialize provider
cat > provider.tf << 'EOF'
terraform {
  required_providers {
    hyperping = {
      source = "develeap/hyperping"
    }
  }
}

provider "hyperping" {}
EOF

terraform init
```

## Phase 1: Discovery

### List All Resources

Use the Hyperping dashboard or API to inventory existing resources:

```bash
# Using curl to list monitors
curl -s -H "Authorization: Bearer $HYPERPING_API_KEY" \
  https://api.hyperping.io/v1/monitors | jq '.[] | {id: .uuid, name: .name}'

# Using curl to list status pages
curl -s -H "Authorization: Bearer $HYPERPING_API_KEY" \
  https://api.hyperping.io/v3/statuspages | jq '.[] | {id: .uuid, name: .name}'
```

Or use Terraform data sources:

```hcl
# discovery.tf
data "hyperping_monitors" "all" {}

output "existing_monitors" {
  value = data.hyperping_monitors.all.monitors[*].id
}
```

```bash
terraform apply -target=data.hyperping_monitors.all
terraform output existing_monitors
```

### Create Resource Inventory

Document your resources in a spreadsheet or file:

| Resource Type | Name | ID | Priority |
|---------------|------|-----|----------|
| Monitor | Production API | mon_abc123 | High |
| Monitor | Database Health | mon_def456 | High |
| Status Page | Public Status | sp_xyz789 | High |
| Healthcheck | Backup Job | hc_123abc | Medium |

## Phase 2: Import

### Import Command Syntax

```bash
terraform import <resource_type>.<name> "<id>"
```

### Monitor Import

```bash
# Single monitor
terraform import hyperping_monitor.api "mon_abc123"

# Multiple monitors (create empty resources first)
for id in mon_abc123 mon_def456 mon_ghi789; do
  name=$(echo $id | sed 's/mon_//')
  terraform import "hyperping_monitor.imported_$name" "$id"
done
```

**Terraform configuration (create before import):**

```hcl
# monitors.tf - Create placeholder resources
resource "hyperping_monitor" "api" {
  name     = "placeholder"  # Will be updated after import
  url      = "https://placeholder.com"
  protocol = "http"
}
```

### Status Page Import

```bash
terraform import hyperping_statuspage.main "sp_abc123"
```

**Configuration:**

```hcl
resource "hyperping_statuspage" "main" {
  name             = "placeholder"
  hosted_subdomain = "placeholder"
  settings = {
    name      = "placeholder"
    languages = ["en"]
  }
}
```

### Incident Import

```bash
terraform import hyperping_incident.outage "inc_abc123"
```

### Maintenance Window Import

```bash
terraform import hyperping_maintenance.scheduled "mw_abc123"
```

### Healthcheck Import

```bash
terraform import hyperping_healthcheck.backup_job "hc_abc123"
```

### Outage Import

```bash
terraform import hyperping_outage.incident "out_abc123"
```

### Status Page Subscriber Import

Subscribers use a composite ID format: `statuspage_uuid:subscriber_id`

```bash
terraform import hyperping_statuspage_subscriber.team "sp_abc123:456"
```

### Incident Update Import

Updates use a composite ID format: `incident_id/update_id`

```bash
terraform import hyperping_incident_update.resolution "inc_abc123/upd_xyz789"
```

## Phase 3: Codify

After import, sync your configuration with actual state:

### Step 1: Generate Configuration from State

```bash
# Show imported resource
terraform state show hyperping_monitor.api
```

Output:

```
# hyperping_monitor.api:
resource "hyperping_monitor" "api" {
    check_frequency = 60
    id              = "mon_abc123"
    name            = "Production API"
    protocol        = "http"
    regions         = ["virginia", "london", "frankfurt"]
    url             = "https://api.example.com/health"
}
```

### Step 2: Update Configuration

Replace placeholder with actual values:

```hcl
resource "hyperping_monitor" "api" {
  name            = "Production API"
  url             = "https://api.example.com/health"
  protocol        = "http"
  check_frequency = 60
  regions         = ["virginia", "london", "frankfurt"]
}
```

### Step 3: Verify No Changes

```bash
terraform plan
```

Expected output:

```
No changes. Your infrastructure matches the configuration.
```

If changes are detected, update your configuration to match the actual state.

## Migration Patterns

### Pattern 1: Incremental Migration

Migrate one resource type at a time:

```bash
# Week 1: Monitors
terraform import hyperping_monitor.api "mon_abc123"
terraform import hyperping_monitor.database "mon_def456"
terraform plan  # Verify

# Week 2: Status Pages
terraform import hyperping_statuspage.main "sp_xyz789"
terraform plan  # Verify

# Week 3: Healthchecks
terraform import hyperping_healthcheck.backup "hc_123abc"
terraform plan  # Verify
```

### Pattern 2: Environment-by-Environment

Migrate non-production first:

```
environments/
├── dev/
│   └── main.tf      # Migrate first
├── staging/
│   └── main.tf      # Migrate second
└── production/
    └── main.tf      # Migrate last
```

### Pattern 3: Big Bang (Small Organizations)

Import everything at once:

```bash
#!/bin/bash
# import-all.sh

set -e

echo "Importing monitors..."
terraform import hyperping_monitor.api "mon_abc123"
terraform import hyperping_monitor.database "mon_def456"

echo "Importing status pages..."
terraform import hyperping_statuspage.main "sp_xyz789"

echo "Importing healthchecks..."
terraform import hyperping_healthcheck.backup "hc_123abc"

echo "Verifying..."
terraform plan

echo "Migration complete!"
```

## Bulk Import Script

For large deployments, use this script to automate discovery and import:

```bash
#!/bin/bash
# bulk-import.sh

API_KEY="${HYPERPING_API_KEY}"
BASE_URL="https://api.hyperping.io"

# Fetch all monitors
echo "Fetching monitors..."
monitors=$(curl -s -H "Authorization: Bearer $API_KEY" "$BASE_URL/v1/monitors")

# Generate import commands
echo "$monitors" | jq -r '.[] | "terraform import hyperping_monitor.\(.name | gsub("[^a-zA-Z0-9_]"; "_") | ascii_downcase) \"\(.uuid)\""'

# Fetch all status pages
echo "Fetching status pages..."
statuspages=$(curl -s -H "Authorization: Bearer $API_KEY" "$BASE_URL/v3/statuspages")

echo "$statuspages" | jq -r '.[] | "terraform import hyperping_statuspage.\(.name | gsub("[^a-zA-Z0-9_]"; "_") | ascii_downcase) \"\(.uuid)\""'
```

## Post-Migration Verification

### 1. Plan Should Show No Changes

```bash
terraform plan
```

### 2. Validate State Integrity

```bash
terraform validate
terraform state list
```

### 3. Test Refresh

```bash
terraform refresh
terraform plan  # Still no changes
```

### 4. Test Minor Update

Make a small change to verify write access:

```hcl
resource "hyperping_monitor" "api" {
  name = "Production API (Terraform Managed)"  # Add suffix
  # ... rest unchanged
}
```

```bash
terraform apply
```

## Rollback Strategy

If migration fails:

### Option 1: Remove from State (Keep Resource)

```bash
# Remove resource from Terraform without destroying
terraform state rm hyperping_monitor.api
```

### Option 2: Restore from Backup

```bash
# Before migration, backup state
cp terraform.tfstate terraform.tfstate.backup

# If migration fails, restore
cp terraform.tfstate.backup terraform.tfstate
```

### Option 3: Manual Recreation

If state becomes corrupted:

1. Delete resource from state: `terraform state rm <resource>`
2. Import again: `terraform import <resource> "<id>"`
3. Verify: `terraform plan`

## Common Pitfalls

### 1. ID Format Confusion

Different resources use different ID formats:

| Resource | ID Format | Example |
|----------|-----------|---------|
| Monitor | `mon_xxx` | `mon_abc123` |
| Incident | `inc_xxx` | `inc_abc123` |
| Maintenance | `mw_xxx` | `mw_abc123` |
| Status Page | `sp_xxx` | `sp_abc123` |
| Healthcheck | `hc_xxx` | `hc_abc123` |
| Subscriber | `sp_xxx:123` | `sp_abc123:456` |
| Inc. Update | `inc_xxx/upd_xxx` | `inc_abc/upd_xyz` |

### 2. Sensitive Fields Not Imported

Some fields aren't imported and must be set manually:

- Status page passwords
- Webhook URLs (may need re-entry)
- Integration tokens

### 3. Computed Fields in Config

Don't include computed (read-only) fields in your configuration:

```hcl
# Wrong - includes computed fields
resource "hyperping_monitor" "api" {
  name   = "API"
  url    = "https://api.example.com"
  status = "up"      # Computed - remove!
  id     = "mon_xxx" # Computed - remove!
}

# Correct - only configurable fields
resource "hyperping_monitor" "api" {
  name     = "API"
  url      = "https://api.example.com"
  protocol = "http"
}
```

### 4. Drift During Migration

If someone modifies resources in the dashboard during migration:

```bash
# Refresh to get latest state
terraform refresh

# Check for drift
terraform plan
```

## Checklist

Before migration:
- [ ] Inventory all existing resources
- [ ] Backup current Terraform state (if any)
- [ ] Set `HYPERPING_API_KEY` environment variable
- [ ] Create placeholder resource configurations
- [ ] Test import on non-critical resource first

During migration:
- [ ] Import resources one at a time
- [ ] Run `terraform plan` after each import
- [ ] Update configuration to match imported state
- [ ] Commit configuration changes to version control

After migration:
- [ ] Verify `terraform plan` shows no changes
- [ ] Test a minor update to confirm write access
- [ ] Document any manual steps required
- [ ] Enable CI/CD for Terraform workflows

## Getting Help

- [Troubleshooting Guide](../TROUBLESHOOTING.md)
- [Validation Guide](validation.md)
- [GitHub Issues](https://github.com/develeap/terraform-provider-hyperping/issues)
