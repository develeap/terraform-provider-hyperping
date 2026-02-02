# Look up a specific healthcheck by ID
data "hyperping_healthcheck" "backup_job" {
  id = "tok_abc123def456"
}

# Use the data to configure alerts
output "backup_status" {
  value = {
    name     = data.hyperping_healthcheck.backup_job.name
    is_down  = data.hyperping_healthcheck.backup_job.is_down
    is_paused = data.hyperping_healthcheck.backup_job.is_paused
    last_ping = data.hyperping_healthcheck.backup_job.last_ping
  }
}
