# Fetch all monitors
data "hyperping_monitors" "all" {
}

# Output all monitor IDs
output "monitor_ids" {
  value = data.hyperping_monitors.all.monitors[*].id
}

# Output all monitor names
output "monitor_names" {
  value = data.hyperping_monitors.all.monitors[*].name
}

# Filter paused monitors (example)
output "paused_monitors" {
  value = [
    for m in data.hyperping_monitors.all.monitors : m.name
    if m.paused
  ]
}

# Count monitors by status
output "monitor_summary" {
  value = {
    total  = length(data.hyperping_monitors.all.monitors)
    paused = length([for m in data.hyperping_monitors.all.monitors : m if m.paused])
    active = length([for m in data.hyperping_monitors.all.monitors : m if !m.paused])
  }
}

# Output monitors grouped by protocol
output "monitors_by_protocol" {
  value = {
    for protocol in distinct([for m in data.hyperping_monitors.all.monitors : m.protocol]) :
    protocol => [for m in data.hyperping_monitors.all.monitors : m.name if m.protocol == protocol]
  }
}
