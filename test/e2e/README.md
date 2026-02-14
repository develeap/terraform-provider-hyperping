# E2E Testing Framework

## Overview

This directory contains the comprehensive End-to-End (E2E) testing framework for validating migration tools. E2E tests simulate complete migration workflows from source platforms (Better Stack, UptimeRobot, Pingdom) to Hyperping, including Terraform execution and resource verification.

## Quick Reference

```bash
# Run all E2E tests
./scripts/run-e2e-tests.sh

# Run specific platform tests
./scripts/run-e2e-tests.sh -p BetterStack
./scripts/run-e2e-tests.sh -p UptimeRobot
./scripts/run-e2e-tests.sh -p Pingdom

# Run small tests only (faster)
./scripts/run-e2e-tests.sh -p Small -t 15m

# Run in short mode (skip slow tests)
./scripts/run-e2e-tests.sh -s
```

## Test Statistics

- **Total E2E Tests:** 12
- **Total Lines of Code:** 1,891
- **Test Scenarios:** Small, Medium, Error Handling, Cleanup
- **Platforms Covered:** Better Stack, UptimeRobot, Pingdom
- **Expected Duration:** 25-35 minutes per platform, 75-105 minutes full suite

## File Structure

```
test/e2e/
├── README.md                 # This file
├── helpers.go                # Shared E2E utilities (520 lines)
├── fixtures/
│   └── monitors.go           # Test data generators (153 lines)
├── betterstack_test.go       # Better Stack E2E tests (423 lines)
├── uptimerobot_test.go       # UptimeRobot E2E tests (378 lines)
└── pingdom_test.go           # Pingdom E2E tests (369 lines)
```

## Test Coverage

### Per Platform Tests

Each platform has the following test suite:

1. **SmallMigration** (1-3 monitors)
   - Quick validation (~5-10 minutes)
   - Full workflow: create → migrate → terraform → verify → cleanup

2. **MediumMigration** (10+ monitors)
   - Realistic workload (~15-20 minutes)
   - Skipped in short mode

3. **ErrorHandling**
   - Invalid credentials
   - Empty source platform
   - Edge cases

4. **IdempotentCleanup**
   - Verifies cleanup can run multiple times
   - Useful for manual cleanup of orphaned resources

### Test Workflow

Each E2E test validates:

1. ✅ Resources created in source platform via API
2. ✅ Migration tool generates all 4 output files
3. ✅ `terraform init` succeeds
4. ✅ `terraform plan` shows no errors
5. ✅ `terraform apply` creates resources
6. ✅ Resources verified in Hyperping (names, URLs, configuration)
7. ✅ Import script executes successfully
8. ✅ Terraform state validated
9. ✅ Cleanup deletes all resources (idempotent)

## Prerequisites

### Required Software

- Go 1.21+
- Terraform 1.8+
- Bash
- Git

### Required API Keys

Set these environment variables before running tests:

```bash
# Required for all tests
export HYPERPING_API_KEY="your_hyperping_api_key"

# Required for specific platforms
export BETTERSTACK_API_TOKEN="your_betterstack_token"
export UPTIMEROBOT_API_KEY="your_uptimerobot_key"
export PINGDOM_API_KEY="your_pingdom_api_key"
```

**Warning:** E2E tests create and delete real resources. Use test/staging accounts.

## Key Components

### 1. Test Helpers (`helpers.go`)

Provides core E2E functionality:

- **TerraformExecutor** - Terraform command execution (init, plan, apply, destroy, import)
- **MigrationToolExecutor** - Migration tool execution and output validation
- **HyperpingResourceManager** - Hyperping resource management (create, list, delete, verify)
- **TestCredentials** - Credential loading and validation
- **Cleanup Management** - Automatic resource cleanup with retry logic
- **Utilities** - Unique name generation, context management, retry logic

### 2. Test Fixtures (`fixtures/monitors.go`)

Generates test data for various scenarios:

- **Small Scenario:** 1-3 monitors with basic configuration
- **Medium Scenario:** 10+ monitors with varied types and frequencies
- **Large Scenario:** 50+ monitors for stress testing (future)

### 3. Platform Test Clients

Extended clients with create/delete capabilities for E2E testing:

- **BetterStackTestClient** - Creates/deletes monitors via Better Stack API
- **UptimeRobotTestClient** - Creates/deletes monitors via UptimeRobot API
- **PingdomTestClient** - Creates/deletes checks via Pingdom API

Each client includes:
- Resource creation methods
- Resource deletion methods
- Bulk cleanup functions
- Test resource filtering (by name prefix)

## Test Scenarios Detail

### Better Stack Tests (`betterstack_test.go`)

- `TestBetterStackE2E_SmallMigration` - 2 monitors, full validation
- `TestBetterStackE2E_MediumMigration` - 12 monitors, realistic load
- `TestBetterStackE2E_ErrorHandling` - Invalid token, empty source
- `TestBetterStackE2E_IdempotentCleanup` - Cleanup verification

### UptimeRobot Tests (`uptimerobot_test.go`)

