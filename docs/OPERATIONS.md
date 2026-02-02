# Production Operations Guide

**Last Updated:** 2026-01-29

This guide provides production-ready guidance for deploying and operating the Hyperping Terraform Provider at scale.

---

## Table of Contents

1. [Overview](#overview)
2. [Common Issues & Solutions](#common-issues--solutions)
3. [Troubleshooting](#troubleshooting)
4. [Rate Limiting](#rate-limiting)
5. [Large-Scale Deployments](#large-scale-deployments)
6. [Performance Characteristics](#performance-characteristics)
7. [Monitoring Best Practices](#monitoring-best-practices)
8. [Debugging Tips](#debugging-tips)

---

## Overview

### Key Components

- **Provider**: Configuration and authentication
- **Resources**: `hyperping_monitor`, `hyperping_incident`, `hyperping_incident_update`, `hyperping_maintenance`, `hyperping_healthcheck`, `hyperping_outage`
- **Data Sources**: Read operations for all resource types
- **API Client**: HTTP client with retry logic, rate limiting, and error handling

### Dependencies

- **Hyperping API**: https://api.hyperping.io
- **Terraform**: v1.0+ (tested with 1.6, 1.7, 1.8)
- **Go Runtime**: 1.21+ (for provider binary)

### Expected Performance

Most operations complete in under 2 seconds. The provider includes:
- Automatic retry logic for transient failures
- Rate limit handling with exponential backoff
- Request timeouts (30s default)
- Clear error messages for debugging

---

## Common Issues & Solutions

### 1. Authentication Failures

**Symptoms:**
```
Error: Failed to create monitor
│ API error (status 401): Unauthorized
```

**Causes:**
- Invalid or expired API key
- Missing `HYPERPING_API_KEY` environment variable
- API key format incorrect (must start with `sk_`)

**Solution:**
1. Verify API key in Hyperping dashboard
2. Check environment variable: `echo $HYPERPING_API_KEY`
3. Ensure key starts with `sk_` prefix
4. If compromised, rotate API key immediately

**Best Practices:**
- Use secret management (AWS Secrets Manager, HashiCorp Vault)
- Rotate keys regularly
- Never commit keys to version control

---

### 2. Rate Limiting

**Symptoms:**
```
Error: API error (status 429): Too Many Requests - retry after 60 seconds
```

**Causes:**
- Too many concurrent operations
- Rapid apply/destroy cycles
- Large-scale deployments exceeding API limits

**Solution:**
1. **Immediate**: Wait (retry is automatic)
2. **Reduce parallelism**:
   ```bash
   terraform apply -parallelism=5
   ```
3. **Long-term**: Batch operations, use staged rollouts

**Prevention:**
- Use `-parallelism=5` for deployments > 50 resources
- Implement gradual rollouts for large changes
- Enable debug logging to monitor rate limits

**Monitoring:**
```bash
export TF_LOG=DEBUG
terraform apply 2>&1 | grep -i "retry after"
```

---

### 3. Resource Drift

**Symptoms:**
```
# terraform plan shows changes even though config unchanged
```

**Causes:**
- Resources modified outside Terraform (via Hyperping UI/API)
- API returns different values than submitted
- Computed fields changed by Hyperping

**Solution:**
1. Review drift: `terraform plan`
2. Accept drift: `terraform apply` to reconcile
3. Investigate: Check Hyperping audit logs
4. Prevent: Use `ignore_changes` for computed fields

**Example:**
```hcl
resource "hyperping_monitor" "api" {
  name = "API Monitor"
  url  = "https://api.example.com"

  lifecycle {
    ignore_changes = [
      updated_at,  # Ignore timestamps
      down,        # Ignore status changes
    ]
  }
}
```

---

### 4. Import Failures

**Symptoms:**
```
Error: Cannot import non-existent remote object
```

**Causes:**
- Resource UUID doesn't exist in Hyperping
- API key doesn't have access to resource
- Typo in resource UUID

**Solution:**
1. Verify UUID exists:
   ```bash
   curl -H "Authorization: Bearer $HYPERPING_API_KEY" \
        https://api.hyperping.io/v1/monitors/{uuid}
   ```
2. Check API key permissions
3. Use correct import syntax:
   ```bash
   terraform import hyperping_monitor.api mon_abc123
   ```

---

### 5. Network Timeouts

**Symptoms:**
```
Error: context deadline exceeded
```

**Causes:**
- Network connectivity issues
- Hyperping API slow to respond
- Request timeout too short (default: 30s)

**Solution:**
1. Check connectivity: `curl https://api.hyperping.io`
2. Terraform automatically retries with exponential backoff
3. For persistent issues, check Hyperping status page

---

### 6. State Lock Conflicts

**Symptoms:**
```
Error: Error acquiring the state lock
```

**Causes:**
- Concurrent Terraform operations
- Stale lock from interrupted operation
- Remote backend issue

**Solution:**
1. Wait - another operation may be in progress
2. Force unlock (use carefully):
   ```bash
   terraform force-unlock LOCK_ID
   ```
3. Verify backend accessibility (S3/Consul/etc.)

---

## Troubleshooting

### Step 1: Enable Debug Logging

```bash
export TF_LOG=DEBUG
export TF_LOG_PATH=terraform-debug.log
terraform apply
```

**Look for:**
- API request/response details
- Retry attempts and backoff timing
- Error messages
- Request duration (duration_ms)

### Step 2: Verify API Connectivity

```bash
# Test authentication
curl -H "Authorization: Bearer $HYPERPING_API_KEY" \
     https://api.hyperping.io/v1/monitors

# Check API status
curl https://api.hyperping.io/
```

### Step 3: Check Provider Version

```bash
terraform version
# Should show: provider registry.terraform.io/develeap/hyperping vX.Y.Z
```

### Step 4: Validate Configuration

```bash
terraform validate
terraform fmt -check
```

### Step 5: Test with Minimal Config

```hcl
provider "hyperping" {
  # api_key from HYPERPING_API_KEY env var
}

resource "hyperping_monitor" "test" {
  name = "Test Monitor"
  url  = "https://httpstat.us/200"
}
```

```bash
terraform init
terraform plan
```

### Step 6: Check State

```bash
# View current state
terraform show

# List resources
terraform state list

# Inspect specific resource
terraform state show hyperping_monitor.api
```

---

## Rate Limiting

### Understanding Rate Limits

The provider implements intelligent rate limit handling:
- **Retry strategy**: 3 retries with exponential backoff
- **Backoff timing**: 1s → 2s → 4s → 8s → 16s → 30s (max)
- **Retry-After header**: Automatically respected

### Monitoring Rate Limits

```bash
export TF_LOG=DEBUG
terraform apply 2>&1 | grep "429"
```

Look for:
```
API error (status 429): Too Many Requests - retry after 60 seconds
```

### Avoiding Rate Limits

#### 1. Reduce Parallelism

```bash
# Default: 10
terraform apply -parallelism=5

# For very large deployments (500+ resources)
terraform apply -parallelism=1
```

#### 2. Batch Operations

Use targeted applies:
```bash
# Create resources in batches
terraform apply -target=module.monitoring[0]
terraform apply -target=module.monitoring[1]
```

#### 3. Staged Rollouts

Use workspaces:
```bash
# Deploy to dev first
terraform workspace select dev
terraform apply

# Then production
terraform workspace select production
terraform apply -parallelism=5
```

### Automatic Recovery

The provider handles rate limits automatically:
1. Detects 429 response
2. Parses Retry-After header
3. Waits specified duration (up to 30s max)
4. Retries request (max 3 retries)
5. Logs retry attempt (visible in DEBUG logs)

**No manual intervention required** for transient rate limits.

---

## Large-Scale Deployments

### Scale Definitions

- **Small**: < 10 resources
- **Medium**: 10-100 resources
- **Large**: 100-500 resources
- **Very Large**: 500+ resources

### Best Practices by Scale

#### Small Deployments (< 10 resources)

✅ Standard workflow:
```bash
terraform apply
```

No special configuration needed.

#### Medium Deployments (10-100 resources)

✅ Monitor for rate limits:
```bash
export TF_LOG=DEBUG
```

✅ Use remote state:
```hcl
terraform {
  backend "s3" {
    bucket = "terraform-state"
    key    = "hyperping/terraform.tfstate"
  }
}
```

#### Large Deployments (100-500 resources)

⚠️ **Required: Reduce parallelism**
```bash
terraform apply -parallelism=5
```

⚠️ **Required: Staged rollouts**
```bash
# Option 1: Use -target
terraform apply -target=module.monitors_batch_1
terraform apply -target=module.monitors_batch_2

# Option 2: Use workspaces
terraform workspace select production-1
terraform apply
```

✅ **Use modules for organization**:
```hcl
module "production_monitors" {
  source   = "./modules/monitors"
  monitors = var.production_monitors
}
```

#### Very Large Deployments (500+ resources)

⚠️ **Special considerations:**
- Use `-parallelism=1` to minimize rate limits
- Consider breaking into multiple Terraform workspaces
- Use Terraform Cloud/Enterprise for better state management
- Monitor memory usage (state grows with resource count)

### Memory Considerations

| Resource Count | Est. State Size | Memory Usage |
|---------------|-----------------|--------------|
| 100           | ~500 KB        | < 100 MB     |
| 500           | ~2.5 MB        | < 500 MB     |
| 1000          | ~5 MB          | < 1 GB       |

**Recommendation:** Use remote state backend for deployments > 100 resources.

---

## Performance Characteristics

### Typical Operation Duration

| Operation                | Duration    | Max Retries | Timeout |
|-------------------------|-------------|-------------|---------|
| Create Monitor          | 0.5-1.5s   | 3           | 30s     |
| Read Monitor            | 0.2-0.6s   | 3           | 30s     |
| Update Monitor          | 0.5-1.5s   | 3           | 30s     |
| Delete Monitor          | 0.3-0.8s   | 3           | 30s     |
| List Monitors (10)      | 0.4-1.0s   | 3           | 30s     |
| List Monitors (100)     | 1.0-3.0s   | 3           | 30s     |

### Retry Timing

**Exponential backoff:**
- Attempt 1: 1 second
- Attempt 2: 2 seconds
- Attempt 3: 4 seconds
- Attempt 4: 8 seconds
- Attempt 5: 16 seconds
- Attempt 6+: 30 seconds (capped)

**Total retry budget:** Up to ~64 seconds for max retries

### Throughput Recommendations

**With parallelism=10** (theoretical max):
- Creates: ~10-20/second
- Reads: ~20-40/second

**Recommended throttling** (avoid rate limits):
- `parallelism=5` for 100+ resources
- `parallelism=1` for 500+ resources

### Network Latency

Provider latency = API latency + network RTT

**Recommendation:** Run Terraform in same region as Hyperping API for best performance.

---

## Monitoring Best Practices

### Key Metrics to Track

#### 1. Apply Success Rate
- **What**: Percentage of successful `terraform apply` runs
- **Target**: > 95%
- **Alert on**: < 90% over 24 hours

#### 2. API Error Rate
- **What**: Count of API errors by status code
- **Target**: < 5% error rate
- **Alert on**: > 10% for 15+ minutes

#### 3. Rate Limit Events
- **What**: Count of 429 responses
- **Target**: < 5 per hour
- **Alert on**: > 10 per hour (indicates need to reduce parallelism)

### Logging Recommendations

#### Production
- Use `TF_LOG=INFO` or `TF_LOG=WARN`
- Aggregate to central system (CloudWatch, Datadog, Splunk)
- Set retention period (30-90 days)
- Parse for metrics

**Avoid:**
- `TF_LOG=DEBUG` in production (too verbose, may contain sensitive data)
- Local-only logs (lost on instance termination)

#### Development
- Use `TF_LOG=DEBUG` for troubleshooting
- Use `TF_LOG_PATH` to save logs

**Example:**
```bash
export TF_LOG=DEBUG
export TF_LOG_PATH=/tmp/terraform-$(date +%Y%m%d-%H%M%S).log
terraform apply
```

---

## Debugging Tips

### 1. Enable Verbose Logging

```bash
export TF_LOG=DEBUG
export TF_LOG_PATH=terraform-debug.log
terraform apply
```

**Look for:**
- `sending API request` - Request details
- `received API response` - Response status, duration
- `retrying request` - Retry attempts

### 2. Use Terraform Console

Test expressions and data sources:
```bash
terraform console

# Test data source
> data.hyperping_monitors.all.monitors[0].name

# Test variable
> var.monitor_url
```

### 3. Validate State vs. Reality

```bash
# Compare state to actual resources
terraform plan

# Refresh state from API
terraform refresh

# Show current state
terraform show
```

### 4. Test API Directly

```bash
# Get monitor
curl -H "Authorization: Bearer $HYPERPING_API_KEY" \
     https://api.hyperping.io/v1/monitors/mon_abc123

# Create monitor
curl -X POST \
     -H "Authorization: Bearer $HYPERPING_API_KEY" \
     -H "Content-Type: application/json" \
     -d '{"name": "Test", "url": "https://example.com"}' \
     https://api.hyperping.io/v1/monitors
```

### 5. Isolate Resource Issues

```bash
# Plan only specific resource
terraform plan -target=hyperping_monitor.api

# Apply only specific resource
terraform apply -target=hyperping_monitor.api

# Destroy only specific resource
terraform destroy -target=hyperping_monitor.api
```

### 6. Check Provider Version

```bash
# Current version
terraform version

# Upgrade to latest
terraform init -upgrade

# Check for known issues
# https://github.com/develeap/terraform-provider-hyperping/issues
```

### 7. Export State for Analysis

```bash
# Export to JSON
terraform show -json > state.json

# Extract specific resource
terraform state show hyperping_monitor.api > monitor.txt
```

### 8. Minimal Test Config

Create `test.tf`:
```hcl
terraform {
  required_providers {
    hyperping = {
      source = "develeap/hyperping"
    }
  }
}

provider "hyperping" {}

resource "hyperping_monitor" "test" {
  name = "Debug Test"
  url  = "https://httpstat.us/200"
}
```

Test:
```bash
terraform init
terraform plan
terraform apply -auto-approve
terraform destroy -auto-approve
```

---

## Quick Reference

### Useful Commands

```bash
# Provider info
terraform version
terraform providers

# State management
terraform state list
terraform state show <resource>
terraform state pull > state.backup.json

# Validation
terraform validate
terraform fmt -check -recursive

# Planning
terraform plan -out=tfplan
terraform show tfplan

# Applying
terraform apply -auto-approve
terraform apply -parallelism=5
terraform apply -target=hyperping_monitor.api

# Debugging
export TF_LOG=DEBUG
export TF_LOG_PATH=terraform.log

# Import
terraform import hyperping_monitor.api mon_abc123
```

### Environment Variables

| Variable               | Purpose           | Example                      |
|-----------------------|-------------------|------------------------------|
| `HYPERPING_API_KEY`   | Authentication    | `sk_abc123...`              |
| `TF_LOG`              | Logging level     | `DEBUG`, `INFO`, `WARN`     |
| `TF_LOG_PATH`         | Log file path     | `/tmp/terraform.log`        |
| `TF_APPEND_USER_AGENT`| Custom user-agent | `my-automation/v1.0`        |

---

## Support

- **Issues**: https://github.com/develeap/terraform-provider-hyperping/issues
- **Documentation**: https://registry.terraform.io/providers/develeap/hyperping/latest/docs
- **Hyperping Support**: https://hyperping.io/support

---

*This guide is maintained by the community. Contributions welcome via pull requests.*
