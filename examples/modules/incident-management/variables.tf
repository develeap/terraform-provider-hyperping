# Incident Management Module - Variables

variable "statuspage_id" {
  description = "UUID of the status page to associate incidents with"
  type        = string
  default     = null
}

variable "incident_templates" {
  description = "Map of incident template configurations"
  type = map(object({
    title               = string
    text                = string
    severity            = optional(string, "major")
    type                = optional(string)
    affected_components = optional(list(string), [])
    status_pages        = optional(list(string))
  }))
  default = {}

  validation {
    condition = alltrue([
      for k, v in var.incident_templates : contains(["minor", "major", "critical"], v.severity)
    ])
    error_message = "Incident severity must be one of: minor, major, critical."
  }

  validation {
    condition = alltrue([
      for k, v in var.incident_templates : v.type == null || contains(["incident", "outage"], v.type)
    ])
    error_message = "Incident type must be one of: incident, outage (or null for auto-detection)."
  }

  validation {
    condition = alltrue([
      for k, v in var.incident_templates : length(v.title) > 0 && length(v.title) <= 255
    ])
    error_message = "Incident title must be between 1 and 255 characters."
  }
}

variable "maintenance_windows" {
  description = "Map of scheduled maintenance window configurations"
  type = map(object({
    title                = string
    text                 = string
    start_date           = string
    end_date             = string
    monitors             = optional(list(string), [])
    status_pages         = optional(list(string))
    notification_option  = optional(string, "scheduled")
    notification_minutes = optional(number, 60)
  }))
  default = {}

  validation {
    condition = alltrue([
      for k, v in var.maintenance_windows : contains(["immediate", "scheduled", "none"], v.notification_option)
    ])
    error_message = "Notification option must be one of: immediate, scheduled, none."
  }

  validation {
    condition = alltrue([
      for k, v in var.maintenance_windows : v.notification_option != "scheduled" || v.notification_minutes != null
    ])
    error_message = "notification_minutes is required when notification_option is 'scheduled'."
  }

  validation {
    condition = alltrue([
      for k, v in var.maintenance_windows : can(formatdate("RFC3339", v.start_date))
    ])
    error_message = "start_date must be a valid ISO 8601 datetime string."
  }

  validation {
    condition = alltrue([
      for k, v in var.maintenance_windows : can(formatdate("RFC3339", v.end_date))
    ])
    error_message = "end_date must be a valid ISO 8601 datetime string."
  }
}

variable "outage_definitions" {
  description = "Map of manual outage configurations for monitors"
  type = map(object({
    monitor_uuid = string
    start_date   = string
    end_date     = optional(string)
    status_code  = number
    description  = string
  }))
  default = {}

  validation {
    condition = alltrue([
      for k, v in var.outage_definitions : v.status_code >= 100 && v.status_code <= 599
    ])
    error_message = "Status code must be a valid HTTP status code (100-599)."
  }

  validation {
    condition = alltrue([
      for k, v in var.outage_definitions : can(formatdate("RFC3339", v.start_date))
    ])
    error_message = "start_date must be a valid ISO 8601 datetime string."
  }

  validation {
    condition = alltrue([
      for k, v in var.outage_definitions : v.end_date == null || can(formatdate("RFC3339", v.end_date))
    ])
    error_message = "end_date must be a valid ISO 8601 datetime string or null."
  }
}

variable "create_incidents" {
  description = "Enable creation of incident resources"
  type        = bool
  default     = true
}

variable "create_maintenance" {
  description = "Enable creation of maintenance window resources"
  type        = bool
  default     = true
}

variable "create_outages" {
  description = "Enable creation of manual outage resources"
  type        = bool
  default     = true
}

variable "default_incident_type" {
  description = "Default incident type when not specified"
  type        = string
  default     = "incident"

  validation {
    condition     = contains(["incident", "outage"], var.default_incident_type)
    error_message = "Default incident type must be either 'incident' or 'outage'."
  }
}
