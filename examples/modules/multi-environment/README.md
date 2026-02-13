# Multi-Environment Monitoring Module

Deploy identical monitors across multiple environments (dev, staging, prod) with environment-specific configuration for frequency, regions, and alerts.

## Features

- **Environment Parity**: Ensure consistent monitoring across dev, staging, and production
- **Environment-Specific Configuration**: Customize frequency, regions, and alerts per environment
- **Terraform Workspace Integration**: Optional workspace-based naming for workspace-per-environment patterns
- **Consistent Naming**: Automatic environment prefix with configurable format
- **Selective Deployment**: Enable/disable specific environments without removing configuration
- **DRY Principles**: Define common defaults once, override only where needed

## Usage

### Basic Example

```hcl
module "api_monitoring" {
  source = "path/to/modules/multi-environment"

  service_name = "UserAPI"

  environments = {
    dev = {
      url       = "https://dev-api.example.com/health"
      frequency = 300
      regions   = ["virginia"]
    }
    staging = {
      url       = "https://staging-api.example.com/health"
      frequency = 120
      regions   = ["virginia", "london"]
    }
    prod = {
      url       = "https://api.example.com/health"
      frequency = 60
      regions   = ["virginia", "london", "singapore"]
    }
  }
}
```

Creates monitors:
- `[DEV] UserAPI` - checks every 5 minutes from Virginia
- `[STAGING] UserAPI` - checks every 2 minutes from Virginia and London
- `[PROD] UserAPI` - checks every minute from Virginia, London, and Singapore

### Advanced Configuration

```hcl
module "payment_api" {
  source = "path/to/modules/multi-environment"

  service_name = "PaymentService"
  name_format  = "[%s] %s Health Check"

  # Defaults applied to all environments
  default_method               = "POST"
  default_expected_status_code = "200"
  default_alerts_wait          = 60
  default_headers = {
    "Content-Type" = "application/json"
    "X-API-Key"    = var.api_key
  }
  default_body = jsonencode({
    action = "health_check"
  })

  environments = {
    dev = {
      url       = "https://dev-payments.example.com/health"
      frequency = 300
      regions   = ["virginia"]
      paused    = false  # Active during development
    }
    staging = {
      url              = "https://staging-payments.example.com/health"
      frequency        = 120
      regions          = ["virginia", "london"]
      required_keyword = "healthy"
      alerts_wait      = 30  # Faster alerts for staging
    }
    prod = {
      url               = "https://payments.example.com/health"
      frequency         = 30
      regions           = ["virginia", "london", "singapore", "sydney"]
      alerts_wait       = 0  # Immediate alerts for production
      escalation_policy = var.prod_escalation_policy_id
    }
  }
}
```

### Workspace Integration

When using Terraform workspaces for environment management:

```hcl
# Use workspace name (e.g., "dev", "staging", "prod")
module "api_monitoring" {
  source = "path/to/modules/multi-environment"

  service_name        = "UserAPI"
  use_workspace_name  = true

  environments = {
    current = {  # Single environment for current workspace
      url       = var.api_url  # Set per workspace
      frequency = var.check_frequency
      regions   = var.regions
    }
  }
}
```

### Selective Environment Deployment

```hcl
module "api_monitoring" {
  source = "path/to/modules/multi-environment"

  service_name = "UserAPI"

  environments = {
    dev = {
      url     = "https://dev-api.example.com/health"
      enabled = true
    }
    staging = {
      url     = "https://staging-api.example.com/health"
      enabled = true
    }
    prod = {
      url     = "https://api.example.com/health"
      enabled = false  # Temporarily disable without removing config
    }
  }
}
```

### Multi-Service Deployment

```hcl
# User API
module "user_api" {
  source = "path/to/modules/multi-environment"

  service_name = "UserAPI"
  environments = {
    dev     = { url = "https://dev-users.example.com/health" }
    staging = { url = "https://staging-users.example.com/health" }
    prod    = { url = "https://users.example.com/health" }
  }
}

# Order API
module "order_api" {
  source = "path/to/modules/multi-environment"

  service_name = "OrderAPI"
  environments = {
    dev     = { url = "https://dev-orders.example.com/health" }
    staging = { url = "https://staging-orders.example.com/health" }
    prod    = { url = "https://orders.example.com/health" }
  }
}

# Payment API
module "payment_api" {
  source = "path/to/modules/multi-environment"

  service_name = "PaymentAPI"
  environments = {
    dev     = { url = "https://dev-payments.example.com/health" }
    staging = { url = "https://staging-payments.example.com/health" }
    prod    = { url = "https://payments.example.com/health" }
  }
}

# Create incidents affecting all production services
resource "hyperping_incident" "prod_outage" {
  title   = "Production API Degradation"
  message = "Investigating elevated error rates across production services"
  status  = "investigating"

  monitor_uuids = concat(
    module.user_api.production_monitor_ids,
    module.order_api.production_monitor_ids,
    module.payment_api.production_monitor_ids
  )
}
```

## Environment-Specific Configuration Examples

### Development Environment

```hcl
dev = {
  url       = "https://dev-api.example.com/health"
  frequency = 300     # 5 minutes - less frequent checks
  regions   = ["virginia"]  # Single region
  paused    = false   # May be paused outside business hours
}
```

### Staging Environment

```hcl
staging = {
  url              = "https://staging-api.example.com/health"
  frequency        = 120     # 2 minutes
  regions          = ["virginia", "london"]  # Multi-region for realistic testing
  required_keyword = "healthy"  # Validate response content
  alerts_wait      = 60      # 1 minute grace period
}
```

