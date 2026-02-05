# Rate Limits and Capacity Planning Guide

This guide documents API call counts per Terraform operation and strategies for managing rate limits in large deployments.

> **Note:** Hyperping doesn't charge per API call. This guide helps you stay within rate limits and plan for large deployments.

## API Call Counts Per Operation

### Summary Table

| Resource | Create | Read | Update | Delete | Import |
|----------|--------|------|--------|--------|--------|
| Monitor | 1 | 1 | 1 | 1 | 1 |
| Healthcheck | 1 | 1 | 1 | 1 | 1 |
| Incident | 1 | 1 | 1 | 1 | 1 |
| Incident Update | 1 | 1 | 1 | 1 | 1 |
| Maintenance | 1 | 1 | 1 | 1 | 1 |
| Status Page | 1 | 1 | 1 | 1 | 1 |
| Subscriber | 1 | 1 | 1 | 1 | 1 |
| Outage | 1 | 1 | 1 | 1 | 1 |

### Data Sources

| Data Source | API Calls | Notes |
|-------------|-----------|-------|
| monitors (list) | 1 | Single paginated request |
| monitor (single) | 1 | Direct lookup by ID |
| healthchecks | 1 | Single paginated request |
| incidents | 1 | Single paginated request |
| statuspages | 1 | Single paginated request |
| outages | 1 | Single paginated request |

## Terraform Operation Costs

### terraform plan

| Scenario | API Calls | Formula |
|----------|-----------|---------|
| Empty state | 0 | No reads needed |
| N resources | N | 1 read per resource (refresh) |
| With `-refresh=false` | 0 | Skips state refresh |

### terraform apply

| Scenario | API Calls | Formula |
|----------|-----------|---------|
| Create N resources | N | 1 create per resource |
| Update N resources | N | 1 update per resource |
| Delete N resources | N | 1 delete per resource |
| Mixed (refresh + changes) | 2N | N reads + N writes |

### terraform import

| Operation | API Calls |
|-----------|-----------|
| Import 1 resource | 1 |
| Import N resources | N |

### terraform refresh

| Scenario | API Calls |
|----------|-----------|
| Refresh N resources | N |

## Rate Limits

### Hyperping API Limits

| Limit Type | Value | Reset Period |
|------------|-------|--------------|
| Requests per minute | 60 | 1 minute |
| Requests per hour | 1000 | 1 hour |
| Concurrent requests | 10 | N/A |

*Note: Contact Hyperping support for enterprise rate limit increases.*

### Provider Retry Behavior

The provider automatically handles rate limits:

```go
// Built-in retry configuration
Max Retries: 3
Initial Delay: 1 second
Max Delay: 30 seconds
Backoff: Exponential with jitter
```

When you see `429 Too Many Requests`:

1. Provider automatically waits and retries
2. Respects `Retry-After` header from API
3. Logs retry attempts for debugging

## Capacity Planning

### Small Deployment (< 50 resources)

| Metric | Value |
|--------|-------|
| Resources | < 50 |
| Plan API calls | ~50 |
| Apply API calls | ~100 (worst case) |
| Daily refresh calls | ~50 |
| **Recommended parallelism** | Default (10) |

```bash
terraform apply  # Default settings work fine
```

### Medium Deployment (50-200 resources)

| Metric | Value |
|--------|-------|
| Resources | 50-200 |
| Plan API calls | ~200 |
| Apply API calls | ~400 (worst case) |
| Daily refresh calls | ~200 |
| **Recommended parallelism** | 5 |

```bash
terraform apply -parallelism=5
```

### Large Deployment (200+ resources)

| Metric | Value |
|--------|-------|
| Resources | 200+ |
| Plan API calls | 200+ |
| Apply API calls | 400+ (worst case) |
| Daily refresh calls | 200+ |
| **Recommended parallelism** | 2-3 |

```bash
terraform apply -parallelism=2
```

### Enterprise Deployment (1000+ resources)

For deployments with 1000+ resources:

1. **Request rate limit increase** from Hyperping support
2. **Split state files** by environment/team:

```
terraform/
├── production/
│   ├── api-monitors/       # 100 monitors
│   ├── database-monitors/  # 100 monitors
│   └── status-pages/       # 10 status pages
├── staging/
│   └── monitors/           # 50 monitors
└── development/
    └── monitors/           # 50 monitors
```

3. **Use targeted applies**:

```bash
# Apply only specific resources
terraform apply -target=hyperping_monitor.critical_api
```

4. **Implement state locking** to prevent concurrent runs:

