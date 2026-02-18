// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// Standard error patterns for consistent user experience across the provider.
// All error messages follow a consistent format with helpful troubleshooting context.

// CRUD Operation Errors

// newCreateError creates a standardized error for Create operations
func newCreateError(resourceType string, err error) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		fmt.Sprintf("Failed to Create %s", resourceType),
		fmt.Sprintf("Unable to create %s, got error: %s\n\n"+
			"Troubleshooting:\n"+
			"- Verify your API key has create permissions\n"+
			"- Check that all required fields are provided\n"+
			"- Review the Hyperping dashboard: https://app.hyperping.io",
			resourceType, err),
	)
}

// newReadError creates a standardized error for Read operations
func newReadError(resourceType, resourceID string, err error) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		fmt.Sprintf("Failed to Read %s", resourceType),
		fmt.Sprintf("Unable to read %s (ID: %s), got error: %s\n\n"+
			"Troubleshooting:\n"+
			"- Verify the resource still exists in Hyperping\n"+
			"- Check your API key has read permissions\n"+
			"- Check network connectivity\n"+
			"- Review the Hyperping dashboard: https://app.hyperping.io",
			resourceType, resourceID, err),
	)
}

// newUpdateError creates a standardized error for Update operations
func newUpdateError(resourceType, resourceID string, err error) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		fmt.Sprintf("Failed to Update %s", resourceType),
		fmt.Sprintf("Unable to update %s (ID: %s), got error: %s\n\n"+
			"Troubleshooting:\n"+
			"- Verify the resource still exists in Hyperping\n"+
			"- Check your API key has update permissions\n"+
			"- Verify the update values are valid\n"+
			"- Review the Hyperping dashboard: https://app.hyperping.io",
			resourceType, resourceID, err),
	)
}

// newDeleteError creates a standardized error for Delete operations
func newDeleteError(resourceType, resourceID string, err error) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		fmt.Sprintf("Failed to Delete %s", resourceType),
		fmt.Sprintf("Unable to delete %s (ID: %s), got error: %s\n\n"+
			"Troubleshooting:\n"+
			"- Verify the resource still exists in Hyperping\n"+
			"- Check your API key has delete permissions\n"+
			"- Check if the resource has dependencies that must be removed first\n"+
			"- Review the Hyperping dashboard: https://app.hyperping.io",
			resourceType, resourceID, err),
	)
}

// newListError creates a standardized error for List operations
func newListError(resourceType string, err error) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		fmt.Sprintf("Failed to List %s", resourceType),
		fmt.Sprintf("Unable to list %s, got error: %s\n\n"+
			"Troubleshooting:\n"+
			"- Verify your API key has read permissions\n"+
			"- Check network connectivity\n"+
			"- Check API service status: https://status.hyperping.app",
			resourceType, err),
	)
}

// Secondary Operation Errors

// newReadAfterCreateError creates a standardized error for reading after successful create
func newReadAfterCreateError(resourceType, resourceID string, err error) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		fmt.Sprintf("%s Created But Read Failed", resourceType),
		fmt.Sprintf("%s was created successfully (ID: %s) but reading it back failed: %s\n\n"+
			"The resource exists in Hyperping but may not be in Terraform state. "+
			"You may need to import it manually:\n"+
			"  terraform import hyperping_%s.example %s",
			resourceType, resourceID, err, resourceType, resourceID),
	)
}

// newReadAfterUpdateError creates a standardized error for read-after-update failures
func newReadAfterUpdateError(resourceType, resourceID string, err error) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		fmt.Sprintf("%s Updated But Read Failed", resourceType),
		fmt.Sprintf("%s was updated successfully (ID: %s) but reading the updated state failed: %s\n\n"+
			"The resource was modified in Hyperping but Terraform state may be inconsistent. "+
			"Run 'terraform refresh' to sync the state:\n"+
			"  terraform refresh",
			resourceType, resourceID, err),
	)
}

// Configuration and Validation Errors

// newConfigError creates a standardized error for configuration issues
func newConfigError(message string) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		"Configuration Error",
		message+"\n\n"+
			"Please review your Terraform configuration and fix the invalid values.",
	)
}

// newValidationError creates a standardized error for input validation failures
func newValidationError(field, message string) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		fmt.Sprintf("Invalid %s", field),
		message,
	)
}

// newImportError creates a standardized error for import operations
func newImportError(resourceType string, err error) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		"Import Failed",
		fmt.Sprintf("Cannot import %s: %s\n\n"+
			"Verify the import ID format is correct. "+
			"See the resource documentation for import examples.",
			resourceType, err),
	)
}

// Provider Configuration Errors

