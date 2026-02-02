# Read a single status page by ID
data "hyperping_statuspage" "existing" {
  id = "sp_abc123def456"
}

# Use status page data in other resources
output "status_page_url" {
  value = data.hyperping_statuspage.existing.url
}

output "status_page_theme" {
  value = data.hyperping_statuspage.existing.theme
}

# Read a status page created by another resource
resource "hyperping_statuspage" "production" {
  name      = "Production Status"
  subdomain = "prod-status"
}

data "hyperping_statuspage" "prod" {
  id = hyperping_statuspage.production.id
}

# Use status page data to create subscribers
resource "hyperping_statuspage_subscriber" "team" {
  statuspage_uuid = data.hyperping_statuspage.existing.id
  type            = "email"
  email           = "team@example.com"
  language        = "en"
}

# Access nested sections data
output "section_count" {
  value       = length(data.hyperping_statuspage.existing.sections)
  description = "Number of sections in the status page"
}

output "first_section_name" {
  value       = try(data.hyperping_statuspage.existing.sections[0].name["en"], "No sections")
  description = "Name of the first section"
}

# Access subscription settings
output "email_subscriptions_enabled" {
  value = data.hyperping_statuspage.existing.subscribe.email
}
