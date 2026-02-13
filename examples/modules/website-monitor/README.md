# Website Monitor Module

Production-ready Terraform module for comprehensive website monitoring with Hyperping. Monitors homepage availability, critical pages, validates expected content, and tracks performance thresholds.

## Features

- **Homepage & Critical Pages**: Monitor key pages like homepage, login, checkout, dashboard
- **Content Validation**: Verify expected text appears on each page
- **Performance Tracking**: Set response time thresholds per page or globally
- **Flexible Configuration**: Override defaults per page or use global settings
- **Multi-Region**: Monitor from multiple geographic locations

## Usage

### Basic Website Monitoring

```hcl
module "website" {
  source = "path/to/modules/website-monitor"

  domain = "example.com"

  pages = {
    homepage = {
      path          = "/"
      expected_text = "Welcome"
    }
    about = {
      path          = "/about"
      expected_text = "About Us"
    }
    contact = {
      path          = "/contact"
      expected_text = "Contact"
    }
  }
}
```

### E-commerce Website with Performance Thresholds

```hcl
module "ecommerce_site" {
  source = "path/to/modules/website-monitor"

  domain      = "shop.example.com"
  name_prefix = "prod"
  frequency   = 30  # Check every 30 seconds

  # Global performance threshold
  performance_threshold_ms = 2000  # Alert if response > 2 seconds

  pages = {
    homepage = {
      path          = "/"
      expected_text = "Shop Now"
    }
    login = {
      path          = "/login"
      expected_text = "Sign In"
      expected_status = "200"
    }
    checkout = {
      path          = "/checkout"
      expected_text = "Shopping Cart"
      # Critical page - stricter performance requirement
      performance_threshold_ms = 1000
    }
    products = {
      path          = "/products"
      expected_text = "Browse Products"
    }
    account = {
      path          = "/account"
      expected_text = "My Account"
    }
  }

  regions = ["virginia", "london", "singapore"]
}
```

### With Authentication Headers

```hcl
module "authenticated_site" {
  source = "path/to/modules/website-monitor"

  domain = "admin.example.com"

  default_headers = {
    "Authorization" = "Bearer ${var.admin_token}"
    "User-Agent"    = "Hyperping-Monitor"
  }

  pages = {
    dashboard = {
      path          = "/dashboard"
      expected_text = "Admin Dashboard"
    }
    users = {
      path          = "/admin/users"
      expected_text = "User Management"
    }
  }
}
```

### Mixed HTTP Methods

```hcl
module "api_website" {
  source = "path/to/modules/website-monitor"

  domain = "api.example.com"

  pages = {
    health = {
      path          = "/health"
      expected_text = "OK"
      expected_status = "200"
    }
    webhook = {
      path   = "/webhook/test"
      method = "POST"
      headers = {
        "Content-Type" = "application/json"
      }
      body          = jsonencode({ test = true })
      expected_status = "200"
    }
  }
}
```

### Staging vs Production

