# API Health Module - Variables

variable "endpoints" {
  description = "Map of endpoint configurations to monitor"
  type = map(object({
    url                  = string
    method               = optional(string, "GET")
    frequency            = optional(number)
    expected_status_code = optional(string, "2xx")
    follow_redirects     = optional(bool, true)
    headers              = optional(map(string))
    body                 = optional(string)
    required_keyword     = optional(string)
    regions              = optional(list(string))
    paused               = optional(bool)
  }))

  validation {
    condition     = length(var.endpoints) > 0
    error_message = "At least one endpoint must be specified."
  }

  validation {
    condition = alltrue([
      for k, v in var.endpoints : can(regex("^https?://", v.url))
    ])
    error_message = "All endpoint URLs must start with http:// or https://."
  }
}

variable "name_prefix" {
  description = "Prefix for monitor names (e.g., 'prod', 'staging')"
  type        = string
  default     = ""

  validation {
    condition     = can(regex("^[a-zA-Z0-9-_]*$", var.name_prefix))
    error_message = "Name prefix must contain only alphanumeric characters, hyphens, and underscores."
  }
}

variable "name_format" {
  description = "Custom name format string. Use %s for the endpoint key. Overrides name_prefix if set."
  type        = string
  default     = ""
}

variable "default_regions" {
  description = "Default regions for monitors (can be overridden per endpoint)"
  type        = list(string)
  default     = ["virginia", "london", "frankfurt"]

  validation {
    condition = alltrue([
      for r in var.default_regions : contains([
        "paris", "frankfurt", "amsterdam", "london",
        "singapore", "sydney", "tokyo", "seoul", "mumbai", "bangalore",
        "virginia", "california", "sanfrancisco", "oregon", "nyc", "toronto", "saopaulo",
        "bahrain", "capetown"
      ], r)
    ])
    error_message = "Invalid region specified."
  }
}

variable "default_frequency" {
  description = "Default check frequency in seconds"
  type        = number
  default     = 60

  validation {
    condition = contains([
      10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400
    ], var.default_frequency)
    error_message = "Frequency must be one of: 10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400 seconds."
  }
}

variable "alerts_wait" {
  description = "Seconds to wait before alerting after outage detection (0 = immediate)"
  type        = number
  default     = null
}

variable "escalation_policy" {
  description = "UUID of escalation policy to use for all monitors"
  type        = string
  default     = null
}

variable "paused" {
  description = "Create monitors in paused state"
  type        = bool
  default     = false
}
