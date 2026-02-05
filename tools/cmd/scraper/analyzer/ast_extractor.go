package analyzer

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
)

// SchemaField represents a field extracted from provider schema
type SchemaField struct {
	TerraformName string
	APIName       string
	GoFieldName   string
	Type          string
	Required      bool
	Optional      bool
	Computed      bool
	ValidValues   []string
	Description   string
}

// ResourceSchema represents extracted schema for a resource
type ResourceSchema struct {
	Name       string
	APISection string
	Fields     []SchemaField
	FilePath   string
}

// ExtractProviderSchemas parses all resource files and extracts schemas
func ExtractProviderSchemas(providerDir string) ([]ResourceSchema, error) {
	var schemas []ResourceSchema

	// Find all resource files
	resourceFiles, err := filepath.Glob(filepath.Join(providerDir, "provider", "*_resource.go"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob resource files: %w", err)
	}

	// Find client model files for API name mapping
	clientModelFiles, err := filepath.Glob(filepath.Join(providerDir, "client", "models_*.go"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob client model files: %w", err)
	}

	// Build API name mapping from client models
	apiNameMap := make(map[string]map[string]string) // resource -> goField -> apiName
	for _, file := range clientModelFiles {
		mappings, err := extractClientModelMappings(file)
		if err != nil {
			continue // Skip files that can't be parsed
		}
		for resource, fields := range mappings {
			if apiNameMap[resource] == nil {
				apiNameMap[resource] = make(map[string]string)
			}
			for goField, apiName := range fields {
				apiNameMap[resource][goField] = apiName
			}
		}
	}

	// Extract schemas from resource files
	for _, file := range resourceFiles {
		schema, err := extractResourceSchema(file, apiNameMap)
		if err != nil {
			continue // Skip files that can't be parsed
		}
		if schema != nil && len(schema.Fields) > 0 {
			schemas = append(schemas, *schema)
		}
	}

	return schemas, nil
}

// extractResourceSchema extracts schema from a single resource file
func extractResourceSchema(filePath string, apiNameMap map[string]map[string]string) (*ResourceSchema, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", filePath, err)
	}

	// Extract resource name from filename
	baseName := filepath.Base(filePath)
	resourceName := strings.TrimSuffix(baseName, "_resource.go")

	schema := &ResourceSchema{
		Name:       "hyperping_" + resourceName,
		APISection: resourceName + "s", // monitors, incidents, etc.
		FilePath:   filePath,
	}

	// Extract schema attributes from Schema() method (Computed/Optional/Required flags)
	schemaAttrs := extractSchemaAttributes(node)

	// Look for ResourceModel struct to get field names and types
	var fields []SchemaField
	ast.Inspect(node, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		// Look for structs ending in ResourceModel
		if !strings.HasSuffix(typeSpec.Name.Name, "ResourceModel") {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		// Extract fields from the struct
		for _, field := range structType.Fields.List {
			if field.Tag == nil {
				continue
			}

			// Parse the tfsdk tag
			tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
			tfsdkTag := tag.Get("tfsdk")
			if tfsdkTag == "" || tfsdkTag == "-" {
				continue
			}

			// Get Go field name
			goFieldName := ""
			if len(field.Names) > 0 {
				goFieldName = field.Names[0].Name
			}

			// Get type
			fieldType := extractFieldType(field.Type)

			// Look up API name from client models
			apiName := tfsdkTag // Default to tfsdk name
			if resourceMap, ok := apiNameMap[resourceName]; ok {
				if mappedName, ok := resourceMap[goFieldName]; ok {
					apiName = mappedName
				}
			}

			schemaField := SchemaField{
				TerraformName: tfsdkTag,
				APIName:       apiName,
				GoFieldName:   goFieldName,
				Type:          fieldType,
			}

			// Apply Computed/Optional/Required from Schema() method
			if attrs, ok := schemaAttrs[tfsdkTag]; ok {
				schemaField.Computed = attrs.Computed
				schemaField.Optional = attrs.Optional
				schemaField.Required = attrs.Required
			}

			fields = append(fields, schemaField)
		}

		return true
	})

	schema.Fields = fields
	return schema, nil
}

// schemaAttrFlags holds the flags extracted from schema.Attribute definitions
type schemaAttrFlags struct {
	Computed bool
	Optional bool
	Required bool
}

// extractSchemaAttributes parses the Schema() method to extract attribute flags
func extractSchemaAttributes(node *ast.File) map[string]schemaAttrFlags {
	result := make(map[string]schemaAttrFlags)

	ast.Inspect(node, func(n ast.Node) bool {
		// Look for function declarations named Schema
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok || funcDecl.Name.Name != "Schema" {
			return true
		}

		// Walk through the function body looking for schema attribute definitions
		ast.Inspect(funcDecl.Body, func(inner ast.Node) bool {
			// Look for key-value expressions in map literals
			keyValueExpr, ok := inner.(*ast.KeyValueExpr)
			if !ok {
				return true
			}

			// Key should be a string literal (attribute name like "id", "name")
			keyLit, ok := keyValueExpr.Key.(*ast.BasicLit)
			if !ok || keyLit.Kind != token.STRING {
				return true
			}

			attrName := strings.Trim(keyLit.Value, `"`)

			// Value should be a composite literal (schema.StringAttribute{...})
			compLit, ok := keyValueExpr.Value.(*ast.CompositeLit)
			if !ok {
				return true
			}

			// Extract Computed/Optional/Required from the composite literal
			flags := schemaAttrFlags{}
			for _, elt := range compLit.Elts {
				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}

				keyIdent, ok := kv.Key.(*ast.Ident)
				if !ok {
					continue
				}

				// Check for boolean true values
				valueIdent, isIdent := kv.Value.(*ast.Ident)
				if isIdent && valueIdent.Name == "true" {
					switch keyIdent.Name {
					case "Computed":
						flags.Computed = true
					case "Optional":
						flags.Optional = true
					case "Required":
						flags.Required = true
					}
				}
			}

			result[attrName] = flags
			return true
		})

		return true
	})

	return result
}

