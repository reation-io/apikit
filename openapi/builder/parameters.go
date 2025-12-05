package builder

import (
	"go/ast"
	"strings"

	"github.com/reation-io/apikit/openapi/spec"
)

// parseRequestStruct extracts parameters and request body from a struct following swagger:route
func (b *Builder) parseRequestStruct(structType *ast.StructType, operation *spec.Operation) {
	b.parseRequestStructWithIgnored(structType, operation, nil)
}

// parseRequestStructWithIgnored extracts parameters, filtering out ignored ones
func (b *Builder) parseRequestStructWithIgnored(structType *ast.StructType, operation *spec.Operation, ignoredParams []string) {
	if structType == nil || structType.Fields == nil {
		return
	}

	// Build a map for quick lookup of ignored parameters
	ignoredMap := make(map[string]bool)
	for _, p := range ignoredParams {
		ignoredMap[p] = true
	}
	// Also add ignored parameters from operation
	for _, p := range operation.IgnoredParameters {
		ignoredMap[p] = true
	}

	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			continue
		}

		fieldName := field.Names[0].Name

		// Check for swagger:ignore directive in field comments
		if hasSwaggerIgnore(field.Doc) {
			continue
		}

		// Parse field comments for in: directive
		inValue := ""
		description := ""
		required := false
		isExplode := false
		defaultVal := ""
		enumValues := []string{}

		if field.Doc != nil {
			for _, comment := range field.Doc.List {
				text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))

				// Check for in: directive
				if strings.HasPrefix(strings.ToLower(text), "in:") {
					inValue = strings.TrimSpace(strings.TrimPrefix(text[3:], " "))
				}
				// Check for required: directive
				if strings.HasPrefix(strings.ToLower(text), "required:") {
					val := strings.TrimSpace(text[9:])
					required = val == "true" || val == "yes"
				}
				// Check for explode: directive
				if strings.HasPrefix(strings.ToLower(text), "explode:") {
					val := strings.TrimSpace(text[8:])
					isExplode = val == "true" || val == "yes"
				}
				// Check for default: directive
				if strings.HasPrefix(strings.ToLower(text), "default:") {
					defaultVal = strings.TrimSpace(text[8:])
				}
				// Check for enum: directive
				if strings.HasPrefix(strings.ToLower(text), "enum:") {
					enumStr := strings.TrimSpace(text[5:])
					for _, v := range strings.Split(enumStr, ",") {
						enumValues = append(enumValues, strings.TrimSpace(v))
					}
				}
				// Collect description (lines that don't match directives)
				if !isDirectiveLine(text) && text != "" {
					if description != "" {
						description += " "
					}
					description += text
				}
			}
		}

		// Skip if no in: directive
		if inValue == "" {
			continue
		}

		// Handle body parameters specially
		if inValue == "body" {
			b.handleBodyParameter(field, operation)
			continue
		}

		// Get parameter name from json tag or field name
		paramName := b.getParameterName(field)
		if paramName == "" {
			paramName = strings.ToLower(fieldName[:1]) + fieldName[1:]
		}

		// Check if this parameter should be ignored
		if ignoredMap[paramName] || ignoredMap[fieldName] {
			continue
		}

		// Path parameters are always required
		if inValue == "path" {
			required = true
		}

		// Create parameter
		param := &spec.Parameter{
			Name:        paramName,
			In:          inValue,
			Description: description,
			Required:    required,
			Schema:      b.parseFieldTypeWithFormat(field.Type),
		}

		// Set explode if specified
		if isExplode {
			explodeVal := true
			param.Explode = &explodeVal
		}

		// Set default value
		if defaultVal != "" {
			param.Schema.Default = defaultVal
		}

		// Set enum values
		if len(enumValues) > 0 {
			enumAny := make([]any, len(enumValues))
			for i, v := range enumValues {
				enumAny[i] = v
			}
			param.Schema.Enum = enumAny
		}

		// Add parameter to operation
		operation.Parameters = append(operation.Parameters, param)
	}
}

