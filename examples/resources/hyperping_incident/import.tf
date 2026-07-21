# Single incident: declarative import block (Terraform 1.5+)
import {
  to = hyperping_incident.api_degradation
  id = "inc_incident123abc"
}

resource "hyperping_incident" "api_degradation" {
  title        = "API Performance Degradation"
  text         = "We are investigating reports of slow API response times."
  type         = "incident"
  status_pages = ["sp_prod111aaa"]
}

# Fleet import: bring a batch of open incidents under management
# using for_each (Terraform 1.7+).
#
# This is useful when migrating from manual incident management to IaC
# or when seeding a new Terraform workspace from an existing account.
locals {
  incident_ids = {
    api_degradation   = "inc_aaa111bbb222"
    database_latency  = "inc_ccc333ddd444"
    cdn_outage        = "inc_eee555fff666"
  }
}

import {
  for_each = local.incident_ids
  to       = hyperping_incident.fleet[each.key]
  id       = each.value
}

resource "hyperping_incident" "fleet" {
  for_each = local.incident_ids

  # Placeholder values: run `terraform plan` after import to see the
  # real values, then update this block to match before applying.
  title        = each.key
  text         = "Imported incident"
  type         = "incident"
  status_pages = []
}
