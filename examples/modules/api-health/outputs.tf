# API Health Module - Outputs

output "monitor_ids" {
  description = "Map of endpoint name to monitor UUID"
  value = {
    for k, v in hyperping_monitor.endpoint : k => v.id
  }
}

output "monitor_names" {
  description = "Map of endpoint name to full monitor name"
  value = {
    for k, v in hyperping_monitor.endpoint : k => v.name
  }
}

output "monitor_urls" {
  description = "Map of endpoint name to monitored URL"
  value = {
    for k, v in hyperping_monitor.endpoint : k => v.url
  }
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

output "active_monitors" {
  description = "List of active (non-paused) monitor IDs"
  value = [
    for k, v in hyperping_monitor.endpoint : v.id
    if !v.paused
  ]
}

output "monitor_count" {
  description = "Total number of monitors created"
  value       = length(hyperping_monitor.endpoint)
}
