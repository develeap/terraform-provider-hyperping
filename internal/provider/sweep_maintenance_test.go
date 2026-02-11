// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// sweepMaintenance deletes all test maintenance windows (those with name prefixed with "tf-acc-test-")
func sweepMaintenance(region string) error {
	ctx := context.Background()
	c, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	maintenanceWindows, err := c.ListMaintenance(ctx)
	if err != nil {
		return fmt.Errorf("error listing maintenance windows: %w", err)
	}

	for _, maint := range maintenanceWindows {
		// Only delete test resources (name prefixed with "tf-acc-test-")
		if strings.HasPrefix(maint.Name, "tf-acc-test-") {
			log.Printf("[INFO] Deleting maintenance window: %s (UUID: %s)", maint.Name, maint.UUID)
			if err := c.DeleteMaintenance(ctx, maint.UUID); err != nil {
				if client.IsNotFound(err) {
					log.Printf("[WARN] Maintenance window %s already deleted", maint.UUID)
					continue
				}
				log.Printf("[ERROR] Failed to delete maintenance window %s: %v", maint.UUID, err)
				// Continue with other deletions even if one fails
			}
		}
	}

	return nil
}
