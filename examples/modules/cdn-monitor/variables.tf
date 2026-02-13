# CDN Monitor Module - Variables

variable "cdn_domain" {
  description = "CDN domain to monitor (e.g., 'cdn.example.com')"
  type        = string

  validation {
    condition     = can(regex("^[a-zA-Z0-9][a-zA-Z0-9.-]+[a-zA-Z0-9]$", var.cdn_domain))
    error_message = "CDN domain must be a valid domain name without protocol."
  }
}

variable "assets" {
  description = "Map of asset paths to monitor (e.g., { logo = '/images/logo.png', css = '/styles/main.css' })"
  type        = map(string)

  validation {
    condition     = length(var.assets) > 0
    error_message = "At least one asset must be specified."
  }

  validation {
    condition = alltrue([
      for k, v in var.assets : can(regex("^/", v))
    ])
    error_message = "All asset paths must start with '/'."
  }
}

variable "protocol" {
  description = "Protocol to use (http or https)"
  type        = string
  default     = "https"

  validation {
    condition     = contains(["http", "https"], var.protocol)
    error_message = "Protocol must be 'http' or 'https'."
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
  description = "Custom name format string. Use %s for the asset key. Overrides name_prefix if set."
  type        = string
  default     = ""
}

variable "regions" {
  description = "List of regions to check from (validates global CDN coverage)"
  type        = list(string)
  default     = ["virginia", "london", "singapore", "sydney", "frankfurt"]

  validation {
    condition = alltrue([
      for r in var.regions : contains([
        "london", "frankfurt", "singapore", "sydney", "tokyo", "virginia", "saopaulo", "bahrain"
      ], r)
    ])
    error_message = "Invalid region specified. Valid regions: london, frankfurt, singapore, sydney, tokyo, virginia, saopaulo, bahrain."
  }

  validation {
    condition     = length(var.regions) >= 3
    error_message = "CDN monitoring should use at least 3 regions for meaningful global coverage."
  }
}

variable "check_frequency" {
  description = "Check frequency in seconds (recommended: 300 for CDN assets)"
  type        = number
  default     = 300

  validation {
    condition = contains([
      10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400
    ], var.check_frequency)
    error_message = "Frequency must be one of: 10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400 seconds."
  }
}

variable "follow_redirects" {
  description = "Follow HTTP redirects"
  type        = bool
  default     = false
}

variable "asset_keywords" {
  description = "Map of asset keys to required keywords in response body (for content validation)"
  type        = map(string)
  default     = {}
}

variable "monitor_root_domain" {
  description = "Create an additional monitor for the CDN root domain"
  type        = bool
  default     = false
}

variable "root_expected_status" {
  description = "Expected status code for root domain monitor"
  type        = string
  default     = "2xx"
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
