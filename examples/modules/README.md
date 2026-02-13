# Hyperping Terraform Modules

Reusable Terraform modules for common Hyperping monitoring patterns.

## Available Modules

| Module | Description | Use Case |
|--------|-------------|----------|
| [api-health](./api-health/) | API endpoint health monitoring | Monitor REST APIs, health endpoints |
| [graphql-monitor](./graphql-monitor/) | GraphQL endpoint monitoring | Monitor GraphQL APIs with query validation |
| [ssl-monitor](./ssl-monitor/) | SSL certificate monitoring | Track certificate expiration, validity |
| [database-monitor](./database-monitor/) | Database TCP port monitoring | Monitor PostgreSQL, MySQL, Redis, MongoDB connectivity |
| [cdn-monitor](./cdn-monitor/) | CDN edge location monitoring | Monitor CDN assets globally, validate cache performance |
| [cron-healthcheck](./cron-healthcheck/) | Cron job dead man's switch | Monitor scheduled jobs, backups, batch processes |
| [statuspage-complete](./statuspage-complete/) | Complete status page setup | Public status page with service monitors |
| [multi-environment](./multi-environment/) | Multi-environment deployments | Deploy same monitors across dev/staging/prod |
| [incident-management](./incident-management/) | Incident and maintenance management | Incident templates, maintenance windows, outage tracking |

## Quick Start

### API Health Monitoring

```hcl
module "api_monitors" {
  source = "path/to/modules/api-health"

  name_prefix = "PROD"
  endpoints = {
    users   = { url = "https://api.example.com/v1/users/health" }
    orders  = { url = "https://api.example.com/v1/orders/health" }
    billing = { url = "https://api.example.com/v1/billing/health" }
  }
}
```

### GraphQL Endpoint Monitoring

```hcl
module "graphql_api" {
  source = "path/to/modules/graphql-monitor"

  endpoint    = "https://api.example.com/graphql"
  name_prefix = "PROD"

  queries = {
    health = {
      query             = "{ health { status } }"
      expected_response = "\"status\":\"ok\""
    }
    users = {
      query             = "{ users { count } }"
      expected_response = "\"count\""
    }
  }
}
```

### SSL Certificate Monitoring

```hcl
module "ssl_monitors" {
  source = "path/to/modules/ssl-monitor"

  domains = [
    "api.example.com",
    "www.example.com",
    "admin.example.com"
  ]
}
```

### Database Port Monitoring

```hcl
module "databases" {
  source = "path/to/modules/database-monitor"

  name_prefix = "PROD"

  databases = {
    postgres = {
      host = "db.example.com"
      port = 5432
      type = "postgresql"
    }
    redis = {
      host = "cache.example.com"
      port = 6379
      type = "redis"
    }
    mongodb = {
      host = "docs.example.com"
      port = 27017
      type = "mongodb"
    }
  }

  default_regions = ["virginia", "london"]
}
```

### CDN Asset Monitoring

```hcl
module "cdn_monitors" {
  source = "path/to/modules/cdn-monitor"

  cdn_domain = "cdn.example.com"

  assets = {
    logo    = "/images/logo.png"
    css     = "/styles/main.css"
    js      = "/scripts/app.js"
    fonts   = "/fonts/roboto.woff2"
  }

  regions = ["virginia", "london", "singapore", "sydney", "frankfurt"]
}
```

### Cron Job Monitoring

```hcl
module "cron_jobs" {
  source = "path/to/modules/cron-healthcheck"

  jobs = {
    daily_backup = {
      cron     = "0 2 * * *"
      timezone = "America/New_York"
      grace    = 30
    }
    hourly_sync = {
      cron     = "0 * * * *"
      timezone = "UTC"
      grace    = 10
    }
  }

  name_prefix = "PROD"
}

# Use ping URLs in cron scripts
output "backup_ping_url" {
  value     = module.cron_jobs.ping_urls["daily_backup"]
  sensitive = true
}
```

### Complete Status Page

```hcl
module "status_page" {
  source = "path/to/modules/statuspage-complete"

  name      = "Acme Corp Status"
  hosted_subdomain = "acme-status"

  services = {
    api = { url = "https://api.example.com/health" }
    web = { url = "https://www.example.com" }
  }
}
```

