# Single status page: declarative import block (Terraform 1.5+)
import {
  to = hyperping_statuspage.main
  id = "sp_status123abc"
}

resource "hyperping_statuspage" "main" {
  name             = "Production Status"
  hosted_subdomain = "status"

  settings = {
    name      = "Production Status"
    languages = ["en"]
  }
}

# Fleet import: bring multiple status pages under management at once
# using for_each (Terraform 1.7+).
locals {
  statuspage_ids = {
    production = "sp_prod111aaa"
    staging    = "sp_stage222bbb"
    internal   = "sp_int333ccc"
  }
}

import {
  for_each = local.statuspage_ids
  to       = hyperping_statuspage.fleet[each.key]
  id       = each.value
}

resource "hyperping_statuspage" "fleet" {
  for_each = local.statuspage_ids

  # Placeholder values: run `terraform plan` after import to see the
  # real values, then update this block to match before applying.
  name             = each.key
  hosted_subdomain = each.key

  settings = {
    name      = each.key
    languages = ["en"]
  }
}
