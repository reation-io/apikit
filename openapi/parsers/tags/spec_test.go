package tags

import (
	"go/ast"
	"testing"

	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/spec"
)

func TestSpecParser_Route_SingleSpec(t *testing.T) {
	parser := NewSpecParser()
	operation := &spec.Operation{}

	comment := &ast.CommentGroup{
		List: []*ast.Comment{
			{Text: "// Spec: admin"},
		},
	}

	value, err := parser.Parse(comment, parsers.ContextRoute)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if err := parser.Apply(operation, value, parsers.ContextRoute); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	specs, ok := operation.Extensions["x-specs"].([]string)
	if !ok {
		t.Fatalf("Expected []string in Extensions[x-specs], got %T", operation.Extensions["x-specs"])
	}

	if len(specs) != 1 || specs[0] != "admin" {
		t.Errorf("Expected [admin], got %v", specs)
	}
}

func TestSpecParser_Route_MultipleSpecs_SpaceSeparated(t *testing.T) {
	parser := NewSpecParser()
	operation := &spec.Operation{}

	comment := &ast.CommentGroup{
		List: []*ast.Comment{
			{Text: "// Spec: admin mobile public"},
		},
	}

	value, err := parser.Parse(comment, parsers.ContextRoute)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if err := parser.Apply(operation, value, parsers.ContextRoute); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	specs, ok := operation.Extensions["x-specs"].([]string)
	if !ok {
		t.Fatalf("Expected []string in Extensions[x-specs], got %T", operation.Extensions["x-specs"])
	}

	expected := []string{"admin", "mobile", "public"}
	if len(specs) != len(expected) {
		t.Fatalf("Expected %d specs, got %d", len(expected), len(specs))
	}

	for i, exp := range expected {
		if specs[i] != exp {
			t.Errorf("Expected specs[%d] = %s, got %s", i, exp, specs[i])
		}
	}
}

func TestSpecParser_Route_MultipleSpecs_CommaSeparated(t *testing.T) {
	parser := NewSpecParser()
	operation := &spec.Operation{}

	comment := &ast.CommentGroup{
		List: []*ast.Comment{
			{Text: "// Spec: admin, mobile, public"},
		},
	}

	value, err := parser.Parse(comment, parsers.ContextRoute)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if err := parser.Apply(operation, value, parsers.ContextRoute); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	specs, ok := operation.Extensions["x-specs"].([]string)
	if !ok {
		t.Fatalf("Expected []string in Extensions[x-specs], got %T", operation.Extensions["x-specs"])
	}

	expected := []string{"admin", "mobile", "public"}
	if len(specs) != len(expected) {
		t.Fatalf("Expected %d specs, got %d", len(expected), len(specs))
	}

	for i, exp := range expected {
		if specs[i] != exp {
			t.Errorf("Expected specs[%d] = %s, got %s", i, exp, specs[i])
		}
	}
}

func TestSpecParser_Route_CaseNormalization(t *testing.T) {
	parser := NewSpecParser()
	operation := &spec.Operation{}

	comment := &ast.CommentGroup{
		List: []*ast.Comment{
			{Text: "// Spec: Admin MOBILE Public"},
		},
	}

	value, err := parser.Parse(comment, parsers.ContextRoute)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if err := parser.Apply(operation, value, parsers.ContextRoute); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	specs, ok := operation.Extensions["x-specs"].([]string)
	if !ok {
		t.Fatalf("Expected []string in Extensions[x-specs], got %T", operation.Extensions["x-specs"])
	}

	// All should be normalized to lowercase
	expected := []string{"admin", "mobile", "public"}
	for i, exp := range expected {
		if specs[i] != exp {
			t.Errorf("Expected specs[%d] = %s (lowercase), got %s", i, exp, specs[i])
		}
	}
}

func TestSpecParser_Meta_SingleSpec(t *testing.T) {
	parser := NewSpecParser()
	info := &spec.Info{}

	comment := &ast.CommentGroup{
		List: []*ast.Comment{
			{Text: "// Spec: admin"},
		},
	}

	value, err := parser.Parse(comment, parsers.ContextMeta)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if err := parser.Apply(info, value, parsers.ContextMeta); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	specs, ok := info.Extensions["x-specs"].([]string)
	if !ok {
		t.Fatalf("Expected []string in Extensions[x-specs], got %T", info.Extensions["x-specs"])
	}

	if len(specs) != 1 || specs[0] != "admin" {
		t.Errorf("Expected [admin], got %v", specs)
	}
}

func TestSpecParser_ValidSpecNames(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "alphanumeric",
			input:    "admin123 mobile456",
			expected: []string{"admin123", "mobile456"},
		},
		{
			name:     "with hyphens",
			input:    "admin-api mobile-app",
			expected: []string{"admin-api", "mobile-app"},
		},
		{
			name:     "with underscores",
			input:    "admin_api mobile_app",
			expected: []string{"admin_api", "mobile_app"},
		},
		{
			name:     "mixed",
			input:    "admin-api_v1 mobile_app-v2",
			expected: []string{"admin-api_v1", "mobile_app-v2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			specs := parseSpecNames(tt.input)
			if len(specs) != len(tt.expected) {
				t.Fatalf("Expected %d specs, got %d", len(tt.expected), len(specs))
			}
			for i, exp := range tt.expected {
				if specs[i] != exp {
					t.Errorf("Expected specs[%d] = %s, got %s", i, exp, specs[i])
				}
			}
		})
	}
}

func TestSpecParser_InvalidSpecNames_Skipped(t *testing.T) {
	// Invalid characters should be skipped silently
	input := "admin valid-name @invalid mobile"
	specs := parseSpecNames(input)

	// Should only get valid names
	expected := []string{"admin", "valid-name", "mobile"}
	if len(specs) != len(expected) {
		t.Fatalf("Expected %d specs (invalid skipped), got %d: %v", len(expected), len(specs), specs)
	}

	for i, exp := range expected {
		if specs[i] != exp {
			t.Errorf("Expected specs[%d] = %s, got %s", i, exp, specs[i])
		}
	}
}
