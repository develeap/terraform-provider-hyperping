# Error Recovery and Partial Failure Handling

This document describes the error recovery features available in the Hyperping migration tools.

## Overview

All three migration tools (Better Stack, UptimeRobot, Pingdom) support:

- **Checkpoint/Resume**: Automatically save progress and resume from interruptions
- **Partial Failure Handling**: Continue processing even if individual resources fail
- **Rollback**: Delete created Hyperping resources to revert a migration
- **Debug Logging**: Detailed logging for troubleshooting
- **Enhanced Dry-Run**: Validate API connectivity before migration

## Checkpoints

### Automatic Checkpointing

Checkpoints are automatically saved every 10 resources processed. Checkpoint files are stored in:

```
~/.hyperping-migrate/checkpoints/
```

Each checkpoint contains:
- Migration ID and tool name
- Timestamp and status (in_progress, completed, failed)
- Total resources and progress count
- List of processed resource IDs
- Failed resources with error details
- Created Hyperping resource UUIDs (for rollback)

### Checkpoint File Format

```json
{
  "migration_id": "betterstack-20260213-120000",
  "tool": "betterstack",
  "timestamp": "2026-02-13T12:05:30Z",
  "status": "in_progress",
  "total_resources": 100,
  "processed": 50,
  "failed": 2,
  "processed_ids": ["monitor-123", "monitor-456", ...],
  "failed_resources": [
    {
      "id": "monitor-789",
      "type": "monitor",
      "name": "API Health Check",
      "error": "unsupported protocol: ftp"
    }
  ],
  "hyperping_created": ["uuid1", "uuid2", ...]
}
```

## Resume Capability

### Resume from Last Checkpoint

To resume an interrupted migration:

```bash
migrate-betterstack --resume
```

This will:
1. Find the most recent checkpoint for the tool
2. Load the previous state
3. Skip already-processed resources
4. Continue from where it left off

### Resume from Specific Checkpoint

To resume from a specific migration:

```bash
migrate-betterstack --resume-id=betterstack-20260213-120000
```

### List Available Checkpoints

View all saved checkpoints:

```bash
migrate-betterstack --list-checkpoints
```

Output:
```
Available checkpoints:

Migration ID: betterstack-20260213-120000
  Tool: betterstack
  Status: in_progress
  Timestamp: 2026-02-13 12:00:00
  Progress: 50/100 processed (2 failed)
  Hyperping resources: 48 created

Migration ID: betterstack-20260213-090000
  Tool: betterstack
  Status: completed
  Timestamp: 2026-02-13 09:00:00
  Progress: 75/75 processed (0 failed)
  Hyperping resources: 75 created
```

## Partial Failure Handling

### Behavior

When a resource fails to convert or migrate:

1. **Error is logged** but migration continues
2. **Failed resource is tracked** in checkpoint
3. **Other resources** continue processing normally
4. **Final report** includes all failures

### Exit Codes

- `0` - All resources migrated successfully
- `1` - Partial failures (some resources failed)
- `2` - Total failure (unable to connect, invalid credentials, etc.)

### Failure Report

At the end of migration, a detailed failure report is displayed:

```
=== Failed Resources (2) ===

1. API Health Check (monitor)
   ID: monitor-789
   Error: unsupported protocol: ftp

2. Database Monitor (monitor)
   ID: monitor-890
   Error: invalid URL format
```

## Rollback

### Rollback Latest Migration

Delete all Hyperping resources created in the most recent migration:

```bash
migrate-betterstack --rollback
```

You will be prompted for confirmation:

```
This will delete 48 resources from Hyperping:
  - mon_abc123
  - mon_def456
  - mon_ghi789
  ... and 45 more

Are you sure you want to delete these resources? [y/N]:
```

### Rollback Specific Migration

Rollback a specific migration by ID:

```bash
migrate-betterstack --rollback --rollback-id=betterstack-20260213-120000
```

### Force Rollback

Skip confirmation prompt:

```bash
migrate-betterstack --rollback --force
```

**WARNING**: This will delete resources without confirmation. Use with caution.

### Rollback Process

1. Load checkpoint file
2. Parse list of created Hyperping resources
3. Confirm with user (unless `--force`)
4. Delete each resource with retry logic
5. Delete checkpoint file (if all deletions successful)

### Retry Logic

