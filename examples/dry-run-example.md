# Dry-Run Example Output

This document shows example output from the enhanced dry-run mode.

## Basic Dry-Run

```bash
$ export BETTERSTACK_API_TOKEN="your_token"
$ migrate-betterstack --dry-run
```

### Sample Output

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🔍 DRY RUN MODE - Migration Preview
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 SUMMARY
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Source Platform:  Better Stack
Target Platform:  Hyperping
Timestamp:        2026-02-14 15:30:00

Total Monitors:   42
Total Heartbeats: 7

Expected Output:
  - hyperping_monitor:     35 resources
  - hyperping_healthcheck: 7 resources
  - Total TF size:         ~1,245 lines (~42 KB)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🎯 COMPATIBILITY SCORE: 85.7% (GOOD)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Migration Complexity: MEDIUM

Breakdown:
  ✅ Clean migrations:  36/42 (85.7%)
  ⚠️  With warnings:     6/42 (14.3%)
  ❌ Errors:            0/42 (0.0%)

By Resource Type:
  hyperping_monitor:     88.6% compatible
  hyperping_healthcheck: 71.4% compatible

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📋 SAMPLE RESOURCE COMPARISONS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Showing 3 of 42 resources (use --verbose for all):

Resource: "prod_api_health" (hyperping_monitor)
┌─────────────────────────────────────┬─────────────────────────────────────┐
│          Better Stack               │             Hyperping               │
├─────────────────────────────────────┼─────────────────────────────────────┤
│ url: https://api.prod.com/health    │ url: https://api.prod.com/health    │
│ check_frequency: 30                 │ → frequency: 30                     │
│ request_timeout: 10                 │ → timeout: 10                       │
│ monitor_type: status                │ → protocol: http                    │
│ expected_status_codes: [200, 201]   │ ⚠ expected_status_code: 200         │
│                                     │   (multi-status not supported)      │
│ regions: [us, eu, asia]             │ → regions: [virginia, frankfurt,    │
│                                     │              singapore]             │
└─────────────────────────────────────┴─────────────────────────────────────┘

Transformations:
  ✓ URL preserved
  ✓ Frequency mapped (30s)
  ✓ Timeout mapped (10s)
  ✓ Protocol converted (status → http)
  ⚠ Multiple status codes → single code (manual review needed)
  ✓ Regions mapped (3/3)

... (2 more sample comparisons)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📝 TERRAFORM PREVIEW
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Showing 3 of 42 resources:

resource "hyperping_monitor" "prod_api_health" {
  name      = "[PROD]-API-Health-Check"
  url       = "https://api.prod.com/health"
  method    = "GET"
  frequency = 30
  timeout   = 10
  regions   = ["virginia", "frankfurt", "singapore"]

  expected_status = 200

  # ⚠ Original monitor had multiple status codes [200, 201]
  # Hyperping only supports single code. Review after migration.
}

resource "hyperping_monitor" "staging_web_check" {
  name      = "[STAGING]-Web-Homepage"
  url       = "https://staging.example.com"
  frequency = 60
  regions   = ["london"]
}

resource "hyperping_healthcheck" "payment_heartbeat" {
  name          = "[PROD]-Payment-Processor"
  cron_schedule = "*/5 * * * *"
  timezone      = "UTC"
  grace_period  = 300

  # ⚠ Cron expression generated from 300s interval - verify correctness
}

... (39 more resources)

📊 Resource Breakdown:
  - hyperping_monitor: 35
  - hyperping_healthcheck: 7

  Total size: ~1,245 lines (~42 KB)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⚠️  WARNINGS & MANUAL STEPS (6)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

CRITICAL (0):
  (none)

WARNINGS (6):
  1. Multiple status codes not supported (4 monitors)
     Resource: "API Gateway Health" (monitor)
     Original: expected_status_codes [200, 201, 204]
     Migrated: expected_status_code 200
     Action: Review and update manually after migration

  2. Heartbeat cron expressions require verification (7 monitors)
     Generated cron from interval may need adjustment
     Action: Review cron_schedule in Terraform output

  3. Custom headers with authentication (2 monitors)
     Monitor: "Internal API Check"
     Header: Authorization
     Action: Hyperping blocks reserved headers for security
     Solution: Use API authentication at endpoint level

INFO (3):
  - 3 monitors use non-standard regions (mapped automatically)

Estimated manual effort: 30-45 minutes

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⏱️  PERFORMANCE ESTIMATES
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Migration Time:        ~2 minutes
API Calls:
  - Better Stack:      3 calls (fetch monitors, contacts, maintenance)
  - Hyperping:         0 calls (dry-run mode)

