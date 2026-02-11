// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func init() {
	// Register all sweepers for cleaning up test resources
	resource.AddTestSweepers("hyperping_monitor", &resource.Sweeper{
		Name: "hyperping_monitor",
		F:    sweepMonitors,
	})

	resource.AddTestSweepers("hyperping_incident", &resource.Sweeper{
		Name: "hyperping_incident",
		F:    sweepIncidents,
	})

	resource.AddTestSweepers("hyperping_maintenance", &resource.Sweeper{
		Name: "hyperping_maintenance",
		F:    sweepMaintenance,
	})

	resource.AddTestSweepers("hyperping_healthcheck", &resource.Sweeper{
		Name: "hyperping_healthcheck",
		F:    sweepHealthchecks,
	})

	resource.AddTestSweepers("hyperping_statuspage", &resource.Sweeper{
		Name: "hyperping_statuspage",
		F:    sweepStatusPages,
	})

	resource.AddTestSweepers("hyperping_outage", &resource.Sweeper{
		Name: "hyperping_outage",
		F:    sweepOutages,
	})
}

// TestMain enables sweep functionality
func TestMain(m *testing.M) {
	resource.TestMain(m)
}
