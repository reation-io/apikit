package extractors

import (
	"strings"
	"testing"

	"github.com/reation-io/apikit/pkg/generator/parser"
)

func TestFormExtractor_Name(t *testing.T) {
	e := &FormExtractor{}
	if e.Name() != "form" {
		t.Errorf("expected name 'form', got %q", e.Name())
	}
}

func TestFormExtractor_Priority(t *testing.T) {
	e := &FormExtractor{}
	if e.Priority() != 15 {
		t.Errorf("expected priority 15, got %d", e.Priority())
	}
}

func TestFormExtractor_CanExtract(t *testing.T) {
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
			name: "field with form comment",
			field: parser.Field{
				Name:      "Title",
				Type:      "string",
				InComment: "form",
			},
			expected: true,
		},
		{
			name: "field without form tag",
			field: parser.Field{
				Name:      "Title",
				Type:      "string",
				StructTag: `json:"title"`,
			},
			expected: false,
		},
		{
			name: "ResponseWriter field",
			field: parser.Field{
				Name:             "W",
				Type:             "http.ResponseWriter",
				IsResponseWriter: true,
			},
			expected: false,
		},
	}

	e := &FormExtractor{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := e.CanExtract(&tt.field)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFormExtractor_GenerateCode_StringField(t *testing.T) {
	e := &FormExtractor{}
	field := &parser.Field{
		Name:      "Title",
		Type:      "string",
		StructTag: `form:"title"`,
	}

	code, imports := e.GenerateCode(field, "TestRequest")

	if !strings.Contains(code, `r.FormValue("title")`) {
		t.Errorf("expected code to contain FormValue, got: %s", code)
	}

	if !strings.Contains(code, "payload.Title") {
		t.Errorf("expected code to assign to payload.Title, got: %s", code)
	}

	if len(imports) > 0 {
		t.Logf("imports: %v", imports)
	}
}

func TestFormExtractor_GenerateCode_FileField(t *testing.T) {
	e := &FormExtractor{}
	field := &parser.Field{
		Name:      "Image",
		Type:      "*multipart.FileHeader",
		StructTag: `form:"image"`,
		IsFile:    true,
	}

	code, imports := e.GenerateCode(field, "UploadRequest")

	if !strings.Contains(code, `r.FormFile("image")`) {
		t.Errorf("expected code to contain FormFile, got: %s", code)
	}

	if !strings.Contains(code, "payload.Image") {
		t.Errorf("expected code to assign to payload.Image, got: %s", code)
	}

	hasMultipartImport := false
	for _, imp := range imports {
		if imp == "mime/multipart" {
			hasMultipartImport = true
			break
		}
	}

	if !hasMultipartImport {
		t.Error("expected mime/multipart import")
	}
}

func TestFormExtractor_GenerateCode_MultipleFiles(t *testing.T) {
	e := &FormExtractor{}
	field := &parser.Field{
		Name:      "Files",
		Type:      "[]*multipart.FileHeader",
		StructTag: `form:"files"`,
		IsFile:    true,
		IsSlice:   true,
	}

	code, imports := e.GenerateCode(field, "UploadRequest")

	if !strings.Contains(code, "MultipartForm") {
		t.Errorf("expected code to use MultipartForm, got: %s", code)
	}

	if !strings.Contains(code, `form.File["files"]`) {
		t.Errorf("expected code to access form.File, got: %s", code)
	}

	hasMultipartImport := false
	for _, imp := range imports {
		if imp == "mime/multipart" {
			hasMultipartImport = true
			break
		}
	}

	if !hasMultipartImport {
		t.Error("expected mime/multipart import")
	}
}
