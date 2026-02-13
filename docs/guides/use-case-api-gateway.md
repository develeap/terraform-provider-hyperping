---
page_title: "Monitoring API Gateways with Terraform - Hyperping Provider"
description: |-
  Complete guide for monitoring API gateway endpoints (Kong, AWS API Gateway, NGINX, Traefik) with the Hyperping Terraform provider.
---

# Monitoring API Gateways with Terraform

Learn how to implement comprehensive monitoring for API gateway endpoints using the Hyperping Terraform provider, with patterns for Kong, AWS API Gateway, NGINX, and Traefik.

## The Challenge

API gateways introduce specific monitoring requirements:

- **Multiple Backends**: Gateway routes to many backend services
- **Authentication**: API keys, JWT tokens, OAuth flows
- **Rate Limiting**: Monitor gateway performance under load
- **Protocol Variety**: REST, GraphQL, WebSocket, gRPC endpoints
- **Version Management**: Multiple API versions (v1, v2, v3)
- **Geographic Distribution**: Edge gateways across multiple regions
- **Dynamic Routes**: Routes added/updated frequently

Manual monitoring doesn't scale with the dynamic nature of modern API gateways. This guide demonstrates automated, infrastructure-as-code monitoring.

## Architecture

```mermaid
graph TB
    subgraph "External Monitoring"
        Hyperping[Hyperping Monitors]
    end

    subgraph "API Gateway Layer"
        Gateway[API Gateway]
        Health[/health Endpoint]
        Metrics[/metrics Endpoint]
    end

    subgraph "API Endpoints"
        REST_V1[REST API v1]
        REST_V2[REST API v2]
        GraphQL[GraphQL API]
        WebSocket[WebSocket API]
    end

    subgraph "Backend Services"
        Auth[Auth Service]
        Users[User Service]
        Orders[Order Service]
        Products[Product Service]
    end

    Hyperping -.->|Health Check| Health
    Hyperping -.->|Metrics Check| Metrics
    Hyperping -.->|REST GET| REST_V1
    Hyperping -.->|REST POST| REST_V2
    Hyperping -.->|GraphQL Query| GraphQL

    Gateway --> REST_V1
    Gateway --> REST_V2
    Gateway --> GraphQL
    Gateway --> WebSocket

    REST_V1 --> Auth
    REST_V1 --> Users
    REST_V2 --> Orders
    REST_V2 --> Products
    GraphQL --> Users
    GraphQL --> Products

    style Hyperping fill:#4CAF50
    style Gateway fill:#2196F3
    style Health fill:#FF9800
    style GraphQL fill:#E91E63
```

**Monitoring Layers:**
- **Gateway Health**: Overall gateway availability
- **API Endpoints**: Individual endpoint monitoring
- **Authentication**: Auth flow validation
- **Rate Limits**: Performance under load

## Solution

### Prerequisites

- Terraform >= 1.0
- Hyperping account with API key
- API Gateway with externally accessible endpoints
- API authentication credentials (if required)

### Step 1: Define API Gateway Structure

Create a comprehensive inventory of your API gateway endpoints:

