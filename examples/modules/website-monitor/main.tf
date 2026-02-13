# Website Monitor Module - Main
#
# Creates comprehensive website monitoring including homepage,
# critical pages, content validation, and performance thresholds.
#
# Usage:
#   module "website" {
#     source = "path/to/modules/website-monitor"
#
#     domain = "example.com"
#     pages = {
#       homepage = {
#         path          = "/"
#         expected_text = "Welcome"
#       }
#       login = {
#         path          = "/login"
#         expected_text = "Sign In"
#       }
#     }
#   }

locals {
  # Build monitor names with consistent format
  name_format = var.name_format != "" ? var.name_format : (
    var.name_prefix != "" ? "[${upper(var.name_prefix)}] ${var.domain} - %s" : "${var.domain} - %s"
  )

  # Build full URLs for each page
  pages_with_urls = {
    for key, page in var.pages : key => merge(page, {
      full_url = "${var.protocol}://${var.domain}${page.path}"
    })
  }

  # Determine which pages should use performance checks
  pages_with_performance = {
    for key, page in local.pages_with_urls : key => page
    if var.performance_threshold_ms != null || page.performance_threshold_ms != null
  }
}

resource "hyperping_monitor" "page" {
  for_each = local.pages_with_urls

  name     = format(local.name_format, each.key)
  url      = each.value.full_url
  protocol = "http"

  http_method          = coalesce(each.value.method, var.default_method)
  check_frequency      = coalesce(each.value.frequency, var.frequency)
  expected_status_code = coalesce(each.value.expected_status, var.default_expected_status)
  follow_redirects     = coalesce(each.value.follow_redirects, var.follow_redirects)
  regions              = coalesce(each.value.regions, var.regions)
  paused               = coalesce(each.value.paused, var.paused)
  alerts_wait          = var.alerts_wait

  # Content validation - check for expected text
  required_keyword = each.value.expected_text

  # Performance threshold (if specified)
  response_time_threshold = coalesce(
    each.value.performance_threshold_ms,
    var.performance_threshold_ms
  )

  # Optional: custom headers (e.g., for authentication, user-agent)
  request_headers = each.value.headers != null ? [
    for k, v in each.value.headers : {
      name  = k
      value = v
    }
    ] : (var.default_headers != null ? [
      for k, v in var.default_headers : {
        name  = k
        value = v
      }
  ] : null)

  # Optional: request body for POST/PUT methods
  request_body = each.value.body

  # Optional: escalation policy
  escalation_policy = var.escalation_policy

  lifecycle {
    create_before_destroy = true
  }
}