Rollback uses exponential backoff retry:
- **Max Retries**: 3
- **Initial Delay**: 1 second
- **Max Delay**: 30 seconds
- **Backoff Factor**: 2.0 (exponential)

Failed deletions are logged but don't stop the rollback process.

## Enhanced Dry-Run

### Validation Checks

The `--dry-run` flag now performs comprehensive validation:

```bash
migrate-betterstack --dry-run
```

Validates:
- API connectivity to source service (Better Stack/UptimeRobot/Pingdom)
- Authentication credentials
- Resource retrieval
- Conversion logic
- Output file generation

Does NOT:
- Create any files
- Create any Hyperping resources
- Save checkpoints

### Example Output

```
Dry run mode: validating API connectivity...
API validation successful

=== DRY RUN MODE ===
Would create migrated-resources.tf (15234 bytes)
Would create import.sh (8956 bytes)
Would create migration-report.json (12456 bytes)
Would create manual-steps.md (3421 bytes)

Summary:
  Total resources: 50
  Migrated monitors: 45
  Migrated healthchecks: 5
  Warnings: 3
  Errors: 0

=== Failed Resources (0) ===
```

## Debug Logging

### Enable Debug Mode

```bash
migrate-betterstack --debug
```

Debug mode enables:
- Verbose logging to stderr
- Detailed log file in `~/.hyperping-migrate/logs/`
- Timestamps on all log messages
- API request/response logging
- Checkpoint save/load operations

### Log File Location

```
~/.hyperping-migrate/logs/migration-20260213-120530.log
```

### Log Format

```
[2026-02-13T12:05:30.123Z] INFO: Starting Better Stack to Hyperping migration...
[2026-02-13T12:05:31.456Z] DEBUG: Fetching Better Stack monitors...
[2026-02-13T12:05:32.789Z] INFO: Found 50 monitors
[2026-02-13T12:05:33.012Z] DEBUG: Converting monitor: monitor-123
[2026-02-13T12:05:33.234Z] DEBUG: Saving checkpoint (processed: 10/50, failed: 0)
[2026-02-13T12:05:35.567Z] ERROR: Failed to convert monitor-789: unsupported protocol
```

### Debug vs Verbose

- `--verbose`: User-facing progress messages (stderr)
- `--debug`: Detailed technical logging (stderr + log file)

Both can be used together for maximum visibility.

## Recovery Scenarios

### Scenario 1: Interrupted Migration

**Problem**: Migration stopped at 50/100 resources due to network issue.

**Solution**:
```bash
# Resume from checkpoint
migrate-betterstack --resume

# Or specify exact checkpoint
migrate-betterstack --resume-id=betterstack-20260213-120000
```

### Scenario 2: Invalid Resources

**Problem**: 5 out of 100 monitors have unsupported configurations.

**Solution**: Migration continues automatically, skipping invalid resources.

**Output**:
```
=== Migration Complete ===
Processed: 95/100
Failed: 5

=== Failed Resources (5) ===
1. FTP Monitor (monitor)
   ID: monitor-123
   Error: unsupported protocol: ftp
...
```

**Action**: Review failed resources and handle manually.

### Scenario 3: Test Migration

**Problem**: Want to test migration without creating resources.

**Solution**:
```bash
# Dry run with debug logging
migrate-betterstack --dry-run --debug
```

Review generated files and logs before running real migration.

### Scenario 4: Rollback After Errors

**Problem**: Migration completed but created resources have issues.

**Solution**:
```bash
# List checkpoints
migrate-betterstack --list-checkpoints

# Rollback specific migration
migrate-betterstack --rollback --rollback-id=betterstack-20260213-120000

# Confirm deletion
```

### Scenario 5: API Rate Limiting

**Problem**: Source API returns rate limit errors.

**Solution**: Migration automatically retries with exponential backoff.

**Retry Configuration**:
- Max retries: 3
- Initial delay: 1s
- Max delay: 30s
- Backoff factor: 2.0

If still failing, migration will:
1. Log the error
2. Mark resource as failed
3. Continue with next resource

## Best Practices

### Before Migration

1. **Test with dry-run**:
   ```bash
   migrate-betterstack --dry-run --verbose
   ```

2. **Enable debug logging** for first run:
   ```bash
   migrate-betterstack --debug
   ```

3. **Backup source configuration** (export from Better Stack/UptimeRobot/Pingdom)

