package builder

import (
	"fmt"
	"strings"

	"github.com/reation-io/apikit/core/definition"
	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/spec"

	// Import all parsers to trigger auto-registration
	_ "github.com/reation-io/apikit/openapi/parsers/tags"
)

// ExtractFromGeneric extracts OpenAPI specification from definition results
// This adapter filters for swagger:meta, swagger:route, and swagger:model directives
func ExtractFromGeneric(results []*definition.Definition) (*spec.OpenAPI, error) {
	openapi := &spec.OpenAPI{
		OpenAPI: "3.0.3",
		Info: &spec.Info{
			Title:   "API",
			Version: "1.0.0",
		},
		Paths: &spec.Paths{
			PathItems: make(map[string]*spec.PathItem),
		},
	}

	for _, result := range results {
		// Process swagger:meta (from Types that serve as meta holders)
		if err := extractMeta(result, openapi); err != nil {
			return nil, fmt.Errorf("failed to extract meta from %s: %w", result.Package, err)
		}

		// Process operations (compiled from functions and structs)
		if err := extractRoutes(result, openapi); err != nil {
			return nil, fmt.Errorf("failed to extract routes from %s: %w", result.Package, err)
		}

		// Process swagger:model (from Types)
		if err := extractModels(result, openapi); err != nil {
			return nil, fmt.Errorf("failed to extract models from %s: %w", result.Package, err)
		}
	}

	return openapi, nil
}

// ExtractMultipleFromGeneric extracts multiple OpenAPI specifications from definition results
// based on Spec: tags in swagger:meta and swagger:route directives
// Returns a map of spec name to OpenAPI specification
func ExtractMultipleFromGeneric(results []*definition.Definition) (map[string]*spec.OpenAPI, error) {
	specs := make(map[string]*spec.OpenAPI)

	// Initialize default spec
	specs["default"] = &spec.OpenAPI{
		OpenAPI: "3.0.3",
		Info: &spec.Info{
			Title:   "API",
			Version: "1.0.0",
		},
		Paths: &spec.Paths{
			PathItems: make(map[string]*spec.PathItem),
		},
	}

	// First pass: collect all meta blocks and their spec tags
	metaBySpec := make(map[string][]*definition.Type)

	for _, result := range results {
		for _, s := range result.Types {
			if !hasDirective(s.Doc, "swagger:meta") {
				continue
			}

			// Parse to get spec names
			tempInfo := &spec.Info{}
			if err := parsers.GlobalRegistry().Parse("swagger:meta", s.Doc, tempInfo, parsers.ContextMeta); err != nil {
				if !isInvalidTargetError(err) {
					return nil, err
				}
			}

			// Get spec names from extensions
			var specNames []string
			if tempInfo.Extensions != nil {
				if specs, ok := tempInfo.Extensions["x-specs"].([]string); ok {
					specNames = specs
				}
			}

			// If no spec tag, apply to default
			if len(specNames) == 0 {
				metaBySpec["default"] = append(metaBySpec["default"], s)
			} else {
				for _, specName := range specNames {
					metaBySpec[specName] = append(metaBySpec[specName], s)
				}
			}
		}
	}

	// Create specs for each meta block
	for specName, metaStructs := range metaBySpec {
		if specs[specName] == nil {
			specs[specName] = &spec.OpenAPI{
				OpenAPI: "3.0.3",
				Info: &spec.Info{
					Title:   "API",
					Version: "1.0.0",
				},
				Paths: &spec.Paths{
					PathItems: make(map[string]*spec.PathItem),
				},
			}
		}

		// Apply meta from all matching meta blocks
		for _, metaStruct := range metaStructs {
			if err := parsers.GlobalRegistry().Parse("swagger:meta", metaStruct.Doc, specs[specName].Info, parsers.ContextMeta); err != nil {
				if !isInvalidTargetError(err) {
					return nil, err
				}
			}

			if err := parsers.GlobalRegistry().Parse("swagger:meta", metaStruct.Doc, specs[specName], parsers.ContextMeta); err != nil {
				if !isInvalidTargetError(err) {
					return nil, err
				}
			}
		}
	}

	// Second pass: extract routes and distribute them
	for _, result := range results {
		if err := extractRoutesMulti(result, specs); err != nil {
			return nil, err
		}
	}

	// Third pass: extract models (shared across all specs)
	allModels := make(map[string]*spec.Schema)
	for _, result := range results {
		for _, s := range result.Types {
			if !hasDirective(s.Doc, "swagger:model") {
				continue
			}

			schema := convertStructToSchema(s)

			// Parse field tags
			for _, field := range s.Fields {
				if field.Doc != nil || field.Comment != nil {
					fieldSchema := schema.Properties[field.JSONName]
					if fieldSchema != nil {
						if field.Doc != nil {
							parsers.GlobalRegistry().Parse("swagger:model", field.Doc, fieldSchema, parsers.ContextField)
						}
						if field.Comment != nil {
							parsers.GlobalRegistry().Parse("swagger:model", field.Comment, fieldSchema, parsers.ContextField)
						}
					}
				}
			}

			allModels[s.Name] = schema
		}
	}

	// Add models to all specs
	for _, openapi := range specs {
		if len(allModels) > 0 {
			if openapi.Components == nil {
				openapi.Components = &spec.Components{}
			}
			if openapi.Components.Schemas == nil {
				openapi.Components.Schemas = make(map[string]*spec.Schema)
			}
			for name, schema := range allModels {
				openapi.Components.Schemas[name] = schema
			}
		}
	}

	return specs, nil
}

