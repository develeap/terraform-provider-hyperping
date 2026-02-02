# Look up a specific monitor by its ID
data "hyperping_monitor" "example" {
  id = "your-monitor-uuid-here"
}

# Use the monitor data in other resources
output "monitor_name" {
  value = data.hyperping_monitor.example.name
}

output "monitor_url" {
  value = data.hyperping_monitor.example.url
}

output "monitor_protocol" {
  value = data.hyperping_monitor.example.protocol
}

output "monitor_status" {
  value = data.hyperping_monitor.example.paused ? "PAUSED" : "ACTIVE"
}

# Example: Create an incident for a specific status page
resource "hyperping_incident" "outage" {
  title        = "Issue with ${data.hyperping_monitor.example.name}"
  text         = "Investigating issues with monitor at ${data.hyperping_monitor.example.url}"
  type         = "incident"
  status_pages = ["your-status-page-uuid"]

  affected_components = ["your-component-uuid"]
}
