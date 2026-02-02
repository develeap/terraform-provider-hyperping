# List all healthchecks
data "hyperping_healthchecks" "all" {}

# Filter down healthchecks (currently in failure state)
locals {
  failing_healthchecks = [
    for hc in data.hyperping_healthchecks.all.healthchecks :
    hc if hc.is_down == true
  ]
}

# Output summary
output "healthcheck_summary" {
  value = {
    total_count   = length(data.hyperping_healthchecks.all.healthchecks)
    failing_count = length(local.failing_healthchecks)
    paused_count  = length([for hc in data.hyperping_healthchecks.all.healthchecks : hc if hc.is_paused])
  }
}

# Create a map of healthcheck IDs to ping URLs
output "healthcheck_ping_urls" {
  value = {
    for hc in data.hyperping_healthchecks.all.healthchecks :
    hc.name => hc.ping_url
  }
  sensitive = true
}
