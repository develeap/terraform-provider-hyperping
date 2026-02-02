# Email subscriber
resource "hyperping_statuspage_subscriber" "team_email" {
  statuspage_uuid = hyperping_statuspage.production.id
  type            = "email"
  email           = "team@example.com"
  language        = "en"
}

# SMS subscriber
resource "hyperping_statuspage_subscriber" "oncall_sms" {
  statuspage_uuid = hyperping_statuspage.production.id
  type            = "sms"
  phone           = "+1234567890"
  language        = "en"
}

# Microsoft Teams webhook subscriber
resource "hyperping_statuspage_subscriber" "teams_channel" {
  statuspage_uuid   = hyperping_statuspage.production.id
  type              = "teams"
  teams_webhook_url = "https://outlook.office.com/webhook/abc123..."
  language          = "en"
}

# Multiple subscribers using for_each
locals {
  email_subscribers = {
    engineering = {
      email    = "engineering@example.com"
      language = "en"
    }
    operations = {
      email    = "ops@example.com"
      language = "en"
    }
    executives = {
      email    = "exec@example.com"
      language = "fr"
    }
  }
}

resource "hyperping_statuspage_subscriber" "team_emails" {
  for_each = local.email_subscribers

  statuspage_uuid = hyperping_statuspage.production.id
  type            = "email"
  email           = each.value.email
  language        = each.value.language
}

# Reference status page from another resource
resource "hyperping_statuspage" "production" {
  name      = "Production Status"
  subdomain = "prod-status"
}

# Note: Slack subscribers must be configured via Hyperping OAuth flow
# They cannot be added through the Terraform provider
# To add Slack notifications:
# 1. Log into Hyperping dashboard
# 2. Navigate to your status page settings
# 3. Connect Slack via OAuth
# 4. Configure channel notifications

# Import existing subscriber
# terraform import hyperping_statuspage_subscriber.existing sp_abc123:456
