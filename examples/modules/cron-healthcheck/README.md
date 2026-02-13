# Cron Healthcheck Module

Reusable Terraform module for dead man's switch monitoring of cron jobs with Hyperping.

This module creates healthchecks that expect periodic pings from your cron jobs. If a ping is not received within the expected schedule plus grace period, an alert is triggered.

## Usage

### Basic

```hcl
module "cron_jobs" {
  source = "path/to/modules/cron-healthcheck"

  jobs = {
    daily_backup = {
      cron     = "0 2 * * *"        # 2 AM every day
      timezone = "America/New_York"
      grace    = 30                  # 30 minutes grace period
    }
    hourly_sync = {
      cron     = "0 * * * *"        # Every hour
      timezone = "UTC"
      grace    = 10                  # 10 minutes grace period
    }
  }
}

# Output ping URLs for use in cron scripts
output "backup_ping_url" {
  value     = module.cron_jobs.ping_urls["daily_backup"]
  sensitive = true
}

output "sync_ping_url" {
  value     = module.cron_jobs.ping_urls["hourly_sync"]
  sensitive = true
}
```

### With Name Prefix

```hcl
module "prod_cron_jobs" {
  source = "path/to/modules/cron-healthcheck"

  jobs = {
    db_backup = {
      cron     = "0 3 * * *"
      timezone = "UTC"
      grace    = 60
    }
    log_rotation = {
      cron     = "0 4 * * 0"        # Weekly on Sunday at 4 AM
      timezone = "America/Los_Angeles"
      grace    = 120
    }
  }

  name_prefix            = "prod"
  default_timezone       = "UTC"
  default_grace_minutes  = 30
}
```

### With Custom Name Format

```hcl
module "cron_jobs" {
  source = "path/to/modules/cron-healthcheck"

  jobs = {
    backup = {
      cron  = "0 2 * * *"
      grace = 30
    }
  }

  name_format = "CRON-%s-healthcheck"
  # Results in: "CRON-backup-healthcheck"
}
```

### With Escalation Policy

```hcl
module "critical_cron_jobs" {
  source = "path/to/modules/cron-healthcheck"

  jobs = {
    payment_processing = {
      cron     = "*/15 * * * *"     # Every 15 minutes
      timezone = "UTC"
      grace    = 5
    }
    fraud_detection = {
      cron              = "*/30 * * * *"  # Every 30 minutes
      timezone          = "UTC"
      grace             = 10
      escalation_policy = "ep_override123"  # Job-specific override
    }
  }

  escalation_policy = "ep_critical123"  # Default for all jobs
}
```

### Paused Jobs (Maintenance)

```hcl
module "cron_jobs" {
  source = "path/to/modules/cron-healthcheck"

  jobs = {
    seasonal_report = {
      cron   = "0 0 1 * *"          # Monthly
      grace  = 60
      paused = true                  # Currently disabled
    }
  }
}
```

## Integration with Cron Jobs

After deploying the module, integrate the ping URLs into your cron scripts:

### Bash Script Integration

```bash
#!/bin/bash
# daily-backup.sh

PING_URL="https://ping.hyperping.io/abc123def456"

# Run your backup
/usr/local/bin/backup.sh

# If backup succeeded, ping Hyperping
if [ $? -eq 0 ]; then
    curl -fsS --retry 3 "$PING_URL" > /dev/null
fi
```

### Python Script Integration

```python
#!/usr/bin/env python3
import os
import requests
import sys

PING_URL = os.environ.get("PING_URL")

def main():
    # Your job logic here
    try:
        # ... do work ...

        # Ping on success
        if PING_URL:
            requests.get(PING_URL, timeout=10)
    except Exception as e:
        print(f"Job failed: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
```

### Crontab Example

```cron
# /etc/crontab

PING_URL=https://ping.hyperping.io/abc123def456

# Daily backup at 2 AM ET
0 2 * * * root /scripts/backup.sh && curl -fsS "$PING_URL" > /dev/null
```

### Using Terraform Outputs in Scripts

