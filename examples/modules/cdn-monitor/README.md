# CDN Monitor Module

Reusable Terraform module for CDN edge location monitoring and cache performance validation with Hyperping.

## Features

- **Global Edge Validation**: Monitor CDN assets from multiple regions worldwide
- **Static Asset Availability**: Check critical assets (images, CSS, JS, fonts)
- **Multi-Region Coverage**: Default configuration spans 5 global regions
- **Content Validation**: Optional keyword matching in response bodies
- **Performance Monitoring**: Track availability and response times across edge locations

## Usage

### Basic CDN Monitoring

```hcl
module "cdn_monitors" {
  source = "path/to/modules/cdn-monitor"

  cdn_domain = "cdn.example.com"

  assets = {
    logo    = "/images/logo.png"
    css     = "/styles/main.css"
    js      = "/scripts/app.js"
  }

  regions = ["virginia", "london", "singapore", "sydney"]
}
```

### Advanced Configuration

```hcl
module "cdn_monitors" {
  source = "path/to/modules/cdn-monitor"

  cdn_domain  = "cdn.example.com"
  name_prefix = "prod"

  assets = {
    logo    = "/images/logo.png"
    css     = "/styles/main.css"
    js      = "/scripts/app.js"
    fonts   = "/fonts/roboto.woff2"
    favicon = "/favicon.ico"
  }

  # Check from 6 global regions
  regions = ["virginia", "london", "singapore", "sydney", "frankfurt", "tokyo"]

  # Check every 5 minutes
  check_frequency = 300

  # Monitor root domain as well
  monitor_root_domain  = true
  root_expected_status = "404" # CDN root typically returns 404
}
```

### With Content Validation

```hcl
module "cdn_monitors" {
  source = "path/to/modules/cdn-monitor"

  cdn_domain = "static.example.com"

  assets = {
    html  = "/index.html"
    json  = "/api/data.json"
  }

  # Validate specific content in responses
  asset_keywords = {
    html = "<!DOCTYPE html>"
    json = "\"version\""
  }

  regions = ["virginia", "london", "singapore"]
}
```

### Custom Naming

```hcl
module "cdn_monitors" {
  source = "path/to/modules/cdn-monitor"

  cdn_domain  = "cdn.example.com"
  name_format = "[PROD-CDN] %s Asset"

  assets = {
    critical_css = "/dist/critical.css"
    app_bundle   = "/dist/app.min.js"
  }

  regions = ["virginia", "london", "frankfurt"]
}
```

### HTTP CDN (Non-HTTPS)

```hcl
module "cdn_monitors" {
  source = "path/to/modules/cdn-monitor"

  cdn_domain = "cdn-legacy.example.com"
  protocol   = "http"  # Default is https

  assets = {
    image = "/legacy/banner.jpg"
  }

  regions = ["virginia", "london"]
}
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| `cdn_domain` | CDN domain to monitor | `string` | n/a | yes |
| `assets` | Map of asset paths to monitor | `map(string)` | n/a | yes |
| `protocol` | Protocol (http or https) | `string` | `"https"` | no |
| `name_prefix` | Prefix for monitor names | `string` | `""` | no |
| `name_format` | Custom name format (use %s for key) | `string` | `""` | no |
| `regions` | Regions to check from | `list(string)` | `["virginia", "london", "singapore", "sydney", "frankfurt"]` | no |
| `check_frequency` | Check frequency (seconds) | `number` | `300` | no |
| `follow_redirects` | Follow HTTP redirects | `bool` | `false` | no |
| `asset_keywords` | Map of asset keys to required keywords | `map(string)` | `{}` | no |
| `monitor_root_domain` | Create monitor for CDN root | `bool` | `false` | no |
| `root_expected_status` | Expected status for root domain | `string` | `"2xx"` | no |
| `alerts_wait` | Seconds before alerting | `number` | `null` | no |
| `escalation_policy` | Escalation policy UUID | `string` | `null` | no |
| `paused` | Create monitors in paused state | `bool` | `false` | no |

### Assets Map

The `assets` variable is a map where:
- **Key**: Short identifier for the asset (used in monitor names)
- **Value**: Path to the asset on the CDN (must start with `/`)

Example:
```hcl
assets = {
  logo    = "/images/logo.png"          # Monitor named "CDN: logo"
  css     = "/styles/main.css"          # Monitor named "CDN: css"
  js      = "/scripts/app.js"           # Monitor named "CDN: js"
  font    = "/fonts/roboto.woff2"       # Monitor named "CDN: font"
}
```

## Outputs

| Name | Description |
|------|-------------|
| `monitor_ids` | Map of asset name to monitor UUID |
| `monitor_ids_list` | List of asset monitor UUIDs |
| `all_monitor_ids` | All monitor UUIDs (including root) |
| `monitors` | Full monitor objects for advanced usage |
| `root_monitor_id` | Root domain monitor UUID (if enabled) |
| `cdn_domain` | CDN domain being monitored |
| `asset_count` | Number of asset monitors |
| `total_monitor_count` | Total monitors (assets + root) |
| `monitored_assets` | Map of asset names to full URLs |

## Valid Regions

```
london, frankfurt, singapore, sydney, tokyo, virginia, saopaulo, bahrain
```

### Recommended Regions for CDN Monitoring

For global coverage:
- **Americas**: virginia, saopaulo
- **Europe**: london, frankfurt
- **Asia-Pacific**: singapore, tokyo, sydney
- **Middle East**: bahrain

## Valid Frequencies (seconds)

```
10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400
```

**Recommended for CDN**: 300 (5 minutes) - balances detection speed with monitoring costs

## Common Use Cases

### Frontend Application CDN

```hcl
module "app_cdn" {
  source = "path/to/modules/cdn-monitor"

