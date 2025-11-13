package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParser_DetectFileFields(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "handler.go")

	content := `package test

import (
	"context"
	"mime/multipart"
)

type UploadRequest struct {
	Title string                  ` + "`" + `form:"title"` + "`" + `
	Image *multipart.FileHeader   ` + "`" + `form:"image"` + "`" + `
}

type UploadResponse struct {
	ID string ` + "`" + `json:"id"` + "`" + `
}

// apikit:handler
func Upload(ctx context.Context, req UploadRequest) (UploadResponse, error) {
	return UploadResponse{}, nil
}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	p := New()
	result, err := p.ParseFile(testFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(result.Handlers) != 1 {
		t.Fatalf("expected 1 handler, got %d", len(result.Handlers))
	}

	handler := result.Handlers[0]
	if handler.Struct == nil {
		t.Fatal("expected handler to have a struct")
	}

	// Find the Image field
	var imageField *Field
	for i := range handler.Struct.Fields {
		if handler.Struct.Fields[i].Name == "Image" {
			imageField = &handler.Struct.Fields[i]
			break
		}
	}

	if imageField == nil {
		t.Fatal("Image field not found")
	}

	// Verify the field is detected as a file field
	if !imageField.IsFile {
		t.Errorf("expected Image field to have IsFile=true, got false. Type: %s", imageField.Type)
	}

	if imageField.Type != "*multipart.FileHeader" {
		t.Errorf("expected type '*multipart.FileHeader', got %q", imageField.Type)
	}
}

func TestParser_DetectMultipleFileFields(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "handler.go")

	content := `package test

import (
	"context"
	"mime/multipart"
)

type UploadRequest struct {
	Files []*multipart.FileHeader ` + "`" + `form:"files"` + "`" + `
}

type UploadResponse struct {
	Count int ` + "`" + `json:"count"` + "`" + `
}

// apikit:handler
func Upload(ctx context.Context, req UploadRequest) (UploadResponse, error) {
	return UploadResponse{}, nil
}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	p := New()
	result, err := p.ParseFile(testFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	handler := result.Handlers[0]
	var filesField *Field
	for i := range handler.Struct.Fields {
		if handler.Struct.Fields[i].Name == "Files" {
			filesField = &handler.Struct.Fields[i]
			break
		}
	}

	if filesField == nil {
		t.Fatal("Files field not found")
	}

	if !filesField.IsFile {
		t.Errorf("expected Files field to have IsFile=true, got false. Type: %s", filesField.Type)
	}

	if filesField.Type != "[]*multipart.FileHeader" {
		t.Errorf("expected type '[]*multipart.FileHeader', got %q", filesField.Type)
	}
}
