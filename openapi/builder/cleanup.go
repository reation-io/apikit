package builder

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/reation-io/apikit/openapi/spec"
)

// schemaRefPattern matches $ref references to schemas
var schemaRefPattern = regexp.MustCompile(`#/components/schemas/([^"]+)`)

// CleanupUnreferencedSchemas removes schemas that are not referenced anywhere in the spec
// This is useful to clean up the spec after filtering routes or for optimization
func (b *Builder) CleanupUnreferencedSchemas() {
	if b.spec.Components == nil || b.spec.Components.Schemas == nil {
		return
	}

	// Find all referenced schemas
	referenced := b.findReferencedSchemas()

	// Remove unreferenced schemas
	for schemaName := range b.spec.Components.Schemas {
		if !referenced[schemaName] {
			delete(b.spec.Components.Schemas, schemaName)
		}
	}
}

// findReferencedSchemas finds all schemas that are referenced in the spec
func (b *Builder) findReferencedSchemas() map[string]bool {
	referenced := make(map[string]bool)

	// Serialize spec to JSON to find all $ref occurrences
	data, err := json.Marshal(b.spec)
	if err != nil {
		// If marshaling fails, keep all schemas
		for name := range b.spec.Components.Schemas {
			referenced[name] = true
		}
		return referenced
	}

	// Find all schema references
	matches := schemaRefPattern.FindAllStringSubmatch(string(data), -1)
	for _, match := range matches {
		if len(match) > 1 {
			referenced[match[1]] = true
		}
	}

	// Also mark schemas that reference other schemas (transitive references)
	changed := true
	for changed {
		changed = false
		for schemaName := range referenced {
			if schema, ok := b.spec.Components.Schemas[schemaName]; ok {
				newRefs := findSchemaReferences(schema)
				for _, ref := range newRefs {
					if !referenced[ref] {
						referenced[ref] = true
						changed = true
					}
				}
			}
		}
	}

	return referenced
}

// findSchemaReferences finds all schema names referenced by a schema
func findSchemaReferences(schema *spec.Schema) []string {
	if schema == nil {
		return nil
	}

	var refs []string

	// Check $ref
	if schema.Ref != "" {
		if name := extractSchemaName(schema.Ref); name != "" {
			refs = append(refs, name)
		}
	}

	// Check properties
	for _, propSchema := range schema.Properties {
		refs = append(refs, findSchemaReferences(propSchema)...)
	}

	// Check items
	if schema.Items != nil {
		refs = append(refs, findSchemaReferences(schema.Items)...)
	}

	// Check allOf, oneOf, anyOf
	for _, s := range schema.AllOf {
		refs = append(refs, findSchemaReferences(s)...)
	}
	for _, s := range schema.OneOf {
		refs = append(refs, findSchemaReferences(s)...)
	}
	for _, s := range schema.AnyOf {
		refs = append(refs, findSchemaReferences(s)...)
	}

	// Check additionalProperties if it's a schema
	if addProps, ok := schema.AdditionalProperties.(*spec.Schema); ok {
		refs = append(refs, findSchemaReferences(addProps)...)
	}

	return refs
}

// extractSchemaName extracts the schema name from a $ref path
func extractSchemaName(ref string) string {
	prefix := "#/components/schemas/"
	if strings.HasPrefix(ref, prefix) {
		return strings.TrimPrefix(ref, prefix)
	}
	return ""
}

// CleanupSpec performs all cleanup operations on the spec
func (b *Builder) CleanupSpec() {
	b.CleanupUnreferencedSchemas()
}

