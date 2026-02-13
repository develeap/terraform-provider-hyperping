# GraphQL Monitor Module

Reusable Terraform module for GraphQL endpoint health monitoring with Hyperping. Supports query validation, introspection checks, and custom headers.

## Features

- GraphQL query health checks with POST requests
- Response validation using expected keywords
- GraphQL introspection monitoring (optional)
- Support for query variables
- Custom headers for GraphQL requests (non-reserved headers only)
- Multi-region monitoring

## Important: Authentication Limitations

The Hyperping provider blocks reserved HTTP headers (Authorization, Cookie, Host, etc.) for security reasons. This module is designed for:

1. **Public GraphQL endpoints** (no authentication)
2. **Custom authentication headers** (X-API-Key, X-Auth-Token, etc.)

If your GraphQL API requires standard `Authorization` headers, you cannot use this module due to provider security restrictions.

## Usage

### Basic Health Check

```hcl
module "graphql_api" {
  source = "path/to/modules/graphql-monitor"

  endpoint = "https://api.example.com/graphql"

  queries = {
    "health" = {
      query             = "{ health { status } }"
      expected_response = "\"status\":\"ok\""
    }
  }
}
```

### Multiple Queries

```hcl
module "graphql_api" {
  source = "path/to/modules/graphql-monitor"

  endpoint = "https://api.example.com/graphql"

  queries = {
    "health" = {
      query             = "{ health { status } }"
      expected_response = "\"status\":\"ok\""
      frequency         = 60
    }
    "user-query" = {
      query             = "{ user(id: 1) { name email } }"
      expected_response = "\"name\""
      frequency         = 120
    }
    "products" = {
      query             = "{ products { id name } }"
      expected_response = "\"products\""
      frequency         = 300
    }
  }

  name_prefix     = "PROD"
  default_regions = ["virginia", "london", "frankfurt"]
}
```

### Query with Variables

```hcl
module "graphql_api" {
  source = "path/to/modules/graphql-monitor"

  endpoint = "https://api.example.com/graphql"

  queries = {
    "user-by-id" = {
      query = <<-GRAPHQL
        query GetUser($userId: ID!) {
          user(id: $userId) {
            name
            email
            status
          }
        }
      GRAPHQL
      variables = {
        userId = "test-user-123"
      }
      expected_response = "\"status\""
    }
  }
}
```

### Custom Authentication Headers

For APIs using non-reserved authentication headers:

```hcl
module "graphql_api" {
  source = "path/to/modules/graphql-monitor"

  endpoint = "https://api.example.com/graphql"

  custom_headers = {
    "X-API-Key"     = var.graphql_api_key
    "X-Client-ID"   = var.client_id
  }

  queries = {
    "health" = {
      query             = "{ health { status } }"
      expected_response = "\"status\":\"ok\""
    }
  }
}
```

### With Custom Headers

```hcl
module "graphql_api" {
  source = "path/to/modules/graphql-monitor"

  endpoint = "https://api.example.com/graphql"

  custom_headers = {
    "X-Request-ID"  = "monitor-check"
    "X-Environment" = "production"
    "Accept"        = "application/json"
  }

  queries = {
    "health" = {
      query             = "{ health { status } }"
      expected_response = "\"status\":\"ok\""
    }
  }
}
```

### With Introspection Check

```hcl
module "graphql_api" {
  source = "path/to/modules/graphql-monitor"

  endpoint = "https://api.example.com/graphql"

  queries = {
    "health" = {
      query             = "{ health { status } }"
      expected_response = "\"status\":\"ok\""
    }
  }

  # Enable introspection monitoring (runs less frequently)
  enable_introspection_check     = true
  introspection_frequency        = 3600  # Once per hour
  introspection_expected_response = "queryType"
}
```

### Advanced Configuration

```hcl
module "graphql_api" {
  source = "path/to/modules/graphql-monitor"

  endpoint = "https://api.example.com/graphql"

  queries = {
    "critical-health" = {
      query                = "{ health { status database cache } }"
      expected_response    = "\"status\":\"ok\""
      frequency            = 30  # Check every 30 seconds
      expected_status_code = "200"
      regions              = ["virginia", "london", "singapore"]  # Override default
    }
    "user-stats" = {
      query             = "{ stats { activeUsers totalUsers } }"
      expected_response = "\"activeUsers\""
      frequency         = 300  # Check every 5 minutes
    }
  }

  name_prefix        = "PROD"
  default_frequency  = 120
  default_regions    = ["virginia", "london"]
  alerts_wait        = 60  # Wait 60 seconds before alerting
  escalation_policy  = var.escalation_policy_id
}
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| `endpoint` | GraphQL endpoint URL | `string` | n/a | yes |
| `queries` | Map of GraphQL query configurations | `map(object)` | n/a | yes |
| `custom_headers` | Additional custom headers (non-reserved only) | `map(string)` | `null` | no |
| `name_prefix` | Prefix for monitor names | `string` | `""` | no |
| `name_format` | Custom name format (use %s for key) | `string` | `""` | no |
| `default_regions` | Default monitoring regions | `list(string)` | `["virginia", "london", "frankfurt"]` | no |
| `default_frequency` | Default check frequency (seconds) | `number` | `120` | no |
| `alerts_wait` | Seconds before alerting | `number` | `null` | no |
| `escalation_policy` | Escalation policy UUID | `string` | `null` | no |
| `paused` | Create monitors in paused state | `bool` | `false` | no |
| `enable_introspection_check` | Enable introspection monitor | `bool` | `false` | no |
| `introspection_query` | Custom introspection query | `string` | `"{ __schema { queryType { name } } }"` | no |
| `introspection_expected_response` | Expected introspection response keyword | `string` | `"queryType"` | no |
| `introspection_frequency` | Introspection check frequency (seconds) | `number` | `3600` | no |
| `introspection_expected_status` | Expected introspection status code | `string` | `"200"` | no |

### Query Object

| Field | Description | Type | Default |
|-------|-------------|------|---------|
| `query` | GraphQL query or mutation string | `string` | required |
| `variables` | Query variables as map | `map(any)` | `null` |
| `expected_response` | Keyword that must appear in response | `string` | required |
| `frequency` | Check frequency (seconds) | `number` | uses `default_frequency` |
| `expected_status_code` | Expected HTTP status | `string` | `"200"` |
| `follow_redirects` | Follow HTTP redirects | `bool` | `true` |
| `regions` | Override default regions | `list(string)` | uses `default_regions` |
| `paused` | Override default paused state | `bool` | uses `paused` |

## Outputs

| Name | Description |
|------|-------------|
| `monitor_ids` | Map of query name to monitor UUID |
| `monitor_ids_list` | List of all query monitor UUIDs |
| `introspection_monitor_id` | UUID of introspection monitor (if enabled) |
| `all_monitor_ids` | List of all monitor UUIDs including introspection |
| `monitors` | Full monitor objects for advanced usage |
| `endpoint` | GraphQL endpoint URL being monitored |
| `query_count` | Total number of query monitors created |
| `total_monitor_count` | Total number of monitors created |

## Authentication Options

### Option 1: Public Endpoints (No Authentication)

Use this module directly for public GraphQL endpoints:

```hcl
module "public_graphql" {
  source   = "path/to/modules/graphql-monitor"
  endpoint = "https://public-api.example.com/graphql"
  # ... queries ...
}
```

### Option 2: Custom Authentication Headers

For APIs using non-reserved headers:

```hcl
module "custom_auth_graphql" {
  source   = "path/to/modules/graphql-monitor"
  endpoint = "https://api.example.com/graphql"

