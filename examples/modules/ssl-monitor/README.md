# SSL Monitor Module

Reusable Terraform module for SSL certificate monitoring with Hyperping.

Monitors HTTPS endpoints to detect SSL certificate issues including:
- Expired certificates
- Invalid certificate chains
- Self-signed certificates (if not trusted)
- Connection failures due to SSL/TLS issues

## Usage

### Basic

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

### With Custom Settings

```hcl
module "ssl_monitors" {
  source = "path/to/modules/ssl-monitor"

  domains = [
    "api.example.com",
    "payments.example.com"
  ]

  name_prefix     = "SSL-PROD"
  check_frequency = 1800  # 30 minutes
  regions         = ["virginia", "london", "tokyo"]
  alerts_wait     = 2     # Alert after 2 failures
}
```

### Multiple Environments

```hcl
module "ssl_prod" {
  source = "path/to/modules/ssl-monitor"

  domains     = ["api.example.com", "www.example.com"]
  name_prefix = "SSL-PROD"
}

module "ssl_staging" {
  source = "path/to/modules/ssl-monitor"

  domains         = ["api.staging.example.com"]
  name_prefix     = "SSL-STAGING"
  check_frequency = 21600  # 6 hours (less frequent for non-prod)
}
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| `domains` | List of domains to monitor | `list(string)` | n/a | yes |
| `name_prefix` | Prefix for monitor names | `string` | `"SSL"` | no |
| `check_frequency` | Check frequency in seconds | `number` | `3600` | no |
| `regions` | Monitoring regions | `list(string)` | `["virginia", "london"]` | no |
| `port` | HTTPS port | `number` | `443` | no |
| `alerts_wait` | Failed checks before alerting | `number` | `1` | no |
| `escalation_policy_uuid` | Escalation policy UUID | `string` | `null` | no |
| `paused` | Create in paused state | `bool` | `false` | no |

## Outputs

| Name | Description |
|------|-------------|
| `monitor_ids` | Map of domain to monitor UUID |
| `monitor_ids_list` | List of all monitor UUIDs |
| `monitors` | Full monitor objects |
| `monitored_domains` | List of domains being monitored |
| `monitor_count` | Total number of monitors created |

## Recommended Frequencies

| Environment | Frequency | Value |
|-------------|-----------|-------|
| Production | Hourly | `3600` |
| Staging | Every 6 hours | `21600` |
| Development | Daily | `86400` |

## Notes

- Domains should not include the protocol (`https://`)
- The module automatically constructs `https://{domain}` URLs
- SSL issues will cause monitor failures, triggering alerts
- Consider using `alerts_wait = 1` or `2` to avoid false positives from transient issues
