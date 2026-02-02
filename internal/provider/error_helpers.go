// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"

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

// Deprecated: Use newCreateError, newReadError, etc. instead
// formatResourceError formats an error message with resource context
func formatResourceError(resourceType, operation, resourceID string, err error) (string, string) {
	summary := fmt.Sprintf("Error %s %s", operation, resourceType)

	var detail string
	if resourceID != "" {
		detail = fmt.Sprintf("Resource ID: %s\n\n%s\n\nCheck the Hyperping dashboard for more details: https://app.hyperping.io",
			resourceID, err.Error())
	} else {
		detail = fmt.Sprintf("%s\n\nCheck the Hyperping dashboard for more details: https://app.hyperping.io", err.Error())
	}

	return summary, detail
}

// Deprecated: Use newClientConfigError instead
// formatAPIError formats an API error with helpful context
func formatAPIError(operation string, err error) (string, string) {
	summary := fmt.Sprintf("API Error During %s", operation)
	detail := fmt.Sprintf("%s\n\nThis may be due to:\n"+
		"- Invalid API key\n"+
		"- Network connectivity issues\n"+
		"- Hyperping API service disruption\n\n"+
		"Check your configuration and https://status.hyperping.app for service status.",
		err.Error())

	return summary, detail
}
