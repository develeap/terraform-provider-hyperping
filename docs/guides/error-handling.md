# Error Handling Guide

This comprehensive guide helps you understand, troubleshoot, and resolve errors when using the Hyperping Terraform provider.

## Table of Contents

1. [Introduction](#introduction)
2. [Common Errors](#common-errors)
3. [Error Types Detailed](#error-types-detailed)
4. [Troubleshooting Guide](#troubleshooting-guide)
5. [Best Practices](#best-practices)
6. [Examples](#examples)

---

## Introduction

The Hyperping Terraform provider implements enhanced error handling to help you quickly identify and resolve issues. When an error occurs, the provider delivers actionable troubleshooting steps tailored to the specific error type and context.

### What Errors Mean in Terraform Provider Context

Terraform providers interact with external APIs (in this case, Hyperping's API) to manage infrastructure resources. Errors can occur at multiple stages:

- **Plan Phase**: Configuration validation and state checking
- **Apply Phase**: Resource creation, updates, or deletions
- **Destroy Phase**: Resource cleanup
- **Import Phase**: Bringing existing resources under Terraform management

Each phase can encounter different types of errors, and understanding the context helps in resolution.

### How Enhanced Errors Help Users

The Hyperping provider's enhanced error system provides:

1. **Context-Specific Guidance**: Tailored troubleshooting steps based on the error type
2. **Resource Links**: Direct links to the Hyperping dashboard for verification
3. **Actionable Steps**: Clear instructions for resolution, not just error messages
4. **Rate Limit Handling**: Automatic retry guidance with specific wait times
5. **API Key Verification**: Step-by-step authentication troubleshooting

### Overview of Error Types

The provider categorizes errors into five main types:

| Error Type | HTTP Status | Description | Common Causes |
|------------|-------------|-------------|---------------|
| **Not Found** | 404 | Resource doesn't exist | Deleted outside Terraform, wrong ID |
| **Authentication** | 401/403 | Invalid or insufficient permissions | Wrong API key, missing permissions |
| **Rate Limit** | 429 | Too many requests | Excessive parallelism, rapid changes |
| **Server Error** | 500-504 | API service issue | Hyperping service disruption |
| **Validation** | 400/422 | Invalid request data | Wrong field values, missing requirements |

---

## Common Errors

This section covers the top 5 errors users encounter and provides quick solutions.

### 1. Resource Not Found (404)

**What It Looks Like:**

```
Error: Failed to Read Monitor

Unable to read Monitor (ID: mon_abc123), got error: API error (status 404): resource not found

Troubleshooting:
- Verify the resource still exists in Hyperping
- Check your API key has read permissions
- Check network connectivity
- Review the Hyperping dashboard: https://app.hyperping.io
```

**Quick Solution:**

The resource was likely deleted outside of Terraform. Options:

1. **If deleted intentionally**: Remove from Terraform state
   ```bash
   terraform state rm hyperping_monitor.example
   ```

2. **If deleted by mistake**: Recreate by removing and re-applying
   ```bash
   terraform apply -replace=hyperping_monitor.example
   ```

3. **If ID is wrong**: Check the Hyperping dashboard and update your configuration

**Prevention:**
- Use `lifecycle { prevent_destroy = true }` for critical resources
- Implement proper access controls in Hyperping
- Coordinate manual changes with your team

### 2. Unauthorized / Invalid API Key (401/403)

**What It Looks Like:**

```
Error: Failed to Create Monitor

Unable to create Monitor, got error: API error (status 401): unauthorized: invalid or missing API key

Troubleshooting:
- Verify your API key has create permissions
- Check that all required fields are provided
- Review the Hyperping dashboard: https://app.hyperping.io
```

**Quick Solution:**

Verify your API key is correctly configured:

```bash
# Check if API key is set
echo $HYPERPING_API_KEY

# Test API key validity
curl -H "Authorization: Bearer $HYPERPING_API_KEY" \
  https://api.hyperping.io/v1/monitors

# Set API key if missing
export HYPERPING_API_KEY="your-api-key-here"
```

**Common Causes:**
- API key not set in environment
- API key has wrong permissions
- API key was revoked or regenerated
- Using wrong API key for different environments

**Prevention:**
- Store API keys in secure secret management (AWS Secrets Manager, HashiCorp Vault)
- Use separate API keys for dev/staging/prod
- Regularly rotate API keys and update configuration
- Document API key requirements for your team

### 3. Rate Limit Exceeded (429)

**What It Looks Like:**

```
Error: Failed to Read Monitor

Unable to read Monitor (ID: mon_def456), got error: API error (status 429): rate limit exceeded - retry after 60 seconds

Troubleshooting:
- Wait 60 seconds before retrying
- Reduce Terraform parallelism: terraform apply -parallelism=1
- Check API service status: https://status.hyperping.app
```

**Quick Solution:**

Wait the specified time and reduce parallelism:

```bash
# Wait for rate limit to reset (check error message for time)
sleep 60

# Re-run with reduced parallelism
terraform apply -parallelism=2

# For large configurations, use even lower parallelism
terraform apply -parallelism=1
```

**Prevention:**
- Set default parallelism in your workflow
- Implement delays between bulk operations
- Use `terraform apply -target` for specific resources when needed
- Monitor your API usage in Hyperping dashboard

### 4. Server Error (500-504)

**What It Looks Like:**

```
Error: Failed to Update Incident

Unable to update Incident (ID: inc_ghi789), got error: API error (status 500): internal server error

Troubleshooting:
- Verify the resource still exists in Hyperping
- Check your API key has update permissions
- Verify the update values are valid
- Review the Hyperping dashboard: https://app.hyperping.io
```

**Quick Solution:**

Server errors are typically temporary. Try these steps:

1. **Check Hyperping Status**: Visit https://status.hyperping.app
2. **Wait and Retry**: Wait 1-2 minutes and try again
3. **Verify State**: Run `terraform plan` to check if the operation completed
4. **Contact Support**: If persistent, contact Hyperping support with the error details

**When to Escalate:**
- Error persists for more than 15 minutes
- Multiple different operations are failing
- Hyperping status page shows ongoing incident

### 5. Validation Error (400/422)

**What It Looks Like:**

```
Error: Failed to Create Monitor

Unable to create Monitor, got error: API error (status 400): validation error - 2 validation errors

Troubleshooting:
- Verify your API key has create permissions
- Check that all required fields are provided
- Review the Hyperping dashboard: https://app.hyperping.io
```

**Quick Solution:**

Check your Terraform configuration for invalid values:

```hcl
# Common validation issues:

# 1. Invalid frequency
resource "hyperping_monitor" "bad" {
  name            = "API Monitor"
  url             = "https://api.example.com"
  protocol        = "http"
  check_frequency = 45  # Invalid! Must be 10, 20, 30, 60, 120, 180, 300, 600, 1800, or 3600
}

# Fix:
resource "hyperping_monitor" "good" {
  name            = "API Monitor"
  url             = "https://api.example.com"
  protocol        = "http"
  check_frequency = 60  # Valid
}

# 2. Invalid region
resource "hyperping_monitor" "bad" {
  regions = ["us-east-1"]  # Invalid!
}

# Fix:
resource "hyperping_monitor" "good" {
  regions = ["virginia", "london"]  # Valid regions
}

# 3. Missing required fields
resource "hyperping_monitor" "bad" {
  name = "Monitor"  # Missing url and protocol!
}

# Fix:
resource "hyperping_monitor" "good" {
  name     = "Monitor"
  url      = "https://api.example.com"
  protocol = "http"
}
```

**Prevention:**
- Use `terraform validate` before applying
- Reference the [Validation Guide](validation.md) for all allowed values
- Use IDE with Terraform HCL support for syntax highlighting

---

## Error Types Detailed

This section provides in-depth information about each error type.

### 404 Not Found Error

**What It Means:**

The resource you're trying to access doesn't exist in Hyperping. This can happen during read, update, or delete operations.

**Common Causes:**

1. **Resource Deleted Outside Terraform**: Someone deleted the resource via Hyperping dashboard
2. **Wrong Resource ID**: The ID in your Terraform state doesn't match reality
3. **State Drift**: Terraform state is out of sync with Hyperping
4. **Import Issue**: Resource wasn't properly imported into Terraform

**Detailed Solutions:**

**Scenario 1: Resource was intentionally deleted**

If you meant to delete the resource, clean up Terraform state:

```bash
# Remove from state
terraform state rm hyperping_monitor.example

# Or refresh state to detect deletion
terraform refresh
terraform apply  # Will detect resource is missing
```

**Scenario 2: Resource exists but wrong ID**

Verify the resource ID in Hyperping dashboard:

1. Log into https://app.hyperping.io
2. Navigate to the resource (Monitors, Incidents, etc.)
3. Click on the resource to view details
4. Copy the correct UUID
5. Import into Terraform:

```bash
terraform import hyperping_monitor.example mon_correct_uuid
```

**Scenario 3: State drift**

Sync Terraform state with Hyperping:

```bash
# Refresh state to detect changes
terraform refresh

# Review what changed
terraform plan

# Apply to sync
terraform apply
```

**Dashboard Links:**

- Monitors: https://app.hyperping.io/monitors
- Incidents: https://app.hyperping.io/incidents
- Maintenance: https://app.hyperping.io/maintenance
- Status Pages: https://app.hyperping.io/statuspages

### 401/403 Authentication Error

**What It Means:**

Your API key is either invalid, missing, or doesn't have sufficient permissions for the requested operation.

**HTTP 401 (Unauthorized):**
- API key is invalid or malformed
- API key is missing from the request
- API key was revoked

**HTTP 403 (Forbidden):**
- API key is valid but lacks necessary permissions
- Resource belongs to different account
- Operation not allowed for your plan level

**API Key Verification Steps:**

**Step 1: Verify API key is set**

```bash
# Check environment variable
echo $HYPERPING_API_KEY

# Should output something like: hp_abc123def456...
# If empty, set it:
export HYPERPING_API_KEY="your-api-key-here"
```

**Step 2: Test API key with curl**

```bash
# Test authentication
curl -v -H "Authorization: Bearer $HYPERPING_API_KEY" \
  https://api.hyperping.io/v1/monitors

# Successful response: HTTP 200 with JSON data
# Failed response: HTTP 401 or 403
```

**Step 3: Verify API key in Hyperping dashboard**

1. Log into https://app.hyperping.io
2. Go to Settings → API Keys
3. Verify your API key exists and is active
4. Check permissions for the API key
5. Regenerate if necessary (update Terraform config after!)

**Step 4: Check permissions**

Ensure your API key has permissions for:
- **Read**: List and read resources
- **Write**: Create and update resources
- **Delete**: Delete resources

**Step 5: Environment-specific keys**

```bash
# Use different keys for different environments
export HYPERPING_API_KEY_DEV="hp_dev_key"
export HYPERPING_API_KEY_PROD="hp_prod_key"

# In your Terraform configuration
# Development
provider "hyperping" {
  api_key = var.hyperping_api_key_dev
}

# Production
provider "hyperping" {
  api_key = var.hyperping_api_key_prod
}
```

**Common Mistakes:**

- Using production API key in development (or vice versa)
- API key exposed in code (use environment variables!)
- API key not set in CI/CD environment
- API key regenerated but not updated in all environments

### 429 Rate Limit Error

**What Triggers It:**

Hyperping's API has rate limits to ensure fair usage and system stability. Rate limits typically trigger when:

1. **Too many requests per second**: Exceeds per-second limit
2. **Bulk operations**: Creating/updating many resources at once
3. **Terraform parallelism**: Default parallelism (10) may be too high
4. **Repeated refreshes**: Running `terraform plan` too frequently

**How to Handle It:**

The error message includes a `Retry-After` value indicating how long to wait:

```
API error (status 429): rate limit exceeded - retry after 60 seconds
```

**Solution 1: Wait and retry**

```bash
# Wait the specified time
sleep 60

# Retry the operation
terraform apply
```

**Solution 2: Reduce parallelism**

```bash
# Reduce concurrent operations
terraform apply -parallelism=2

# For very large configurations
terraform apply -parallelism=1
```

**Solution 3: Implement retry logic in CI/CD**

```bash
#!/bin/bash
# retry-terraform.sh

MAX_RETRIES=3
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
  terraform apply -auto-approve
  EXIT_CODE=$?

  if [ $EXIT_CODE -eq 0 ]; then
    echo "Apply succeeded"
    exit 0
  fi

  # Check if rate limited
  if grep -q "rate limit exceeded" terraform.log; then
    RETRY_COUNT=$((RETRY_COUNT + 1))
    echo "Rate limited, waiting 60 seconds... (attempt $RETRY_COUNT/$MAX_RETRIES)"
    sleep 60
  else
    echo "Apply failed with non-rate-limit error"
    exit $EXIT_CODE
  fi
done

echo "Max retries exceeded"
exit 1
```

**Best Practices:**

- Set reasonable parallelism in production: `terraform apply -parallelism=3`
- Implement exponential backoff in automation
- Batch large changes into smaller groups
- Monitor rate limit usage in Hyperping dashboard
- Use `-target` flag to limit operations when appropriate

### 500 Server Error

**What It Means:**

The Hyperping API encountered an internal error processing your request. These are typically temporary and not caused by your configuration.

**HTTP Status Codes:**

- **500 Internal Server Error**: General server-side error
- **502 Bad Gateway**: Upstream service unavailable
- **503 Service Unavailable**: Service temporarily down
- **504 Gateway Timeout**: Request took too long

**Troubleshooting Steps:**

**Step 1: Check Hyperping status**

Visit https://status.hyperping.app to check for:
- Ongoing incidents
- Scheduled maintenance
- Service degradation

**Step 2: Verify the operation completed**

Sometimes the operation succeeds despite the error:

```bash
# Check current state
terraform plan

# If plan shows no changes, the operation completed
# If plan shows changes, retry the operation
```

**Step 3: Wait and retry**

```bash
# Wait 1-2 minutes
sleep 120

# Retry
terraform apply
```

**Step 4: Check if it's a specific resource**

```bash
# If a specific resource is failing, try others
terraform apply -target=hyperping_monitor.working

# Then retry the failing resource
terraform apply -target=hyperping_monitor.failing
```

**When to Escalate:**

Contact Hyperping support if:
- Error persists for more than 15 minutes
- Status page shows no incidents
- Multiple different resources are affected
- Error happens consistently for the same resource

**What to Provide to Support:**

- Full error message
- Timestamp of the error
- Resource type and operation (create/read/update/delete)
- Your account ID or email
- Whether the error is consistent or intermittent

### 400 Validation Error

**What It Means:**

Your request data doesn't meet Hyperping's validation requirements. The API validates all fields before processing.

**Common Validation Errors:**

**1. Invalid check_frequency**

```hcl
# Invalid
resource "hyperping_monitor" "example" {
  check_frequency = 45  # Not in allowed list
}

# Valid values: 10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600
resource "hyperping_monitor" "example" {
  check_frequency = 60  # Valid
}
```

**2. Invalid regions**

```hcl
# Invalid
resource "hyperping_monitor" "example" {
  regions = ["us-east-1", "eu-west-1"]  # Wrong format
}

# Valid regions: london, frankfurt, singapore, sydney, virginia, oregon, saopaulo, tokyo, bahrain
resource "hyperping_monitor" "example" {
  regions = ["virginia", "london"]  # Valid
}
```

**3. Invalid protocol/URL combination**

```hcl
# Invalid - HTTP URL with port protocol
resource "hyperping_monitor" "example" {
  protocol = "port"
  url      = "https://example.com"  # Should not have https://
}

# Valid
resource "hyperping_monitor" "example" {
  protocol = "port"
  url      = "example.com"
  port     = 443
}
```

**4. Missing required fields**

```hcl
# Invalid - missing required fields
resource "hyperping_monitor" "example" {
  name = "My Monitor"
  # Missing: url, protocol
}

# Valid
resource "hyperping_monitor" "example" {
  name     = "My Monitor"
  url      = "https://api.example.com"
  protocol = "http"
}
```

**5. Invalid datetime format**

```hcl
# Invalid - wrong datetime format
resource "hyperping_maintenance" "example" {
  scheduled_start = "2024-01-15 10:00:00"  # Wrong format
}

# Valid - ISO 8601 format with timezone
resource "hyperping_maintenance" "example" {
  scheduled_start = "2024-01-15T10:00:00Z"  # Correct
}
```

**Field-Specific Guidance:**

The provider validates many fields client-side before sending to the API. Use `terraform validate` to catch these early:

```bash
terraform validate

# Example output:
# Error: Invalid value for "check_frequency"
#   on main.tf line 5:
#   5:   check_frequency = 45
#
# Allowed values: 10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600
```

---

## Troubleshooting Guide

This section provides step-by-step troubleshooting for each error type.

### Troubleshooting Decision Tree

```
Error occurred
    |
    ├─ Is it a 404 Not Found?
    │  ├─ Yes → Check Hyperping dashboard
    │  │       → Verify resource exists
    │  │       → Check resource ID
    │  │       → Consider state import
    │  │
    │  └─ No → Continue
    │
    ├─ Is it a 401/403 Auth Error?
    │  ├─ Yes → Verify API key is set
    │  │       → Test API key with curl
    │  │       → Check API key permissions
    │  │       → Verify correct environment
    │  │
    │  └─ No → Continue
    │
    ├─ Is it a 429 Rate Limit?
    │  ├─ Yes → Check retry-after time
    │  │       → Reduce parallelism
    │  │       → Wait and retry
    │  │       → Implement backoff
    │  │
    │  └─ No → Continue
    │
    ├─ Is it a 500 Server Error?
    │  ├─ Yes → Check status.hyperping.app
    │  │       → Wait 1-2 minutes
    │  │       → Retry operation
    │  │       → Contact support if persistent
    │  │
    │  └─ No → Continue
    │
    └─ Is it a 400 Validation Error?
       └─ Yes → Run terraform validate
              → Check field requirements
              → Review documentation
              → Fix configuration
```

### Verification Commands

Use these commands to verify your setup:

```bash
# 1. Verify API key is set
echo $HYPERPING_API_KEY

# 2. Test API connectivity
curl -H "Authorization: Bearer $HYPERPING_API_KEY" \
  https://api.hyperping.io/v1/monitors

# 3. Validate Terraform configuration
terraform validate

# 4. Check Terraform state
terraform state list

# 5. Preview changes
terraform plan

# 6. Refresh state from Hyperping
terraform refresh

# 7. Show specific resource
terraform state show hyperping_monitor.example
```

---

## Best Practices

Follow these practices to minimize errors and improve reliability.

### 1. Error Prevention Strategies

**Use Terraform Validate**

Always validate before applying:

```bash
# Add to your workflow
terraform fmt -check
terraform validate
terraform plan
# Only then:
terraform apply
```

**Implement Pre-Commit Hooks**

```bash
# .git/hooks/pre-commit
#!/bin/bash
terraform fmt -check -recursive && terraform validate
```

**Use Terraform Workspaces**

Separate environments to prevent accidental changes:

```bash
# Development workspace
terraform workspace new dev
terraform workspace select dev

# Production workspace
terraform workspace new prod
terraform workspace select prod
```

### 2. Proper API Key Management

**Use Environment Variables**

```bash
# Never hardcode API keys
# Bad:
provider "hyperping" {
  api_key = "hp_abc123..."  # Don't do this!
}

# Good:
provider "hyperping" {
  api_key = var.hyperping_api_key
}

# Set in environment
export HYPERPING_API_KEY="hp_abc123..."
```

**Use Secret Management**

```bash
# AWS Secrets Manager
export HYPERPING_API_KEY=$(aws secretsmanager get-secret-value \
  --secret-id hyperping-api-key \
  --query SecretString \
  --output text)

# HashiCorp Vault
export HYPERPING_API_KEY=$(vault kv get \
  -field=api_key secret/hyperping)
```

**Rotate Keys Regularly**

1. Generate new API key in Hyperping dashboard
2. Update secret in secret management system
3. Test new key in non-production environment
4. Update production environment
5. Revoke old key

### 3. Rate Limit Avoidance

**Set Parallelism Appropriately**

```bash
# For small configs (< 10 resources)
terraform apply -parallelism=5

# For medium configs (10-50 resources)
terraform apply -parallelism=3

# For large configs (> 50 resources)
terraform apply -parallelism=1
```

**Implement Delays in Automation**

```bash
# In CI/CD pipeline
terraform apply -auto-approve
sleep 30  # Cool-down period
terraform destroy -auto-approve
```

**Use Targeted Applies**

```bash
# Apply specific resources
terraform apply -target=hyperping_monitor.critical

# Apply by module
terraform apply -target=module.monitors
```

### 4. State Management Tips

**Enable Remote State**

```hcl
terraform {
  backend "s3" {
    bucket = "my-terraform-state"
    key    = "hyperping/terraform.tfstate"
    region = "us-east-1"
  }
}
```

**Use State Locking**

```hcl
terraform {
  backend "s3" {
    bucket         = "my-terraform-state"
    key            = "hyperping/terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "terraform-state-lock"
  }
}
```

**Regular State Backups**

```bash
# Backup state before major changes
terraform state pull > terraform.tfstate.backup

# Restore if needed
terraform state push terraform.tfstate.backup
```

### 5. Documentation and Communication

**Document Your Resources**

```hcl
resource "hyperping_monitor" "api" {
  name     = "[PROD] API Health Check"  # Environment prefix
  url      = "https://api.example.com/health"
  protocol = "http"

  # Document purpose in name or tags
  description = "Critical API health endpoint - DO NOT DELETE"
}
```

**Use Consistent Naming**

```hcl
# Pattern: [ENVIRONMENT]-[SERVICE]-[TYPE]
resource "hyperping_monitor" "prod_api_health" {
  name = "[PROD]-API-Health"
}

resource "hyperping_monitor" "prod_api_latency" {
  name = "[PROD]-API-Latency"
}
```

**Maintain Change Log**

Document why changes were made:

```bash
# Git commit messages
git commit -m "feat: add health check for new API endpoint

- Monitors GET /v2/health
- Checks from virginia and london regions
- 60 second frequency
- Relates to ticket INFRA-123"
```

---

## Examples

Real-world error scenarios and complete resolution walkthroughs.

### Example 1: 404 Error - Resource Deleted Outside Terraform

**Scenario:**

You run `terraform plan` and see:

```
Error: Failed to Read Monitor

Unable to read Monitor (ID: mon_abc123), got error: API error (status 404): resource not found

Troubleshooting:
- Verify the resource still exists in Hyperping
- Check your API key has read permissions
- Check network connectivity
- Review the Hyperping dashboard: https://app.hyperping.io
```

**Investigation:**

```bash
# 1. Check Hyperping dashboard
# Visit https://app.hyperping.io/monitors
# Search for mon_abc123
# Result: Monitor not found

# 2. Check who deleted it
# Review Hyperping audit logs
# Finding: Deleted manually by teammate yesterday

# 3. Check if we want to keep it in Terraform
# Decision: Resource no longer needed
```

**Resolution:**

```bash
# Option A: Remove from Terraform (if deletion was intentional)
terraform state rm hyperping_monitor.old_api

# Verify removal
terraform state list | grep old_api
# Should return nothing

# Option B: Recreate (if deletion was a mistake)
# The resource is already in config, so just apply
terraform apply
# This will recreate the monitor with a new ID
```

**Before/After:**

Before:
```
terraform plan
# Error: Resource mon_abc123 not found
```

After:
```
terraform plan
# No changes. Infrastructure is up-to-date.
```

### Example 2: 401 Error - Wrong API Key

**Scenario:**

Deploying to a new environment:

```
Error: Failed to Create Monitor

Unable to create Monitor, got error: API error (status 401): unauthorized: invalid or missing API key
```

**Investigation:**

```bash
# 1. Check if API key is set
echo $HYPERPING_API_KEY
# Output: (empty)

# Problem identified: API key not set in this environment
```

**Resolution:**

```bash
# 1. Get correct API key from secret manager
export HYPERPING_API_KEY=$(aws secretsmanager get-secret-value \
  --secret-id prod/hyperping/api-key \
  --query SecretString \
  --output text)

# 2. Verify API key works
curl -H "Authorization: Bearer $HYPERPING_API_KEY" \
  https://api.hyperping.io/v1/monitors
# Should return 200 OK with JSON

# 3. Retry Terraform operation
terraform apply
```

**Before/After:**

Before:
```
$ terraform apply
# Error: 401 Unauthorized
```

After:
```
$ export HYPERPING_API_KEY="..."
$ terraform apply
# Apply complete! Resources: 5 added, 0 changed, 0 destroyed.
```

**Lesson Learned:**

Always verify environment variables are set in new environments. Add to deployment checklist.

### Example 3: 429 Rate Limit - Bulk Monitor Creation

**Scenario:**

Creating 50 monitors at once:

```
Error: Failed to Create Monitor

Unable to create Monitor, got error: API error (status 429): rate limit exceeded - retry after 60 seconds
```

**Investigation:**

```bash
# 1. Check how many resources are being created
terraform plan | grep "will be created"
# Output: 50 resources will be created

# 2. Check current parallelism
# Default is 10, which is too high for this many creates
```

**Resolution:**

```bash
# 1. Reduce parallelism significantly
terraform apply -parallelism=2

# 2. If still rate limited, go to 1
terraform apply -parallelism=1

# 3. Alternative: Create in batches
terraform apply -target=hyperping_monitor.api_01
terraform apply -target=hyperping_monitor.api_02
# ... etc

# 4. Better: Use count with delays
# Update Terraform config to create in smaller batches
```

**Improved Configuration:**

```hcl
# Create monitors in smaller batches
resource "hyperping_monitor" "batch_1" {
  count = 10

  name     = "Monitor ${count.index + 1}"
  url      = "https://api${count.index + 1}.example.com"
  protocol = "http"
}

# Apply this batch first, then create batch_2, etc.
```

**Before/After:**

Before:
```
$ terraform apply
# Creates 50 monitors at once
# Error: Rate limit exceeded
```

After:
```
$ terraform apply -parallelism=2
# Successfully creates all 50 monitors
# Takes longer but completes successfully
```

### Example 4: 400 Validation - Invalid Frequency

**Scenario:**

```
Error: Failed to Create Monitor

Unable to create Monitor, got error: API error (status 400): validation error - 1 validation error

Troubleshooting:
- Verify your API key has create permissions
- Check that all required fields are provided
- Review the Hyperping dashboard: https://app.hyperping.io
```

**Configuration:**

```hcl
resource "hyperping_monitor" "api" {
  name            = "API Health"
  url             = "https://api.example.com"
  protocol        = "http"
  check_frequency = 45  # Invalid!
}
```

**Investigation:**

```bash
# 1. Run terraform validate
terraform validate
# May not catch this if it's an API-side validation

# 2. Review documentation
# Check allowed values for check_frequency
# Valid: 10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600
```

**Resolution:**

```hcl
# Fix configuration
resource "hyperping_monitor" "api" {
  name            = "API Health"
  url             = "https://api.example.com"
  protocol        = "http"
  check_frequency = 60  # Valid: 1 minute
}
```

```bash
# Apply fixed configuration
terraform validate
terraform plan
terraform apply
```

**Before/After:**

Before:
```hcl
check_frequency = 45  # Invalid
# Error: validation error
```

After:
```hcl
check_frequency = 60  # Valid
# Apply complete! Resources: 1 added, 0 changed, 0 destroyed.
```

---

## Additional Resources

- [Validation Guide](validation.md) - Complete field validation reference
- [Rate Limits Guide](rate-limits.md) - Detailed rate limiting information
- [Troubleshooting Guide](../TROUBLESHOOTING.md) - General troubleshooting
- [Hyperping API Documentation](https://hyperping.io/docs/api) - Official API docs
- [Hyperping Status Page](https://status.hyperping.app) - Service status

## Getting Help

If you encounter errors not covered in this guide:

1. Check the [GitHub Issues](https://github.com/develeap/terraform-provider-hyperping/issues)
2. Search for similar errors in closed issues
3. Open a new issue with:
   - Full error message
   - Terraform configuration (sanitized)
   - Steps to reproduce
   - Terraform and provider versions
4. Contact Hyperping support for API-related issues

---

**Last Updated:** 2026-02-12
**Provider Version:** 1.0.9+
