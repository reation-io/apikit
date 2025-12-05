package builder

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/reation-io/apikit/core/definition"
	"github.com/reation-io/apikit/core/parser"
)

func TestMultipartFormDataGeneration(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "handler.go")

	content := `package test

import (
	"mime/multipart"
)

// UploadRequest represents a file upload request
// swagger:model
type UploadRequest struct {
	// in:form title
	Title string ` + "`" + `form:"title" validate:"required"` + "`" + `
	
	// in:form description
	Description string ` + "`" + `form:"description"` + "`" + `
	
	// in:form file
	File *multipart.FileHeader ` + "`" + `form:"file" validate:"required"` + "`" + `
}

// UploadResponse represents the upload response
// swagger:model
type UploadResponse struct {
	ID string ` + "`" + `json:"id"` + "`" + `
}

// swagger:route POST /upload files uploadFile
// Consumes: multipart/form-data
// Summary: Upload a file
// Responses:
// - 200: UploadResponse
type UploadFileRoute struct{}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	p := parser.New()
	def, err := p.ParseFile(testFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	openapi, err := ExtractFromGeneric([]*definition.Definition{def})
	if err != nil {
		t.Fatalf("ExtractFromGeneric failed: %v", err)
	}

	// Verify the /upload path exists
	if openapi.Paths == nil || openapi.Paths.PathItems["/upload"] == nil {
		t.Fatal("expected /upload path to exist")
	}

	pathItem := openapi.Paths.PathItems["/upload"]
	if pathItem.Post == nil {
		t.Fatal("expected POST operation on /upload")
	}

	operation := pathItem.Post

	// Verify requestBody exists
	if operation.RequestBody == nil {
		t.Fatal("expected requestBody to exist")
	}

	// Verify multipart/form-data content type
	mediaType := operation.RequestBody.Content["multipart/form-data"]
	if mediaType == nil {
		t.Fatal("expected multipart/form-data content type")
	}

	// Verify schema
	if mediaType.Schema == nil {
		t.Fatal("expected schema in multipart/form-data")
	}

	schema := mediaType.Schema
	t.Logf("Schema: %+v", schema)
	t.Logf("Properties: %+v", schema.Properties)
	if schema.Type != "object" {
		t.Errorf("expected schema type 'object', got '%s'", schema.Type)
	}

	// Verify properties
	if schema.Properties == nil {
		t.Fatal("expected properties in schema")
	}

	// Check title field
	titleProp := schema.Properties["title"]
	if titleProp == nil {
		t.Error("expected 'title' property")
	} else if titleProp.Type != "string" {
		t.Errorf("expected title type 'string', got '%s'", titleProp.Type)
	}

	// Check description field
	descProp := schema.Properties["description"]
	if descProp == nil {
		t.Error("expected 'description' property")
	} else if descProp.Type != "string" {
		t.Errorf("expected description type 'string', got '%s'", descProp.Type)
	}

	// Check file field
	fileProp := schema.Properties["file"]
	if fileProp == nil {
		t.Error("expected 'file' property")
	} else {
		if fileProp.Type != "string" {
			t.Errorf("expected file type 'string', got '%s'", fileProp.Type)
		}
		if fileProp.Format != "binary" {
			t.Errorf("expected file format 'binary', got '%s'", fileProp.Format)
		}
	}

	// Verify required fields
	if len(schema.Required) != 2 {
		t.Errorf("expected 2 required fields, got %d", len(schema.Required))
	}

	hasTitle := false
	hasFile := false
	for _, req := range schema.Required {
		if req == "title" {
			hasTitle = true
		}
		if req == "file" {
			hasFile = true
		}
	}

	if !hasTitle {
		t.Error("expected 'title' to be required")
	}
	if !hasFile {
		t.Error("expected 'file' to be required")
	}
}
