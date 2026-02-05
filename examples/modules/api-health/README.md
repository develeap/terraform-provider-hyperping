# API Health Module

Reusable Terraform module for HTTP/HTTPS API health monitoring with Hyperping.

## Usage

### Basic

```hcl
module "api_monitors" {
  source = "path/to/modules/api-health"

  endpoints = {
    "main-api" = {
      url = "https://api.example.com/health"
    }
    "payment-api" = {
      url       = "https://payments.example.com/ping"
      frequency = 30
    }
  }
}
```

### With Authentication

```hcl
module "api_monitors" {
  source = "path/to/modules/api-health"

  endpoints = {
    "authenticated-api" = {
      url     = "https://api.example.com/health"
      headers = {
        "Authorization" = "Bearer ${var.api_token}"
        "Accept"        = "application/json"
      }
    }
  }

  name_prefix     = "prod"
  default_regions = ["virginia", "london", "tokyo"]
}
```

### POST Request with Body

```hcl
module "api_monitors" {
  source = "path/to/modules/api-health"

  endpoints = {
    "webhook-test" = {
      url    = "https://api.example.com/webhook"
      method = "POST"
      headers = {
        "Content-Type" = "application/json"
      }
      body = jsonencode({ test = true })
    }
  }
}
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| `endpoints` | Map of endpoint configurations | `map(object)` | n/a | yes |
| `name_prefix` | Prefix for monitor names | `string` | `""` | no |
| `default_regions` | Default monitoring regions | `list(string)` | `["virginia", "london", "frankfurt"]` | no |
| `default_frequency` | Default check frequency (seconds) | `number` | `60` | no |
| `alerts_wait` | Failed checks before alerting | `number` | `0` | no |
| `escalation_policy_uuid` | Escalation policy UUID | `string` | `null` | no |

### Endpoint Object

| Field | Description | Type | Default |
|-------|-------------|------|---------|
| `url` | URL to monitor | `string` | required |
| `method` | HTTP method | `string` | `"GET"` |
| `frequency` | Check frequency (seconds) | `number` | uses `default_frequency` |
| `expected_status_code` | Expected HTTP status | `string` | `"200"` |
| `follow_redirects` | Follow HTTP redirects | `bool` | `true` |
| `headers` | Request headers | `map(string)` | `{}` |
| `body` | Request body | `string` | `null` |
| `required_keyword` | Keyword that must appear in response | `string` | `null` |
| `regions` | Override default regions | `list(string)` | uses `default_regions` |
| `paused` | Create monitor in paused state | `bool` | `false` |

## Outputs

| Name | Description |
|------|-------------|
| `monitor_ids` | Map of endpoint name to monitor UUID |
| `monitor_names` | Map of endpoint name to full monitor name |
| `monitor_urls` | Map of endpoint name to monitored URL |
| `monitors` | Full monitor objects for advanced usage |
| `active_monitors` | List of active (non-paused) monitor IDs |
| `monitor_count` | Total number of monitors created |

## Valid Regions

`virginia`, `london`, `frankfurt`, `singapore`, `sydney`, `tokyo`, `saopaulo`, `oregon`, `bahrain`

## Valid Frequencies

`10`, `20`, `30`, `60`, `120`, `180`, `300`, `600`, `1800`, `3600`, `21600`, `43200`, `86400` (seconds)
