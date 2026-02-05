# API Health Module - Outputs

output "monitor_ids" {
  description = "Map of endpoint name to monitor UUID"
  value = {
    for k, v in hyperping_monitor.endpoint : k => v.id
  }
}

output "monitor_ids_list" {
  description = "List of all monitor UUIDs"
  value       = [for v in hyperping_monitor.endpoint : v.id]
}

output "monitors" {
  description = "Full monitor objects for advanced usage"
  value = {
    for k, v in hyperping_monitor.endpoint : k => {
      id              = v.id
      name            = v.name
      url             = v.url
      protocol        = v.protocol
      check_frequency = v.check_frequency
      regions         = v.regions
      paused          = v.paused
    }
  }
}

output "endpoint_count" {
  description = "Total number of monitors created"
  value       = length(hyperping_monitor.endpoint)
}
