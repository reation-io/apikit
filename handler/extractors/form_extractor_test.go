package extractors

import (
	"testing"

	"github.com/reation-io/apikit/core/definition"
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
		field    definition.Field
		expected bool
	}{
		{
			name: "field with form tag",
			field: definition.Field{
				Name: "Title",
				Type: &definition.Type{GoType: "string", Kind: "primitive"},
				Tags: `form:"title"`,
			},
			expected: true,
		},
		{
			name: "field with in:form comment",
			field: definition.Field{
				Name:     "Description",
				Type:     &definition.Type{GoType: "string", Kind: "primitive"},
				Metadata: map[string]string{"in": "form"},
			},
			expected: true,
		},
		{
			name: "file field",
			field: definition.Field{
				Name: "Image",
				Type: &definition.Type{GoType: "*multipart.FileHeader", Kind: "pointer"},
				Tags: `form:"image"`,
			},
			expected: true,
		},
		{
			name: "field without form tag or comment",
			field: definition.Field{
				Name: "Other",
				Type: &definition.Type{GoType: "string", Kind: "primitive"},
			},
			expected: false,
		},
		{
			name: "request field",
			field: definition.Field{
				Name: "Request",
				Type: &definition.Type{GoType: "*http.Request", Kind: "pointer"},
			},
			expected: false,
		},
		{
			name: "response writer field",
			field: definition.Field{
				Name: "Writer",
				Type: &definition.Type{GoType: "http.ResponseWriter", Kind: "interface"},
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

	field := &definition.Field{
		Name: "Title",
		Type: &definition.Type{GoType: "string", Kind: "primitive"},
		Tags: `form:"title"`,
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

	field := &definition.Field{
		Name: "Image",
		Type: &definition.Type{GoType: "*multipart.FileHeader", Kind: "pointer"},
		Tags: `form:"image"`,
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
