# Status Page Complete Module - Outputs

output "statuspage_id" {
  description = "Status page UUID"
  value       = hyperping_statuspage.main.id
}

output "statuspage_url" {
  description = "Status page public URL"
  value       = var.hostname != null ? "https://${var.hostname}" : "https://${var.subdomain}.hyperping.app"
}

output "statuspage_subdomain" {
  description = "Status page subdomain"
  value       = var.subdomain
}

output "monitor_ids" {
  description = "Map of service name to monitor UUID"
  value = {
    for k, v in hyperping_monitor.service : k => v.id
  }
}

output "monitor_ids_list" {
  description = "List of all monitor UUIDs"
  value       = [for v in hyperping_monitor.service : v.id]
}

output "monitors" {
  description = "Full monitor objects with details"
  value = {
    for k, v in hyperping_monitor.service : k => {
      id              = v.id
      name            = v.name
      url             = v.url
      check_frequency = v.check_frequency
      regions         = v.regions
    }
  }
}

output "service_count" {
  description = "Number of services monitored"
  value       = length(hyperping_monitor.service)
}