// extractClientModelMappings extracts json tag mappings from client model files
func extractClientModelMappings(filePath string) (map[string]map[string]string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// Extract resource name from filename (models_monitor.go -> monitor)
	baseName := filepath.Base(filePath)
	resourceName := strings.TrimPrefix(baseName, "models_")
	resourceName = strings.TrimSuffix(resourceName, ".go")

	mappings := make(map[string]map[string]string)

	ast.Inspect(node, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		// Look for structs that are likely API models
		structName := typeSpec.Name.Name
		if !isAPIModelStruct(structName) {
			return true
		}

		if mappings[resourceName] == nil {
			mappings[resourceName] = make(map[string]string)
		}

		for _, field := range structType.Fields.List {
			if field.Tag == nil || len(field.Names) == 0 {
				continue
			}

			goFieldName := field.Names[0].Name
			tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
			jsonTag := tag.Get("json")
			if jsonTag == "" || jsonTag == "-" {
				continue
			}

			// Extract just the field name (remove ,omitempty etc.)
			jsonFieldName := strings.Split(jsonTag, ",")[0]
			mappings[resourceName][goFieldName] = jsonFieldName
		}

		return true
	})

	return mappings, nil
}

// isAPIModelStruct checks if a struct name looks like an API model
func isAPIModelStruct(name string) bool {
	// Include main response structs and request structs
	patterns := []string{
		"Monitor", "Incident", "Maintenance", "Healthcheck",
		"StatusPage", "Outage", "Report", "Subscriber",
	}

	for _, p := range patterns {
		if strings.Contains(name, p) {
			return true
		}
	}
	return false
}

// extractFieldType converts AST type to string representation
func extractFieldType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return mapGoTypeToSchema(t.Name)
	case *ast.SelectorExpr:
		// types.String, types.Int64, etc.
		if ident, ok := t.X.(*ast.Ident); ok && ident.Name == "types" {
			return mapTerraformType(t.Sel.Name)
		}
		return t.Sel.Name
	case *ast.StarExpr:
		return extractFieldType(t.X)
	case *ast.ArrayType:
		return "array"
	case *ast.MapType:
		return "object"
	default:
		return "unknown"
	}
}

// mapGoTypeToSchema maps Go types to schema types
func mapGoTypeToSchema(goType string) string {
	switch goType {
	case "string":
		return "string"
	case "int", "int32", "int64":
		return "number"
	case "bool":
		return "boolean"
	case "float32", "float64":
		return "number"
	default:
		return goType
	}
}

// mapTerraformType maps Terraform SDK types to schema types
func mapTerraformType(tfType string) string {
	switch tfType {
	case "String":
		return "string"
	case "Int64", "Int32", "Number":
		return "number"
	case "Bool":
		return "boolean"
	case "List":
		return "array"
	case "Map", "Object":
		return "object"
	default:
		return strings.ToLower(tfType)
	}
}

// NormalizeFieldName converts field names to a common format for comparison
func NormalizeFieldName(name string) string {
	// Convert camelCase to snake_case
	re := regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := re.ReplaceAllString(name, "${1}_${2}")
	return strings.ToLower(snake)
}

// ExtractAllowedValues extracts allowed values from models_common.go
func ExtractAllowedValues(providerDir string) (map[string][]string, error) {
	commonFile := filepath.Join(providerDir, "client", "models_common.go")

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, commonFile, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	allowedValues := make(map[string][]string)

	ast.Inspect(node, func(n ast.Node) bool {
		// Look for var declarations like: var AllowedRegions = []string{...}
		genDecl, ok := n.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			return true
		}

		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok || len(valueSpec.Names) == 0 {
				continue
			}

			name := valueSpec.Names[0].Name
			if !strings.HasPrefix(name, "Allowed") {
				continue
			}

			// Extract values from composite literal
			if len(valueSpec.Values) == 0 {
				continue
			}

			compLit, ok := valueSpec.Values[0].(*ast.CompositeLit)
			if !ok {
				continue
			}

			var values []string
			for _, elt := range compLit.Elts {
				if basicLit, ok := elt.(*ast.BasicLit); ok && basicLit.Kind == token.STRING {
					// Remove quotes
					val := strings.Trim(basicLit.Value, `"`)
					values = append(values, val)
				}
			}

			if len(values) > 0 {
				allowedValues[name] = values
			}
		}

		return true
	})

	return allowedValues, nil
}
