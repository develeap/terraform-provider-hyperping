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

// sweepOutages deletes all test outages (those with title prefixed with "tf-acc-test-")
func sweepOutages(region string) error {
	ctx := context.Background()
	c, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	outages, err := c.ListOutages(ctx)
	if err != nil {
		return fmt.Errorf("error listing outages: %w", err)
	}

	for _, outage := range outages {
		// Only delete test resources (monitor name prefixed with "tf-acc-test-")
		// Outages are associated with monitors, so we filter by monitor name
		if strings.HasPrefix(outage.Monitor.Name, "tf-acc-test-") {
			log.Printf("[INFO] Deleting outage: %s (UUID: %s, Monitor: %s)", outage.Description, outage.UUID, outage.Monitor.Name)
			if err := c.DeleteOutage(ctx, outage.UUID); err != nil {
				if client.IsNotFound(err) {
					log.Printf("[WARN] Outage %s already deleted", outage.UUID)
					continue
				}
				log.Printf("[ERROR] Failed to delete outage %s: %v", outage.UUID, err)
				// Continue with other deletions even if one fails
			}
		}
	}

	return nil
}
