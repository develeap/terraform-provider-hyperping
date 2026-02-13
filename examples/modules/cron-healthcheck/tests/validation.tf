# Validation Tests for Cron Healthcheck Module
#
# These tests verify input validation works correctly.
#
# Usage:
#   terraform init
#   terraform validate (should pass)
#
# To test validation errors, uncomment invalid examples at the bottom

# Test 1: Valid configurations (should succeed)
module "valid_jobs" {
  source = "../"

  jobs = {
    test1 = {
      cron     = "0 * * * *"
      timezone = "UTC"
      grace    = 15
    }
    test2 = {
      cron     = "*/30 * * * *"
      timezone = "America/New_York"
      grace    = 10
    }
  }
}

# Test 2: Minimum values (should succeed)
module "minimum_values" {
  source = "../"

  jobs = {
    minimal = {
      cron  = "0 0 * * *"
      grace = 1 # Minimum grace period
    }
  }
}

# Test 3: Maximum values (should succeed)
module "maximum_values" {
  source = "../"

  jobs = {
    maximal = {
      cron  = "0 0 * * *"
      grace = 1440 # Maximum grace period (24 hours)
    }
  }
}

# Test 4: Complex cron expressions (should succeed)
module "complex_cron" {
  source = "../"

  jobs = {
    every_15_min = {
      cron = "*/15 * * * *"
    }
    weekdays_9am = {
      cron = "0 9 * * 1-5"
    }
    first_of_month = {
      cron = "0 0 1 * *"
    }
    quarterly = {
      cron = "0 0 1 */3 *"
    }
  }
}

# Test 5: All supported timezones (should succeed)
module "timezone_test" {
  source = "../"

  jobs = {
    utc       = { cron = "0 0 * * *", timezone = "UTC" }
    ny        = { cron = "0 0 * * *", timezone = "America/New_York" }
    chicago   = { cron = "0 0 * * *", timezone = "America/Chicago" }
    denver    = { cron = "0 0 * * *", timezone = "America/Denver" }
    la        = { cron = "0 0 * * *", timezone = "America/Los_Angeles" }
    toronto   = { cron = "0 0 * * *", timezone = "America/Toronto" }
    saopaulo  = { cron = "0 0 * * *", timezone = "America/Sao_Paulo" }
    london    = { cron = "0 0 * * *", timezone = "Europe/London" }
    paris     = { cron = "0 0 * * *", timezone = "Europe/Paris" }
    berlin    = { cron = "0 0 * * *", timezone = "Europe/Berlin" }
    tokyo     = { cron = "0 0 * * *", timezone = "Asia/Tokyo" }
    singapore = { cron = "0 0 * * *", timezone = "Asia/Singapore" }
    sydney    = { cron = "0 0 * * *", timezone = "Australia/Sydney" }
  }
}

# Test 6: Custom name formats (should succeed)
module "name_format_test" {
  source = "../"

  jobs = {
    test = { cron = "0 0 * * *" }
  }

  name_format = "CUSTOM-%s-FORMAT"
}

# Test 7: Name prefix validation (should succeed)
module "valid_name_prefix" {
  source = "../"

  jobs = {
    test = { cron = "0 0 * * *" }
  }

  name_prefix = "prod-env_123"
}

# Test 8: Paused jobs (should succeed)
module "paused_jobs" {
  source = "../"

  jobs = {
    paused_job = {
      cron   = "0 0 * * *"
      paused = true
    }
  }
}

# Test 9: Override defaults (should succeed)
module "override_defaults" {
  source = "../"

  jobs = {
    override1 = {
      cron     = "0 0 * * *"
      timezone = "Europe/Paris"
      grace    = 60
    }
  }

  default_timezone      = "UTC"
  default_grace_minutes = 15
}

# -----------------------------------------------------------------------------
# INVALID CONFIGURATIONS (commented out - uncomment to test validation)
# These should fail terraform validate
# -----------------------------------------------------------------------------

# # FAIL: Invalid cron format (4 fields instead of 5)
# module "invalid_cron_format" {
#   source = "../"
#
#   jobs = {
#     bad = {
#       cron = "0 * * *"  # Missing weekday field
#     }
#   }
# }

# # FAIL: Grace period too small
# module "invalid_grace_min" {
#   source = "../"
#
#   jobs = {
#     bad = {
#       cron  = "0 0 * * *"
#       grace = 0  # Must be >= 1
#     }
#   }
# }

# # FAIL: Grace period too large
# module "invalid_grace_max" {
#   source = "../"
#
#   jobs = {
#     bad = {
#       cron  = "0 0 * * *"
#       grace = 1441  # Must be <= 1440
#     }
#   }
# }

# # FAIL: Invalid timezone
# module "invalid_timezone" {
#   source = "../"
#
#   jobs = {
#     bad = {
#       cron     = "0 0 * * *"
#       timezone = "Invalid/Timezone"
#     }
#   }
# }

# # FAIL: Invalid name prefix (contains spaces)
# module "invalid_name_prefix" {
#   source = "../"
#
#   jobs = {
#     test = { cron = "0 0 * * *" }
#   }
#
#   name_prefix = "invalid prefix"
# }

# # FAIL: No jobs specified
# module "no_jobs" {
#   source = "../"
#
#   jobs = {}
# }

# # FAIL: Default grace period out of range
# module "invalid_default_grace" {
#   source = "../"
#
#   jobs = {
#     test = { cron = "0 0 * * *" }
#   }
#
#   default_grace_minutes = 2000
# }
