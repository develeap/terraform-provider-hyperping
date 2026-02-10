# Complete Example - Hyperping Terraform Provider

This example demonstrates a production-ready setup of the Hyperping Terraform provider with all major features.

## What This Example Includes

### Monitors (4 types)
1. **Website Monitor** - Basic HTTP GET check with global regions
2. **API Health Check** - Authenticated endpoint with custom headers
3. **API Endpoint** - POST request with JSON body
4. **Database Health** - Pausable monitor for maintenance windows

### Data Sources
- Query all monitors in your account
- Filter and analyze monitor configurations

### Incident Management
- Create incidents for status page communication
- Add updates to track incident progress
- Link incidents to monitors and status pages

### Maintenance Windows
- Schedule planned downtime
- Link to affected monitors
- Configure notifications for subscribers

## Prerequisites

1. **Hyperping Account**: Sign up at [hyperping.io](https://hyperping.io)
2. **API Key**: Generate from [Hyperping Dashboard](https://app.hyperping.io/settings/api)
3. **Terraform**: Install version 1.0 or higher
4. **Status Page ID**: Create a status page and note its UUID

## Setup Instructions

### 1. Set Environment Variables

```bash
export HYPERPING_API_KEY="sk_your_api_key_here"
```

Optional: Set API token for authenticated health checks
```bash
export TF_VAR_api_token="your_api_bearer_token"
```

### 2. Configure Variables

Create `terraform.tfvars`:

```hcl
environment      = "production"
status_page_id   = "sp_your_status_page_uuid"
api_token        = "your_api_bearer_token"  # Optional
```

### 3. Initialize and Apply

```bash
# Initialize Terraform
terraform init

# Review planned changes
terraform plan

# Apply configuration
terraform apply
```

## Example Outputs

After applying, you'll see:

```hcl
website_monitor = {
  id       = "mon_abc123def456"
  name     = "Website - production"
  protocol = "http"
  status   = "ACTIVE"
  url      = "https://example.com"
}

api_monitors = {
  database = "mon_db123"
  endpoint = "mon_api456"
  health   = "mon_health789"
}

total_monitors = 4

active_monitors = [
  "Website - production",
  "API Health - production",
  "API Create Endpoint - production",
  "Database Health - production"
]

monitors_by_status = {
  active = 4
  paused = 0
  total  = 4
}

incident_info = {
  date = "2026-02-02T10:30:00.000Z"
  id   = "inci_xyz789"
  type = "incident"
}

maintenance_info = {
  end_date   = "2026-02-01T04:00:00.000Z"
  id         = "mw_abc123"
  name       = "database-maintenance-production"
  start_date = "2026-02-01T02:00:00.000Z"
}
```

## Key Features Demonstrated

### 1. Multiple Monitor Types

**Basic Website Monitoring:**
```hcl
resource "hyperping_monitor" "website" {
  name                 = "Website - ${var.environment}"
  url                  = "https://example.com"
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 60
  expected_status_code = "200"
  regions              = ["london", "virginia", "singapore", "tokyo"]
}
```

**Authenticated API Checks:**
```hcl
resource "hyperping_monitor" "api_health" {
  name                 = "API Health - ${var.environment}"
  url                  = "https://api.example.com/health"
  check_frequency      = 30

  request_headers = [
    { name = "Authorization", value = "Bearer ${var.api_token}" }
  ]
}
```

**POST Requests with Body:**
```hcl
resource "hyperping_monitor" "api_endpoint" {
  http_method  = "POST"
  request_body = jsonencode({ test = true })

  request_headers = [
    { name = "Content-Type", value = "application/json" }
  ]
}
```

### 2. Data Source Queries

```hcl
data "hyperping_monitors" "all" {}

output "active_monitors" {
  value = [
    for m in data.hyperping_monitors.all.monitors : m.name
    if !m.paused
  ]
}
```

### 3. Incident Workflow

```hcl
# Create incident
resource "hyperping_incident" "api_degradation" {
  title        = "API Performance Monitoring"
  text         = "Monitoring API response times..."
  status_pages = [var.status_page_id]
}

# Add update
resource "hyperping_incident_update" "investigating" {
  incident_id = hyperping_incident.api_degradation.id
  text        = "We have identified the cause..."
  type        = "identified"
}
```

### 4. Maintenance Windows

```hcl
resource "hyperping_maintenance" "database_upgrade" {
  start_date = "2026-02-01T02:00:00.000Z"
  end_date   = "2026-02-01T04:00:00.000Z"

  monitors = [
    hyperping_monitor.database.id,
    hyperping_monitor.api_health.id,
  ]

  notification_option  = "scheduled"
  notification_minutes = 60
}
```

## Customization

### Adjusting Check Frequencies

Valid check frequencies (in seconds):
- `10` - Every 10 seconds (high frequency)
- `30` - Every 30 seconds
- `60` - Every minute (default)
- `300` - Every 5 minutes
- `600` - Every 10 minutes
- `3600` - Every hour
- `86400` - Daily

### Available Regions

Choose from 19 global regions:
- Americas: `sanfrancisco`, `california`, `virginia`, `nyc`, `tokyo`, `toronto`, `saopaulo`
- Europe: `london`, `paris`, `frankfurt`, `amsterdam`
- Asia: `tokyo`, `singapore`, `seoul`, `mumbai`, `bangalore`, `bahrain`
- Africa: `capetown`
- Oceania: `sydney`

### Expected Status Codes

Support for flexible matching:
- Exact: `"200"`, `"201"`, `"204"`
- Pattern: `"2xx"`, `"3xx"`, `"4xx"`

## Production Best Practices

### 1. Use Separate State per Environment

```hcl
terraform {
  backend "s3" {
    bucket = "terraform-state"
    key    = "hyperping/production/terraform.tfstate"
    region = "us-east-1"
  }
}
```

### 2. Protect Sensitive Variables

Never commit API keys or tokens. Use:
- Environment variables
- Secret management systems (AWS Secrets Manager, HashiCorp Vault)
- Terraform Cloud/Enterprise variables

### 3. Tag Resources with Environment

```hcl
locals {
  name_prefix = "[${upper(var.environment)}]"
}

resource "hyperping_monitor" "api" {
  name = "${local.name_prefix} API Health"
}
```

### 4. Use Workspaces for Multi-Environment

```bash
# Create workspaces
terraform workspace new production
terraform workspace new staging

# Switch and apply
terraform workspace select production
terraform apply -var-file="production.tfvars"
```

### 5. Monitor the Monitors

Set up alerting:
- Configure escalation policies in Hyperping
- Integrate with PagerDuty, Slack, or webhooks
- Use maintenance windows during deployments

## Troubleshooting

### Authentication Errors

```bash
# Verify API key
curl -H "Authorization: Bearer $HYPERPING_API_KEY" \
  https://api.hyperping.io/v1/monitors

# Check permissions
echo $HYPERPING_API_KEY | grep "^sk_"
```

### Rate Limiting

If you hit rate limits:

```bash
# Reduce Terraform parallelism
terraform apply -parallelism=1
```

### Invalid Status Page ID

Verify status page exists:

```bash
curl -H "Authorization: Bearer $HYPERPING_API_KEY" \
  https://api.hyperping.io/v1/statuspages
```

### Maintenance Date Validation

Ensure dates are:
- In ISO 8601 format with `.000Z` suffix
- In the future (for scheduling)
- End date after start date

## Cleanup

To destroy all resources:

```bash
# Preview what will be destroyed
terraform plan -destroy

# Destroy all resources
terraform destroy
```

**Warning:** This will permanently delete all monitors, incidents, and maintenance windows created by this configuration.

## Next Steps

1. **Explore More Examples**: See other examples in the `examples/` directory
2. **Add Status Pages**: Configure public status pages for your services
3. **Set Up Integrations**: Connect to Slack, PagerDuty, or webhooks
4. **Automate in CI/CD**: Integrate Terraform into your deployment pipeline

## Resources

- [Provider Documentation](https://registry.terraform.io/providers/develeap/hyperping/latest/docs)
- [Hyperping API Docs](https://hyperping.notion.site/Hyperping-API-documentation-a0dc48fb818e4542a8f7fb4163ede2c3)
- [Terraform Best Practices](https://developer.hashicorp.com/terraform/language/best-practices)

## Support

- Issues: [GitHub Issues](https://github.com/develeap/terraform-provider-hyperping/issues)
- Questions: [GitHub Discussions](https://github.com/develeap/terraform-provider-hyperping/discussions)
- Documentation: [Provider Docs](https://registry.terraform.io/providers/develeap/hyperping/latest/docs)
