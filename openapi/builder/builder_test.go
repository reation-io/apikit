package builder

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestBuilder_Meta(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create a test file with swagger:meta
	testFile := filepath.Join(tmpDir, "api.go")
	content := `package main

// swagger:meta
// Title: My Test API
// Version: 1.0.0
// Description: This is a test API
//   with multiple lines
type API struct{}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Build the spec
	builder := NewBuilder(filepath.Join(tmpDir, "*.go"))
	openapi, err := builder.Build()
	if err != nil {
		t.Fatalf("failed to build spec: %v", err)
	}

	// Verify Info
	if openapi.Info.Title != "My Test API" {
		t.Errorf("expected title 'My Test API', got %q", openapi.Info.Title)
	}
	if openapi.Info.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", openapi.Info.Version)
	}
	if openapi.Info.Description != "This is a test API\nwith multiple lines" {
		t.Errorf("expected description, got %q", openapi.Info.Description)
	}
}

func TestBuilder_Route(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create a test file with swagger:route
	testFile := filepath.Join(tmpDir, "handlers.go")
	content := `package main

// swagger:route POST /users user createUser
// Summary: Create a new user
// Tags: users, admin
type CreateUserRequest struct{}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Build the spec
	builder := NewBuilder(filepath.Join(tmpDir, "*.go"))
	openapi, err := builder.Build()
	if err != nil {
		t.Fatalf("failed to build spec: %v", err)
	}

	// Verify operation
	pathItem := openapi.Paths.PathItems["/users"]
	if pathItem == nil {
		t.Fatal("expected /users path to exist")
	}
	if pathItem.Post == nil {
		t.Fatal("expected POST operation to exist")
	}

	operation := pathItem.Post
	if operation.OperationID != "createUser" {
		t.Errorf("expected operationId 'createUser', got %q", operation.OperationID)
	}
	if operation.Summary != "Create a new user" {
		t.Errorf("expected summary 'Create a new user', got %q", operation.Summary)
	}
	if len(operation.Tags) != 2 || operation.Tags[0] != "users" || operation.Tags[1] != "admin" {
		t.Errorf("expected tags [users, admin], got %v", operation.Tags)
	}
}

func TestBuilder_Model(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create a test file with swagger:model
	testFile := filepath.Join(tmpDir, "models.go")
	content := `package main

// swagger:model
type User struct {
	// Example: user@example.com
	// Format: email
	Email string ` + "`json:\"email\"`" + `

	// MinLength: 3
	// MaxLength: 50
	Name string ` + "`json:\"name\"`" + `

	// Minimum: 0
	// Maximum: 150
	Age int ` + "`json:\"age\"`" + `
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Build the spec
	builder := NewBuilder(filepath.Join(tmpDir, "*.go"))
	openapi, err := builder.Build()
	if err != nil {
		t.Fatalf("failed to build spec: %v", err)
	}

	// Verify schema
	if openapi.Components == nil {
		t.Fatal("expected components to exist")
	}
	schema := openapi.Components.Schemas["User"]
	if schema == nil {
		t.Fatal("expected User schema to exist")
	}
	if schema.Type != "object" {
		t.Errorf("expected type 'object', got %q", schema.Type)
	}

	// Verify email field
	emailSchema := schema.Properties["email"]
	if emailSchema == nil {
		t.Fatal("expected email property to exist")
	}
	if emailSchema.Example != "user@example.com" {
		t.Errorf("expected example 'user@example.com', got %v", emailSchema.Example)
	}
	if emailSchema.Format != "email" {
		t.Errorf("expected format 'email', got %q", emailSchema.Format)
	}
}

func TestBuilder_JSON(t *testing.T) {
	// Create a simple spec
	builder := NewBuilder()
	builder.spec.Info.Title = "Test API"
	builder.spec.Info.Version = "1.0.0"

	// Marshal to JSON
	data, err := json.MarshalIndent(builder.spec, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}

	// Verify JSON contains expected fields
	jsonStr := string(data)
	if !contains(jsonStr, "Test API") {
		t.Error("expected JSON to contain 'Test API'")
	}
	if !contains(jsonStr, "1.0.0") {
		t.Error("expected JSON to contain '1.0.0'")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr))))
}
