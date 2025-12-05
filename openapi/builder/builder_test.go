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

func TestNewBuilderWithOptions(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		builder := NewBuilderWithOptions()

		if builder.config == nil {
			t.Fatal("expected config to be set")
		}
		if builder.spec == nil {
			t.Fatal("expected spec to be initialized")
		}
		if builder.spec.OpenAPI != "3.0.3" {
			t.Errorf("expected OpenAPI version '3.0.3', got %q", builder.spec.OpenAPI)
		}
	})

	t.Run("with pattern option enables scanner", func(t *testing.T) {
		builder := NewBuilderWithOptions(
			WithPattern("./..."),
		)

		if !builder.config.UseScanner {
			t.Error("expected UseScanner to be true when WithPattern is used")
		}
		if builder.config.ScannerConfig.Pattern != "./..." {
			t.Errorf("expected pattern './...', got %q", builder.config.ScannerConfig.Pattern)
		}
	})

	t.Run("with dir option", func(t *testing.T) {
		builder := NewBuilderWithOptions(
			WithDir("/project"),
		)

		if !builder.config.UseScanner {
			t.Error("expected UseScanner to be true when WithDir is used")
		}
		if builder.config.ScannerConfig.Dir != "/project" {
			t.Errorf("expected dir '/project', got %q", builder.config.ScannerConfig.Dir)
		}
	})

	t.Run("with ignore paths", func(t *testing.T) {
		builder := NewBuilderWithOptions(
			WithPattern("./..."),
			WithIgnorePaths("vendor/**", "test/**"),
		)

		if len(builder.config.ScannerConfig.IgnorePaths) != 2 {
			t.Errorf("expected 2 ignore paths, got %d", len(builder.config.ScannerConfig.IgnorePaths))
		}
	})

	t.Run("with validation", func(t *testing.T) {
		builder := NewBuilderWithOptions(
			WithValidation(true),
		)

		if !builder.config.Validation {
			t.Error("expected Validation to be true")
		}
	})

	t.Run("with legacy patterns", func(t *testing.T) {
		builder := NewBuilderWithOptions(
			WithLegacyPatterns("*.go", "handlers/*.go"),
		)

		if builder.config.UseScanner {
			t.Error("expected UseScanner to be false when WithLegacyPatterns is used")
		}
		if len(builder.config.Patterns) != 2 {
			t.Errorf("expected 2 patterns, got %d", len(builder.config.Patterns))
		}
	})
}

func TestBuilder_WithScanner(t *testing.T) {
	t.Run("scan current package", func(t *testing.T) {
		// This test scans the builder package itself
		builder := NewBuilderWithOptions(
			WithPattern("."),
			WithDir("."),
		)

		spec, err := builder.Build()
		if err != nil {
			t.Fatalf("failed to build with scanner: %v", err)
		}

		// The spec should be valid even if no swagger annotations are found
		if spec == nil {
			t.Fatal("expected spec to be returned")
		}
		if spec.OpenAPI != "3.0.3" {
			t.Errorf("expected OpenAPI version '3.0.3', got %q", spec.OpenAPI)
		}
	})
}

func TestBuilder_Validation(t *testing.T) {
	t.Run("validation enabled catches errors", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a test file with invalid route (operation without responses)
		testFile := filepath.Join(tmpDir, "api.go")
		content := `package main

// swagger:route GET /users users listUsers
// Summary: List all users
type ListUsersRequest struct{}
`
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		builder := NewBuilderWithOptions(
			WithLegacyPatterns(filepath.Join(tmpDir, "*.go")),
			WithValidation(true),
		)

		_, err := builder.Build()
		if err == nil {
			t.Error("expected validation error for operation without responses")
		}
	})

	t.Run("validation disabled allows invalid spec", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a test file with invalid spec (missing title)
		testFile := filepath.Join(tmpDir, "api.go")
		content := `package main

// swagger:meta
// Version: 1.0.0
type API struct{}
`
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		builder := NewBuilderWithOptions(
			WithLegacyPatterns(filepath.Join(tmpDir, "*.go")),
			WithValidation(false),
		)

		spec, err := builder.Build()
		if err != nil {
			t.Fatalf("build failed: %v", err)
		}
		if spec == nil {
			t.Error("expected spec to be returned")
		}
	})

	t.Run("Validate method returns result", func(t *testing.T) {
		builder := NewBuilder()
		// Default spec should be valid
		result := builder.Validate()

		if result == nil {
			t.Error("expected validation result")
		}
	})
}

func TestBuilder_SwaggerIgnore(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("ignores fields in models", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "models.go")
		content := `package test

// User represents a user in the system
// swagger:model
type User struct {
	// ID of the user
	ID string ` + "`json:\"id\"`" + `

	// Name of the user
	Name string ` + "`json:\"name\"`" + `

	// swagger:ignore
	// Internal token - should not appear in spec
	InternalToken string ` + "`json:\"internal_token\"`" + `
}
`
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		builder := NewBuilder(filepath.Join(tmpDir, "*.go"))
		openapi, err := builder.Build()
		if err != nil {
			t.Fatalf("failed to build spec: %v", err)
		}

		schema := openapi.Components.Schemas["User"]
		if schema == nil {
			t.Fatal("expected User schema")
		}

		// Check that id and name are present
		if _, ok := schema.Properties["id"]; !ok {
			t.Error("expected 'id' property")
		}
		if _, ok := schema.Properties["name"]; !ok {
			t.Error("expected 'name' property")
		}

		// Check that internal_token is NOT present
		if _, ok := schema.Properties["internal_token"]; ok {
			t.Error("'internal_token' should be ignored via swagger:ignore")
		}
	})
}
