# Incident Management Module - Maintenance Window Tests

provider "hyperping" {
  api_key = "test_mock_api_key_for_plan_only"
}

variables {
  statuspage_id = "sp_test123"

  maintenance_windows = {
    test_maintenance = {
      title      = "Test Maintenance"
      text       = "Routine maintenance"
      start_date = "2026-02-20T02:00:00.000Z"
      end_date   = "2026-02-20T04:00:00.000Z"
    }
  }
}

run "creates_maintenance_with_defaults" {
  command = plan

  assert {
    condition     = hyperping_maintenance.window["test_maintenance"].name == "test_maintenance"
    error_message = "Maintenance name should match key"
  }

  assert {
    condition     = hyperping_maintenance.window["test_maintenance"].title == "Test Maintenance"
    error_message = "Maintenance title should match input"
  }

  assert {
    condition     = hyperping_maintenance.window["test_maintenance"].start_date == "2026-02-20T02:00:00.000Z"
    error_message = "Start date should match input"
  }

  assert {
    condition     = hyperping_maintenance.window["test_maintenance"].end_date == "2026-02-20T04:00:00.000Z"
    error_message = "End date should match input"
  }

  assert {
    condition     = hyperping_maintenance.window["test_maintenance"].notification_option == "scheduled"
    error_message = "Default notification option should be scheduled"
  }

  assert {
    condition     = hyperping_maintenance.window["test_maintenance"].notification_minutes == 60
    error_message = "Default notification minutes should be 60"
  }
}

run "creates_maintenance_with_monitors" {
  command = plan

  variables {
    maintenance_windows = {
      db_maintenance = {
        title      = "Database Maintenance"
        text       = "DB upgrade"
        start_date = "2026-02-25T01:00:00.000Z"
        end_date   = "2026-02-25T03:00:00.000Z"
        monitors   = ["mon_123", "mon_456"]
      }
    }
  }

  assert {
    condition     = length(hyperping_maintenance.window["db_maintenance"].monitors) == 2
    error_message = "Should associate with 2 monitors"
  }

  assert {
    condition     = contains(hyperping_maintenance.window["db_maintenance"].monitors, "mon_123")
    error_message = "Should include first monitor"
  }
}

run "creates_maintenance_with_immediate_notification" {
  command = plan

  variables {
    maintenance_windows = {
      urgent = {
        title               = "Emergency Maintenance"
        text                = "Urgent fix required"
        start_date          = "2026-02-15T00:00:00.000Z"
        end_date            = "2026-02-15T06:00:00.000Z"
        notification_option = "immediate"
      }
    }
  }

  assert {
    condition     = hyperping_maintenance.window["urgent"].notification_option == "immediate"
    error_message = "Notification option should be immediate"
  }

  assert {
    condition     = hyperping_maintenance.window["urgent"].notification_minutes == null
    error_message = "Notification minutes should be null for immediate"
  }
}

run "creates_maintenance_with_custom_notification_time" {
  command = plan

  variables {
    maintenance_windows = {
      planned = {
        title                = "Planned Upgrade"
        text                 = "System upgrade"
        start_date           = "2026-03-01T02:00:00.000Z"
        end_date             = "2026-03-01T04:00:00.000Z"
        notification_option  = "scheduled"
        notification_minutes = 120
      }
    }
  }

  assert {
    condition     = hyperping_maintenance.window["planned"].notification_minutes == 120
    error_message = "Should use custom notification minutes"
  }
}

run "outputs_maintenance_count" {
  command = plan

  variables {
    maintenance_windows = {
      maint1 = {
        title      = "Maintenance 1"
        text       = "Text"
        start_date = "2026-02-20T02:00:00.000Z"
        end_date   = "2026-02-20T04:00:00.000Z"
      }
      maint2 = {
        title      = "Maintenance 2"
        text       = "Text"
        start_date = "2026-02-21T02:00:00.000Z"
        end_date   = "2026-02-21T04:00:00.000Z"
      }
    }
  }

  assert {
    condition     = output.maintenance_count == 2
    error_message = "Should output correct maintenance count"
  }

  assert {
    condition     = length(output.maintenance_ids) == 2
    error_message = "Should output maintenance IDs map"
  }
}
