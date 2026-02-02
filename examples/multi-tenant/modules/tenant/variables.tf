# Tenant Module Variables

variable "tenant_id" {
  description = "Unique tenant identifier (lowercase, alphanumeric, hyphens)"
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9-]+$", var.tenant_id))
    error_message = "Tenant ID must be lowercase alphanumeric with hyphens."
  }
}

variable "tenant_name" {
  description = "Human-readable tenant name"
  type        = string
}

variable "active" {
  description = "Whether this tenant is active"
  type        = bool
  default     = true
}

# Status Page (UUIDs from Hyperping - manual until API available)
variable "status_page" {
  description = "Status page configuration"
  type = object({
    uuid       = optional(string)  # sp_xxx from Hyperping dashboard
    subdomain  = string
    components = optional(list(object({
      name         = string
      uuid         = optional(string)  # comp_xxx from Hyperping dashboard
      group        = optional(string)
      display_order = optional(number, 0)
    })), [])
  })
}

# Monitors
variable "monitors" {
  description = "List of monitors for this tenant"
  type = list(object({
    name            = string
    url             = string
    method          = optional(string, "GET")
    frequency       = optional(number, 60)
    regions         = optional(list(string), ["london", "virginia", "singapore"])
    expected_status = optional(string, "2xx")
    category        = optional(string)
    headers         = optional(list(object({
      name  = string
      value = string
    })), [])
    body            = optional(string)
    component       = optional(string)  # Component name to link
    enabled         = optional(bool, true)
  }))
  default = []
}

# Shared monitors to reference (created in shared/ workspace)
variable "shared_monitor_ids" {
  description = "Map of shared monitor name to UUID"
  type        = map(string)
  default     = {}
}

# Tags for organization
variable "tags" {
  description = "Tags for this tenant"
  type        = list(string)
  default     = []
}
