# Basic incident
resource "hyperping_incident" "example" {
  title        = "API Performance Degradation"
  text         = "We are investigating reports of slow API response times."
  type         = "incident"
  status_pages = [hyperping_status_page.main.id]
}

# Full incident with all options
resource "hyperping_incident" "outage" {
  title               = "Database Connectivity Issues"
  text                = "Our database cluster is experiencing connectivity problems. We are working to restore service."
  type                = "outage"
  status_pages        = [hyperping_status_page.main.id]
  affected_components = [hyperping_component.database.id, hyperping_component.api.id]
}

# Output incident details
output "incident_id" {
  description = "The ID of the created incident"
  value       = hyperping_incident.example.id
}

output "incident_date" {
  description = "When the incident was created"
  value       = hyperping_incident.example.date
}
