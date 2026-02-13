# Cron Healthcheck Module - Main
#
# Creates dead man's switch monitors for cron jobs.
#
# Usage:
#   module "cron_jobs" {
#     source = "path/to/modules/cron-healthcheck"
#
#     jobs = {
#       daily_backup = {
#         cron     = "0 2 * * *"
#         timezone = "America/New_York"
#         grace    = 30
#       }
#       hourly_sync = {
#         cron     = "0 * * * *"
#         timezone = "UTC"
#         grace    = 10
#       }
#     }
#   }

locals {
  # Build the full name with format string
  name_format = var.name_format != "" ? var.name_format : (var.name_prefix != "" ? "[${upper(var.name_prefix)}] %s" : "%s")
}

resource "hyperping_healthcheck" "job" {
  for_each = var.jobs

  name = format(local.name_format, each.key)

  # Cron schedule configuration
  cron     = each.value.cron
  timezone = coalesce(each.value.timezone, var.default_timezone)

  # Grace period configuration
  grace_period_value = coalesce(each.value.grace, var.default_grace_minutes)
  grace_period_type  = "minutes"

  # Optional: escalation policy
  escalation_policy = coalesce(each.value.escalation_policy, var.escalation_policy)

  # Paused state
  is_paused = coalesce(each.value.paused, var.paused)

  lifecycle {
    create_before_destroy = true
  }
}
