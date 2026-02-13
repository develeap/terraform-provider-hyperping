# Database Monitor Module

Reusable Terraform module for TCP port monitoring of database connectivity with Hyperping.

## Features

- TCP port monitoring for popular databases (PostgreSQL, MySQL, Redis, MongoDB, etc.)
- Support for multiple databases using `for_each` pattern
- Regional redundancy with configurable regions
- Configurable check frequency and timeout
- Optional type-based naming (e.g., "PostgreSQL - mydb")
- Built-in validation for ports and database types

## Usage

### Basic PostgreSQL and Redis

```hcl
module "databases" {
  source = "path/to/modules/database-monitor"

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

  default_regions = ["virginia", "london"]
}
```

### Multiple Databases with Custom Frequencies

```hcl
module "databases" {
  source = "path/to/modules/database-monitor"

  name_prefix = "PROD"

  databases = {
    primary-db = {
      host      = "db-primary.example.com"
      port      = 5432
      type      = "postgresql"
      frequency = 30  # Check every 30 seconds
    }
    replica-db = {
      host      = "db-replica.example.com"
      port      = 5432
      type      = "postgresql"
      frequency = 60
    }
    session-store = {
      host      = "redis.example.com"
      port      = 6379
      type      = "redis"
      frequency = 30  # Check more frequently
    }
    document-db = {
      host = "mongo.example.com"
      port = 27017
      type = "mongodb"
    }
  }

  default_regions  = ["virginia", "london", "frankfurt"]
  default_frequency = 60
}
```

### Multi-Region Database Cluster

```hcl
module "databases" {
  source = "path/to/modules/database-monitor"

  name_prefix = "PROD"

  databases = {
    db-us-east = {
      host    = "db-us-east.example.com"
      port    = 5432
      type    = "postgresql"
      regions = ["virginia", "nyc"]  # Monitor from nearby regions
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
      regions = ["singapore", "sydney"]
    }
  }

  include_type_in_name = true
}
```

### With Escalation Policy

```hcl
resource "hyperping_escalation_policy" "critical_db" {
  name = "Critical Database Team"
  # ... escalation policy configuration
}

module "databases" {
  source = "path/to/modules/database-monitor"

  name_prefix = "PROD"

  databases = {
    postgres = {
      host = "db.example.com"
      port = 5432
      type = "postgresql"
    }
  }

  escalation_policy = hyperping_escalation_policy.critical_db.id
  alerts_wait       = 0  # Alert immediately on failure
}
```

### Custom Name Format

```hcl
module "databases" {
  source = "path/to/modules/database-monitor"

  name_format = "Database: %s"  # Override default [PREFIX] format

  databases = {
    main = {
      host = "db.example.com"
      port = 5432
      type = "postgresql"
    }
  }

  include_type_in_name = false  # Results in "Database: main"
}
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| `databases` | Map of database configurations | `map(object)` | n/a | yes |
| `name_prefix` | Prefix for monitor names | `string` | `""` | no |
| `name_format` | Custom name format (use %s for key) | `string` | `""` | no |
| `default_regions` | Default monitoring regions | `list(string)` | `["virginia", "london"]` | no |
| `default_frequency` | Default check frequency (seconds) | `number` | `60` | no |
| `alerts_wait` | Seconds before alerting | `number` | `null` | no |
| `escalation_policy` | Escalation policy UUID | `string` | `null` | no |
| `paused` | Create monitors in paused state | `bool` | `false` | no |
| `include_type_in_name` | Include database type in name | `bool` | `true` | no |

### Database Object

| Field | Description | Type | Default |
|-------|-------------|------|---------|
| `host` | Database hostname or IP | `string` | required |
| `port` | Database port number | `number` | required |
| `type` | Database type | `string` | required |
| `frequency` | Check frequency (seconds) | `number` | uses `default_frequency` |
| `regions` | Override default regions | `list(string)` | uses `default_regions` |
| `paused` | Override default paused state | `bool` | uses `paused` |

## Outputs

| Name | Description |
|------|-------------|
| `monitor_ids` | Map of database name to monitor UUID |
| `monitor_ids_list` | List of all monitor UUIDs |
| `monitors` | Full monitor objects for advanced usage |
| `database_count` | Total number of monitors created |
| `monitors_by_type` | Monitors grouped by database type |
| `connection_strings` | Database endpoints (host:port format) |

## Supported Database Types

| Type | Default Port | Aliases |
|------|--------------|---------|
| PostgreSQL | 5432 | `postgresql`, `postgres` |
| MySQL | 3306 | `mysql` |
| MariaDB | 3306 | `mariadb` |
| Redis | 6379 | `redis` |
| MongoDB | 27017 | `mongodb`, `mongo` |
| Cassandra | 9042 | `cassandra` |
| Elasticsearch | 9200 | `elasticsearch` |
| Memcached | 11211 | `memcached` |
| RabbitMQ | 5672 | `rabbitmq` |

## Valid Regions

```
london, frankfurt, singapore, sydney, tokyo, virginia, saopaulo, bahrain
```

## Valid Frequencies (seconds)

```
10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400
```

## Examples

### Complete Production Setup

```hcl
module "prod_databases" {
  source = "path/to/modules/database-monitor"

  name_prefix = "PROD"

  databases = {
    primary-postgres = {
      host = "prod-db-primary.example.com"
      port = 5432
      type = "postgresql"
      frequency = 30
      regions = ["virginia", "london", "tokyo"]
    }
    replica-postgres = {
      host = "prod-db-replica.example.com"
      port = 5432
      type = "postgresql"
      frequency = 60
    }
    session-redis = {
      host = "prod-redis-session.example.com"
      port = 6379
      type = "redis"
      frequency = 30
    }
    cache-redis = {
      host = "prod-redis-cache.example.com"
      port = 6379
      type = "redis"
    }
    analytics-mongo = {
      host = "prod-mongo.example.com"
      port = 27017
      type = "mongodb"
      frequency = 300  # Less critical, check every 5 minutes
    }
  }

  default_frequency = 60
  alerts_wait       = 30  # Wait 30 seconds before alerting

  escalation_policy = var.database_escalation_policy_id
}

output "database_monitor_ids" {
  value = module.prod_databases.monitor_ids
}
```

### Using Outputs with Incidents

```hcl
module "databases" {
  source = "path/to/modules/database-monitor"

  databases = {
    postgres = { host = "db.example.com", port = 5432, type = "postgresql" }
  }
}

resource "hyperping_incident" "db_maintenance" {
  title    = "Database Maintenance"
  message  = "Scheduled PostgreSQL upgrade"
  severity = "minor"

  # Attach incident to all database monitors
  monitor_uuids = module.databases.monitor_ids_list
}
```

### Testing Configuration

```hcl
module "databases" {
  source = "path/to/modules/database-monitor"

  name_prefix = "TEST"
  paused      = true  # Start paused for testing

  databases = {
    test-db = {
      host = "localhost"
      port = 5432
      type = "postgresql"
    }
  }

  default_regions = ["virginia"]  # Single region for testing
}
```

## Notes

- Port monitors check TCP connectivity only, not database health or queries
- For database-specific health checks, use the `api-health` module with health endpoints
- The `type` field is used for naming and validation, but doesn't affect the port check
- Monitors use `create_before_destroy` lifecycle to prevent downtime during updates
- Port validation ensures values are between 1-65535
- The protocol is set to `port` for TCP port monitoring
