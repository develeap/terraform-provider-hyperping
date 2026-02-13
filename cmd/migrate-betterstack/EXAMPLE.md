# Migration Tool Example

This document provides a complete, real-world example of migrating from Better Stack to Hyperping.

## Scenario

**Company**: Acme Corp
**Environment**: Production
**Resources to migrate**:
- 8 HTTP monitors
- 2 TCP port monitors
- 1 ICMP ping monitor
- 3 heartbeats (cron jobs)

## Step 1: Prepare Environment

```bash
# Export API credentials
export BETTERSTACK_API_TOKEN="bt_1234567890abcdef"
export HYPERPING_API_KEY="sk_hp_abcdef1234567890"

# Create migration directory
mkdir acme-hyperping-migration
cd acme-hyperping-migration
```

## Step 2: Run Migration Tool

```bash
# Dry run first to validate
go run github.com/develeap/terraform-provider-hyperping/cmd/migrate-betterstack \
  --dry-run \
  --verbose

# Output:
# Starting Better Stack to Hyperping migration...
# Fetching Better Stack monitors...
# Found 11 monitors
# Fetching Better Stack heartbeats...
# Found 3 heartbeats
# Converting monitors to Hyperping format...
#
# === DRY RUN MODE ===
# Would create migrated-resources.tf (3456 bytes)
# Would create import.sh (2145 bytes)
# Would create migration-report.json (1234 bytes)
# Would create manual-steps.md (890 bytes)
#
# Summary:
# Migration Summary:
#   Monitors:     11 migrated from 11
#   Healthchecks: 3 migrated from 3 heartbeats
#   Total Issues: 2 (0 critical, 2 warnings)
#
# ⚠️  Warnings found. Review migration-report.json for details
```

## Step 3: Actual Migration

```bash
# Run actual migration
go run github.com/develeap/terraform-provider-hyperping/cmd/migrate-betterstack \
  --output=acme-production.tf \
  --import-script=import-acme.sh \
  --report=acme-migration-report.json \
  --verbose

# Output:
# Starting Better Stack to Hyperping migration...
# Fetching Better Stack monitors...
# Found 11 monitors
# Fetching Better Stack heartbeats...
# Found 3 heartbeats
# Converting monitors to Hyperping format...
# Generated acme-production.tf
# Generated import-acme.sh
# Generated acme-migration-report.json
# Generated manual-steps.md
#
# === Migration Complete ===
# Migration Summary:
#   Monitors:     11 migrated from 11
#   Healthchecks: 3 migrated from 3 heartbeats
#   Total Issues: 2 (0 critical, 2 warnings)
#
# ⚠️  Warnings found. Review acme-migration-report.json for details
#
# Generated files:
#   - acme-production.tf (Terraform configuration)
#   - import-acme.sh (import script)
#   - acme-migration-report.json (migration report)
#   - manual-steps.md (manual steps)
#
# Next steps:
#   1. Review acme-production.tf and adjust as needed
#   2. Review manual-steps.md for any manual actions
#   3. Run: terraform init
#   4. Run: terraform plan
#   5. Run: terraform apply
```

## Step 4: Review Generated Files

### acme-production.tf (excerpt)

```hcl
# Auto-generated from Better Stack migration
# Generated at: 2026-02-13T00:00:00Z
# Review and customize before applying

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

# ===== MONITORS =====

resource "hyperping_monitor" "prod_api_gateway" {
  name                 = "[PROD]-API-Gateway"
  url                  = "https://api.acme.com/health"
  check_frequency      = 30

  regions = [
    "london",
    "virginia",
    "singapore",
  ]

  request_headers = [
    {
      name  = "Authorization"
      value = "Bearer prod_token_123"
    },
  ]
}

resource "hyperping_monitor" "prod_web_app" {
  name                 = "[PROD]-Web-Application"
  url                  = "https://www.acme.com"
  check_frequency      = 60

  regions = [
    "london",
    "virginia",
    "frankfurt",
  ]
}

resource "hyperping_monitor" "prod_database" {
  name                 = "[PROD]-PostgreSQL-Database"
  url                  = "db.acme.com"
  protocol             = "port"
  check_frequency      = 60
  port                 = 5432

  regions = [
    "virginia",
  ]
}

# MIGRATION NOTES:
# - Keyword monitoring not fully supported in Hyperping. Using HTTP protocol with expected status code validation. Review required_keyword field manually.
resource "hyperping_monitor" "prod_search_page" {
  name                 = "[PROD]-Search-Page"
  url                  = "https://www.acme.com/search"
  check_frequency      = 60

  regions = [
    "london",
    "virginia",
  ]
}

# ===== HEALTHCHECKS =====

resource "hyperping_healthcheck" "daily_database_backup" {
  name               = "Daily Database Backup"
  cron               = "0 2 * * *"
  timezone           = "UTC"
  grace_period_value = 30
  grace_period_type  = "minutes"
}

resource "hyperping_healthcheck" "hourly_data_sync" {
  name               = "Hourly Data Sync"
  cron               = "0 * * * *"
  timezone           = "UTC"
  grace_period_value = 15
  grace_period_type  = "minutes"
}
```

