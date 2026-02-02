# Create a manual outage for planned maintenance
resource "hyperping_outage" "planned_maintenance" {
  monitor_uuid = "mon_abc123def456"
  start_date   = "2026-02-15T02:00:00Z"
  end_date     = "2026-02-15T04:00:00Z"
  status_code  = 503
  description  = "Planned database migration - service temporarily unavailable"
}

# Create an ongoing outage (no end_date)
resource "hyperping_outage" "incident_tracking" {
  monitor_uuid = hyperping_monitor.api.id
  start_date   = "2026-01-29T10:00:00Z"
  status_code  = 500
  description  = "API gateway experiencing intermittent 500 errors"
}

# Note: All outage fields are ForceNew - any change triggers destroy and recreate
# To modify an outage, update it via the Hyperping dashboard or API directly
