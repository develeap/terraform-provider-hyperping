# Database Monitor Module - Custom Configuration Tests

provider "hyperping" {
  api_key = "test_mock_api_key_for_plan_only"
}

run "custom_name_format" {
  command = plan

  variables {
    name_format = "DB Monitor: %s"
    databases = {
      maindb = {
        host = "db.example.com"
        port = 5432
        type = "postgresql"
      }
    }
    include_type_in_name = false
  }

  assert {
    condition     = hyperping_monitor.database["maindb"].name == "DB Monitor: maindb"
    error_message = "Should use custom name format without type prefix"
  }
}

run "custom_name_format_with_type" {
  command = plan

  variables {
    name_format = "PROD: %s"
    databases = {
      cache = {
        host = "redis.example.com"
        port = 6379
        type = "redis"
      }
    }
    include_type_in_name = true
  }

  assert {
    condition     = hyperping_monitor.database["cache"].name == "PROD: Redis - cache"
    error_message = "Should use custom format with type prefix"
  }
}

run "without_type_in_name" {
  command = plan

  variables {
    name_prefix = "PROD"
    databases = {
      postgres = {
        host = "db.example.com"
        port = 5432
        type = "postgresql"
      }
    }
    include_type_in_name = false
  }

  assert {
    condition     = hyperping_monitor.database["postgres"].name == "[PROD] postgres"
    error_message = "Should exclude type from name when include_type_in_name is false"
  }
}

run "paused_state" {
  command = plan

  variables {
    databases = {
      postgres = {
        host = "db.example.com"
        port = 5432
        type = "postgresql"
      }
    }
    paused = true
  }

  assert {
    condition     = hyperping_monitor.database["postgres"].paused == true
    error_message = "Should create monitor in paused state"
  }
}

run "per_database_paused_override" {
  command = plan

  variables {
    databases = {
      active = {
        host   = "db1.example.com"
        port   = 5432
        type   = "postgresql"
        paused = false
      }
      maintenance = {
        host   = "db2.example.com"
        port   = 5432
        type   = "postgresql"
        paused = true
      }
    }
    paused = false
  }

  assert {
    condition     = hyperping_monitor.database["active"].paused == false
    error_message = "Active database should not be paused"
  }

  assert {
    condition     = hyperping_monitor.database["maintenance"].paused == true
    error_message = "Maintenance database should be paused"
  }
}

run "database_type_aliases" {
  command = plan

  variables {
    databases = {
      pg1 = {
        host = "db1.example.com"
        port = 5432
        type = "postgresql"
      }
      pg2 = {
        host = "db2.example.com"
        port = 5432
        type = "postgres"
      }
      mongo1 = {
        host = "db3.example.com"
        port = 27017
        type = "mongodb"
      }
      mongo2 = {
        host = "db4.example.com"
        port = 27017
        type = "mongo"
      }
    }
  }

  assert {
    condition = (
      hyperping_monitor.database["pg1"].name == "[TEST] PostgreSQL - pg1" &&
      hyperping_monitor.database["pg2"].name == "[TEST] PostgreSQL - pg2"
    )
    error_message = "Both postgresql and postgres aliases should result in PostgreSQL name"
  }

  assert {
    condition = (
      hyperping_monitor.database["mongo1"].name == "[TEST] MongoDB - mongo1" &&
      hyperping_monitor.database["mongo2"].name == "[TEST] MongoDB - mongo2"
    )
    error_message = "Both mongodb and mongo aliases should result in MongoDB name"
  }
}

run "mixed_database_types" {
  command = plan

  variables {
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
      mongo = {
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
  }

  assert {
    condition     = length(hyperping_monitor.database) == 5
    error_message = "Should create monitors for all database types"
  }

  assert {
    condition     = hyperping_monitor.database["postgres"].name == "[TEST] PostgreSQL - postgres"
    error_message = "PostgreSQL name should be correct"
  }

  assert {
    condition     = hyperping_monitor.database["mysql"].name == "[TEST] MySQL - mysql"
    error_message = "MySQL name should be correct"
  }

  assert {
    condition     = hyperping_monitor.database["redis"].name == "[TEST] Redis - redis"
    error_message = "Redis name should be correct"
  }

  assert {
    condition     = hyperping_monitor.database["mongo"].name == "[TEST] MongoDB - mongo"
    error_message = "MongoDB name should be correct"
  }

  assert {
    condition     = hyperping_monitor.database["cassandra"].name == "[TEST] Cassandra - cassandra"
    error_message = "Cassandra name should be correct"
  }
}

variables {
  name_prefix = "TEST"
}