### migration-report.json (excerpt)

```json
{
  "summary": {
    "total_monitors": 11,
    "converted_monitors": 11,
    "total_heartbeats": 3,
    "converted_healthchecks": 3,
    "total_issues": 2,
    "critical_issues": 0,
    "warnings": 2
  },
  "monitors": [
    {
      "betterstack_id": "123456",
      "betterstack_name": "[PROD]-API-Gateway",
      "hyperping_name": "[PROD]-API-Gateway",
      "resource_name": "prod_api_gateway",
      "protocol": "http",
      "issues": []
    },
    {
      "betterstack_id": "123457",
      "betterstack_name": "[PROD]-Search-Page",
      "hyperping_name": "[PROD]-Search-Page",
      "resource_name": "prod_search_page",
      "protocol": "http",
      "issues": [
        "Keyword monitoring not fully supported in Hyperping. Using HTTP protocol with expected status code validation. Review required_keyword field manually."
      ]
    }
  ],
  "healthchecks": [
    {
      "betterstack_id": "789012",
      "betterstack_name": "Daily Database Backup",
      "hyperping_name": "Daily Database Backup",
      "resource_name": "daily_database_backup",
      "period": 86400,
      "issues": []
    }
  ],
  "conversion_issues": [
    {
      "resource_name": "prod_search_page",
      "resource_type": "monitor",
      "severity": "warning",
      "message": "Keyword monitoring not fully supported in Hyperping. Using HTTP protocol with expected status code validation. Review required_keyword field manually."
    },
    {
      "resource_name": "prod_api_gateway",
      "resource_type": "monitor",
      "severity": "warning",
      "message": "Check frequency 45s rounded to nearest supported value 60s"
    }
  ]
}
```

### manual-steps.md (excerpt)

```markdown
# Manual Migration Steps

This document outlines manual steps required to complete the migration from Better Stack to Hyperping.

## Before You Begin

1. **Review the generated Terraform configuration** (`acme-production.tf`)
2. **Update the import script** with actual resource UUIDs after creating resources in Hyperping
3. **Set environment variables**:
   ```bash
   export HYPERPING_API_KEY="sk_your_api_key"
   ```

## ⚠️ Warnings (Review Recommended)

### prod_search_page (monitor)
- Keyword monitoring not fully supported in Hyperping. Using HTTP protocol with expected status code validation. Review required_keyword field manually.

### prod_api_gateway (monitor)
- Check frequency 45s rounded to nearest supported value 60s

## Update Import Script

The import script contains placeholder UUIDs. You need to:

1. **Create resources in Hyperping** by running `terraform apply`
2. **Note down the UUIDs** from the Terraform output
3. **Update the import script** replacing `PLACEHOLDER_UUID` with actual UUIDs

Alternatively, skip the import step if you're creating new resources (not importing existing ones).

## Configure Notifications

Better Stack notification channels are not automatically migrated. You need to:

1. **Email notifications**: Configure in Hyperping dashboard
2. **Slack integration**: Set up webhooks in Hyperping
3. **PagerDuty**: Configure webhook integration
4. **SMS alerts**: Not supported in Hyperping (use email/Slack instead)
```

## Step 5: Customize Configuration

```bash
# Review and edit the Terraform config
vim acme-production.tf

# Changes made:
# 1. Updated API Gateway check frequency from 60s to 30s for faster detection
# 2. Added custom region configuration for critical services
# 3. Updated database backup grace period from 30 to 60 minutes
```

## Step 6: Initialize Terraform

```bash
terraform init

# Output:
# Initializing the backend...
# Initializing provider plugins...
# - Finding develeap/hyperping versions matching "~> 1.0"...
# - Installing develeap/hyperping v1.0.7...
# - Installed develeap/hyperping v1.0.7
#
# Terraform has been successfully initialized!
```

## Step 7: Plan Changes

