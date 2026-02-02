# ACME Corporation - Terraform Variables

tenant_id   = "acme"
tenant_name = "ACME Corporation"
active      = true
tags        = ["enterprise", "priority"]

status_page = {
  subdomain = "acme"
  uuid      = "sp_abc123def456"  # From Hyperping dashboard

  components = [
    {
      name          = "API Services"
      uuid          = "comp_api_001"
      group         = "Core Services"
      display_order = 1
    },
    {
      name          = "Web Application"
      uuid          = "comp_web_001"
      group         = "Core Services"
      display_order = 2
    },
    {
      name          = "Database"
      uuid          = "comp_db_001"
      group         = "Infrastructure"
      display_order = 3
    }
  ]
}

monitors = [
  {
    name            = "API-Health"
    url             = "https://api.acme.com/health"
    method          = "GET"
    frequency       = 30
    category        = "API"
    component       = "API Services"
    expected_status = "2xx"
    regions         = ["london", "virginia", "frankfurt", "singapore"]
  },
  {
    name            = "API-GraphQL"
    url             = "https://api.acme.com/graphql"
    method          = "POST"
    frequency       = 60
    category        = "API"
    component       = "API Services"
    headers = [
      { name = "Content-Type", value = "application/json" }
    ]
    body = "{\"query\": \"{ __typename }\"}"
  },
  {
    name      = "Website-Home"
    url       = "https://www.acme.com/"
    method    = "GET"
    frequency = 60
    category  = "Website"
    component = "Web Application"
  },
  {
    name      = "DB-Health"
    url       = "https://api.acme.com/db/health"
    method    = "GET"
    frequency = 120
    category  = "Database"
    component = "Database"
  }
]

# Reference to shared monitors (created in shared/ workspace)
shared_monitor_ids = {
  # "Auth-Service" = "mon_shared_auth_xxx"
  # "CDN-Health"   = "mon_shared_cdn_xxx"
}
