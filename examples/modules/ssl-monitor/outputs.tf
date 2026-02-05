# SSL Monitor Module - Outputs

output "monitor_ids" {
  description = "Map of domain to monitor UUID"
  value = {
    for k, v in hyperping_monitor.ssl : k => v.id
  }
}

output "monitor_ids_list" {
  description = "List of all monitor UUIDs"
  value       = [for v in hyperping_monitor.ssl : v.id]
}

output "monitors" {
  description = "Full monitor objects"
  value = {
    for k, v in hyperping_monitor.ssl : k => {
      id              = v.id
      name            = v.name
      url             = v.url
      check_frequency = v.check_frequency
      regions         = v.regions
      paused          = v.paused
    }
  }
}

output "monitored_domains" {
  description = "List of domains being monitored"
  value       = var.domains
}

output "monitor_count" {
  description = "Total number of SSL monitors created"
  value       = length(hyperping_monitor.ssl)
}
