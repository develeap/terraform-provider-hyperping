// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestOutageAttrTypes(t *testing.T) {
	attrTypes := OutageAttrTypes()

	expectedKeys := []string{
		"count",
		"total_downtime",
		"total_downtime_formatted",
		"longest_outage",
		"longest_outage_formatted",
	}

	if len(attrTypes) != len(expectedKeys) {
		t.Errorf("expected %d attributes, got %d", len(expectedKeys), len(attrTypes))
	}

	for _, key := range expectedKeys {
		if _, ok := attrTypes[key]; !ok {
			t.Errorf("missing expected attribute: %s", key)
		}
	}

	// Verify specific types
	if attrTypes["count"] != types.Int64Type {
		t.Error("count should be Int64Type")
	}
	if attrTypes["total_downtime"] != types.Int64Type {
		t.Error("total_downtime should be Int64Type")
	}
	if attrTypes["total_downtime_formatted"] != types.StringType {
		t.Error("total_downtime_formatted should be StringType")
	}
	if attrTypes["longest_outage"] != types.Int64Type {
		t.Error("longest_outage should be Int64Type")
	}
	if attrTypes["longest_outage_formatted"] != types.StringType {
		t.Error("longest_outage_formatted should be StringType")
	}
}

func TestNewMonitorReportDataSource(t *testing.T) {
	ds := NewMonitorReportDataSource()
	if ds == nil {
		t.Fatal("NewMonitorReportDataSource returned nil")
	}
	if _, ok := ds.(*MonitorReportDataSource); !ok {
		t.Errorf("expected *MonitorReportDataSource, got %T", ds)
	}
}
