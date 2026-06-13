# Single monitor: declarative import block (Terraform 1.5+)
import {
  to = hyperping_monitor.api
  id = "mon_abc123def456"
}

resource "hyperping_monitor" "api" {
  name                 = "Production API"
  url                  = "https://api.example.com/health"
  protocol             = "http"
  http_method          = "GET"
  check_frequency      = 300
  expected_status_code = "200"

  regions = ["london", "virginia", "singapore"]
}

# Fleet import: bring an entire monitor inventory under management
# in one pass using for_each (Terraform 1.7+).
#
# Populate monitor_ids from your Hyperping dashboard or the data source
# hyperping_monitors (see examples/data-sources/hyperping_monitors/).
locals {
  monitor_ids = {
    api_health  = "mon_aaabbbccc111"
    web_home    = "mon_dddeeefff222"
    db_primary  = "mon_ggghhh333444"
    cdn_assets  = "mon_iiijjj555666"
  }
}

import {
  for_each = local.monitor_ids
  to       = hyperping_monitor.fleet[each.key]
  id       = each.value
}

resource "hyperping_monitor" "fleet" {
  for_each = local.monitor_ids

  # Placeholder values: run `terraform plan` after import to see the
  # real values, then update this block to match before applying.
  name     = each.key
  url      = "https://placeholder.example.com"
  protocol = "http"
}
