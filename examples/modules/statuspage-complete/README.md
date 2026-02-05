# Status Page Complete Module

Comprehensive Terraform module for creating a fully configured Hyperping status page with associated service monitors.

This module creates:
- A public status page with customizable branding
- Individual monitors for each service
- Automatic association of monitors to the status page

## Usage

### Basic

```hcl
module "status_page" {
  source = "path/to/modules/statuspage-complete"

  name      = "Acme Corp Status"
  subdomain = "acme-status"

  services = {
    api = {
      url         = "https://api.example.com/health"
      description = "Core API"
    }
    web = {
      url         = "https://www.example.com"
      description = "Web Application"
    }
  }
}
```

### With Custom Branding

```hcl
module "status_page" {
  source = "path/to/modules/statuspage-complete"

  name      = "Acme Corp Status"
  subdomain = "status"
  hostname  = "status.example.com"  # Custom domain

  services = {
    api = { url = "https://api.example.com/health" }
    web = { url = "https://www.example.com" }
    cdn = { url = "https://cdn.example.com/health" }
  }

  # Branding
  theme        = "dark"
  accent_color = "#3B82F6"  # Blue

  # Features
  hide_powered_by      = true
  enable_subscriptions = true

  # Multi-language
  languages = ["en", "es", "fr"]

  # Monitoring regions
  regions = ["virginia", "london", "singapore", "sydney"]
}
```

### Multi-Environment Setup

```hcl
module "status_prod" {
  source = "path/to/modules/statuspage-complete"

  name      = "Production Status"
  subdomain = "status"
  hostname  = "status.example.com"

  services = {
    api     = { url = "https://api.example.com/health" }
    web     = { url = "https://www.example.com" }
    billing = { url = "https://billing.example.com/health" }
  }

  hide_powered_by = true
}

module "status_staging" {
  source = "path/to/modules/statuspage-complete"

  name      = "Staging Status"
  subdomain = "status-staging"

  services = {
    api = {
      url       = "https://api.staging.example.com/health"
      frequency = 300  # Less frequent for non-prod
    }
    web = {
      url       = "https://staging.example.com"
      frequency = 300
    }
  }
}
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| `name` | Status page name | `string` | n/a | yes |
| `subdomain` | Subdomain for hosted status page | `string` | n/a | yes |
| `hostname` | Custom domain for status page | `string` | `null` | no |
| `services` | Map of services to monitor | `map(object)` | n/a | yes |
| `theme` | Status page theme | `string` | `"system"` | no |
| `accent_color` | Accent color (hex format) | `string` | `"#36b27e"` | no |
| `languages` | Supported languages | `list(string)` | `["en"]` | no |
| `regions` | Monitoring regions | `list(string)` | `["virginia", "london", "frankfurt"]` | no |
| `hide_powered_by` | Hide branding | `bool` | `false` | no |
| `enable_subscriptions` | Allow status subscriptions | `bool` | `true` | no |

### Service Object

Each service in the `services` map accepts:

| Field | Description | Type | Default |
|-------|-------------|------|---------|
| `url` | Service URL to monitor | `string` | required |
| `description` | Service description | `string` | `""` |
| `method` | HTTP method | `string` | `"GET"` |
| `frequency` | Check frequency (seconds) | `number` | `60` |
| `expected_status_code` | Expected response code | `string` | `"200"` |
| `headers` | Custom request headers | `map(string)` | `{}` |

## Outputs

| Name | Description |
|------|-------------|
| `statuspage_id` | Status page UUID |
| `statuspage_url` | Public URL for the status page |
| `statuspage_subdomain` | Status page subdomain |
| `monitor_ids` | Map of service name to monitor UUID |
| `monitor_ids_list` | List of all monitor UUIDs |
| `monitors` | Full monitor objects with details |
| `service_count` | Number of services monitored |

## Custom Domain Setup

To use a custom domain:

1. Set the `hostname` variable to your domain
2. Create a CNAME record pointing to `hyperping.app`
3. Wait for DNS propagation and SSL certificate provisioning

```hcl
module "status_page" {
  source = "path/to/modules/statuspage-complete"

  name      = "My Company Status"
  subdomain = "mycompany"
  hostname  = "status.mycompany.com"  # Point CNAME to hyperping.app

  services = {
    main = { url = "https://api.mycompany.com/health" }
  }
}
```

## Available Themes

| Theme | Description |
|-------|-------------|
| `light` | Light background |
| `dark` | Dark background |
| `system` | Follows user's system preference |

## Available Regions

```
virginia, london, frankfurt, singapore, sydney, tokyo, saopaulo, oregon, bahrain
```

## Notes

- At least one service must be specified
- Subdomain must be lowercase alphanumeric with hyphens only
- Accent color must be in hex format (e.g., `#36b27e`)
- Custom hostname requires DNS CNAME configuration
