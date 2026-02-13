---
page_title: "Input Validation Guide"
subcategory: "Guides"
description: |-
  Comprehensive guide to input validation in the Hyperping Terraform provider.
---

# Input Validation Guide

This guide covers input validation in the Hyperping Terraform provider, explaining how plan-time validation catches errors early and improves the user experience.

## Table of Contents

- [Overview](#overview)
- [Complete Validator Reference](#complete-validator-reference)
- [Common Validation Errors and Fixes](#common-validation-errors-and-fixes)
- [Best Practices](#best-practices)
- [CI/CD Integration](#cicd-integration)
- [Complete Examples](#complete-examples)
- [FAQ](#frequently-asked-questions)

---

## Overview

### What is Plan-Time Validation?

Plan-time validation checks your Terraform configuration **before** making any API calls. This means:

- **Faster feedback**: Errors appear in seconds during `terraform plan`, not minutes later during `terraform apply`
- **No API quota waste**: Invalid configurations never reach the Hyperping API
- **Better error messages**: Specific guidance on what's wrong and how to fix it
- **Local testing**: Validate configurations completely offline

### Benefits Over Apply-Time Errors

| Without Validation | With Validation |
|-------------------|-----------------|
| Error discovered during `terraform apply` | Error discovered during `terraform plan` |
| API call made with invalid data | No API call made |
| Generic API error: "400 Bad Request" | Specific error: "URL must start with http:// or https://" |
| Slower feedback loop | Instant feedback |

### How Validation Works

```bash
# Step 1: Terraform parses your .tf files
terraform init

# Step 2: Validators check attribute values (happens instantly)
terraform plan
# ✅ All validations pass
# → Shows planned changes

# OR
# ❌ Validation fails
# → Shows specific error with fix suggestion
# → No API calls made
```

---

## Complete Validator Reference

The Hyperping provider includes 9 core validators and 3 conditional validators. All validators run automatically during `terraform plan`.

### Core Validators

| Validator | Applied To | Valid Examples | Invalid Examples | Error Message |
|-----------|-----------|----------------|------------------|---------------|
| **URLFormat** | `monitor.url`<br>`statuspage.hostname`<br>`statuspage.settings.website` | `https://example.com`<br>`http://api.example.com/health`<br>`https://example.com:8080/path` | `example.com`<br>`ftp://example.com`<br>`//example.com` | Must be a valid HTTP or HTTPS URL |
| **StringLength** | `monitor.name`<br>`incident.title`<br>`maintenance.title`<br>`statuspage.name` | `"My Monitor"` (1-255 chars)<br>`"API"` (3 chars)<br>`"Production Database Monitor"` (28 chars) | `""` (empty string)<br>Strings over 255 characters | Must be between 1 and 255 characters |
| **CronExpression** | `healthcheck.cron` | `0 0 * * *` (daily at midnight)<br>`*/15 * * * *` (every 15 min)<br>`30 2 * * 1-5` (weekdays at 2:30 AM) | `invalid`<br>`60 * * * *` (invalid minute)<br>`* * *` (too few fields) | Must be valid cron (minute hour day month weekday) |
| **Timezone** | `healthcheck.timezone` | `America/New_York`<br>`Europe/London`<br>`Asia/Tokyo`<br>`UTC` | `EST`<br>`PST`<br>`GMT+5`<br>`Eastern` | Must be valid IANA timezone |
| **PortRange** | `monitor.port` | `80`<br>`443`<br>`3000`<br>`8080`<br>`5432` | `0`<br>`-1`<br>`65536`<br>`99999` | Must be between 1 and 65535 |
| **HexColor** | `statuspage.settings.accent_color` | `#ff5733`<br>`#000000`<br>`#FFFFFF`<br>`#3498db` | `#fff` (too short)<br>`red` (not hex)<br>`#gggggg` (invalid chars)<br>`ff5733` (missing #) | Must be 6-digit hex (#RRGGBB) |
| **NoControlCharacters** | `monitor.request_headers` (keys/values) | `"Authorization"`<br>`"X-Custom-Header"`<br>`"Bearer token123"` | Headers containing `\n`<br>Headers containing `\r`<br>Headers containing `\x00` | No control characters allowed (security) |
| **ReservedHeaderName** | `monitor.request_headers` (keys) | `"X-Custom-Header"`<br>`"X-API-Key"`<br>`"X-Request-ID"` | `"Authorization"`<br>`"Host"`<br>`"Cookie"`<br>`"Set-Cookie"` | Cannot override reserved HTTP headers |
| **ISO8601** | `maintenance.scheduled_start`<br>`maintenance.scheduled_end`<br>`incident.created_at` (computed) | `2026-01-29T10:00:00Z`<br>`2026-02-13T15:30:00Z`<br>`2026-03-01T00:00:00+00:00` | `2026-01-29` (missing time)<br>`10:00:00` (missing date)<br>`Jan 29, 2026` (wrong format) | Must be ISO 8601 format (YYYY-MM-DDTHH:MM:SSZ) |

### Conditional Validators

These validators only apply when certain conditions are met:

| Validator | Condition | Applied To | Description |
|-----------|-----------|-----------|-------------|
| **RequiredWhenProtocolPort** | When `protocol = "port"` | `monitor.port` | Port field becomes required for TCP/UDP checks |
| **RequiredWhenValueIs** | When field X = specific value | Various fields | Makes field Y required based on field X value |
| **AtLeastOneOf** | Always | `statuspage.hostname` OR `statuspage.subdomain` | At least one must be set (both can be set) |

---

## Common Validation Errors and Fixes

### 1. URL Format Errors

**Problem:** `The value "example.com" must be a valid HTTP or HTTPS URL`

**Cause:** Missing protocol prefix (http:// or https://)

**Solution:** Always include the protocol:

```hcl
# ❌ Wrong - missing protocol
resource "hyperping_monitor" "api" {
  name     = "API Monitor"
  url      = "example.com"
  protocol = "http"
}

# ✅ Correct - includes protocol
resource "hyperping_monitor" "api" {
  name     = "API Monitor"
  url      = "https://example.com"
  protocol = "http"
}

# ✅ Also correct - with path
resource "hyperping_monitor" "api" {
  name     = "API Monitor"
  url      = "https://api.example.com/health"
  protocol = "http"
}
```

**Common mistakes:**
- Using `example.com` instead of `https://example.com`
- Using `//example.com` instead of `https://example.com`
- Using unsupported protocols like `ftp://example.com`

---

### 2. String Length Errors

**Problem:** `The value must be between 1 and 255 characters, got 0`

**Cause:** Empty string or string exceeding maximum length

**Solution:** Ensure strings are 1-255 characters:

```hcl
# ❌ Wrong - empty name
resource "hyperping_monitor" "api" {
  name     = ""
  url      = "https://example.com"
  protocol = "http"
}

# ❌ Wrong - name too long (300 characters)
resource "hyperping_monitor" "api" {
  name     = "This is an extremely long monitor name that exceeds the maximum allowed length of 255 characters which will cause a validation error during terraform plan because the Hyperping API has strict limits on field lengths to ensure database integrity and performance so we need to keep our names concise"
  url      = "https://example.com"
  protocol = "http"
}

# ✅ Correct - appropriate length
resource "hyperping_monitor" "api" {
  name     = "Production API Health Check"
  url      = "https://example.com"
  protocol = "http"
}
```

**Tip:** Use descriptive but concise names. Include environment and service type:
- `Production-API-Payment`
- `Staging-DB-Primary`
- `Dev-Frontend-Web`

---

### 3. Cron Expression Errors

**Problem:** `The value "invalid" is not a valid cron expression`

**Cause:** Invalid cron syntax (must be 5 fields: minute hour day month weekday)

**Solution:** Use standard 5-field cron format:

```hcl
# ❌ Wrong - invalid format
resource "hyperping_healthcheck" "daily" {
  name        = "Daily Health Check"
  url         = "https://example.com/health"
  cron        = "invalid"
  timezone    = "UTC"
  http_method = "GET"
}

# ❌ Wrong - too many fields (6-field cron with seconds)
resource "hyperping_healthcheck" "daily" {
  name        = "Daily Health Check"
  url         = "https://example.com/health"
  cron        = "0 0 0 * * *"  # Includes seconds field
  timezone    = "UTC"
  http_method = "GET"
}

# ✅ Correct - daily at midnight
resource "hyperping_healthcheck" "daily" {
  name        = "Daily Health Check"
  url         = "https://example.com/health"
  cron        = "0 0 * * *"
  timezone    = "UTC"
  http_method = "GET"
}

# ✅ Correct - every 15 minutes
resource "hyperping_healthcheck" "frequent" {
  name        = "Frequent Health Check"
  url         = "https://example.com/health"
  cron        = "*/15 * * * *"
  timezone    = "UTC"
  http_method = "GET"
}

# ✅ Correct - weekdays at 9 AM
resource "hyperping_healthcheck" "business_hours" {
  name        = "Business Hours Check"
  url         = "https://example.com/health"
  cron        = "0 9 * * 1-5"
  timezone    = "America/New_York"
  http_method = "GET"
}
```

**Common cron patterns:**

| Pattern | Description | Cron Expression |
|---------|-------------|-----------------|
| Every minute | Runs every minute | `* * * * *` |
| Every 5 minutes | Runs at :00, :05, :10, etc. | `*/5 * * * *` |
| Every 15 minutes | Runs at :00, :15, :30, :45 | `*/15 * * * *` |
| Every hour | Runs at the top of every hour | `0 * * * *` |
| Daily at midnight | Runs once per day at 00:00 | `0 0 * * *` |
| Daily at 3 AM | Runs once per day at 03:00 | `0 3 * * *` |
| Weekdays at 9 AM | Monday-Friday at 09:00 | `0 9 * * 1-5` |
| Weekly on Monday | Every Monday at midnight | `0 0 * * 1` |

**Cron format reference:**
```
* * * * *
│ │ │ │ │
│ │ │ │ └─── Day of week (0-6, 0=Sunday)
│ │ │ └───── Month (1-12)
│ │ └─────── Day of month (1-31)
│ └───────── Hour (0-23)
└─────────── Minute (0-59)
```

---

### 4. Timezone Errors

**Problem:** `The value "EST" is not a valid IANA timezone`

**Cause:** Using abbreviated timezone names instead of IANA timezone database names

**Solution:** Use IANA timezone names (continent/city format):

```hcl
# ❌ Wrong - abbreviated timezone
resource "hyperping_healthcheck" "daily" {
  name        = "Daily Check"
  url         = "https://example.com/health"
  cron        = "0 9 * * *"
  timezone    = "EST"  # Not a valid IANA timezone
  http_method = "GET"
}

# ❌ Wrong - offset format
resource "hyperping_healthcheck" "daily" {
  name        = "Daily Check"
  url         = "https://example.com/health"
  cron        = "0 9 * * *"
  timezone    = "GMT+5"  # Not a valid IANA timezone
  http_method = "GET"
}

# ✅ Correct - IANA timezone
resource "hyperping_healthcheck" "daily" {
  name        = "Daily Check"
  url         = "https://example.com/health"
  cron        = "0 9 * * *"
  timezone    = "America/New_York"
  http_method = "GET"
}

# ✅ Correct - UTC is always valid
resource "hyperping_healthcheck" "daily" {
  name        = "Daily Check"
  url         = "https://example.com/health"
  cron        = "0 9 * * *"
  timezone    = "UTC"
  http_method = "GET"
}
```

**Common timezone mappings:**

| Abbreviated | IANA Timezone | Region |
|-------------|---------------|--------|
| EST/EDT | `America/New_York` | Eastern US |
| CST/CDT | `America/Chicago` | Central US |
| MST/MDT | `America/Denver` | Mountain US |
| PST/PDT | `America/Los_Angeles` | Pacific US |
| GMT/BST | `Europe/London` | United Kingdom |
| CET/CEST | `Europe/Paris` | Central Europe |
| JST | `Asia/Tokyo` | Japan |
| AEST/AEDT | `Australia/Sydney` | Eastern Australia |
| IST | `Asia/Kolkata` | India |
| UTC | `UTC` | Universal |

**Find your timezone:** https://en.wikipedia.org/wiki/List_of_tz_database_time_zones

---

### 5. Port Range Errors

**Problem:** `Port must be between 1 and 65535, got 0`

**Cause:** Port number outside valid TCP/UDP range

**Solution:** Use port numbers between 1 and 65535:

```hcl
# ❌ Wrong - port 0 is reserved
resource "hyperping_monitor" "database" {
  name     = "Database Monitor"
  url      = "db.example.com"
  protocol = "port"
  port     = 0
}

# ❌ Wrong - exceeds maximum
resource "hyperping_monitor" "database" {
  name     = "Database Monitor"
  url      = "db.example.com"
  protocol = "port"
  port     = 99999
}

# ✅ Correct - PostgreSQL default port
resource "hyperping_monitor" "database" {
  name     = "Database Monitor"
  url      = "db.example.com"
  protocol = "port"
  port     = 5432
}

# ✅ Correct - MySQL default port
resource "hyperping_monitor" "mysql" {
  name     = "MySQL Monitor"
  url      = "mysql.example.com"
  protocol = "port"
  port     = 3306
}

# ✅ Correct - Redis default port
resource "hyperping_monitor" "redis" {
  name     = "Redis Monitor"
  url      = "redis.example.com"
  protocol = "port"
  port     = 6379
}
```

**Common ports:**

| Service | Port | Example |
|---------|------|---------|
| HTTP | 80 | Standard web traffic |
| HTTPS | 443 | Secure web traffic |
| SSH | 22 | Secure shell |
| FTP | 21 | File transfer |
| SMTP | 25 | Email sending |
| MySQL | 3306 | MySQL database |
| PostgreSQL | 5432 | PostgreSQL database |
| MongoDB | 27017 | MongoDB database |
| Redis | 6379 | Redis cache |
| Elasticsearch | 9200 | Elasticsearch |

---

### 6. Hex Color Errors

**Problem:** `The value "#fff" must be a 6-digit hex color (e.g., '#ff5733')`

**Cause:** Hex color not in 6-digit #RRGGBB format

**Solution:** Use 6-digit hex colors with # prefix:

```hcl
# ❌ Wrong - only 3 digits
resource "hyperping_statuspage" "main" {
  name             = "Status Page"
  hosted_subdomain = "status"
  settings = {
    name         = "Settings"
    languages    = ["en"]
    accent_color = "#fff"  # Too short
  }
}

# ❌ Wrong - missing # prefix
resource "hyperping_statuspage" "main" {
  name             = "Status Page"
  hosted_subdomain = "status"
  settings = {
    name         = "Settings"
    languages    = ["en"]
    accent_color = "ff5733"  # Missing #
  }
}

# ❌ Wrong - color name instead of hex
resource "hyperping_statuspage" "main" {
  name             = "Status Page"
  hosted_subdomain = "status"
  settings = {
    name         = "Settings"
    languages    = ["en"]
    accent_color = "red"  # Not hex format
  }
}

# ✅ Correct - 6-digit hex
resource "hyperping_statuspage" "main" {
  name             = "Status Page"
  hosted_subdomain = "status"
  settings = {
    name         = "Settings"
    languages    = ["en"]
    accent_color = "#ff5733"
  }
}

# ✅ Correct - white
resource "hyperping_statuspage" "main" {
  name             = "Status Page"
  hosted_subdomain = "status"
  settings = {
    name         = "Settings"
    languages    = ["en"]
    accent_color = "#ffffff"
  }
}

# ✅ Correct - case-insensitive
resource "hyperping_statuspage" "main" {
  name             = "Status Page"
  hosted_subdomain = "status"
  settings = {
    name         = "Settings"
    languages    = ["en"]
    accent_color = "#3498DB"  # Uppercase works
  }
}
```

**Popular brand colors:**

| Brand/Color | Hex Code | Visual |
|-------------|----------|--------|
| Red | `#ff0000` | Primary red |
| Green | `#00ff00` | Primary green |
| Blue | `#0000ff` | Primary blue |
| Orange | `#ff5733` | Hyperping default |
| Slack Purple | `#4a154b` | Brand color |
| GitHub Dark | `#24292e` | Brand color |
| Twitter Blue | `#1da1f2` | Brand color |
| Black | `#000000` | Pure black |
| White | `#ffffff` | Pure white |

**Tip:** Use online color pickers to find hex codes: https://htmlcolorcodes.com/color-picker/

---

### 7. Control Characters in Headers

**Problem:** `Header value must not contain control characters (CR, LF, NULL)`

**Cause:** Header contains newline (`\n`), carriage return (`\r`), or null byte (`\x00`) characters

**Solution:** Remove control characters from header values:

```hcl
# ❌ Wrong - contains newline (security vulnerability)
resource "hyperping_monitor" "api" {
  name     = "API Monitor"
  url      = "https://api.example.com"
  protocol = "http"
  request_headers = {
    "X-Custom-Header" = "value\nInjected-Header: malicious"
  }
}

# ✅ Correct - clean header value
resource "hyperping_monitor" "api" {
  name     = "API Monitor"
  url      = "https://api.example.com"
  protocol = "http"
  request_headers = {
    "X-Custom-Header" = "clean-value-123"
  }
}

# ✅ Correct - multiple headers (use separate keys)
resource "hyperping_monitor" "api" {
  name     = "API Monitor"
  url      = "https://api.example.com"
  protocol = "http"
  request_headers = {
    "X-Custom-Header" = "value1"
    "X-Another-Header" = "value2"
  }
}
```

**Why this validation exists:** Control characters in HTTP headers enable **header injection attacks** (CVE-style vulnerability). This validator prevents security issues by rejecting headers with `\r`, `\n`, or `\x00`.

---

### 8. Reserved Header Names

**Problem:** `The header name "Authorization" is reserved and cannot be overridden`

**Cause:** Attempting to override protected HTTP headers

**Solution:** Use custom header names (X- prefix recommended):

```hcl
# ❌ Wrong - overriding reserved header
resource "hyperping_monitor" "api" {
  name     = "API Monitor"
  url      = "https://api.example.com"
  protocol = "http"
  request_headers = {
    "Authorization" = "Bearer mytoken"  # Reserved!
  }
}

# ❌ Wrong - overriding Host header
resource "hyperping_monitor" "api" {
  name     = "API Monitor"
  url      = "https://api.example.com"
  protocol = "http"
  request_headers = {
    "Host" = "different.example.com"  # Reserved!
  }
}

# ✅ Correct - custom header
resource "hyperping_monitor" "api" {
  name     = "API Monitor"
  url      = "https://api.example.com"
  protocol = "http"
  request_headers = {
    "X-API-Key" = "mytoken"
  }
}

# ✅ Correct - multiple custom headers
resource "hyperping_monitor" "api" {
  name     = "API Monitor"
  url      = "https://api.example.com"
  protocol = "http"
  request_headers = {
    "X-API-Key"    = var.api_key
    "X-Request-ID" = "12345"
    "X-Environment" = "production"
  }
}
```

**Reserved headers (cannot override):**
- `Authorization` - Authentication credentials (managed by provider)
- `Host` - Target hostname (derived from URL)
- `Cookie` / `Set-Cookie` - Cookie management (security risk)
- `Proxy-Authorize` - Proxy authentication
- `Transfer-Encoding` - Message body encoding

**Why this validation exists:** Overriding these headers could:
- Leak API credentials
- Bypass authentication
- Manipulate request routing
- Break HTTP protocol compliance

---

### 9. ISO 8601 Date Format Errors

**Problem:** `The value "2026-01-29" does not appear to be in ISO 8601 format`

**Cause:** Missing time component or wrong datetime format

**Solution:** Use ISO 8601 format with date AND time:

```hcl
# ❌ Wrong - date only (missing time)
resource "hyperping_maintenance" "upgrade" {
  title           = "Database Upgrade"
  text            = "Upgrading to PostgreSQL 15"
  scheduled_start = "2026-01-29"
  scheduled_end   = "2026-01-29"
  status_pages    = [hyperping_statuspage.main.id]
}

# ❌ Wrong - wrong format
resource "hyperping_maintenance" "upgrade" {
  title           = "Database Upgrade"
  text            = "Upgrading to PostgreSQL 15"
  scheduled_start = "January 29, 2026 10:00 AM"
  scheduled_end   = "January 29, 2026 12:00 PM"
  status_pages    = [hyperping_statuspage.main.id]
}

# ✅ Correct - ISO 8601 with timezone
resource "hyperping_maintenance" "upgrade" {
  title           = "Database Upgrade"
  text            = "Upgrading to PostgreSQL 15"
  scheduled_start = "2026-01-29T10:00:00Z"
  scheduled_end   = "2026-01-29T12:00:00Z"
  status_pages    = [hyperping_statuspage.main.id]
}

# ✅ Correct - with timezone offset
resource "hyperping_maintenance" "upgrade" {
  title           = "Database Upgrade"
  text            = "Upgrading to PostgreSQL 15"
  scheduled_start = "2026-01-29T10:00:00+00:00"
  scheduled_end   = "2026-01-29T12:00:00+00:00"
  status_pages    = [hyperping_statuspage.main.id]
}

# ✅ Correct - using Terraform functions
resource "hyperping_maintenance" "upgrade" {
  title           = "Database Upgrade"
  text            = "Upgrading to PostgreSQL 15"
  scheduled_start = timeadd(timestamp(), "24h")
  scheduled_end   = timeadd(timestamp(), "26h")
  status_pages    = [hyperping_statuspage.main.id]
}
```

**ISO 8601 format:** `YYYY-MM-DDTHH:MM:SSZ`

Components:
- `YYYY-MM-DD` - Date (year-month-day)
- `T` - Separator between date and time
- `HH:MM:SS` - Time (hour:minute:second)
- `Z` - UTC timezone (or use `+HH:MM` / `-HH:MM` for offsets)

**Examples:**
- `2026-02-13T15:30:00Z` - February 13, 2026 at 3:30 PM UTC
- `2026-12-25T00:00:00Z` - December 25, 2026 at midnight UTC
- `2026-06-15T09:00:00-05:00` - June 15, 2026 at 9 AM EST

---

## Best Practices

### 1. Always Validate Locally First

Run validation before applying to catch errors immediately:

```bash
# Step 1: Initialize (downloads provider)
terraform init

# Step 2: Validate (checks syntax and validators)
terraform validate

# Step 3: Plan (shows what will change)
terraform plan

# Step 4: Apply (only if plan looks good)
terraform apply
```

**Benefits:**
- Catch errors in seconds, not minutes
- No wasted API calls
- No partial state changes from failed applies

---

### 2. Use Variables with Validation

Add validation to your variables for extra safety:

```hcl
# variables.tf
variable "monitor_url" {
  type        = string
  description = "URL to monitor"

  validation {
    condition     = can(regex("^https?://", var.monitor_url))
    error_message = "URL must start with http:// or https://"
  }
}

variable "monitor_name" {
  type        = string
  description = "Monitor name"

  validation {
    condition     = length(var.monitor_name) >= 1 && length(var.monitor_name) <= 255
    error_message = "Name must be 1-255 characters"
  }
}

variable "check_frequency" {
  type        = number
  description = "Check frequency in seconds"

  validation {
    condition     = contains([10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600], var.check_frequency)
    error_message = "Frequency must be one of: 10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600"
  }
}

# main.tf
resource "hyperping_monitor" "api" {
  name            = var.monitor_name
  url             = var.monitor_url
  protocol        = "http"
  check_frequency = var.check_frequency
}
```

**Benefits:**
- Double validation (variable + resource)
- Reusable validation logic
- Better documentation

---

### 3. Test with Invalid Data

Create test cases with intentionally invalid data to verify validation works:

```hcl
# tests/invalid_configs/
# File: empty_name.tf
resource "hyperping_monitor" "test" {
  name     = ""  # Should fail: StringLength validator
  url      = "https://example.com"
  protocol = "http"
}

# File: invalid_url.tf
resource "hyperping_monitor" "test" {
  name     = "Test"
  url      = "example.com"  # Should fail: URLFormat validator
  protocol = "http"
}

# File: invalid_port.tf
resource "hyperping_monitor" "test" {
  name     = "Test"
  url      = "db.example.com"
  protocol = "port"
  port     = 99999  # Should fail: PortRange validator
}
```

Run tests:
```bash
cd tests/invalid_configs
terraform init
terraform validate
# Should show validation errors
```

---

### 4. Check Terraform Plan Output

Review plan output for validation errors before applying:

```bash
terraform plan
```

**Good plan output (validation passed):**
```
Terraform will perform the following actions:

  # hyperping_monitor.api will be created
  + resource "hyperping_monitor" "api" {
      + name     = "API Monitor"
      + url      = "https://api.example.com"
      + protocol = "http"
    }

Plan: 1 to add, 0 to change, 0 to destroy.
```

**Bad plan output (validation failed):**
```
Error: Invalid URL Format

  on main.tf line 3, in resource "hyperping_monitor" "api":
   3:   url = "example.com"

The value "example.com" must be a valid HTTP or HTTPS URL
```

---

### 5. Use Pre-Commit Hooks

Automate validation with git pre-commit hooks to catch errors before committing:

**Option A: Using pre-commit framework**

Create `.pre-commit-config.yaml`:
```yaml
repos:
  - repo: https://github.com/antonbabenko/pre-commit-terraform
    rev: v1.83.5
    hooks:
      - id: terraform_fmt
      - id: terraform_validate
      - id: terraform_docs
```

Install:
```bash
pip install pre-commit
pre-commit install
```

**Option B: Simple shell script**

Create `.git/hooks/pre-commit`:
```bash
#!/bin/bash
set -e

echo "Running Terraform validation..."

# Find all directories with .tf files
for dir in $(find . -name "*.tf" -not -path "*/.terraform/*" -exec dirname {} \; | sort -u); do
  echo "Validating $dir..."
  (cd "$dir" && terraform init -backend=false -input=false >/dev/null 2>&1 && terraform validate)
done

echo "All validations passed!"
```

Make executable:
```bash
chmod +x .git/hooks/pre-commit
```

**Benefits:**
- Automatic validation on every commit
- Prevents invalid configurations from entering version control
- Faster feedback for team members

---

## CI/CD Integration

### GitHub Actions

Add validation to your GitHub Actions workflow:

```yaml
# .github/workflows/terraform-validate.yml
name: Terraform Validate

on:
  push:
    branches: [ main ]
  pull_request:
    paths:
      - '**/*.tf'
      - '**/*.tfvars'

jobs:
  validate:
    name: Validate Terraform
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: 1.8.0

      - name: Terraform Format Check
        run: terraform fmt -check -recursive

      - name: Terraform Init
        run: terraform init -backend=false

      - name: Terraform Validate
        run: terraform validate

      - name: Run tflint
        uses: terraform-linters/setup-tflint@v4
        with:
          tflint_version: latest

      - name: TFLint Run
        run: tflint --recursive
```

---

### GitLab CI

Add validation to your GitLab CI pipeline:

```yaml
# .gitlab-ci.yml
stages:
  - validate
  - plan
  - apply

variables:
  TF_VERSION: "1.8.0"

.terraform_base:
  image:
    name: hashicorp/terraform:${TF_VERSION}
    entrypoint: [""]

validate:
  extends: .terraform_base
  stage: validate
  script:
    - terraform --version
    - terraform fmt -check -recursive
    - terraform init -backend=false
    - terraform validate
  only:
    changes:
      - "**/*.tf"
      - "**/*.tfvars"
  rules:
    - if: '$CI_PIPELINE_SOURCE == "merge_request_event"'

plan:
  extends: .terraform_base
  stage: plan
  script:
    - terraform init
    - terraform plan -out=tfplan
  artifacts:
    paths:
      - tfplan
  only:
    - merge_requests
```

---

### CircleCI

Add validation to your CircleCI config:

```yaml
# .circleci/config.yml
version: 2.1

orbs:
  terraform: circleci/terraform@3.2.0

jobs:
  validate:
    docker:
      - image: hashicorp/terraform:1.8.0
    steps:
      - checkout
      - run:
          name: Format Check
          command: terraform fmt -check -recursive
      - run:
          name: Initialize
          command: terraform init -backend=false
      - run:
          name: Validate
          command: terraform validate

workflows:
  version: 2
  validate_and_plan:
    jobs:
      - validate:
          filters:
            branches:
              only: /.*/
```

---

### Azure DevOps

Add validation to your Azure Pipeline:

```yaml
# azure-pipelines.yml
trigger:
  branches:
    include:
      - main
  paths:
    include:
      - '**/*.tf'

pool:
  vmImage: 'ubuntu-latest'

steps:
  - task: TerraformInstaller@0
    inputs:
      terraformVersion: '1.8.0'

  - script: |
      terraform fmt -check -recursive
    displayName: 'Terraform Format Check'

  - script: |
      terraform init -backend=false
    displayName: 'Terraform Init'

  - script: |
      terraform validate
    displayName: 'Terraform Validate'

  - task: PublishTestResults@2
    condition: always()
    inputs:
      testResultsFormat: 'JUnit'
      testResultsFiles: '**/validate-results.xml'
```

---

## Complete Examples

### Before Validation (Apply-Time Errors)

**Configuration:**
```hcl
# main.tf (INVALID - multiple errors)
resource "hyperping_monitor" "broken" {
  name     = ""                    # Error 1: empty string
  url      = "example.com"         # Error 2: missing protocol
  protocol = "http"
  port     = 99999                 # Error 3: port out of range
  request_headers = {
    "Authorization" = "Bearer token"  # Error 4: reserved header
  }
}

resource "hyperping_healthcheck" "broken" {
  name        = "Health Check"
  url         = "https://example.com"
  cron        = "invalid cron"     # Error 5: invalid cron
  timezone    = "EST"              # Error 6: invalid timezone
  http_method = "GET"
}

resource "hyperping_statuspage" "broken" {
  name             = "Status"
  hosted_subdomain = "status"
  settings = {
    name         = "Settings"
    languages    = ["en"]
    accent_color = "red"           # Error 7: not hex format
  }
}
```

**Applying without validation:**
```bash
terraform apply
```

**Result (slow, multiple API calls):**
```
hyperping_monitor.broken: Creating...
Error: API returned 400 Bad Request: Invalid name

hyperping_monitor.broken: Creating...
Error: API returned 400 Bad Request: Invalid URL format

hyperping_monitor.broken: Creating...
Error: API returned 400 Bad Request: Port out of range

hyperping_healthcheck.broken: Creating...
Error: API returned 400 Bad Request: Invalid cron expression

hyperping_statuspage.broken: Creating...
Error: API returned 400 Bad Request: Invalid accent_color

# Total time: 2-3 minutes (multiple API calls)
# Result: Frustration, wasted time, API quota used
```

---

### After Validation (Plan-Time Errors)

**Same configuration, but with validators active:**

```bash
terraform validate
```

**Result (instant feedback, no API calls):**
```
Error: Invalid String Length
  on main.tf line 2, in resource "hyperping_monitor" "broken":
   2:   name = ""

  The value must be between 1 and 255 characters, got 0

Error: Invalid URL Format
  on main.tf line 3, in resource "hyperping_monitor" "broken":
   3:   url = "example.com"

  The value "example.com" must be a valid HTTP or HTTPS URL

Error: Invalid Port Number
  on main.tf line 5, in resource "hyperping_monitor" "broken":
   5:   port = 99999

  Port must be between 1 and 65535, got 99999

Error: Reserved Header Name
  on main.tf line 7, in resource "hyperping_monitor" "broken":
   7:     "Authorization" = "Bearer token"

  The header name "Authorization" is reserved and cannot be overridden

Error: Invalid Cron Expression
  on main.tf line 14, in resource "hyperping_healthcheck" "broken":
  14:   cron = "invalid cron"

  The value "invalid cron" is not a valid cron expression
  Expected format: 'minute hour day month weekday' (e.g., '0 0 * * *')

Error: Invalid Timezone
  on main.tf line 15, in resource "hyperping_healthcheck" "broken":
  15:   timezone = "EST"

  The value "EST" is not a valid IANA timezone.
  Use standard timezone names like 'America/New_York', 'Europe/London', or 'UTC'

Error: Invalid Hex Color
  on main.tf line 24, in resource "hyperping_statuspage" "broken":
  24:     accent_color = "red"

  The value "red" must be a 6-digit hex color (e.g., '#ff5733')

# Total time: 1-2 seconds
# Result: Clear actionable errors, no API calls, no quota wasted
```

---

### Fixed Configuration

```hcl
# main.tf (VALID - all errors fixed)
resource "hyperping_monitor" "api" {
  name     = "Production API"           # ✅ 1-255 chars
  url      = "https://api.example.com"  # ✅ Valid HTTP URL
  protocol = "http"
  request_headers = {
    "X-API-Key" = "Bearer token"        # ✅ Custom header (not reserved)
  }
}

resource "hyperping_monitor" "database" {
  name     = "PostgreSQL"
  url      = "db.example.com"
  protocol = "port"
  port     = 5432                       # ✅ Valid port (1-65535)
}

resource "hyperping_healthcheck" "daily" {
  name        = "Daily Health Check"
  url         = "https://api.example.com/health"
  cron        = "0 0 * * *"             # ✅ Valid cron (daily at midnight)
  timezone    = "America/New_York"      # ✅ Valid IANA timezone
  http_method = "GET"
}

resource "hyperping_statuspage" "main" {
  name             = "Status Page"
  hosted_subdomain = "status"
  settings = {
    name         = "Settings"
    languages    = ["en"]
    accent_color = "#ff5733"            # ✅ Valid hex color
  }
}
```

**Applying fixed configuration:**
```bash
terraform validate
# Success! No configuration errors

terraform plan
# Shows planned changes (validation passed)

terraform apply
# Creates resources successfully
```

---

## Frequently Asked Questions

### Why do I get validation errors when my configuration worked before?

**Answer:** Plan-time validation was added in provider version v1.0.10 (released February 2026). Existing configurations that worked with earlier versions may contain invalid data that the API was lenient about accepting. The provider now catches these errors earlier to:

1. Prevent future API breaking changes
2. Improve error messages
3. Catch issues before API calls
4. Align with Terraform best practices

**Fix:** Update your configuration to pass validation. The error messages show exactly what needs to change.

---

### Can I disable validation for specific resources?

**Answer:** No. Validation cannot be disabled because:

1. **Data integrity**: Validators prevent invalid data from reaching the API
2. **Security**: Validators like `NoControlCharacters` and `ReservedHeaderName` prevent security vulnerabilities
3. **API compatibility**: Hyperping API will reject invalid data anyway
4. **Better UX**: Catching errors at plan time is faster than apply time

**Alternative:** If you believe a validator is too strict for your use case, please [open an issue](https://github.com/develeap/terraform-provider-hyperping/issues) with:
- Your use case
- Example configuration
- Why the current validation is problematic

---

### Do validators catch all possible errors?

**Answer:** No. Validators catch **format errors** (wrong structure, invalid characters, out-of-range values), but they don't validate business logic. The Hyperping API still performs additional checks:

**What validators catch:**
- ✅ Invalid URL format (`example.com` instead of `https://example.com`)
- ✅ Port out of range (`99999`)
- ✅ Invalid cron syntax
- ✅ Reserved header names

**What validators don't catch:**
- ❌ URL returns 404 (API checks this)
- ❌ Duplicate monitor names (API checks this)
- ❌ Quota limits exceeded (API checks this)
- ❌ Invalid API key (authentication check)

**Best practice:** Use validators to catch format errors early, but still expect occasional API errors for business logic issues.

---

### What if I need to test with invalid data?

**Answer:** For testing purposes, you can:

1. **Test validation itself**: Create separate test directories with intentionally invalid configs and verify they fail validation
2. **Use development environment**: Test against a separate Hyperping account
3. **Mock the API**: Use tools like Terraform's `mock_provider` for unit testing

You cannot bypass validation for production use.

---

### How do I find valid values for validators?

**Answer:** Refer to the [Complete Validator Reference](#complete-validator-reference) table above, or check resource documentation:

- **URLs**: Must be HTTP/HTTPS with protocol
- **Cron**: Standard 5-field format (minute hour day month weekday)
- **Timezones**: [IANA timezone database](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones)
- **Ports**: 1-65535 (standard TCP/UDP range)
- **Hex colors**: 6-digit format (#RRGGBB)
- **String length**: 1-255 characters for most fields

---

### Does validation work with Terraform variables and functions?

**Answer:** Yes, but with limitations:

**Variables (work perfectly):**
```hcl
variable "monitor_url" {
  type = string
}

resource "hyperping_monitor" "api" {
  name     = "API"
  url      = var.monitor_url  # ✅ Validated
  protocol = "http"
}
```

**Functions (validated at plan time):**
```hcl
resource "hyperping_maintenance" "upgrade" {
  title           = "Maintenance"
  text            = "Description"
  scheduled_start = timeadd(timestamp(), "24h")  # ✅ Validated
  scheduled_end   = timeadd(timestamp(), "26h")  # ✅ Validated
  status_pages    = []
}
```

**Unknown values (validation skipped):**
```hcl
resource "hyperping_monitor" "api" {
  name     = "API"
  url      = data.external.dynamic_url.result.url  # ⚠️ Validation skipped (unknown)
  protocol = "http"
}
```

When values are unknown at plan time (from data sources, external commands, etc.), validation is skipped. Errors will appear at apply time.

---

### How do I report a bug in a validator?

**Answer:** [Open an issue](https://github.com/develeap/terraform-provider-hyperping/issues/new) with:

1. **Validator name** (e.g., URLFormat, PortRange)
2. **Your input value** (what you tried)
3. **Expected behavior** (should pass validation)
4. **Actual behavior** (validation error)
5. **Provider version** (`terraform version`)

Example:
```
Title: URLFormat validator rejects valid IPv6 URLs

Description:
The URLFormat validator rejects valid IPv6 URLs like http://[::1]:8080

Input: url = "http://[::1]:8080"
Expected: Validation passes
Actual: Error "must be a valid HTTP or HTTPS URL"
Provider version: v1.0.10
```

---

## Additional Resources

- **Provider Documentation**: [Resource Reference](../resources/)
- **Troubleshooting Guide**: [TROUBLESHOOTING.md](../TROUBLESHOOTING.md)
- **GitHub Issues**: https://github.com/develeap/terraform-provider-hyperping/issues
- **Terraform Plugin Framework**: https://developer.hashicorp.com/terraform/plugin/framework
- **Cron Expression Reference**: https://crontab.guru/
- **IANA Timezones**: https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
- **Hex Color Picker**: https://htmlcolorcodes.com/color-picker/

---

## Summary

Plan-time validation in the Hyperping Terraform provider:

✅ **Catches errors instantly** (seconds, not minutes)
✅ **No wasted API calls** (saves quota and time)
✅ **Clear error messages** (tells you exactly what to fix)
✅ **Security protection** (prevents header injection, reserved headers)
✅ **Works with CI/CD** (GitHub Actions, GitLab CI, etc.)
✅ **Improves developer experience** (faster feedback loop)

**Remember:** Run `terraform validate` before `terraform apply` to catch errors early!
