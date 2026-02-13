# Database Monitor Module - Main
#
# Creates TCP port monitors for database connectivity checking.
#
# Usage:
#   module "database_monitors" {
#     source = "path/to/modules/database-monitor"
#
#     databases = {
#       "postgres" = { host = "db.example.com", port = 5432, type = "postgresql" }
#       "redis"    = { host = "cache.example.com", port = 6379, type = "redis" }
#     }
#
#     name_prefix     = "PROD"
#     default_regions = ["virginia", "london"]
#   }

locals {
  # Database type display names
  type_names = {
    postgresql    = "PostgreSQL"
    postgres      = "PostgreSQL"
    mysql         = "MySQL"
    redis         = "Redis"
    mongodb       = "MongoDB"
    mongo         = "MongoDB"
    mariadb       = "MariaDB"
    cassandra     = "Cassandra"
    elasticsearch = "Elasticsearch"
    memcached     = "Memcached"
    rabbitmq      = "RabbitMQ"
  }

  # Default ports for common databases (for reference/validation)
  default_ports = {
    postgresql    = 5432
    postgres      = 5432
    mysql         = 3306
    redis         = 6379
    mongodb       = 27017
    mongo         = 27017
    mariadb       = 3306
    cassandra     = 9042
    elasticsearch = 9200
    memcached     = 11211
    rabbitmq      = 5672
  }

  # Build the monitor name
  name_format = var.name_format != "" ? var.name_format : (
    var.name_prefix != "" ? "[${upper(var.name_prefix)}] %s" : "%s"
  )

  # Build full names with optional type prefix
  monitor_names = {
    for k, v in var.databases : k => var.include_type_in_name ? format(
      local.name_format,
      "${local.type_names[lower(v.type)]} - ${k}"
    ) : format(local.name_format, k)
  }
}

resource "hyperping_monitor" "database" {
  for_each = var.databases

  name     = local.monitor_names[each.key]
  url      = each.value.host
  protocol = "port"
  port     = each.value.port

  # Monitoring configuration
  check_frequency = coalesce(each.value.frequency, var.default_frequency)
  regions         = coalesce(each.value.regions, var.default_regions)
  paused          = coalesce(each.value.paused, var.paused)
  alerts_wait     = var.alerts_wait

  # Optional: escalation policy
  escalation_policy = var.escalation_policy

  lifecycle {
    create_before_destroy = true
  }
}