- `TestUptimeRobotE2E_SmallMigration` - 2 monitors, full validation
- `TestUptimeRobotE2E_MediumMigration` - 12 monitors, realistic load
- `TestUptimeRobotE2E_ErrorHandling` - Invalid API key
- `TestUptimeRobotE2E_IdempotentCleanup` - Cleanup verification

### Pingdom Tests (`pingdom_test.go`)

- `TestPingdomE2E_SmallMigration` - 2 checks, full validation
- `TestPingdomE2E_MediumMigration` - 12 checks, realistic load
- `TestPingdomE2E_ErrorHandling` - Invalid API key
- `TestPingdomE2E_IdempotentCleanup` - Cleanup verification

## Running Tests

### Using Test Runner Script

```bash
# Run all tests
./scripts/run-e2e-tests.sh

# Run with options
./scripts/run-e2e-tests.sh -p SmallMigration -s -t 15m
```

### Using Go Test Directly

```bash
# Run all E2E tests
go test -tags=e2e ./test/e2e -v -timeout=30m

# Run specific test
go test -tags=e2e ./test/e2e -run TestBetterStackE2E_SmallMigration -v

# Run in short mode
go test -tags=e2e ./test/e2e -short -v

# Run with pattern
go test -tags=e2e ./test/e2e -run "Small|Error" -v
```

## Cleanup

### Automatic Cleanup

Tests automatically clean up resources via `t.Cleanup()`. Resources are deleted on:
- Test completion
- Test failure
- Test timeout

### Manual Cleanup

If tests are interrupted:

```bash
# Run cleanup tests
go test -tags=e2e ./test/e2e -run IdempotentCleanup -v

# Or manually delete via API/UI:
# - Search for resources with "E2E-Test-" prefix
# - Delete from source platform
# - Delete from Hyperping
```

## Test Naming Convention

All test resources use this prefix:

```
E2E-Test-{Platform}-{Scenario}-{Timestamp}-{UUID}
```

Examples:
- `E2E-Test-BS-Small-20260213-150405-a1b2c3d4`
- `E2E-Test-UR-Med-20260213-151230-e5f6g7h8`

This ensures:
- Easy identification of test resources
- No conflicts with production
- Automatic cleanup by prefix

## CI/CD Integration

### GitHub Actions

See `.github/workflows/e2e-tests.yml` (example):

```yaml
- name: Run E2E Tests
  env:
    BETTERSTACK_API_TOKEN: ${{ secrets.BETTERSTACK_API_TOKEN }}
    UPTIMEROBOT_API_KEY: ${{ secrets.UPTIMEROBOT_API_KEY }}
    PINGDOM_API_KEY: ${{ secrets.PINGDOM_API_KEY }}
    HYPERPING_API_KEY: ${{ secrets.HYPERPING_API_KEY }}
  run: ./scripts/run-e2e-tests.sh -s -t 45m
```

### Recommended CI Strategy

1. **PR Tests:** Run small migration tests only (fast feedback)
2. **Merge to Main:** Run full test suite
3. **Nightly:** Run all tests including large scenarios
4. **Manual Trigger:** Run specific platform tests on demand

## Troubleshooting

### Tests Skip with "credential not set"

**Solution:** Export required API keys before running tests

### Tests Fail with "terraform init failed"

**Solution:** Ensure Terraform is installed and in PATH

### Tests Fail with "429 Too Many Requests"

**Solution:** Add delays between tests or use separate API keys

### Resources Not Cleaned Up

**Solution:** Run `go test -tags=e2e ./test/e2e -run IdempotentCleanup -v`

## Performance Benchmarks

Measured on standard development machine:

| Test Type | Per Platform | All Platforms |
|-----------|-------------|---------------|
| Small Migration | 5-10 min | 15-30 min |
| Medium Migration | 15-20 min | 45-60 min |
| Error Handling | 2-5 min | 6-15 min |
| Full Suite | 25-35 min | 75-105 min |

## Best Practices

1. **Use Test Accounts** - Never run E2E tests against production
2. **Run Small Tests First** - Quick validation before full suite
3. **Monitor Rate Limits** - Watch for 429 errors
4. **Clean Up Regularly** - Run cleanup tests to prevent resource leaks
5. **Check Logs** - Use `-v` flag for detailed output
6. **Isolate Tests** - Use separate API keys per environment

## Contributing

When adding new E2E tests:

1. Follow existing test structure
2. Use fixtures from `fixtures/monitors.go`
3. Implement proper cleanup with `SetupTestCleanup()`
4. Add appropriate timeouts
5. Document new environment variables
6. Update test count in this README

## Documentation

- **Full Guide:** [docs/E2E_TESTING.md](../../docs/E2E_TESTING.md)
- **Integration Tests:** [test/integration/README.md](../integration/README.md)
- **Migration Tools:** [docs/MIGRATION_TOOLS.md](../../docs/MIGRATION_TOOLS.md)

## Support

For issues with E2E tests:

1. Check [docs/E2E_TESTING.md](../../docs/E2E_TESTING.md) troubleshooting section
2. Review test logs with `-v` flag
3. Verify API credentials are valid
4. Run cleanup tests if resources remain
5. Open an issue with test output

## License

Copyright (c) 2026 Develeap. Licensed under MPL-2.0.