```hcl
# Production monitoring - frequent checks
module "website_prod" {
  source = "path/to/modules/website-monitor"

  domain      = "example.com"
  name_prefix = "PROD"
  frequency   = 30

  performance_threshold_ms = 1000

  pages = {
    homepage = { path = "/", expected_text = "Welcome" }
    login    = { path = "/login", expected_text = "Sign In" }
  }

  regions = ["virginia", "london", "singapore", "frankfurt"]
}

# Staging monitoring - less frequent
module "website_staging" {
  source = "path/to/modules/website-monitor"

  domain      = "staging.example.com"
  name_prefix = "STAGING"
  frequency   = 300  # Every 5 minutes

  pages = {
    homepage = { path = "/", expected_text = "Welcome" }
    login    = { path = "/login", expected_text = "Sign In" }
  }

  regions = ["virginia"]
}
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| `domain` | Domain name to monitor | `string` | n/a | yes |
| `pages` | Map of page configurations | `map(object)` | n/a | yes |
| `protocol` | Protocol (http or https) | `string` | `"https"` | no |
| `name_prefix` | Prefix for monitor names | `string` | `""` | no |
| `name_format` | Custom name format (use %s for key) | `string` | `""` | no |
| `frequency` | Default check frequency (seconds) | `number` | `60` | no |
| `regions` | Default monitoring regions | `list(string)` | `["virginia", "london", "frankfurt"]` | no |
| `default_method` | Default HTTP method | `string` | `"GET"` | no |
| `default_expected_status` | Default expected status | `string` | `"2xx"` | no |
| `follow_redirects` | Follow HTTP redirects | `bool` | `true` | no |
| `performance_threshold_ms` | Global performance threshold (milliseconds) | `number` | `null` | no |
| `default_headers` | Default HTTP headers for all pages | `map(string)` | `null` | no |
| `alerts_wait` | Seconds before alerting | `number` | `null` | no |
| `escalation_policy` | Escalation policy UUID | `string` | `null` | no |
| `paused` | Create monitors in paused state | `bool` | `false` | no |

### Page Object

Each page in the `pages` map supports:

| Field | Description | Type | Default |
|-------|-------------|------|---------|
| `path` | URL path (must start with /) | `string` | required |
| `expected_text` | Text that must appear in response | `string` | `null` |
| `expected_status` | Expected HTTP status | `string` | uses `default_expected_status` |
| `method` | HTTP method | `string` | uses `default_method` |
| `frequency` | Check frequency (seconds) | `number` | uses `frequency` |
| `performance_threshold_ms` | Response time threshold | `number` | uses `performance_threshold_ms` |
| `follow_redirects` | Follow redirects | `bool` | uses `follow_redirects` |
| `headers` | Custom headers | `map(string)` | uses `default_headers` |
| `body` | Request body | `string` | `null` |
| `regions` | Override regions | `list(string)` | uses `regions` |
| `paused` | Override paused state | `bool` | uses `paused` |

## Outputs

| Name | Description |
|------|-------------|
| `monitor_ids` | Map of page name to monitor UUID |
| `monitor_ids_list` | List of all monitor UUIDs |
| `monitors` | Full monitor objects with details |
| `page_count` | Total number of page monitors |
| `domain` | Domain being monitored |
| `pages_monitored` | List of page paths monitored |
| `performance_monitoring_enabled` | Whether performance thresholds are set |

## Valid Regions

```
paris, frankfurt, amsterdam, london, singapore, sydney, tokyo, seoul,
mumbai, bangalore, virginia, california, sanfrancisco, nyc, toronto,
saopaulo, bahrain, capetown
```

## Valid Frequencies (seconds)

```
10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400
```

## Common Patterns

### Critical E-commerce Pages

```hcl
pages = {
  homepage = {
    path          = "/"
    expected_text = "Welcome"
  }
  products = {
    path          = "/products"
    expected_text = "Browse"
  }
  cart = {
    path          = "/cart"
    expected_text = "Shopping Cart"
  }
  checkout = {
    path          = "/checkout"
    expected_text = "Payment"
    performance_threshold_ms = 1000  # Critical page
  }
  login = {
    path          = "/login"
    expected_text = "Sign In"
  }
  register = {
    path          = "/register"
    expected_text = "Create Account"
  }
}
```

### SaaS Application Pages

```hcl
pages = {
  homepage = {
    path          = "/"
    expected_text = "Get Started"
  }
  login = {
    path          = "/login"
    expected_text = "Sign In"
  }
  dashboard = {
    path          = "/dashboard"
    expected_text = "Dashboard"
  }
  settings = {
    path          = "/settings"
    expected_text = "Settings"
  }
  billing = {
    path          = "/billing"
    expected_text = "Subscription"
  }
}
```

### Marketing Website Pages

```hcl
pages = {
  homepage = {
    path          = "/"
    expected_text = "Welcome"
  }
  about = {
    path          = "/about"
    expected_text = "About Us"
  }
  services = {
    path          = "/services"
    expected_text = "Our Services"
  }
  contact = {
    path          = "/contact"
    expected_text = "Contact Us"
  }
  blog = {
    path          = "/blog"
    expected_text = "Latest Posts"
  }
}
```

## Testing

See the `tests/` directory for example test configurations.

```bash
# Run module tests
cd tests
terraform init
terraform test
```

## Related Modules

- **api-health**: For REST API endpoint monitoring
- **ssl-monitor**: For SSL certificate monitoring
- **statuspage-complete**: For status page creation with website monitors
