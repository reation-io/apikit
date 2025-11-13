package codegen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/reation-io/apikit/pkg/generator/parser"
)

func TestGenerate_WithMultipartForm(t *testing.T) {
	// Create a temporary test file with multipart form handler
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "handler.go")

	content := `package test

import (
	"context"
	"mime/multipart"
)

type UploadImageRequest struct {
	Title       string                  ` + "`" + `form:"title" validate:"required"` + "`" + `
	Description string                  ` + "`" + `form:"description"` + "`" + `
	Image       *multipart.FileHeader   ` + "`" + `form:"image" validate:"required"` + "`" + `
	Category    string                  ` + "`" + `form:"category"` + "`" + `
}

type UploadImageResponse struct {
	ID  string ` + "`" + `json:"id"` + "`" + `
	URL string ` + "`" + `json:"url"` + "`" + `
}

// apikit:handler
func UploadImage(ctx context.Context, req UploadImageRequest) (UploadImageResponse, error) {
	return UploadImageResponse{}, nil
}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Parse the file
	p := parser.New()
	result, err := p.ParseFile(testFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Debug: Check if Image field is detected as file
	if len(result.Handlers) > 0 && result.Handlers[0].Struct != nil {
		for _, field := range result.Handlers[0].Struct.Fields {
			if field.Name == "Image" {
				t.Logf("Image field: Type=%s, IsFile=%v", field.Type, field.IsFile)
			}
		}
	}

	// Generate code
	gen, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	code, err := gen.Generate(result)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	codeStr := string(code)

	// Verify generated code contains multipart parsing
	expectedElements := []string{
		"ParseMultipartForm",
		`r.FormValue("title")`,
		`r.FormValue("description")`,
		`r.FormValue("category")`,
		`r.FormFile("image")`,
		"payload.Title",
		"payload.Description",
		"payload.Image",
		"payload.Category",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(codeStr, expected) {
			t.Errorf("expected generated code to contain %q\nGenerated code:\n%s", expected, codeStr)
		}
	}

	// Note: mime/multipart import is not needed in generated code because
	// the type *multipart.FileHeader is only used in the original struct definition,
	// not in the generated wrapper code. The generated code only uses r.FormFile()
	// which returns the header, but doesn't need to import the package.
}

func TestGenerate_WithMultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "handler.go")

	content := `package test

import (
	"context"
	"mime/multipart"
)

type UploadMultipleRequest struct {
	Title string                    ` + "`" + `form:"title" validate:"required"` + "`" + `
	Files []*multipart.FileHeader   ` + "`" + `form:"files" validate:"required,min=1,max=10"` + "`" + `
}

type UploadMultipleResponse struct {
	Count int ` + "`" + `json:"count"` + "`" + `
}

// apikit:handler
func UploadMultiple(ctx context.Context, req UploadMultipleRequest) (UploadMultipleResponse, error) {
	return UploadMultipleResponse{}, nil
}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	p := parser.New()
	result, err := p.ParseFile(testFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	gen, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	code, err := gen.Generate(result)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	codeStr := string(code)

	// Verify multiple files handling
	if !strings.Contains(codeStr, "MultipartForm") {
		t.Error("expected code to use MultipartForm for multiple files")
	}

	if !strings.Contains(codeStr, `form.File["files"]`) {
		t.Error("expected code to access form.File for multiple files")
	}

	if !strings.Contains(codeStr, "payload.Files") {
		t.Error("expected code to assign to payload.Files")
	}
}
