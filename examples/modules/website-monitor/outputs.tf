# Website Monitor Module - Outputs

output "monitor_ids" {
  description = "Map of page name to monitor UUID"
  value = {
    for k, v in hyperping_monitor.page : k => v.id
  }
}

output "monitor_ids_list" {
  description = "List of all monitor UUIDs"
  value       = [for v in hyperping_monitor.page : v.id]
}

output "monitors" {
  description = "Full monitor objects for advanced usage"
  value = {
    for k, v in hyperping_monitor.page : k => {
      id                      = v.id
      name                    = v.name
      url                     = v.url
      protocol                = v.protocol
      check_frequency         = v.check_frequency
      regions                 = v.regions
      paused                  = v.paused
      required_keyword        = v.required_keyword
      response_time_threshold = v.response_time_threshold
    }
  }
}

output "page_count" {
  description = "Total number of page monitors created"
  value       = length(hyperping_monitor.page)
}

output "domain" {
  description = "Domain being monitored"
  value       = var.domain
}

output "pages_monitored" {
  description = "List of page paths being monitored"
  value       = [for k, v in var.pages : v.path]
}

output "performance_monitoring_enabled" {
  description = "Whether performance thresholds are configured"
  value       = var.performance_threshold_ms != null || anytrue([for k, v in var.pages : v.performance_threshold_ms != null])
}
