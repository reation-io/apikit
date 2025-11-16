package extractors

import (
	"testing"

	"github.com/reation-io/apikit/handler/parser"
)

func TestFormExtractor_Name(t *testing.T) {
	e := &FormExtractor{}
	if e.Name() != "form" {
		t.Errorf("expected name 'form', got '%s'", e.Name())
	}
}

func TestFormExtractor_Priority(t *testing.T) {
	e := &FormExtractor{}
	if e.Priority() != 15 {
		t.Errorf("expected priority 15, got %d", e.Priority())
	}
}

func TestFormExtractor_CanExtract(t *testing.T) {
	e := &FormExtractor{}

	tests := []struct {
		name     string
		field    parser.Field
		expected bool
	}{
		{
			name: "field with form tag",
			field: parser.Field{
				Name:      "Title",
				Type:      "string",
				StructTag: `form:"title"`,
			},
			expected: true,
		},
		{
			name: "field with in:form comment",
			field: parser.Field{
				Name:      "Description",
				Type:      "string",
				InComment: "form",
			},
			expected: true,
		},
		{
			name: "file field",
			field: parser.Field{
				Name:      "Image",
				Type:      "*multipart.FileHeader",
				StructTag: `form:"image"`,
				IsFile:    true,
			},
			expected: true,
		},
		{
			name: "field without form tag or comment",
			field: parser.Field{
				Name: "Other",
				Type: "string",
			},
			expected: false,
		},
		{
			name: "request field",
			field: parser.Field{
				Name:      "Request",
				Type:      "*http.Request",
				IsRequest: true,
			},
			expected: false,
		},
		{
			name: "response writer field",
			field: parser.Field{
				Name:             "Writer",
				Type:             "http.ResponseWriter",
				IsResponseWriter: true,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := e.CanExtract(&tt.field)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFormExtractor_GenerateCode_RegularField(t *testing.T) {
	e := &FormExtractor{}

	field := &parser.Field{
		Name:      "Title",
		Type:      "string",
		StructTag: `form:"title"`,
	}

	code, imports := e.GenerateCode(field, "TestStruct")

	if code == "" {
		t.Error("expected code to be generated")
	}

	if len(imports) == 0 {
		t.Log("No imports needed for regular string field (expected)")
	}

	t.Logf("Generated code:\n%s", code)
	t.Logf("Imports: %v", imports)
}

func TestFormExtractor_GenerateCode_FileField(t *testing.T) {
	e := &FormExtractor{}

	field := &parser.Field{
		Name:      "Image",
		Type:      "*multipart.FileHeader",
		StructTag: `form:"image"`,
		IsFile:    true,
	}

	code, imports := e.GenerateCode(field, "TestStruct")

	if code == "" {
		t.Error("expected code to be generated")
	}

	// Should import mime/multipart for file fields
	hasMultipartImport := false
	for _, imp := range imports {
		if imp == "mime/multipart" {
			hasMultipartImport = true
			break
		}
	}

	if !hasMultipartImport {
		t.Errorf("expected mime/multipart import, got: %v", imports)
	}

	t.Logf("Generated code:\n%s", code)
}

