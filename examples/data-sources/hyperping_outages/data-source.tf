# List all outages
data "hyperping_outages" "all" {}

# Filter for unresolved outages
locals {
  active_outages = [
    for outage in data.hyperping_outages.all.outages :
    outage if !outage.is_resolved
  ]

  # Group outages by monitor
  outages_by_monitor = {
    for outage in data.hyperping_outages.all.outages :
    outage.monitor.name => outage...
  }
}

# Output active incidents for monitoring
output "active_incidents" {
  value = {
    count = length(local.active_outages)
    details = [
      for outage in local.active_outages : {
        id          = outage.id
        monitor     = outage.monitor.name
        started     = outage.start_date
        description = outage.description
      }
    ]
  }
}

# Calculate total downtime in the last period
output "downtime_statistics" {
  value = {
    total_outages     = length(data.hyperping_outages.all.outages)
    total_downtime_ms = sum([for o in data.hyperping_outages.all.outages : o.duration_ms])
    manual_outages    = length([for o in data.hyperping_outages.all.outages : o if o.outage_type == "manual"])
  }
}
