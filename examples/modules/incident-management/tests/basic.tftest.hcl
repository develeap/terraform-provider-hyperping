# Incident Management Module - Basic Tests

provider "hyperping" {
  api_key = "test_mock_api_key_for_plan_only"
}

variables {
  statuspage_id = "sp_test123"

  incident_templates = {
    test_incident = {
      title    = "Test Incident"
      text     = "This is a test incident"
      severity = "major"
    }
  }
}

run "creates_incident_with_defaults" {
  command = plan

  assert {
    condition     = hyperping_incident.template["test_incident"].title == "Test Incident"
    error_message = "Incident title should match input"
  }

  assert {
    condition     = hyperping_incident.template["test_incident"].text == "This is a test incident"
    error_message = "Incident text should match input"
  }

  assert {
    condition     = hyperping_incident.template["test_incident"].type == "incident"
    error_message = "Major severity should map to incident type"
  }

  assert {
    condition     = contains(hyperping_incident.template["test_incident"].status_pages, "sp_test123")
    error_message = "Incident should be linked to status page"
  }
}

run "creates_multiple_incidents" {
  command = plan

  variables {
    incident_templates = {
      api_issue = {
        title    = "API Performance Issue"
        text     = "API is slow"
        severity = "major"
      }
      db_outage = {
        title    = "Database Outage"
        text     = "Database is down"
        severity = "critical"
      }
      minor_bug = {
        title    = "Minor Bug"
        text     = "Small issue found"
        severity = "minor"
      }
    }
  }

  assert {
    condition     = length(hyperping_incident.template) == 3
    error_message = "Should create 3 incidents"
  }

  assert {
    condition     = hyperping_incident.template["db_outage"].type == "outage"
    error_message = "Critical severity should map to outage type"
  }

  assert {
    condition     = hyperping_incident.template["minor_bug"].type == "incident"
    error_message = "Minor severity should map to incident type"
  }
}

run "severity_to_type_mapping" {
  command = plan

  variables {
    incident_templates = {
      minor_inc = {
        title    = "Minor"
        text     = "Text"
        severity = "minor"
      }
      major_inc = {
        title    = "Major"
        text     = "Text"
        severity = "major"
      }
      critical_inc = {
        title    = "Critical"
        text     = "Text"
        severity = "critical"
      }
    }
  }

  assert {
    condition     = hyperping_incident.template["minor_inc"].type == "incident"
    error_message = "Minor severity should map to incident type"
  }

  assert {
    condition     = hyperping_incident.template["major_inc"].type == "incident"
    error_message = "Major severity should map to incident type"
  }

  assert {
    condition     = hyperping_incident.template["critical_inc"].type == "outage"
    error_message = "Critical severity should map to outage type"
  }
}

run "outputs_correct_count" {
  command = plan

  variables {
    incident_templates = {
      inc1 = {
        title    = "Incident 1"
        text     = "Text 1"
        severity = "major"
      }
      inc2 = {
        title    = "Incident 2"
        text     = "Text 2"
        severity = "minor"
      }
    }
  }

  assert {
    condition     = output.incident_count == 2
    error_message = "Should output correct incident count"
  }

  assert {
    condition     = length(output.incident_ids) == 2
    error_message = "Should output incident IDs map"
  }
}
