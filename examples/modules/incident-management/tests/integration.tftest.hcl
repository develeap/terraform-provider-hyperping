# Incident Management Module - Integration Tests

provider "hyperping" {
  api_key = "test_mock_api_key_for_plan_only"
}

variables {
  statuspage_id = "sp_integration_test"

  incident_templates = {
    api_degradation = {
      title               = "API Performance Degraded"
      text                = "Investigating slow API response times"
      severity            = "major"
      affected_components = ["comp_api", "comp_db"]
    }
  }

  maintenance_windows = {
    routine_maintenance = {
      title                = "Routine Maintenance"
      text                 = "Weekly system updates"
      start_date           = "2026-02-20T02:00:00.000Z"
      end_date             = "2026-02-20T04:00:00.000Z"
      monitors             = ["mon_api", "mon_web"]
      notification_option  = "scheduled"
      notification_minutes = 120
    }
  }

  outage_definitions = {
    planned_outage = {
      monitor_uuid = "mon_api"
      start_date   = "2026-02-20T02:00:00Z"
      end_date     = "2026-02-20T04:00:00Z"
      status_code  = 503
      description  = "Planned API downtime"
    }
  }
}

run "creates_all_resource_types" {
  command = plan

  assert {
    condition     = length(hyperping_incident.template) == 1
    error_message = "Should create incident"
  }

  assert {
    condition     = length(hyperping_maintenance.window) == 1
    error_message = "Should create maintenance window"
  }

  assert {
    condition     = length(hyperping_outage.manual) == 1
    error_message = "Should create outage"
  }
}

run "links_resources_to_statuspage" {
  command = plan

  assert {
    condition     = contains(hyperping_incident.template["api_degradation"].status_pages, "sp_integration_test")
    error_message = "Incident should be linked to status page"
  }

  assert {
    condition     = contains(hyperping_maintenance.window["routine_maintenance"].status_pages, "sp_integration_test")
    error_message = "Maintenance should be linked to status page"
  }
}

run "associates_incidents_with_components" {
  command = plan

  assert {
    condition     = length(hyperping_incident.template["api_degradation"].affected_components) == 2
    error_message = "Incident should have 2 affected components"
  }

  assert {
    condition     = contains(hyperping_incident.template["api_degradation"].affected_components, "comp_api")
    error_message = "Should include API component"
  }

  assert {
    condition     = contains(hyperping_incident.template["api_degradation"].affected_components, "comp_db")
    error_message = "Should include DB component"
  }
}

run "associates_maintenance_with_monitors" {
  command = plan

  assert {
    condition     = length(hyperping_maintenance.window["routine_maintenance"].monitors) == 2
    error_message = "Maintenance should have 2 monitors"
  }

  assert {
    condition     = contains(hyperping_maintenance.window["routine_maintenance"].monitors, "mon_api")
    error_message = "Should include API monitor"
  }
}

run "summary_output_shows_all_resources" {
  command = plan

  assert {
    condition     = output.summary.incidents == 1
    error_message = "Summary should show 1 incident"
  }

  assert {
    condition     = output.summary.maintenance == 1
    error_message = "Summary should show 1 maintenance window"
  }

  assert {
    condition     = output.summary.outages == 1
    error_message = "Summary should show 1 outage"
  }

  assert {
    condition     = output.summary.statuspage_linked == true
    error_message = "Summary should indicate status page is linked"
  }
}

run "conditional_creation_works" {
  command = plan

  variables {
    create_incidents   = false
    create_maintenance = false
    create_outages     = false
  }

  assert {
    condition     = length(hyperping_incident.template) == 0
    error_message = "Should not create incidents when disabled"
  }

  assert {
    condition     = length(hyperping_maintenance.window) == 0
    error_message = "Should not create maintenance when disabled"
  }

  assert {
    condition     = length(hyperping_outage.manual) == 0
    error_message = "Should not create outages when disabled"
  }

  assert {
    condition     = output.summary.incidents == 0
    error_message = "Summary should show 0 incidents"
  }
}

run "works_without_statuspage" {
  command = plan

  variables {
    statuspage_id = null

    incident_templates = {
      test = {
        title    = "Test Incident"
        text     = "Test"
        severity = "major"
      }
    }
  }

  assert {
    condition     = length(hyperping_incident.template["test"].status_pages) == 0
    error_message = "Should create incident without status page"
  }

  assert {
    condition     = output.summary.statuspage_linked == false
    error_message = "Summary should indicate no status page link"
  }
}