// newUnexpectedConfigTypeError creates a standardized error for type assertion failures
func newUnexpectedConfigTypeError(expected string, actual interface{}) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		"Unexpected Resource Configure Type",
		fmt.Sprintf("Expected %s, got: %T. "+
			"This is a provider bug - please report this issue to the provider developers.",
			expected, actual),
	)
}

// Warning Helpers

// newDeleteWarning creates a standardized warning for delete operations
func newDeleteWarning(resourceType, message string) diag.Diagnostic {
	return diag.NewWarningDiagnostic(
		fmt.Sprintf("%s Not Found", resourceType),
		message+" The resource may have already been deleted outside of Terraform.",
	)
}

// Enhanced Error Helpers with Context-Specific Troubleshooting

// BuildTroubleshootingSteps generates context-specific troubleshooting guidance.
func BuildTroubleshootingSteps(ctx ErrorContext) string {
	var steps []string

	switch ctx.Type {
	case "not_found":
		steps = buildNotFoundSteps(ctx)
	case "auth_error":
		steps = buildAuthErrorSteps(ctx)
	case "rate_limit":
		steps = buildRateLimitSteps(ctx)
	case "server_error":
		steps = buildServerErrorSteps(ctx)
	case "validation":
		steps = buildValidationErrorSteps(ctx)
	default:
		steps = buildGenericSteps(ctx)
	}

	return "Troubleshooting:\n" + joinSteps(steps)
}

// buildNotFoundSteps generates troubleshooting steps for 404 errors.
func buildNotFoundSteps(ctx ErrorContext) []string {
	dashboardURL := getDashboardURL(ctx.ResourceType)

	steps := []string{
		fmt.Sprintf("1. Verify the %s still exists in the Hyperping dashboard: %s", ctx.ResourceType, dashboardURL),
		"2. Check if the resource was deleted outside of Terraform",
	}

	if ctx.ResourceID != "" {
		steps = append(steps,
			fmt.Sprintf("3. Verify the resource ID is correct: %s", ctx.ResourceID),
			"4. Try viewing the resource state: terraform state show <resource_address>",
		)
	}

	steps = append(steps, "5. If the resource was deleted manually, remove it from Terraform state or recreate it")

	return steps
}

// buildAuthErrorSteps generates troubleshooting steps for 401/403 errors.
func buildAuthErrorSteps(ctx ErrorContext) []string {
	steps := []string{
		"1. Verify your HYPERPING_API_KEY environment variable is set:",
		"   $ echo $HYPERPING_API_KEY",
		"2. Confirm your API key format is correct (starts with 'sk_')",
		"3. Test your API key with curl:",
		"   $ curl -H \"Authorization: Bearer $HYPERPING_API_KEY\" https://api.hyperping.io/v1/monitors",
	}

	if ctx.HTTPStatus == 403 {
		steps = append(steps,
			"4. Verify your API key has the required permissions for this operation",
			fmt.Sprintf("5. Check if your account has access to %s resources", ctx.ResourceType),
		)
	} else {
		steps = append(steps, "4. If the key is invalid, generate a new one at: https://app.hyperping.io/settings/api")
	}

	return steps
}

// buildRateLimitSteps generates troubleshooting steps for 429 errors.
func buildRateLimitSteps(ctx ErrorContext) []string {
	waitTime := ctx.RetryAfter
	if waitTime == 0 {
		waitTime = 60
	}

	steps := []string{
		fmt.Sprintf("1. Wait %d seconds before retrying", waitTime),
		"2. Reduce the number of parallel operations:",
		"   $ terraform apply -parallelism=1",
		"3. Consider batching your operations to reduce API calls",
		"4. Review Hyperping rate limits documentation: https://docs.hyperping.io/api/rate-limits",
	}

	return steps
}

// buildServerErrorSteps generates troubleshooting steps for 5xx errors.
func buildServerErrorSteps(ctx ErrorContext) []string {
	steps := []string{
		"1. Check Hyperping service status: https://status.hyperping.app",
		"2. Wait a few moments and retry the operation",
		"3. If the error persists, check for any ongoing incidents",
		"4. Consider implementing retry logic with exponential backoff",
		"5. Contact Hyperping support if the issue continues: https://hyperping.io/support",
	}

	return steps
}

