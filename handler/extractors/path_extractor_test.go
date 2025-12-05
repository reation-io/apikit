package extractors

import (
	"strings"
	"testing"

	"github.com/reation-io/apikit/core/definition"
)

func TestPathExtractor_Name(t *testing.T) {
	e := &PathExtractor{}
	if e.Name() != "path" {
		t.Errorf("expected name 'path', got %q", e.Name())
	}
}

func TestPathExtractor_Priority(t *testing.T) {
	e := &PathExtractor{}
	if e.Priority() != 10 {
		t.Errorf("expected priority 10, got %d", e.Priority())
	}
}

func TestPathExtractor_CanExtract(t *testing.T) {
	e := &PathExtractor{}

	tests := []struct {
		name     string
		field    *definition.Field
		expected bool
	}{
		{
			name:     "with path tag",
			field:    &definition.Field{Tags: `path:"id"`},
			expected: true,
		},
		{
			name:     "with in:path comment",
			field:    &definition.Field{Metadata: map[string]any{"in": "path"}},
			expected: true,
		},
		{
			name:     "without path tag or comment",
			field:    &definition.Field{Tags: `json:"id"`},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := e.CanExtract(tt.field)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPathExtractor_GenerateCode(t *testing.T) {
	e := &PathExtractor{}

	tests := []struct {
		name           string
		field          *definition.Field
		expectedInCode []string
	}{
		{
			name: "string field",
			field: &definition.Field{
				Name: "UserID",
				Type: &definition.Type{GoType: "string", Kind: "primitive"},
				Tags: `path:"userId"`,
			},
			expectedInCode: []string{
				`r.PathValue("userId")`,
				"payload.UserID",
			},
		},
		{
			name: "int field",
			field: &definition.Field{
				Name: "ID",
				Type: &definition.Type{GoType: "int64", Kind: "primitive"},
				Tags: `path:"id"`,
			},
			expectedInCode: []string{
				`r.PathValue("id")`,
				"strconv.ParseInt",
				"payload.ID",
			},
		},
		{
			name: "field with comment name",
			field: &definition.Field{
				Name:     "UserID",
				Type:     &definition.Type{GoType: "string", Kind: "primitive"},
				Metadata: map[string]any{"in": "path", "in_name": "user_id"},
			},
			expectedInCode: []string{
				`r.PathValue("user_id")`,
				"payload.UserID",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _ := e.GenerateCode(tt.field, "Request")

			for _, expected := range tt.expectedInCode {
				if !strings.Contains(code, expected) {
					t.Errorf("expected code to contain %q, got:\n%s", expected, code)
				}
			}
		})
	}
}

func TestPathExtractor_GenerateCode_Imports(t *testing.T) {
	e := &PathExtractor{}

	// Int field should require strconv import
	field := &definition.Field{
		Name: "ID",
		Type: &definition.Type{GoType: "int", Kind: "primitive"},
		Tags: `path:"id"`,
	}

	_, imports := e.GenerateCode(field, "Request")

	found := false
	for _, imp := range imports {
		if imp == "strconv" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected strconv import for int field")
	}
}