// handleBodyParameter creates a request body from a body field
func (b *Builder) handleBodyParameter(field *ast.Field, operation *spec.Operation) {
	// Get the type name for the body
	typeName := b.getTypeName(field.Type)
	if typeName == "" {
		return
	}

	// Get description from comment
	description := ""
	if field.Doc != nil {
		for _, comment := range field.Doc.List {
			text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
			if !isDirectiveLine(text) && text != "" {
				if description != "" {
					description += " "
				}
				description += text
			}
		}
	}

	// Check if it's an array type
	isArray := false
	if _, ok := field.Type.(*ast.ArrayType); ok {
		isArray = true
	}

	// Create request body
	operation.RequestBody = &spec.RequestBody{
		Description: description,
		Required:    true,
		Content:     make(map[string]*spec.MediaType),
	}

	var schema *spec.Schema
	if isArray {
		// For array types, use array schema with items ref
		schema = &spec.Schema{
			Type: "array",
			Items: &spec.Schema{
				Ref: "#/components/schemas/" + typeName,
			},
		}
	} else {
		// For object types, use $ref directly
		schema = &spec.Schema{
			Ref: "#/components/schemas/" + typeName,
		}
	}

	// Add common content types
	operation.RequestBody.Content["application/json"] = &spec.MediaType{
		Schema: schema,
	}
	operation.RequestBody.Content["application/xml"] = &spec.MediaType{
		Schema: schema,
	}
	operation.RequestBody.Content["application/x-www-form-urlencoded"] = &spec.MediaType{
		Schema: schema,
	}
}

// getTypeName extracts the type name from an ast expression
func (b *Builder) getTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return b.getTypeName(t.X)
	case *ast.ArrayType:
		return b.getTypeName(t.Elt)
	case *ast.SelectorExpr:
		return t.Sel.Name
	case *ast.MapType:
		return "object"
	}
	return ""
}

// parseFieldTypeWithFormat parses a field type and includes format information
func (b *Builder) parseFieldTypeWithFormat(expr ast.Expr) *spec.Schema {
	schema := &spec.Schema{}

	switch t := expr.(type) {
	case *ast.Ident:
		// Check if this type is an enum
		if enumInfo := b.enumRegistry.GetByTypeName(t.Name); enumInfo != nil {
			return b.createEnumSchema(enumInfo)
		}

		// Basic types with format
		mapping := MapGoTypeToOpenAPI(t.Name)
		schema.Type = mapping.Type
		schema.Format = mapping.Format

	case *ast.ArrayType:
		schema.Type = "array"
		schema.Items = b.parseFieldTypeWithFormat(t.Elt)

	case *ast.StarExpr:
		// Pointer type
		return b.parseFieldTypeWithFormat(t.X)

	case *ast.SelectorExpr:
		// Check if this qualified type is an enum
		if enumInfo := b.enumRegistry.GetByTypeName(t.Sel.Name); enumInfo != nil {
			return b.createEnumSchema(enumInfo)
		}

		// Check for registered type processors (e.g. time.Time, uuid.UUID)
		tempInfo := spec.NewFieldInfo()
		if ProcessTypeWithProcessors(tempInfo, expr) {
			schema.Type = tempInfo.Type
			if format, ok := tempInfo.Validations["format"]; ok {
				schema.Format = format
			}
			return schema
		}

	case *ast.MapType:
		schema.Type = "object"
		if t.Value != nil {
			schema.AdditionalProperties = b.parseFieldTypeWithFormat(t.Value)
		}
	}

	return schema
}

// getParameterName gets the parameter name from json tag or field name
func (b *Builder) getParameterName(field *ast.Field) string {
	if field.Tag != nil {
		tag := field.Tag.Value
		tag = strings.Trim(tag, "`")

		// Try json tag first
		for _, part := range strings.Fields(tag) {
			if strings.HasPrefix(part, "json:") {
				jsonTag := strings.TrimPrefix(part, "json:")
				jsonTag = strings.Trim(jsonTag, `"`)
				parts := strings.Split(jsonTag, ",")
				if len(parts) > 0 && parts[0] != "-" && parts[0] != "" {
					return parts[0]
				}
			}
		}
	}

	// Fallback to field name (camelCase)
	if len(field.Names) > 0 {
		name := field.Names[0].Name
		// Convert first letter to lowercase for camelCase
		if len(name) > 0 {
			return strings.ToLower(name[:1]) + name[1:]
		}
	}

	return ""
}

// isDirectiveLine checks if a line is a directive (in:, required:, etc.)
func isDirectiveLine(text string) bool {
	directives := []string{
		"in:", "required:", "explode:", "default:", "enum:",
		"example:", "format:", "minimum:", "maximum:",
		"minlength:", "maxlength:", "pattern:",
	}
	lower := strings.ToLower(text)
	for _, d := range directives {
		if strings.HasPrefix(lower, d) {
			return true
		}
	}
	return false
}

// hasSwaggerIgnore checks if a comment group contains swagger:ignore directive
func hasSwaggerIgnore(comments *ast.CommentGroup) bool {
	if comments == nil {
		return false
	}
	for _, comment := range comments.List {
		text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
		lower := strings.ToLower(text)
		if lower == "swagger:ignore" || strings.HasPrefix(lower, "swagger:ignore ") {
			return true
		}
	}
	return false
}
