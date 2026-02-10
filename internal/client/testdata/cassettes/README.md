# VCR Cassettes

This directory contains recorded HTTP interactions for hermetic testing of the Hyperping API client.

## Overview

VCR (Video Cassette Recorder) cassettes allow us to:
- **Test without API keys** - Replay recorded interactions
- **Fast test execution** - No network calls
- **Deterministic tests** - Same results every time
- **CI/CD friendly** - No rate limiting concerns
- **Contract validation** - Detect API breaking changes

## Cassette Index

### Authentication & Authorization

| Cassette | Size | Recorded | Description |
|----------|------|----------|-------------|
| `auth_missing_key.yaml` | 2.3KB | 2026-02-10 | Tests behavior with missing API key |
| `auth_unauthorized.yaml` | 2.3KB | 2026-02-10 | Tests behavior with invalid API key |
| `auth_valid_key.yaml` | 20KB | 2026-02-10 | Validates correct API key across resources |

### Monitor Resource

| Cassette | Size | Recorded | Description |
|----------|------|----------|-------------|
| `monitor_crud.yaml` | 8.3KB | 2026-02-10 | Create, Read, Delete monitor |
| `monitor_list.yaml` | 12KB | 2026-02-10 | List all monitors |
| `monitor_update.yaml` | 8.4KB | 2026-02-10 | Update monitor properties |
| `monitor_pause_resume.yaml` | 12KB | 2026-02-10 | Pause and resume monitor |
| `monitor_not_found.yaml` | 2.5KB | 2026-02-10 | 404 error when monitor doesn't exist |
| `monitor_validation_error.yaml` | 2.7KB | 2026-02-10 | Validation errors on invalid input |

### Incident Resource

| Cassette | Size | Recorded | Description |
|----------|------|----------|-------------|
| `incident_crud.yaml` | 12KB | 2026-02-10 | Create, Read, Delete incident |
| `incident_list.yaml` | 3.4KB | 2026-02-10 | List all incidents |

### Maintenance Resource

| Cassette | Size | Recorded | Description |
|----------|------|----------|-------------|
| `maintenance_crud.yaml` | 26KB | 2026-02-10 | Create, Read, Delete maintenance window |
| `maintenance_list.yaml` | 5.2KB | 2026-02-10 | List all maintenance windows |

### Healthcheck Resource

| Cassette | Size | Recorded | Description |
|----------|------|----------|-------------|
| `healthcheck_crud.yaml` | 14KB | 2026-02-10 | Create, Read, Delete healthcheck |
| `healthcheck_list.yaml` | 9.1KB | 2026-02-10 | List all healthchecks |
| `healthcheck_not_found.yaml` | 2.5KB | 2026-02-10 | 404 error handling |

### Outage Resource

| Cassette | Size | Recorded | Description |
|----------|------|----------|-------------|
| `outage_crud.yaml` | 20KB | 2026-02-10 | Create, Read, Delete outage |
| `outage_list.yaml` | 12KB | 2026-02-10 | List all outages |

### StatusPage Resource

| Cassette | Size | Recorded | Description |
|----------|------|----------|-------------|
| `statuspage_crud.yaml` | 9.2KB | 2026-02-10 | Create, Read, Delete status page |
| `statuspage_list.yaml` | 3.8KB | 2026-02-10 | List status pages |
| `statuspage_list_pagination.yaml` | 2.5KB | 2026-02-10 | Test pagination |
| `statuspage_list_search.yaml` | 2.6KB | 2026-02-10 | Test search functionality |
| `statuspage_update.yaml` | 9.2KB | 2026-02-10 | Update status page properties |
| `statuspage_subscribers.yaml` | 12KB | 2026-02-10 | Subscriber management operations |
| `statuspage_not_found.yaml` | 2.5KB | 2026-02-10 | 404 error handling |

### StatusPage Subscriber Contract Tests