// extractMeta extracts swagger:meta information
func extractMeta(result *definition.Definition, openapi *spec.OpenAPI) error {
	for _, s := range result.Types {
		if !hasDirective(s.Doc, "swagger:meta") {
			continue
		}

		// Parse meta tags into Info
		if err := parsers.GlobalRegistry().Parse("swagger:meta", s.Doc, openapi.Info, parsers.ContextMeta); err != nil {
			if !isInvalidTargetError(err) {
				return err
			}
		}

		// Parse meta tags that target OpenAPI root
		if err := parsers.GlobalRegistry().Parse("swagger:meta", s.Doc, openapi, parsers.ContextMeta); err != nil {
			if !isInvalidTargetError(err) {
				return err
			}
		}
	}

	return nil
}

// extractRoutes extracts swagger:route information
func extractRoutes(result *definition.Definition, openapi *spec.OpenAPI) error {
	for _, op := range result.Operations {
		// Create operation using data already parsed by core/parser
		operation := &spec.Operation{
			OperationID: op.ID,
			Tags:        op.Tags,
			Responses: &spec.Responses{
				StatusCodeResponses: make(map[string]*spec.Response),
			},
		}

		// Parse additional operation tags from Docs (Summary, Description, Responses, etc.)
		if op.Doc != nil {
			if err := parsers.GlobalRegistry().Parse("swagger:route", op.Doc, operation, parsers.ContextRoute); err != nil {
				if !isInvalidTargetError(err) {
					return err
				}
			}
		}

		// Auto-generate multipart/form-data requestBody if Consumes includes it
		if hasMultipartFormData(operation) {
			generateMultipartRequestBody(operation, result)
		}

		// Add operation to path
		if openapi.Paths.PathItems[op.Path] == nil {
			openapi.Paths.PathItems[op.Path] = &spec.PathItem{}
		}

		pathItem := openapi.Paths.PathItems[op.Path]
		switch strings.ToUpper(op.Method) {
		case "GET":
			pathItem.Get = operation
		case "POST":
			pathItem.Post = operation
		case "PUT":
			pathItem.Put = operation
		case "DELETE":
			pathItem.Delete = operation
		case "PATCH":
			pathItem.Patch = operation
		case "OPTIONS":
			pathItem.Options = operation
		case "HEAD":
			pathItem.Head = operation
		}
	}

	return nil
}

