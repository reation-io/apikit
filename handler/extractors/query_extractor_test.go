package extractors

import (
	"strings"
	"testing"

	"github.com/reation-io/apikit/core/definition"
)

func TestQueryExtractor_Name(t *testing.T) {
	e := &QueryExtractor{}
	if e.Name() != "query" {
		t.Errorf("expected name 'query', got %q", e.Name())
	}
}

func TestQueryExtractor_Priority(t *testing.T) {
	e := &QueryExtractor{}
	if e.Priority() != 20 {
		t.Errorf("expected priority 20, got %d", e.Priority())
	}
}

func TestQueryExtractor_CanExtract(t *testing.T) {
	e := &QueryExtractor{}

	tests := []struct {
		name     string
		field    *definition.Field
		expected bool
	}{
		{
			name:     "with query tag",
			field:    &definition.Field{Tags: `query:"search"`},
			expected: true,
		},
		{
			name:     "with in:query comment",
			field:    &definition.Field{Metadata: map[string]string{"in": "query"}},
			expected: true,
		},
		{
			name:     "without query tag or comment",
			field:    &definition.Field{Tags: `json:"search"`},
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

func TestQueryExtractor_GenerateCode_SingleValue(t *testing.T) {
	e := &QueryExtractor{}

	field := &definition.Field{
		Name: "Search",
		Type: &definition.Type{GoType: "string", Kind: "primitive"},
		Tags: `query:"q"`,
	}

	code, _ := e.GenerateCode(field, "Request")

	expectedParts := []string{
		`r.URL.Query().Get("q")`,
		"payload.Search",
	}

	for _, expected := range expectedParts {
		if !strings.Contains(code, expected) {
			t.Errorf("expected code to contain %q, got:\n%s", expected, code)
		}
	}
}

func TestQueryExtractor_GenerateCode_Slice(t *testing.T) {
	e := &QueryExtractor{}

	field := &definition.Field{
		Name: "Tags",
		Type: &definition.Type{
			GoType:      "[]string",
			Kind:        "slice",
			ElementType: &definition.Type{GoType: "string", Kind: "primitive"},
		},
		Tags: `query:"tags"`,
	}

	code, _ := e.GenerateCode(field, "Request")

	expectedParts := []string{
		`r.URL.Query()["tags"]`,
		"payload.Tags",
	}

	for _, expected := range expectedParts {
		if !strings.Contains(code, expected) {
			t.Errorf("expected code to contain %q, got:\n%s", expected, code)
		}
	}
}

func TestQueryExtractor_GenerateCode_IntSlice(t *testing.T) {
	e := &QueryExtractor{}

	field := &definition.Field{
		Name: "IDs",
		Type: &definition.Type{
			GoType:      "[]int",
			Kind:        "slice",
			ElementType: &definition.Type{GoType: "int", Kind: "primitive"},
		},
		Tags: `query:"ids"`,
	}

	code, imports := e.GenerateCode(field, "Request")

	// Should contain parsing logic for int slice
	if !strings.Contains(code, "strconv.ParseInt") {
		t.Error("expected ParseInt for int slice")
	}

	// Should have strconv import
	found := false
	for _, imp := range imports {
		if imp == "strconv" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected strconv import for int slice")
	}
}
