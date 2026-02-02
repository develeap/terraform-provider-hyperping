# Basic maintenance window
resource "hyperping_maintenance" "example" {
  name       = "database-maintenance"
  title      = "Database Maintenance"
  text       = "Routine database maintenance window"
  start_date = "2026-01-20T02:00:00.000Z"
  end_date   = "2026-01-20T04:00:00.000Z"
  monitors   = [hyperping_monitor.database.id]
}

# Full maintenance window with all options
resource "hyperping_maintenance" "full" {
  name                 = "infrastructure-upgrade"
  title                = "Infrastructure Upgrade"
  text                 = "Upgrading cloud infrastructure for improved performance"
  start_date           = "2026-01-25T01:00:00.000Z"
  end_date             = "2026-01-25T05:00:00.000Z"
  monitors             = [hyperping_monitor.api.id, hyperping_monitor.web.id]
  status_pages         = [hyperping_status_page.main.id]
  notification_option  = "scheduled"
  notification_minutes = 120
}

# Maintenance with immediate notification
resource "hyperping_maintenance" "urgent" {
  name                = "emergency-maintenance"
  title               = "Emergency Network Maintenance"
  text                = "Addressing critical network issues"
  start_date          = "2026-01-15T00:00:00.000Z"
  end_date            = "2026-01-15T06:00:00.000Z"
  monitors            = [hyperping_monitor.network.id]
  notification_option = "immediate"
}

# Output maintenance details
output "maintenance_id" {
  description = "The ID of the scheduled maintenance"
  value       = hyperping_maintenance.example.id
}

output "maintenance_name" {
  description = "The name of the maintenance window"
  value       = hyperping_maintenance.example.name
}