// extractRoutesMulti extracts swagger:route information and distributes to multiple specs
func extractRoutesMulti(result *definition.Definition, specs map[string]*spec.OpenAPI) error {
	for _, op := range result.Operations {

		// Create operation
		operation := &spec.Operation{
			OperationID: op.ID,
			Tags:        op.Tags,
			Responses: &spec.Responses{
				StatusCodeResponses: make(map[string]*spec.Response),
			},
		}

		// Parse operation tags
		if op.Doc != nil {
			if err := parsers.GlobalRegistry().Parse("swagger:route", op.Doc, operation, parsers.ContextRoute); err != nil {
				if !isInvalidTargetError(err) {
					return err
				}
			}
		}

		// Auto-generate multipart/form-data requestBody if Consumes includes it
		if hasMultipartFormData(operation) {
			generateMultipartRequestBody(operation, result)
		}

		// Get spec names from operation extensions
		var specNames []string
		if operation.Extensions != nil {
			if specs, ok := operation.Extensions["x-specs"].([]string); ok {
				specNames = specs
			}
		}

		// If no spec tag, add to default
		if len(specNames) == 0 {
			specNames = []string{"default"}
		}

		// Add operation to each specified spec
		for _, specName := range specNames {
			// Ensure spec exists
			if specs[specName] == nil {
				specs[specName] = &spec.OpenAPI{
					OpenAPI: "3.0.3",
					Info: &spec.Info{
						Title:   "API",
						Version: "1.0.0",
					},
					Paths: &spec.Paths{
						PathItems: make(map[string]*spec.PathItem),
					},
				}
			}

			targetSpec := specs[specName]

			// Clone operation to avoid sharing references
			// Reuse internal clone from builder
			clonedOp := cloneOperation(operation)

			// Add operation to path
			if targetSpec.Paths.PathItems[op.Path] == nil {
				targetSpec.Paths.PathItems[op.Path] = &spec.PathItem{}
			}

			pathItem := targetSpec.Paths.PathItems[op.Path]
			switch strings.ToUpper(op.Method) {
			case "GET":
				pathItem.Get = clonedOp
			case "POST":
				pathItem.Post = clonedOp
			case "PUT":
				pathItem.Put = clonedOp
			case "DELETE":
				pathItem.Delete = clonedOp
			case "PATCH":
				pathItem.Patch = clonedOp
			case "OPTIONS":
				pathItem.Options = clonedOp
			case "HEAD":
				pathItem.Head = clonedOp
			}
		}
	}

	return nil
}

// extractModels extracts swagger:model information
func extractModels(result *definition.Definition, openapi *spec.OpenAPI) error {
	for _, s := range result.Types {
		if !hasDirective(s.Doc, "swagger:model") {
			continue
		}

		// Convert struct to schema
		schema := convertStructToSchema(s)

		// Parse field tags
		for _, field := range s.Fields {
			if field.Doc != nil || field.Comment != nil {
				fieldSchema := schema.Properties[field.JSONName]
				if fieldSchema != nil {
					// Parse field documentation
					if field.Doc != nil {
						parsers.GlobalRegistry().Parse("swagger:model", field.Doc, fieldSchema, parsers.ContextField)
					}
					if field.Comment != nil {
						parsers.GlobalRegistry().Parse("swagger:model", field.Comment, fieldSchema, parsers.ContextField)
					}
				}
			}
		}

		// Add to components
		if openapi.Components == nil {
			openapi.Components = &spec.Components{
				Schemas: make(map[string]*spec.Schema),
			}
		}
		openapi.Components.Schemas[s.Name] = schema
	}

	return nil
}

// convertStructToSchema converts a generic struct to OpenAPI schema
func convertStructToSchema(s *definition.Type) *spec.Schema {
	schema := &spec.Schema{
		Type:       "object",
		Properties: make(map[string]*spec.Schema),
	}

	for _, field := range s.Fields {
		if field.JSONName == "-" {
			continue
		}

		name := field.JSONName
		if name == "" {
			name = field.Name
		}

		fieldSchema := typeToSchema(field.Type)
		schema.Properties[name] = fieldSchema
	}

	return schema
}