```bash
terraform plan -out=acme.tfplan

# Output:
# Terraform will perform the following actions:
#
#   # hyperping_monitor.prod_api_gateway will be created
#   + resource "hyperping_monitor" "prod_api_gateway" {
#       + id                   = (known after apply)
#       + name                 = "[PROD]-API-Gateway"
#       + url                  = "https://api.acme.com/health"
#       + check_frequency      = 30
#       + regions              = [
#           + "london",
#           + "virginia",
#           + "singapore",
#         ]
#       ...
#     }
#
#   # ... (10 more monitors)
#   # ... (3 healthchecks)
#
# Plan: 14 to add, 0 to change, 0 to destroy.
```

## Step 8: Apply Configuration

```bash
terraform apply acme.tfplan

# Output:
# hyperping_monitor.prod_api_gateway: Creating...
# hyperping_monitor.prod_web_app: Creating...
# hyperping_monitor.prod_database: Creating...
# ...
# hyperping_healthcheck.daily_database_backup: Creating...
# ...
# hyperping_monitor.prod_api_gateway: Creation complete after 2s [id=mon_abc123...]
# ...
#
# Apply complete! Resources: 14 added, 0 changed, 0 destroyed.
```

## Step 9: Configure Notifications (Manual)

```bash
# Log in to Hyperping dashboard
open https://app.hyperping.io

# Navigate to Settings → Notifications
# Add Slack webhook: https://hooks.slack.com/services/T00/B00/xxx

# Add email addresses:
# - ops@acme.com
# - oncall@acme.com
# - alerts@acme.com
```

## Step 10: Parallel Testing

```bash
# Both Better Stack and Hyperping are now active
# Monitor for 1 week to compare

# Create comparison script
cat > compare-uptime.sh << 'EOF'
#!/bin/bash
echo "Better Stack Uptime:"
curl -s -H "Authorization: Bearer $BETTERSTACK_API_TOKEN" \
  "https://betteruptime.com/api/v2/monitors" | \
  jq -r '.data[] | "\(.attributes.pronounceable_name): \(.attributes.availability)%"'

echo ""
echo "Hyperping Status:"
echo "Check: https://app.hyperping.io/monitors"
EOF

chmod +x compare-uptime.sh
./compare-uptime.sh
```

## Step 11: Validation

After 1 week of parallel operation:

```bash
# Verify all monitors are working
terraform plan
# Expected: "No changes. Your infrastructure matches the configuration."

# Check for any drift
# Expected: No drift detected

# Verify alert delivery
# Trigger test alert in Hyperping dashboard
# Confirm receipt in Slack and email
```

## Step 12: Cutover

```bash
# Pause all Better Stack monitors
curl -X PATCH \
  -H "Authorization: Bearer $BETTERSTACK_API_TOKEN" \
  -H "Content-Type: application/json" \
  "https://betteruptime.com/api/v2/monitors/bulk" \
  -d '{"paused": true}'

# Monitor Hyperping-only for 48 hours
# Confirm no issues

# Export Better Stack data for records
curl -H "Authorization: Bearer $BETTERSTACK_API_TOKEN" \
  "https://betteruptime.com/api/v2/monitors" > betterstack-export-final.json

# Cancel Better Stack subscription
# (via Better Stack dashboard)
```

## Results

**Migration Completed**: ✓
**Downtime**: 0 minutes
**Resources Migrated**:
- 11 monitors → 11 Hyperping monitors
- 3 heartbeats → 3 Hyperping healthchecks

**Issues Encountered**:
- 1 keyword monitor required manual review (converted to HTTP)
- 1 frequency rounded from 45s to 60s (acceptable)

**Cost Savings**:
- Better Stack: $79/month (5 seats)
- Hyperping: $49/month (unlimited seats)
- **Savings**: $30/month ($360/year)

**Team Feedback**:
- ✓ Easier Terraform management
- ✓ Faster check frequencies available
- ✓ Simpler API
- ✓ Better status page customization
- ⚠️ Missing on-call scheduling (using PagerDuty instead)

## Lessons Learned

1. **Start with dry-run**: Always validate before creating files
2. **Review conversion issues**: Pay attention to warnings in report
3. **Test notifications**: Verify alert delivery before cutover
4. **Parallel operation**: Run both systems for at least 1 week
5. **Document changes**: Keep migration report for future reference

## Next Steps

1. ✓ Migration complete
2. ✓ Better Stack decommissioned
3. Update runbooks with Hyperping URLs
4. Train team on Hyperping dashboard
5. Set up additional status pages
6. Consider migrating staging environment