  custom_headers = {
    "X-API-Key" = var.api_key
  }
  # ... queries ...
}
```

### Option 3: Not Supported - Standard Authorization

**This module cannot be used** for APIs requiring standard `Authorization: Bearer <token>` headers due to provider security restrictions.

For such APIs, you must use the `hyperping_monitor` resource directly without this module.

## GraphQL Query Best Practices

### Keep Queries Simple

Monitor queries should be lightweight and fast:

```hcl
# GOOD: Simple health check
query = "{ health { status } }"

# AVOID: Complex queries with many nested fields
query = "{ users { posts { comments { author { profile { ... } } } } } }"
```

### Use Specific Expected Responses

Choose unique keywords that confirm the query executed correctly:

```hcl
# GOOD: Specific field
expected_response = "\"status\":\"ok\""

# LESS IDEAL: Generic field
expected_response = "data"
```

### Validate Critical Fields

```hcl
queries = {
  "health-check" = {
    query = "{ health { status database cache queue } }"
    # Validate that all systems are reporting
    expected_response = "\"database\""
  }
}
```

### Use Query Variables for Dynamic Values

```hcl
queries = {
  "test-user" = {
    query = "query GetUser($id: ID!) { user(id: $id) { status } }"
    variables = {
      id = "test-user-123"
    }
    expected_response = "\"status\""
  }
}
```

## Valid Regions

```
paris, frankfurt, amsterdam, london, singapore, sydney, tokyo, seoul,
mumbai, bangalore, virginia, california, sanfrancisco, nyc,
toronto, saopaulo, bahrain, capetown
```

## Valid Frequencies (seconds)

```
10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400
```

## Common GraphQL Monitoring Queries

### Health Check

```graphql
{
  health {
    status
  }
}
```

### Service Status Check

```graphql
{
  health {
    status
    database
    cache
    queue
  }
}
```

### Introspection (Schema Validation)

```graphql
{
  __schema {
    queryType {
      name
    }
  }
}
```

### Version Check

```graphql
{
  version {
    api
    schema
  }
}
```

## Examples Directory Structure

```
examples/modules/graphql-monitor/
├── main.tf           # Main resource definitions
├── variables.tf      # Input variable definitions
├── outputs.tf        # Output value definitions
├── versions.tf       # Provider version requirements
├── README.md         # This file
└── tests/
    ├── basic.tftest.hcl          # Basic functionality tests
    └── variables.tftest.hcl      # Variable handling tests
```

## Testing

Run tests using Terraform's native testing framework:

```bash
cd examples/modules/graphql-monitor
terraform init
terraform test
```

## Limitations

1. **No Standard Authorization Headers**: The provider blocks `Authorization`, `Cookie`, `Host`, and other reserved headers for security
2. **Public or Custom Auth Only**: This module works only with public endpoints or custom authentication headers (X-API-Key, etc.)
3. **No OAuth/JWT Support**: Standard OAuth2 Bearer tokens cannot be used through this module

## Alternatives for Authenticated APIs

If your GraphQL API requires standard `Authorization` headers, use the `hyperping_monitor` resource directly:

```hcl
resource "hyperping_monitor" "graphql_with_auth" {
  name     = "GraphQL API Health"
  url      = "https://api.example.com/graphql"
  protocol = "http"
  http_method = "POST"

  request_headers = [
    { name = "Content-Type", value = "application/json" }
    # Note: Authorization header will be rejected by provider
  ]

  request_body = jsonencode({
    query = "{ health { status } }"
  })

  required_keyword = "\"status\":\"ok\""
}
```

## Notes

- All GraphQL requests are sent as POST with JSON body
- Content-Type header is automatically set to `application/json`
- Response validation uses the `required_keyword` field to check for expected content
- Introspection checks can help detect schema changes or API availability issues
- Default frequency is 120 seconds (2 minutes), suitable for GraphQL endpoints
- Use different regions for global availability monitoring
