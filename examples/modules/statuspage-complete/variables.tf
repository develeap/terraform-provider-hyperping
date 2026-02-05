# Status Page Complete Module - Variables

variable "name" {
  description = "Status page name"
  type        = string

  validation {
    condition     = length(var.name) >= 1 && length(var.name) <= 255
    error_message = "Name must be between 1 and 255 characters."
  }
}

variable "subdomain" {
  description = "Subdomain for hosted status page (e.g., 'status' for status.hyperping.app)"
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9-]+$", var.subdomain))
    error_message = "Subdomain must be lowercase alphanumeric with hyphens only."
  }
}

variable "hostname" {
  description = "Custom domain for status page (e.g., 'status.example.com')"
  type        = string
  default     = null
}

variable "services" {
  description = "Map of services to monitor and display on status page"
  type = map(object({
    url                  = string
    description          = optional(string, "")
    method               = optional(string, "GET")
    frequency            = optional(number, 60)
    expected_status_code = optional(string, "200")
    headers              = optional(map(string), {})
  }))

  validation {
    condition     = length(var.services) > 0
    error_message = "At least one service must be specified."
  }
}

variable "theme" {
  description = "Status page theme"
  type        = string
  default     = "system"

  validation {
    condition     = contains(["light", "dark", "system"], var.theme)
    error_message = "Theme must be 'light', 'dark', or 'system'."
  }
}

variable "accent_color" {
  description = "Accent color (hex format)"
  type        = string
  default     = "#36b27e"

  validation {
    condition     = can(regex("^#[0-9a-fA-F]{6}$", var.accent_color))
    error_message = "Accent color must be in hex format (e.g., '#36b27e')."
  }
}

variable "languages" {
  description = "Supported languages"
  type        = list(string)
  default     = ["en"]
}

variable "regions" {
  description = "Monitoring regions for all services"
  type        = list(string)
  default     = ["virginia", "london", "frankfurt"]

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

variable "hide_powered_by" {
  description = "Hide 'Powered by Hyperping' branding"
  type        = bool
  default     = false
}

variable "enable_subscriptions" {
  description = "Allow visitors to subscribe to status updates"
  type        = bool
  default     = true
}
