# Dry-Run Guide

Comprehensive guide to using the enhanced dry-run mode for migration preview and validation.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [What Dry-Run Shows](#what-dry-run-shows)
- [Interpreting Results](#interpreting-results)
- [Command-Line Options](#command-line-options)
- [Output Formats](#output-formats)
- [Understanding Compatibility Scores](#understanding-compatibility-scores)
- [Common Scenarios](#common-scenarios)
- [Troubleshooting](#troubleshooting)

## Overview

Dry-run mode provides a comprehensive preview of your migration **before** executing any changes. It validates your source platform credentials, analyzes compatibility, and shows exactly what will be created in Terraform.

### Key Benefits

- **Zero Risk**: No files created, no API calls to Hyperping
- **Full Preview**: See generated Terraform code before committing
- **Compatibility Analysis**: Understand which features map cleanly
- **Manual Work Estimation**: Know upfront what requires manual configuration
- **Performance Estimates**: Time and resource projections

## Quick Start

### Basic Dry-Run

```bash
# Better Stack migration
export BETTERSTACK_API_TOKEN="your_token"
migrate-betterstack --dry-run

# UptimeRobot migration
export UPTIMEROBOT_API_KEY="your_key"
migrate-uptimerobot --dry-run

# Pingdom migration
export PINGDOM_API_TOKEN="your_token"
migrate-pingdom --dry-run
```

### Verbose Mode

Show detailed comparison for **all** resources (not just samples):

```bash
migrate-betterstack --dry-run --verbose
```

### JSON Output

For automation or further processing:

```bash
migrate-betterstack --dry-run --format > report.json
jq '.compatibility.overall_score' report.json
```

## What Dry-Run Shows

### 1. Summary Statistics

Overview of source and target resources:

```
Source Platform:  Better Stack
Total Monitors:   42
Total Heartbeats: 7

Expected Output:
  - hyperping_monitor:     35 resources
  - hyperping_healthcheck: 7 resources
  - Total TF size:         ~1,245 lines (~42 KB)
```

### 2. Compatibility Score

Overall migration compatibility with breakdown:

```
COMPATIBILITY SCORE: 85.7% (GOOD)

Migration Complexity: MEDIUM

Breakdown:
  ✅ Clean migrations:  36/42 (85.7%)
  ⚠️  With warnings:     6/42 (14.3%)
  ❌ Errors:            0/42 (0.0%)

By Resource Type:
  HTTP/HTTPS:   88.6% compatible (31/35 clean)
  Heartbeats:   71.4% compatible (5/7 clean)
```

### 3. Side-by-Side Comparisons

Shows source vs target configuration with transformations:

```
Resource: "prod_api_health" (hyperping_monitor)
┌─────────────────────────────────────┬─────────────────────────────────────┐
│          Source Platform            │             Hyperping               │
├─────────────────────────────────────┼─────────────────────────────────────┤
│ url: https://api.prod.com/health    │ url: https://api.prod.com/health    │
│ check_frequency: 30                 │ → frequency: 30                     │
│ request_timeout: 10                 │ → timeout: 10                       │
│ monitor_type: status                │ → protocol: http                    │
│ expected_status_codes: [200, 201]   │ ⚠ expected_status_code: 200         │
│                                     │   (multi-status not supported)      │
└─────────────────────────────────────┴─────────────────────────────────────┘

Transformations:
  ✓ URL preserved
  ✓ Frequency mapped (30s)
  ⚠ Multiple status codes → single code (manual review needed)
```

### 4. Terraform Preview

Sample of generated Terraform configuration:

```hcl
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
```

### 5. Warnings & Manual Steps

Categorized list of issues requiring attention:

```
⚠️  WARNINGS & MANUAL STEPS (6)

CRITICAL (0):
  (none)

WARNINGS (6):
  1. Multiple status codes not supported (4 monitors)
     Resource: "API Gateway Health" (monitor)
     Action: Review and update manually after migration
     Impact: May require manual configuration

  2. Heartbeat cron expressions require verification (7 monitors)
     ...

Estimated manual effort: 30-45 minutes
```

### 6. Performance Estimates

Time and resource projections:

```
⏱️  PERFORMANCE ESTIMATES

Migration Time:        ~2 minutes
API Calls:
  - Better Stack:      3 calls
  - Hyperping:         0 calls (dry-run mode)

Terraform Operations:
  - terraform plan:    ~10 seconds (estimated)
  - terraform apply:   ~45 seconds (42 resources)

Generated Files:
  - Terraform config:  42 KB
  - Import script:     3 KB
  - Manual steps:      8 KB
```

## Interpreting Results

### Compatibility Score Ratings

| Score | Rating | Description |
|-------|--------|-------------|
| 90-100% | EXCELLENT | Migration is straightforward, minimal manual work |
| 75-89% | GOOD | Most features supported, some manual review needed |
| 50-74% | FAIR | Significant manual work required |
| 0-49% | NEEDS REVIEW | Complex migration, consider alternatives |

### Complexity Levels

| Level | Criteria | What to Expect |
|-------|----------|----------------|
| **Simple** | <25% warnings, no errors | Quick migration, minor adjustments |
| **Medium** | 25-50% warnings, or 1-10% errors | Moderate manual work, plan 1-2 hours |
| **Complex** | >50% warnings, or >10% errors | Significant effort, consider phased approach |

### Transformation Icons

| Icon | Meaning | Description |
|------|---------|-------------|
| ✓ | Preserved | Value copied directly |
| → | Mapped | Field name or format changed |
| ~ | Rounded | Value adjusted to nearest supported |
| + | Defaulted | Value set to default (source had none) |
| ✗ | Dropped | Feature not available in target |
| ⚠ | Warning | Requires manual review |

## Command-Line Options

### Basic Flags

```bash
--dry-run              # Enable dry-run mode (no files created)
--verbose              # Show all resources, not just samples
--format               # Output as JSON instead of formatted text
--debug                # Enable detailed debug logging
```

### Migration Flags

```bash
--betterstack-token    # Better Stack API token
--hyperping-api-key    # Not required in dry-run mode
--output               # Terraform file path (shown in preview)
--import-script        # Import script path (shown in preview)
```

### Example Combinations

```bash
# Basic preview
migrate-betterstack --dry-run

# Detailed preview with all resources
migrate-betterstack --dry-run --verbose

# JSON output for CI/CD
migrate-betterstack --dry-run --format > report.json

# Debug mode to troubleshoot API issues
migrate-betterstack --dry-run --debug
```

## Output Formats

### Terminal (Default)

Human-readable formatted output with colors and structure:

```bash
migrate-betterstack --dry-run
```

Features:
- Color-coded warnings and errors
- ASCII tables for side-by-side comparisons
- Progress indicators
- Summary statistics

### JSON Format

Machine-readable for automation:

```bash
migrate-betterstack --dry-run --format > report.json
```

JSON structure:

```json
{
  "summary": {
    "total_monitors": 42,
    "total_healthchecks": 7,
    "expected_tf_resources": 49
  },
  "compatibility": {
    "overall_score": 85.7,
    "complexity": "Medium",
    "clean_migrations": 36,
    "warning_count": 6,
    "error_count": 0
  },
  "warnings": [...],
  "estimates": {...}
}
```

### Processing JSON Output

```bash
# Extract compatibility score
jq '.compatibility.overall_score' report.json

# Count warnings
jq '.warnings | length' report.json

# List resources with errors
jq '.comparison[] | select(.has_errors == true)' report.json

# Get migration time estimate
jq -r '.estimates.migration_time' report.json
```

## Understanding Compatibility Scores

### Factors Affecting Score

1. **Feature Support**: Does Hyperping support this feature?
2. **Value Mapping**: Can values be converted automatically?
3. **Protocol Support**: HTTP, ICMP, Port checks vs platform-specific
4. **Region Mapping**: Geographic availability
5. **Frequency Support**: Supported check intervals

### By Monitor Type

Different monitor types have different compatibility:

| Type | Typical Score | Notes |
|------|---------------|-------|
| HTTP/HTTPS | 85-95% | Generally excellent compatibility |
| Heartbeat/Healthcheck | 70-85% | Cron expression conversion needed |
| TCP/Port | 80-90% | Direct mapping available |
| Ping/ICMP | 85-95% | Good support |
| Keyword Monitoring | 50-70% | Limited content matching in Hyperping |

### Common Warnings

#### Multiple Status Codes

**Issue**: Hyperping supports single expected status code

```
Original: expected_status_codes: [200, 201, 204]
Migrated: expected_status_code: 200
```

**Action**: Review and decide which single status code to use

#### Heartbeat Intervals

**Issue**: Converted to cron expressions

```
Original: period: 300 (5 minutes)
Migrated: cron_schedule: "*/5 * * * *"
```

**Action**: Verify cron expression matches intent

#### Region Mapping

**Issue**: Region names differ between platforms

```
Original: regions: ["us", "eu", "asia"]
Migrated: regions: ["virginia", "frankfurt", "singapore"]
```

**Action**: Confirm geographic coverage is acceptable

#### Custom Headers

**Issue**: Some headers may be restricted

```
Warning: Authorization header blocked for security
Action: Use endpoint-level authentication instead
```

## Common Scenarios

### Scenario 1: High Score, Low Complexity

```
Score: 95% (EXCELLENT)
Complexity: Simple
Warnings: 2
```

**Interpretation**: Smooth migration ahead
**Next Steps**: Review warnings, then proceed with full migration
**Time Estimate**: 30-45 minutes total

### Scenario 2: Good Score, Medium Complexity

```
Score: 78% (GOOD)
Complexity: Medium
Warnings: 12
```

**Interpretation**: Manageable migration with some manual work
**Next Steps**:
1. Review all warnings carefully
2. Plan 1-2 hours for manual adjustments
3. Consider migrating in phases

### Scenario 3: Fair Score, Complex

```
Score: 62% (FAIR)
Complexity: Complex
Warnings: 28
Critical: 3
```

**Interpretation**: Significant manual effort required
**Next Steps**:
1. Review critical issues first
2. Consider whether all monitors need migration
3. Plan phased approach: migrate cleanest 50% first
4. Allocate 4-6 hours for full migration

### Scenario 4: Low Score

```
Score: 45% (NEEDS REVIEW)
Complexity: Complex
Errors: 8
```

**Interpretation**: Platform mismatch or specialized features
**Next Steps**:
1. Review error details
2. Check if Hyperping is right fit for use case
3. Consider keeping some monitors on original platform
4. Contact support for guidance

## Troubleshooting

### Dry-Run Fails with "Invalid API Token"

**Symptom**: Error before any output
**Cause**: Source platform credentials invalid
**Solution**:

```bash
# Verify token is set
echo $BETTERSTACK_API_TOKEN

# Test token directly
curl -H "Authorization: Bearer $BETTERSTACK_API_TOKEN" \
  https://betteruptime.com/api/v2/monitors

# Regenerate token if needed
```

### No Resources Found

**Symptom**: "Total Monitors: 0"
**Cause**: Token lacks permissions or wrong account
**Solution**:
1. Verify token has read permissions
2. Check you're using correct account
3. Test with `--debug` for detailed API logs

### Compatibility Score Lower Than Expected

**Symptom**: Many warnings for simple monitors
**Cause**: Platform-specific features in use
**Solution**:
1. Run with `--verbose` to see all warnings
2. Review warnings category by category
3. Check if features are essential
4. Consider simplifying monitors pre-migration

### JSON Output Malformed

**Symptom**: `jq` fails to parse
**Cause**: Debug output mixed with JSON
**Solution**:

```bash
# Ensure only JSON is output
migrate-betterstack --dry-run --format 2>/dev/null > report.json

# Or redirect stderr
migrate-betterstack --dry-run --format > report.json 2>errors.log
```

### Performance Estimates Seem Off

**Symptom**: Time estimates don't match reality
**Cause**: Estimates are based on averages
**Solution**:
- Actual time varies with API latency
- Complex monitors take longer
- Use estimates for planning, not SLAs
- Add 50% buffer for manual work

## Best Practices

### Before Running Dry-Run

1. **Verify Credentials**: Test source API access
2. **Review Documentation**: Understand platform differences
3. **Plan Downtime**: If coordinating with migration
4. **Backup Data**: Export existing configurations

### Analyzing Results

1. **Start with Score**: Overall indicator of effort
2. **Review Critical Issues**: Address blockers first
3. **Categorize Warnings**: Group by type for batch fixes
4. **Check Estimates**: Plan time accordingly
5. **Compare Samples**: Verify transformations make sense

### Acting on Results

1. **Document Decisions**: Note why you chose specific values
2. **Test Iteratively**: Migrate subset first
3. **Monitor Closely**: Watch for gaps post-migration
4. **Keep Records**: Save dry-run reports for audit

### Automation

```bash
#!/bin/bash
# CI/CD dry-run check

SCORE=$(migrate-betterstack --dry-run --format | jq -r '.compatibility.overall_score')

if (( $(echo "$SCORE < 75" | bc -l) )); then
  echo "Compatibility score too low: $SCORE%"
  exit 1
fi

echo "Migration compatible: $SCORE%"
```

## Next Steps

After reviewing dry-run results:

1. **If Score > 90%**: Proceed with migration
   ```bash
   migrate-betterstack --output=hyperping.tf
   ```

2. **If Score 75-90%**: Review warnings, then migrate
   ```bash
   migrate-betterstack --dry-run --verbose > review.txt
   # Review manually
   migrate-betterstack --output=hyperping.tf
   ```

3. **If Score < 75%**: Plan carefully
   - Address critical issues
   - Consider phased approach
   - Contact support if needed

## Additional Resources

- [Automated Migration Guide](guides/automated-migration.md)
- [Better Stack Migration Guide](guides/migrate-from-betterstack.md)
- [UptimeRobot Migration Guide](guides/migrate-from-uptimerobot.md)
- [Pingdom Migration Guide](guides/migrate-from-pingdom.md)
- [Manual Steps Guide](guides/post-migration-manual-steps.md)

## Support

Questions or issues with dry-run mode?

- GitHub Issues: https://github.com/develeap/terraform-provider-hyperping/issues
- Documentation: https://registry.terraform.io/providers/develeap/hyperping/latest/docs
- Discord Community: [Link]

---

**Tip**: Always run `--dry-run` before actual migration. It's free, fast, and prevents surprises!
