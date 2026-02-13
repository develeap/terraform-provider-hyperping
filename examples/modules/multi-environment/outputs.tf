# Multi-Environment Module - Outputs

output "monitor_ids" {
  description = "Map of environment name to monitor UUID"
  value = {
    for env_name, monitor in hyperping_monitor.environment : env_name => monitor.id
  }
}

output "monitor_ids_list" {
  description = "List of all monitor UUIDs across all environments"
  value       = [for monitor in hyperping_monitor.environment : monitor.id]
}

output "monitors" {
  description = "Full monitor objects for advanced usage"
  value = {
    for env_name, monitor in hyperping_monitor.environment : env_name => {
      id              = monitor.id
      name            = monitor.name
      url             = monitor.url
      protocol        = monitor.protocol
      check_frequency = monitor.check_frequency
      regions         = monitor.regions
      paused          = monitor.paused
      environment     = env_name
    }
  }
}

output "environments" {
  description = "List of enabled environment names"
  value       = keys(hyperping_monitor.environment)
}

output "environment_count" {
  description = "Total number of environments (monitors) created"
  value       = length(hyperping_monitor.environment)
}

output "service_name" {
  description = "Service name being monitored"
  value       = var.service_name
}

# Organized by environment for easy reference
output "by_environment" {
  description = "Environment-specific details organized by environment name"
  value = {
    for env_name, monitor in hyperping_monitor.environment : env_name => {
      monitor_id = monitor.id
      url        = monitor.url
      regions    = monitor.regions
      frequency  = monitor.check_frequency
      paused     = monitor.paused
    }
  }
}

# Useful for creating incidents or maintenance windows affecting specific environments
output "production_monitor_ids" {
  description = "Monitor IDs for production-like environments (prod, production)"
  value = [
    for env_name, monitor in hyperping_monitor.environment : monitor.id
    if contains(["prod", "production"], lower(env_name))
  ]
}

output "non_production_monitor_ids" {
  description = "Monitor IDs for non-production environments (dev, staging, test, etc.)"
  value = [
    for env_name, monitor in hyperping_monitor.environment : monitor.id
    if !contains(["prod", "production"], lower(env_name))
  ]
}
