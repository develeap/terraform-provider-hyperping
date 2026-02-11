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

// sweepHealthchecks deletes all test healthchecks (those with name prefixed with "tf-acc-test-")
func sweepHealthchecks(_ string) error {
	ctx := context.Background()
	c, err := sharedClientForRegion("")
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	healthchecks, err := c.ListHealthchecks(ctx)
	if err != nil {
		return fmt.Errorf("error listing healthchecks: %w", err)
	}

	for _, hc := range healthchecks {
		// Only delete test resources (name prefixed with "tf-acc-test-")
		if strings.HasPrefix(hc.Name, "tf-acc-test-") {
			log.Printf("[INFO] Deleting healthcheck: %s (UUID: %s)", hc.Name, hc.UUID)
			if err := c.DeleteHealthcheck(ctx, hc.UUID); err != nil {
				if client.IsNotFound(err) {
					log.Printf("[WARN] Healthcheck %s already deleted", hc.UUID)
					continue
				}
				log.Printf("[ERROR] Failed to delete healthcheck %s: %v", hc.UUID, err)
				// Continue with other deletions even if one fails
			}
		}
	}

	return nil
}