// typeToSchema converts a Go type to OpenAPI schema
func typeToSchema(t *definition.Type) *spec.Schema {
	if t.GoType == "" {
		return &spec.Schema{Type: "string"}
	}

	goType := t.GoType

	// Remove pointer prefix
	goType = strings.TrimPrefix(goType, "*")

	// Handle slices
	if strings.HasPrefix(goType, "[]") {
		elemType := strings.TrimPrefix(goType, "[]")
		// simplified recursion
		return &spec.Schema{
			Type:  "array",
			Items: typeToSchema(&definition.Type{GoType: elemType}),
		}
	}

	// Map Go types to JSON Schema types
	switch goType {
	case "string":
		return &spec.Schema{Type: "string"}
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		return &spec.Schema{Type: "integer"}
	case "float32", "float64":
		return &spec.Schema{Type: "number"}
	case "bool":
		return &spec.Schema{Type: "boolean"}
	default:
		// Assume it's a reference to another schema
		return &spec.Schema{
			Ref: "#/components/schemas/" + goType,
		}
	}
}

// hasMultipartFormData checks if operation consumes multipart/form-data
func hasMultipartFormData(operation *spec.Operation) bool {
	if operation.RequestBody == nil {
		return false
	}

	_, ok := operation.RequestBody.Content["multipart/form-data"]
	return ok
}

// generateMultipartRequestBody generates multipart/form-data schema from struct fields
func generateMultipartRequestBody(operation *spec.Operation, result *definition.Definition) {
	if operation.RequestBody == nil || operation.RequestBody.Content["multipart/form-data"] == nil {
		return
	}

	properties := make(map[string]*spec.Schema)
	requiredMap := make(map[string]bool)

	// Iterate through all structs to find form fields
	for _, s := range result.Types {
		for _, field := range s.Fields {
			// Check if field has form tag or in:form comment
			isFormField := false
			formName := ""

			// Check for form tag
			if field.Tags != "" {
				tag := strings.Trim(field.Tags, "`")
				if strings.Contains(tag, "form:") {
					isFormField = true
					// Extract form name from tag
					formName = extractTagValue(tag, "form")
				}
			}

			// Check for // in:form comment
			if field.Doc != nil {
				for _, comment := range field.Doc.List {
					trimmed := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
					if strings.HasPrefix(trimmed, "in:form") {
						isFormField = true
						// Extract custom name if provided: // in:form custom_name
						parts := strings.Fields(trimmed)
						if len(parts) > 1 && formName == "" {
							formName = parts[1]
						}
						break
					}
				}
			}

			if !isFormField {
				continue
			}

			// Use field name if no custom form name
			if formName == "" {
				formName = toSnakeCase(field.Name)
			}

			// Detect file upload fields
			isFile := strings.Contains(field.Type.GoType, "multipart.FileHeader")

			if isFile {
				// File upload field
				if strings.HasPrefix(field.Type.GoType, "[]") {
					// Multiple files
					properties[formName] = &spec.Schema{
						Type: "array",
						Items: &spec.Schema{
							Type:   "string",
							Format: "binary",
						},
					}
				} else {
					// Single file
					properties[formName] = &spec.Schema{
						Type:   "string",
						Format: "binary",
					}
				}
			} else {
				// Regular form field
				properties[formName] = typeToSchema(field.Type)
			}

			// Check if required (validate tag)
			if field.Tags != "" {
				tag := strings.Trim(field.Tags, "`")
				if strings.Contains(tag, "validate:") {
					validateTag := extractTagValue(tag, "validate")
					if strings.Contains(validateTag, "required") {
						requiredMap[formName] = true
					}
				}
			}
		}
	}

	// Convert required map to slice
	required := make([]string, 0, len(requiredMap))
	for fieldName := range requiredMap {
		required = append(required, fieldName)
	}

	// Update the schema
	mediaType := operation.RequestBody.Content["multipart/form-data"]
	mediaType.Schema = &spec.Schema{
		Type:       "object",
		Properties: properties,
		Required:   required,
	}
}

// extractTagValue extracts value from struct tag like `form:"value"`
func extractTagValue(tag, key string) string {
	// Find key: pattern
	keyPattern := key + `:"`
	start := strings.Index(tag, keyPattern)
	if start == -1 {
		return ""
	}
	start += len(keyPattern)

	// Find closing quote
	end := strings.Index(tag[start:], `"`)
	if end == -1 {
		return ""
	}

	return tag[start : start+end]
}

// toSnakeCase converts CamelCase to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
