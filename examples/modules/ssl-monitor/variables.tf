# SSL Monitor Module - Variables

variable "domains" {
  description = "List of domains to monitor for SSL certificate health"
  type        = list(string)

  validation {
    condition     = length(var.domains) > 0
    error_message = "At least one domain must be specified."
  }

  validation {
    condition = alltrue([
      for d in var.domains : can(regex("^[a-zA-Z0-9][a-zA-Z0-9.-]+[a-zA-Z0-9]$", d))
    ])
    error_message = "Invalid domain format. Domains should not include protocol (https://)."
  }
}

variable "name_prefix" {
  description = "Prefix for monitor names"
  type        = string
  default     = "SSL"
}

variable "check_frequency" {
  description = "Check frequency in seconds (recommended: 3600 for hourly)"
  type        = number
  default     = 3600

  validation {
    condition = contains([
      10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400
    ], var.check_frequency)
    error_message = "Frequency must be one of: 10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400 seconds."
  }
}

variable "regions" {
  description = "Regions to check from"
  type        = list(string)
  default     = ["virginia", "london"]

  validation {
    condition = alltrue([
      for r in var.regions : contains([
        "virginia", "london", "frankfurt", "singapore",
        "sydney", "tokyo", "saopaulo", "oregon", "bahrain"
      ], r)
    ])
    error_message = "Invalid region specified."
  }
}

variable "port" {
  description = "HTTPS port to check"
  type        = number
  default     = 443
}

variable "alerts_wait" {
  description = "Number of failed checks before alerting"
  type        = number
  default     = 1
}

variable "escalation_policy_uuid" {
  description = "UUID of escalation policy for SSL alerts"
  type        = string
  default     = null
}

variable "paused" {
  description = "Create monitors in paused state"
  type        = bool
  default     = false
}
