# Advanced Cron Healthcheck Module Test
#
# Tests advanced features including custom formats, escalation policies,
# and paused jobs
#
# Usage:
#   export HYPERPING_API_KEY="your-api-key"
#   terraform init
#   terraform apply

# Example: Create an escalation policy first
# (In real usage, this would already exist or be created separately)
# resource "hyperping_escalation_policy" "critical" {
#   name = "Critical Jobs Policy"
#   # ... escalation policy configuration
# }

module "production_cron_jobs" {
  source = "../"

  jobs = {
    payment_processing = {
      cron     = "*/15 * * * *" # Every 15 minutes
      timezone = "UTC"
      grace    = 5
    }
    fraud_detection = {
      cron     = "*/30 * * * *" # Every 30 minutes
      timezone = "UTC"
      grace    = 10
    }
    db_backup_postgres = {
      cron     = "0 3 * * *" # Daily at 3 AM
      timezone = "America/New_York"
      grace    = 60
    }
    db_backup_mysql = {
      cron     = "0 4 * * *" # Daily at 4 AM
      timezone = "America/New_York"
      grace    = 60
    }
    log_rotation = {
      cron     = "0 0 * * 0" # Weekly on Sunday
      timezone = "UTC"
      grace    = 120
    }
  }

  name_format           = "[PROD] Cron - %s"
  default_timezone      = "UTC"
  default_grace_minutes = 15
  # escalation_policy     = hyperping_escalation_policy.critical.id
}

module "staging_cron_jobs" {
  source = "../"

  jobs = {
    nightly_tests = {
      cron     = "0 0 * * *"
      timezone = "America/Los_Angeles"
      grace    = 30
    }
    cache_warmup = {
      cron     = "0 6 * * *"
      timezone = "America/Los_Angeles"
      grace    = 15
    }
  }

  name_prefix = "STAGING"
}

module "maintenance_jobs" {
  source = "../"

  jobs = {
    seasonal_cleanup = {
      cron   = "0 0 1 * *" # Monthly on 1st
      grace  = 120
      paused = true # Currently disabled
    }
    quarterly_report = {
      cron   = "0 0 1 */3 *" # Every 3 months
      grace  = 240
      paused = true
    }
  }

  name_format = "MAINT - %s"
  paused      = true # All maintenance jobs paused by default
}

# Custom name format examples
module "custom_names" {
  source = "../"

  jobs = {
    user-export = {
      cron  = "0 1 * * *"
      grace = 30
    }
    inventory-sync = {
      cron  = "0 */6 * * *"
      grace = 20
    }
  }

  name_format = "CRON::%s::healthcheck"
  # Results in: "CRON::user-export::healthcheck"
}

# Outputs
output "prod_job_count" {
  value = module.production_cron_jobs.job_count
}

output "prod_healthcheck_ids" {
  value = module.production_cron_jobs.healthcheck_ids
}

output "prod_ping_urls" {
  value     = module.production_cron_jobs.ping_urls
  sensitive = true
}

output "staging_ping_urls" {
  value     = module.staging_cron_jobs.ping_urls
  sensitive = true
}

output "all_healthchecks" {
  description = "All healthcheck details across all modules"
  value = {
    production  = module.production_cron_jobs.healthchecks
    staging     = module.staging_cron_jobs.healthchecks
    maintenance = module.maintenance_jobs.healthchecks
    custom      = module.custom_names.healthchecks
  }
  sensitive = true
}
