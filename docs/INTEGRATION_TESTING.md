# Integration Testing for Migration Tools

This document describes the integration testing framework for the terraform-provider-hyperping migration tools.

## Overview

The integration testing framework validates end-to-end workflows for all 3 migration tools:
- **Better Stack** migration tool
- **UptimeRobot** migration tool
- **Pingdom** migration tool

Each tool has comprehensive integration tests that validate real API interactions and generated Terraform configurations.

## Test Coverage

### Better Stack Migration Tests

Located in: `cmd/migrate-betterstack/integration_test.go`

**Test Scenarios:**
- `TestBetterStackMigration_SmallScenario` - 1-3 monitors, 1 heartbeat
- `TestBetterStackMigration_MediumScenario` - 5-10 monitors, 3 heartbeats
- `TestBetterStackMigration_LargeScenario` - 20-30 monitors, 10 heartbeats
- `TestBetterStackMigration_DryRun` - Validates dry-run mode
- `TestBetterStackMigration_InvalidCredentials` - Error handling

### UptimeRobot Migration Tests

Located in: `cmd/migrate-uptimerobot/integration_test.go`

**Test Scenarios:**
- `TestUptimeRobotMigration_SmallScenario` - 1-3 monitors
- `TestUptimeRobotMigration_MediumScenario` - 5-10 monitors (mixed types)
- `TestUptimeRobotMigration_LargeScenario` - 20-30 monitors (all 5 types)
- `TestUptimeRobotMigration_MonitorTypes` - Tests all 5 monitor types (HTTP, Keyword, Ping, Port, Heartbeat)
- `TestUptimeRobotMigration_ValidateMode` - Validate-only mode
- `TestUptimeRobotMigration_DryRun` - Dry-run mode
- `TestUptimeRobotMigration_InvalidCredentials` - Error handling

### Pingdom Migration Tests

Located in: `cmd/migrate-pingdom/integration_test.go`

**Test Scenarios:**
- `TestPingdomMigration_SmallScenario` - 1-3 checks
- `TestPingdomMigration_MediumScenario` - 5-10 checks
- `TestPingdomMigration_LargeScenario` - 20-30 checks
- `TestPingdomMigration_CheckTypes` - Tests all check types (HTTP, TCP, PING, SMTP)
- `TestPingdomMigration_DryRun` - Dry-run mode
- `TestPingdomMigration_WithPrefix` - Resource name prefix feature
- `TestPingdomMigration_InvalidCredentials` - Error handling

## Validation Steps

Each integration test performs the following validation steps:

1. **API Connection Test** - Verifies connection to source platform API
2. **Migration Tool Execution** - Runs the migration tool without errors
3. **Output Files Generated** - Validates all 4 output files were created:
   - `.tf` (Terraform configuration)
   - `import.sh` (Import script)
   - `report.json` (Migration report)
   - `manual-steps.md` (Manual steps documentation)
4. **Terraform Validation** - Runs `terraform validate` on generated config
5. **Terraform Plan** - Verifies `terraform plan` shows expected resources (0 errors)
6. **Import Script Validation** - Checks script is executable with valid bash syntax
7. **File Content Validation** - Validates JSON and Markdown file formats
8. **Resource Count Validation** - Verifies expected number of resources were generated

## Running Tests Locally

### Prerequisites

```bash
# Required tools
go 1.24+
terraform 1.8+
bash (for import script validation)
```

### Setup Credentials

Create a `.env` file or export environment variables:

```bash
# Better Stack
export BETTERSTACK_API_TOKEN="your_betterstack_token"

# UptimeRobot
export UPTIMEROBOT_API_KEY="your_uptimerobot_key"

# Pingdom
export PINGDOM_API_KEY="your_pingdom_key"

# Hyperping (required for all tools)
export HYPERPING_API_KEY="sk_your_hyperping_key"
```

### Run All Integration Tests

```bash
# Run all integration tests for all 3 tools
go test -v -tags=integration -timeout=30m ./cmd/migrate-betterstack/...
go test -v -tags=integration -timeout=30m ./cmd/migrate-uptimerobot/...
go test -v -tags=integration -timeout=30m ./cmd/migrate-pingdom/...
```

### Run Specific Scenarios

```bash
# Run only small scenario tests
go test -v -tags=integration -run=".*SmallScenario" ./cmd/migrate-betterstack/...

# Run only Better Stack medium scenario
go test -v -tags=integration -run="TestBetterStackMigration_MediumScenario" \
  ./cmd/migrate-betterstack/

# Run only UptimeRobot monitor types test
go test -v -tags=integration -run="TestUptimeRobotMigration_MonitorTypes" \
  ./cmd/migrate-uptimerobot/

# Run only Pingdom dry-run test
go test -v -tags=integration -run="TestPingdomMigration_DryRun" \
  ./cmd/migrate-pingdom/
```

### Run in Short Mode

```bash
# Skip integration tests when running short mode
go test -short ./cmd/migrate-betterstack/...
```

## Running Tests in CI/CD

