# Example Usage - Incident Management Module
#
# This file demonstrates various usage patterns for the incident-management module.
# Not meant to be executed directly - copy relevant sections to your configuration.

# ============================================================================
# Example 1: Basic Incident Templates
# ============================================================================

module "basic_incidents" {
  source = "./modules/incident-management"

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

# ============================================================================
# Example 2: Maintenance Windows with Monitors
# ============================================================================

module "maintenance_scheduling" {
  source = "./modules/incident-management"

  statuspage_id = hyperping_statuspage.main.id

  maintenance_windows = {
    weekly_db_maintenance = {
      title                = "Weekly Database Maintenance"
      text                 = "Routine database optimization and backup verification."
      start_date           = "2026-02-23T02:00:00.000Z"
      end_date             = "2026-02-23T04:00:00.000Z"
      monitors             = [hyperping_monitor.database.id]
      notification_option  = "scheduled"
      notification_minutes = 120
    }

    emergency_patch = {
      title               = "Emergency Security Patch"
      text                = "Critical security update being applied immediately."
      start_date          = "2026-02-15T14:00:00.000Z"
      end_date            = "2026-02-15T14:30:00.000Z"
      monitors            = [hyperping_monitor.api.id, hyperping_monitor.web.id]
      notification_option = "immediate"
    }
  }
}

# ============================================================================
# Example 3: Complete Incident Management with All Features
# ============================================================================

module "complete_incident_mgmt" {
  source = "./modules/incident-management"

  statuspage_id = hyperping_statuspage.main.id

  # Incident templates for common issues
  incident_templates = {
    api_slow = {
      title               = "API Response Times Elevated"
      text                = "API endpoints experiencing increased latency. Engineering team investigating."
      severity            = "major"
      affected_components = [hyperping_component.api.id]
    }

    payment_down = {
      title    = "Payment Processing Unavailable"
      text     = "Payment gateway is currently offline. Payments cannot be processed."
      severity = "critical"
      type     = "outage"
      affected_components = [
        hyperping_component.payment.id,
        hyperping_component.checkout.id
      ]
    }

    cache_issue = {
      title    = "Cache Performance Degraded"
      text     = "Cache layer experiencing elevated miss rates."
      severity = "minor"
    }
  }

  # Scheduled maintenance
  maintenance_windows = {
    infrastructure_upgrade = {
      title      = "Infrastructure Upgrade"
      text       = "Upgrading cloud infrastructure for improved performance and reliability."
      start_date = "2026-03-01T02:00:00.000Z"
      end_date   = "2026-03-01T05:00:00.000Z"
      monitors = [
        hyperping_monitor.api.id,
        hyperping_monitor.web.id,
        hyperping_monitor.database.id
      ]
      notification_option  = "scheduled"
      notification_minutes = 180
    }
  }

  # Manual outages for testing
  outage_definitions = {
    staging_test = {
      monitor_uuid = hyperping_monitor.staging_api.id
      start_date   = "2026-02-14T10:00:00Z"
      end_date     = "2026-02-14T10:15:00Z"
      status_code  = 503
      description  = "Testing incident workflow in staging environment"
    }
  }
}

# ============================================================================
# Example 4: Environment-Specific Configuration
# ============================================================================

module "production_incidents" {
  source = "./modules/incident-management"

  statuspage_id = hyperping_statuspage.production.id

  # Only create incidents in production
  create_incidents   = var.environment == "production"
  create_maintenance = var.environment == "production"
  create_outages     = false

  incident_templates = {
    production_outage = {
      title    = "Production Service Outage"
      text     = "Critical production services are experiencing an outage."
      severity = "critical"
    }
  }

  maintenance_windows = {
    production_maintenance = {
      title      = "Production Maintenance Window"
      text       = "Scheduled maintenance for production infrastructure."
      start_date = "2026-02-20T04:00:00.000Z"
      end_date   = "2026-02-20T06:00:00.000Z"
      monitors   = var.production_monitor_ids
    }
  }
}

module "staging_testing" {
  source = "./modules/incident-management"

  # No status page for staging
  statuspage_id = null

  # Only create test outages in staging
  create_incidents   = false
  create_maintenance = false
  create_outages     = var.environment == "staging"

  outage_definitions = {
    incident_test = {
      monitor_uuid = hyperping_monitor.staging_api.id
      start_date   = "2026-02-14T15:00:00Z"
      end_date     = "2026-02-14T15:10:00Z"
      status_code  = 503
      description  = "Testing incident response procedures"
    }
  }
}

# ============================================================================
# Example 5: Integration with Other Modules
# ============================================================================

# Create monitors
module "production_monitors" {
  source = "./modules/api-health"

  name_prefix = "PROD"
  endpoints = {
    api      = { url = "https://api.example.com/health" }
    web      = { url = "https://www.example.com" }
    database = { url = "https://db.example.com/health" }
  }
}

# Create status page
module "status_page" {
  source = "./modules/statuspage-complete"

  name             = "Example Corp Status"
  hosted_subdomain = "status"

  services = {
    api      = { url = "https://api.example.com/health" }
    web      = { url = "https://www.example.com" }
    database = { url = "https://db.example.com/health" }
  }
}

# Link incident management
module "incident_mgmt" {
  source = "./modules/incident-management"

  statuspage_id = module.status_page.statuspage_id

  incident_templates = {
    api_degradation = {
      title               = "API Performance Degraded"
      text                = "Investigating slow API response times."
      severity            = "major"
      affected_components = [hyperping_component.api.id]
    }
  }

  maintenance_windows = {
    routine_maintenance = {
      title      = "Weekly System Maintenance"
      text       = "Routine updates and optimization."
      start_date = "2026-02-16T03:00:00.000Z"
      end_date   = "2026-02-16T05:00:00.000Z"
      monitors   = module.production_monitors.monitor_ids_list
    }
  }
}

# ============================================================================
# Example 6: Dynamic Maintenance Window Generation
# ============================================================================

locals {
  # Generate 4 weekly maintenance windows
  weekly_maintenance_windows = {
    for i in range(4) : "week_${i + 1}" => {
      title = "Weekly Maintenance - Week ${i + 1}"
      text  = "Routine weekly system maintenance and updates."
      start_date = formatdate("YYYY-MM-DD'T'03:00:00.000'Z'",
      timeadd(timestamp(), "${7 * i * 24}h"))
      end_date = formatdate("YYYY-MM-DD'T'05:00:00.000'Z'",
      timeadd(timestamp(), "${(7 * i * 24) + 2}h"))
      monitors             = var.all_monitor_ids
      notification_option  = "scheduled"
      notification_minutes = 120
    }
  }
}

module "recurring_maintenance" {
  source = "./modules/incident-management"

  statuspage_id       = hyperping_statuspage.main.id
  maintenance_windows = local.weekly_maintenance_windows
}

# ============================================================================
# Outputs (Example - use in root module, not in this file)
# ============================================================================

# output "incident_ids" {
#   description = "Created incident IDs"
#   value       = module.complete_incident_mgmt.incident_ids
# }

# output "maintenance_schedule" {
#   description = "Scheduled maintenance windows"
#   value       = module.complete_incident_mgmt.maintenance_windows
# }

# output "management_summary" {
#   description = "Summary of incident management resources"
#   value       = module.complete_incident_mgmt.summary
# }
