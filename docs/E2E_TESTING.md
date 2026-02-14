# End-to-End (E2E) Testing Guide

## Overview

This guide covers the comprehensive E2E testing framework for validating migration tools from source platforms (Better Stack, UptimeRobot, Pingdom) to Hyperping. E2E tests simulate complete migration workflows including Terraform execution and resource verification.

## What E2E Tests Validate

The E2E testing pipeline validates the entire migration workflow:

1. **Resource Creation** - Programmatically creates monitors in source platform
2. **Migration Tool Execution** - Runs migration tool to generate all output files
3. **Terraform Initialization** - `terraform init` succeeds
4. **Terraform Planning** - `terraform plan` shows no errors
5. **Terraform Application** - `terraform apply` creates resources successfully
6. **Resource Verification** - Created resources match source configuration
7. **Import Script Execution** - Import script runs without errors
8. **State Validation** - Terraform state matches configuration
9. **Cleanup** - All resources are deleted (idempotent)

## Prerequisites

### Required Software

- **Go 1.21+** - For running tests
- **Terraform 1.8+** - For infrastructure validation
- **Bash** - For import script execution
- **Git** - For version control

### Required API Keys

You must have valid API credentials for all platforms you want to test:

| Platform | Environment Variable | How to Obtain |
|----------|---------------------|---------------|
| Better Stack | `BETTERSTACK_API_TOKEN` | https://betteruptime.com/team/api |
| UptimeRobot | `UPTIMEROBOT_API_KEY` | https://uptimerobot.com/dashboard#mySettings |
| Pingdom | `PINGDOM_API_KEY` | https://my.pingdom.com/app/api-tokens |
| Hyperping | `HYPERPING_API_KEY` | https://hyperping.io/settings/api |

**Important:** E2E tests create and delete real resources. Use test/staging accounts when possible.

## Quick Start

### 1. Set Environment Variables

```bash
# Required for all E2E tests
export HYPERPING_API_KEY="your_hyperping_api_key"

# Required for Better Stack tests
export BETTERSTACK_API_TOKEN="your_betterstack_token"

# Required for UptimeRobot tests
export UPTIMEROBOT_API_KEY="your_uptimerobot_key"

# Required for Pingdom tests
export PINGDOM_API_KEY="your_pingdom_api_key"
```

### 2. Run E2E Tests

```bash
# Run all E2E tests (requires all API keys)
go test -tags=e2e ./test/e2e/... -v -timeout=30m

# Run specific platform tests
go test -tags=e2e ./test/e2e -run TestBetterStackE2E -v -timeout=30m
go test -tags=e2e ./test/e2e -run TestUptimeRobotE2E -v -timeout=30m
go test -tags=e2e ./test/e2e -run TestPingdomE2E -v -timeout=30m

# Run small tests only (faster, good for CI)
go test -tags=e2e ./test/e2e -run Small -v -timeout=15m

# Skip medium/large tests in short mode
go test -tags=e2e ./test/e2e -short -v -timeout=15m
```

## Test Scenarios

### Small Migration (1-3 monitors)

- **Duration:** ~5-10 minutes per platform
- **Purpose:** Quick validation of core functionality
- **Coverage:** Basic monitor creation, migration, and cleanup

```bash
go test -tags=e2e ./test/e2e -run SmallMigration -v
```

### Medium Migration (10+ monitors)

- **Duration:** ~15-20 minutes per platform
- **Purpose:** Stress test with realistic workload
- **Coverage:** Multiple monitor types, various configurations
- **Skipped in:** `go test -short` mode

```bash
go test -tags=e2e ./test/e2e -run MediumMigration -v
```

### Large Migration (50+ monitors)

- **Duration:** ~30+ minutes per platform
- **Purpose:** Performance validation at scale
- **Status:** Optional, may hit API rate limits
- **Note:** Not yet implemented for all platforms

### Error Handling Tests

- **Duration:** ~2-5 minutes per platform
- **Purpose:** Validate error scenarios and edge cases
- **Coverage:** Invalid credentials, empty sources, etc.

```bash
go test -tags=e2e ./test/e2e -run ErrorHandling -v
```

## Test Architecture

### Directory Structure

```
test/e2e/
├── helpers.go                 # Shared E2E utilities
├── fixtures/
│   └── monitors.go            # Test data generators
├── betterstack_test.go        # Better Stack E2E tests
├── uptimerobot_test.go        # UptimeRobot E2E tests
└── pingdom_test.go            # Pingdom E2E tests
```

### Key Components

#### 1. Test Helpers (`helpers.go`)

Provides shared utilities for E2E testing:

- **TerraformExecutor** - Runs Terraform commands (init, plan, apply, destroy)
- **MigrationToolExecutor** - Runs migration tools with configuration
- **HyperpingResourceManager** - Manages Hyperping resources (create, list, delete)
- **TestCredentials** - Retrieves and validates API credentials
- **Cleanup Management** - Ensures resources are deleted after tests

