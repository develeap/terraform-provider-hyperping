# List all status pages
data "hyperping_statuspages" "all" {}

# Output all status page names
output "all_status_pages" {
  value = [
    for sp in data.hyperping_statuspages.all.statuspages : sp.name
  ]
}

# List status pages with search filter
data "hyperping_statuspages" "production" {
  search = "prod"
}

# Output filtered status pages
output "production_status_pages" {
  value = data.hyperping_statuspages.production.statuspages
}

# List status pages with pagination
data "hyperping_statuspages" "page_one" {
  page = 0
}

# Check if more pages exist
output "has_more_pages" {
  value = data.hyperping_statuspages.page_one.has_next_page
}

output "total_status_pages" {
  value = data.hyperping_statuspages.page_one.total
}

# Find a specific status page by name
locals {
  prod_status = [
    for sp in data.hyperping_statuspages.all.statuspages :
    sp if sp.name == "Production Status"
  ]
}

output "prod_status_id" {
  value = length(local.prod_status) > 0 ? local.prod_status[0].id : null
}

# Use status pages data to create conditional resources
resource "hyperping_statuspage_subscriber" "team_notifications" {
  for_each = {
    for sp in data.hyperping_statuspages.production.statuspages :
    sp.id => sp if sp.subdomain != null
  }

  statuspage_uuid = each.value.id
  type            = "email"
  email           = "team@example.com"
  language        = "en"
}

# Monitor status page count
output "status_page_count" {
  value       = length(data.hyperping_statuspages.all.statuspages)
  description = "Total number of status pages in your account"
}

# Extract all status page URLs
output "status_page_urls" {
  value = {
    for sp in data.hyperping_statuspages.all.statuspages :
    sp.name => sp.url
  }
}
