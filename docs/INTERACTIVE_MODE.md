# Interactive Migration Mode

**Interactive mode** provides a guided, user-friendly wizard for migrating monitoring resources to Hyperping. It eliminates the need to remember command-line flags and provides real-time validation and progress feedback.

## Features

✅ **Automatic activation** when no flags provided
✅ **API connection testing** before migration starts
✅ **Real-time progress bars** for long operations
✅ **Input validation** with helpful error messages
✅ **Migration preview** with confirmation prompt
✅ **Detailed summary** with next steps
✅ **Full backward compatibility** with existing workflows

## Quick Start

Simply run a migration tool without any flags:

```bash
# Better Stack
cd cmd/migrate-betterstack
go run .

# UptimeRobot
cd cmd/migrate-uptimerobot
go run .

# Pingdom
cd cmd/migrate-pingdom
go run .
```

The tool will automatically enter interactive mode and guide you through the migration process.

## Interactive Workflow

### Step-by-Step Process

#### Step 1: Source Platform Configuration

You'll be prompted to enter your source platform API credentials:

```
Step 1/5: Source Platform Configuration

? Enter your Better Stack API token: ••••••••••••••••
✅ Testing Better Stack API connection... Done! Found 42 monitors
```

The tool will:
- Prompt for API credentials (input is hidden)
- Test the connection immediately
- Display the number of resources found
- Show helpful error messages if connection fails

#### Step 2: Destination Platform Configuration

For full migrations, you'll provide Hyperping credentials:

```
Step 2/5: Destination Platform Configuration

? Perform dry run only (validate without creating files)? No
? Enter your Hyperping API key: ••••••••••••••••
```

For dry runs, this step is skipped.

#### Step 3: Output Configuration

Configure where migration files will be saved:

```
Step 3/5: Output Configuration

? Terraform output file: (migrated-resources.tf)
? Import script file: (import.sh)
? Migration report file: (migration-report.json)
? Manual steps file: (manual-steps.md)
```

Default values are provided - just press Enter to accept them.

#### Step 4: Migration Preview

Review what will happen before proceeding:

```
Step 4/5: Migration Preview

  📊 Summary:
    - Total monitors: 42
    - Total heartbeats: 7
    - Total resources: 49
    - Mode: Full migration

  📁 Output files:
    - migrated-resources.tf (Terraform configuration)
    - import.sh (Import script)
    - migration-report.json (Migration report)
    - manual-steps.md (Manual steps)

? Proceed with migration? (Y/n)
```

You can abort safely at this point by selecting "No".

#### Step 5: Running Migration

Watch real-time progress as the migration runs:

```
Step 5/5: Running Migration

Converting monitors... ████████████████████ 100% (42/42)
Converting heartbeats... ████████████████████ 100% (7/7)
✅ Writing output files... Done!
```

### Final Summary

After completion, you'll see a detailed summary:

```
✅ Migration complete!

Generated files:
  📄 migrated-resources.tf - Terraform configuration (49 resources)
  📜 import.sh - Import script
  📊 migration-report.json - Migration report
  📝 manual-steps.md - Manual configuration steps

⚠️  2 warnings - see migration-report.json for details

Next steps:
  1. Review migrated-resources.tf and adjust as needed
  2. Review manual-steps.md for manual configuration steps
  3. Run: terraform init && terraform plan
  4. Run: terraform apply

📚 Documentation: https://github.com/develeap/terraform-provider-hyperping/tree/main/docs/guides
```

## Platform-Specific Features

### Better Stack

- **Monitors**: HTTP/HTTPS checks
- **Heartbeats**: Cron-job monitoring
- **Warnings**: Displays conversion issues for heartbeats requiring manual cron review

### UptimeRobot

- **Monitor Types**: Shows breakdown by type (HTTP, Keyword, Ping, Port, Heartbeat)
- **Alert Contacts**: Fetches and includes in manual steps
- **Validation Mode**: Option to validate monitors without migrating

### Pingdom

