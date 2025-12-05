package builder

import (
	"go/ast"
	"reflect"
	"strings"
	"sync"

	"github.com/reation-io/apikit/openapi/spec"
)

// TypeProcessor is a function that processes a specific type and updates the FieldInfo
// Returns true if the type was processed, false otherwise
type TypeProcessor func(fieldInfo *spec.FieldInfo, expr ast.Expr) bool

// TypeMapping maps a Go type to OpenAPI type and format
type TypeMapping struct {
	Type   string
	Format string
}

// Global registry of type processors and mappings
var (
	typeProcessors      []TypeProcessor
	typeProcessorsMutex sync.RWMutex

	typeMappings      = make(map[string]TypeMapping)
	typeMappingsMutex sync.RWMutex
)

func init() {
	// Register default type mappings
	RegisterTypeMapping("string", TypeMapping{Type: "string"})
	RegisterTypeMapping("int", TypeMapping{Type: "integer"})
	RegisterTypeMapping("int8", TypeMapping{Type: "integer", Format: "int8"})
	RegisterTypeMapping("int16", TypeMapping{Type: "integer", Format: "int16"})
	RegisterTypeMapping("int32", TypeMapping{Type: "integer", Format: "int32"})
	RegisterTypeMapping("int64", TypeMapping{Type: "integer", Format: "int64"})
	RegisterTypeMapping("uint", TypeMapping{Type: "integer"})
	RegisterTypeMapping("uint8", TypeMapping{Type: "integer", Format: "uint8"})
	RegisterTypeMapping("uint16", TypeMapping{Type: "integer", Format: "uint16"})
	RegisterTypeMapping("uint32", TypeMapping{Type: "integer", Format: "uint32"})
	RegisterTypeMapping("uint64", TypeMapping{Type: "integer", Format: "uint64"})
	RegisterTypeMapping("float32", TypeMapping{Type: "number", Format: "float"})
	RegisterTypeMapping("float64", TypeMapping{Type: "number", Format: "double"})
	RegisterTypeMapping("bool", TypeMapping{Type: "boolean"})
	RegisterTypeMapping("byte", TypeMapping{Type: "string", Format: "byte"})

	// Register time.Time as string with date-time format
	RegisterTypeHandler("time.Time", func(f *spec.FieldInfo) {
		f.Type = "string"
		f.Validations["format"] = "date-time"
	})

	// Register uuid.UUID as string with uuid format
	RegisterTypeHandler("uuid.UUID", func(f *spec.FieldInfo) {
		f.Type = "string"
		f.Validations["format"] = "uuid"
	})
}

// RegisterTypeProcessor registers a custom type processor
func RegisterTypeProcessor(processor TypeProcessor) {
	typeProcessorsMutex.Lock()
	defer typeProcessorsMutex.Unlock()
	typeProcessors = append(typeProcessors, processor)
}

// RegisterTypeMapping registers a type mapping for a Go type name
func RegisterTypeMapping(goType string, mapping TypeMapping) {
	typeMappingsMutex.Lock()
	defer typeMappingsMutex.Unlock()
	typeMappings[goType] = mapping
}

// RegisterTypeHandler registers a handler for a specific type (by example)
// This allows registering handlers using type instances like time.Time{}
func RegisterTypeHandler(typeName string, handler func(fieldInfo *spec.FieldInfo)) {
	RegisterTypeProcessor(func(fieldInfo *spec.FieldInfo, expr ast.Expr) bool {
		var name string
		switch t := expr.(type) {
		case *ast.Ident:
			name = t.Name
		case *ast.SelectorExpr:
			if ident, ok := t.X.(*ast.Ident); ok {
				name = ident.Name + "." + t.Sel.Name
			}
		}

		if name == typeName {
			handler(fieldInfo)
			return true
		}
		return false
	})
}

// RegisterTypeFromExample registers a type handler using a type instance
func RegisterTypeFromExample(example any, handler func(fieldInfo *spec.FieldInfo)) {
	t := reflect.TypeOf(example)
	typeName := t.String()

	// Remove package path if present
	if idx := strings.LastIndex(typeName, "/"); idx >= 0 {
		typeName = typeName[idx+1:]
	}

	RegisterTypeHandler(typeName, handler)
}

// GetTypeMapping returns the OpenAPI type mapping for a Go type
func GetTypeMapping(goType string) (TypeMapping, bool) {
	typeMappingsMutex.RLock()
	defer typeMappingsMutex.RUnlock()
	mapping, ok := typeMappings[goType]
	return mapping, ok
}

// ProcessTypeWithProcessors runs all registered type processors on a field
// Returns true if any processor handled the type
func ProcessTypeWithProcessors(fieldInfo *spec.FieldInfo, expr ast.Expr) bool {
	typeProcessorsMutex.RLock()
	processors := typeProcessors
	typeProcessorsMutex.RUnlock()

	for _, processor := range processors {
		if processor(fieldInfo, expr) {
			return true
		}
	}
	return false
}

// MapGoTypeToOpenAPI maps a Go type name to OpenAPI type and format
func MapGoTypeToOpenAPI(goType string) TypeMapping {
	if mapping, ok := GetTypeMapping(goType); ok {
		return mapping
	}
	// Default to object for unknown types
	return TypeMapping{Type: "object"}
}

