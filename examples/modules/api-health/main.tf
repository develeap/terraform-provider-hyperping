# API Health Module - Main
#
# Creates HTTP/HTTPS monitors for API health checking.
#
# Usage:
#   module "api_monitors" {
#     source = "path/to/modules/api-health"
#
#     endpoints = {
#       "api" = { url = "https://api.example.com/health" }
#       "web" = { url = "https://www.example.com", frequency = 120 }
#     }
#
#     name_prefix     = "prod"
#     default_regions = ["virginia", "london"]
#   }

locals {
  # Build the full name with optional prefix
  name_format = var.name_prefix != "" ? "[${upper(var.name_prefix)}] %s" : "%s"
}

resource "hyperping_monitor" "endpoint" {
  for_each = var.endpoints

  name     = format(local.name_format, each.key)
  url      = each.value.url
  protocol = "http"

  http_method          = each.value.method
  check_frequency      = coalesce(each.value.frequency, var.default_frequency)
  expected_status_code = each.value.expected_status_code
  follow_redirects     = each.value.follow_redirects
  regions              = coalesce(each.value.regions, var.default_regions)
  paused               = each.value.paused
  alerts_wait          = var.alerts_wait

  # Optional: required keyword for content validation
  required_keyword = each.value.required_keyword

  # Optional: escalation policy
  escalation_policy_uuid = var.escalation_policy_uuid

  # Dynamic request headers
  dynamic "request_headers" {
    for_each = each.value.headers != null ? each.value.headers : {}
    content {
      name  = request_headers.key
      value = request_headers.value
    }
  }

  # Optional request body (for POST/PUT methods)
  request_body = each.value.body

  lifecycle {
    create_before_destroy = true
  }
}
