# Cron Healthcheck Module - Variables

variable "jobs" {
  description = "Map of cron job configurations to monitor"
  type = map(object({
    cron              = string
    timezone          = optional(string)
    grace             = optional(number)
    escalation_policy = optional(string)
    paused            = optional(bool)
  }))

  validation {
    condition     = length(var.jobs) > 0
    error_message = "At least one job must be specified."
  }

  validation {
    condition = alltrue([
      for k, v in var.jobs : can(regex("^[^\\s]+\\s+[^\\s]+\\s+[^\\s]+\\s+[^\\s]+\\s+[^\\s]+$", v.cron))
    ])
    error_message = "All cron schedules must be in standard 5-field format (minute hour day month weekday)."
  }

  validation {
    condition = alltrue([
      for k, v in var.jobs : v.grace == null || (v.grace >= 1 && v.grace <= 1440)
    ])
    error_message = "Grace period must be between 1 and 1440 minutes (24 hours)."
  }
}

variable "name_prefix" {
  description = "Prefix for healthcheck names (e.g., 'prod', 'staging')"
  type        = string
  default     = ""

  validation {
    condition     = can(regex("^[a-zA-Z0-9-_]*$", var.name_prefix))
    error_message = "Name prefix must contain only alphanumeric characters, hyphens, and underscores."
  }
}

variable "name_format" {
  description = "Custom name format string. Use %s for the job key. Overrides name_prefix if set."
  type        = string
  default     = ""
}

variable "default_timezone" {
  description = "Default timezone for cron schedules (IANA timezone, e.g., 'America/New_York', 'UTC')"
  type        = string
  default     = "UTC"

  validation {
    condition = contains([
      "UTC",
      "America/New_York", "America/Chicago", "America/Denver", "America/Los_Angeles",
      "America/Toronto", "America/Sao_Paulo", "America/Mexico_City",
      "Europe/London", "Europe/Paris", "Europe/Berlin", "Europe/Amsterdam", "Europe/Rome",
      "Asia/Tokyo", "Asia/Singapore", "Asia/Shanghai", "Asia/Hong_Kong", "Asia/Dubai",
      "Asia/Kolkata", "Asia/Bangkok", "Asia/Seoul", "Asia/Manila",
      "Australia/Sydney", "Australia/Melbourne", "Pacific/Auckland"
    ], var.default_timezone)
    error_message = "Invalid timezone. Must be a valid IANA timezone identifier."
  }
}

variable "default_grace_minutes" {
  description = "Default grace period in minutes (time allowed after expected run before alerting)"
  type        = number
  default     = 15

  validation {
    condition     = var.default_grace_minutes >= 1 && var.default_grace_minutes <= 1440
    error_message = "Grace period must be between 1 and 1440 minutes (24 hours)."
  }
}

variable "escalation_policy" {
  description = "UUID of escalation policy to use for all healthchecks"
  type        = string
  default     = null
}

variable "paused" {
  description = "Create healthchecks in paused state"
  type        = bool
  default     = false
}
