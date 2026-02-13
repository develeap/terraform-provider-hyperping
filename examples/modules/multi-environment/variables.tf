# Multi-Environment Module - Variables

variable "service_name" {
  description = "Name of the service being monitored (e.g., 'UserAPI', 'PaymentService')"
  type        = string

  validation {
    condition     = length(var.service_name) > 0 && length(var.service_name) <= 100
    error_message = "Service name must be between 1 and 100 characters."
  }
}

variable "environments" {
  description = "Map of environment configurations. Key is environment name (e.g., 'dev', 'staging', 'prod')"
  type = map(object({
    url                  = string
    method               = optional(string)
    frequency            = optional(number)
    regions              = optional(list(string))
    expected_status_code = optional(string)
    follow_redirects     = optional(bool)
    headers              = optional(map(string))
    body                 = optional(string)
    required_keyword     = optional(string)
    alerts_wait          = optional(number)
    escalation_policy    = optional(string)
    paused               = optional(bool)
    enabled              = optional(bool, true)
  }))

  validation {
    condition     = length(var.environments) > 0
    error_message = "At least one environment must be specified."
  }

  validation {
    condition = alltrue([
      for env_name, env in var.environments : can(regex("^[a-zA-Z0-9-_]+$", env_name))
    ])
    error_message = "Environment names must contain only alphanumeric characters, hyphens, and underscores."
  }

  validation {
    condition = alltrue([
      for env_name, env in var.environments : can(regex("^https?://", env.url))
    ])
    error_message = "All environment URLs must start with http:// or https://."
  }

  validation {
    condition = alltrue([
      for env_name, env in var.environments :
      env.frequency == null || contains([
        10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400
      ], env.frequency)
    ])
    error_message = "Environment frequency must be one of: 10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400 seconds."
  }
}

variable "name_format" {
  description = "Custom name format string. Use first %s for environment, second %s for service_name. Example: '[%s] %s Health Check'"
  type        = string
  default     = ""
}

variable "use_workspace_name" {
  description = "Use Terraform workspace name instead of environment key for naming. Useful when using workspace-per-environment pattern."
  type        = bool
  default     = false
}

# Default values for all environments (can be overridden per environment)

variable "default_method" {
  description = "Default HTTP method for all environments"
  type        = string
  default     = "GET"

  validation {
    condition     = contains(["GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"], var.default_method)
    error_message = "Method must be one of: GET, POST, PUT, PATCH, DELETE, HEAD."
  }
}

variable "default_frequency" {
  description = "Default check frequency in seconds for all environments"
  type        = number
  default     = 60

  validation {
    condition = contains([
      10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400
    ], var.default_frequency)
    error_message = "Frequency must be one of: 10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400 seconds."
  }
}

variable "default_regions" {
  description = "Default monitoring regions for all environments"
  type        = list(string)
  default     = ["virginia", "london", "frankfurt"]

  validation {
    condition = alltrue([
      for r in var.default_regions : contains([
        "paris", "frankfurt", "amsterdam", "london",
        "singapore", "sydney", "tokyo", "seoul", "mumbai", "bangalore",
        "virginia", "california", "sanfrancisco", "nyc", "toronto", "saopaulo",
        "bahrain", "capetown"
      ], r)
    ])
    error_message = "Invalid region specified."
  }
}

variable "default_expected_status_code" {
  description = "Default expected HTTP status code for all environments"
  type        = string
  default     = "2xx"
}

variable "default_follow_redirects" {
  description = "Default redirect following behavior for all environments"
  type        = bool
  default     = true
}

variable "default_headers" {
  description = "Default HTTP headers for all environments"
  type        = map(string)
  default     = null
}

variable "default_body" {
  description = "Default request body for all environments (for POST/PUT/PATCH methods)"
  type        = string
  default     = null
}

variable "default_required_keyword" {
  description = "Default required keyword in response for all environments"
  type        = string
  default     = null
}

variable "default_alerts_wait" {
  description = "Default seconds to wait before alerting after outage detection for all environments (0 = immediate)"
  type        = number
  default     = null
}

variable "default_escalation_policy" {
  description = "Default UUID of escalation policy for all environments"
  type        = string
  default     = null
}

variable "default_paused" {
  description = "Default paused state for all environments"
  type        = bool
  default     = false
}
