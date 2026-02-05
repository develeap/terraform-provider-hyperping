# SSL Monitor Module - Main
#
# Creates HTTPS monitors for SSL certificate health checking.
# Monitors verify that HTTPS connections can be established,
# which will fail if certificates are expired or invalid.
#
# Usage:
#   module "ssl_monitors" {
#     source = "path/to/modules/ssl-monitor"
#
#     domains = [
#       "api.example.com",
#       "www.example.com",
#       "admin.example.com"
#     ]
#
#     check_frequency = 3600  # hourly
#   }

locals {
  # Create a map from the list for for_each
  domain_map = { for d in var.domains : d => d }
}

resource "hyperping_monitor" "ssl" {
  for_each = local.domain_map

  name     = "[${var.name_prefix}] ${each.key}"
  url      = "https://${each.key}"
  protocol = "http"
  port     = var.port

  http_method          = "GET"
  check_frequency      = var.check_frequency
  expected_status_code = "2xx"
  follow_redirects     = true
  regions              = var.regions
  paused               = var.paused
  alerts_wait          = var.alerts_wait

  escalation_policy = var.escalation_policy

  lifecycle {
    create_before_destroy = true
  }
}
