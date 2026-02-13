---
page_title: "Quick Start - Monitor Your First Service in 5 Minutes"
subcategory: "Getting Started"
description: |-
  Get started with the Hyperping Terraform provider in 5 minutes. This guide walks you through creating your first uptime monitor.
---

# Quick Start - Monitor Your First Service in 5 Minutes

Get started with the Hyperping Terraform provider in 5 minutes. This guide walks you through creating your first uptime monitor.

## Prerequisites (30 seconds)

Before you begin, ensure you have:

- [x] Terraform 1.8 or later installed ([Install Terraform](https://developer.hashicorp.com/terraform/install))
- [x] A Hyperping account ([Sign up free](https://hyperping.io))
- [x] A Hyperping API key ([Create in Settings → API Keys](https://app.hyperping.io/settings/api-keys))

Check your Terraform version:

```bash
terraform version
# Should show: Terraform v1.8.0 or higher
```

## Step 1: Configure the Provider (2 minutes)

Create a new directory for your configuration:

```bash
mkdir hyperping-quickstart
cd hyperping-quickstart
```

Create `main.tf`:

```hcl
terraform {
  required_version = ">= 1.8"

  required_providers {
    hyperping = {
      source  = "develeap/hyperping"
      version = "~> 1.0"
    }
  }
}

provider "hyperping" {
  # API key will be read from HYPERPING_API_KEY environment variable
  # Alternatively, you can set it here: api_key = "sk_your_api_key"
}
```

Set your API key as an environment variable:

```bash
export HYPERPING_API_KEY="sk_your_api_key_here"
```

> **💡 Tip:** Never commit your API key to version control. Always use environment variables or a secrets manager.

## Step 2: Create Your First Monitor (1 minute)

Add this to your `main.tf`:

```hcl
resource "hyperping_monitor" "my_first_monitor" {
  name     = "My Website"
  url      = "https://example.com"
  protocol = "http"

  # Check every 60 seconds
  check_frequency = 60

  # Monitor from multiple regions for redundancy
  regions = [
    "virginia", # US East
    "london",   # Europe
    "singapore" # Asia
  ]

  # Expect HTTP 200 OK
  expected_status_code = "200"
}

output "monitor_id" {
  value       = hyperping_monitor.my_first_monitor.id
  description = "The UUID of your monitor"
}

output "monitor_url" {
  value       = "https://app.hyperping.io/monitors/${hyperping_monitor.my_first_monitor.id}"
  description = "View your monitor in the Hyperping dashboard"
}
```

## Step 3: Deploy Your Monitor (1 minute)

Initialize Terraform:

```bash
terraform init
```

Preview the changes:

```bash
terraform plan
```

Apply the configuration:

```bash
terraform apply
```

Type `yes` when prompted.

**Expected output:**

```
Apply complete! Resources: 1 added, 0 changed, 0 destroyed.

Outputs:

monitor_id = "mon_abc123def456"
monitor_url = "https://app.hyperping.io/monitors/mon_abc123def456"
```

## Step 4: Verify in Dashboard (30 seconds)

1. Open the monitor URL from the output
2. You should see your monitor actively checking https://example.com
3. Check results will appear within 60 seconds

**Congratulations!** You've created your first uptime monitor with Terraform.

---

## Next Steps

Now that you have your first monitor, explore more features:

### Add More Monitors

Monitor multiple services by adding more resources:

```hcl
resource "hyperping_monitor" "api" {
  name                 = "API Health Endpoint"
  url                  = "https://api.example.com/health"
  protocol             = "http"
  check_frequency      = 60
  expected_status_code = "200"
  regions              = ["london", "virginia"]

  request_headers = [
    {
      name  = "Accept"
      value = "application/json"
    }
  ]
}

resource "hyperping_monitor" "database" {
  name            = "Database Port"
  url             = "db.example.com"
  protocol        = "port"
  port            = 5432
  check_frequency = 300
  regions         = ["virginia"]
}
```

### Create a Status Page

Share uptime status with your users. Learn more in the [Status Page Resource Documentation](https://registry.terraform.io/providers/develeap/hyperping/latest/docs/resources/statuspage).

```hcl
resource "hyperping_statuspage" "public" {
  name      = "Service Status"
  subdomain = "status" # status.hyperping.app
  theme     = "dark"

  sections = [{
    name     = { en = "Core Services" }
    is_split = true
    services = [{
      monitor_uuid        = hyperping_monitor.my_first_monitor.id
      show_uptime         = true
      show_response_times = true
    }]
  }]
}
```

### Set Up Maintenance Windows

Schedule maintenance to prevent false alerts:

```hcl
resource "hyperping_maintenance" "upgrade" {
  name       = "Database Upgrade"
  title      = "Scheduled Database Maintenance"
  text       = "We will be upgrading our database infrastructure."
  start_date = "2026-02-20T02:00:00Z"
  end_date   = "2026-02-20T04:00:00Z"
  monitors   = [hyperping_monitor.my_first_monitor.id]
}
```

### Monitor Cron Jobs

Use healthchecks to monitor scheduled tasks (dead man's switch):

```hcl
resource "hyperping_healthcheck" "daily_backup" {
  name               = "Daily Backup Job"
  cron               = "0 2 * * *" # 2 AM every day
  timezone           = "America/New_York"
  grace_period_value = 30
  grace_period_type  = "minutes"
}

output "backup_ping_url" {
  value       = hyperping_healthcheck.daily_backup.ping_url
  description = "Add this URL to your backup script: curl $PING_URL"
  sensitive   = true
}
```

### Advanced Use Cases

Explore comprehensive examples:

- [Complete Example](../../examples/complete/main.tf) - All features in one configuration
- [Multi-Tenant Setup](../../examples/multi-tenant/) - Managing monitors for multiple clients
- [Advanced Patterns](../../examples/advanced-patterns/main.tf) - Complex monitoring scenarios

---

## Common Issues

### "Authentication failed" or "401 Unauthorized"

**Problem:** Invalid or missing API key

**Solution:**

1. Verify your API key in [Hyperping Settings](https://app.hyperping.io/settings/api-keys)
2. Ensure the environment variable is set:
   ```bash
   echo $HYPERPING_API_KEY
   ```
3. Check for typos (API keys start with `sk_`)

### "Invalid URL format"

**Problem:** URL missing protocol or malformed

**Solution:** Always include the protocol:

```hcl
# ❌ Wrong
url = "example.com"

# ✅ Correct
url = "https://example.com"
```

### "Region not available" or "Invalid region"

**Problem:** Specified region doesn't exist

**Solution:** Use valid regions from the list below:

```hcl
regions = [
  "virginia",   # US East
  "oregon",     # US West
  "london",     # Europe West
  "frankfurt",  # Europe Central
  "singapore",  # Asia Pacific
  "sydney",     # Australia
  "tokyo",      # Asia Pacific
  "saopaulo"    # South America
]
```

### "Error: Provider configuration not present"

**Problem:** Terraform cannot find the provider

**Solution:** Run `terraform init` to download the provider:

```bash
terraform init -upgrade
```

### Monitor not appearing in dashboard

**Problem:** Dashboard may take a moment to update

**Solution:**

1. Wait 30-60 seconds for the first check
2. Refresh your browser
3. Verify the monitor ID in the output matches the dashboard URL

---

## What You Learned

✅ How to configure the Hyperping provider
✅ How to create an uptime monitor
✅ How to monitor from multiple regions
✅ How to view monitors in the dashboard

**Time to first monitor:** ~5 minutes

---

## Additional Resources

### Documentation

- [Provider Configuration](https://registry.terraform.io/providers/develeap/hyperping/latest/docs)
- [Monitor Resource Reference](https://registry.terraform.io/providers/develeap/hyperping/latest/docs/resources/monitor)
- [All Resources and Data Sources](https://registry.terraform.io/providers/develeap/hyperping/latest/docs)

### Guides

- [Importing Existing Resources](./importing-resources.md)
- [Error Handling Patterns](./error-handling.md)
- [Filtering Data Sources](./filtering-data-sources.md)
- [Rate Limit Management](./rate-limits.md)

### Examples

Browse the [examples directory](../../examples/) for real-world patterns:

- [Complete setup](../../examples/complete/main.tf) - All features
- [Reusable modules](../../examples/modules/) - Modular configurations
- [Validation tests](../../examples/validation-test/main.tf) - Testing patterns

### Community

- [GitHub Issues](https://github.com/develeap/terraform-provider-hyperping/issues) - Report bugs or request features
- [Hyperping Support](https://hyperping.io/support) - Official Hyperping help

---

## Ready for Production?

Once you're comfortable with the basics, check out these production-ready patterns:

### Use Remote State

Store your Terraform state in a remote backend:

```hcl
terraform {
  backend "s3" {
    bucket = "my-terraform-state"
    key    = "hyperping/terraform.tfstate"
    region = "us-east-1"
  }
}
```

### Modularize Your Configuration

Create reusable modules for common patterns:

```hcl
module "api_monitoring" {
  source = "./modules/api-health"

  api_name    = "Payment API"
  api_url     = "https://api.example.com/health"
  regions     = ["london", "virginia", "singapore"]
  alert_email = "ops@example.com"
}
```

See [examples/modules/](../../examples/modules/) for pre-built modules.

### Implement CI/CD

Automate deployments with GitHub Actions:

```yaml
name: Terraform
on:
  push:
    branches: [main]

jobs:
  terraform:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3
      - run: terraform init
      - run: terraform plan
      - run: terraform apply -auto-approve
    env:
      HYPERPING_API_KEY: ${{ secrets.HYPERPING_API_KEY }}
```

---

**Need help?** Check out the [Troubleshooting Guide](./error-handling.md) or [open an issue](https://github.com/develeap/terraform-provider-hyperping/issues).
