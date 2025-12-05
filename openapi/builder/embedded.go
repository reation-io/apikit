package builder

import (
	"go/ast"
	"strings"

	"github.com/reation-io/apikit/openapi/spec"
)

// EmbeddedProcessor handles resolution of embedded struct fields.
// This provides explicit control over how embedded fields are processed
// and merged into parent schemas.
type EmbeddedProcessor struct {
	builder *Builder
}

// NewEmbeddedProcessor creates a new embedded field processor
func NewEmbeddedProcessor(b *Builder) *EmbeddedProcessor {
	return &EmbeddedProcessor{builder: b}
}

// ProcessEmbeddedFields resolves and merges embedded fields from a struct type
// into the target schema. It handles:
// - Direct embedded types (Type)
// - Pointer embedded types (*Type)
// - Qualified types (pkg.Type)
// - Field override detection (parent fields take precedence)
func (p *EmbeddedProcessor) ProcessEmbeddedFields(
	structType *ast.StructType,
	targetSchema *spec.Schema,
	existingFields map[string]bool,
) (embeddedRequired []string) {
	if structType == nil || structType.Fields == nil {
		return nil
	}

	for _, field := range structType.Fields.List {
		// Only process embedded fields (no names)
		if len(field.Names) > 0 {
			continue
		}

		// Check for swagger:ignore directive
		if hasSwaggerIgnore(field.Doc) || hasSwaggerIgnore(field.Comment) {
			continue
		}

		embeddedSchema := p.resolveEmbeddedType(field.Type)
		if embeddedSchema == nil {
			continue
		}

		// Merge properties, respecting field overrides
		for propName, propSchema := range embeddedSchema.Properties {
			// Check if this field is overridden by parent
			jsonName := propName
			if existingFields[jsonName] {
				continue // Skip overridden fields
			}

			if targetSchema.Properties == nil {
				targetSchema.Properties = make(map[string]*spec.Schema)
			}
			targetSchema.Properties[propName] = cloneSchema(propSchema)
		}

		// Collect required fields from embedded schema
		for _, req := range embeddedSchema.Required {
			if !existingFields[req] {
				embeddedRequired = append(embeddedRequired, req)
			}
		}
	}

	return embeddedRequired
}

// resolveEmbeddedType resolves an embedded type expression to its schema
func (p *EmbeddedProcessor) resolveEmbeddedType(expr ast.Expr) *spec.Schema {
	switch t := expr.(type) {
	case *ast.Ident:
		return p.resolveTypeName(t.Name)

	case *ast.StarExpr:
		// Pointer to embedded type
		return p.resolveEmbeddedType(t.X)

	case *ast.SelectorExpr:
		// Qualified type (package.Type)
		if pkgIdent, ok := t.X.(*ast.Ident); ok {
			qualifiedName := pkgIdent.Name + "." + t.Sel.Name
			return p.resolveTypeName(qualifiedName)
		}
	}

	return nil
}

// resolveTypeName looks up a type name in the builder's registries
func (p *EmbeddedProcessor) resolveTypeName(typeName string) *spec.Schema {
	b := p.builder

	// Try to get from components schemas first (registered models)
	if b.spec.Components != nil && b.spec.Components.Schemas != nil {
		// Try exact name match
		if schema, ok := b.spec.Components.Schemas[typeName]; ok {
			// Ensure model is parsed if not yet
			if !b.parsedModels[typeName] && b.modelRegistry[typeName] != nil {
				_ = b.ensureModelParsed(typeName)
				return b.spec.Components.Schemas[typeName]
			}
			return schema
		}

		// Try without package prefix
		if idx := strings.LastIndex(typeName, "."); idx >= 0 {
			shortName := typeName[idx+1:]
			if schema, ok := b.spec.Components.Schemas[shortName]; ok {
				return schema
			}
		}
	}

	// Try type registry for untagged structs
	if structType, ok := b.typeRegistry[typeName]; ok {
		return b.parseStruct(structType)
	}

	// Try without package prefix in type registry
	if idx := strings.LastIndex(typeName, "."); idx >= 0 {
		shortName := typeName[idx+1:]
		if structType, ok := b.typeRegistry[shortName]; ok {
			return b.parseStruct(structType)
		}
	}

	return nil
}

// CopyFieldInfo creates a deep copy of a FieldInfo
func CopyFieldInfo(field *spec.FieldInfo) *spec.FieldInfo {
	if field == nil {
		return nil
	}

	newField := &spec.FieldInfo{
		Name:        field.Name,
		Type:        field.Type,
		Description: field.Description,
		Required:    field.Required,
		Example:     field.Example,
		Default:     field.Default,
		Enum:        field.Enum,
		IsArray:     field.IsArray,
		IsPointer:   field.IsPointer,
		IsMap:       field.IsMap,
		MapKeyType:  field.MapKeyType,
	}

	// Deep copy validations map
	if field.Validations != nil {
		newField.Validations = make(map[string]string)
		for k, v := range field.Validations {
			newField.Validations[k] = v
		}
	}

	return newField
}

