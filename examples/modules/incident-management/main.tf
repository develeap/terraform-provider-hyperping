# Incident Management Module - Main
#
# Provides pre-configured incident templates and workflow automation
# for incident management integrated with status pages and monitors.
#
# Usage:
#   module "incident_mgmt" {
#     source = "path/to/modules/incident-management"
#
#     statuspage_id = hyperping_statuspage.main.id
#
#     incident_templates = {
#       api_degradation = {
#         title    = "API Performance Degraded"
#         severity = "major"
#         monitors = [hyperping_monitor.api.id]
#       }
#     }
#   }

locals {
  # Flatten incident templates for resource creation
  incidents = {
    for k, v in var.incident_templates : k => merge(
      {
        type                = "incident"
        affected_components = []
        status_pages        = var.statuspage_id != null ? [var.statuspage_id] : []
      },
      v
    )
  }

  # Flatten maintenance windows for resource creation
  maintenance_windows = {
    for k, v in var.maintenance_windows : k => merge(
      {
        notification_option  = "scheduled"
        notification_minutes = 60
        status_pages         = var.statuspage_id != null ? [var.statuspage_id] : []
        monitors             = []
      },
      v
    )
  }

  # Map severity to incident type
  severity_to_type = {
    minor    = "incident"
    major    = "incident"
    critical = "outage"
  }

  # Create outages for monitors when enabled
  outages = var.create_outages ? {
    for k, v in var.outage_definitions : k => v
  } : {}
}

# Create incident templates
resource "hyperping_incident" "template" {
  for_each = var.create_incidents ? local.incidents : {}

  title               = each.value.title
  text                = each.value.text
  type                = lookup(each.value, "type", lookup(local.severity_to_type, each.value.severity, "incident"))
  status_pages        = each.value.status_pages
  affected_components = each.value.affected_components

  lifecycle {
    create_before_destroy = true
    # Prevent accidental deletion of active incidents
    prevent_destroy = false
  }
}

# Create maintenance windows
resource "hyperping_maintenance" "window" {
  for_each = var.create_maintenance ? local.maintenance_windows : {}

  name                 = each.key
  title                = each.value.title
  text                 = each.value.text
  start_date           = each.value.start_date
  end_date             = each.value.end_date
  monitors             = each.value.monitors
  status_pages         = each.value.status_pages
  notification_option  = each.value.notification_option
  notification_minutes = each.value.notification_option == "scheduled" ? each.value.notification_minutes : null

  lifecycle {
    create_before_destroy = true
  }
}

# Create manual outages for monitors
resource "hyperping_outage" "manual" {
  for_each = local.outages

  monitor_uuid = each.value.monitor_uuid
  start_date   = each.value.start_date
  end_date     = each.value.end_date
  status_code  = each.value.status_code
  description  = each.value.description

  lifecycle {
    # All fields are ForceNew - any change triggers recreate
    create_before_destroy = false
  }
}
