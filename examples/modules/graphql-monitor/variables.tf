# GraphQL Monitor Module - Variables

variable "endpoint" {
  description = "GraphQL endpoint URL"
  type        = string

  validation {
    condition     = can(regex("^https?://", var.endpoint))
    error_message = "Endpoint URL must start with http:// or https://."
  }
}

variable "queries" {
  description = "Map of GraphQL queries to monitor"
  type = map(object({
    query                = string
    variables            = optional(map(any))
    expected_response    = string
    frequency            = optional(number)
    expected_status_code = optional(string)
    follow_redirects     = optional(bool)
    regions              = optional(list(string))
    paused               = optional(bool)
  }))

  validation {
    condition     = length(var.queries) > 0
    error_message = "At least one GraphQL query must be specified."
  }

  validation {
    condition = alltrue([
      for k, v in var.queries : can(regex("^\\s*\\{", v.query)) || can(regex("^\\s*query", v.query)) || can(regex("^\\s*mutation", v.query))
    ])
    error_message = "All queries must be valid GraphQL query or mutation syntax."
  }
}

variable "custom_headers" {
  description = "Additional custom headers to include in GraphQL requests. Note: Reserved headers (Authorization, Cookie, Host, etc.) are not allowed by the provider for security reasons. Use non-reserved headers like X-API-Key for authentication if needed."
  type        = map(string)
  default     = null
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
  description = "Custom name format string. Use %s for the query key. Overrides name_prefix if set."
  type        = string
  default     = ""
}

variable "default_regions" {
  description = "Default regions for monitors (can be overridden per query)"
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

variable "default_frequency" {
  description = "Default check frequency in seconds"
  type        = number
  default     = 120

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

# Introspection check variables
variable "enable_introspection_check" {
  description = "Enable GraphQL introspection check monitor"
  type        = bool
  default     = false
}

variable "introspection_query" {
  description = "GraphQL introspection query to use (default: standard __schema query)"
  type        = string
  default     = "{ __schema { queryType { name } } }"
}

variable "introspection_expected_response" {
  description = "Expected response keyword for introspection check"
  type        = string
  default     = "queryType"
}

variable "introspection_frequency" {
  description = "Check frequency for introspection monitor (seconds)"
  type        = number
  default     = 3600

  validation {
    condition = contains([
      10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400
    ], var.introspection_frequency)
    error_message = "Frequency must be one of: 10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400 seconds."
  }
}

variable "introspection_expected_status" {
  description = "Expected HTTP status code for introspection check"
  type        = string
  default     = "200"
}