```hcl
# variables.tf
variable "gateway_domain" {
  description = "Primary API gateway domain"
  type        = string
  default     = "api.example.com"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "production"
}

variable "api_endpoints" {
  description = "API gateway endpoints to monitor"
  type = map(object({
    path                 = string
    http_method          = string
    expected_status_code = string
    check_frequency      = number
    critical             = bool
    regions              = list(string)
    request_body         = optional(string)
    request_headers      = optional(list(object({
      name  = string
      value = string
    })), [])
  }))
}

# terraform.tfvars
gateway_domain = "api.example.com"
environment    = "production"

api_endpoints = {
  gateway_health = {
    path                 = "/health"
    http_method          = "GET"
    expected_status_code = "200"
    check_frequency      = 30
    critical             = true
    regions              = ["london", "virginia", "singapore", "tokyo"]
  }

  api_v1_health = {
    path                 = "/v1/health"
    http_method          = "GET"
    expected_status_code = "200"
    check_frequency      = 60
    critical             = true
    regions              = ["london", "virginia", "singapore"]
  }

  api_v2_health = {
    path                 = "/v2/health"
    http_method          = "GET"
    expected_status_code = "200"
    check_frequency      = 60
    critical             = true
    regions              = ["london", "virginia", "singapore"]
  }

  graphql_introspection = {
    path                 = "/graphql"
    http_method          = "POST"
    expected_status_code = "200"
    check_frequency      = 300
    critical             = false
    regions              = ["london", "virginia"]
    request_body         = "{\"query\": \"{ __typename }\"}"
    request_headers = [
      {
        name  = "Content-Type"
        value = "application/json"
      }
    ]
  }

  auth_endpoint = {
    path                 = "/v1/auth/health"
    http_method          = "GET"
    expected_status_code = "200"
    check_frequency      = 30
    critical             = true
    regions              = ["london", "virginia", "singapore"]
  }

  rate_limit_test = {
    path                 = "/v1/ping"
    http_method          = "GET"
    expected_status_code = "2xx"  # Allow 200 or 429
    check_frequency      = 60
    critical             = false
    regions              = ["london"]
  }
}
```

### Step 2: Create Gateway Monitors

Implement monitors for all gateway endpoints:

```hcl
# main.tf
terraform {
  required_version = ">= 1.0"

  required_providers {
    hyperping = {
      source  = "develeap/hyperping"
      version = "~> 1.0"
    }
  }
}

provider "hyperping" {
  # Set HYPERPING_API_KEY environment variable
}

locals {
  name_prefix = "[${upper(var.environment)}]"

  # Group by criticality
  critical_endpoints = {
    for k, v in var.api_endpoints : k => v if v.critical
  }

  standard_endpoints = {
    for k, v in var.api_endpoints : k => v if !v.critical
  }
}

resource "hyperping_monitor" "api_gateway" {
  for_each = var.api_endpoints

  name     = "${local.name_prefix}-APIGateway-${replace(title(replace(each.key, "_", " ")), " ", "")}"
  url      = "https://${var.gateway_domain}${each.value.path}"
  protocol = "http"

  http_method     = each.value.http_method
  check_frequency = each.value.check_frequency
  regions         = each.value.regions

  expected_status_code = each.value.expected_status_code
  follow_redirects     = true

  request_body = each.value.request_body

  request_headers = each.value.request_headers
}
```

### Step 3: Add Authentication Monitoring

Monitor authenticated endpoints with proper credentials:

```hcl
# authenticated-endpoints.tf
variable "api_key" {
  description = "API gateway API key for monitoring"
  type        = string
  sensitive   = true
  # Set via: export TF_VAR_api_key=xxx
}

variable "jwt_token" {
  description = "JWT token for authenticated endpoint monitoring"
  type        = string
  sensitive   = true
  # Set via: export TF_VAR_jwt_token=xxx
}

variable "authenticated_endpoints" {
  description = "Endpoints requiring authentication"
  type = map(object({
    path      = string
    auth_type = string  # "api_key" or "bearer"
  }))

  default = {
    user_profile = {
      path      = "/v1/user/profile"
      auth_type = "bearer"
    }
    admin_dashboard = {
      path      = "/v1/admin/health"
      auth_type = "api_key"
    }
  }
}

resource "hyperping_monitor" "authenticated" {
  for_each = var.authenticated_endpoints

  name            = "${local.name_prefix}-APIGateway-Auth-${replace(title(replace(each.key, "_", " ")), " ", "")}"
  url             = "https://${var.gateway_domain}${each.value.path}"
  protocol        = "http"
  http_method     = "GET"
  check_frequency = 60
  regions         = ["london", "virginia"]

  expected_status_code = "2xx"

  request_headers = each.value.auth_type == "bearer" ? [
    {
      name  = "Authorization"
      value = "Bearer ${var.jwt_token}"
    }
  ] : [
    {
      name  = "X-API-Key"
      value = var.api_key
    }
  ]
}
```

### Step 4: Monitor Multiple API Versions

Track different API versions independently:

