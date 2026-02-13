# Incident Management Module - Validation Tests

provider "hyperping" {
  api_key = "test_mock_api_key_for_plan_only"
}

# Test valid severity values
run "accepts_valid_severity_values" {
  command = plan

  variables {
    incident_templates = {
      minor_inc = {
        title    = "Minor Issue"
        text     = "Minor problem"
        severity = "minor"
      }
      major_inc = {
        title    = "Major Issue"
        text     = "Major problem"
        severity = "major"
      }
      critical_inc = {
        title    = "Critical Issue"
        text     = "Critical problem"
        severity = "critical"
      }
    }
  }

  assert {
    condition     = length(hyperping_incident.template) == 3
    error_message = "Should accept all valid severity values"
  }
}

# Test valid notification options
run "accepts_valid_notification_options" {
  command = plan

  variables {
    maintenance_windows = {
      immediate = {
        title               = "Immediate"
        text                = "Test"
        start_date          = "2026-02-20T02:00:00.000Z"
        end_date            = "2026-02-20T04:00:00.000Z"
        notification_option = "immediate"
      }
      scheduled = {
        title                = "Scheduled"
        text                 = "Test"
        start_date           = "2026-02-21T02:00:00.000Z"
        end_date             = "2026-02-21T04:00:00.000Z"
        notification_option  = "scheduled"
        notification_minutes = 60
      }
      none = {
        title               = "None"
        text                = "Test"
        start_date          = "2026-02-22T02:00:00.000Z"
        end_date            = "2026-02-22T04:00:00.000Z"
        notification_option = "none"
      }
    }
  }

  assert {
    condition     = length(hyperping_maintenance.window) == 3
    error_message = "Should accept all valid notification options"
  }
}

# Test valid status codes
run "accepts_valid_status_codes" {
  command = plan

  variables {
    outage_definitions = {
      code_400 = {
        monitor_uuid = "mon_1"
        start_date   = "2026-02-15T00:00:00Z"
        status_code  = 400
        description  = "Bad request"
      }
      code_500 = {
        monitor_uuid = "mon_2"
        start_date   = "2026-02-15T00:00:00Z"
        status_code  = 500
        description  = "Server error"
      }
      code_503 = {
        monitor_uuid = "mon_3"
        start_date   = "2026-02-15T00:00:00Z"
        status_code  = 503
        description  = "Service unavailable"
      }
    }
  }

  assert {
    condition     = length(hyperping_outage.manual) == 3
    error_message = "Should accept valid HTTP status codes"
  }
}

# Test custom type override
run "accepts_custom_incident_type" {
  command = plan

  variables {
    incident_templates = {
      custom_outage = {
        title    = "Custom Outage"
        text     = "This is a custom outage"
        severity = "minor"
        type     = "outage"
      }
    }
  }

  assert {
    condition     = hyperping_incident.template["custom_outage"].type == "outage"
    error_message = "Should allow custom type override"
  }
}

# Test empty configurations
run "handles_empty_configurations" {
  command = plan

  variables {
    incident_templates  = {}
    maintenance_windows = {}
    outage_definitions  = {}
  }

  assert {
    condition     = output.incident_count == 0
    error_message = "Should handle empty incident templates"
  }

  assert {
    condition     = output.maintenance_count == 0
    error_message = "Should handle empty maintenance windows"
  }

  assert {
    condition     = output.outage_count == 0
    error_message = "Should handle empty outage definitions"
  }
}

# Test output structure
run "outputs_have_correct_structure" {
  command = plan

  variables {
    incident_templates = {
      test = {
        title    = "Test"
        text     = "Test"
        severity = "major"
      }
    }

    maintenance_windows = {
      test = {
        title      = "Test"
        text       = "Test"
        start_date = "2026-02-20T02:00:00.000Z"
        end_date   = "2026-02-20T04:00:00.000Z"
      }
    }

    outage_definitions = {
      test = {
        monitor_uuid = "mon_test"
        start_date   = "2026-02-15T00:00:00Z"
        status_code  = 503
        description  = "Test"
      }
    }
  }

  assert {
    condition     = can(output.incident_ids)
    error_message = "Should output incident_ids"
  }

  assert {
    condition     = can(output.incident_ids_list)
    error_message = "Should output incident_ids_list"
  }

  assert {
    condition     = can(output.incidents)
    error_message = "Should output incidents"
  }

  assert {
    condition     = can(output.maintenance_ids)
    error_message = "Should output maintenance_ids"
  }

  assert {
    condition     = can(output.maintenance_windows)
    error_message = "Should output maintenance_windows"
  }

  assert {
    condition     = can(output.outage_ids)
    error_message = "Should output outage_ids"
  }

  assert {
    condition     = can(output.summary)
    error_message = "Should output summary"
  }
}
