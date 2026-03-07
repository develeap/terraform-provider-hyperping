# Add an update to an existing incident
resource "hyperping_incident_update" "investigating" {
  incident_id = hyperping_incident.outage.id
  type        = "investigating"
  text        = "We are investigating the issue affecting our API services."
}

# Follow-up update
resource "hyperping_incident_update" "identified" {
  incident_id = hyperping_incident.outage.id
  type        = "identified"
  text        = "The root cause has been identified. A fix is being deployed."

  depends_on = [hyperping_incident_update.investigating]
}