```hcl
# api-versions.tf
variable "api_versions" {
  description = "API versions to monitor"
  type        = list(string)
  default     = ["v1", "v2", "v3"]
}

locals {
  # Generate version-specific endpoints
  version_endpoints = flatten([
    for version in var.api_versions : [
      {
        key      = "${version}_health"
        version  = version
        path     = "/${version}/health"
        critical = version == "v3"  # Latest version is critical
      },
      {
        key      = "${version}_status"
        version  = version
        path     = "/${version}/status"
        critical = false
      }
    ]
  ])

  version_endpoint_map = {
    for endpoint in local.version_endpoints :
    endpoint.key => endpoint
  }
}

resource "hyperping_monitor" "api_versions" {
  for_each = local.version_endpoint_map

  name            = "${local.name_prefix}-APIGateway-${upper(each.value.version)}-${each.key}"
  url             = "https://${var.gateway_domain}${each.value.path}"
  protocol        = "http"
  http_method     = "GET"
  check_frequency = each.value.critical ? 30 : 300
  regions         = each.value.critical ? ["london", "virginia", "singapore"] : ["london"]

  expected_status_code = "200"
}
```

### Step 5: Create Status Page

Set up a public status page for your API:

```hcl
# statuspage.tf
resource "hyperping_statuspage" "api_gateway" {
  name      = "${title(var.environment)} API Gateway Status"
  subdomain = "${var.environment}-api"
  theme     = "dark"

  sections = [
    {
      name = {
        en = "Gateway Health"
      }
      is_split = false
      services = [
        for key in ["gateway_health"] : {
          monitor_uuid        = hyperping_monitor.api_gateway[key].id
          show_uptime         = true
          show_response_times = true
        } if contains(keys(hyperping_monitor.api_gateway), key)
      ]
    },
    {
      name = {
        en = "API Endpoints"
      }
      is_split = true
      services = [
        for key in ["api_v1_health", "api_v2_health"] : {
          monitor_uuid        = hyperping_monitor.api_gateway[key].id
          show_uptime         = true
          show_response_times = true
        } if contains(keys(hyperping_monitor.api_gateway), key)
      ]
    },
    {
      name = {
        en = "GraphQL & Special Endpoints"
      }
      is_split = true
      services = [
        for key in ["graphql_introspection"] : {
          monitor_uuid        = hyperping_monitor.api_gateway[key].id
          show_uptime         = true
          show_response_times = false
        } if contains(keys(hyperping_monitor.api_gateway), key)
      ]
    }
  ]
}

output "api_status_page_url" {
  description = "API Gateway status page URL"
  value       = "https://${hyperping_statuspage.api_gateway.subdomain}.hyperping.app"
}
```

## Complete Example

Here's a full working example for monitoring an API gateway:

```hcl
# main.tf
terraform {
  required_version = ">= 1.0"

  required_providers {
    hyperping = {
      source  = "develeap/hyperping"
      version = "~> 1.0"
    }
  }
}

provider "hyperping" {
  # Set HYPERPING_API_KEY environment variable
}

variable "gateway_domain" {
  description = "API Gateway domain"
  type        = string
  default     = "api.example.com"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "production"
}

locals {
  name_prefix = "[${upper(var.environment)}]"

  # Define all endpoints to monitor
  endpoints = {
    gateway_health = {
      path        = "/health"
      method      = "GET"
      status_code = "200"
      frequency   = 30
      critical    = true
    }
    api_v1_users = {
      path        = "/v1/users"
      method      = "GET"
      status_code = "2xx"
      frequency   = 60
      critical    = true
    }
    api_v2_orders = {
      path        = "/v2/orders"
      method      = "GET"
      status_code = "2xx"
      frequency   = 60
      critical    = true
    }
    graphql = {
      path        = "/graphql"
      method      = "POST"
      status_code = "200"
      frequency   = 300
      critical    = false
      body        = jsonencode({ query = "{ __typename }" })
      headers = [
        {
          name  = "Content-Type"
          value = "application/json"
        }
      ]
    }
  }
}

resource "hyperping_monitor" "gateway" {
  for_each = local.endpoints

  name            = "${local.name_prefix}-APIGateway-${replace(title(replace(each.key, "_", " ")), " ", "")}"
  url             = "https://${var.gateway_domain}${each.value.path}"
  protocol        = "http"
  http_method     = each.value.method
  check_frequency = each.value.frequency
  regions         = each.value.critical ? ["london", "virginia", "singapore"] : ["london"]

  expected_status_code = each.value.status_code
  follow_redirects     = true

  request_body = try(each.value.body, null)

  request_headers = try(each.value.headers, [])
}

resource "hyperping_statuspage" "api" {
  name      = "${title(var.environment)} API"
  subdomain = "${var.environment}-api"
  theme     = "dark"

  sections = [{
    name = { en = "API Gateway" }
    is_split = true
    services = [
      for k, v in hyperping_monitor.gateway : {
        monitor_uuid        = v.id
        show_uptime         = true
        show_response_times = local.endpoints[k].critical
      }
    ]
  }]
}

output "status_page_url" {
  value = "https://${hyperping_statuspage.api.subdomain}.hyperping.app"
}

output "monitors" {
  value = {
    for k, v in hyperping_monitor.gateway : k => {
      id   = v.id
      name = v.name
      url  = v.url
    }
  }
}
```

