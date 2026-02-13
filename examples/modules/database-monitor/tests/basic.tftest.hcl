# Database Monitor Module - Basic Tests

provider "hyperping" {
  api_key = "test_mock_api_key_for_plan_only"
}

variables {
  name_prefix = "TEST"
  databases = {
    postgres = {
      host = "db.example.com"
      port = 5432
      type = "postgresql"
    }
  }
}

run "creates_monitor_with_defaults" {
  command = plan

  assert {
    condition     = hyperping_monitor.database["postgres"].name == "[TEST] PostgreSQL - postgres"
    error_message = "Monitor name should include prefix, type, and database key"
  }

  assert {
    condition     = hyperping_monitor.database["postgres"].url == "db.example.com"
    error_message = "URL should match input host"
  }

  assert {
    condition     = hyperping_monitor.database["postgres"].port == 5432
    error_message = "Port should match input"
  }

  assert {
    condition     = hyperping_monitor.database["postgres"].protocol == "port"
    error_message = "Protocol should be port"
  }

  assert {
    condition     = hyperping_monitor.database["postgres"].check_frequency == 60
    error_message = "Default frequency should be 60 seconds"
  }
}

run "creates_multiple_databases" {
  command = plan

  variables {
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
      mongo = {
        host = "docs.example.com"
        port = 27017
        type = "mongodb"
      }
    }
  }

  assert {
    condition     = length(hyperping_monitor.database) == 3
    error_message = "Should create 3 monitors"
  }

  assert {
    condition     = hyperping_monitor.database["postgres"].port == 5432
    error_message = "PostgreSQL port should be 5432"
  }

  assert {
    condition     = hyperping_monitor.database["redis"].port == 6379
    error_message = "Redis port should be 6379"
  }

  assert {
    condition     = hyperping_monitor.database["mongo"].port == 27017
    error_message = "MongoDB port should be 27017"
  }
}

run "respects_custom_frequency" {
  command = plan

  variables {
    databases = {
      postgres = {
        host      = "db.example.com"
        port      = 5432
        type      = "postgresql"
        frequency = 30
      }
    }
    default_frequency = 120
  }

  assert {
    condition     = hyperping_monitor.database["postgres"].check_frequency == 30
    error_message = "Should use per-database frequency when specified"
  }
}

run "uses_default_frequency_when_not_specified" {
  command = plan

  variables {
    databases = {
      postgres = {
        host = "db.example.com"
        port = 5432
        type = "postgresql"
      }
    }
    default_frequency = 120
  }

  assert {
    condition     = hyperping_monitor.database["postgres"].check_frequency == 120
    error_message = "Should use default_frequency when per-database frequency not specified"
  }
}

run "supports_region_override" {
  command = plan

  variables {
    databases = {
      postgres = {
        host    = "db.example.com"
        port    = 5432
        type    = "postgresql"
        regions = ["virginia", "london", "tokyo"]
      }
    }
    default_regions = ["frankfurt"]
  }

  assert {
    condition     = length(hyperping_monitor.database["postgres"].regions) == 3
    error_message = "Should use per-database regions when specified"
  }

  assert {
    condition     = contains(hyperping_monitor.database["postgres"].regions, "virginia")
    error_message = "Should include virginia region"
  }
}

run "uses_default_regions_when_not_specified" {
  command = plan

  variables {
    databases = {
      postgres = {
        host = "db.example.com"
        port = 5432
        type = "postgresql"
      }
    }
    default_regions = ["frankfurt", "london"]
  }

  assert {
    condition     = length(hyperping_monitor.database["postgres"].regions) == 2
    error_message = "Should use default_regions when per-database regions not specified"
  }
}
