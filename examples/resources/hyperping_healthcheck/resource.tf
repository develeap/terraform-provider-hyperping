# Create a healthcheck with cron schedule
resource "hyperping_healthcheck" "daily_backup" {
  name             = "Daily Backup Job"
  cron             = "0 2 * * *" # 2 AM every day
  tz               = "America/New_York"
  grace_period_value = 30
  grace_period_type  = "minutes"
}

# Create a healthcheck with period-based schedule
resource "hyperping_healthcheck" "hourly_sync" {
  name             = "Hourly Data Sync"
  period_value     = 1
  period_type      = "hours"
  grace_period_value = 15
  grace_period_type  = "minutes"
  escalation_policy  = "ep_abc123def456" # Optional
}

# Paused healthcheck example
resource "hyperping_healthcheck" "maintenance_job" {
  name             = "Maintenance Job (Paused)"
  period_value     = 7
  period_type      = "days"
  grace_period_value = 1
  grace_period_type  = "hours"
  is_paused        = true
}

# Output the ping URL for use in cron jobs
output "backup_ping_url" {
  value       = hyperping_healthcheck.daily_backup.ping_url
  description = "Add this URL to your backup script: curl $PING_URL"
  sensitive   = true
}
