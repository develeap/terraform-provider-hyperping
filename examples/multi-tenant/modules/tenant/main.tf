# Tenant Module - Manages monitors for a single tenant
#
# Usage:
#   module "acme" {
#     source    = "../modules/tenant"
#     tenant_id = "acme"
#     ...
#   }

terraform {
  required_providers {
    hyperping = {
      source  = "develeap/hyperping"
      version = "~> 1.0"
    }
  }
}

locals {
  # Filter to enabled monitors only
  enabled_monitors = [for m in var.monitors : m if m.enabled]

  # Build component UUID lookup (name -> uuid)
  component_uuids = {
    for c in var.status_page.components : c.name => c.uuid
    if c.uuid != null
  }

  # Monitor name prefix
  name_prefix = upper(var.tenant_id)
}

# Create monitors for this tenant
resource "hyperping_monitor" "tenant" {
  for_each = {
    for m in local.enabled_monitors : m.name => m
  }

  name                 = "[${local.name_prefix}]-${each.value.category != null ? "${each.value.category}-" : ""}${each.value.name}"
  url                  = each.value.url
  protocol             = "http"
  http_method          = each.value.method
  check_frequency      = each.value.frequency
  expected_status_code = each.value.expected_status
  follow_redirects     = true
  regions              = each.value.regions

  dynamic "request_headers" {
    for_each = each.value.headers != null ? each.value.headers : []
    content {
      name  = request_headers.value.name
      value = request_headers.value.value
    }
  }

  request_body = each.value.body

  lifecycle {
    # Prevent accidental deletion
    prevent_destroy = false
  }
}

# Output mapping for use in incidents
locals {
  monitor_to_component = {
    for name, m in local.enabled_monitors : name => m.component
    if m.component != null
  }
}
