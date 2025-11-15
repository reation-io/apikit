package builder

import (
	"fmt"
	"strings"

	coreast "github.com/reation-io/apikit/core/ast"
	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/spec"

	// Import all parsers to trigger auto-registration
	_ "github.com/reation-io/apikit/openapi/parsers/tags"
)

// ExtractFromGeneric extracts OpenAPI specification from generic parse results
// This adapter filters for swagger:meta, swagger:route, and swagger:model directives
func ExtractFromGeneric(results []*coreast.ParseResult) (*spec.OpenAPI, error) {
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
		// Process swagger:meta
		if err := extractMeta(result, openapi); err != nil {
			return nil, fmt.Errorf("failed to extract meta from %s: %w", result.Filename, err)
		}

		// Process swagger:route
		if err := extractRoutes(result, openapi); err != nil {
			return nil, fmt.Errorf("failed to extract routes from %s: %w", result.Filename, err)
		}

		// Process swagger:model
		if err := extractModels(result, openapi); err != nil {
			return nil, fmt.Errorf("failed to extract models from %s: %w", result.Filename, err)
		}
	}

	return openapi, nil
}

// ExtractMultipleFromGeneric extracts multiple OpenAPI specifications from generic parse results
// based on Spec: tags in swagger:meta and swagger:route directives
// Returns a map of spec name to OpenAPI specification
func ExtractMultipleFromGeneric(results []*coreast.ParseResult) (map[string]*spec.OpenAPI, error) {
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
	metaBySpec := make(map[string][]*coreast.Struct)

	for _, result := range results {
		for _, s := range result.Structs {
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
		for _, s := range result.Structs {
			if !hasDirective(s.Doc, "swagger:model") {
				continue
			}

			schema := convertStructToSchema(s)

			// Parse field tags
			for _, field := range s.Fields {
				if field.Doc != nil || field.Comment != nil {
					fieldSchema := schema.Properties[getJSONName(field)]
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
func extractMeta(result *coreast.ParseResult, openapi *spec.OpenAPI) error {
	for _, s := range result.Structs {
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
func extractRoutes(result *coreast.ParseResult, openapi *spec.OpenAPI) error {
	for _, s := range result.Structs {
		if !hasDirective(s.Doc, "swagger:route") {
			continue
		}

		// Parse the route line: swagger:route METHOD PATH TAG OPERATION_ID
		routeInfo, err := parseRouteLine(s.Doc)
		if err != nil {
			return err
		}

		// Create operation
		operation := &spec.Operation{
			OperationID: routeInfo.OperationID,
			Tags:        []string{routeInfo.Tag},
			Responses: &spec.Responses{
				StatusCodeResponses: make(map[string]*spec.Response),
			},
		}

		// Parse operation tags
		if err := parsers.GlobalRegistry().Parse("swagger:route", s.Doc, operation, parsers.ContextRoute); err != nil {
			if !isInvalidTargetError(err) {
				return err
			}
		}

		// Add operation to path
		if openapi.Paths.PathItems[routeInfo.Path] == nil {
			openapi.Paths.PathItems[routeInfo.Path] = &spec.PathItem{}
		}

		pathItem := openapi.Paths.PathItems[routeInfo.Path]
		switch strings.ToUpper(routeInfo.Method) {
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
func extractRoutesMulti(result *coreast.ParseResult, specs map[string]*spec.OpenAPI) error {
	for _, s := range result.Structs {
		if !hasDirective(s.Doc, "swagger:route") {
			continue
		}

		// Parse the route line: swagger:route METHOD PATH TAG OPERATION_ID
		routeInfo, err := parseRouteLine(s.Doc)
		if err != nil {
			return err
		}

		// Create operation
		operation := &spec.Operation{
			OperationID: routeInfo.OperationID,
			Tags:        []string{routeInfo.Tag},
			Responses: &spec.Responses{
				StatusCodeResponses: make(map[string]*spec.Response),
			},
		}

		// Parse operation tags
		if err := parsers.GlobalRegistry().Parse("swagger:route", s.Doc, operation, parsers.ContextRoute); err != nil {
			if !isInvalidTargetError(err) {
				return err
			}
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
			clonedOp := cloneOperationForAdapter(operation)

			// Add operation to path
			if targetSpec.Paths.PathItems[routeInfo.Path] == nil {
				targetSpec.Paths.PathItems[routeInfo.Path] = &spec.PathItem{}
			}

			pathItem := targetSpec.Paths.PathItems[routeInfo.Path]
			switch strings.ToUpper(routeInfo.Method) {
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

// cloneOperationForAdapter creates a copy of an operation (without x-specs extension)
func cloneOperationForAdapter(op *spec.Operation) *spec.Operation {
	if op == nil {
		return nil
	}

	cloned := &spec.Operation{
		Tags:        make([]string, len(op.Tags)),
		Summary:     op.Summary,
		Description: op.Description,
		OperationID: op.OperationID,
		Deprecated:  op.Deprecated,
	}

	copy(cloned.Tags, op.Tags)

	// Clone parameters
	if op.Parameters != nil {
		cloned.Parameters = make([]*spec.Parameter, len(op.Parameters))
		copy(cloned.Parameters, op.Parameters)
	}

	// Clone request body
	cloned.RequestBody = op.RequestBody

	// Clone responses
	if op.Responses != nil {
		cloned.Responses = &spec.Responses{
			StatusCodeResponses: make(map[string]*spec.Response),
			Default:             op.Responses.Default,
		}
		for code, resp := range op.Responses.StatusCodeResponses {
			cloned.Responses.StatusCodeResponses[code] = resp
		}
	}

	// Clone security
	if op.Security != nil {
		cloned.Security = make([]spec.SecurityRequirement, len(op.Security))
		copy(cloned.Security, op.Security)
	}

	// Clone servers
	if op.Servers != nil {
		cloned.Servers = make([]*spec.Server, len(op.Servers))
		copy(cloned.Servers, op.Servers)
	}

	// Don't copy Extensions (we don't want x-specs in the output)

	return cloned
}

// extractModels extracts swagger:model information
func extractModels(result *coreast.ParseResult, openapi *spec.OpenAPI) error {
	for _, s := range result.Structs {
		if !hasDirective(s.Doc, "swagger:model") {
			continue
		}

		// Convert struct to schema
		schema := convertStructToSchema(s)

		// Parse field tags
		for i, field := range s.Fields {
			if field.Doc != nil || field.Comment != nil {
				fieldSchema := schema.Properties[getJSONName(field)]
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
			_ = i // unused
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
func convertStructToSchema(s *coreast.Struct) *spec.Schema {
	schema := &spec.Schema{
		Type:       "object",
		Properties: make(map[string]*spec.Schema),
	}

	for _, field := range s.Fields {
		// Skip embedded fields for now
		if field.IsEmbedded {
			continue
		}

		jsonName := getJSONName(field)
		if jsonName == "-" {
			continue
		}

		fieldSchema := typeToSchema(field.Type, field.IsPointer, field.IsSlice)
		schema.Properties[jsonName] = fieldSchema
	}

	return schema
}

// getJSONName extracts the JSON name from struct tag
func getJSONName(field *coreast.Field) string {
	if field.Tag == "" {
		return field.Name
	}

	// Parse json tag
	tag := field.Tag
	if idx := strings.Index(tag, "json:"); idx != -1 {
		rest := tag[idx+5:]
		rest = strings.TrimPrefix(rest, "\"")
		if endIdx := strings.Index(rest, "\""); endIdx != -1 {
			jsonTag := rest[:endIdx]
			// Split by comma to get just the name
			parts := strings.Split(jsonTag, ",")
			if len(parts) > 0 && parts[0] != "" {
				return parts[0]
			}
		}
	}

	return field.Name
}

// typeToSchema converts a Go type to OpenAPI schema
func typeToSchema(goType string, isPointer bool, isSlice bool) *spec.Schema {
	// Remove pointer prefix
	goType = strings.TrimPrefix(goType, "*")

	// Handle slices
	if isSlice {
		elemType := strings.TrimPrefix(goType, "[]")
		return &spec.Schema{
			Type:  "array",
			Items: typeToSchema(elemType, false, false),
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
