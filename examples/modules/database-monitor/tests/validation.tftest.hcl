# Database Monitor Module - Validation Tests

provider "hyperping" {
  api_key = "test_mock_api_key_for_plan_only"
}

variables {
  name_prefix = "TEST"
}

run "validates_empty_databases" {
  command = plan

  variables {
    databases = {}
  }

  expect_failures = [
    var.databases,
  ]
}

run "validates_invalid_database_type" {
  command = plan

  variables {
    databases = {
      invalid = {
        host = "db.example.com"
        port = 5432
        type = "invalid_database"
      }
    }
  }

  expect_failures = [
    var.databases,
  ]
}

run "validates_port_range_too_low" {
  command = plan

  variables {
    databases = {
      postgres = {
        host = "db.example.com"
        port = 0
        type = "postgresql"
      }
    }
  }

  expect_failures = [
    var.databases,
  ]
}

run "validates_port_range_too_high" {
  command = plan

  variables {
    databases = {
      postgres = {
        host = "db.example.com"
        port = 65536
        type = "postgresql"
      }
    }
  }

  expect_failures = [
    var.databases,
  ]
}

run "validates_invalid_region" {
  command = plan

  variables {
    databases = {
      postgres = {
        host = "db.example.com"
        port = 5432
        type = "postgresql"
      }
    }
    default_regions = ["invalid_region"]
  }

  expect_failures = [
    var.default_regions,
  ]
}

run "validates_invalid_frequency" {
  command = plan

  variables {
    databases = {
      postgres = {
        host = "db.example.com"
        port = 5432
        type = "postgresql"
      }
    }
    default_frequency = 99
  }

  expect_failures = [
    var.default_frequency,
  ]
}

run "validates_name_prefix_special_chars" {
  command = plan

  variables {
    name_prefix = "PROD@#$%"
    databases = {
      postgres = {
        host = "db.example.com"
        port = 5432
        type = "postgresql"
      }
    }
  }

  expect_failures = [
    var.name_prefix,
  ]
}

run "accepts_valid_configuration" {
  command = plan

  variables {
    name_prefix = "PROD-US-EAST-1"
    databases = {
      postgres = {
        host      = "db.example.com"
        port      = 5432
        type      = "postgresql"
        frequency = 60
        timeout   = 10
        regions   = ["virginia", "london"]
      }
    }
    default_regions   = ["virginia"]
    default_frequency = 60
  }

  assert {
    condition     = hyperping_monitor.database["postgres"].name == "[PROD-US-EAST-1] PostgreSQL - postgres"
    error_message = "Should accept valid configuration with hyphenated prefix"
  }
}

run "accepts_underscore_in_prefix" {
  command = plan

  variables {
    name_prefix = "PROD_EAST_1"
    databases = {
      postgres = {
        host = "db.example.com"
        port = 5432
        type = "postgresql"
      }
    }
  }

  assert {
    condition     = hyperping_monitor.database["postgres"].name == "[PROD_EAST_1] PostgreSQL - postgres"
    error_message = "Should accept underscores in name prefix"
  }
}

run "accepts_edge_ports" {
  command = plan

  variables {
    databases = {
      min_port = {
        host = "db1.example.com"
        port = 1
        type = "postgresql"
      }
      max_port = {
        host = "db2.example.com"
        port = 65535
        type = "postgresql"
      }
    }
  }

  assert {
    condition     = hyperping_monitor.database["min_port"].port == 1
    error_message = "Should accept port 1"
  }

  assert {
    condition     = hyperping_monitor.database["max_port"].port == 65535
    error_message = "Should accept port 65535"
  }
}