Integration tests are automatically run via GitHub Actions when:
- Pushing to `main` branch
- Pull requests that modify migration tool code
- Manual workflow dispatch with scenario selection

### Manual Trigger

1. Go to: **Actions** → **Integration Tests** → **Run workflow**
2. Select scenario: `small`, `medium`, `large`, or `all`
3. Click **Run workflow**

### Required Secrets

Configure these secrets in your GitHub repository:

```
Settings → Secrets and variables → Actions → New repository secret
```

**Required secrets:**
- `BETTERSTACK_API_TOKEN` - Better Stack API token
- `UPTIMEROBOT_API_KEY` - UptimeRobot API key
- `PINGDOM_API_KEY` - Pingdom API key/token
- `HYPERPING_API_KEY` - Hyperping API key (starts with `sk_`)

**Security Note:** These secrets are only available to workflows running on the main repository, not from forks.

## Test Infrastructure

### Shared Helpers

Located in: `test/integration/helpers.go`

**Key functions:**
- `GetTestCredentials()` - Retrieve credentials from environment
- `CreateTempTestDir()` - Create temporary test output directory
- `ValidateTerraformFile()` - Run `terraform validate` on generated config
- `ValidateImportScript()` - Validate import script syntax
- `ValidateGeneratedFiles()` - Check all expected files exist
- `RunWithRetry()` - Retry logic for flaky API calls
- `CountTerraformResources()` - Count resource blocks in Terraform file

### Test Configuration

```go
const (
    maxRetries      = 3                // API call retries
    retryDelay      = 2 * time.Second  // Initial retry delay
    testTimeout     = 15 * time.Minute // Test timeout
    cleanupTimeout  = 5 * time.Minute  // Cleanup timeout
)
```

## Troubleshooting

### Tests Skipped (Missing Credentials)

```
⚠️  Skipping test: BETTERSTACK_API_TOKEN not set (required for integration tests)
```

**Solution:** Set the required environment variable or secret.

### API Connection Failed

```
Better Stack API connection failed: 401 Unauthorized
```

**Solution:** Verify your API token is valid and has correct permissions.

### Terraform Validation Failed

```
terraform validate failed: Error: Invalid resource block
```

**Solution:** Check the migration tool's Terraform generation logic. This indicates a bug in the generator.

### Import Script Syntax Error

```
import script has syntax errors: line 10: unexpected token
```

**Solution:** Check the import script generator. Verify bash syntax is correct.

### Test Timeout

```
panic: test timed out after 15m0s
```

**Solution:**
- Increase timeout: `go test -timeout=30m ...`
- Check for API rate limiting
- Verify network connectivity

## Best Practices

### Writing New Integration Tests

1. **Use build tags:** Always add `//go:build integration` at the top
2. **Skip when appropriate:** Use `SkipIfCredentialsMissing()` helper
3. **Create temp dirs:** Use `CreateTempTestDir()` for isolated test output
4. **Cleanup resources:** Register cleanup functions with `t.Cleanup()`
5. **Add retry logic:** Use `RunWithRetry()` for API calls
6. **Validate thoroughly:** Run all 8 validation steps
7. **Log output:** Use `t.Logf()` to log migration tool output

### Test Isolation

- Each test creates its own temporary directory
- Tests run in parallel by default
- No shared state between tests
- Cleanup runs automatically via `t.Cleanup()`

### Performance Optimization

- Use `testing.Short()` to skip slow tests
- Set reasonable timeouts (15 minutes default)
- Run smaller scenarios in PR checks
- Run full suite only on main branch

## Examples

### Example: Adding a New Test Scenario

```go
//go:build integration

func TestBetterStackMigration_CustomScenario(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    creds := integration.GetTestCredentials(t)
    integration.SkipIfCredentialsMissing(t, "BETTERSTACK_API_TOKEN", creds.BetterstackToken)
    integration.SkipIfCredentialsMissing(t, "HYPERPING_API_KEY", creds.HyperpingAPIKey)

    scenario := integration.TestScenario{
        Name:                 "Custom",
        Description:          "Custom test scenario",
        ExpectedMonitors:     10,
        ExpectedHealthchecks: 5,
        MinWarnings:          0,
        MaxWarnings:          10,
    }

    runBetterStackMigrationTest(t, creds, scenario)
}
```

### Example: Running Tests with Verbose Output

```bash
# Run with verbose Go test output
go test -v -tags=integration ./cmd/migrate-betterstack/...

# Run with migration tool verbose output
# (Tests automatically use --verbose flag)
```

## Continuous Improvement

Integration tests should be updated when:
- Adding new migration features
- Supporting new resource types
- Changing output file formats
- Updating API integrations
- Fixing migration bugs

## Support

For issues with integration tests:
1. Check test output logs carefully
2. Verify all credentials are set correctly
3. Test migration tool manually first
4. Review the migration tool's unit tests
5. Open an issue with full test output

## Related Documentation

- [Migration Tools README](../cmd/migrate-betterstack/README.md)
- [Testing Guide](./TESTING.md)
- [Contributing Guide](../CONTRIBUTING.md)
