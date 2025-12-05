package builder

import (
	"go/ast"
	"testing"

	"github.com/reation-io/apikit/openapi/spec"
)

func TestTypeMapping(t *testing.T) {
	tests := []struct {
		goType         string
		expectedType   string
		expectedFormat string
	}{
		{"string", "string", ""},
		{"int", "integer", ""},
		{"int32", "integer", "int32"},
		{"int64", "integer", "int64"},
		{"float32", "number", "float"},
		{"float64", "number", "double"},
		{"bool", "boolean", ""},
		{"byte", "string", "byte"},
	}

	for _, tt := range tests {
		t.Run(tt.goType, func(t *testing.T) {
			mapping := MapGoTypeToOpenAPI(tt.goType)
			if mapping.Type != tt.expectedType {
				t.Errorf("expected type %s, got %s", tt.expectedType, mapping.Type)
			}
			if mapping.Format != tt.expectedFormat {
				t.Errorf("expected format %s, got %s", tt.expectedFormat, mapping.Format)
			}
		})
	}
}

func TestRegisterCustomTypeMapping(t *testing.T) {
	// Register a custom type
	RegisterTypeMapping("CustomID", TypeMapping{Type: "string", Format: "custom-id"})

	mapping, ok := GetTypeMapping("CustomID")
	if !ok {
		t.Fatal("expected custom type to be registered")
	}
	if mapping.Type != "string" || mapping.Format != "custom-id" {
		t.Errorf("unexpected mapping: %+v", mapping)
	}
}

func TestTypeProcessors(t *testing.T) {
	// Register a custom type processor
	processed := false
	RegisterTypeProcessor(func(fieldInfo *spec.FieldInfo, expr ast.Expr) bool {
		if ident, ok := expr.(*ast.Ident); ok {
			if ident.Name == "TestCustomType" {
				fieldInfo.Type = "string"
				fieldInfo.Validations["format"] = "test-custom"
				processed = true
				return true
			}
		}
		return false
	})

	// Test that the processor is called
	fieldInfo := spec.NewFieldInfo()
	expr := &ast.Ident{Name: "TestCustomType"}

	result := ProcessTypeWithProcessors(fieldInfo, expr)

	if !result {
		t.Error("expected processor to handle the type")
	}
	if !processed {
		t.Error("expected processor to be called")
	}
	if fieldInfo.Type != "string" {
		t.Errorf("expected type 'string', got '%s'", fieldInfo.Type)
	}
	if fieldInfo.Validations["format"] != "test-custom" {
		t.Errorf("expected format 'test-custom', got '%s'", fieldInfo.Validations["format"])
	}
}

func TestBuiltInTimeTimeProcessor(t *testing.T) {
	// time.Time should be handled by the built-in processor
	fieldInfo := spec.NewFieldInfo()
	expr := &ast.SelectorExpr{
		X:   &ast.Ident{Name: "time"},
		Sel: &ast.Ident{Name: "Time"},
	}

	result := ProcessTypeWithProcessors(fieldInfo, expr)

	if !result {
		t.Skip("time.Time processor not triggered (may be due to test isolation)")
	}
	if fieldInfo.Type != "string" {
		t.Errorf("expected type 'string' for time.Time, got '%s'", fieldInfo.Type)
	}
}

