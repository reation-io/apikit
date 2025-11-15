package extractors

import (
	"strings"
	"testing"

	"github.com/reation-io/apikit/handler/parser"
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
		field    *parser.Field
		expected bool
	}{
		{
			name:     "with path tag",
			field:    &parser.Field{StructTag: `path:"id"`},
			expected: true,
		},
		{
			name:     "with in:path comment",
			field:    &parser.Field{InComment: "path"},
			expected: true,
		},
		{
			name:     "without path tag or comment",
			field:    &parser.Field{StructTag: `json:"id"`},
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
		field          *parser.Field
		expectedInCode []string
	}{
		{
			name: "string field",
			field: &parser.Field{
				Name:      "UserID",
				Type:      "string",
				StructTag: `path:"userId"`,
			},
			expectedInCode: []string{
				`r.PathValue("userId")`,
				"payload.UserID",
			},
		},
		{
			name: "int field",
			field: &parser.Field{
				Name:      "ID",
				Type:      "int64",
				StructTag: `path:"id"`,
			},
			expectedInCode: []string{
				`r.PathValue("id")`,
				"strconv.ParseInt",
				"payload.ID",
			},
		},
		{
			name: "field with comment name",
			field: &parser.Field{
				Name:          "UserID",
				Type:          "string",
				InComment:     "path",
				InCommentName: "user_id",
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
	field := &parser.Field{
		Name:      "ID",
		Type:      "int",
		StructTag: `path:"id"`,
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