// buildValidationErrorSteps generates troubleshooting steps for 400/422 errors.
func buildValidationErrorSteps(ctx ErrorContext) []string {
	steps := []string{
		"1. Review the error message for specific field validation failures",
		"2. Check that all required fields are provided",
		fmt.Sprintf("3. Consult the %s documentation for valid field values", ctx.ResourceType),
		"4. Common validation issues:",
	}

	// Add resource-specific validation guidance
	switch ctx.ResourceType {
	case "Monitor":
		steps = append(steps,
			"   - URL must be valid and accessible",
			"   - Frequency must be one of: 10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400",
			"   - Timeout must be one of: 5, 10, 15, 20",
			"   - Regions must be valid Hyperping region codes",
		)
	case "Incident":
		steps = append(steps,
			"   - Title must be 1-255 characters",
			"   - Status must be one of: investigating, identified, monitoring, resolved",
			"   - Severity must be one of: minor, major, critical",
		)
	case "Maintenance":
		steps = append(steps,
			"   - scheduledStart and scheduledEnd must be valid ISO 8601 timestamps",
			"   - scheduledEnd must be after scheduledStart",
			"   - Title must be 1-255 characters",
		)
	default:
		steps = append(steps, "   - Check field types and value constraints")
	}

	steps = append(steps, "5. Review documentation: https://docs.hyperping.io")

	return steps
}

// buildGenericSteps generates generic troubleshooting steps for unknown errors.
func buildGenericSteps(ctx ErrorContext) []string {
	steps := []string{
		"1. Review the error message for specific details",
		"2. Check network connectivity to Hyperping API",
		"3. Verify your API key is valid and has required permissions",
		"4. Check Hyperping service status: https://status.hyperping.app",
		"5. Review Terraform provider logs for additional context",
		"6. If the issue persists, report it to the provider maintainers: https://github.com/develeap/terraform-provider-hyperping/issues",
	}

	return steps
}

// getDashboardURL returns the dashboard URL for a specific resource type.
func getDashboardURL(resourceType string) string {
	switch resourceType {
	case "Monitor":
		return "https://app.hyperping.io/monitors"
	case "Incident":
		return "https://app.hyperping.io/incidents"
	case "Maintenance":
		return "https://app.hyperping.io/maintenance"
	case "Statuspage":
		return "https://app.hyperping.io/statuspages"
	case "Healthcheck":
		return "https://app.hyperping.io/healthchecks"
	case "Outage":
		return "https://app.hyperping.io/outages"
	default:
		return "https://app.hyperping.io"
	}
}

// joinSteps formats troubleshooting steps into a readable string.
func joinSteps(steps []string) string {
	return strings.Join(steps, "\n") + "\n"
}

// Enhanced error creation functions with context-specific troubleshooting

// NewReadErrorWithContext creates an enhanced read error with troubleshooting steps.
func NewReadErrorWithContext(resourceType, resourceID string, err error) diag.Diagnostic {
	ctx := DetectErrorContext(resourceType, resourceID, "read", err)
	troubleshooting := BuildTroubleshootingSteps(ctx)

	return diag.NewErrorDiagnostic(
		fmt.Sprintf("Failed to Read %s", resourceType),
		fmt.Sprintf("Unable to read %s (ID: %s), got error: %s\n\n%s",
			resourceType, resourceID, err, troubleshooting),
	)
}

// NewCreateErrorWithContext creates an enhanced create error with troubleshooting steps.
func NewCreateErrorWithContext(resourceType string, err error) diag.Diagnostic {
	ctx := DetectErrorContext(resourceType, "", "create", err)
	troubleshooting := BuildTroubleshootingSteps(ctx)

	return diag.NewErrorDiagnostic(
		fmt.Sprintf("Failed to Create %s", resourceType),
		fmt.Sprintf("Unable to create %s, got error: %s\n\n%s",
			resourceType, err, troubleshooting),
	)
}

// NewUpdateErrorWithContext creates an enhanced update error with troubleshooting steps.
func NewUpdateErrorWithContext(resourceType, resourceID string, err error) diag.Diagnostic {
	ctx := DetectErrorContext(resourceType, resourceID, "update", err)
	troubleshooting := BuildTroubleshootingSteps(ctx)

	return diag.NewErrorDiagnostic(
		fmt.Sprintf("Failed to Update %s", resourceType),
		fmt.Sprintf("Unable to update %s (ID: %s), got error: %s\n\n%s",
			resourceType, resourceID, err, troubleshooting),
	)
}

// NewDeleteErrorWithContext creates an enhanced delete error with troubleshooting steps.
func NewDeleteErrorWithContext(resourceType, resourceID string, err error) diag.Diagnostic {
	ctx := DetectErrorContext(resourceType, resourceID, "delete", err)
	troubleshooting := BuildTroubleshootingSteps(ctx)

	return diag.NewErrorDiagnostic(
		fmt.Sprintf("Failed to Delete %s", resourceType),
		fmt.Sprintf("Unable to delete %s (ID: %s), got error: %s\n\n%s",
			resourceType, resourceID, err, troubleshooting),
	)
}
