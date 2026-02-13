# Database Monitor Module - Variables

variable "databases" {
  description = "Map of database configurations to monitor"
  type = map(object({
    host      = string
    port      = number
    type      = string
    frequency = optional(number)
    regions   = optional(list(string))
    paused    = optional(bool)
  }))

  validation {
    condition     = length(var.databases) > 0
    error_message = "At least one database must be specified."
  }

  validation {
    condition = alltrue([
      for k, v in var.databases : contains([
        "postgresql", "postgres", "mysql", "redis", "mongodb", "mongo",
        "mariadb", "cassandra", "elasticsearch", "memcached", "rabbitmq"
      ], lower(v.type))
    ])
    error_message = "Database type must be one of: postgresql, postgres, mysql, redis, mongodb, mongo, mariadb, cassandra, elasticsearch, memcached, rabbitmq."
  }

  validation {
    condition = alltrue([
      for k, v in var.databases : v.port >= 1 && v.port <= 65535
    ])
    error_message = "Port must be between 1 and 65535."
  }
}

variable "name_prefix" {
  description = "Prefix for monitor names (e.g., 'PROD', 'STAGING')"
  type        = string
  default     = ""

  validation {
    condition     = can(regex("^[a-zA-Z0-9-_]*$", var.name_prefix))
    error_message = "Name prefix must contain only alphanumeric characters, hyphens, and underscores."
  }
}

variable "name_format" {
  description = "Custom name format string. Use %s for the database key. Overrides name_prefix if set."
  type        = string
  default     = ""
}

variable "default_regions" {
  description = "Default regions for monitors (can be overridden per database)"
  type        = list(string)
  default     = ["virginia", "london"]

  validation {
    condition = alltrue([
      for r in var.default_regions : contains([
        "london", "frankfurt", "singapore", "sydney", "tokyo",
        "virginia", "saopaulo", "bahrain"
      ], r)
    ])
    error_message = "Invalid region specified. Valid regions: london, frankfurt, singapore, sydney, tokyo, virginia, saopaulo, bahrain."
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

variable "include_type_in_name" {
  description = "Include database type in monitor name (e.g., 'PostgreSQL - mydb')"
  type        = bool
  default     = true
}