#### 2. Test Fixtures (`fixtures/monitors.go`)

Generates test data for various scenarios:

- `GetSmallScenarioMonitors()` - 1-3 monitors
- `GetMediumScenarioMonitors()` - 10+ monitors with variety
- `GetLargeScenarioMonitors()` - 50+ monitors for stress testing

#### 3. Platform-Specific Test Clients

Each platform has an extended test client with create/delete methods:

- **BetterStackTestClient** - Extends betterstack.Client
- **UptimeRobotTestClient** - Extends uptimerobot.Client
- **PingdomTestClient** - Extends pingdom.Client

## Test Workflow

Each E2E test follows this workflow:

```
1. Setup Phase
   ├── Load credentials from environment
   ├── Create test clients
   ├── Register cleanup handlers
   └── Create temporary work directory

2. Resource Creation Phase
   ├── Generate unique test resource names
   ├── Create monitors in source platform
   └── Verify creation successful

3. Migration Phase
   ├── Execute migration tool
   ├── Validate all output files generated
   │   ├── migrated-resources.tf
   │   ├── import.sh
   │   ├── migration-report.json
   │   └── manual-steps.md
   └── Validate Terraform syntax

4. Terraform Execution Phase
   ├── terraform init
   ├── terraform plan (validate no errors)
   ├── terraform apply (create resources)
   └── Validate resources in Hyperping

5. Import Validation Phase
   ├── Execute import script
   ├── Verify Terraform state
   └── Run terraform plan (should show no changes)

6. Cleanup Phase
   ├── terraform destroy
   ├── Delete from source platform
   ├── Delete from Hyperping
   └── Remove temporary directories
```

## Cleanup Strategy

E2E tests use comprehensive cleanup to prevent resource leaks:

### Automatic Cleanup

Tests use `t.Cleanup()` to automatically delete resources on:
- Test completion
- Test failure
- Test timeout
- Context cancellation

### Manual Cleanup

If tests are interrupted (e.g., Ctrl+C), resources may remain. Clean up manually:

```bash
# Cleanup Hyperping test resources
# List monitors with E2E-Test prefix and delete them via API or UI

# Cleanup source platform test resources
# Better Stack: Delete monitors with "E2E-Test-" prefix
# UptimeRobot: Delete monitors with "E2E-Test-" prefix
# Pingdom: Delete checks with "E2E-Test-" prefix
```

### Idempotent Cleanup

Cleanup functions are idempotent - safe to run multiple times:

```bash
# Run cleanup tests to clean up orphaned resources
go test -tags=e2e ./test/e2e -run IdempotentCleanup -v
```

## Test Naming Convention

All test resources use a unique prefix:

```
E2E-Test-{Platform}-{Scenario}-{Timestamp}-{UUID}
```

Example:
```
E2E-Test-BS-Small-20260213-150405-a1b2c3d4
```

This ensures:
- Resources are identifiable as test resources
- No conflicts with production resources
- Easy cleanup via name prefix filtering

## Troubleshooting

### Tests Failing Due to API Rate Limits

**Symptom:** Tests fail with 429 (Too Many Requests) errors

**Solutions:**
- Add delays between tests: `time.Sleep(5 * time.Second)`
- Run fewer tests in parallel
- Use separate API keys for testing
- Contact platform support to increase rate limits

### Terraform Init Failures

**Symptom:** `terraform init` fails with provider download errors

**Solutions:**
```bash
# Clear Terraform cache
rm -rf ~/.terraform.d/plugin-cache

# Ensure Terraform provider is available
terraform init -upgrade

# Check provider registry availability
curl -I https://registry.terraform.io/v1/providers/develeap/hyperping
```

### Resources Not Cleaned Up

**Symptom:** Test resources remain after test completion

**Solutions:**
```bash
# Run idempotent cleanup tests
go test -tags=e2e ./test/e2e -run IdempotentCleanup -v

# Or use cleanup scripts per platform
# See cleanup strategy section above
```

### Import Script Failures

**Symptom:** Import script execution fails

**Common causes:**
1. Resources already imported
2. Resource IDs changed
3. Terraform state out of sync

**Solutions:**
```bash
# Remove state and re-import
terraform state rm 'hyperping_monitor.example'

# Verify resource exists
terraform state list

# Re-run import
bash import.sh
```

### Context Deadline Exceeded

**Symptom:** Tests timeout with "context deadline exceeded"