```hcl
terraform {
  backend "s3" {
    bucket         = "terraform-state"
    key            = "hyperping/production.tfstate"
    region         = "us-east-1"
    dynamodb_table = "terraform-locks"
    encrypt        = true
  }
}
```

## Optimization Strategies

### 1. Use -refresh=false for Large Plans

When you only need to see what would change (not current state):

```bash
terraform plan -refresh=false
```

**Savings:** Eliminates N read API calls

### 2. Target Specific Resources

When updating a single resource:

```bash
terraform apply -target=hyperping_monitor.api
```

**Savings:** Only 2 API calls (read + update) instead of N+1

### 3. Batch Changes

Group related changes together rather than running multiple applies:

```bash
# Bad: Multiple applies
terraform apply -target=hyperping_monitor.api1
terraform apply -target=hyperping_monitor.api2
terraform apply -target=hyperping_monitor.api3

# Good: Single apply
terraform apply -target=hyperping_monitor.api1 \
                -target=hyperping_monitor.api2 \
                -target=hyperping_monitor.api3
```

### 4. Use Data Sources Sparingly

Data sources refresh on every plan/apply. If you don't need live data:

```hcl
# Avoid: Refreshes every plan
data "hyperping_monitors" "all" {}

# Better: Use resource references
resource "hyperping_incident" "outage" {
  status_pages = [hyperping_statuspage.main.id]  # Direct reference
}
```

### 5. State Import vs. Recreate

Importing existing resources uses 1 API call. Recreating uses 2 (delete + create):

```bash
# Better: Import existing (1 API call)
terraform import hyperping_monitor.api "mon_abc123"

# Worse: Delete and recreate (2 API calls)
# (happens if resource is removed from state and re-added to config)
```

## CI/CD Considerations

### GitHub Actions Rate Limiting

For CI/CD pipelines, add delays between jobs:

```yaml
jobs:
  plan:
    runs-on: ubuntu-latest
    steps:
      - uses: hashicorp/setup-terraform@v3
      - run: terraform plan -out=plan.tfplan

  apply:
    needs: plan
    runs-on: ubuntu-latest
    steps:
      - run: sleep 60  # Wait 1 minute between plan and apply
      - run: terraform apply plan.tfplan
```

### Concurrency Limits

Prevent multiple Terraform runs:

```yaml
concurrency:
  group: terraform-${{ github.ref }}
  cancel-in-progress: false  # Don't cancel running applies!
```

### Scheduled Refreshes

For drift detection, schedule during off-peak hours:

```yaml
on:
  schedule:
    - cron: '0 3 * * *'  # 3 AM UTC daily

jobs:
  drift-check:
    runs-on: ubuntu-latest
    steps:
      - run: terraform plan -detailed-exitcode
```

## Monitoring API Usage

### Enable Debug Logging

```bash
TF_LOG=DEBUG terraform apply 2>&1 | grep -i "api\|request\|response"
```

### Count API Calls

```bash
TF_LOG=DEBUG terraform plan 2>&1 | grep -c "HTTP"
```

### Track Over Time

Log API calls in CI/CD:

```yaml
- name: Terraform Plan
  run: |
    TF_LOG=DEBUG terraform plan 2>&1 | tee plan.log
    echo "API calls: $(grep -c 'HTTP' plan.log)"
```

## Troubleshooting

### Error: 429 Too Many Requests

```
Error: API error: 429 Too Many Requests
```

**Solutions:**

1. Reduce parallelism:
   ```bash
   terraform apply -parallelism=1
   ```

2. Wait and retry:
   ```bash
   sleep 60 && terraform apply
   ```

3. Split into smaller batches:
   ```bash
   terraform apply -target=module.monitors
   sleep 60
   terraform apply -target=module.statuspages
   ```

### Error: Timeout During Apply

```
Error: context deadline exceeded
```

**Solutions:**

1. Increase timeout (if supported)
2. Apply resources individually
3. Check network connectivity

## Quick Reference

| Deployment Size | Resources | Parallelism | Daily API Budget |
|-----------------|-----------|-------------|------------------|
| Small | < 50 | 10 (default) | ~100 calls |
| Medium | 50-200 | 5 | ~400 calls |
| Large | 200-500 | 2-3 | ~1000 calls |
| Enterprise | 500+ | 1-2 | Contact Hyperping |

## Getting Help

- [Rate Limit Support](https://hyperping.io/support) - Request limit increases
- [Troubleshooting Guide](../TROUBLESHOOTING.md) - Common issues
- [GitHub Issues](https://github.com/develeap/terraform-provider-hyperping/issues)
