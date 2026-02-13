# Website Monitor Module - Variables

variable "domain" {
  description = "Domain name to monitor (e.g., 'example.com')"
  type        = string

  validation {
    condition     = can(regex("^[a-zA-Z0-9][a-zA-Z0-9-\\.]*[a-zA-Z0-9]$", var.domain))
    error_message = "Domain must be a valid hostname."
  }
}

variable "protocol" {
  description = "Protocol to use for requests (http or https)"
  type        = string
  default     = "https"

  validation {
    condition     = contains(["http", "https"], var.protocol)
    error_message = "Protocol must be either 'http' or 'https'."
  }
}

variable "pages" {
  description = "Map of page configurations to monitor"
  type = map(object({
    path                     = string
    expected_text            = optional(string)
    expected_status          = optional(string)
    method                   = optional(string)
    frequency                = optional(number)
    performance_threshold_ms = optional(number)
    follow_redirects         = optional(bool)
    headers                  = optional(map(string))
    body                     = optional(string)
    regions                  = optional(list(string))
    paused                   = optional(bool)
  }))

  validation {
    condition     = length(var.pages) > 0
    error_message = "At least one page must be specified."
  }

  validation {
    condition = alltrue([
      for k, v in var.pages : can(regex("^/", v.path))
    ])
    error_message = "All page paths must start with '/'."
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
  description = "Custom name format string. Use %s for the page key. Overrides name_prefix if set."
  type        = string
  default     = ""
}

variable "frequency" {
  description = "Default check frequency in seconds"
  type        = number
  default     = 60

  validation {
    condition = contains([
      10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400
    ], var.frequency)
    error_message = "Frequency must be one of: 10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400 seconds."
  }
}

variable "regions" {
  description = "Default regions for monitors (can be overridden per page)"
  type        = list(string)
  default     = ["virginia", "london", "frankfurt"]

  validation {
    condition = alltrue([
      for r in var.regions : contains([
        "paris", "frankfurt", "amsterdam", "london",
        "singapore", "sydney", "tokyo", "seoul", "mumbai", "bangalore",
        "virginia", "california", "sanfrancisco", "nyc", "toronto", "saopaulo",
        "bahrain", "capetown"
      ], r)
    ])
    error_message = "Invalid region specified."
  }
}

variable "default_method" {
  description = "Default HTTP method for requests"
  type        = string
  default     = "GET"

  validation {
    condition     = contains(["GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"], var.default_method)
    error_message = "Method must be one of: GET, POST, PUT, PATCH, DELETE, HEAD."
  }
}

variable "default_expected_status" {
  description = "Default expected HTTP status code or pattern (e.g., '200', '2xx')"
  type        = string
  default     = "2xx"
}

variable "follow_redirects" {
  description = "Follow HTTP redirects by default"
  type        = bool
  default     = true
}

variable "performance_threshold_ms" {
  description = "Response time threshold in milliseconds (null = no threshold)"
  type        = number
  default     = null

  validation {
    condition     = var.performance_threshold_ms == null || var.performance_threshold_ms > 0
    error_message = "Performance threshold must be greater than 0 if specified."
  }
}

variable "default_headers" {
  description = "Default HTTP headers to include in all requests"
  type        = map(string)
  default     = null
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