**Usage:**

```bash
# Set API key
export HYPERPING_API_KEY="sk_your_api_key"

# Initialize
terraform init

# Plan
terraform plan -var="gateway_domain=api.example.com"

# Apply
terraform apply -var="gateway_domain=api.example.com"

# Check status
terraform output status_page_url
```

## Customization

### Kong Gateway Integration

Monitor Kong-specific endpoints:

```hcl
# kong.tf
locals {
  kong_admin_endpoints = {
    status = {
      path   = "/status"
      port   = 8001
      public = false
    }
    services = {
      path   = "/services"
      port   = 8001
      public = false
    }
  }

  kong_proxy_routes = {
    users = {
      path   = "/api/users"
      port   = 8000
      public = true
    }
    orders = {
      path   = "/api/orders"
      port   = 8000
      public = true
    }
  }
}

# Monitor Kong admin API (internal)
resource "hyperping_monitor" "kong_admin" {
  for_each = local.kong_admin_endpoints

  name            = "${local.name_prefix}-Kong-Admin-${title(each.key)}"
  url             = "http://admin.kong.internal:${each.value.port}${each.value.path}"
  protocol        = "http"
  check_frequency = 300
  regions         = ["london"]

  expected_status_code = "200"
}

# Monitor Kong proxy routes (public)
resource "hyperping_monitor" "kong_proxy" {
  for_each = local.kong_proxy_routes

  name            = "${local.name_prefix}-Kong-${title(each.key)}"
  url             = "https://${var.gateway_domain}${each.value.path}"
  protocol        = "http"
  check_frequency = 60
  regions         = ["london", "virginia", "singapore"]

  expected_status_code = "2xx"
}
```

### AWS API Gateway

Monitor AWS API Gateway with custom domain:

```hcl
# aws-api-gateway.tf
variable "aws_api_gateway_id" {
  description = "AWS API Gateway ID"
  type        = string
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

locals {
  aws_gateway_url = "https://${var.aws_api_gateway_id}.execute-api.${var.aws_region}.amazonaws.com"

  aws_stages = ["prod", "staging"]

  aws_endpoints = flatten([
    for stage in local.aws_stages : [
      {
        key   = "${stage}_health"
        stage = stage
        path  = "/${stage}/health"
      },
      {
        key   = "${stage}_api"
        stage = stage
        path  = "/${stage}/api/v1/status"
      }
    ]
  ])
}

resource "hyperping_monitor" "aws_gateway" {
  for_each = {
    for endpoint in local.aws_endpoints :
    endpoint.key => endpoint
  }

  name            = "${local.name_prefix}-AWS-APIGateway-${title(each.value.stage)}"
  url             = "${local.aws_gateway_url}${each.value.path}"
  protocol        = "http"
  check_frequency = each.value.stage == "prod" ? 30 : 300
  regions         = each.value.stage == "prod" ? ["virginia", "london", "singapore"] : ["virginia"]

  expected_status_code = "200"
}
```

### GraphQL Endpoint Monitoring

Monitor GraphQL with health queries:

