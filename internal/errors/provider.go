// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package terraflyerrors

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// AddDiagnosticError adds an enhanced error to Terraform diagnostics.
func AddDiagnosticError(diags *diag.Diagnostics, summary string, err error) {
	if err == nil {
		return
	}

	// Check if already enhanced
	var enhanced *EnhancedError
	if errors, ok := err.(*EnhancedError); ok {
		enhanced = errors
	}

	if enhanced != nil {
		// Use enhanced error's title as summary if not provided
		if summary == "" {
			summary = enhanced.Title
		}

		// Format detailed error message
		detail := enhanced.Format()

		diags.AddError(summary, detail)
	} else {
		// Fall back to standard error
		diags.AddError(summary, err.Error())
	}
}

// ResourceError creates an enhanced error for a resource operation.
func ResourceError(err error, operation, resourceType, resourceName string) error {
	if err == nil {
		return nil
	}

	resource := fmt.Sprintf("%s.%s", resourceType, resourceName)
	return EnhanceClientError(err, operation, resource, "")
}

// FieldError creates an enhanced error for a specific field.
func FieldError(err error, operation, resourceType, resourceName, field string) error {
	if err == nil {
		return nil
	}

	resource := fmt.Sprintf("%s.%s", resourceType, resourceName)
	return EnhanceClientError(err, operation, resource, field)
}

// ImportError creates an enhanced error for import operations.
func ImportError(err error, resourceType, importID string) error {
	if err == nil {
		return nil
	}

	return EnhanceError(err, CategoryNotFound,
		WithTitle("Import Failed"),
		WithDescription(fmt.Sprintf("Unable to import resource with ID: %s", importID)),
		WithOperation("import"),
		WithResource(resourceType),
		WithSuggestions(
			"Verify the resource ID is correct",
			"Check if the resource exists in the Hyperping dashboard",
			"Ensure your API key has permission to access this resource",
		),
		WithCommands(
			fmt.Sprintf("terraform import %s.name %s", resourceType, importID),
		),
		WithDocLinks(
			fmt.Sprintf("https://registry.terraform.io/providers/develeap/hyperping/latest/docs/resources/%s#import",
				resourceType),
		),
	)
}

// CreateError creates an enhanced error for create operations.
func CreateError(err error, resourceType, resourceName string) error {
	return ResourceError(err, "create", resourceType, resourceName)
}

// ReadError creates an enhanced error for read operations.
func ReadError(err error, resourceType, resourceName string) error {
	return ResourceError(err, "read", resourceType, resourceName)
}

// UpdateError creates an enhanced error for update operations.
func UpdateError(err error, resourceType, resourceName string) error {
	return ResourceError(err, "update", resourceType, resourceName)
}

// DeleteError creates an enhanced error for delete operations.
func DeleteError(err error, resourceType, resourceName string) error {
	return ResourceError(err, "delete", resourceType, resourceName)
}

// PlanError creates an enhanced error for plan operations.
func PlanError(err error, field string) error {
	if err == nil {
		return nil
	}

	return EnhanceError(err, CategoryValidation,
		WithTitle("Plan Error"),
		WithDescription("Error during plan phase"),
		WithOperation("plan"),
		WithField(field),
	)
}

// ConfigError creates an enhanced error for configuration issues.
func ConfigError(field, message string) error {
	return EnhanceError(
		fmt.Errorf("configuration error: %s", message),
		CategoryValidation,
		WithTitle("Configuration Error"),
		WithDescription(message),
		WithField(field),
		WithSuggestions(
			"Check your provider configuration",
			"Verify all required fields are set",
		),
		WithDocLinks(
			"https://registry.terraform.io/providers/develeap/hyperping/latest/docs",
		),
	)
}

// GetResourceAddress extracts resource address from context or path.
// This is a helper for getting the resource address in provider code.
func GetResourceAddress(ctx context.Context, resourceType string) string {
	// In Terraform provider framework, we don't have direct access to
	// resource address from context. Return the resource type.
	// The actual resource name should be passed explicitly.
	return resourceType
}

// FormatResourceAddress formats a resource address for error messages.
func FormatResourceAddress(resourceType, resourceName string) string {
	if resourceName == "" {
		return resourceType
	}
	return fmt.Sprintf("%s.%s", resourceType, resourceName)
}
