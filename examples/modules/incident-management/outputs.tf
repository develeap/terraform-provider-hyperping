# Incident Management Module - Outputs

output "incident_ids" {
  description = "Map of incident template name to incident ID"
  value = {
    for k, v in hyperping_incident.template : k => v.id
  }
}

output "incident_ids_list" {
  description = "List of all incident IDs"
  value       = [for v in hyperping_incident.template : v.id]
}

output "incidents" {
  description = "Full incident objects for advanced usage"
  value = {
    for k, v in hyperping_incident.template : k => {
      id                  = v.id
      title               = v.title
      text                = v.text
      type                = v.type
      date                = v.date
      affected_components = v.affected_components
    }
  }
}

output "maintenance_ids" {
  description = "Map of maintenance window name to maintenance ID"
  value = {
    for k, v in hyperping_maintenance.window : k => v.id
  }
}

output "maintenance_ids_list" {
  description = "List of all maintenance window IDs"
  value       = [for v in hyperping_maintenance.window : v.id]
}

output "maintenance_windows" {
  description = "Full maintenance window objects for advanced usage"
  value = {
    for k, v in hyperping_maintenance.window : k => {
      id         = v.id
      name       = v.name
      title      = v.title
      text       = v.text
      start_date = v.start_date
      end_date   = v.end_date
      monitors   = v.monitors
    }
  }
}

output "outage_ids" {
  description = "Map of outage name to outage ID"
  value = {
    for k, v in hyperping_outage.manual : k => v.id
  }
}

output "outage_ids_list" {
  description = "List of all outage IDs"
  value       = [for v in hyperping_outage.manual : v.id]
}

output "outages" {
  description = "Full outage objects for advanced usage"
  value = {
    for k, v in hyperping_outage.manual : k => {
      id           = v.id
      monitor_uuid = v.monitor_uuid
      start_date   = v.start_date
      end_date     = v.end_date
      status_code  = v.status_code
      description  = v.description
    }
  }
}

output "incident_count" {
  description = "Total number of incidents created"
  value       = length(hyperping_incident.template)
}

output "maintenance_count" {
  description = "Total number of maintenance windows created"
  value       = length(hyperping_maintenance.window)
}

output "outage_count" {
  description = "Total number of outages created"
  value       = length(hyperping_outage.manual)
}

output "summary" {
  description = "Summary of all resources created by the module"
  value = {
    incidents         = length(hyperping_incident.template)
    maintenance       = length(hyperping_maintenance.window)
    outages           = length(hyperping_outage.manual)
    statuspage_linked = var.statuspage_id != null
  }
}