### Production Environment

```hcl
prod = {
  url               = "https://api.example.com/health"
  frequency         = 30      # 30 seconds - frequent checks
  regions           = ["virginia", "london", "singapore", "sydney"]  # Global coverage
  alerts_wait       = 0       # Immediate alerts
  escalation_policy = "pol_abc123"  # Production escalation
}
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| service_name | Name of the service being monitored | string | - | yes |
| environments | Map of environment configurations | map(object) | - | yes |
| name_format | Custom name format string (first %s = environment, second %s = service_name) | string | "" | no |
| use_workspace_name | Use Terraform workspace name instead of environment key | bool | false | no |
| default_method | Default HTTP method for all environments | string | "GET" | no |
| default_frequency | Default check frequency in seconds | number | 60 | no |
| default_regions | Default monitoring regions | list(string) | ["virginia", "london", "frankfurt"] | no |
| default_expected_status_code | Default expected HTTP status code | string | "2xx" | no |
| default_follow_redirects | Default redirect following behavior | bool | true | no |
| default_headers | Default HTTP headers | map(string) | null | no |
| default_body | Default request body | string | null | no |
| default_required_keyword | Default required keyword in response | string | null | no |
| default_alerts_wait | Default seconds to wait before alerting | number | null | no |
| default_escalation_policy | Default escalation policy UUID | string | null | no |
| default_paused | Default paused state | bool | false | no |

### Environment Object Schema

```hcl
{
  url                  = string           # Required: Environment-specific URL
  method               = optional(string) # HTTP method (GET, POST, etc.)
  frequency            = optional(number) # Check frequency in seconds
  regions              = optional(list(string)) # Monitoring regions
  expected_status_code = optional(string) # Expected HTTP status code
  follow_redirects     = optional(bool)   # Follow HTTP redirects
  headers              = optional(map(string)) # HTTP headers
  body                 = optional(string) # Request body (for POST/PUT/PATCH)
  required_keyword     = optional(string) # Required keyword in response
  alerts_wait          = optional(number) # Seconds to wait before alerting
  escalation_policy    = optional(string) # Escalation policy UUID
  paused               = optional(bool)   # Monitor paused state
  enabled              = optional(bool)   # Deploy this environment (default: true)
}
```

## Outputs

| Name | Description |
|------|-------------|
| monitor_ids | Map of environment name to monitor UUID |
| monitor_ids_list | List of all monitor UUIDs |
| monitors | Full monitor objects with details |
| environments | List of enabled environment names |
| environment_count | Total number of monitors created |
| service_name | Service name being monitored |
| by_environment | Environment-specific details organized by name |
| production_monitor_ids | Monitor IDs for production environments |
| non_production_monitor_ids | Monitor IDs for non-production environments |

## Common Patterns

### Pattern 1: Progressive Deployment

```hcl
module "new_service" {
  source = "path/to/modules/multi-environment"

  service_name = "NewService"

  environments = {
    dev = {
      url     = "https://dev-new.example.com/health"
      enabled = true  # Start with dev
    }
    staging = {
      url     = "https://staging-new.example.com/health"
      enabled = false  # Enable after dev validation
    }
    prod = {
      url     = "https://new.example.com/health"
      enabled = false  # Enable after staging validation
    }
  }
}
```

### Pattern 2: Environment-Specific Escalations

```hcl
module "critical_api" {
  source = "path/to/modules/multi-environment"

  service_name = "CriticalAPI"

  environments = {
    dev = {
      url               = "https://dev-critical.example.com/health"
      escalation_policy = null  # No escalation for dev
    }
    staging = {
      url               = "https://staging-critical.example.com/health"
      escalation_policy = var.staging_escalation_id
    }
    prod = {
      url               = "https://critical.example.com/health"
      escalation_policy = var.prod_escalation_id
    }
  }
}
```

### Pattern 3: Canary Deployments

```hcl
module "api_canary" {
  source = "path/to/modules/multi-environment"

  service_name = "API-Canary"

  environments = {
    prod-stable = {
      url     = "https://api.example.com/health"
      regions = ["virginia", "london", "singapore"]
    }
    prod-canary = {
      url              = "https://canary.api.example.com/health"
      regions          = ["virginia"]  # Single region for canary
      required_keyword = "canary-ok"
      alerts_wait      = 120  # More lenient for canary
    }
  }
}
```

## Best Practices

1. **Consistent URLs**: Use predictable URL patterns (e.g., `{env}-{service}.example.com`)
2. **Progressive Monitoring**: Use less frequent checks for dev, more frequent for prod
3. **Regional Strategy**: Single region for dev, multi-region for staging/prod
4. **Alert Tuning**: Longer wait times for non-prod to reduce noise
5. **Escalation Policies**: Different policies per environment criticality
6. **Required Keywords**: Validate response content in staging/prod
7. **Documentation**: Document environment-specific URLs and settings

## Testing

Run native Terraform tests:

```bash
cd examples/modules/multi-environment
terraform test
```

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.0 |
| hyperping | >= 1.0 |

## Related Modules

- [api-health](../api-health/) - Monitor multiple endpoints in a single environment
- [ssl-monitor](../ssl-monitor/) - Monitor SSL certificates
- [statuspage-complete](../statuspage-complete/) - Complete status page with monitors

## License

Same as the Terraform provider.