- **Check Types**: Displays supported vs unsupported checks
- **Resource Prefix**: Optional prefix for Terraform resource names
- **Resource Creation**: Can create Hyperping monitors directly or dry-run only

## Disabling Interactive Mode

### Use Existing Workflows

Interactive mode is **automatically disabled** when:

1. **Any flag is provided**:
   ```bash
   go run . --output=custom.tf  # Uses non-interactive mode
   ```

2. **Environment variables are set**:
   ```bash
   export BETTERSTACK_API_TOKEN="token"
   export HYPERPING_API_KEY="sk_key"
   go run .  # Uses non-interactive mode
   ```

3. **Not running in a TTY**:
   ```bash
   go run . < input.txt  # Non-interactive
   go run . | tee output.log  # Non-interactive
   ```

### Explicit Non-Interactive

For automation or CI/CD pipelines, always use flags:

```bash
# Better Stack
go run ./cmd/migrate-betterstack \
  --betterstack-token="$TOKEN" \
  --hyperping-api-key="$API_KEY" \
  --output=migrated.tf \
  --dry-run

# UptimeRobot
go run ./cmd/migrate-uptimerobot \
  --uptimerobot-api-key="$UR_KEY" \
  --hyperping-api-key="$HP_KEY" \
  --validate

# Pingdom
go run ./cmd/migrate-pingdom \
  --pingdom-api-key="$PD_KEY" \
  --hyperping-api-key="$HP_KEY" \
  --output=./migration \
  --dry-run
```

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| Enter | Accept default value / Continue |
| ↑/↓ | Navigate options in select menus |
| Ctrl+C | Cancel operation and exit |
| Tab | Auto-complete (when available) |

## Error Handling

Interactive mode provides helpful error messages:

### API Connection Failures

```
❌ Connection failed: invalid API token
Unable to connect to Better Stack API
ℹ️  Please verify your API token and try again
```

### Invalid Input

```
? Enter your Hyperping API key: test123
✗ Hyperping API keys must start with 'sk_'
```

### File Write Errors

```
❌ Failed to write migrated-resources.tf
Error: permission denied
```

You can safely abort at any point using Ctrl+C.

## Advanced Features

### Progress Indicators

Interactive mode uses appropriate visual indicators:

- **Spinners**: For indeterminate operations (API calls, file writes)
- **Progress bars**: For batch operations (converting resources)
- **Status messages**: Success (✅), Error (❌), Warning (⚠️), Info (ℹ️)

### Color Support

Interactive mode automatically detects terminal capabilities:

- **TTY terminals**: Full color and progress bars
- **Piped output**: Plain text without ANSI codes
- **NO_COLOR environment**: Respects `NO_COLOR=1` to disable colors

### Validation

All inputs are validated before proceeding:

- **API keys**: Format validation (e.g., Hyperping keys must start with `sk_`)
- **File paths**: Checked for invalid characters
- **Confirmations**: Yes/No validation
- **Connection tests**: API connectivity verified before migration

## Examples

### Example 1: Full Better Stack Migration

```bash
$ cd cmd/migrate-betterstack
$ go run .

🚀 Hyperping Migration Tool - Better Stack Edition
═══════════════════════════════════════════════════

This wizard will guide you through migrating your Better Stack
monitors to Hyperping.

Step 1/5: Source Platform Configuration

? Enter your Better Stack API token: ••••••••••••••••
✅ Testing Better Stack API connection... Done! Found 42 monitors and 7 heartbeats

Step 2/5: Destination Platform Configuration

? Perform dry run only (validate without creating files)? No
? Enter your Hyperping API key: ••••••••••••••••
...
```

### Example 2: UptimeRobot Validation Only

