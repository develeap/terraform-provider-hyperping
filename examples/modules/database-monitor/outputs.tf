# Database Monitor Module - Outputs

output "monitor_ids" {
  description = "Map of database name to monitor UUID"
  value = {
    for k, v in hyperping_monitor.database : k => v.id
  }
}

output "monitor_ids_list" {
  description = "List of all monitor UUIDs"
  value       = [for v in hyperping_monitor.database : v.id]
}

output "monitors" {
  description = "Full monitor objects for advanced usage"
  value = {
    for k, v in hyperping_monitor.database : k => {
      id              = v.id
      name            = v.name
      protocol        = v.protocol
      url             = v.url
      port            = v.port
      check_frequency = v.check_frequency
      regions         = v.regions
      paused          = v.paused
    }
  }
}

output "database_count" {
  description = "Total number of database monitors created"
  value       = length(hyperping_monitor.database)
}

output "monitors_by_type" {
  description = "Database monitors grouped by type"
  value = {
    for type in distinct([for k, v in var.databases : lower(v.type)]) : type => {
      for k, v in var.databases : k => hyperping_monitor.database[k].id
      if lower(v.type) == type
    }
  }
}

output "connection_strings" {
  description = "Database connection endpoints (host:port)"
  value = {
    for k, v in hyperping_monitor.database : k => "${v.url}:${v.port}"
  }
}
