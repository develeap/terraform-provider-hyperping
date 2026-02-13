# CDN Monitor Module - Outputs

output "monitor_ids" {
  description = "Map of asset name to monitor UUID"
  value = {
    for k, v in hyperping_monitor.cdn_asset : k => v.id
  }
}

output "monitor_ids_list" {
  description = "List of all monitor UUIDs (assets only)"
  value       = [for v in hyperping_monitor.cdn_asset : v.id]
}

output "all_monitor_ids" {
  description = "List of all monitor UUIDs including root domain if enabled"
  value = concat(
    [for v in hyperping_monitor.cdn_asset : v.id],
    [for v in hyperping_monitor.cdn_root : v.id]
  )
}

output "monitors" {
  description = "Full monitor objects for advanced usage"
  value = {
    for k, v in hyperping_monitor.cdn_asset : k => {
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

output "root_monitor_id" {
  description = "UUID of root domain monitor (if enabled)"
  value       = var.monitor_root_domain ? hyperping_monitor.cdn_root[0].id : null
}

output "cdn_domain" {
  description = "CDN domain being monitored"
  value       = var.cdn_domain
}

output "asset_count" {
  description = "Total number of asset monitors created"
  value       = length(hyperping_monitor.cdn_asset)
}

output "total_monitor_count" {
  description = "Total number of monitors created (including root domain if enabled)"
  value       = length(hyperping_monitor.cdn_asset) + length(hyperping_monitor.cdn_root)
}

output "monitored_assets" {
  description = "Map of asset names to their full URLs"
  value = {
    for k, v in var.assets : k => "${var.protocol}://${var.cdn_domain}${v}"
  }
}
