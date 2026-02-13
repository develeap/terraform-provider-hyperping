# Incident Management Module - Manual Outage Tests

provider "hyperping" {
  api_key = "test_mock_api_key_for_plan_only"
}

variables {
  outage_definitions = {
    test_outage = {
      monitor_uuid = "mon_test123"
      start_date   = "2026-02-15T02:00:00Z"
      end_date     = "2026-02-15T04:00:00Z"
      status_code  = 503
      description  = "Planned downtime"
    }
  }
}

run "creates_outage_with_end_date" {
  command = plan

  assert {
    condition     = hyperping_outage.manual["test_outage"].monitor_uuid == "mon_test123"
    error_message = "Outage should be linked to correct monitor"
  }

  assert {
    condition     = hyperping_outage.manual["test_outage"].start_date == "2026-02-15T02:00:00Z"
    error_message = "Start date should match input"
  }

  assert {
    condition     = hyperping_outage.manual["test_outage"].end_date == "2026-02-15T04:00:00Z"
    error_message = "End date should match input"
  }

  assert {
    condition     = hyperping_outage.manual["test_outage"].status_code == 503
    error_message = "Status code should match input"
  }

  assert {
    condition     = hyperping_outage.manual["test_outage"].description == "Planned downtime"
    error_message = "Description should match input"
  }
}

run "creates_ongoing_outage" {
  command = plan

  variables {
    outage_definitions = {
      ongoing = {
        monitor_uuid = "mon_abc456"
        start_date   = "2026-02-15T10:00:00Z"
        status_code  = 500
        description  = "Investigating server errors"
      }
    }
  }

  assert {
    condition     = hyperping_outage.manual["ongoing"].end_date == null
    error_message = "Ongoing outage should have null end_date"
  }
}

run "creates_multiple_outages" {
  command = plan

  variables {
    outage_definitions = {
      api_outage = {
        monitor_uuid = "mon_api"
        start_date   = "2026-02-16T00:00:00Z"
        end_date     = "2026-02-16T02:00:00Z"
        status_code  = 503
        description  = "API maintenance"
      }
      db_outage = {
        monitor_uuid = "mon_db"
        start_date   = "2026-02-17T00:00:00Z"
        end_date     = "2026-02-17T04:00:00Z"
        status_code  = 500
        description  = "Database upgrade"
      }
    }
  }

  assert {
    condition     = length(hyperping_outage.manual) == 2
    error_message = "Should create 2 outages"
  }
}

run "respects_create_outages_flag" {
  command = plan

  variables {
    create_outages = false

    outage_definitions = {
      should_not_create = {
        monitor_uuid = "mon_test"
        start_date   = "2026-02-15T00:00:00Z"
        status_code  = 503
        description  = "Test"
      }
    }
  }

  assert {
    condition     = length(hyperping_outage.manual) == 0
    error_message = "Should not create outages when create_outages is false"
  }
}

run "outputs_outage_count" {
  command = plan

  variables {
    outage_definitions = {
      outage1 = {
        monitor_uuid = "mon_1"
        start_date   = "2026-02-15T00:00:00Z"
        status_code  = 503
        description  = "Outage 1"
      }
      outage2 = {
        monitor_uuid = "mon_2"
        start_date   = "2026-02-16T00:00:00Z"
        status_code  = 500
        description  = "Outage 2"
      }
    }
  }

  assert {
    condition     = output.outage_count == 2
    error_message = "Should output correct outage count"
  }

  assert {
    condition     = length(output.outage_ids) == 2
    error_message = "Should output outage IDs map"
  }
}
