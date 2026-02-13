# Incident Management Module

Production-ready Terraform module for comprehensive incident management with Hyperping. Provides pre-configured incident templates, maintenance window scheduling, manual outage tracking, and status page integration.

## Features

- Pre-configured incident templates (investigating, identified, monitoring, resolved)
- Scheduled maintenance window management
- Manual outage creation and tracking
- Status page integration
- Severity-based incident type mapping
- Flexible notification configuration
- Multiple incident/maintenance workflow support

## Usage

### Basic Incident Management

```hcl
module "incidents" {
  source = "path/to/modules/incident-management"

  statuspage_id = hyperping_statuspage.main.id

  incident_templates = {
    api_degradation = {
      title    = "API Performance Degraded"
      text     = "We are investigating reports of slow API response times."
      severity = "major"
    }
    database_outage = {
      title    = "Database Connection Issues"
      text     = "Database cluster experiencing connectivity problems."
      severity = "critical"
    }
  }
}
```

### Complete Example with All Features

```hcl
module "incident_mgmt" {
  source = "path/to/modules/incident-management"

  statuspage_id = hyperping_statuspage.main.id

  # Incident templates for common issues
  incident_templates = {
    api_degradation = {
      title               = "API Performance Degraded"
      text                = "Investigating slow response times across API endpoints."
      severity            = "major"
      affected_components = [hyperping_component.api.id]
    }

    payment_outage = {
      title               = "Payment Processing Unavailable"
      text                = "Payment gateway is currently offline. We are working to restore service."
      severity            = "critical"
      type                = "outage"
      affected_components = [
        hyperping_component.payment.id,
        hyperping_component.checkout.id
      ]
    }

    minor_issue = {
      title    = "Minor Connectivity Issue"
      text     = "Intermittent connectivity to secondary services."
      severity = "minor"
    }
  }

  # Scheduled maintenance windows
  maintenance_windows = {
    database_upgrade = {
      title                = "Database Maintenance"
      text                 = "Routine database optimization and upgrades."
      start_date           = "2026-02-20T02:00:00.000Z"
      end_date             = "2026-02-20T04:00:00.000Z"
      monitors             = [hyperping_monitor.db.id]
      notification_option  = "scheduled"
      notification_minutes = 120
    }

    emergency_maintenance = {
      title               = "Emergency Network Maintenance"
      text                = "Addressing critical network infrastructure issues."
      start_date          = "2026-02-15T00:00:00.000Z"
      end_date            = "2026-02-15T06:00:00.000Z"
      monitors            = [hyperping_monitor.network.id]
      notification_option = "immediate"
    }
  }

  # Manual outages for specific monitors
  outage_definitions = {
    planned_api_downtime = {
      monitor_uuid = hyperping_monitor.api.id
      start_date   = "2026-02-15T02:00:00Z"
      end_date     = "2026-02-15T04:00:00Z"
      status_code  = 503
      description  = "Planned API downtime for infrastructure migration"
    }
  }
}

# Access created resources
output "incident_ids" {
  value = module.incident_mgmt.incident_ids
}

output "maintenance_schedule" {
  value = module.incident_mgmt.maintenance_windows
}
```

### Incident Workflow Integration

```hcl
# Monitor infrastructure
module "api_monitors" {
  source = "path/to/modules/api-health"

  endpoints = {
    users   = { url = "https://api.example.com/users/health" }
    orders  = { url = "https://api.example.com/orders/health" }
    payment = { url = "https://api.example.com/payment/health" }
  }
}

# Status page
resource "hyperping_statuspage" "main" {
  name             = "Example Corp Status"
  hosted_subdomain = "status"

  settings = {
    name             = "Example Corp Status"
    languages        = ["en"]
    default_language = "en"
    theme            = "light"

    subscribe = {
      enabled = true
      email   = true
    }
  }

  sections = [{
    name = { en = "Services" }
    services = [
      for k, v in module.api_monitors.monitors : {
        uuid                = v.id
        name                = { en = k }
        show_uptime         = true
        show_response_times = true
      }
    ]
  }]
}

# Incident management
module "incidents" {
  source = "path/to/modules/incident-management"

  statuspage_id = hyperping_statuspage.main.id

  incident_templates = {
    api_degradation = {
      title    = "API Performance Issues"
      text     = "Investigating degraded performance across API services."
      severity = "major"
    }
  }

  maintenance_windows = {
    weekly_maintenance = {
      title      = "Weekly Maintenance Window"
      text       = "Routine system maintenance and updates."
      start_date = "2026-02-16T03:00:00.000Z"
      end_date   = "2026-02-16T05:00:00.000Z"
      monitors   = module.api_monitors.monitor_ids_list
    }
  }
}
```

### Conditional Resource Creation

