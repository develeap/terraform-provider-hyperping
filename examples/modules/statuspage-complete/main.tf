# Status Page Complete Module - Main
#
# Creates a fully configured status page with monitors for all services.
# Handles service grouping, branding, and subscription configuration.
#
# Usage:
#   module "status_page" {
#     source = "path/to/modules/statuspage-complete"
#
#     name             = "Acme Corp Status"
#     hosted_subdomain = "acme-status"
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

  # Build description map for status page settings
  description_map = var.description != null ? var.description : {
    for lang in var.languages : lang => var.name
  }

  # Build services list for sections
  services_list = [
    for k, v in hyperping_monitor.service : {
      uuid = v.id
      name = {
        for lang in var.languages : lang => k
      }
      show_uptime         = true
      show_response_times = true
    }
  ]
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

  # Request headers as list of objects (convert from map)
  request_headers = each.value.headers != null ? [
    for k, v in each.value.headers : {
      name  = k
      value = v
    }
  ] : null

  lifecycle {
    create_before_destroy = true
  }
}

# Create the status page
resource "hyperping_statuspage" "main" {
  name             = var.name
  hosted_subdomain = var.hosted_subdomain
  hostname         = var.hostname

  settings = {
    name             = var.name
    languages        = var.languages
    default_language = var.languages[0]
    theme            = var.theme
    accent_color     = var.accent_color
    description      = local.description_map
    hide_powered_by  = var.hide_powered_by

    subscribe = {
      enabled = var.enable_subscriptions
      email   = var.enable_subscriptions
    }
  }

  # Create a single section with all services
  sections = [
    {
      name = {
        for lang in var.languages : lang => "Services"
      }
      is_split = true
      services = local.services_list
    }
  ]

  lifecycle {
    create_before_destroy = true
  }
}
