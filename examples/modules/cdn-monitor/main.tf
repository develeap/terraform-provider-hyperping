# CDN Monitor Module - Main
#
# Creates HTTP/HTTPS monitors for CDN edge location validation and cache performance.
# Monitors check static asset availability globally and validate cache headers.
#
# Usage:
#   module "cdn_monitors" {
#     source = "path/to/modules/cdn-monitor"
#
#     cdn_domain = "cdn.example.com"
#     assets = {
#       logo = "/images/logo.png"
#       css  = "/styles/main.css"
#       js   = "/scripts/app.js"
#     }
#
#     regions = ["virginia", "london", "singapore", "sydney"]
#   }

locals {
  # Build full URLs for each asset
  asset_urls = {
    for k, v in var.assets : k => "${var.protocol}://${var.cdn_domain}${v}"
  }

  # Build the full name with format string
  name_format = var.name_format != "" ? var.name_format : (var.name_prefix != "" ? "[${upper(var.name_prefix)}] CDN: %s" : "CDN: %s")
}

resource "hyperping_monitor" "cdn_asset" {
  for_each = local.asset_urls

  name     = format(local.name_format, each.key)
  url      = each.value
  protocol = "http"

  http_method          = "GET"
  check_frequency      = var.check_frequency
  expected_status_code = "2xx"
  follow_redirects     = var.follow_redirects
  regions              = var.regions
  paused               = var.paused
  alerts_wait          = var.alerts_wait

  # Optional: validate specific content in response
  # Note: Can include cache-related keywords in response body
  required_keyword = lookup(var.asset_keywords, each.key, null)

  # Optional: escalation policy
  escalation_policy = var.escalation_policy

  lifecycle {
    create_before_destroy = true
  }
}

# Optional: Create a single monitor for the CDN root domain
resource "hyperping_monitor" "cdn_root" {
  count = var.monitor_root_domain ? 1 : 0

  name     = format(local.name_format, "root")
  url      = "${var.protocol}://${var.cdn_domain}"
  protocol = "http"

  http_method          = "GET"
  check_frequency      = var.check_frequency
  expected_status_code = var.root_expected_status
  follow_redirects     = var.follow_redirects
  regions              = var.regions
  paused               = var.paused
  alerts_wait          = var.alerts_wait

  escalation_policy = var.escalation_policy

  lifecycle {
    create_before_destroy = true
  }
}