### During Migration

1. **Monitor progress**: Watch stderr output for errors
2. **Don't interrupt**: Let migration complete or use Ctrl+C (checkpoint will save)
3. **Check disk space**: Ensure enough space for checkpoint files

### After Migration

1. **Review failure report**: Address any failed resources
2. **Verify created resources**: Check Hyperping dashboard
3. **Save checkpoint ID**: Note migration ID for potential rollback
4. **Test monitors**: Verify migrated monitors work correctly

### Error Recovery

1. **Resume from checkpoint** if interrupted:
   ```bash
   migrate-betterstack --resume
   ```

2. **Review logs** if errors occur:
   ```bash
   cat ~/.hyperping-migrate/logs/migration-*.log
   ```

3. **Rollback if needed**:
   ```bash
   migrate-betterstack --rollback
   ```

4. **Re-run migration** after fixing issues:
   ```bash
   migrate-betterstack
   ```

## Troubleshooting

### Checkpoint Not Found

**Error**: `No checkpoint found to resume from`

**Solution**:
```bash
# List available checkpoints
migrate-betterstack --list-checkpoints

# Use specific checkpoint ID
migrate-betterstack --resume-id=CHECKPOINT_ID
```

### Rollback Failed

**Error**: `Failed to delete resource mon_abc123`

**Possible Causes**:
- Resource already deleted
- Invalid Hyperping API key
- Network connectivity issues
- Hyperping API error

**Solution**:
1. Check Hyperping API key is valid
2. Verify network connectivity
3. Manually delete remaining resources from Hyperping dashboard
4. Check debug logs for details

### API Validation Failed

**Error**: `Failed to connect to Better Stack API`

**Solution**:
1. Verify API token is correct
2. Check network connectivity
3. Verify source service is accessible
4. Enable debug mode to see detailed error

### Permission Denied

**Error**: `Failed to create checkpoint directory`

**Solution**:
```bash
# Check directory permissions
ls -la ~/.hyperping-migrate/

# Create directory manually
mkdir -p ~/.hyperping-migrate/checkpoints
chmod 700 ~/.hyperping-migrate/checkpoints
```

## File Locations

### Checkpoint Files

```
~/.hyperping-migrate/checkpoints/
├── betterstack-20260213-120000.json
├── uptimerobot-20260213-090000.json
└── pingdom-20260213-150000.json
```

### Log Files

```
~/.hyperping-migrate/logs/
├── migration-20260213-120530.log
├── migration-20260213-090245.log
└── migration-20260213-150812.log
```

### Permissions

All files are created with `0600` (user read/write only) for security.

## Advanced Usage

### Custom Checkpoint Interval

The checkpoint interval (default: 10 resources) is defined in code:

```go
const checkpointInterval = 10
```

To modify, edit the respective `checkpoint.go` file for each tool.

### Programmatic Access

Checkpoint manager can be used programmatically:

```go
import "github.com/develeap/terraform-provider-hyperping/pkg/checkpoint"

mgr, _ := checkpoint.NewManager()

// Save checkpoint
cp := &checkpoint.Checkpoint{
    MigrationID: "custom-migration",
    Tool:        "custom-tool",
    Status:      checkpoint.StatusInProgress,
    // ...
}
mgr.Save(cp)

// Load checkpoint
loaded, _ := mgr.Load("custom-migration")

// List all checkpoints
checkpoints, _ := mgr.List()

// Find latest
latest, _ := mgr.FindLatest("custom-tool")
```

## Migration Tool Comparison

| Feature | Better Stack | UptimeRobot | Pingdom |
|---------|--------------|-------------|---------|
| Checkpoint/Resume | ✓ | ✓ | ✓ |
| Partial Failure Handling | ✓ | ✓ | ✓ |
| Rollback | ✓ | ✓ | ✓ |
| Debug Logging | ✓ | ✓ | ✓ |
| Enhanced Dry-Run | ✓ | ✓ | ✓ |
| Retry Logic | ✓ | ✓ | ✓ |

All tools share the same recovery infrastructure.

## See Also

- [Migration Guide](../README.md)
- [Better Stack Migration](../cmd/migrate-betterstack/README.md)
- [UptimeRobot Migration](../cmd/migrate-uptimerobot/README.md)
- [Pingdom Migration](../cmd/migrate-pingdom/README.md)
