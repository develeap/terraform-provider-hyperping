---
page_title: "Migrating from Pingdom to Hyperping - Terraform Provider"
description: |-
  Complete migration guide for moving from Pingdom to Hyperping using Terraform, including check type conversion, tag mapping, and bulk import workflows.
---

# Migrating from Pingdom to Hyperping

This guide provides a comprehensive migration path from Pingdom to Hyperping using Terraform. Whether you're migrating a handful of checks or hundreds, this guide covers everything from exporting Pingdom data to automating the bulk creation of equivalent Hyperping monitors.

## Automated Migration Tool

**🚀 New: Automated CLI Tool Available!**

Simplify your Pingdom migration with our automated CLI tool:

```bash
# Install the Pingdom migration tool
go install github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom@latest

# Run full automated migration
migrate-pingdom migrate \
  --source-api-key $PINGDOM_API_KEY \
  --dest-api-key $HYPERPING_API_KEY \
  --output ./hyperping-migration

# Apply generated configuration
cd hyperping-migration
terraform init
terraform plan
terraform apply
```

**Automation benefits:**
- ✅ **Export all checks** from Pingdom API automatically
- ✅ **Convert check types** (HTTP, TCP, Transaction, etc.)
- ✅ **Map tags and naming** to structured conventions
- ✅ **Generate Terraform** with proper resource dependencies
- ✅ **Create import scripts** for easy resource management
- ✅ **Validate compatibility** and report any issues

**Migration time comparison:**

| Organization Size | Checks | Manual Time | Automated Time | Savings |
|-------------------|--------|-------------|----------------|---------|
| Small | 1-25 | 1-2 days | 30 minutes | 95% |
| Medium | 26-100 | 3-5 days | 1 hour | 90% |
| Large | 100-500 | 1-2 weeks | 2-3 hours | 85% |
| Enterprise | 500+ | 2-4 weeks | 4-6 hours | 80% |

**📚 Comprehensive documentation:** [Automated Migration Tools Guide](./automated-migration.md)

**Tool vs. Manual comparison:**

| Feature | Automated Tool | Manual Process |
|---------|---------------|----------------|
| **Export** | Single command | Manual API calls + scripting |
| **Conversion** | Built-in mapping logic | Write custom scripts |
| **Validation** | Automatic compatibility checks | Manual review |
| **Time** | 30 minutes - 6 hours | 1-4 weeks |
| **Accuracy** | Validated conversions | Prone to human error |
| **Best for** | Standard checks, bulk migrations | Custom configs, learning |

**When to use automated tool:**
- ✅ Migrating 10+ checks
- ✅ Standard HTTP/HTTPS/TCP checks
- ✅ Time-sensitive migration
- ✅ Need repeatable process

**When to use manual process (this guide):**
- ⚠️ <10 checks (manual may be comparable)
- ⚠️ Highly customized check configurations
- ⚠️ Learning migration process in detail
- ⚠️ Need granular control over every aspect

---

## Table of Contents

