# Example: Password-protected status page

resource "hyperping_statuspage" "password_protected" {
  name             = "Private Status Page"
  hosted_subdomain = "private-status"

  # Set the password (sensitive field)
  password = "SecurePassword123!"

  settings = {
    name              = "Private Status Settings"
    languages         = ["en"]
    default_language  = "en"
    theme             = "dark"
    accent_color      = "#ff6b6b"

    # Enable password protection
    authentication = {
      password_protection = true
    }

    subscribe = {
      enabled = true
      email   = true
      sms     = false
      slack   = false
      teams   = false
    }
  }

  sections = []
}

# Output the status page URL (password not shown in output due to Sensitive flag)
output "status_page_url" {
  value = hyperping_statuspage.password_protected.url
}

# Note: The password field is marked as Sensitive in Terraform,
# so it will not appear in logs or console output.
# Visitors to the status page will need to enter this password.
