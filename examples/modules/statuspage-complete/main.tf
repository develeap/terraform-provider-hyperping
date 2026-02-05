# Status Page Complete Module - Main
#
# Creates a fully configured status page with monitors for all services.
# Handles service grouping, branding, and subscription configuration.
#
# Usage:
#   module "status_page" {
#     source = "path/to/modules/statuspage-complete"
#
#     name      = "Acme Corp Status"
#     subdomain = "status"
#
#     services = {
#       api = {
#         url         = "https://api.example.com/health"
#         description = "Core API"
#       }
#       web = {
#         url         = "https://www.example.com"
#         description = "Web Application"
#       }
#     }
#   }

locals {
  # Build region list for monitors
  monitor_regions = var.regions
}

# Create monitors for each service
resource "hyperping_monitor" "service" {
  for_each = var.services

  name                 = each.key
  url                  = each.value.url
  protocol             = "http"
  http_method          = each.value.method
  check_frequency      = each.value.frequency
  expected_status_code = each.value.expected_status_code
  regions              = local.monitor_regions
  follow_redirects     = true

  dynamic "request_headers" {
    for_each = each.value.headers
    content {
      key   = request_headers.key
      value = request_headers.value
    }
  }

  lifecycle {
    create_before_destroy = true
  }
}

# Create the status page
resource "hyperping_statuspage" "main" {
  name      = var.name
  subdomain = var.subdomain
  hostname  = var.hostname

  # Branding
  theme        = var.theme
  accent_color = var.accent_color

  # Languages
  languages = var.languages

  # Features
  hide_powered_by      = var.hide_powered_by
  enable_subscriptions = var.enable_subscriptions

  # Associate all service monitors
  monitor_ids = [for m in hyperping_monitor.service : m.id]

  lifecycle {
    create_before_destroy = true
  }
}