```bash
# Extract ping URL from Terraform
terraform output -raw backup_ping_url > /etc/cron-secrets/backup-ping-url

# Use in cron script
#!/bin/bash
PING_URL=$(cat /etc/cron-secrets/backup-ping-url)
/usr/local/bin/backup.sh && curl -fsS "$PING_URL" > /dev/null
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| `jobs` | Map of cron job configurations | `map(object)` | n/a | yes |
| `name_prefix` | Prefix for healthcheck names | `string` | `""` | no |
| `name_format` | Custom name format (use %s for key) | `string` | `""` | no |
| `default_timezone` | Default timezone for cron schedules | `string` | `"UTC"` | no |
| `default_grace_minutes` | Default grace period (minutes) | `number` | `15` | no |
| `escalation_policy` | Escalation policy UUID | `string` | `null` | no |
| `paused` | Create healthchecks in paused state | `bool` | `false` | no |

### Job Object

| Field | Description | Type | Default |
|-------|-------------|------|---------|
| `cron` | Cron schedule (5-field format) | `string` | required |
| `timezone` | IANA timezone | `string` | uses `default_timezone` |
| `grace` | Grace period in minutes | `number` | uses `default_grace_minutes` |
| `escalation_policy` | Override escalation policy | `string` | uses `escalation_policy` |
| `paused` | Override paused state | `bool` | uses `paused` |

## Outputs

| Name | Description | Sensitive |
|------|-------------|-----------|
| `healthcheck_ids` | Map of job name to healthcheck UUID | no |
| `healthcheck_ids_list` | List of all healthcheck UUIDs | no |
| `ping_urls` | Map of job name to ping URL | yes |
| `ping_urls_list` | List of all ping URLs | yes |
| `healthchecks` | Full healthcheck objects | yes |
| `job_count` | Total number of healthchecks created | no |

## Cron Schedule Format

Standard 5-field cron format:

```
┌───────────── minute (0-59)
│ ┌───────────── hour (0-23)
│ │ ┌───────────── day of month (1-31)
│ │ │ ┌───────────── month (1-12)
│ │ │ │ ┌───────────── day of week (0-6, Sunday=0)
│ │ │ │ │
* * * * *
```

### Common Examples

| Schedule | Cron Expression | Description |
|----------|----------------|-------------|
| Every hour | `0 * * * *` | Top of every hour |
| Every 15 minutes | `*/15 * * * *` | Four times per hour |
| Daily at 2 AM | `0 2 * * *` | Once per day |
| Weekly on Sunday | `0 0 * * 0` | Once per week |
| Monthly on 1st | `0 0 1 * *` | Once per month |
| Weekdays at 9 AM | `0 9 * * 1-5` | Monday-Friday |
| Every 6 hours | `0 */6 * * *` | Four times per day |

## Supported Timezones

Common IANA timezone identifiers (full list available at [IANA Time Zone Database](https://www.iana.org/time-zones)):

### Americas
```
America/New_York, America/Chicago, America/Denver, America/Los_Angeles,
America/Toronto, America/Sao_Paulo, America/Mexico_City
```

### Europe
```
Europe/London, Europe/Paris, Europe/Berlin, Europe/Amsterdam, Europe/Rome
```

### Asia
```
Asia/Tokyo, Asia/Singapore, Asia/Shanghai, Asia/Hong_Kong, Asia/Dubai,
Asia/Kolkata, Asia/Bangkok, Asia/Seoul, Asia/Manila
```

### Oceania
```
Australia/Sydney, Australia/Melbourne, Pacific/Auckland
```

### UTC
```
UTC
```

## Grace Period Guidelines

The grace period is the additional time allowed after the expected run time before an alert is triggered.

| Job Frequency | Recommended Grace Period |
|--------------|-------------------------|
| Every 5-15 min | 5-10 minutes |
| Every 30-60 min | 10-15 minutes |
| Hourly | 15-30 minutes |
| Daily | 30-60 minutes |
| Weekly | 1-2 hours |
| Monthly | 2-4 hours |

**Factors to consider:**
- Job execution time variability
- System load variations
- Network latency
- Acceptable alert delay

## Best Practices

### 1. Use Descriptive Job Names

```hcl
jobs = {
  db_backup_postgres_prod = { ... }
  log_rotation_nginx      = { ... }
  cache_warmup_redis      = { ... }
}
```

### 2. Set Appropriate Grace Periods

```hcl
jobs = {
  quick_sync = {
    cron  = "*/5 * * * *"  # Every 5 minutes
    grace = 2               # Short grace for frequent job
  }
  long_report = {
    cron  = "0 0 * * 0"    # Weekly
    grace = 120             # 2 hours for slow job
  }
}
```

### 3. Use Environment-Specific Prefixes

```hcl
# Production
module "prod_cron" {
  name_prefix = "PROD"
  jobs = { ... }
}

# Staging
module "staging_cron" {
  name_prefix = "STAGING"
  jobs = { ... }
}
```

### 4. Secure Ping URLs

```hcl
# Store in secrets manager
resource "aws_secretsmanager_secret_version" "backup_ping_url" {
  secret_id     = aws_secretsmanager_secret.backup.id
  secret_string = module.cron_jobs.ping_urls["daily_backup"]
}

# Or use Terraform output with encryption
output "ping_urls" {
  value     = module.cron_jobs.ping_urls
  sensitive = true
}
```

### 5. Test Before Production

```hcl
# Start paused, verify configuration
module "cron_jobs" {
  jobs = {
    new_job = {
      cron   = "0 3 * * *"
      paused = true
    }
  }
}

# Manually trigger job and verify ping works
# Then unpause: paused = false
```

### 6. Handle Ping Failures Gracefully

```bash
#!/bin/bash
# Don't fail the entire job if ping fails

/usr/local/bin/backup.sh

if [ $? -eq 0 ]; then
    curl -fsS --retry 3 --max-time 10 "$PING_URL" > /dev/null || true
fi
```

## Troubleshooting

### Job Not Pinging

1. **Check cron is running**: `systemctl status cron`
2. **Verify script execution**: Check cron logs (`/var/log/cron` or `journalctl`)
3. **Test ping URL manually**: `curl -v "$PING_URL"`
4. **Check network connectivity**: Ensure outbound HTTPS allowed

### False Alerts

1. **Increase grace period**: Job may take longer than expected
2. **Check timezone**: Ensure cron and healthcheck use same timezone
3. **Verify cron schedule**: Use [crontab.guru](https://crontab.guru) to validate

### Ping URL Not Found

1. **Apply Terraform**: `terraform apply` to create healthchecks
2. **Check outputs**: `terraform output ping_urls`
3. **Verify sensitive flag**: Use `-raw` for non-JSON output

## Examples

See the `tests/` directory for complete working examples:

- `tests/basic.tf` - Simple daily/hourly jobs
- `tests/advanced.tf` - Complex schedules with escalations
- `tests/integration.sh` - Shell script integration examples

## Related Resources

- [Hyperping Healthcheck Resource](https://registry.terraform.io/providers/develeap/hyperping/latest/docs/resources/healthcheck)
- [Crontab Format Reference](https://crontab.guru)
- [IANA Time Zone Database](https://www.iana.org/time-zones)