### Multi-Environment Deployment

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

### Incident Management

```hcl
module "incidents" {
  source = "path/to/modules/incident-management"

  statuspage_id = hyperping_statuspage.main.id

  incident_templates = {
    api_degradation = {
      title    = "API Performance Degraded"
      text     = "Investigating slow API response times"
      severity = "major"
    }
  }

  maintenance_windows = {
    routine = {
      title      = "Database Maintenance"
      text       = "Routine database optimization"
      start_date = "2026-02-20T02:00:00.000Z"
      end_date   = "2026-02-20T04:00:00.000Z"
    }
  }
}
```

## Module Selection Guide

| Scenario | Recommended Module |
|----------|-------------------|
| Monitor multiple API endpoints | `api-health` |
| Monitor GraphQL APIs with query validation | `graphql-monitor` |
| Track SSL certificates for multiple domains | `ssl-monitor` |
| Monitor database TCP connectivity | `database-monitor` |
| Monitor CDN assets globally | `cdn-monitor` |
| Monitor cron jobs, backups, scheduled tasks | `cron-healthcheck` |
| Create public status page with monitors | `statuspage-complete` |
| Deploy monitors across dev/staging/prod | `multi-environment` |
| Manage incidents and maintenance windows | `incident-management` |
| Custom monitoring setup | Use provider resources directly |

## Module Composition

Modules can be combined for comprehensive monitoring:

```hcl
# Monitor internal APIs
module "internal_apis" {
  source = "path/to/modules/api-health"

  name_prefix = "INTERNAL"
  endpoints = {
    auth    = { url = "https://auth.internal.example.com/health" }
    cache   = { url = "https://cache.internal.example.com/health" }
    queue   = { url = "https://queue.internal.example.com/health" }
  }
}

# Monitor SSL certificates
module "ssl" {
  source = "path/to/modules/ssl-monitor"

  domains = [
    "api.example.com",
    "www.example.com"
  ]
}

# Public status page for customer-facing services
module "status_page" {
  source = "path/to/modules/statuspage-complete"

  name      = "Example Corp Status"
  hosted_subdomain = "status"
  hostname  = "status.example.com"

  services = {
    api = { url = "https://api.example.com/health" }
    web = { url = "https://www.example.com" }
  }

  hide_powered_by = true
}

# Incident management for status page
module "incidents" {
  source = "path/to/modules/incident-management"

  statuspage_id = module.status_page.statuspage_id

  incident_templates = {
    api_degradation = {
      title    = "API Performance Issues"
      text     = "Investigating degraded API performance"
      severity = "major"
    }
  }

  maintenance_windows = {
    weekly_maintenance = {
      title      = "Weekly Maintenance"
      text       = "Routine system maintenance and updates"
      start_date = "2026-02-16T03:00:00.000Z"
      end_date   = "2026-02-16T05:00:00.000Z"
    }
  }
}
```

## Testing Modules

Each module includes native Terraform tests. Run tests with:

```bash
cd examples/modules/api-health
terraform test

cd examples/modules/graphql-monitor
terraform test

cd examples/modules/ssl-monitor
terraform test

cd examples/modules/database-monitor
terraform test

cd examples/modules/cdn-monitor
terraform test

cd examples/modules/cron-healthcheck
terraform test

cd examples/modules/statuspage-complete
terraform test

cd examples/modules/multi-environment
terraform test

cd examples/modules/incident-management
terraform test
```

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.0 |
| hyperping | >= 1.0 |

## Contributing

When adding new modules:

1. Follow the standard structure:
   ```
   module-name/
   ├── main.tf          # Resources
   ├── variables.tf     # Input variables with validation
   ├── outputs.tf       # Output values
   ├── versions.tf      # Provider requirements
   ├── README.md        # Documentation
   └── tests/           # Native Terraform tests
       └── basic.tftest.hcl
   ```

2. Include variable validation for inputs
3. Provide comprehensive outputs
4. Document all inputs and outputs in README
5. Include usage examples
6. Add native Terraform tests
