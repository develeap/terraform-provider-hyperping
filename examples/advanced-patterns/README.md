# Advanced Patterns - Hyperping Terraform Provider

This example demonstrates advanced patterns and best practices for production monitoring with Hyperping.

## Patterns Demonstrated

### 1. Dynamic Monitor Creation

Create monitors dynamically from a data structure:

```hcl
variable "services" {
  type = map(object({
    url             = string
    method          = string
    frequency       = number
    expected_status = string
    critical        = bool
    regions         = list(string)
  }))
}

resource "hyperping_monitor" "services" {
  for_each = var.services

  name            = "${local.name_prefix} ${title(each.key)} Service"
  url             = each.value.url
  check_frequency = each.value.frequency
  # ... other attributes
}
```

**Benefits:**
- Single source of truth for service definitions
- Easy to add/remove services
- Consistent naming and configuration
- DRY (Don't Repeat Yourself)

### 2. Regional Redundancy

Monitor multiple regional deployments:

```hcl
resource "hyperping_monitor" "regional_primary" {
  url     = "https://us-east.api.example.com/health"
  regions = ["virginia"]
}

resource "hyperping_monitor" "regional_secondary" {
  url     = "https://eu-west.api.example.com/health"
  regions = ["london"]
}
```

**Use Cases:**
- Multi-region active-active deployments
- Failover detection
- Regional performance monitoring
- Geo-distributed systems

### 3. Multi-Protocol Monitoring

Combine HTTP, TCP, and ICMP checks:

```hcl
# HTTP endpoint
resource "hyperping_monitor" "http_endpoint" {
  protocol   = "http"
  url        = "https://api.example.com"
}

# TCP port (database)
resource "hyperping_monitor" "database_port" {
  protocol = "port"
  url      = "tcp://db.example.com:5432"
  port     = 5432
}
```

**Use Cases:**
- Full stack monitoring (web + database + cache)
- Infrastructure health checks
- Network connectivity validation

### 4. Criticality-Based Monitoring

Separate critical and standard services:

```hcl
locals {
  critical_services = {
    for k, v in var.services : k => v if v.critical
  }

  standard_services = {
    for k, v in var.services : k => v if !v.critical
  }
}

# Higher frequency for critical services
resource "hyperping_monitor" "services" {
  check_frequency = each.value.critical ? 30 : 300
}
```

**Benefits:**
- Optimize monitoring costs
- Prioritize critical systems
- Flexible alerting strategies
- Resource allocation

### 5. Conditional Resource Creation

Create resources based on configuration:

```hcl
# Only create incident if status page configured
resource "hyperping_incident" "degradation" {
  count = var.status_page_id != "" ? 1 : 0

  status_pages = [var.status_page_id]
}

# Only affect critical services during maintenance
resource "hyperping_maintenance" "upgrade" {
  monitors = [
    for k, v in hyperping_monitor.services : v.id
    if var.services[k].critical
  ]
}
```

**Use Cases:**
- Environment-specific configurations
- Feature flags
- Phased rollouts
- Cost optimization

### 6. Smart Outputs and Analysis

Extract insights from your monitoring setup:

```hcl
# Group by frequency
output "monitor_by_frequency" {
  value = {
    high_frequency = [
      for k, v in hyperping_monitor.services : v
      if v.check_frequency <= 60
    ]
  }
}

# Regional coverage
output "regional_coverage" {
  value = distinct(flatten([
    for k, v in hyperping_monitor.services : v.regions
  ]))
}

# Cost estimation
output "estimated_monthly_checks" {
  value = sum([
    for k, v in hyperping_monitor.services :
    (2592000 / v.check_frequency) * length(v.regions)
  ])
}
```

## Configuration

### Basic Setup

```bash
export HYPERPING_API_KEY="sk_your_api_key"
```

### Custom Variables

Create `terraform.tfvars`:

```hcl
environment = "production"

services = {
  api = {
    url             = "https://api.example.com/health"
    method          = "GET"
    frequency       = 30
    expected_status = "200"
    critical        = true
    regions         = ["london", "virginia", "singapore", "tokyo"]
  }
  website = {
    url             = "https://www.example.com"
    method          = "GET"
    frequency       = 60
    expected_status = "2xx"
    critical        = false
    regions         = ["london", "virginia"]
  }
  admin = {
    url             = "https://admin.example.com"
    method          = "GET"
    frequency       = 300
    expected_status = "200"
    critical        = false
    regions         = ["virginia"]
  }
}

status_page_id = "sp_your_status_page_uuid"
```

### Environment-Specific Configurations

```bash
# Development
terraform workspace select development
terraform apply -var-file="dev.tfvars"

# Staging
terraform workspace select staging
terraform apply -var-file="staging.tfvars"

# Production
terraform workspace select production
terraform apply -var-file="production.tfvars"
```

## Usage

### Initialize

```bash
terraform init
```

### Plan

```bash
terraform plan -out=plan.tfplan
```

### Apply

```bash
terraform apply plan.tfplan
```

### Verify Outputs

```bash
terraform output environment_summary
terraform output critical_monitors
terraform output regional_coverage
```

## Example Outputs

```hcl
environment_summary = {
  critical_count = 1
  environment    = "production"
  standard_count = 2
  status_page_id = "sp_abc123"
  total_monitors = 9
}

service_monitors = {
  api = {
    critical  = true
    frequency = 30
    id        = "mon_api123"
    name      = "[PRODUCTION] Api Service"
    regions   = ["london", "virginia", "singapore", "tokyo"]
    url       = "https://api.example.com/health"
  }
  website = {
    critical  = false
    frequency = 60
    id        = "mon_web456"
    name      = "[PRODUCTION] Website Service"
    regions   = ["london", "virginia"]
    url       = "https://www.example.com"
  }
}

critical_monitors = [
  "mon_api123"
]

regional_monitors = {
  asia_pacific = "mon_ap789"
  eu_west      = "mon_eu456"
  us_east      = "mon_us123"
}

monitor_by_frequency = {
  high_frequency = [
    { frequency = 30, name = "[PRODUCTION] Api Service" },
    { frequency = 30, name = "[PRODUCTION] Regional Check - US East" }
  ]
  low_frequency = [
    { frequency = 300, name = "[PRODUCTION] Admin Service" }
  ]
  standard_frequency = [
    { frequency = 60, name = "[PRODUCTION] Website Service" }
  ]
}

regional_coverage = [
  "london",
  "singapore",
  "tokyo",
  "virginia"
]

estimated_monthly_checks = 14976000

maintenance_window = {
  affected_monitors    = 1
  end_date             = "2026-02-15T04:00:00.000Z"
  id                   = "mw_upgrade123"
  name                 = "critical-services-maintenance-production"
  notification_minutes = 120
  start_date           = "2026-02-15T02:00:00.000Z"
}

monitor_urls_by_criticality = {
  critical = ["https://api.example.com/health"]
  standard = ["https://www.example.com", "https://admin.example.com"]
}
```

## Best Practices

### 1. Use Locals for Computed Values

```hcl
locals {
  name_prefix = "[${upper(var.environment)}]"

  critical_services = {
    for k, v in var.services : k => v if v.critical
  }
}
```

### 2. Validate Input Variables

```hcl
variable "services" {
  validation {
    condition = alltrue([
      for k, v in var.services :
      contains([30, 60, 120, 300], v.frequency)
    ])
    error_message = "Check frequencies must be 30, 60, 120, or 300 seconds."
  }
}
```

### 3. Document Your Patterns

Add clear comments explaining complex logic:

```hcl
# Use service-specific regions or defaults based on criticality
regions = length(each.value.regions) > 0 ? each.value.regions : (
  each.value.critical ? local.tier1_regions : local.tier2_regions
)
```

### 4. Use Meaningful Outputs

Provide insights, not just raw data:

```hcl
output "estimated_monthly_checks" {
  description = "Estimated total checks per month for cost planning"
  value       = sum([...])
}
```

### 5. Organize by Pattern

Group resources by their pattern or purpose:

```hcl
# =============================================================================
# Pattern 1: Dynamic Monitor Creation
# =============================================================================

# =============================================================================
# Pattern 2: Regional Redundancy
# =============================================================================
```

## Real-World Use Cases

### Microservices Monitoring

```hcl
services = {
  auth_service    = { url = "...", critical = true, frequency = 30 }
  user_service    = { url = "...", critical = true, frequency = 30 }
  payment_service = { url = "...", critical = true, frequency = 10 }
  email_service   = { url = "...", critical = false, frequency = 300 }
  analytics       = { url = "...", critical = false, frequency = 600 }
}
```

### Multi-Region Deployment

```hcl
# Primary region (US East)
us_east = {
  api      = "https://us-east.api.example.com"
  database = "tcp://us-east.db.example.com:5432"
}

# DR region (EU West)
eu_west = {
  api      = "https://eu-west.api.example.com"
  database = "tcp://eu-west.db.example.com:5432"
}
```

### Staged Rollout Monitoring

```hcl
# Canary deployment
canary_monitors = var.enable_canary ? {
  canary_v2 = {
    url       = "https://canary.api.example.com/v2/health"
    critical  = false
    frequency = 30
  }
} : {}
```

## Troubleshooting

### Too Many Monitors

If you exceed your plan limits:

```hcl
# Reduce frequency for non-critical services
locals {
  optimized_frequency = each.value.critical ? 60 : 600
}
```

### Regional Rate Limits

Spread checks across regions:

```hcl
# Use different regions for each service
locals {
  region_rotation = {
    service1 = ["london", "virginia"]
    service2 = ["frankfurt", "singapore"]
    service3 = ["oregon", "tokyo"]
  }
}
```

### Dynamic Updates

Use `terraform apply -target` for specific resources:

```bash
# Update only critical services
terraform apply -target='hyperping_monitor.services["api"]'
```

## Next Steps

1. Adapt patterns to your infrastructure
2. Add authentication for private endpoints
3. Integrate with your status page
4. Set up maintenance windows
5. Configure alerting integrations

## Resources

- [Main README](../../README.md)
- [Complete Example](../complete/)
- [Multi-Tenant Example](../multi-tenant/)
- [Provider Documentation](https://registry.terraform.io/providers/develeap/hyperping/latest/docs)
