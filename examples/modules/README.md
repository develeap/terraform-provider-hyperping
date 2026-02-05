# Hyperping Terraform Modules

Reusable Terraform modules for common Hyperping monitoring patterns.

## Available Modules

| Module | Description | Use Case |
|--------|-------------|----------|
| [api-health](./api-health/) | API endpoint health monitoring | Monitor REST APIs, health endpoints |
| [ssl-monitor](./ssl-monitor/) | SSL certificate monitoring | Track certificate expiration, validity |
| [statuspage-complete](./statuspage-complete/) | Complete status page setup | Public status page with service monitors |

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

## Module Selection Guide

| Scenario | Recommended Module |
|----------|-------------------|
| Monitor multiple API endpoints | `api-health` |
| Track SSL certificates for multiple domains | `ssl-monitor` |
| Create public status page with monitors | `statuspage-complete` |
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
```

## Testing Modules

Each module includes native Terraform tests. Run tests with:

```bash
cd examples/modules/api-health
terraform test

cd examples/modules/ssl-monitor
terraform test

cd examples/modules/statuspage-complete
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