- [Automated Migration Tool](#automated-migration-tool)
- [Why Migrate from Pingdom?](#why-migrate-from-pingdom)
- [Migration Overview](#migration-overview)
- [Prerequisites](#prerequisites)
- [Pingdom Data Export](#pingdom-data-export)
- [Check Type Conversion](#check-type-conversion)
- [Tag to Naming Convention Mapping](#tag-to-naming-convention-mapping)
- [Bulk Import Workflow](#bulk-import-workflow)
- [Advanced Migration Patterns](#advanced-migration-patterns)
- [Validation and Testing](#validation-and-testing)
- [Post-Migration Cleanup](#post-migration-cleanup)
- [Troubleshooting](#troubleshooting)

## Why Migrate from Pingdom?

### Common Migration Drivers

Organizations migrate from Pingdom to Hyperping for several reasons:

**Cost Reduction**
- Hyperping offers competitive pricing with unlimited checks on higher tiers
- No per-check pricing surprise as monitoring scales
- Better value for comprehensive monitoring coverage

**Modern Infrastructure-as-Code Support**
- Native Terraform provider for Hyperping
- Pingdom's Terraform provider has limited functionality
- Full resource lifecycle management (monitors, incidents, status pages)

**Simpler Management**
- Unified monitoring and status page platform
- Easier integration with modern DevOps workflows
- Better API design for automation

**Feature Parity**
- All essential Pingdom features available in Hyperping
- Modern incident management
- Comprehensive status page capabilities
- Better real-time alerting

### Migration Timeline

Typical migration timeline by organization size:

| Organization Size | Checks | Timeline | Approach |
|-------------------|--------|----------|----------|
| Small (1-25 checks) | 1-25 | 1-2 days | Manual migration |
| Medium (26-100 checks) | 26-100 | 3-5 days | Semi-automated |
| Large (100-500 checks) | 100-500 | 1-2 weeks | Fully automated |
| Enterprise (500+ checks) | 500+ | 2-4 weeks | Phased migration |

## Migration Overview

### High-Level Process

```mermaid
graph LR
    A[Export Pingdom Data] --> B[Convert Check Types]
    B --> C[Map Tags to Names]
    C --> D[Generate Terraform]
    D --> E[Apply Resources]
    E --> F[Validate Monitors]
    F --> G[Cutover Alerting]
    G --> H[Decommission Pingdom]

    style A fill:#FFC107
    style D fill:#2196F3
    style E fill:#4CAF50
    style H fill:#F44336
```

### Migration Phases

**Phase 1: Discovery (Day 1)**
- Export all Pingdom checks via API
- Inventory alerting contacts and integrations
- Document tag naming conventions
- Plan naming scheme for Hyperping

**Phase 2: Configuration (Days 2-3)**
- Convert Pingdom checks to Hyperping equivalents
- Map tags to structured naming conventions
- Generate Terraform configuration
- Review and validate configs

**Phase 3: Deployment (Days 4-5)**
- Create Hyperping monitors via Terraform
- Validate monitors are functioning
- Parallel run with Pingdom for verification
- Update alert routing

**Phase 4: Cutover (Day 6+)**
- Switch primary alerting to Hyperping
- Monitor for any gaps or issues
- Decommission Pingdom checks
- Document new workflows

## Prerequisites

### Required Tools

```bash
# 1. Terraform 1.8 or higher
terraform version

# 2. jq for JSON processing
jq --version

# 3. curl for API access
curl --version

# 4. Python 3.8+ (for migration scripts)
python3 --version

# 5. Git (for version control)
git --version
```

### Required Credentials

```bash
# Pingdom API credentials
export PINGDOM_API_TOKEN="your_pingdom_token"

# Hyperping API key
export HYPERPING_API_KEY="sk_your_hyperping_key"
```

### Terraform Setup

```hcl
# provider.tf
terraform {
  required_version = ">= 1.8"

  required_providers {
    hyperping = {
      source  = "develeap/hyperping"
      version = "~> 1.0"
    }
  }
}

provider "hyperping" {
  # API key from HYPERPING_API_KEY environment variable
}
```

Initialize Terraform:

```bash
terraform init
```

## Pingdom Data Export

### Export All Checks

Use the Pingdom API to export all existing checks:

```bash
#!/bin/bash
# export-pingdom-checks.sh

API_TOKEN="${PINGDOM_API_TOKEN}"
BASE_URL="https://api.pingdom.com/api/3.1"

# Fetch all checks
curl -s -X GET "$BASE_URL/checks" \
  -H "Authorization: Bearer $API_TOKEN" \
  | jq . > pingdom-checks.json

echo "Exported $(jq '.checks | length' pingdom-checks.json) checks to pingdom-checks.json"
```

### Export Check Details

For each check, fetch detailed configuration:

```bash
#!/bin/bash
# export-pingdom-details.sh

API_TOKEN="${PINGDOM_API_TOKEN}"
BASE_URL="https://api.pingdom.com/api/3.1"

# Create directory for detailed exports
mkdir -p pingdom-exports

# Get list of check IDs
check_ids=$(jq -r '.checks[].id' pingdom-checks.json)

# Export details for each check
for check_id in $check_ids; do
  echo "Exporting check $check_id..."
  curl -s -X GET "$BASE_URL/checks/$check_id" \
    -H "Authorization: Bearer $API_TOKEN" \
    | jq . > "pingdom-exports/check-${check_id}.json"

  sleep 0.5  # Rate limiting
done

echo "Detailed exports saved to pingdom-exports/"
```

### Sample Pingdom Check Export

```json
{
  "check": {
    "id": 12345678,
    "name": "Production API - /health",
    "type": "http",
    "hostname": "api.example.com",
    "resolution": 5,
    "sendnotificationwhendown": 2,
    "notifyagainevery": 0,
    "notifywhenbackup": true,
    "url": "/health",
    "encryption": true,
    "port": 443,
    "verify_certificate": true,
    "ssl_down_days_before": 7,
    "auth": "",
    "shouldcontain": "",
    "shouldnotcontain": "",
    "postdata": "",
    "requestheaders": {
      "User-Agent": "Pingdom.com_bot"
    },
    "tags": [
      {
        "name": "production",
        "type": "u"
      },
      {
        "name": "api",
        "type": "u"
      },
      {
        "name": "critical",
        "type": "u"
      }
    ],
    "probe_filters": [
      "region:NA",
      "region:EU"
    ],
    "userids": [],
    "teamids": [],
    "integrationids": []
  }
}
```

### Export Contact Integrations

```bash
#!/bin/bash
# export-pingdom-integrations.sh

API_TOKEN="${PINGDOM_API_TOKEN}"
BASE_URL="https://api.pingdom.com/api/3.1"

# Export integrations
curl -s -X GET "$BASE_URL/alerting/contacts" \
  -H "Authorization: Bearer $API_TOKEN" \
  | jq . > pingdom-contacts.json

curl -s -X GET "$BASE_URL/alerting/integrations" \
  -H "Authorization: Bearer $API_TOKEN" \
  | jq . > pingdom-integrations.json

echo "Exported alerting configuration"
```

## Check Type Conversion

Pingdom and Hyperping have different check type models. This section maps Pingdom check types to Hyperping equivalents.

### Check Type Mapping Table

| Pingdom Check Type | Hyperping Equivalent | Protocol | Notes |
|--------------------|---------------------|----------|-------|
| HTTP | Monitor (protocol: http) | HTTP | Direct 1:1 mapping |
| HTTPS | Monitor (protocol: http, URL with https://) | HTTP | HTTPS detected from URL |
| TCP | Monitor (protocol: port) | TCP | Port monitoring |
| PING | Monitor (protocol: icmp) | ICMP | ICMP ping checks |
| DNS | Not directly supported | N/A | Use HTTP check to DNS-over-HTTPS service |
| UDP | Not directly supported | N/A | Use TCP alternative or external monitoring |
| SMTP | Monitor (protocol: port, port: 25/587) | TCP | Port-based check on SMTP port |
| POP3 | Monitor (protocol: port, port: 110/995) | TCP | Port-based check on POP3 port |
| IMAP | Monitor (protocol: port, port: 143/993) | TCP | Port-based check on IMAP port |
| Transaction | Multiple HTTP Monitors | HTTP | Split into individual endpoint checks |

### HTTP/HTTPS Check Conversion

**Pingdom HTTP Check:**

```json
{
  "type": "http",
  "hostname": "api.example.com",
  "url": "/v1/health",
  "encryption": true,
  "port": 443,
  "resolution": 5,
  "shouldcontain": "healthy",
  "requestheaders": {
    "Authorization": "Bearer token123",
    "Content-Type": "application/json"
  }
}
```

**Hyperping Equivalent:**

```hcl
resource "hyperping_monitor" "api_health" {
  name                 = "[PROD]-API-Health"
  url                  = "https://api.example.com/v1/health"
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 300  # 5 minutes (Pingdom resolution: 5)
  expected_status_code = "200"

  request_headers = [
    {
      name  = "Authorization"
      value = "Bearer token123"
    },
    {
      name  = "Content-Type"
      value = "application/json"
    }
  ]

  body_regex = "healthy"  # Maps to "shouldcontain"

  regions = [
    "virginia",
    "london",
    "frankfurt"
  ]
}
```

### Frequency/Resolution Conversion

Pingdom uses "resolution" (in minutes), Hyperping uses "check_frequency" (in seconds):

| Pingdom Resolution | Hyperping Frequency | Use Case |
|--------------------|---------------------|----------|
| 1 minute | 60 seconds | Critical endpoints |
| 5 minutes | 300 seconds | Standard monitoring |
| 10 minutes | 600 seconds | Non-critical resources |
| 15 minutes | 900 seconds | Internal services |
| 30 minutes | 1800 seconds | Low-priority checks |
| 60 minutes | 3600 seconds | Daily health checks |

**Conversion formula:**

```
hyperping_frequency_seconds = pingdom_resolution_minutes * 60
```

### TCP Check Conversion

**Pingdom TCP Check:**

```json
{
  "type": "tcp",
  "hostname": "db.example.com",
  "port": 5432,
  "resolution": 5
}
```

**Hyperping Equivalent:**

```hcl
resource "hyperping_monitor" "database_tcp" {
  name            = "[PROD]-Database-TCP"
  url             = "db.example.com"
  protocol        = "port"
  port            = 5432
  check_frequency = 300

  regions = [
    "virginia",
    "london"
  ]
}
```

### PING/ICMP Check Conversion

**Pingdom PING Check:**

```json
{
  "type": "ping",
  "hostname": "server.example.com",
  "resolution": 1
}
```

**Hyperping Equivalent:**

```hcl
resource "hyperping_monitor" "server_ping" {
  name            = "[PROD]-Server-ICMP"
  url             = "server.example.com"
  protocol        = "icmp"
  check_frequency = 60

  regions = [
    "virginia",
    "london",
    "frankfurt"
  ]
}
```

### Transaction Check Conversion

Pingdom "Transaction" checks simulate multi-step user journeys. Hyperping doesn't have a direct equivalent, so break them into individual HTTP monitor steps.

**Pingdom Transaction (conceptual):**

```
1. GET https://example.com/login (expect 200)
2. POST https://example.com/login (with credentials, expect 302)
3. GET https://example.com/dashboard (expect 200, contain "Welcome")
```

**Hyperping Equivalent (multiple monitors):**

```hcl
# Step 1: Login page availability
resource "hyperping_monitor" "transaction_login_page" {
  name                 = "[PROD]-Transaction-LoginPage"
  url                  = "https://example.com/login"
  protocol             = "http"
  http_method          = "GET"
  expected_status_code = "200"
  check_frequency      = 300
}

# Step 2: Login form submission
resource "hyperping_monitor" "transaction_login_post" {
  name                 = "[PROD]-Transaction-LoginSubmit"
  url                  = "https://example.com/login"
  protocol             = "http"
  http_method          = "POST"
  expected_status_code = "302"
  check_frequency      = 300

  request_headers = [
    {
      name  = "Content-Type"
      value = "application/x-www-form-urlencoded"
    }
  ]

  request_body = "username=test&password=test123"
}

# Step 3: Dashboard access
resource "hyperping_monitor" "transaction_dashboard" {
  name                 = "[PROD]-Transaction-Dashboard"
  url                  = "https://example.com/dashboard"
  protocol             = "http"
  http_method          = "GET"
  expected_status_code = "200"
  check_frequency      = 300
  body_regex           = "Welcome"
}
```

**Note:** For true transaction monitoring, consider using external tools like Playwright/Selenium with a healthcheck that reports success/failure to Hyperping.

### Real User Monitoring (RUM) Conversion

Pingdom Real User Monitoring (RUM) tracks actual user performance metrics from browsers. Hyperping focuses on synthetic monitoring from external probes.

**Migration Path:**

1. **Keep Pingdom RUM temporarily** - RUM and synthetic monitoring serve different purposes
2. **Add client-side monitoring** - Use services like Sentry, DataDog RUM, or Google Analytics
3. **Synthetic equivalent** - Create HTTP monitors for critical user paths

**Example:**

```hcl
# Synthetic check for page load time
resource "hyperping_monitor" "homepage_performance" {
  name                 = "[PROD]-Homepage-LoadTime"
  url                  = "https://example.com"
  protocol             = "http"
  http_method          = "GET"
  expected_status_code = "200"
  check_frequency      = 300
  timeout              = 10

  # Monitor response time thresholds
  # Configure alerts if response time exceeds acceptable limits
}
```

## Tag to Naming Convention Mapping

Pingdom uses tags for organization. Hyperping uses structured naming conventions for better filtering and management.

### Pingdom Tag Structure

Common Pingdom tag patterns:

```
Tags: production, api, critical
Tags: staging, web, medium
Tags: dev, database, low
Tags: customer-acme, frontend, high
```

### Hyperping Naming Convention

Hyperping recommended format: `[ENVIRONMENT]-Category-ServiceName`

```
[PROD]-API-Health
[STAGING]-Web-Homepage
[DEV]-Database-Connection
[PROD-ACME]-Frontend-Login
```

### Tag Conversion Patterns

#### Pattern 1: Environment + Service Type + Priority

**Pingdom:**
```json
{
  "name": "API Health Check",
  "tags": ["production", "api", "critical"]
}
```

**Hyperping:**
```hcl
resource "hyperping_monitor" "prod_api_health" {
  name = "[PROD]-API-Health-Critical"
  # ... configuration
}
```

#### Pattern 2: Customer/Tenant + Component

**Pingdom:**
```json
{
  "name": "ACME Corp Website",
  "tags": ["customer-acme", "website", "frontend"]
}
```

**Hyperping:**
```hcl
resource "hyperping_monitor" "acme_website_frontend" {
  name = "[ACME]-Website-Frontend"
  # ... configuration
}
```

#### Pattern 3: Service + Region

**Pingdom:**
```json
{
  "name": "API Gateway US-East",
  "tags": ["api-gateway", "us-east", "production"]
}
```

**Hyperping:**
```hcl
resource "hyperping_monitor" "prod_api_gateway_us_east" {
  name = "[PROD]-APIGateway-USEast"
  # ... configuration
}
```

### Tag-to-Name Conversion Script

Automate tag-to-name conversion:

```python
#!/usr/bin/env python3
# convert-tags-to-names.py

import json
import re

def load_pingdom_checks(filename):
    with open(filename, 'r') as f:
        return json.load(f)['checks']

def extract_environment(tags):
    """Extract environment from tags"""
    env_map = {
        'production': 'PROD',
        'prod': 'PROD',
        'staging': 'STAGING',
        'stage': 'STAGING',
        'development': 'DEV',
        'dev': 'DEV',
        'qa': 'QA',
        'test': 'TEST'
    }

    for tag in tags:
        tag_name = tag['name'].lower()
        if tag_name in env_map:
            return env_map[tag_name]

    return 'UNKNOWN'

def extract_category(tags):
    """Extract service category from tags"""
    categories = {
        'api': 'API',
        'web': 'Web',
        'website': 'Web',
        'database': 'Database',
        'db': 'Database',
        'cache': 'Cache',
        'redis': 'Cache',
        'queue': 'Queue',
        'worker': 'Worker',
        'cdn': 'CDN',
        'dns': 'DNS',
        'mail': 'Mail',
        'smtp': 'Mail',
        'frontend': 'Frontend',
        'backend': 'Backend'
    }

    for tag in tags:
        tag_name = tag['name'].lower()
        if tag_name in categories:
            return categories[tag_name]

    return 'Service'

def extract_customer(tags):
    """Extract customer/tenant from tags"""
    for tag in tags:
        tag_name = tag['name']
        if tag_name.startswith('customer-'):
            customer = tag_name.replace('customer-', '').upper()
            return customer

    return None

def sanitize_service_name(name):
    """Clean up service name for use in Hyperping naming"""
    # Remove common prefixes
    name = re.sub(r'^(Production|Staging|Dev|QA)\s*-?\s*', '', name, flags=re.IGNORECASE)

    # Remove special characters
    name = re.sub(r'[^\w\s-]', '', name)

    # Convert to title case and remove spaces
    name = ''.join(word.capitalize() for word in name.split())

    # Limit length
    if len(name) > 30:
        name = name[:30]

    return name

def generate_hyperping_name(check):
    """Generate Hyperping-style name from Pingdom check"""
    tags = check.get('tags', [])
    original_name = check['name']

    environment = extract_environment(tags)
    category = extract_category(tags)
    customer = extract_customer(tags)
    service_name = sanitize_service_name(original_name)

    if customer:
        return f"[{environment}-{customer}]-{category}-{service_name}"
    else:
        return f"[{environment}]-{category}-{service_name}"

def generate_terraform_resource_name(hyperping_name):
    """Generate valid Terraform resource name"""
    # Remove brackets and special chars
    name = re.sub(r'[\[\]]', '', hyperping_name)
    name = re.sub(r'[^a-zA-Z0-9_]', '_', name)
    name = name.lower()

    # Ensure doesn't start with number
    if name[0].isdigit():
        name = 'monitor_' + name

    return name

# Main conversion
if __name__ == '__main__':
    checks = load_pingdom_checks('pingdom-checks.json')

    print("# Tag to Name Conversion Results\n")
    print(f"{'Pingdom Name':<40} {'Tags':<40} {'Hyperping Name':<50}")
    print("=" * 130)

    for check in checks:
        original = check['name']
        tags = ', '.join([t['name'] for t in check.get('tags', [])])
        hyperping_name = generate_hyperping_name(check)

        print(f"{original:<40} {tags:<40} {hyperping_name:<50}")
```

### Example Conversion Output

```
Pingdom Name                             Tags                                     Hyperping Name
==================================================================================================================================
Production API - /health                 production, api, critical                [PROD]-API-Health
Staging Website Homepage                 staging, web, frontend                   [STAGING]-Web-Homepage
Database Connection Check                production, database, critical           [PROD]-Database-ConnectionCheck
ACME Corp Login                          customer-acme, frontend, production      [PROD-ACME]-Frontend-Login
Dev Redis Cache                          dev, cache, redis                        [DEV]-Cache-Redis
```

## Bulk Import Workflow

Automate the conversion and creation of Hyperping monitors from Pingdom data.

### Step 1: Generate Terraform Configuration

Python script to convert Pingdom JSON to Terraform HCL:

```python
#!/usr/bin/env python3
# generate-terraform-from-pingdom.py

import json
import sys
from convert_tags_to_names import (
    generate_hyperping_name,
    generate_terraform_resource_name,
    load_pingdom_checks
)

def pingdom_to_hyperping_protocol(pingdom_type, encryption=False):
    """Convert Pingdom check type to Hyperping protocol"""
    type_map = {
        'http': 'http',
        'https': 'http',
        'tcp': 'port',
        'ping': 'icmp'
    }
    return type_map.get(pingdom_type.lower(), 'http')

def pingdom_to_hyperping_frequency(resolution):
    """Convert Pingdom resolution (minutes) to Hyperping frequency (seconds)"""
    frequency = resolution * 60

    # Round to nearest allowed frequency
    allowed = [60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400]
    return min(allowed, key=lambda x: abs(x - frequency))

def pingdom_regions_to_hyperping(probe_filters):
    """Convert Pingdom probe regions to Hyperping regions"""
    region_map = {
        'NA': ['virginia', 'oregon'],
        'EU': ['london', 'frankfurt'],
        'APAC': ['singapore', 'sydney', 'tokyo'],
        'LATAM': ['saopaulo']
    }

    if not probe_filters:
        # Default to all regions if not specified
        return ['virginia', 'london', 'frankfurt', 'singapore']

    hyperping_regions = []
    for pf in probe_filters:
        if pf.startswith('region:'):
            region = pf.split(':')[1]
            if region in region_map:
                hyperping_regions.extend(region_map[region])

    return list(set(hyperping_regions)) if hyperping_regions else ['virginia', 'london']

def escape_hcl_string(s):
    """Escape special characters for HCL strings"""
    if s is None:
        return ""
    return s.replace('\\', '\\\\').replace('"', '\\"').replace('\n', '\\n')

def generate_monitor_hcl(check):
    """Generate HCL for a single Hyperping monitor"""
    hyperping_name = generate_hyperping_name(check)
    tf_resource_name = generate_terraform_resource_name(hyperping_name)

    # Build URL
    if check['type'] in ['http', 'https']:
        protocol = 'https' if check.get('encryption', False) else 'http'
        url = f"{protocol}://{check['hostname']}{check.get('url', '/')}"
    else:
        url = check['hostname']

    # Basic configuration
    hcl = f'resource "hyperping_monitor" "{tf_resource_name}" {{\n'
    hcl += f'  name = "{escape_hcl_string(hyperping_name)}"\n'
    hcl += f'  url  = "{escape_hcl_string(url)}"\n'

    # Protocol
    protocol = pingdom_to_hyperping_protocol(check['type'], check.get('encryption', False))
    hcl += f'  protocol = "{protocol}"\n'

    # HTTP-specific settings
    if protocol == 'http':
        hcl += f'  http_method = "GET"\n'
        hcl += f'  expected_status_code = "200"\n'

    # Port for TCP checks
    if protocol == 'port' and 'port' in check:
        hcl += f'  port = {check["port"]}\n'

    # Frequency
    frequency = pingdom_to_hyperping_frequency(check.get('resolution', 5))
    hcl += f'  check_frequency = {frequency}\n'

    # Regions
    regions = pingdom_regions_to_hyperping(check.get('probe_filters', []))
    hcl += f'  regions = [\n'
    for region in regions:
        hcl += f'    "{region}",\n'
    hcl += f'  ]\n'

    # Request headers
    if check.get('requestheaders'):
        hcl += f'  request_headers = [\n'
        for header_name, header_value in check['requestheaders'].items():
            hcl += f'    {{\n'
            hcl += f'      name  = "{escape_hcl_string(header_name)}"\n'
            hcl += f'      value = "{escape_hcl_string(header_value)}"\n'
            hcl += f'    }},\n'
        hcl += f'  ]\n'

    # Body regex (shouldcontain)
    if check.get('shouldcontain'):
        hcl += f'  body_regex = "{escape_hcl_string(check["shouldcontain"])}"\n'

    # SSL verification
    if protocol == 'http' and check.get('encryption'):
        hcl += f'  verify_ssl = {str(check.get("verify_certificate", True)).lower()}\n'

    hcl += '}\n'

    return hcl

def generate_all_monitors(checks, output_file):
    """Generate Terraform configuration for all checks"""
    with open(output_file, 'w') as f:
        f.write('# Generated from Pingdom export\n')
        f.write('# Review and adjust as needed before applying\n\n')

        for i, check in enumerate(checks):
            f.write(f'# Pingdom Check ID: {check["id"]}\n')
            f.write(f'# Original Name: {check["name"]}\n')

            hcl = generate_monitor_hcl(check)
            f.write(hcl)
            f.write('\n')

        print(f"Generated Terraform configuration for {len(checks)} monitors")
        print(f"Saved to: {output_file}")

if __name__ == '__main__':
    if len(sys.argv) > 1:
        input_file = sys.argv[1]
        output_file = sys.argv[2] if len(sys.argv) > 2 else 'generated-monitors.tf'
    else:
        input_file = 'pingdom-checks.json'
        output_file = 'generated-monitors.tf'

    checks = load_pingdom_checks(input_file)
    generate_all_monitors(checks, output_file)
```

### Step 2: Review Generated Configuration

The script generates a Terraform file. Review it:

```bash
python3 generate-terraform-from-pingdom.py pingdom-checks.json generated-monitors.tf

# Review output
less generated-monitors.tf
```

**Example Generated Configuration:**

```hcl
# Generated from Pingdom export
# Review and adjust as needed before applying

# Pingdom Check ID: 12345678
# Original Name: Production API - /health
resource "hyperping_monitor" "prod_api_health" {
  name                 = "[PROD]-API-Health"
  url                  = "https://api.example.com/health"
  protocol             = "http"
  http_method          = "GET"
  expected_status_code = "200"
  check_frequency      = 300

  regions = [
    "virginia",
    "london",
    "frankfurt",
  ]

  request_headers = [
    {
      name  = "User-Agent"
      value = "Hyperping Monitor"
    },
  ]

  verify_ssl = true
}

# Pingdom Check ID: 87654321
# Original Name: Staging Web Homepage
resource "hyperping_monitor" "staging_web_homepage" {
  name                 = "[STAGING]-Web-Homepage"
  url                  = "https://staging.example.com/"
  protocol             = "http"
  http_method          = "GET"
  expected_status_code = "200"
  check_frequency      = 300

  regions = [
    "virginia",
    "london",
  ]

  verify_ssl = true
}
```

### Step 3: Validate Configuration

```bash
# Format Terraform files
terraform fmt

# Validate syntax
terraform validate

# Plan to see what will be created
terraform plan
```

Review the plan output carefully:

```
Plan: 47 to add, 0 to change, 0 to destroy.

Changes to Outputs:
  + monitor_count = 47
```

### Step 4: Apply Configuration

Create monitors in Hyperping:

```bash
# Apply in stages for large migrations
terraform apply -target=hyperping_monitor.prod_api_health

# Or apply all at once (small migrations)
terraform apply
```

### Step 5: Verify Monitors

Check that monitors are created and functioning:

```bash
# List all created monitors
terraform state list | grep hyperping_monitor

# Check specific monitor
terraform state show hyperping_monitor.prod_api_health

# Verify in Hyperping dashboard
# https://app.hyperping.io/monitors
```

### Complete Migration Script

Automated end-to-end migration:

```bash
#!/bin/bash
# migrate-pingdom-to-hyperping.sh

set -e

echo "=== Pingdom to Hyperping Migration ==="
echo ""

# Step 1: Export Pingdom data
echo "[1/6] Exporting Pingdom checks..."
./export-pingdom-checks.sh

# Step 2: Convert tags to names
echo "[2/6] Converting tags to naming convention..."
python3 convert-tags-to-names.py > tag-conversion-report.txt
cat tag-conversion-report.txt

# Step 3: Generate Terraform configuration
echo "[3/6] Generating Terraform configuration..."
python3 generate-terraform-from-pingdom.py pingdom-checks.json monitors.tf

# Step 4: Validate configuration
echo "[4/6] Validating Terraform configuration..."
terraform fmt monitors.tf
terraform validate

# Step 5: Plan
echo "[5/6] Planning Terraform apply..."
terraform plan -out=migration.tfplan

# Step 6: Prompt for apply
echo "[6/6] Ready to create monitors in Hyperping"
read -p "Apply changes? (yes/no): " confirm

if [ "$confirm" = "yes" ]; then
  terraform apply migration.tfplan
  echo ""
  echo "Migration complete!"
  echo "Created $(terraform state list | grep hyperping_monitor | wc -l) monitors"
else
  echo "Apply cancelled. Run 'terraform apply migration.tfplan' when ready."
fi
```

## Advanced Migration Patterns

### Phased Migration (Large Deployments)

For organizations with hundreds of checks, migrate in phases:

**Phase 1: Critical Production Monitors**

```bash
# Filter for production + critical tags
jq '.checks | map(select(.tags | map(.name) | contains(["production", "critical"])))' \
  pingdom-checks.json > phase1-critical.json

python3 generate-terraform-from-pingdom.py phase1-critical.json phase1.tf

terraform apply -target=module.critical_monitors
```

**Phase 2: Standard Production Monitors**

```bash
# Production but not critical
jq '.checks | map(select((.tags | map(.name) | contains(["production"])) and (.tags | map(.name) | contains(["critical"]) | not)))' \
  pingdom-checks.json > phase2-production.json

python3 generate-terraform-from-pingdom.py phase2-production.json phase2.tf

terraform apply -target=module.production_monitors
```

**Phase 3: Non-Production Environments**

```bash
# Staging, dev, QA
jq '.checks | map(select(.tags | map(.name) | contains(["production"]) | not))' \
  pingdom-checks.json > phase3-nonprod.json

python3 generate-terraform-from-pingdom.py phase3-nonprod.json phase3.tf

terraform apply
```

### Multi-Region Migration

For global deployments with region-specific monitors:

```hcl
# regions.tf
locals {
  regions = {
    us_east = {
      pingdom_filter = "region:NA"
      hyperping_regions = ["virginia"]
    }
    eu_west = {
      pingdom_filter = "region:EU"
      hyperping_regions = ["london", "frankfurt"]
    }
    asia_pacific = {
      pingdom_filter = "region:APAC"
      hyperping_regions = ["singapore", "sydney", "tokyo"]
    }
  }
}

module "monitors_us_east" {
  source = "./modules/monitors"

  region_filter = local.regions.us_east.pingdom_filter
  hyperping_regions = local.regions.us_east.hyperping_regions
}

module "monitors_eu_west" {
  source = "./modules/monitors"

  region_filter = local.regions.eu_west.pingdom_filter
  hyperping_regions = local.regions.eu_west.hyperping_regions
}
```

### Multi-Tenant Migration

For MSPs or platforms with customer-specific monitors:

```bash
# Extract unique customers from tags
jq -r '.checks[].tags[] | select(.name | startswith("customer-")) | .name' \
  pingdom-checks.json | sort -u > customers.txt

# Generate per-customer configurations
while read customer; do
  customer_id=$(echo $customer | sed 's/customer-//')

  jq --arg customer "$customer" \
    '.checks | map(select(.tags | map(.name) | contains([$customer])))' \
    pingdom-checks.json > "customer-${customer_id}.json"

  python3 generate-terraform-from-pingdom.py \
    "customer-${customer_id}.json" \
    "monitors-${customer_id}.tf"
done < customers.txt
```

### Parallel Run Strategy

Run Pingdom and Hyperping in parallel during migration:

```hcl
# Enable all Hyperping monitors but keep Pingdom active
resource "hyperping_monitor" "prod_api_health" {
  name            = "[PROD]-API-Health"
  url             = "https://api.example.com/health"
  protocol        = "http"
  check_frequency = 300

  # Initially pause for validation
  paused = true
}

# After validating Hyperping monitors work correctly:
# 1. Unpause Hyperping monitors: paused = false
# 2. Configure alerting to route to Hyperping
# 3. Monitor both systems for 1-2 weeks
# 4. Decommission Pingdom checks
```

## Validation and Testing

### Pre-Migration Validation

Validate exported data before conversion:

```python
#!/usr/bin/env python3
# validate-pingdom-export.py

import json

def validate_export(filename):
    with open(filename, 'r') as f:
        data = json.load(f)

    checks = data.get('checks', [])

    print(f"Total checks: {len(checks)}")
    print("\nCheck type distribution:")

    type_counts = {}
    for check in checks:
        check_type = check.get('type', 'unknown')
        type_counts[check_type] = type_counts.get(check_type, 0) + 1

    for check_type, count in sorted(type_counts.items()):
        print(f"  {check_type}: {count}")

    print("\nTag distribution:")
    tag_counts = {}
    for check in checks:
        for tag in check.get('tags', []):
            tag_name = tag['name']
            tag_counts[tag_name] = tag_counts.get(tag_name, 0) + 1

    for tag_name, count in sorted(tag_counts.items(), key=lambda x: -x[1])[:20]:
        print(f"  {tag_name}: {count}")

    # Validation checks
    print("\nValidation:")
    issues = []

    for check in checks:
        if check['type'] in ['dns', 'udp']:
            issues.append(f"Check {check['id']} ({check['name']}) uses unsupported type: {check['type']}")

    if issues:
        print(f"  Found {len(issues)} potential issues:")
        for issue in issues[:10]:
            print(f"    - {issue}")
    else:
        print("  No issues found")

if __name__ == '__main__':
    validate_export('pingdom-checks.json')
```

### Post-Migration Testing

Verify monitors are functioning:

```bash
#!/bin/bash
# test-monitors.sh

echo "Testing Hyperping monitors..."

# Get all monitor IDs from Terraform state
monitor_ids=$(terraform state list | grep hyperping_monitor | \
  xargs -I {} terraform state show {} | grep 'id.*=' | awk '{print $3}' | tr -d '"')

# Check each monitor status
for monitor_id in $monitor_ids; do
  status=$(curl -s -H "Authorization: Bearer $HYPERPING_API_KEY" \
    "https://api.hyperping.io/monitors/$monitor_id" | jq -r '.status')

  echo "Monitor $monitor_id: $status"
done
```

### Comparison Testing

Compare Pingdom vs Hyperping monitoring results:

```python
#!/usr/bin/env python3
# compare-monitoring-results.py

import requests
import os
from datetime import datetime, timedelta

PINGDOM_TOKEN = os.getenv('PINGDOM_API_TOKEN')
HYPERPING_KEY = os.getenv('HYPERPING_API_KEY')

def get_pingdom_uptime(check_id, days=7):
    """Get uptime percentage from Pingdom"""
    url = f"https://api.pingdom.com/api/3.1/summary.average/{check_id}"
    headers = {"Authorization": f"Bearer {PINGDOM_TOKEN}"}

    end = datetime.now()
    start = end - timedelta(days=days)

    params = {
        "from": int(start.timestamp()),
        "to": int(end.timestamp())
    }

    response = requests.get(url, headers=headers, params=params)
    data = response.json()

    return data['summary']['status']['totalup']

def get_hyperping_uptime(monitor_id, days=7):
    """Get uptime percentage from Hyperping"""
    url = f"https://api.hyperping.io/monitors/{monitor_id}/uptime"
    headers = {"Authorization": f"Bearer {HYPERPING_KEY}"}

    params = {"days": days}

    response = requests.get(url, headers=headers, params=params)
    data = response.json()

    return data.get('uptime_percentage', 0)

# Compare specific checks
comparisons = [
    {"pingdom_id": 12345678, "hyperping_id": "mon_abc123", "name": "API Health"},
    {"pingdom_id": 87654321, "hyperping_id": "mon_def456", "name": "Web Homepage"}
]

print(f"{'Check Name':<30} {'Pingdom Uptime':<15} {'Hyperping Uptime':<15} {'Difference':<10}")
print("=" * 70)

for comp in comparisons:
    pingdom_uptime = get_pingdom_uptime(comp['pingdom_id'])
    hyperping_uptime = get_hyperping_uptime(comp['hyperping_id'])
    difference = abs(pingdom_uptime - hyperping_uptime)

    print(f"{comp['name']:<30} {pingdom_uptime:>13.2f}% {hyperping_uptime:>13.2f}% {difference:>8.2f}%")
```

## Post-Migration Cleanup

### Decommission Pingdom Checks

After successful migration and validation period:

```bash
#!/bin/bash
# decommission-pingdom.sh

API_TOKEN="${PINGDOM_API_TOKEN}"
BASE_URL="https://api.pingdom.com/api/3.1"

echo "WARNING: This will DELETE all Pingdom checks"
echo "Ensure Hyperping is fully operational before proceeding"
read -p "Type 'DELETE' to confirm: " confirm

if [ "$confirm" != "DELETE" ]; then
  echo "Cancelled"
  exit 0
fi

# Pause all checks first (safety measure)
check_ids=$(jq -r '.checks[].id' pingdom-checks.json)

for check_id in $check_ids; do
  echo "Pausing check $check_id..."
  curl -s -X PUT "$BASE_URL/checks/$check_id" \
    -H "Authorization: Bearer $API_TOKEN" \
    -d "paused=true"

  sleep 0.5
done

echo "All checks paused. Monitor for 24 hours, then run deletion."
echo ""
read -p "Delete all paused checks now? (yes/no): " delete_confirm

if [ "$delete_confirm" = "yes" ]; then
  for check_id in $check_ids; do
    echo "Deleting check $check_id..."
    curl -s -X DELETE "$BASE_URL/checks/$check_id" \
      -H "Authorization: Bearer $API_TOKEN"

    sleep 0.5
  done

  echo "All checks deleted from Pingdom"
fi
```

### Update Documentation

Update team documentation:

```markdown
# Monitoring Migration Complete

## New Process

Monitoring is now managed via Terraform in the `terraform-hyperping/` directory.

### Make Changes

1. Edit `.tf` files in `terraform-hyperping/monitors/`
2. Run `terraform plan` to preview changes
3. Run `terraform apply` to apply changes
4. Commit changes to Git

### Add New Monitor

```hcl
resource "hyperping_monitor" "new_service" {
  name     = "[PROD]-Service-NewAPI"
  url      = "https://api.example.com/new"
  protocol = "http"
}
```

### Emergency Disable

```bash
# Pause a monitor
terraform apply -var="monitor_name_paused=true"
```

### View Monitors

- Dashboard: https://app.hyperping.io
- Terraform State: `terraform state list | grep monitor`

## Old Pingdom Access

Pingdom account decommissioned as of YYYY-MM-DD.
All historical data exported to `pingdom-exports/` directory.
```

### Archive Pingdom Data

```bash
#!/bin/bash
# archive-pingdom-data.sh

ARCHIVE_DIR="pingdom-archive-$(date +%Y%m%d)"
mkdir -p "$ARCHIVE_DIR"

# Copy exports
cp -r pingdom-exports "$ARCHIVE_DIR/"
cp pingdom-checks.json "$ARCHIVE_DIR/"
cp pingdom-contacts.json "$ARCHIVE_DIR/"
cp pingdom-integrations.json "$ARCHIVE_DIR/"

# Export historical uptime data
echo "Exporting historical uptime data..."

check_ids=$(jq -r '.checks[].id' pingdom-checks.json)

for check_id in $check_ids; do
  echo "Exporting uptime for check $check_id..."
  curl -s -X GET \
    "https://api.pingdom.com/api/3.1/summary.average/$check_id?from=0" \
    -H "Authorization: Bearer $PINGDOM_API_TOKEN" \
    > "$ARCHIVE_DIR/uptime-${check_id}.json"

  sleep 0.5
done

# Create archive
tar -czf "${ARCHIVE_DIR}.tar.gz" "$ARCHIVE_DIR"

echo "Archive created: ${ARCHIVE_DIR}.tar.gz"
echo "Store this archive for compliance/historical reference"
```

## Troubleshooting

### Issue: Unsupported Check Types

**Problem:** Pingdom DNS or UDP checks cannot be directly migrated.

**Solution:**

For DNS checks:

```hcl
# Option 1: Monitor DNS over HTTPS service
resource "hyperping_monitor" "dns_check_via_doh" {
  name                 = "[PROD]-DNS-ExampleCom"
  url                  = "https://dns.google/resolve?name=example.com&type=A"
  protocol             = "http"
  http_method          = "GET"
  expected_status_code = "200"
  body_regex           = "example\\.com"
}

# Option 2: Monitor the service that relies on DNS
resource "hyperping_monitor" "service_check" {
  name     = "[PROD]-Service-UsingDNS"
  url      = "https://example.com/health"
  protocol = "http"
}
```

For UDP checks:

```hcl
# Alternative: Use TCP check if service also listens on TCP
resource "hyperping_monitor" "tcp_alternative" {
  name     = "[PROD]-Service-TCP"
  url      = "service.example.com"
  protocol = "port"
  port     = 53  # DNS TCP instead of UDP
}
```

### Issue: Regional Probe Differences

**Problem:** Pingdom and Hyperping have different probe locations.

**Pingdom Regions:**
- North America: Dallas, San Jose, Washington DC
- Europe: London, Stockholm, Amsterdam
- Asia: Tokyo, Sydney, Singapore
- South America: São Paulo

**Hyperping Regions:**
- North America: Virginia, Oregon
- Europe: London, Frankfurt
- Asia: Singapore, Sydney, Tokyo
- Middle East: Bahrain
- South America: São Paulo

**Solution:**

Map regions based on coverage needs:

```python
def map_pingdom_to_hyperping_regions(pingdom_filters):
    """Smart region mapping"""
    if not pingdom_filters:
        return ["virginia", "london", "singapore"]

    region_mapping = {
        # Pingdom → Hyperping
        "NA": ["virginia", "oregon"],
        "EU": ["london", "frankfurt"],
        "APAC": ["singapore", "sydney", "tokyo"],
        "LATAM": ["saopaulo"]
    }

    hyperping_regions = set()
    for pf in pingdom_filters:
        if pf.startswith('region:'):
            region = pf.split(':')[1]
            hyperping_regions.update(region_mapping.get(region, []))

    return list(hyperping_regions)
```

### Issue: Alert Integration Migration

**Problem:** Pingdom integrations (PagerDuty, Slack, email) need to be recreated.

**Solution:**

Hyperping uses webhook-based integrations. Configure in Terraform:

```hcl
# PagerDuty integration
resource "hyperping_webhook" "pagerduty" {
  name = "PagerDuty Production"
  url  = "https://events.pagerduty.com/v2/enqueue"

  headers = {
    "Content-Type"  = "application/json"
    "Authorization" = "Token token=${var.pagerduty_token}"
  }

  payload_template = jsonencode({
    event_action = "trigger"
    payload = {
      summary  = "{{monitor_name}} is {{status}}"
      severity = "error"
      source   = "Hyperping"
    }
  })
}

# Attach webhook to monitor
resource "hyperping_monitor" "critical_api" {
  name     = "[PROD]-API-Critical"
  url      = "https://api.example.com/health"
  protocol = "http"

  webhooks = [hyperping_webhook.pagerduty.id]
}
```

### Issue: Rate Limiting During Migration

**Problem:** Hitting Hyperping API rate limits when creating many monitors.

**Solution:**

Implement batching with delays:

```bash
#!/bin/bash
# apply-with-rate-limiting.sh

# Get list of monitors to create
monitors=$(terraform state list | grep hyperping_monitor || \
  grep 'resource "hyperping_monitor"' monitors.tf | \
  awk '{print $3}' | tr -d '"')

# Apply in batches of 10 with delays
batch_size=10
count=0

for monitor in $monitors; do
  echo "Creating monitor: $monitor"
  terraform apply -target="hyperping_monitor.$monitor" -auto-approve

  count=$((count + 1))

  if [ $((count % batch_size)) -eq 0 ]; then
    echo "Batch complete. Waiting 60 seconds..."
    sleep 60
  else
    sleep 5
  fi
done

echo "All monitors created"
```

### Issue: Configuration Drift After Migration

**Problem:** `terraform plan` shows changes after import.

**Solution:**

Common drift sources and fixes:

1. **Default values not matching:**

```hcl
# Explicitly set all defaults
resource "hyperping_monitor" "api" {
  name                 = "[PROD]-API-Health"
  url                  = "https://api.example.com/health"
  protocol             = "http"
  http_method          = "GET"           # Explicit
  check_frequency      = 300             # Explicit
  expected_status_code = "200"           # Explicit
  follow_redirects     = true            # Explicit
  verify_ssl           = true            # Explicit
  paused               = false           # Explicit
}
```

2. **Region ordering:**

```hcl
# Use sorted list
regions = sort([
  "virginia",
  "london",
  "frankfurt"
])
```

3. **Header ordering:**

Use consistent header order:

```hcl
request_headers = [
  {
    name  = "Authorization"
    value = var.auth_token
  },
  {
    name  = "User-Agent"
    value = "Hyperping"
  }
]
```

### Issue: Transaction Check Equivalent

**Problem:** Pingdom transaction checks don't have direct equivalent.

**Solution:**

Create external transaction checker with healthcheck:

```python
#!/usr/bin/env python3
# transaction-checker.py
# Run as scheduled job (cron/k8s cronjob)

import requests
import sys

HEALTHCHECK_URL = "https://ping.hyperping.io/hc_transaction123"

def run_transaction():
    """Simulate Pingdom transaction"""
    session = requests.Session()

    # Step 1: GET login page
    r1 = session.get("https://example.com/login")
    assert r1.status_code == 200, "Login page failed"

    # Step 2: POST login
    r2 = session.post("https://example.com/login", data={
        "username": "test",
        "password": "test123"
    })
    assert r2.status_code == 302, "Login POST failed"

    # Step 3: GET dashboard
    r3 = session.get("https://example.com/dashboard")
    assert r3.status_code == 200, "Dashboard failed"
    assert "Welcome" in r3.text, "Dashboard content missing"

    return True

if __name__ == '__main__':
    try:
        success = run_transaction()
        # Ping healthcheck on success
        requests.get(HEALTHCHECK_URL)
        sys.exit(0)
    except Exception as e:
        print(f"Transaction failed: {e}")
        # Don't ping healthcheck on failure
        sys.exit(1)
```

Deploy as Kubernetes CronJob:

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: transaction-checker
spec:
  schedule: "*/5 * * * *"  # Every 5 minutes
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: checker
            image: python:3.11
            command: ["python", "/scripts/transaction-checker.py"]
            volumeMounts:
            - name: scripts
              mountPath: /scripts
          volumes:
          - name: scripts
            configMap:
              name: transaction-scripts
          restartPolicy: OnFailure
```

Hyperping healthcheck:

```hcl
resource "hyperping_healthcheck" "transaction_check" {
  name   = "[PROD]-Transaction-LoginFlow"
  period = 300  # 5 minutes
  grace  = 60   # 1 minute grace
}

output "healthcheck_url" {
  value = hyperping_healthcheck.transaction_check.webhook_url
  sensitive = true
}
```

---

## Summary

You now have a complete migration path from Pingdom to Hyperping:

- ✅ **Export Pingdom data** - Using API to extract all checks
- ✅ **Convert check types** - Map Pingdom → Hyperping equivalents
- ✅ **Transform tags to names** - Structured naming conventions
- ✅ **Generate Terraform** - Automated configuration generation
- ✅ **Bulk import** - Scripts for large-scale migrations
- ✅ **Validation** - Pre and post-migration testing
- ✅ **Cleanup** - Decommission old Pingdom checks

**Migration Checklist:**

- [ ] Export all Pingdom checks and configuration
- [ ] Review unsupported check types (DNS, UDP, Transactions)
- [ ] Plan naming convention for Hyperping
- [ ] Generate Terraform configuration from exports
- [ ] Validate generated configuration
- [ ] Apply in test/staging environment first
- [ ] Parallel run Pingdom + Hyperping (1-2 weeks)
- [ ] Migrate alerting integrations
- [ ] Cutover to Hyperping as primary
- [ ] Archive Pingdom historical data
- [ ] Decommission Pingdom account
- [ ] Update team documentation

**Estimated Migration Times:**

- Small (1-25 checks): 1-2 days
- Medium (26-100 checks): 3-5 days
- Large (100-500 checks): 1-2 weeks
- Enterprise (500+ checks): 2-4 weeks

**Need Help?**

- [Hyperping Provider Documentation](https://registry.terraform.io/providers/develeap/hyperping)
- [Importing Resources Guide](importing-resources.md)
- [Validation Guide](validation.md)
- [Troubleshooting](../TROUBLESHOOTING.md)

Happy migrating!
