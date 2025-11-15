package tags

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/spec"
)

func TestAllMetaParsers(t *testing.T) {
	// Create a comment with multiple meta tags
	src := `
package main

// swagger:meta
// Title: My API
// Version: 1.0.0
// Description: This is a comprehensive API
//   with multiple features
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

	// Create targets
	info := &spec.Info{}

	// Parse all meta tags
	err = parsers.GlobalRegistry().Parse("swagger:meta", comments, info, parsers.ContextMeta)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	// Verify Info
	if info.Title != "My API" {
		t.Errorf("expected title 'My API', got %q", info.Title)
	}
	if info.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", info.Version)
	}
	if info.Description != "This is a comprehensive API\nwith multiple features" {
		t.Errorf("expected description, got %q", info.Description)
	}
}

func TestAllRouteParsers(t *testing.T) {
	// Create a comment with multiple route tags
	src := `
package main

// swagger:route POST /users user createUser
// Summary: Create a new user
// Tags: users, admin
// Deprecated: false
// Consumes: application/json
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

	// Parse all route tags
	err = parsers.GlobalRegistry().Parse("swagger:route", comments, operation, parsers.ContextRoute)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	// Verify Operation
	if operation.Summary != "Create a new user" {
		t.Errorf("expected summary 'Create a new user', got %q", operation.Summary)
	}
	if len(operation.Tags) != 2 || operation.Tags[0] != "users" || operation.Tags[1] != "admin" {
		t.Errorf("expected tags [users, admin], got %v", operation.Tags)
	}
	if operation.Deprecated {
		t.Error("expected deprecated to be false")
	}
	if operation.RequestBody == nil {
		t.Fatal("expected RequestBody to be set")
	}
	if _, ok := operation.RequestBody.Content["application/json"]; !ok {
		t.Error("expected application/json in RequestBody.Content")
	}
}

func TestAllFieldParsers(t *testing.T) {
	// Create a comment with multiple field tags
	src := `
package main

type User struct {
	// Example: user@example.com
	// Format: email
	// MinLength: 5
	// MaxLength: 100
	// Pattern: ^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$
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

	// Parse all field tags
	err = parsers.GlobalRegistry().Parse("swagger:model", comments, schema, parsers.ContextField)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	// Verify Schema
	if schema.Example != "user@example.com" {
		t.Errorf("expected example 'user@example.com', got %v", schema.Example)
	}
	if schema.Format != "email" {
		t.Errorf("expected format 'email', got %q", schema.Format)
	}
	if schema.MinLength == nil || *schema.MinLength != 5 {
		t.Errorf("expected minLength 5, got %v", schema.MinLength)
	}
	if schema.MaxLength == nil || *schema.MaxLength != 100 {
		t.Errorf("expected maxLength 100, got %v", schema.MaxLength)
	}
	if schema.Pattern != "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$" {
		t.Errorf("expected pattern, got %q", schema.Pattern)
	}
}
