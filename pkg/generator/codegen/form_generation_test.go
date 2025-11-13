package codegen

import (
	"strings"
	"testing"

	"github.com/reation-io/apikit/pkg/generator/parser"
)

func TestGenerateExtractionCode_WithFileField(t *testing.T) {
	gen, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Create a struct with a file field
	s := &parser.Struct{
		Name: "UploadRequest",
		Fields: []parser.Field{
			{
				Name:      "Title",
				Type:      "string",
				StructTag: `form:"title"`,
			},
			{
				Name:      "Image",
				Type:      "*multipart.FileHeader",
				StructTag: `form:"image"`,
				IsFile:    true,
			},
		},
	}

	importsMap := make(map[string]bool)
	code := gen.generateExtractionCode(s, importsMap)

	t.Logf("Generated code:\n%s", code)
	t.Logf("Imports map: %v", importsMap)

	// Check if code contains FormFile
	if !strings.Contains(code, "FormFile") {
		t.Error("expected code to contain FormFile")
	}

	// Check if mime/multipart import was added
	if !importsMap["mime/multipart"] {
		t.Errorf("expected mime/multipart in imports map, got: %v", importsMap)
	}
}

func TestPrepareTemplateData_WithFileField(t *testing.T) {
	gen, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Create a parse result with a file field
	result := &parser.ParseResult{
		Source: parser.Source{
			Package: "test",
		},
		Handlers: []parser.Handler{
			{
				Name:       "UploadImage",
				ParamType:  "UploadImageRequest",
				ReturnType: "UploadImageResponse",
				Struct: &parser.Struct{
					Name: "UploadImageRequest",
					Fields: []parser.Field{
						{
							Name:      "Image",
							Type:      "*multipart.FileHeader",
							StructTag: `form:"image"`,
							IsFile:    true,
						},
					},
				},
			},
		},
	}

	data := gen.prepareTemplateData(result)

	t.Logf("Imports: %v", data.Imports)

	// Check if mime/multipart is in the imports
	hasMultipartImport := false
	for _, imp := range data.Imports {
		if imp == "mime/multipart" {
			hasMultipartImport = true
			break
		}
	}

	if !hasMultipartImport {
		t.Errorf("expected mime/multipart in imports, got: %v", data.Imports)
	}
}
