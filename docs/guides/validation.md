# Terraform Validation Guide

This guide covers common validation patterns and how to catch configuration errors early using `terraform validate`.

## Quick Start

Always validate your configuration before applying:

```bash
# Initialize Terraform (required before validate)
terraform init

# Validate configuration syntax and references
terraform validate

# Format check (optional but recommended)
terraform fmt -check
```

## Common Validation Errors and Fixes

### 1. Missing Required Arguments

**Error:**
```
Error: Missing required argument

  on main.tf line 1, in resource "hyperping_monitor" "api":
   1: resource "hyperping_monitor" "api" {

The argument "url" is required, but no definition was found.
```

**Fix:** Ensure all required arguments are provided:

```hcl
# Bad - missing required fields
resource "hyperping_monitor" "api" {
  name = "API Monitor"
  # Missing: url, protocol
}

# Good - all required fields present
resource "hyperping_monitor" "api" {
  name     = "API Monitor"
  url      = "https://api.example.com/health"
  protocol = "http"
}
```

### 2. Invalid Attribute Values

**Error:**
```
Error: Invalid value for variable

  check_frequency must be one of: [10 20 30 60 120 180 300 600 1800 3600]
```

**Fix:** Use only allowed values:

```hcl
# Bad - invalid frequency
resource "hyperping_monitor" "api" {
  name            = "API Monitor"
  url             = "https://api.example.com"
  protocol        = "http"
  check_frequency = 45  # Invalid!
}

# Good - valid frequency value
resource "hyperping_monitor" "api" {
  name            = "API Monitor"
  url             = "https://api.example.com"
  protocol        = "http"
  check_frequency = 60  # Valid: 1 minute
}
```

**Valid `check_frequency` values (in seconds):**
| Value | Description |
|-------|-------------|
| 10 | 10 seconds |
| 20 | 20 seconds |
| 30 | 30 seconds |
| 60 | 1 minute |
| 120 | 2 minutes |
| 180 | 3 minutes |
| 300 | 5 minutes |
| 600 | 10 minutes |
| 1800 | 30 minutes |
| 3600 | 1 hour |

### 3. Invalid Resource References

**Error:**
```
Error: Reference to undeclared resource

  on main.tf line 5, in resource "hyperping_incident" "outage":
   5:   status_pages = [hyperping_statuspage.main.id]

A managed resource "hyperping_statuspage" "main" has not been declared.
```

**Fix:** Ensure referenced resources exist:

```hcl
# Bad - referencing non-existent resource
resource "hyperping_incident" "outage" {
  title        = "API Outage"
  text         = "We are investigating the issue."
  type         = "incident"
  status_pages = [hyperping_statuspage.main.id]  # Resource doesn't exist!
}

# Good - define the referenced resource first
resource "hyperping_statuspage" "main" {
  name             = "Status Page"
  hosted_subdomain = "status"
  settings = {
    name      = "Settings"
    languages = ["en"]
  }
}

resource "hyperping_incident" "outage" {
  title        = "API Outage"
  text         = "We are investigating the issue."
  type         = "incident"
  status_pages = [hyperping_statuspage.main.id]  # Now valid
}
```

### 4. Invalid Region Values

**Error:**
```
Error: Invalid region specified

  regions contains invalid value "us-east"
```

**Fix:** Use valid Hyperping region identifiers:

```hcl
# Bad - invalid region names
resource "hyperping_monitor" "api" {
  name     = "API Monitor"
  url      = "https://api.example.com"
  protocol = "http"
  regions  = ["us-east", "eu-west"]  # Invalid!
}

# Good - valid region names
resource "hyperping_monitor" "api" {
  name     = "API Monitor"
  url      = "https://api.example.com"
  protocol = "http"
  regions  = ["virginia", "london"]  # Valid
}
```

**Valid regions:**
- `london`, `frankfurt`, `singapore`, `sydney`
- `virginia`, `oregon`, `saopaulo`, `tokyo`, `bahrain`

### 5. Invalid Protocol/URL Combination

**Error:**
```
Error: Invalid URL for protocol

  Protocol "http" requires URL starting with "http://" or "https://"
```

**Fix:** Match URL scheme to protocol:

```hcl
# Bad - mismatched protocol and URL
resource "hyperping_monitor" "tcp" {
  name     = "Database"
  url      = "https://db.example.com"  # HTTP URL with TCP protocol
  protocol = "port"
}

# Good - TCP protocol with proper format
resource "hyperping_monitor" "tcp" {
  name     = "Database"
  url      = "db.example.com"
  protocol = "port"
  port     = 5432
}

# Good - HTTP protocol with HTTP URL
resource "hyperping_monitor" "api" {
  name     = "API"
  url      = "https://api.example.com"
  protocol = "http"
}
```

### 6. Circular Dependencies

**Error:**
```
Error: Cycle detected

  hyperping_monitor.a depends on hyperping_monitor.b
  hyperping_monitor.b depends on hyperping_monitor.a
```

**Fix:** Break the dependency cycle:

```hcl
# Bad - circular dependency
resource "hyperping_statuspage" "main" {
  name             = "Status"
  hosted_subdomain = "status"
  settings = {
    name      = "Settings"
    languages = ["en"]
  }
  sections = [{
    name = { en = "Services" }
    services = [{
      uuid = hyperping_monitor.api.id  # Depends on monitor
    }]
  }]
}

resource "hyperping_monitor" "api" {
  name     = "API"
  url      = "https://${hyperping_statuspage.main.hosted_subdomain}.example.com"  # Depends on statuspage!
  protocol = "http"
}

# Good - no circular dependency
resource "hyperping_monitor" "api" {
  name     = "API"
  url      = "https://api.example.com"  # Independent URL
  protocol = "http"
}

resource "hyperping_statuspage" "main" {
  name             = "Status"
  hosted_subdomain = "status"
  settings = {
    name      = "Settings"
    languages = ["en"]
  }
  sections = [{
    name = { en = "Services" }
    services = [{
      uuid = hyperping_monitor.api.id
    }]
  }]
}
```

## Pre-Commit Validation Script

Create a `.pre-commit-config.yaml` for automated validation:

```yaml
repos:
  - repo: https://github.com/antonbabenko/pre-commit-terraform
    rev: v1.83.5
    hooks:
      - id: terraform_fmt
      - id: terraform_validate
      - id: terraform_tflint
```

Or use a simple shell script (`.git/hooks/pre-commit`):

```bash
#!/bin/bash
set -e

echo "Running Terraform validation..."

# Find all directories with .tf files
for dir in $(find . -name "*.tf" -exec dirname {} \; | sort -u); do
  echo "Validating $dir..."
  (cd "$dir" && terraform init -backend=false -input=false >/dev/null 2>&1 && terraform validate)
done

echo "All validations passed!"
```

## CI/CD Validation

Add validation to your GitHub Actions workflow:

```yaml
name: Terraform Validate

on:
  pull_request:
    paths:
      - '**/*.tf'

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: 1.8.0

      - name: Terraform Init
        run: terraform init -backend=false

      - name: Terraform Validate
        run: terraform validate

      - name: Terraform Format Check
        run: terraform fmt -check -recursive
```

## Validation Checklist

Before running `terraform apply`, verify:

- [ ] `terraform init` completed successfully
- [ ] `terraform validate` returns no errors
- [ ] `terraform fmt -check` passes (consistent formatting)
- [ ] `terraform plan` shows expected changes
- [ ] All resource references resolve correctly
- [ ] Required variables have values or defaults
- [ ] Sensitive values use environment variables, not hardcoded

## Common Mistakes to Avoid

| Mistake | Impact | Prevention |
|---------|--------|------------|
| Hardcoded API keys | Security breach | Use `HYPERPING_API_KEY` env var |
| Missing `depends_on` | Race conditions | Explicit dependencies for ordering |
| Typos in resource names | Apply failures | Use IDE with HCL support |
| Wrong frequency values | Validation errors | Copy from documentation |
| Invalid regions | API rejection | Use region constants |

## Getting Help

If validation errors persist:

1. Check the [Troubleshooting Guide](../TROUBLESHOOTING.md)
2. Review [Resource Documentation](../resources/)
3. Open an issue: https://github.com/develeap/terraform-provider-hyperping/issues