```hcl
module "incidents" {
  source = "path/to/modules/incident-management"

  statuspage_id = hyperping_statuspage.main.id

  # Only create incidents in production
  create_incidents = var.environment == "production"

  # Only create maintenance windows when scheduled
  create_maintenance = var.has_scheduled_maintenance

  # Disable outage creation
  create_outages = false

  incident_templates = {
    api_issue = {
      title    = "API Service Issue"
      text     = "Investigating API connectivity problems."
      severity = "major"
    }
  }
}
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| `statuspage_id` | UUID of status page to associate with | `string` | `null` | no |
| `incident_templates` | Map of incident configurations | `map(object)` | `{}` | no |
| `maintenance_windows` | Map of maintenance window configurations | `map(object)` | `{}` | no |
| `outage_definitions` | Map of manual outage configurations | `map(object)` | `{}` | no |
| `create_incidents` | Enable incident resource creation | `bool` | `true` | no |
| `create_maintenance` | Enable maintenance resource creation | `bool` | `true` | no |
| `create_outages` | Enable outage resource creation | `bool` | `true` | no |
| `default_incident_type` | Default incident type | `string` | `"incident"` | no |

### Incident Template Object

| Field | Description | Type | Default |
|-------|-------------|------|---------|
| `title` | Incident title (1-255 chars) | `string` | required |
| `text` | Incident description/message | `string` | required |
| `severity` | Incident severity level | `string` | `"major"` |
| `type` | Incident type | `string` | auto (based on severity) |
| `affected_components` | List of affected component UUIDs | `list(string)` | `[]` |
| `status_pages` | Override status pages | `list(string)` | uses module `statuspage_id` |

**Valid Severity Values:**
- `minor` - Minor issues, limited impact
- `major` - Significant degradation (maps to "incident" type)
- `critical` - Severe outage (maps to "outage" type)

**Valid Type Values:**
- `incident` - Service degradation
- `outage` - Complete service unavailability

### Maintenance Window Object

| Field | Description | Type | Default |
|-------|-------------|------|---------|
| `title` | Maintenance window title | `string` | required |
| `text` | Maintenance description | `string` | required |
| `start_date` | Start datetime (ISO 8601) | `string` | required |
| `end_date` | End datetime (ISO 8601) | `string` | required |
| `monitors` | List of monitor UUIDs | `list(string)` | `[]` |
| `status_pages` | Override status pages | `list(string)` | uses module `statuspage_id` |
| `notification_option` | Notification timing | `string` | `"scheduled"` |
| `notification_minutes` | Minutes before start to notify | `number` | `60` |

**Valid Notification Options:**
- `immediate` - Notify immediately when created
- `scheduled` - Notify X minutes before start
- `none` - No notification

### Outage Definition Object

| Field | Description | Type | Default |
|-------|-------------|------|---------|
| `monitor_uuid` | Monitor UUID to create outage for | `string` | required |
| `start_date` | Outage start datetime (ISO 8601) | `string` | required |
| `end_date` | Outage end datetime (ISO 8601) | `string` | `null` (ongoing) |
| `status_code` | HTTP status code for outage | `number` | required |
| `description` | Outage description | `string` | required |

**Note:** All outage fields are ForceNew - any change triggers destroy and recreate.

## Outputs

| Name | Description |
|------|-------------|
| `incident_ids` | Map of template name to incident ID |
| `incident_ids_list` | List of all incident IDs |
| `incidents` | Full incident objects with metadata |
| `maintenance_ids` | Map of window name to maintenance ID |
| `maintenance_ids_list` | List of all maintenance IDs |
| `maintenance_windows` | Full maintenance window objects |
| `outage_ids` | Map of outage name to outage ID |
| `outage_ids_list` | List of all outage IDs |
| `outages` | Full outage objects |
| `incident_count` | Total incidents created |
| `maintenance_count` | Total maintenance windows created |
| `outage_count` | Total outages created |
| `summary` | Summary of all created resources |

## Common Patterns

### Multi-Severity Incident Templates

```hcl
module "incidents" {
  source = "path/to/modules/incident-management"

  statuspage_id = hyperping_statuspage.main.id

