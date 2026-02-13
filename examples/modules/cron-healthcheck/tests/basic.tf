# Basic Cron Healthcheck Module Test
#
# Tests basic functionality with simple cron schedules
#
# Usage:
#   export HYPERPING_API_KEY="your-api-key"
#   terraform init
#   terraform apply

module "basic_cron_jobs" {
  source = "../"

  jobs = {
    daily_backup = {
      cron     = "0 2 * * *"
      timezone = "America/New_York"
      grace    = 30
    }
    hourly_sync = {
      cron     = "0 * * * *"
      timezone = "UTC"
      grace    = 10
    }
    weekly_report = {
      cron     = "0 0 * * 0"
      timezone = "America/Los_Angeles"
      grace    = 60
    }
  }

  name_prefix = "TEST"
}

# Outputs for verification
output "healthcheck_count" {
  description = "Number of healthchecks created"
  value       = module.basic_cron_jobs.job_count
}

output "healthcheck_ids" {
  description = "Map of job names to healthcheck IDs"
  value       = module.basic_cron_jobs.healthcheck_ids
}

output "backup_ping_url" {
  description = "Ping URL for daily backup job"
  value       = module.basic_cron_jobs.ping_urls["daily_backup"]
  sensitive   = true
}

output "all_ping_urls" {
  description = "All ping URLs for cron integration"
  value       = module.basic_cron_jobs.ping_urls
  sensitive   = true
}

output "healthcheck_details" {
  description = "Full healthcheck details"
  value       = module.basic_cron_jobs.healthchecks
  sensitive   = true
}
