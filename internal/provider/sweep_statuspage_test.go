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

// sweepStatusPages deletes all test status pages (those with name prefixed with "tf-acc-test-")
func sweepStatusPages(region string) error {
	ctx := context.Background()
	c, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	// List all status pages (pagination not needed for sweepers - we want all)
	resp, err := c.ListStatusPages(ctx, nil, nil)
	if err != nil {
		return fmt.Errorf("error listing status pages: %w", err)
	}

	for _, sp := range resp.StatusPages {
		// Only delete test resources (name prefixed with "tf-acc-test-")
		if strings.HasPrefix(sp.Name, "tf-acc-test-") {
			log.Printf("[INFO] Deleting status page: %s (UUID: %s)", sp.Name, sp.UUID)
			if err := c.DeleteStatusPage(ctx, sp.UUID); err != nil {
				if client.IsNotFound(err) {
					log.Printf("[WARN] Status page %s already deleted", sp.UUID)
					continue
				}
				log.Printf("[ERROR] Failed to delete status page %s: %v", sp.UUID, err)
				// Continue with other deletions even if one fails
			}
		}
	}

	return nil
}