| Cassette | Size | Recorded | Description |
|----------|------|----------|-------------|
| `contract_addsubscriber_email.yaml` | 4.1KB | 2026-02-10 | Add email subscriber |
| `contract_addsubscriber_types.yaml` | 6.3KB | 2026-02-10 | Add multiple subscriber types |
| `contract_addsubscriber_language.yaml` | 4.1KB | 2026-02-10 | Language preference validation |
| `contract_addsubscriber_validation.yaml` | 1.8KB | 2026-02-10 | Input validation errors |
| `contract_deletesubscriber_success.yaml` | 5.2KB | 2026-02-10 | Delete subscriber successfully |
| `contract_deletesubscriber_invalid.yaml` | 1.8KB | 2026-02-10 | Invalid subscriber ID error |
| `contract_deletesubscriber_nonexistent.yaml` | 2.9KB | 2026-02-10 | Non-existent subscriber error |
| `contract_listsubscribers_success.yaml` | 3.2KB | 2026-02-10 | List subscribers |
| `contract_listsubscribers_pagination.yaml` | 3.1KB | 2026-02-10 | Pagination support |
| `contract_listsubscribers_filter.yaml` | 6.6KB | 2026-02-10 | Filter by subscriber type |
| `contract_listsubscribers_invalid.yaml` | 1.1KB | 2026-02-10 | Invalid status page UUID |
| `contract_updatestatuspage_name.yaml` | 4.8KB | 2026-02-10 | Update status page name |
| `contract_updatestatuspage_multiple.yaml` | 4.8KB | 2026-02-10 | Update multiple fields |
| `contract_updatestatuspage_invalid.yaml` | 1.2KB | 2026-02-10 | Validation errors |

### Reporting

| Cassette | Size | Recorded | Description |
|----------|------|----------|-------------|
| `report_get.yaml` | 15KB | 2026-02-10 | Get monitor report for date range |
| `report_list.yaml` | 7.0KB | 2026-02-10 | List all monitor reports |

## Security

All cassettes have sensitive data properly masked:

### Masked Fields
- ✅ **Authorization headers:** `Bearer [MASKED]`
- ✅ **API keys in requests:** `[MASKED]`
- ✅ **Sensitive tokens:** `[MASKED]`

### Security Verification

```bash
# Verify no exposed API keys
grep -r "Bearer hp_" . || echo "✅ No exposed API keys"

# Verify masking is present
grep -l "Bearer \[MASKED\]" *.yaml | wc -l
# Should show: 40 (all cassettes with auth)
```

### What's NOT Masked
- Email addresses (needed for subscriber tests)
- Monitor URLs (public endpoints)
- UUIDs (needed for resource identification)
- Timestamps (needed for validation)

## Using Cassettes

### Replay Mode (Default)

No API key required - uses recorded interactions:

```bash
export VCR_MODE=replay
go test -v -run TestContract ./internal/client/
```

### Record Mode

Records new interactions (requires valid API key):

```bash
export HYPERPING_API_KEY=hp_your_key_here
export VCR_MODE=record
go test -v -run TestLiveContract_Monitor_CRUD ./internal/client/
```

### Verify Recording

```bash
# Check cassette was created
ls -lh monitor_crud.yaml

# Verify sensitive data is masked
grep "Authorization" monitor_crud.yaml
# Should show: - Bearer [MASKED]
```

## Regenerating Cassettes

### When to Regenerate

Regenerate cassettes when:
1. **API changes** - New fields, different response structure
2. **Test changes** - Adding new test scenarios
3. **Cassette corruption** - Malformed YAML
4. **Missing interactions** - Tests fail with "requested interaction not found"

### How to Regenerate