```hcl
# graphql.tf
variable "graphql_queries" {
  description = "GraphQL health check queries"
  type = map(object({
    query           = string
    expected_status = string
    frequency       = number
  }))

  default = {
    introspection = {
      query           = "{ __typename }"
      expected_status = "200"
      frequency       = 300
    }
    health_check = {
      query           = "{ health { status version } }"
      expected_status = "200"
      frequency       = 60
    }
  }
}

resource "hyperping_monitor" "graphql" {
  for_each = var.graphql_queries

  name            = "${local.name_prefix}-GraphQL-${replace(title(replace(each.key, "_", " ")), " ", "")}"
  url             = "https://${var.gateway_domain}/graphql"
  protocol        = "http"
  http_method     = "POST"
  check_frequency = each.value.frequency
  regions         = ["london", "virginia"]

  request_body = jsonencode({
    query = each.value.query
  })

  request_headers = [
    {
      name  = "Content-Type"
      value = "application/json"
    },
    {
      name  = "Accept"
      value = "application/json"
    }
  ]

  expected_status_code = each.value.expected_status
}
```

### Rate Limit Monitoring

Monitor gateway rate limit behavior:

```hcl
# rate-limits.tf
resource "hyperping_monitor" "rate_limit_test" {
  name     = "${local.name_prefix}-APIGateway-RateLimitTest"
  url      = "https://${var.gateway_domain}/v1/ping"
  protocol = "http"

  http_method     = "GET"
  check_frequency = 60
  regions         = ["london"]

  # Accept both success and rate limit responses
  expected_status_code = "2xx"  # Will match 200 or 429

  request_headers = [
    {
      name  = "X-Test-Rate-Limit"
      value = "true"
    }
  ]
}

# Monitor rate limit metrics endpoint
resource "hyperping_monitor" "rate_limit_metrics" {
  name     = "${local.name_prefix}-APIGateway-RateLimitMetrics"
  url      = "https://${var.gateway_domain}/metrics/rate-limits"
  protocol = "http"

  http_method     = "GET"
  check_frequency = 300
  regions         = ["london"]

  expected_status_code = "200"
}
```

## Best Practices

### 1. Separate Public and Private Endpoints

Monitor internal and external endpoints differently:

```hcl
locals {
  public_endpoints = {
    for k, v in var.api_endpoints : k => v
    if try(v.public, true)
  }

  private_endpoints = {
    for k, v in var.api_endpoints : k => v
    if !try(v.public, true)
  }
}

# Public endpoints: multi-region
resource "hyperping_monitor" "public" {
  for_each = local.public_endpoints
  regions  = ["london", "virginia", "singapore"]
  # ...
}

# Private endpoints: single region
resource "hyperping_monitor" "private" {
  for_each = local.private_endpoints
  regions  = ["london"]
  # ...
}
```

### 2. Version Your API Monitoring

Track deprecation of old API versions:

```hcl
locals {
  api_version_status = {
    v1 = { supported = true, critical = false }   # Legacy
    v2 = { supported = true, critical = true }    # Current
    v3 = { supported = true, critical = true }    # Latest
  }
}

resource "hyperping_monitor" "versioned" {
  for_each = local.api_version_status

  check_frequency = each.value.critical ? 30 : 600
  # ...
}
```

### 3. Use Realistic Test Data

Send realistic payloads for POST/PUT monitoring:

```hcl
resource "hyperping_monitor" "create_user" {
  name        = "${local.name_prefix}-APIGateway-CreateUserTest"
  url         = "https://${var.gateway_domain}/v1/users/test"
  http_method = "POST"

  request_body = jsonencode({
    email    = "test@hyperping-monitor.example.com"
    username = "hyperping_test_${timestamp()}"
    test     = true
  })

  request_headers = [
    {
      name  = "Content-Type"
      value = "application/json"
    },
    {
      name  = "X-Test-Mode"
      value = "true"  # Ensure test mode is enabled
    }
  ]
}
```

### 4. Monitor Edge Locations

For CDN/edge gateways, monitor from appropriate regions:

