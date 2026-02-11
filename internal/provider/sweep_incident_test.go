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

// sweepIncidents deletes all test incidents (those with title prefixed with "tf-acc-test-")
func sweepIncidents(region string) error {
	ctx := context.Background()
	c, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	incidents, err := c.ListIncidents(ctx)
	if err != nil {
		return fmt.Errorf("error listing incidents: %w", err)
	}

	for _, incident := range incidents {
		// Only delete test resources (title prefixed with "tf-acc-test-")
		if strings.HasPrefix(incident.Title.En, "tf-acc-test-") {
			log.Printf("[INFO] Deleting incident: %s (UUID: %s)", incident.Title.En, incident.UUID)
			if err := c.DeleteIncident(ctx, incident.UUID); err != nil {
				if client.IsNotFound(err) {
					log.Printf("[WARN] Incident %s already deleted", incident.UUID)
					continue
				}
				log.Printf("[ERROR] Failed to delete incident %s: %v", incident.UUID, err)
				// Continue with other deletions even if one fails
			}
		}
	}

	return nil
}
