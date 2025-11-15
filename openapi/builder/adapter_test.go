package builder

import (
	"os"
	"path/filepath"
	"testing"

	coreast "github.com/reation-io/apikit/core/ast"
)

func TestExtractFromGeneric(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package test

// swagger:meta
type Meta struct{}

// Title: My API
// Version: 1.0.0
// Description: This is my API

// User represents a user in the system
// swagger:model
type User struct {
	// ID is the user identifier
	ID int ` + "`json:\"id\"`" + `

	// Name is the user's name
	Name string ` + "`json:\"name\"`" + `

	// Email is the user's email
	Email string ` + "`json:\"email\"`" + `
}

// CreateUserRequest is the request for creating a user
// swagger:route POST /users user createUser
// Summary: Create a new user
// Description: Creates a new user in the system
type CreateUserRequest struct {
	Name  string ` + "`json:\"name\"`" + `
	Email string ` + "`json:\"email\"`" + `
}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Parse with generic parser
	genericParser := coreast.New()
	genericResult, err := genericParser.Parse(testFile)
	if err != nil {
		t.Fatalf("generic parse failed: %v", err)
	}

	// Extract OpenAPI-specific information
	openapi, err := ExtractFromGeneric([]*coreast.ParseResult{genericResult})
	if err != nil {
		t.Fatalf("ExtractFromGeneric failed: %v", err)
	}

	// Check OpenAPI version
	if openapi.OpenAPI != "3.0.3" {
		t.Errorf("expected OpenAPI version '3.0.3', got %q", openapi.OpenAPI)
	}

	// Check Info (should have defaults)
	if openapi.Info == nil {
		t.Fatal("expected Info to be set")
	}

	// Check that we have paths
	if openapi.Paths == nil {
		t.Fatal("expected Paths to be set")
	}

	// Check for the POST /users route
	pathItem, ok := openapi.Paths.PathItems["/users"]
	if !ok {
		t.Fatal("expected /users path to exist")
	}

	if pathItem.Post == nil {
		t.Fatal("expected POST operation on /users")
	}

	// Check operation details
	post := pathItem.Post
	if post.OperationID != "createUser" {
		t.Errorf("expected operation ID 'createUser', got %q", post.OperationID)
	}

	if len(post.Tags) != 1 || post.Tags[0] != "user" {
		t.Errorf("expected tags ['user'], got %v", post.Tags)
	}

	// Check components/schemas
	if openapi.Components == nil {
		t.Fatal("expected Components to be set")
	}

	if openapi.Components.Schemas == nil {
		t.Fatal("expected Schemas to be set")
	}

	// Check User schema
	userSchema, ok := openapi.Components.Schemas["User"]
	if !ok {
		t.Fatal("expected User schema to exist")
	}

	if userSchema.Type != "object" {
		t.Errorf("expected User schema type 'object', got %q", userSchema.Type)
	}

	if len(userSchema.Properties) != 3 {
		t.Errorf("expected 3 properties in User schema, got %d", len(userSchema.Properties))
	}

	// Check ID property
	idProp, ok := userSchema.Properties["id"]
	if !ok {
		t.Fatal("expected 'id' property in User schema")
	}

	if idProp.Type != "integer" {
		t.Errorf("expected 'id' type 'integer', got %q", idProp.Type)
	}

	// Check Name property
	nameProp, ok := userSchema.Properties["name"]
	if !ok {
		t.Fatal("expected 'name' property in User schema")
	}

	if nameProp.Type != "string" {
		t.Errorf("expected 'name' type 'string', got %q", nameProp.Type)
	}

	// Check Email property
	emailProp, ok := userSchema.Properties["email"]
	if !ok {
		t.Fatal("expected 'email' property in User schema")
	}

	if emailProp.Type != "string" {
		t.Errorf("expected 'email' type 'string', got %q", emailProp.Type)
	}
}