**Solutions:**
```bash
# Increase timeout
go test -tags=e2e ./test/e2e -timeout=45m -v

# Or run fewer tests
go test -tags=e2e ./test/e2e -run SmallMigration -v
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: E2E Tests

on:
  pull_request:
    branches: [main]
  schedule:
    - cron: '0 2 * * *'  # Run nightly

jobs:
  e2e-tests:
    runs-on: ubuntu-latest
    timeout-minutes: 60

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: '1.8'

      - name: Run E2E Tests
        env:
          BETTERSTACK_API_TOKEN: ${{ secrets.BETTERSTACK_API_TOKEN }}
          UPTIMEROBOT_API_KEY: ${{ secrets.UPTIMEROBOT_API_KEY }}
          PINGDOM_API_KEY: ${{ secrets.PINGDOM_API_KEY }}
          HYPERPING_API_KEY: ${{ secrets.HYPERPING_API_KEY }}
        run: |
          go test -tags=e2e ./test/e2e/... -v -timeout=45m -short

      - name: Cleanup on Failure
        if: failure()
        run: |
          go test -tags=e2e ./test/e2e -run IdempotentCleanup -v
```

### GitLab CI Example

```yaml
e2e:tests:
  stage: test
  image: golang:1.21
  timeout: 60m
  before_script:
    - apt-get update && apt-get install -y wget unzip
    - wget https://releases.hashicorp.com/terraform/1.8.0/terraform_1.8.0_linux_amd64.zip
    - unzip terraform_1.8.0_linux_amd64.zip -d /usr/local/bin
  script:
    - go test -tags=e2e ./test/e2e/... -v -timeout=45m -short
  variables:
    BETTERSTACK_API_TOKEN: $BETTERSTACK_API_TOKEN
    UPTIMEROBOT_API_KEY: $UPTIMEROBOT_API_KEY
    PINGDOM_API_KEY: $PINGDOM_API_KEY
    HYPERPING_API_KEY: $HYPERPING_API_KEY
  only:
    - merge_requests
    - main
```

## Best Practices

### 1. Use Dedicated Test Accounts

Create separate test accounts for each platform to avoid interfering with production:
- Isolates test resources from production
- Prevents accidental deletion of production monitors
- Allows higher API rate limits

### 2. Run Tests in Isolated Environments

```bash
# Use temporary directories
export TMPDIR=/tmp/e2e-tests

# Clean up before running
go clean -testcache
```

### 3. Monitor Test Duration

```bash
# Track test duration
time go test -tags=e2e ./test/e2e -run SmallMigration -v

# Profile slow tests
go test -tags=e2e ./test/e2e -cpuprofile=cpu.prof -v
```

### 4. Validate Before Committing

Run E2E tests locally before pushing:

```bash
# Quick validation
go test -tags=e2e ./test/e2e -run Small -v -short

# Full validation (slower)
go test -tags=e2e ./test/e2e -v -timeout=30m
```

### 5. Handle Flaky Tests

If tests are flaky due to network issues:

```bash
# Retry failed tests
go test -tags=e2e ./test/e2e -count=2 -v

# Or increase retry delays in helpers.go
```

## Performance Benchmarks

Expected test durations (approximate):

| Test Scenario | Per Platform | All Platforms |
|---------------|-------------|---------------|
| Small Migration | 5-10 min | 15-30 min |
| Medium Migration | 15-20 min | 45-60 min |
| Error Handling | 2-5 min | 6-15 min |
| Full Suite | 25-35 min | 75-105 min |

## Contributing

When adding new E2E tests:

1. Follow existing test structure and naming conventions
2. Use test fixtures from `fixtures/monitors.go`
3. Implement proper cleanup with `SetupTestCleanup()`
4. Add timeouts appropriate for test complexity
5. Document any new environment variables needed
6. Update this guide with new scenarios

## FAQ

**Q: Can I run E2E tests without all API keys?**
A: Yes, tests will skip if credentials are missing. Set only the keys you have.

**Q: Are resources really created in the platforms?**
A: Yes, E2E tests create real resources. Always use test accounts.

**Q: What if I don't have a Terraform provider published?**
A: E2E tests work with local provider builds. No registry required.

**Q: Can I run tests in parallel?**
A: Tests within a file run sequentially. Multiple files can run in parallel with `-p` flag.

**Q: How do I debug a failing E2E test?**
A: Add `-v` flag for verbose output. Check temporary directories for generated files.

**Q: What about cost?**
A: Most monitoring platforms have free tiers. E2E tests typically stay within free limits.

## Support

For issues with E2E tests:

1. Check this documentation first
2. Review test logs for error messages
3. Verify API credentials are valid
4. Check platform status pages
5. Open an issue with test output

## Related Documentation

- [Integration Testing](./INTEGRATION_TESTING.md) - Integration test framework
- [Testing Strategy](./TESTING_STRATEGY.md) - Overall testing approach
- [Migration Tools](./MIGRATION_TOOLS.md) - Migration tool usage
- [Contributing](../CONTRIBUTING.md) - Contribution guidelines
