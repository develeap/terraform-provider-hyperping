# Cron Healthcheck Module - Outputs

output "healthcheck_ids" {
  description = "Map of job name to healthcheck UUID"
  value = {
    for k, v in hyperping_healthcheck.job : k => v.id
  }
}

output "healthcheck_ids_list" {
  description = "List of all healthcheck UUIDs"
  value       = [for v in hyperping_healthcheck.job : v.id]
}

output "ping_urls" {
  description = "Map of job name to ping URL for use in cron scripts (curl $PING_URL)"
  value = {
    for k, v in hyperping_healthcheck.job : k => v.ping_url
  }
  sensitive = true
}

output "ping_urls_list" {
  description = "List of all ping URLs"
  value       = [for v in hyperping_healthcheck.job : v.ping_url]
  sensitive   = true
}

output "healthchecks" {
  description = "Full healthcheck objects for advanced usage"
  value = {
    for k, v in hyperping_healthcheck.job : k => {
      id                 = v.id
      name               = v.name
      cron               = v.cron
      timezone           = v.timezone
      grace_period_value = v.grace_period_value
      grace_period_type  = v.grace_period_type
      ping_url           = v.ping_url
      is_paused          = v.is_paused
    }
  }
  sensitive = true
}

output "job_count" {
  description = "Total number of healthchecks created"
  value       = length(hyperping_healthcheck.job)
}
