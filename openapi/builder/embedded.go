package builder

import (
	"go/ast"
	"strings"

	"github.com/reation-io/apikit/openapi/spec"
)

// resolveEmbeddedFields resolves embedded struct fields and collects all properties
// This includes fields from embedded structs that should be "flattened" into the parent
func (b *Builder) resolveEmbeddedFields(structType *ast.StructType) []*spec.FieldInfo {
	if structType == nil || structType.Fields == nil {
		return nil
	}

	var allFields []*spec.FieldInfo

	for _, field := range structType.Fields.List {
		// Check if this is an embedded field (no name)
		if len(field.Names) == 0 {
			// This is an embedded field
			embeddedFields := b.resolveEmbeddedType(field.Type)
			allFields = append(allFields, embeddedFields...)
			continue
		}

		// Regular field
		fieldInfo := b.parseFieldInfo(field)
		if fieldInfo != nil {
			allFields = append(allFields, fieldInfo)
		}
	}

	return allFields
}

// resolveEmbeddedType resolves fields from an embedded type
func (b *Builder) resolveEmbeddedType(expr ast.Expr) []*spec.FieldInfo {
	// Handle pointer types
	if starExpr, ok := expr.(*ast.StarExpr); ok {
		return b.resolveEmbeddedType(starExpr.X)
	}

	// Get the type name
	typeName := ""
	switch t := expr.(type) {
	case *ast.Ident:
		typeName = t.Name
	case *ast.SelectorExpr:
		// External package type (e.g., pkg.Type)
		if ident, ok := t.X.(*ast.Ident); ok {
			typeName = ident.Name + "." + t.Sel.Name
		}
	}

	if typeName == "" {
		return nil
	}

	// Look up the struct in our discovered types
	return b.getFieldsFromType(typeName)
}

// getFieldsFromType retrieves fields from a known type
func (b *Builder) getFieldsFromType(typeName string) []*spec.FieldInfo {
	// First check if we have this as a processed schema
	if b.document != nil && b.document.Components != nil && b.document.Components.Schemas != nil {
		if schema, ok := b.document.Components.Schemas[typeName]; ok {
			return b.schemaToFieldInfos(schema)
		}
	}

	// Return empty if type not found
	return nil
}

// schemaToFieldInfos converts a schema's properties to FieldInfo slice
func (b *Builder) schemaToFieldInfos(schema *spec.Schema) []*spec.FieldInfo {
	if schema == nil || schema.Properties == nil {
		return nil
	}

	requiredMap := make(map[string]bool)
	for _, r := range schema.Required {
		requiredMap[r] = true
	}

	var fields []*spec.FieldInfo
	for name, propSchema := range schema.Properties {
		field := &spec.FieldInfo{
			Name:        name,
			Type:        propSchema.Type,
			Description: propSchema.Description,
			Required:    requiredMap[name],
			Nullable:    propSchema.Nullable,
			Tags:        make(map[string]string),
			Validations: make(map[string]string),
		}

		// Copy validation rules
		if propSchema.Format != "" {
			field.Validations["format"] = propSchema.Format
		}
		if propSchema.Pattern != "" {
			field.Validations["pattern"] = propSchema.Pattern
		}
		if propSchema.Minimum != nil {
			field.Validations["minimum"] = strings.TrimSuffix(strings.TrimSuffix(string(rune(*propSchema.Minimum)), ".0"), "0")
		}
		if propSchema.Maximum != nil {
			field.Validations["maximum"] = strings.TrimSuffix(strings.TrimSuffix(string(rune(*propSchema.Maximum)), ".0"), "0")
		}

		// Check if array
		if propSchema.Type == "array" {
			field.IsArray = true
		}

		fields = append(fields, field)
	}

	return fields
}

// parseFieldInfo parses a single field into FieldInfo
func (b *Builder) parseFieldInfo(field *ast.Field) *spec.FieldInfo {
	if len(field.Names) == 0 {
		return nil
	}

	fieldInfo := spec.NewFieldInfo()
	fieldInfo.Name = field.Names[0].Name

	// Parse struct tags
	if field.Tag != nil {
		fieldInfo.Tags = parseStructTags(field.Tag.Value)
		// Get JSON name
		if jsonTag, ok := fieldInfo.Tags["json"]; ok {
			parts := strings.Split(jsonTag, ",")
			if len(parts) > 0 && parts[0] != "-" && parts[0] != "" {
				fieldInfo.Name = parts[0]
			}
			// Check for omitempty
			for _, part := range parts[1:] {
				if part == "omitempty" || part == "omitzero" {
					fieldInfo.HasOmitempty = true
					break
				}
			}
		}
	}

	// Parse field type
	b.populateFieldTypeInfo(fieldInfo, field.Type)

	// Parse field comments
	if field.Doc != nil {
		fieldInfo.Description, fieldInfo.ExplicitRequired, fieldInfo.ExplicitOptional =
			parseFieldComments(field.Doc)
	}

	return fieldInfo
}

