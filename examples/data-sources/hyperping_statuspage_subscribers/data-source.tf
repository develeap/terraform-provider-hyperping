# List all subscribers for a status page
data "hyperping_statuspage_subscribers" "all" {
  statuspage_uuid = hyperping_statuspage.production.id
}

# Output subscriber count
output "total_subscribers" {
  value = data.hyperping_statuspage_subscribers.all.total
}

# Output all subscriber emails
output "subscriber_emails" {
  value = [
    for sub in data.hyperping_statuspage_subscribers.all.subscribers :
    sub.email if sub.type == "email"
  ]
}

# Filter subscribers by type
data "hyperping_statuspage_subscribers" "email_only" {
  statuspage_uuid = hyperping_statuspage.production.id
  type            = "email"
}

data "hyperping_statuspage_subscribers" "sms_only" {
  statuspage_uuid = hyperping_statuspage.production.id
  type            = "sms"
}

data "hyperping_statuspage_subscribers" "teams_only" {
  statuspage_uuid = hyperping_statuspage.production.id
  type            = "teams"
}

# Count subscribers by type
output "subscriber_breakdown" {
  value = {
    email = length(data.hyperping_statuspage_subscribers.email_only.subscribers)
    sms   = length(data.hyperping_statuspage_subscribers.sms_only.subscribers)
    teams = length(data.hyperping_statuspage_subscribers.teams_only.subscribers)
  }
}

# Reference existing status page
resource "hyperping_statuspage" "production" {
  name      = "Production Status"
  subdomain = "prod-status"
}

# Find specific subscriber by value
locals {
  team_subscriber = [
    for sub in data.hyperping_statuspage_subscribers.all.subscribers :
    sub if sub.email == "team@example.com"
  ]
}

output "team_subscriber_exists" {
  value = length(local.team_subscriber) > 0
}

# Use subscriber data for conditional logic
output "should_add_subscriber" {
  value       = length(data.hyperping_statuspage_subscribers.email_only.subscribers) == 0
  description = "True if no email subscribers exist"
}

# Group subscribers by language
locals {
  subscribers_by_language = {
    for sub in data.hyperping_statuspage_subscribers.all.subscribers :
    sub.language => sub...
  }
}

output "english_subscriber_count" {
  value = length(lookup(local.subscribers_by_language, "en", []))
}

# Export all subscriber details
output "all_subscribers" {
  value = [
    for sub in data.hyperping_statuspage_subscribers.all.subscribers : {
      id         = sub.id
      type       = sub.type
      value      = sub.value
      language   = sub.language
      created_at = sub.created_at
    }
  ]
  description = "Detailed information about all subscribers"
}
