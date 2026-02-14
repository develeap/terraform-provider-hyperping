# Import Generator - Complete Guide

**Version:** 2.0
**Last Updated:** 2026-02-14

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Modes of Operation](#modes-of-operation)
- [Filtering](#filtering)
- [Parallel Execution](#parallel-execution)
- [Drift Detection](#drift-detection)
- [Checkpoint & Resume](#checkpoint--resume)
- [Rollback](#rollback)
- [Advanced Usage](#advanced-usage)
- [Performance Benchmarks](#performance-benchmarks)
- [Troubleshooting](#troubleshooting)
- [Examples](#examples)

---

## Overview

The Import Generator is a powerful tool that:

- **Generates** Terraform import commands and HCL configurations from existing Hyperping resources
- **Executes** bulk imports with parallel processing (5-8x faster)
- **Filters** resources by name pattern, type, or exclusion rules
- **Detects drift** before and after imports
- **Resumes** from checkpoints after interruptions
- **Rollbacks** failed imports cleanly

### Key Features

| Feature | Description | Benefit |
|---------|-------------|---------|
| Filtering | Import specific resource subsets | Avoid importing unnecessary resources |
| Parallel Execution | Concurrent terraform imports | 5-8x faster for 100+ resources |
| Drift Detection | Pre/post-import drift checks | Catch configuration conflicts early |
| Checkpointing | Auto-save progress every 10 imports | Resume after failures or interruptions |
| Rollback | Undo imports with one command | Clean recovery from errors |
| Progress Tracking | Real-time progress bars | Monitor long-running imports |

---

## Installation

### Build from Source

```bash
cd cmd/import-generator
go build -o import-generator
```

### Install Globally

```bash
go install ./cmd/import-generator
```

### Prerequisites

- **Go 1.21+** (for building)
- **Terraform 1.8+** (for execution mode)
- **Hyperping API Key** (set as environment variable)

---

## Quick Start

### 1. Set API Key

```bash
export HYPERPING_API_KEY="sk_your_api_key_here"
```

### 2. Generate Import Commands

```bash
# Generate import commands for all resources
import-generator

# Generate and save to file
import-generator -format=both -output=import.tf
```

### 3. Execute Import (New in v2.0)

```bash
# Import with parallel execution
import-generator --execute --parallel=10

# Import PROD resources only
import-generator --execute --filter-name="PROD-.*"

# Dry run (see what would be imported)
import-generator --dry-run --filter-name="PROD-.*"
```

---

## Modes of Operation

### Generation Mode (Default)

Generates import commands and HCL without executing:

```bash
# Generate import commands
import-generator -format=import

# Generate HCL configuration
import-generator -format=hcl

# Generate both
import-generator -format=both

# Generate executable shell script
import-generator -format=script -output=import.sh
```

**Output formats:**
- `import` - Terraform import commands only
- `hcl` - HCL resource configurations only
- `both` - Import commands + HCL (default)
- `script` - Executable bash script with error handling

### Execution Mode

Directly executes terraform imports:

```bash
# Execute imports sequentially
import-generator --execute

# Execute with 10 parallel workers
import-generator --execute --parallel=10

# Show what would be imported without executing
import-generator --dry-run
```

**Flags:**
- `--execute` - Enable execution mode
- `--dry-run` - Show plan without executing
- `--parallel=N` - Number of concurrent workers (default: 5, max: 20)
- `--sequential` - Disable parallelization

### Validation Mode

Validate resource IDs without generating output:

```bash
import-generator --validate
```

**Checks:**
- Resource ID format (e.g., `mon_*`, `hc_*`, `sp_*`)
- API connectivity
- Resource accessibility

### Rollback Mode

Remove previously imported resources from state:

```bash
# Show rollback plan
import-generator --rollback-plan

# Execute rollback
import-generator --rollback

# Rollback with custom log file
import-generator --rollback --rollback-file=.my-import-log
```

---

## Filtering

### Filter by Name Pattern

Import resources matching a regex pattern:

```bash
# Import all PROD resources
import-generator --execute --filter-name="^PROD-.*"

# Import staging resources
import-generator --execute --filter-name="^STAGING-.*"

# Import API-related resources
import-generator --execute --filter-name=".*-API-.*"
```

**Regex Examples:**
- `^PROD-.*` - Starts with PROD-
- `.*-API$` - Ends with -API
- `.*(prod|production).*` - Contains prod or production
- `^[A-Z]{4}-.*` - Four uppercase letters followed by dash

### Exclude Pattern

Exclude resources matching a pattern:

```bash
# Exclude test resources
import-generator --execute --filter-exclude="test-.*"

# Exclude legacy resources
import-generator --execute --filter-exclude="legacy-.*"

# Combine include and exclude
import-generator --execute \
  --filter-name="^PROD-.*" \
  --filter-exclude=".*-deprecated$"
```

### Filter by Resource Type

Import specific resource types only:

```bash
# Import monitors only
import-generator --execute --filter-type=hyperping_monitor

# Import healthchecks only
import-generator --execute --filter-type=hyperping_healthcheck

# Import status pages only
import-generator --execute --filter-type=hyperping_statuspage
```

**Available resource types:**
- `hyperping_monitor`
- `hyperping_healthcheck`
- `hyperping_statuspage`
- `hyperping_incident`
- `hyperping_maintenance`
- `hyperping_outage`

### Combining Filters

All filters use AND logic:

```bash
# PROD monitors only, excluding tests
import-generator --execute \
  --filter-type=hyperping_monitor \
  --filter-name="^PROD-.*" \
  --filter-exclude="test-.*"
```

---

## Parallel Execution

### Why Parallel?

| Resources | Sequential | Parallel (5) | Parallel (10) | Speedup |
|-----------|-----------|--------------|---------------|---------|
| 10        | 30s       | 15s          | 12s           | 2.5x    |
| 50        | 2m 30s    | 45s          | 30s           | 5x      |
| 100       | 5m        | 1m 15s       | 45s           | 6.7x    |
| 500       | 25m       | 5m 30s       | 3m            | 8.3x    |

### Usage

```bash
# Default: 5 workers
import-generator --execute

# Custom worker count
import-generator --execute --parallel=10

# Maximum workers (20)
import-generator --execute --parallel=20

# Sequential (debugging)
import-generator --execute --sequential
```

### Recommendations

- **Small imports (<20 resources):** Use sequential
- **Medium imports (20-100):** Use 5-10 workers
- **Large imports (100+):** Use 10-20 workers
- **Testing/debugging:** Use sequential with `--verbose`

### Safety

Parallel imports are safe because:
- Terraform state locking prevents conflicts
- Each import is independent
- Failed imports don't affect successful ones
- Checkpoints enable recovery

---

## Drift Detection

### Pre-Import Drift Detection

Detect configuration drift before importing:

```bash
# Detect drift and prompt to continue
import-generator --execute --detect-drift

# Abort if drift detected
import-generator --execute --detect-drift --abort-on-drift

# Refresh state before drift detection
import-generator --execute --detect-drift --refresh-first
```

**What it checks:**
- Resources that will be updated
- Resources that will be created
- Resources that will be destroyed
- Resources that must be replaced

**Example output:**

```
=================================================================================
DRIFT DETECTION
=================================================================================
Running terraform plan to detect configuration drift...

⚠️  Configuration drift detected!

  UPDATE (2):
    - hyperping_monitor.staging_api (frequency changed)
    - hyperping_monitor.dev_web (paused status changed)

  CREATE (1):
    - hyperping_healthcheck.new_heartbeat (will be created)

Total drifted resources: 3
=================================================================================

⚠️  WARNING: Importing resources with existing drift may cause unexpected results.
It's recommended to resolve drift before importing new resources.

Do you want to continue anyway? (yes/no):
```

### Post-Import Drift Check

Verify imported resources match configuration:

```bash
import-generator --execute --post-import-check
```

**Recommended for:**
- Production imports
- Validating HCL accuracy
- Compliance verification

---

## Checkpoint & Resume

### How It Works

Checkpoints automatically save progress:
- Every 10 imports
- On completion
- Before exit (if enabled)

### Resume After Interruption

```bash
# Start import
import-generator --execute --parallel=10
# (press Ctrl+C after 50% complete)

# Resume from checkpoint
import-generator --execute --resume
```

**Example output:**

```
Checkpoint Information:
=======================
Created:          2026-02-14T10:30:00Z
Total resources:  100
Imported:         47
Failed:           3
Progress:         50/100 (50.0%)
Completed:        false

Resume from checkpoint? (y/N): y

Resuming import: 47/100 completed, 50 remaining
```

### Checkpoint Management

```bash
# Disable checkpointing
import-generator --execute --no-checkpoint

# Custom checkpoint file
import-generator --execute --checkpoint-file=.my-checkpoint

# Manually clean up checkpoint
rm .import-checkpoint
```

### Checkpoint File Format

```json
{
  "timestamp": "2026-02-14T10:30:00Z",
  "total_resources": 100,
  "imported_ids": [
    {
      "id": "mon_abc123",
      "resource_type": "hyperping_monitor",
      "resource_name": "prod_api"
    }
  ],
  "failed_ids": ["mon_xyz789"],
  "current_index": 50,
  "completed": false,
  "filter_summary": "Name pattern: PROD-.*"
}
```

---

## Rollback

### When to Rollback

- Import failed mid-way
- Wrong resources imported
- Need to start over
- Testing import process

### Show Rollback Plan

```bash
import-generator --rollback-plan
```

**Example output:**

```
=================================================================================
ROLLBACK PLAN
=================================================================================
Import log created: 2026-02-14T10:30:00Z
Resources that would be removed: 42

  - hyperping_monitor.prod_api (ID: mon_123, imported at: 2026-02-14T10:31:00Z)
  - hyperping_monitor.prod_web (ID: mon_456, imported at: 2026-02-14T10:31:05Z)
  - hyperping_healthcheck.prod_heartbeat (ID: hc_789, imported at: 2026-02-14T10:31:10Z)
  ...
=================================================================================
```

### Execute Rollback

```bash
# Rollback with confirmation
import-generator --rollback

# Verbose output
import-generator --rollback --verbose
```

**Example output:**

```
=================================================================================
ROLLBACK PLAN
=================================================================================
Import log created: 2026-02-14T10:30:00Z
Resources to remove: 42

  hyperping_monitor: 35 resource(s)
  hyperping_healthcheck: 7 resource(s)
=================================================================================

This will remove all listed resources from Terraform state.
Are you sure you want to proceed? (yes/no): yes

=================================================================================
EXECUTING ROLLBACK
=================================================================================
✓ Removed hyperping_monitor.prod_api
✓ Removed hyperping_monitor.prod_web
✓ Removed hyperping_healthcheck.prod_heartbeat
...

=================================================================================
ROLLBACK SUMMARY
=================================================================================
Successfully removed: 42
Failed to remove:     0
=================================================================================

Import log deleted
```

### Important Notes

- Rollback removes from **state only**, not from Hyperping
- Actual resources remain unchanged
- Re-run import after fixing issues
- Rollback is reversible (just import again)

---

## Advanced Usage

### Complex Filtering

```bash
# Import PROD monitors, exclude tests, with drift detection
import-generator --execute \
  --filter-type=hyperping_monitor \
  --filter-name="^PROD-.*" \
  --filter-exclude=".*-test$" \
  --detect-drift \
  --abort-on-drift \
  --parallel=10 \
  --verbose

# Import specific resource types with checkpointing
import-generator --execute \
  --resources=monitors,healthchecks \
  --filter-name="^(PROD|STAGING)-.*" \
  --checkpoint-file=.prod-import-checkpoint \
  --parallel=15
```

### Multi-Stage Import

```bash
# Stage 1: Import PROD monitors
import-generator --execute \
  --filter-type=hyperping_monitor \
  --filter-name="^PROD-.*" \
  --checkpoint-file=.prod-monitors-checkpoint

# Stage 2: Import PROD healthchecks
import-generator --execute \
  --filter-type=hyperping_healthcheck \
  --filter-name="^PROD-.*" \
  --checkpoint-file=.prod-healthchecks-checkpoint

# Stage 3: Import status pages
import-generator --execute \
  --filter-type=hyperping_statuspage \
  --checkpoint-file=.statuspages-checkpoint
```

### Testing Import Strategy

```bash
# 1. Dry run to see what would be imported
import-generator --dry-run --filter-name="^PROD-.*"

# 2. Validate resources
import-generator --validate

# 3. Execute with drift detection
import-generator --execute \
  --filter-name="^PROD-.*" \
  --detect-drift \
  --post-import-check \
  --verbose

# 4. Rollback if needed
import-generator --rollback
```

### JSON Output (for automation)

```bash
# Generate JSON for processing
import-generator --execute --json > import-results.json

# Parse with jq
jq '.summary.success_count' import-results.json
```

---

## Performance Benchmarks

### Test Environment

- **Provider:** terraform-provider-hyperping v1.0.7
- **Terraform:** v1.8.0
- **CPU:** 8 cores
- **Network:** 100 Mbps
- **API Latency:** ~200ms average

### Sequential vs Parallel

| Resources | Sequential | Parallel (5) | Parallel (10) | Parallel (20) |
|-----------|-----------|--------------|---------------|---------------|
| 10        | 30s       | 15s          | 12s           | 11s           |
| 25        | 1m 15s    | 25s          | 18s           | 16s           |
| 50        | 2m 30s    | 45s          | 30s           | 25s           |
| 100       | 5m        | 1m 15s       | 45s           | 35s           |
| 250       | 12m 30s   | 3m           | 1m 50s        | 1m 20s        |
| 500       | 25m       | 5m 30s       | 3m            | 2m 15s        |
| 1000      | 50m       | 11m          | 6m            | 4m 30s        |

### Speedup by Worker Count

| Workers | 50 Resources | 100 Resources | 500 Resources |
|---------|-------------|--------------|---------------|
| 1       | 1.0x (baseline) | 1.0x (baseline) | 1.0x (baseline) |
| 5       | 3.3x        | 4.0x         | 4.5x          |
| 10      | 5.0x        | 6.7x         | 8.3x          |
| 20      | 6.0x        | 8.6x         | 11.1x         |

### Optimal Worker Count

| Resource Count | Recommended Workers | Rationale |
|---------------|-------------------|-----------|
| < 20          | 1 (sequential)    | Overhead > benefit |
| 20-50         | 5                 | Good balance |
| 50-100        | 10                | Optimal speedup |
| 100-500       | 15                | Maximum efficiency |
| 500+          | 20                | Diminishing returns |

### Factors Affecting Performance

1. **API Rate Limits:** Hyperping API has rate limits (check documentation)
2. **Network Latency:** Higher latency benefits more from parallelization
3. **Terraform State Size:** Larger states slow down state operations
4. **Resource Complexity:** Complex resources take longer to import

---

## Troubleshooting

### Common Issues

#### 1. Import Fails with "Resource Not Found"

**Symptom:**
```
Error: Resource not found: mon_abc123
```

**Cause:** Resource was deleted from Hyperping but still in generated commands.

**Solution:**
```bash
# Validate resources first
import-generator --validate

# Filter out problematic resources
import-generator --execute --continue-on-error
```

#### 2. "Too Many Requests" Error

**Symptom:**
```
Error: 429 Too Many Requests
```

**Cause:** Hitting API rate limits with too many parallel workers.

**Solution:**
```bash
# Reduce worker count
import-generator --execute --parallel=3

# Or use sequential
import-generator --execute --sequential
```

#### 3. Import Interrupted

**Symptom:** Process killed or network disconnected mid-import.

**Solution:**
```bash
# Resume from checkpoint
import-generator --execute --resume
```

#### 4. Drift Detected

**Symptom:**
```
⚠️  Configuration drift detected!
```

**Solutions:**

**Option A: Update HCL to match actual state**
```bash
# Generate fresh HCL from current state
import-generator -format=hcl -output=updated.tf

# Compare with existing
diff main.tf updated.tf
```

**Option B: Continue anyway**
```bash
# Continue despite drift
import-generator --execute --detect-drift
# (answer 'yes' when prompted)
```

**Option C: Abort and fix**
```bash
# Abort on drift
import-generator --execute --detect-drift --abort-on-drift

# Fix drift
terraform apply

# Retry import
import-generator --execute
```

#### 5. State Locked

**Symptom:**
```
Error: Error acquiring the state lock
```

**Cause:** Another terraform process is running or crashed.

**Solution:**
```bash
# Check for running processes
ps aux | grep terraform

# Force unlock (use with caution)
terraform force-unlock <LOCK_ID>

# Retry import
import-generator --execute
```

#### 6. Memory Issues (Large Imports)

**Symptom:** Process killed or OOM error.

**Solution:**
```bash
# Use smaller batches
import-generator --execute \
  --resources=monitors \
  --checkpoint-file=.monitors-checkpoint

import-generator --execute \
  --resources=healthchecks \
  --checkpoint-file=.healthchecks-checkpoint

# Or reduce parallel workers
import-generator --execute --parallel=3
```

### Debugging

#### Enable Verbose Output

```bash
import-generator --execute --verbose
```

#### Enable Terraform Debug Logging

```bash
export TF_LOG=DEBUG
export TF_LOG_PATH=terraform.log
import-generator --execute
```

#### Check Import Log

```bash
# View import log
cat .import-log | jq .

# Count imported resources
jq '.resources | length' .import-log
```

#### Validate Environment

```bash
# Check Terraform version
terraform version

# Check provider version
terraform providers

# Validate configuration
terraform validate

# Check API connectivity
curl -H "Authorization: Bearer $HYPERPING_API_KEY" \
  https://api.hyperping.io/v1/monitors
```

---

## Examples

### Example 1: Basic Import

```bash
# Set API key
export HYPERPING_API_KEY="sk_abc123"

# Generate and review
import-generator -format=both > import-plan.txt
less import-plan.txt

# Execute import
import-generator --execute --progress
```

### Example 2: Production Import with Safety Checks

```bash
# Import PROD resources with all safety features
import-generator --execute \
  --filter-name="^PROD-.*" \
  --detect-drift \
  --abort-on-drift \
  --post-import-check \
  --parallel=10 \
  --verbose
```

### Example 3: Multi-Environment Import

```bash
# Create separate files for each environment
import-generator \
  --filter-name="^PROD-.*" \
  -format=hcl \
  -output=prod.tf

import-generator \
  --filter-name="^STAGING-.*" \
  -format=hcl \
  -output=staging.tf

import-generator \
  --filter-name="^DEV-.*" \
  -format=hcl \
  -output=dev.tf

# Execute imports with checkpoints
import-generator --execute \
  --filter-name="^PROD-.*" \
  --checkpoint-file=.prod-checkpoint

import-generator --execute \
  --filter-name="^STAGING-.*" \
  --checkpoint-file=.staging-checkpoint
```

### Example 4: Selective Resource Type Import

```bash
# Import only monitors and healthchecks
import-generator --execute \
  --resources=monitors,healthchecks \
  --parallel=10

# Import only status pages
import-generator --execute \
  --filter-type=hyperping_statuspage \
  --sequential
```

### Example 5: Recovery from Failed Import

```bash
# Import fails at 60%
import-generator --execute --parallel=10
# (import fails)

# Resume from checkpoint
import-generator --execute --resume
# (complete remaining 40%)

# If still failing, rollback
import-generator --rollback

# Fix issues and retry
import-generator --execute --parallel=5 --verbose
```

### Example 6: Automated CI/CD Integration

```bash
#!/bin/bash
set -e

# Validate environment
if [ -z "$HYPERPING_API_KEY" ]; then
  echo "Error: HYPERPING_API_KEY not set"
  exit 1
fi

# Validate resources
echo "Validating resources..."
import-generator --validate || exit 1

# Dry run
echo "Planning import..."
import-generator --dry-run --filter-name="^PROD-.*" || exit 1

# Execute with drift detection
echo "Executing import..."
import-generator --execute \
  --filter-name="^PROD-.*" \
  --detect-drift \
  --abort-on-drift \
  --parallel=10 \
  --json > import-results.json || {
    echo "Import failed, rolling back..."
    import-generator --rollback
    exit 1
  }

# Verify success
SUCCESS=$(jq '.summary.success_count' import-results.json)
TOTAL=$(jq '.summary.total_jobs' import-results.json)

if [ "$SUCCESS" -eq "$TOTAL" ]; then
  echo "Import successful: $SUCCESS/$TOTAL resources"
else
  echo "Import partially failed: $SUCCESS/$TOTAL resources"
  import-generator --rollback
  exit 1
fi
```

---

## FAQ

### Q: What happens if import fails mid-way?

**A:** Checkpoints save progress every 10 imports. Use `--resume` to continue:

```bash
import-generator --execute --resume
```

### Q: Can I import resources from multiple Hyperping accounts?

**A:** Yes, change the `HYPERPING_API_KEY` environment variable:

```bash
# Account 1
export HYPERPING_API_KEY="sk_account1"
import-generator --execute --checkpoint-file=.account1-checkpoint

# Account 2
export HYPERPING_API_KEY="sk_account2"
import-generator --execute --checkpoint-file=.account2-checkpoint
```

### Q: How do I import resources with special characters in names?

**A:** The generator automatically sanitizes names. Use `--prefix` to avoid conflicts:

```bash
import-generator --execute --prefix="imported_"
```

### Q: Can I import to a remote Terraform state?

**A:** Yes, configure your backend before importing:

```hcl
# backend.tf
terraform {
  backend "s3" {
    bucket = "my-terraform-state"
    key    = "hyperping/terraform.tfstate"
    region = "us-east-1"
  }
}
```

```bash
terraform init
import-generator --execute
```

### Q: What's the maximum number of parallel workers?

**A:** 20 workers (hard-coded limit). Beyond this, diminishing returns occur.

### Q: Does rollback delete resources from Hyperping?

**A:** No, rollback only removes from Terraform state. Hyperping resources remain unchanged.

### Q: Can I customize the generated HCL?

**A:** Yes, generate HCL and edit before importing:

```bash
# Generate HCL
import-generator -format=hcl -output=resources.tf

# Edit as needed
vim resources.tf

# Import manually
terraform import hyperping_monitor.custom_name mon_abc123
```

---

## Best Practices

1. **Always validate first:** `import-generator --validate`
2. **Use dry-run for testing:** `--dry-run`
3. **Enable drift detection for production:** `--detect-drift --abort-on-drift`
4. **Use checkpoints for large imports:** Default enabled
5. **Keep rollback logs:** Don't delete `.import-log` until verified
6. **Test in staging first:** Import staging resources before production
7. **Use filters to avoid bloat:** Don't import test/legacy resources
8. **Monitor API rate limits:** Reduce workers if hitting limits
9. **Verify with terraform plan:** After import, always run `terraform plan`
10. **Document your imports:** Keep logs of what was imported when

---

## Support

- **Issues:** https://github.com/develeap/terraform-provider-hyperping/issues
- **Documentation:** https://github.com/develeap/terraform-provider-hyperping/tree/main/docs
- **Examples:** https://github.com/develeap/terraform-provider-hyperping/tree/main/examples

---

## Changelog

### v2.0 (2026-02-14)

**New Features:**
- Parallel import execution (5-8x faster)
- Resource filtering (name, type, exclusion)
- Drift detection (pre/post-import)
- Checkpoint/resume capability
- Rollback functionality
- Progress tracking
- Dry-run mode
- JSON output

**Improvements:**
- Better error handling
- Enhanced progress reporting
- Comprehensive logging

**Breaking Changes:**
- None (backward compatible)

### v1.0

- Initial release
- Sequential import generation
- Basic HCL generation
- Validation mode

---

**Last Updated:** 2026-02-14
**Version:** 2.0
**License:** MPL-2.0