// populateFieldTypeInfo fills in type information for a field
func (b *Builder) populateFieldTypeInfo(fieldInfo *spec.FieldInfo, expr ast.Expr) {
	// Try registered type processors first
	if ProcessTypeWithProcessors(fieldInfo, expr) {
		return
	}

	switch t := expr.(type) {
	case *ast.Ident:
		// Check for enum
		if enumInfo := b.enumRegistry.GetByTypeName(t.Name); enumInfo != nil {
			fieldInfo.Enum = enumInfo.Name
			fieldInfo.Type = enumInfo.BaseType
			return
		}
		// Check for type mapping
		mapping := MapGoTypeToOpenAPI(t.Name)
		fieldInfo.Type = mapping.Type
		if mapping.Format != "" {
			fieldInfo.Validations["format"] = mapping.Format
		}

	case *ast.StarExpr:
		fieldInfo.IsPointer = true
		fieldInfo.Nullable = true
		b.populateFieldTypeInfo(fieldInfo, t.X)

	case *ast.ArrayType:
		fieldInfo.IsArray = true
		// Create a temp field info for the element type
		elemInfo := spec.NewFieldInfo()
		b.populateFieldTypeInfo(elemInfo, t.Elt)
		fieldInfo.Type = elemInfo.Type

	case *ast.MapType:
		fieldInfo.IsMap = true
		fieldInfo.Type = "object"
		// Get the key type
		if keyIdent, ok := t.Key.(*ast.Ident); ok {
			fieldInfo.MapKeyType = keyIdent.Name
		}

	case *ast.SelectorExpr:
		// Qualified type (pkg.Type)
		if ident, ok := t.X.(*ast.Ident); ok {
			qualifiedName := ident.Name + "." + t.Sel.Name
			// Try type processors again with qualified name
			tempInfo := spec.NewFieldInfo()
			if ProcessTypeWithProcessors(tempInfo, expr) {
				*fieldInfo = *tempInfo
				return
			}
			// Default handling for specific known types
			if qualifiedName == "time.Time" {
				fieldInfo.Type = "string"
				fieldInfo.Validations["format"] = "date-time"
			} else {
				fieldInfo.Type = "object"
			}
		}

	case *ast.StructType:
		// Inline struct
		fieldInfo.IsInlineStruct = true
		fieldInfo.Type = "object"
		fieldInfo.InlineStruct = b.parseInlineStruct(t)
	}
}

// parseInlineStruct parses an inline struct definition
func (b *Builder) parseInlineStruct(structType *ast.StructType) *spec.StructInfo {
	if structType == nil || structType.Fields == nil {
		return nil
	}

	structInfo := spec.NewStructInfo()
	structInfo.Fields = b.resolveEmbeddedFields(structType)

	return structInfo
}

// parseStructTags parses struct tags into a map
func parseStructTags(tagStr string) map[string]string {
	tags := make(map[string]string)
	tagStr = strings.Trim(tagStr, "`")

	// Simple parser for struct tags
	for len(tagStr) > 0 {
		tagStr = strings.TrimSpace(tagStr)
		if tagStr == "" {
			break
		}

		// Find the key
		i := 0
		for i < len(tagStr) && tagStr[i] != ':' && tagStr[i] != '"' && tagStr[i] != ' ' {
			i++
		}
		if i >= len(tagStr) {
			break
		}
		if tagStr[i] != ':' {
			tagStr = tagStr[i:]
			continue
		}

		key := tagStr[:i]
		tagStr = tagStr[i+1:]

		// Skip to the opening quote
		if len(tagStr) == 0 || tagStr[0] != '"' {
			continue
		}
		tagStr = tagStr[1:]

		// Find the closing quote
		j := 0
		for j < len(tagStr) && tagStr[j] != '"' {
			if tagStr[j] == '\\' && j+1 < len(tagStr) {
				j++
			}
			j++
		}

		value := tagStr[:j]
		tags[key] = value

		if j < len(tagStr) {
			tagStr = tagStr[j+1:]
		} else {
			break
		}
	}

	return tags
}

// parseFieldComments extracts description and required info from field comments
func parseFieldComments(comments *ast.CommentGroup) (description string, explicitRequired, explicitOptional bool) {
	if comments == nil {
		return "", false, false
	}

	var lines []string
	for _, comment := range comments.List {
		text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
		lower := strings.ToLower(text)

		// Check for required: directive
		if strings.HasPrefix(lower, "required:") {
			val := strings.TrimSpace(text[9:])
			if val == "true" || val == "yes" {
				explicitRequired = true
			} else if val == "false" || val == "no" {
				explicitOptional = true
			}
			continue
		}

		// Skip other directives
		if isDirectiveLine(text) {
			continue
		}

		if text != "" {
			lines = append(lines, text)
		}
	}

	description = strings.Join(lines, " ")
	return
}
