# Error Reference

This document provides a comprehensive reference for all errors that may occur when using the Hyperping Terraform Provider, along with solutions and troubleshooting steps.

## Table of Contents

- [Authentication Errors](#authentication-errors)
- [Rate Limit Errors](#rate-limit-errors)
- [Validation Errors](#validation-errors)
- [Not Found Errors](#not-found-errors)
- [Server Errors](#server-errors)
- [Network Errors](#network-errors)
- [Circuit Breaker Errors](#circuit-breaker-errors)

---

## Authentication Errors

### ERR-AUTH-001: Invalid API Key

**Error Message:**
```
❌ Authentication Failed

Your Hyperping API key is invalid or has expired.

Resource:  hyperping_monitor.prod_api
Operation: create

💡 Suggestions:
  • Verify your API key is correct (should start with 'sk_')
  • Check if the API key has been revoked in the Hyperping dashboard
  • Ensure HYPERPING_API_KEY environment variable is set correctly
  • Generate a new API key if needed

🔧 Try:
  $ echo $HYPERPING_API_KEY                    # Verify key is set
  $ terraform plan                             # Test with valid credentials

📚 Documentation:
  https://registry.terraform.io/providers/develeap/hyperping/latest/docs#authentication
  https://app.hyperping.io/settings/api        # Generate new API key
```

**Cause**: The provided API key is invalid, expired, or has been revoked.

**Solutions**:

1. **Verify API Key Format**
   ```bash
   echo $HYPERPING_API_KEY
   # Should start with 'sk_' and be a long alphanumeric string
   ```

2. **Generate New API Key**
   - Visit [Hyperping Dashboard](https://app.hyperping.io/settings/api)
   - Click "Generate New API Key"
   - Copy the key and set it in your environment
   ```bash
   export HYPERPING_API_KEY="sk_your_new_key_here"
   ```

3. **Set in Terraform**
   ```hcl
   provider "hyperping" {
     api_key = var.hyperping_api_key  # Or use environment variable
   }
   ```

**Related Errors**: 401 Unauthorized, 403 Forbidden

---

### ERR-AUTH-002: Missing API Key

**Error Message:**
```
❌ Authentication Failed

No API key provided. Please set the HYPERPING_API_KEY environment variable or configure the provider.
```

**Cause**: API key is not configured in provider or environment.

**Solutions**:

1. **Set Environment Variable**
   ```bash
   export HYPERPING_API_KEY="sk_your_key_here"
   ```

2. **Configure in Provider**
   ```hcl
   provider "hyperping" {
     api_key = "sk_your_key_here"  # Not recommended for production
   }
   ```

3. **Use Variable**
   ```hcl
   variable "hyperping_api_key" {
     type      = string
     sensitive = true
   }

   provider "hyperping" {
     api_key = var.hyperping_api_key
   }
   ```

---

## Rate Limit Errors

### ERR-RATE-001: Rate Limit Exceeded

**Error Message:**
```
❌ Rate Limit Exceeded

You've hit the Hyperping API rate limit. Your request will automatically retry.

Resource:  hyperping_monitor.staging_api
Operation: create

⏰ Auto-retry after: 23 seconds (automatically retrying with exponential backoff)

💡 Suggestions:
  • Reduce the number of parallel terraform operations
  • Use terraform apply with -parallelism=1 flag for serial execution
  • Consider upgrading your Hyperping plan for higher rate limits
  • Use bulk operations where possible instead of individual creates

🔧 Try:
  $ terraform apply -parallelism=1             # Reduce concurrent requests
  $ terraform apply -refresh=false             # Skip refresh to reduce API calls

📚 Documentation:
  https://github.com/develeap/terraform-provider-hyperping/tree/main/docs/guides/rate-limits.md
  https://api.hyperping.io/docs#rate-limits
```

**Cause**: Too many API requests in a short period.

**Solutions**:

1. **Reduce Parallelism**
   ```bash
   terraform apply -parallelism=1
   ```

2. **Skip Refresh**
   ```bash
   terraform apply -refresh=false
   ```

3. **Use Workspaces Carefully**
   - Avoid running multiple `terraform apply` commands simultaneously
   - Wait for one operation to complete before starting another

4. **Upgrade Plan**
   - Higher-tier Hyperping plans have increased rate limits
   - Contact Hyperping support for enterprise rate limits

**Auto-Remediation**: The provider automatically retries with exponential backoff when rate limits are hit. The `Retry-After` header is respected.

---

## Validation Errors

### ERR-VAL-001: Invalid Monitor Frequency

**Error Message:**
```
❌ Invalid Monitor Frequency

The 'check_frequency' field must be one of the allowed values.

Resource: hyperping_monitor.prod_api
Field:    check_frequency
Value:    45 (invalid)

💡 Allowed values (in seconds):
  10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400

Closest valid values to your input (45):
  • 30 seconds (15 seconds faster)
  • 60 seconds (15 seconds slower)

🔧 Try:
  frequency = 30  # Check every 30 seconds
  frequency = 60  # Check every minute

📚 Documentation:
  https://registry.terraform.io/providers/develeap/hyperping/latest/docs/resources/monitor#check_frequency
```

**Cause**: The `check_frequency` value is not one of the allowed intervals.

**Allowed Values** (in seconds):
- `10` - Every 10 seconds
- `20` - Every 20 seconds
- `30` - Every 30 seconds
- `60` - Every minute (default)
- `120` - Every 2 minutes
- `180` - Every 3 minutes
- `300` - Every 5 minutes
- `600` - Every 10 minutes
- `1800` - Every 30 minutes
- `3600` - Every hour
- `21600` - Every 6 hours
- `43200` - Every 12 hours
- `86400` - Every 24 hours

**Example:**
```hcl
resource "hyperping_monitor" "api" {
  name           = "API Health Check"
  url            = "https://api.example.com/health"
  check_frequency = 60  # Every minute
}
```

---

### ERR-VAL-002: Invalid Region

**Error Message:**
```
❌ Invalid Region

The region 'frenkfurt' is not valid.

Resource: hyperping_monitor.api
Field:    regions

Allowed regions: london, frankfurt, singapore, sydney, virginia, oregon, saopaulo, tokyo, bahrain

💡 Did you mean 'frankfurt'?

📝 Examples:
  regions = ["london", "frankfurt"]          # Europe
  regions = ["virginia", "oregon"]           # North America
  regions = ["singapore", "sydney", "tokyo"] # Asia-Pacific
```

**Cause**: Invalid region name.

**Allowed Regions**:
- `london` - Europe (London)
- `frankfurt` - Europe (Frankfurt)
- `singapore` - Asia (Singapore)
- `sydney` - Australia (Sydney)
- `virginia` - North America (Virginia)
- `oregon` - North America (Oregon)
- `saopaulo` - South America (São Paulo)
- `tokyo` - Asia (Tokyo)
- `bahrain` - Middle East (Bahrain)

**Example:**
```hcl
resource "hyperping_monitor" "global" {
  name    = "Global Health Check"
  url     = "https://api.example.com/health"
  regions = ["london", "frankfurt", "virginia"]
}
```

---

### ERR-VAL-003: Invalid HTTP Method

**Error Message:**
```
❌ Invalid HTTP Method

The HTTP method 'get' is not valid.

Field: http_method

Allowed methods: GET, POST, PUT, PATCH, DELETE, HEAD

💡 Did you mean 'GET'? (check capitalization)

📝 Examples:
  http_method = "GET"   # Most common for health checks
  http_method = "POST"  # For endpoints requiring POST
  http_method = "HEAD"  # For lightweight checks
```

**Cause**: HTTP method is lowercase or misspelled.

**Allowed Methods**: `GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `HEAD` (uppercase)

---

### ERR-VAL-004: Invalid Status Code

**Error Message:**
```
❌ Invalid Status Code

The 'expected_status_code' field must be a valid HTTP status code or range.

Current value: 999
Valid formats:
  • Single code: "200", "404", "500"
  • Range: "2xx", "4xx", "5xx"
  • Multiple: "200,201,202"

📝 Examples:
  expected_status_code = "200"       # Expect OK
  expected_status_code = "2xx"       # Any success code
  expected_status_code = "200,201"   # Multiple codes
```

**Cause**: Invalid status code format.

**Valid Formats**:
- Single code: `"200"`, `"404"`, `"301"`
- Range: `"2xx"`, `"3xx"`, `"4xx"`, `"5xx"`
- Multiple: `"200,201,202"`

---

### ERR-VAL-005: Invalid Incident Status

**Error Message:**
```
❌ Invalid Incident Status

The incident status 'investgating' is not valid.

Allowed statuses: investigating, identified, monitoring, resolved

Status workflow: investigating → identified → monitoring → resolved

💡 Did you mean 'investigating'?
```

**Cause**: Misspelled or invalid incident status.

**Status Workflow**:
1. `investigating` - Initial status when incident is created
2. `identified` - Root cause has been identified
3. `monitoring` - Fix deployed, monitoring for stability
4. `resolved` - Incident fully resolved

---

### ERR-VAL-006: Invalid Incident Severity

**Error Message:**
```
❌ Invalid Incident Severity

The incident severity 'criticl' is not valid.

Allowed severities: minor, major, critical

💡 Suggestions:
  • minor - Low impact, degraded performance
  • major - Significant impact, partial outage
  • critical - Severe impact, complete outage
```

**Cause**: Misspelled or invalid severity.

**Severity Levels**:
- `minor` - Low impact, degraded performance
- `major` - Significant impact, partial outage
- `critical` - Severe impact, complete outage

---

## Not Found Errors

### ERR-404-001: Resource Not Found

**Error Message:**
```
❌ Resource Not Found

The requested resource does not exist or has been deleted.

Resource:  hyperping_monitor.prod_api
Operation: read

💡 Suggestions:
  • Verify the resource ID is correct
  • Check if the resource was deleted outside of Terraform
  • Use 'terraform import' to sync existing resources with your state
  • Check the Hyperping dashboard to confirm resource existence

🔧 Try:
  $ terraform import hyperping_monitor.prod_api <resource_id>

📚 Documentation:
  https://registry.terraform.io/providers/develeap/hyperping/latest/docs/guides/import
  https://app.hyperping.io                   # View resources in dashboard
```

**Cause**: Resource does not exist or has been deleted outside Terraform.

**Solutions**:

1. **Verify Resource Exists**
   - Check the Hyperping dashboard
   - Confirm the resource ID is correct

2. **Import Existing Resource**
   ```bash
   terraform import hyperping_monitor.name mon_abc123
   ```

3. **Recreate Resource**
   ```bash
   terraform apply
   ```

---

## Server Errors

### ERR-5XX-001: Server Error

**Error Message:**
```
❌ Server Error

The Hyperping API returned a server error. Your request will automatically retry.

Resource:  hyperping_monitor.api
Operation: create

💡 Suggestions:
  • The error is temporary and will automatically retry
  • Check the Hyperping status page for any ongoing incidents
  • If the problem persists, contact Hyperping support

📚 Documentation:
  https://status.hyperping.io                # Check service status
  https://hyperping.io/support               # Contact support
```

**Cause**: Temporary server-side issue.

**Auto-Remediation**: Automatically retries with exponential backoff (max 3 retries).

**Solutions**:
1. Wait for automatic retry
2. Check [Hyperping Status](https://status.hyperping.io)
3. Contact support if issue persists

---

## Network Errors

### ERR-NET-001: Connection Failed

**Error Message:**
```
❌ Network Error

Unable to connect to the Hyperping API.

Resource:  hyperping_monitor.api
Operation: create

💡 Suggestions:
  • Check your internet connection
  • Verify firewall settings allow connections to api.hyperping.io
  • Check if a proxy is required and configured correctly
  • Try accessing the API directly with curl to diagnose

🔧 Try:
  $ curl -H "Authorization: Bearer $HYPERPING_API_KEY" https://api.hyperping.io/v1/monitors
  $ ping api.hyperping.io                      # Test DNS resolution
```

**Cause**: Network connectivity issue.

**Solutions**:

1. **Test Connectivity**
   ```bash
   ping api.hyperping.io
   curl https://api.hyperping.io
   ```

2. **Check Firewall**
   - Ensure port 443 (HTTPS) is open
   - Allow connections to `api.hyperping.io`

3. **Configure Proxy** (if needed)
   ```bash
   export HTTPS_PROXY=http://proxy.example.com:8080
   ```

---

## Circuit Breaker Errors

### ERR-CB-001: Circuit Breaker Open

**Error Message:**
```
❌ Circuit Breaker Open

Too many consecutive failures have occurred. The circuit breaker is temporarily blocking requests to prevent cascading failures.

💡 Suggestions:
  • Wait 30 seconds for the circuit breaker to reset
  • Check the Hyperping status page for any ongoing incidents
  • Verify your API credentials are correct
  • Review recent errors in terraform output for the root cause

📚 Documentation:
  https://status.hyperping.io                # Check service status
```

**Cause**: Multiple consecutive API failures triggered the circuit breaker.

**Auto-Remediation**: Circuit breaker automatically resets after 30 seconds.

**Solutions**:
1. Wait 30 seconds
2. Fix underlying issue (auth, network, etc.)
3. Retry operation

---

## Getting Help

If you encounter an error not listed here:

1. **Check Terraform Output**: Error messages include suggestions and documentation links
2. **Review Logs**: Run with `TF_LOG=DEBUG` for detailed logs
   ```bash
   TF_LOG=DEBUG terraform apply
   ```
3. **Check Status Page**: [status.hyperping.io](https://status.hyperping.io)
4. **GitHub Issues**: [Report an issue](https://github.com/develeap/terraform-provider-hyperping/issues)
5. **Hyperping Support**: [hyperping.io/support](https://hyperping.io/support)

---

## Error Code Index

| Code | Category | Description |
|------|----------|-------------|
| ERR-AUTH-001 | Authentication | Invalid API key |
| ERR-AUTH-002 | Authentication | Missing API key |
| ERR-RATE-001 | Rate Limit | Rate limit exceeded |
| ERR-VAL-001 | Validation | Invalid monitor frequency |
| ERR-VAL-002 | Validation | Invalid region |
| ERR-VAL-003 | Validation | Invalid HTTP method |
| ERR-VAL-004 | Validation | Invalid status code |
| ERR-VAL-005 | Validation | Invalid incident status |
| ERR-VAL-006 | Validation | Invalid incident severity |
| ERR-404-001 | Not Found | Resource not found |
| ERR-5XX-001 | Server | Server error |
| ERR-NET-001 | Network | Connection failed |
| ERR-CB-001 | Circuit Breaker | Circuit breaker open |