  incident_templates = {
    # Minor - investigating
    cache_issue = {
      title    = "Cache Performance Degraded"
      text     = "Investigating cache layer performance issues."
      severity = "minor"
    }

    # Major - service degradation
    api_slow = {
      title    = "API Response Times Elevated"
      text     = "API endpoints experiencing increased latency."
      severity = "major"
    }

    # Critical - complete outage
    database_down = {
      title    = "Database Cluster Offline"
      text     = "Primary database cluster is unavailable. Working to restore."
      severity = "critical"
    }
  }
}
```

### Recurring Maintenance Windows

```hcl
# Use Terraform functions to generate recurring schedules
locals {
  # Generate next 4 weekly maintenance windows
  weekly_maintenance = {
    for i in range(4) : "week_${i + 1}" => {
      title      = "Weekly Maintenance - Week ${i + 1}"
      text       = "Routine weekly system maintenance and updates."
      start_date = formatdate("YYYY-MM-DD'T'03:00:00.000'Z'",
        timeadd(timestamp(), "${7 * i}h"))
      end_date   = formatdate("YYYY-MM-DD'T'05:00:00.000'Z'",
        timeadd(timestamp(), "${7 * i + 2}h"))
      monitors   = var.all_monitor_ids
    }
  }
}

module "incidents" {
  source = "path/to/modules/incident-management"

  statuspage_id      = hyperping_statuspage.main.id
  maintenance_windows = local.weekly_maintenance
}
```

### Integration with Monitor Modules

```hcl
# Create monitors
module "production_monitors" {
  source = "path/to/modules/api-health"

  name_prefix = "PROD"
  endpoints = {
    api      = { url = "https://api.example.com/health" }
    web      = { url = "https://www.example.com" }
    database = { url = "https://db.example.com/health" }
  }
}

# Link incidents to monitors via outages
module "incident_mgmt" {
  source = "path/to/modules/incident-management"

  statuspage_id = hyperping_statuspage.main.id

  # Create incidents for common issues
  incident_templates = {
    api_degradation = {
      title    = "API Performance Issues"
      text     = "Investigating degraded API performance."
      severity = "major"
    }
  }

  # Schedule maintenance for database monitor
  maintenance_windows = {
    db_maintenance = {
      title      = "Database Upgrade"
      text       = "Upgrading database to latest version."
      start_date = "2026-03-01T02:00:00.000Z"
      end_date   = "2026-03-01T04:00:00.000Z"
      monitors   = [module.production_monitors.monitor_ids["database"]]
    }
  }
}
```

## Best Practices

### 1. Incident Severity Guidelines

| Severity | Use When | Notification | Example |
|----------|----------|--------------|---------|
| `minor` | Minimal impact, workarounds available | Standard | Cache delay, non-critical features |
| `major` | Significant degradation, user impact | Elevated | Slow API, partial outages |
| `critical` | Complete service unavailability | Immediate | Database down, total outage |

### 2. Maintenance Window Scheduling

- Schedule maintenance during low-traffic periods
- Use `scheduled` notification with adequate notice (120+ minutes)
- Group related monitors in same maintenance window
- Test maintenance procedures in staging first

### 3. Status Page Integration

Always link incidents to status pages for transparency:

```hcl
module "incidents" {
  source = "path/to/modules/incident-management"

  # Link to primary status page
  statuspage_id = hyperping_statuspage.public.id

  incident_templates = {
    # Incidents automatically appear on status page
    api_issue = {
      title    = "API Service Degraded"
      text     = "Investigating performance issues."
      severity = "major"
    }
  }
}
```

### 4. Manual Outages for Testing

Create manual outages for testing incident workflows:

```hcl
module "test_incidents" {
  source = "path/to/modules/incident-management"

  # Only create in non-production
  create_incidents   = false
  create_maintenance = false
  create_outages     = var.environment != "production"

  outage_definitions = {
    test_outage = {
      monitor_uuid = hyperping_monitor.staging_api.id
      start_date   = "2026-02-14T10:00:00Z"
      end_date     = "2026-02-14T10:05:00Z"
      status_code  = 503
      description  = "Test incident workflow"
    }
  }
}
```

## Datetime Format

All datetime fields must use ISO 8601 format with UTC timezone:

```
YYYY-MM-DDTHH:MM:SS.sssZ

Examples:
  2026-02-20T02:00:00.000Z
  2026-03-15T14:30:00.000Z
```

Use Terraform's `formatdate()` function for dynamic dates:

```hcl
start_date = formatdate("YYYY-MM-DD'T'HH:mm:ss.SSS'Z'", timestamp())
```

## Limitations

1. **Outage Immutability**: Outage resources cannot be updated - any change triggers destroy/recreate
2. **Incident Updates**: This module creates initial incidents only - use Hyperping API/UI for status updates
3. **Manual Cleanup**: Resolved incidents must be manually removed from configuration
4. **Timezone**: All times must be in UTC (Z suffix required)

## Testing

Run module tests with:

```bash
cd examples/modules/incident-management
terraform test
```

See `tests/` directory for test cases covering:
- Basic incident creation
- Maintenance window scheduling
- Manual outage management
- Status page integration
- Validation rules
- Conditional resource creation

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.0 |
| hyperping | >= 1.0 |

## License

Same as parent Hyperping Terraform Provider.
