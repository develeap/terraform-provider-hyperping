# Look up a specific outage by ID
data "hyperping_outage" "last_incident" {
  id = "out_abc123def456"
}

# Use outage data for incident reporting
output "incident_details" {
  value = {
    monitor_name      = data.hyperping_outage.last_incident.monitor.name
    started_at        = data.hyperping_outage.last_incident.start_date
    duration_seconds  = data.hyperping_outage.last_incident.duration_ms / 1000
    status_code       = data.hyperping_outage.last_incident.status_code
    is_resolved       = data.hyperping_outage.last_incident.is_resolved
    acknowledged_by   = data.hyperping_outage.last_incident.acknowledged_by
  }
}
