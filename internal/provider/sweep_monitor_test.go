// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// sweepMonitors deletes all test monitors (those prefixed with "tf-acc-test-")
func sweepMonitors(region string) error {
	ctx := context.Background()
	c, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	monitors, err := c.ListMonitors(ctx)
	if err != nil {
		return fmt.Errorf("error listing monitors: %w", err)
	}

	for _, monitor := range monitors {
		// Only delete test resources (prefix with "tf-acc-test-")
		if strings.HasPrefix(monitor.Name, "tf-acc-test-") {
			log.Printf("[INFO] Deleting monitor: %s (UUID: %s)", monitor.Name, monitor.UUID)
			if err := c.DeleteMonitor(ctx, monitor.UUID); err != nil {
				if client.IsNotFound(err) {
					log.Printf("[WARN] Monitor %s already deleted", monitor.UUID)
					continue
				}
				log.Printf("[ERROR] Failed to delete monitor %s: %v", monitor.UUID, err)
				// Continue with other deletions even if one fails
			}
		}
	}

	return nil
}

// sharedClientForRegion creates a Hyperping client for sweeper operations
func sharedClientForRegion(region string) (*client.Client, error) {
	apiKey := os.Getenv("HYPERPING_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("HYPERPING_API_KEY must be set for sweepers")
	}

	baseURL := os.Getenv("HYPERPING_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.hyperping.io"
	}

	return client.NewClient(apiKey, client.WithBaseURL(baseURL)), nil
}
