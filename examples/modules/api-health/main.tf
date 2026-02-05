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
  # Build the full name with format string
  name_format = var.name_format != "" ? var.name_format : (var.name_prefix != "" ? "[${upper(var.name_prefix)}] %s" : "%s")
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
  paused               = coalesce(each.value.paused, var.paused)
  alerts_wait          = var.alerts_wait

  # Optional: required keyword for content validation
  required_keyword = each.value.required_keyword

  # Optional: escalation policy
  escalation_policy = var.escalation_policy

  # Request headers as list of objects (convert from map)
  request_headers = each.value.headers != null ? [
    for k, v in each.value.headers : {
      name  = k
      value = v
    }
  ] : null

  # Optional request body (for POST/PUT methods)
  request_body = each.value.body

  lifecycle {
    create_before_destroy = true
  }
}