Terraform Operations:
  - terraform plan:    ~10 seconds (estimated)
  - terraform apply:   ~45 seconds (42 resources)

Generated Files:
  - hyperping.tf:           42 KB
  - import.sh:             3 KB
  - migration-report.json: 12 KB
  - manual-steps.md:       8 KB

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📚 NEXT STEPS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Ready to proceed? Remove --dry-run flag to execute migration:

  migrate-betterstack --output=hyperping.tf

Review documentation:
  - Migration guide: docs/guides/migrate-from-betterstack.md
  - Dry-run guide:   docs/DRY_RUN_GUIDE.md
```

## Verbose Mode

```bash
$ migrate-betterstack --dry-run --verbose
```

Shows **all** resource comparisons, not just samples (3 → 42 in this example).

## JSON Output

```bash
$ migrate-betterstack --dry-run --format > report.json
```

### Sample JSON Structure

```json
{
  "summary": {
    "total_monitors": 42,
    "total_healthchecks": 7,
    "expected_tf_resources": 49,
    "expected_tf_lines": 1245,
    "expected_tf_size_bytes": 42983,
    "resource_breakdown": {
      "hyperping_monitor": 35,
      "hyperping_healthcheck": 7
    },
    "frequency_distribution": {
      "30": 15,
      "60": 20,
      "120": 7
    },
    "region_distribution": {
      "london": 30,
      "virginia": 35,
      "singapore": 28,
      "frankfurt": 25
    }
  },
  "compatibility": {
    "overall_score": 85.7,
    "by_type": {
      "hyperping_monitor": 88.6,
      "hyperping_healthcheck": 71.4
    },
    "clean_migrations": 36,
    "warning_count": 6,
    "error_count": 0,
    "complexity": "Medium",
    "details": "36 of 42 resources convert cleanly. 6 resources require review."
  },
  "warnings": [
    {
      "severity": "warning",
      "resource_name": "api_gateway_health",
      "resource_type": "monitor",
      "category": "conversion",
      "message": "Multiple expected status codes not supported",
      "action": "Review and update manually after migration",
      "impact": "May require manual configuration"
    }
  ],
  "estimates": {
    "migration_time": 120000000000,
    "source_api_calls": 3,
    "target_api_calls": 0,
    "terraform_plan_time": 10000000000,
    "terraform_apply_time": 45000000000,
    "terraform_file_size": 42983,
    "import_script_size": 3360,
    "manual_steps_size": 8192,
    "report_size": 12000
  },
  "source_platform": "Better Stack",
  "target_platform": "Hyperping",
  "timestamp": "2026-02-14T15:30:00Z"
}
```

## Using JSON Output in Scripts

```bash
#!/bin/bash

# Run dry-run and capture JSON
migrate-betterstack --dry-run --format > report.json

# Extract compatibility score
SCORE=$(jq -r '.compatibility.overall_score' report.json)

echo "Compatibility: $SCORE%"

# Check if migration is recommended
if (( $(echo "$SCORE < 75" | bc -l) )); then
    echo "⚠️  Score below 75% - manual review required"
    exit 1
fi

# Count warnings
WARNINGS=$(jq -r '.compatibility.warning_count' report.json)
echo "Warnings: $WARNINGS"

# List critical warnings
CRITICAL=$(jq -r '.warnings[] | select(.severity == "critical") | .message' report.json)
if [ -n "$CRITICAL" ]; then
    echo "Critical issues found:"
    echo "$CRITICAL"
    exit 1
fi

echo "✅ Migration approved (score: $SCORE%)"
```

## Interpreting Results

### High Compatibility (≥90%)

- **Action**: Proceed with migration
- **Risk**: Low
- **Manual Work**: Minimal (< 30 minutes)

### Good Compatibility (75-89%)

- **Action**: Review warnings, then proceed
- **Risk**: Low-Medium
- **Manual Work**: Moderate (30-60 minutes)

### Fair Compatibility (50-74%)

- **Action**: Plan carefully, consider phased approach
- **Risk**: Medium
- **Manual Work**: Significant (1-3 hours)

### Poor Compatibility (<50%)

- **Action**: Reconsider migration approach
- **Risk**: High
- **Manual Work**: Extensive (>3 hours)

## Next Steps After Dry-Run

1. **Review output carefully**
2. **Check warnings and manual steps**
3. **If score ≥ 75%, proceed with migration**:
   ```bash
   migrate-betterstack --output=hyperping.tf
   ```
4. **If score < 75%, investigate issues**:
   ```bash
   migrate-betterstack --dry-run --verbose > detailed-review.txt
   ```
