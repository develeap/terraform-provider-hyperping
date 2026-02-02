# Tenant Module Outputs

output "tenant_id" {
  description = "Tenant identifier"
  value       = var.tenant_id
}

output "tenant_name" {
  description = "Tenant display name"
  value       = var.tenant_name
}

output "status_page_uuid" {
  description = "Status page UUID (for incidents)"
  value       = var.status_page.uuid
}

output "monitor_ids" {
  description = "Map of monitor name to UUID"
  value = {
    for name, monitor in hyperping_monitor.tenant : name => monitor.id
  }
}

output "monitor_count" {
  description = "Number of monitors created"
  value       = length(hyperping_monitor.tenant)
}

output "component_uuids" {
  description = "Map of component name to UUID"
  value       = local.component_uuids
}

output "monitor_to_component" {
  description = "Map of monitor name to component name"
  value       = local.monitor_to_component
}

# Useful for creating incidents with affected components
output "affected_components_for_incident" {
  description = "Component UUIDs for use in incidents (only those with UUIDs)"
  value = [
    for c in var.status_page.components : c.uuid
    if c.uuid != null
  ]
}

# Full registry data (can be exported to state.json)
output "registry_data" {
  description = "Complete registry data for this tenant"
  value = {
    tenant_id             = var.tenant_id
    tenant_name           = var.tenant_name
    status_page_uuid      = var.status_page.uuid
    status_page_subdomain = var.status_page.subdomain
    monitor_uuids         = [for m in hyperping_monitor.tenant : m.id]
    component_uuids       = local.component_uuids
  }
}