```hcl
variable "edge_locations" {
  description = "Edge gateway locations"
  type = map(list(string))

  default = {
    us_east      = ["virginia"]
    us_west      = ["oregon"]
    eu_west      = ["london", "frankfurt"]
    asia_pacific = ["singapore", "tokyo", "sydney"]
  }
}

resource "hyperping_monitor" "edge" {
  for_each = var.edge_locations

  name    = "${local.name_prefix}-EdgeGateway-${replace(title(replace(each.key, "_", " ")), " ", "")}"
  regions = each.value
  # ...
}
```

### 5. Implement Circuit Breaker Monitoring

Monitor gateway circuit breaker status:

```hcl
resource "hyperping_monitor" "circuit_breaker_status" {
  name            = "${local.name_prefix}-APIGateway-CircuitBreakerStatus"
  url             = "https://${var.gateway_domain}/internal/circuit-breakers"
  protocol        = "http"
  check_frequency = 60
  regions         = ["london"]

  # Expect healthy response
  expected_status_code = "200"

  request_headers = [
    {
      name  = "X-Internal-Health-Check"
      value = var.internal_health_check_token
    }
  ]
}
```

## Troubleshooting

### Issue: Authentication Failures

**Problem:** Monitors fail with 401/403 errors.

**Solution:** Verify credentials and token expiration:

```bash
# Test endpoint manually
curl -H "Authorization: Bearer $TOKEN" https://api.example.com/v1/health

# Check token expiration
echo $JWT_TOKEN | cut -d. -f2 | base64 -d | jq .exp

# Refresh token if needed
export TF_VAR_jwt_token=$(./scripts/refresh-token.sh)
terraform apply
```

### Issue: GraphQL Queries Failing

**Problem:** GraphQL monitors return errors.

**Solution:** Test query format:

```bash
# Test GraphQL query
curl -X POST https://api.example.com/graphql \
  -H "Content-Type: application/json" \
  -d '{"query": "{ __typename }"}'

# Verify response format
# Should return: {"data": {"__typename": "Query"}}
```

### Issue: Rate Limit False Positives

**Problem:** Monitors frequently hit rate limits.

**Solution:** Use dedicated health check endpoints or adjust frequency:

```hcl
# Use dedicated health endpoint
resource "hyperping_monitor" "health" {
  url = "https://${var.gateway_domain}/health"  # Exempt from rate limits

  request_headers = [
    {
      name  = "X-Health-Check"
      value = "hyperping"  # Gateway exempts this header
    }
  ]
}

# Or reduce check frequency
resource "hyperping_monitor" "standard" {
  check_frequency = 300  # 5 minutes instead of 60 seconds
}
```

### Issue: CORS Errors

**Problem:** Monitors fail due to CORS.

**Solution:** CORS only affects browsers, verify actual error:

```bash
# Test without browser
curl -v https://api.example.com/v1/health

# If failing, likely not CORS but auth/network issue
```

### Issue: Gateway Timeouts

**Problem:** Monitors timeout before gateway responds.

**Solution:** Gateway may need longer timeout setting (not configurable in Hyperping, use appropriate check frequency):

```hcl
# For slow endpoints, use less frequent checks
resource "hyperping_monitor" "slow_endpoint" {
  check_frequency = 300  # Give more time between checks
  # Hyperping default timeout is adequate for most cases
}
```

## Next Steps

- **[Microservices Monitoring Guide](./use-case-microservices.md)** - Monitor service mesh
- **[Kubernetes Monitoring Guide](./use-case-kubernetes.md)** - Monitor K8s ingress
- **[Rate Limits Guide](./rate-limits.md)** - Handle API rate limits
- **[Error Handling Guide](./error-handling.md)** - Handle monitoring failures

## Additional Resources

- [Hyperping Provider Documentation](https://registry.terraform.io/providers/develeap/hyperping/latest/docs)
- [Kong Gateway Documentation](https://docs.konghq.com/)
- [AWS API Gateway Monitoring](https://docs.aws.amazon.com/apigateway/latest/developerguide/monitoring-cloudwatch.html)
- [GraphQL Best Practices](https://graphql.org/learn/best-practices/)
- [API Gateway Patterns](https://microservices.io/patterns/apigateway.html)
