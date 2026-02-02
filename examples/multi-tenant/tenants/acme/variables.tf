# Variables passed to tenant module
# Values come from terraform.tfvars

variable "tenant_id" {
  type = string
}

variable "tenant_name" {
  type = string
}

variable "active" {
  type    = bool
  default = true
}

variable "status_page" {
  type = object({
    uuid       = optional(string)
    subdomain  = string
    components = optional(list(object({
      name          = string
      uuid          = optional(string)
      group         = optional(string)
      display_order = optional(number, 0)
    })), [])
  })
}

variable "monitors" {
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
    component       = optional(string)
    enabled         = optional(bool, true)
  }))
  default = []
}

variable "shared_monitor_ids" {
  type    = map(string)
  default = {}
}

variable "tags" {
  type    = list(string)
  default = []
}
