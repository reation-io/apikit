package tags

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/spec"
)

func TestDescriptionParser_Meta(t *testing.T) {
	// Create a comment with description
	src := `
package main

// Description: This is a test API
// for testing purposes
type API struct{}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse file: %v", err)
	}

	// Get the comment group
	var comments *ast.CommentGroup
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			if genDecl.Doc != nil {
				comments = genDecl.Doc
				break
			}
		}
	}

	if comments == nil {
		t.Fatal("no comments found")
	}

	// Create target
	info := &spec.Info{}

	// Parse
	descParser := NewDescriptionParser()
	value, err := descParser.Parse(comments, parsers.ContextMeta)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	// Apply
	err = descParser.Apply(info, value, parsers.ContextMeta)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	// Verify
	expected := "This is a test API\nfor testing purposes"
	if info.Description != expected {
		t.Errorf("expected description %q, got %q", expected, info.Description)
	}
}

func TestDescriptionParser_Route(t *testing.T) {
	// Create a comment with description
	src := `
package main

// Description: Creates a new user
type CreateUser struct{}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse file: %v", err)
	}

	// Get the comment group
	var comments *ast.CommentGroup
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			if genDecl.Doc != nil {
				comments = genDecl.Doc
				break
			}
		}
	}

	if comments == nil {
		t.Fatal("no comments found")
	}

	// Create target
	operation := &spec.Operation{}

	// Parse
	descParser := NewDescriptionParser()
	value, err := descParser.Parse(comments, parsers.ContextRoute)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	// Apply
	err = descParser.Apply(operation, value, parsers.ContextRoute)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	// Verify
	expected := "Creates a new user"
	if operation.Description != expected {
		t.Errorf("expected description %q, got %q", expected, operation.Description)
	}
}

func TestDescriptionParser_Field(t *testing.T) {
	// Create a comment with description
	src := `
package main

type User struct {
	// Description: The user's email address
	Email string
}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse file: %v", err)
	}

	// Get the comment group from the field
	var comments *ast.CommentGroup
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if structType, ok := typeSpec.Type.(*ast.StructType); ok {
						if len(structType.Fields.List) > 0 {
							comments = structType.Fields.List[0].Doc
							break
						}
					}
				}
			}
		}
	}

	if comments == nil {
		t.Fatal("no comments found")
	}

	// Create target
	schema := &spec.Schema{}

	// Parse
	descParser := NewDescriptionParser()
	value, err := descParser.Parse(comments, parsers.ContextField)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	// Apply
	err = descParser.Apply(schema, value, parsers.ContextField)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	// Verify
	expected := "The user's email address"
	if schema.Description != expected {
		t.Errorf("expected description %q, got %q", expected, schema.Description)
	}
}
