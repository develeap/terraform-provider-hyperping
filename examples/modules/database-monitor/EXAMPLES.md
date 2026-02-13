# Database Monitor Module - Usage Examples

This file demonstrates various ways to use the database-monitor module.

## Setup

First, configure the Hyperping provider in your root module:

```hcl
terraform {
  required_version = ">= 1.0"

  required_providers {
    hyperping = {
      source  = "develeap/hyperping"
      version = ">= 1.0"
    }
  }
}

provider "hyperping" {
  # API key should be set via HYPERPING_API_KEY environment variable
}
```

## Examples

### Example 1: Basic database monitoring

```hcl
module "basic_databases" {
  source = "./modules/database-monitor"

  name_prefix = "PROD"

  databases = {
    postgres = {
      host = "db.example.com"
      port = 5432
      type = "postgresql"
    }
    redis = {
      host = "cache.example.com"
      port = 6379
      type = "redis"
    }
  }
}
```

### Example 2: Multi-region setup with custom frequencies

```hcl
module "prod_databases" {
  source = "./modules/database-monitor"

  name_prefix = "PROD"

  databases = {
    primary-postgres = {
      host      = "prod-db-primary.example.com"
      port      = 5432
      type      = "postgresql"
      frequency = 30 # Check every 30 seconds for critical DB
      regions   = ["virginia", "london", "tokyo"]
    }
    replica-postgres = {
      host = "prod-db-replica.example.com"
      port = 5432
      type = "postgresql"
    }
    session-redis = {
      host      = "prod-redis-session.example.com"
      port      = 6379
      type      = "redis"
      frequency = 30
    }
    cache-redis = {
      host = "prod-redis-cache.example.com"
      port = 6379
      type = "redis"
    }
  }

  default_frequency = 60
  default_regions   = ["virginia", "london"]
  alerts_wait       = 30 # Wait 30 seconds before alerting
}
```

### Example 3: Staging environment (paused for cost savings)

```hcl
module "staging_databases" {
  source = "./modules/database-monitor"

  name_prefix = "STAGING"

  databases = {
    postgres = {
      host = "staging-db.example.com"
      port = 5432
      type = "postgresql"
    }
  }

  paused          = true # Start paused
  default_regions = ["virginia"]
}
```

### Example 4: Multiple database types

```hcl
module "all_databases" {
  source = "./modules/database-monitor"

  name_prefix = "PROD"

  databases = {
    postgres = {
      host = "db.example.com"
      port = 5432
      type = "postgresql"
    }
    mysql = {
      host = "mysql.example.com"
      port = 3306
      type = "mysql"
    }
    redis = {
      host = "cache.example.com"
      port = 6379
      type = "redis"
    }
    mongodb = {
      host = "docs.example.com"
      port = 27017
      type = "mongodb"
    }
    cassandra = {
      host = "wide.example.com"
      port = 9042
      type = "cassandra"
    }
  }

  default_regions = ["virginia", "london"]
}
```

### Example 5: Custom naming without type prefix

```hcl
module "simple_names" {
  source = "./modules/database-monitor"

  name_format = "DB - %s"

  databases = {
    main = {
      host = "db.example.com"
      port = 5432
      type = "postgresql"
    }
  }

  include_type_in_name = false
}
```

### Example 6: Regional database clusters

```hcl
module "multi_region_clusters" {
  source = "./modules/database-monitor"

  name_prefix = "PROD"

  databases = {
    db-us-east = {
      host    = "db-us-east.example.com"
      port    = 5432
      type    = "postgresql"
      regions = ["virginia"]
    }
    db-eu-west = {
      host    = "db-eu-west.example.com"
      port    = 5432
      type    = "postgresql"
      regions = ["london", "frankfurt"]
    }
    db-ap-southeast = {
      host    = "db-ap-southeast.example.com"
      port    = 5432
      type    = "postgresql"
      regions = ["singapore", "sydney", "tokyo"]
    }
  }
}
```

## Outputs

```hcl
output "basic_monitor_ids" {
  description = "Monitor UUIDs from basic example"
  value       = module.basic_databases.monitor_ids
}

output "prod_connection_strings" {
  description = "Connection strings for production databases"
  value       = module.prod_databases.connection_strings
}

output "all_databases_by_type" {
  description = "Databases grouped by type"
  value       = module.all_databases.monitors_by_type
}
```