  cdn_domain  = "app-cdn.example.com"
  name_prefix = "production"

  assets = {
    index     = "/index.html"
    app_js    = "/static/js/main.bundle.js"
    app_css   = "/static/css/main.css"
    vendor_js = "/static/js/vendor.bundle.js"
    fonts     = "/static/fonts/inter.woff2"
    logo      = "/static/images/logo.svg"
  }

  regions = ["virginia", "london", "singapore", "tokyo", "frankfurt"]
}
```

### Multi-CDN Setup

```hcl
module "primary_cdn" {
  source = "path/to/modules/cdn-monitor"

  cdn_domain  = "cdn1.example.com"
  name_prefix = "primary"

  assets = {
    logo = "/images/logo.png"
    css  = "/styles/main.css"
  }

  regions = ["virginia", "london", "singapore"]
}

module "backup_cdn" {
  source = "path/to/modules/cdn-monitor"

  cdn_domain  = "cdn2.example.com"
  name_prefix = "backup"

  assets = {
    logo = "/images/logo.png"
    css  = "/styles/main.css"
  }

  regions = ["virginia", "london", "singapore"]
}
```

### API CDN (JSON/Data Files)

```hcl
module "data_cdn" {
  source = "path/to/modules/cdn-monitor"

  cdn_domain = "data.example.com"

  assets = {
    config      = "/config.json"
    translations = "/i18n/en.json"
    api_schema  = "/api/schema.json"
  }

  asset_keywords = {
    config       = "\"version\""
    translations = "\"language\""
    api_schema   = "\"openapi\""
  }

  regions = ["virginia", "london", "frankfurt"]
}
```

## Best Practices

1. **Region Selection**: Choose regions that match your user base geography
2. **Check Frequency**: 300s (5 min) is optimal for CDN monitoring
3. **Content Validation**: Use `asset_keywords` for critical assets to detect corruption
4. **Root Domain**: Enable `monitor_root_domain` only if your CDN serves content at root
5. **Redirects**: Set `follow_redirects = false` for CDN assets (should be direct)
6. **Global Coverage**: Use at least 3-5 regions for meaningful CDN performance insights

## Troubleshooting

### Monitor Failing: Content Not Found

Verify asset paths:
```bash
curl https://cdn.example.com/images/logo.png
```

Ensure paths in `assets` map start with `/`.

### False Positives from Specific Regions

Some CDN edges may have delayed cache propagation. Consider:
- Reducing region count for new CDN deployments
- Increasing `alerts_wait` to allow cache propagation time

## Examples

See the `tests/` directory for additional configuration examples.
