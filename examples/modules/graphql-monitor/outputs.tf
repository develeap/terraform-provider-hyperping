# GraphQL Monitor Module - Outputs

output "monitor_ids" {
  description = "Map of query name to monitor UUID"
  value = {
    for k, v in hyperping_monitor.graphql_query : k => v.id
  }
}

output "monitor_ids_list" {
  description = "List of all query monitor UUIDs"
  value       = [for v in hyperping_monitor.graphql_query : v.id]
}

output "introspection_monitor_id" {
  description = "UUID of introspection monitor (if enabled)"
  value       = var.enable_introspection_check ? hyperping_monitor.introspection[0].id : null
}

output "all_monitor_ids" {
  description = "List of all monitor UUIDs including introspection"
  value = concat(
    [for v in hyperping_monitor.graphql_query : v.id],
    var.enable_introspection_check ? [hyperping_monitor.introspection[0].id] : []
  )
}

output "monitors" {
  description = "Full monitor objects for advanced usage"
  value = {
    for k, v in hyperping_monitor.graphql_query : k => {
      id              = v.id
      name            = v.name
      url             = v.url
      protocol        = v.protocol
      check_frequency = v.check_frequency
      regions         = v.regions
      paused          = v.paused
      required_keyword = v.required_keyword
    }
  }
}

output "endpoint" {
  description = "GraphQL endpoint URL being monitored"
  value       = var.endpoint
}

output "query_count" {
  description = "Total number of query monitors created"
  value       = length(hyperping_monitor.graphql_query)
}

output "total_monitor_count" {
  description = "Total number of monitors created (queries + introspection)"
  value       = length(hyperping_monitor.graphql_query) + (var.enable_introspection_check ? 1 : 0)
}