```bash
# Step 1: Delete old cassette
rm internal/client/testdata/cassettes/monitor_crud.yaml

# Step 2: Set up environment
export HYPERPING_API_KEY=hp_your_key_here
export VCR_MODE=record

# Step 3: Run test to record
go test -v -run TestLiveContract_Monitor_CRUD ./internal/client/

# Step 4: Verify cassette
ls -lh internal/client/testdata/cassettes/monitor_crud.yaml

# Step 5: Test in replay mode
unset HYPERPING_API_KEY
export VCR_MODE=replay
go test -v -run TestContract_Monitor ./internal/client/
```

### Batch Regeneration

To regenerate all cassettes:

```bash
#!/bin/bash
export HYPERPING_API_KEY=hp_your_key_here
export VCR_MODE=record

# Remove all cassettes
rm internal/client/testdata/cassettes/*.yaml

# Record all LiveContract tests
go test -v -run TestLiveContract ./internal/client/

# Verify in replay mode
unset HYPERPING_API_KEY
export VCR_MODE=replay
go test -v -run TestContract ./internal/client/
```

## Validation Checklist

After recording or updating cassettes:

- [ ] **No exposed secrets** - `grep -r "Bearer hp_" .` returns nothing
- [ ] **Proper masking** - All auth headers show `[MASKED]`
- [ ] **Valid YAML** - No syntax errors
- [ ] **Complete interactions** - All test scenarios recorded
- [ ] **Replay works** - Tests pass without API key
- [ ] **Deterministic** - Multiple runs produce same results

## Troubleshooting

### "requested interaction not found"

**Cause:** Cassette doesn't contain the requested interaction

**Solutions:**
1. Re-record the cassette in record mode
2. Check if test scenario changed
3. Verify VCR_MODE is set to "replay"

### Test fails in replay but works in record mode

**Cause:** Cassette contains interactions that depend on external state

**Solutions:**
1. Use fixed timestamps instead of `time.Now()`
2. Use predictable UUIDs from setup
3. Ensure test is idempotent

### Cassette contains exposed secrets

**Cause:** VCR matcher not configured to mask sensitive headers

**Solutions:**
1. Check VCR configuration in `testutil` package
2. Verify `BeforeSave` hook is masking headers
3. Re-record cassette with proper configuration

### Tests are slow even in replay mode

**Cause:** VCR might be falling back to real HTTP calls

**Solutions:**
1. Verify `VCR_MODE=replay` is set
2. Check cassette file exists
3. Verify cassette path is correct

## Best Practices

### Recording Best Practices

1. **Use unique names** - Avoid conflicts with existing resources
2. **Clean up after tests** - Delete created resources
3. **Use fixed timestamps** - Avoid time-dependent failures
4. **Test idempotency** - Tests should work multiple times
5. **Verify masking** - Check cassettes before committing

### Cassette Organization

1. **One cassette per test** - Easy to understand and maintain
2. **Descriptive names** - `{resource}_{operation}.yaml`
3. **Separate contract tests** - Prefix with `contract_`
4. **Group by resource** - Easy to find related cassettes

### Maintenance

1. **Regular updates** - Keep cassettes current with API
2. **Version control** - Commit cassettes to git
3. **Documentation** - Update README when adding cassettes
4. **Validation** - Run security checks regularly

## Statistics

- **Total Cassettes:** 41
- **Total Size:** 364KB
- **Average Size:** 8.9KB
- **Largest:** `maintenance_crud.yaml` (26KB)
- **Smallest:** `contract_listsubscribers_invalid.yaml` (1.1KB)
- **Resources Covered:** 7 (Monitor, Incident, Maintenance, Healthcheck, Outage, StatusPage, Reports)
- **Contract Tests:** 356
- **Security:** ✅ All secrets masked

## Related Documentation

- **Contract Tests:** See `../contract_test_helpers.go`
- **VCR Configuration:** See `../../provider/testutil/vcr.go`
- **Test Guide:** See `/tmp/VCR_CONTRACT_TEST_REPORT.md`

---

**Last Updated:** 2026-02-10
**Maintainer:** Development Team
**Questions:** See project documentation or open an issue