```bash
$ cd cmd/migrate-uptimerobot
$ go run .

🚀 Hyperping Migration Tool - UptimeRobot Edition
═════════════════════════════════════════════════

Step 1/5: Source Platform Configuration

? Enter your UptimeRobot API key: ••••••••••••••••
✅ Connected! Found 28 monitors and 3 alert contacts

  Monitor types:
    - HTTP/HTTPS: 20
    - Heartbeat: 5
    - Ping (ICMP): 3

Step 2/5: Migration Mode

? Select migration mode:
  ▸ Full migration (generate files and proceed)
    Dry run (validate without creating files)
    Validate only (check monitors)
```

### Example 3: Pingdom Dry Run

```bash
$ cd cmd/migrate-pingdom
$ go run .

🚀 Hyperping Migration Tool - Pingdom Edition
═════════════════════════════════════════════

Step 1/5: Source Platform Configuration

? Enter your Pingdom API token: ••••••••••••••••
✅ Connected! Found 15 checks

  Check types:
    - http: 12
    - ping: 3

Step 2/5: Migration Mode

? Select migration mode:
    Full migration (create resources in Hyperping)
  ▸ Dry run (generate configs only)
```

## Troubleshooting

### Interactive Mode Not Activating

**Problem**: Tool goes straight to error about missing API key.

**Solution**: Make sure you're not setting environment variables:

```bash
# Check for environment variables
env | grep -E "(BETTERSTACK|UPTIMEROBOT|PINGDOM|HYPERPING)"

# Unset them
unset BETTERSTACK_API_TOKEN
unset HYPERPING_API_KEY
```

### Terminal Not Detected as TTY

**Problem**: Progress bars or prompts not showing.

**Solution**: Run directly in terminal, not through pipes:

```bash
# ❌ Bad
echo "" | go run .

# ✅ Good
go run .
```

### Progress Bars Show Garbled Characters

**Problem**: Terminal doesn't support UTF-8 or ANSI codes.

**Solution**: Disable colors:

```bash
NO_COLOR=1 go run .
```

Or use non-interactive mode:

```bash
go run . --dry-run --verbose
```

### API Connection Timeout

**Problem**: Spinner shows "Testing API connection..." but hangs.

**Solution**: Check network connectivity and firewall rules. Default timeout is 30 seconds.

## Best Practices

### Development

- Use **interactive mode** for initial migrations and testing
- Use **dry-run mode** first to validate configurations
- Review all generated files before applying

### Production

- Use **non-interactive mode** with explicit flags
- Store API keys in secure vaults (not environment variables)
- Run in CI/CD pipelines with `--dry-run` for validation
- Use `--validate` (UptimeRobot) to check resources without migration

### Security

- **Never** share terminal recordings showing API keys
- Use **password managers** to generate and store API keys
- Rotate API keys if accidentally exposed
- Review generated files for sensitive data before committing

## Technical Details

### Interactive Libraries

Interactive mode uses the following Go libraries:

- **survey/v2**: User prompts and input validation
- **spinner**: Loading indicators for async operations
- **progressbar/v3**: Progress bars for batch operations
- **go-isatty**: Terminal detection

### Fallback Behavior

If interactive mode cannot be used:

1. Terminal check fails → Falls back to non-interactive mode
2. Library initialization fails → Shows error and exits
3. User cancels (Ctrl+C) → Graceful shutdown with message

### Performance

Interactive mode adds minimal overhead:

- **API validation**: 1-5 seconds (network dependent)
- **Progress bars**: Updates every 65ms
- **File generation**: Same as non-interactive mode

## See Also

- [Automated Migration Guide](./guides/automated-migration.md)
- [Better Stack Migration Guide](./guides/migrate-from-betterstack.md)
- [UptimeRobot Migration Guide](./guides/migrate-from-uptimerobot.md)
- [Pingdom Migration Guide](./guides/migrate-from-pingdom.md)
- [Error Handling](./guides/error-handling.md)

## Support

For issues or questions:

- **GitHub Issues**: https://github.com/develeap/terraform-provider-hyperping/issues
- **Documentation**: https://github.com/develeap/terraform-provider-hyperping/tree/main/docs
- **Examples**: https://github.com/develeap/terraform-provider-hyperping/tree/main/examples
