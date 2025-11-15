package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/reation-io/apikit/openapi/spec"
)

func TestOpenAPICommand(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package test

// swagger:meta
type Meta struct{}

// Title: Test API
// Version: 1.0.0

// User represents a user
// swagger:model
type User struct {
	ID   int    ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
}

// CreateUserRequest creates a user
// swagger:route POST /users user createUser
// Summary: Create user
type CreateUserRequest struct {
	Name string ` + "`json:\"name\"`" + `
}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Set output file
	outputFile := filepath.Join(tmpDir, "openapi.json")
	openapiOutput = outputFile
	openapiFormat = "json"
	openapiTitle = ""
	openapiVer = ""

	// Change to temp directory so relative paths work
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	// Run command with relative path
	if err := runOpenAPI(nil, []string{"test.go"}); err != nil {
		t.Fatalf("runOpenAPI failed: %v", err)
	}

	// Verify output file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatal("output file was not created")
	}

	// Read and parse output
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	var openapi spec.OpenAPI
	if err := json.Unmarshal(data, &openapi); err != nil {
		t.Fatalf("failed to parse OpenAPI JSON: %v", err)
	}

	// Verify OpenAPI version
	if openapi.OpenAPI != "3.0.3" {
		t.Errorf("expected OpenAPI version '3.0.3', got %q", openapi.OpenAPI)
	}

	// Verify Info
	if openapi.Info == nil {
		t.Fatal("expected Info to be set")
	}

	// Verify paths
	if openapi.Paths == nil || len(openapi.Paths.PathItems) == 0 {
		t.Logf("Generated OpenAPI: %s", string(data))
		t.Fatal("expected at least one path")
	}

	// Verify /users path
	usersPath, ok := openapi.Paths.PathItems["/users"]
	if !ok {
		t.Fatal("expected /users path")
	}

	if usersPath.Post == nil {
		t.Fatal("expected POST operation on /users")
	}

	if usersPath.Post.OperationID != "createUser" {
		t.Errorf("expected operation ID 'createUser', got %q", usersPath.Post.OperationID)
	}

	// Verify schemas
	if openapi.Components == nil || openapi.Components.Schemas == nil {
		t.Fatal("expected schemas to be set")
	}

	userSchema, ok := openapi.Components.Schemas["User"]
	if !ok {
		t.Fatal("expected User schema")
	}

	if userSchema.Type != "object" {
		t.Errorf("expected User schema type 'object', got %q", userSchema.Type)
	}

	if len(userSchema.Properties) != 2 {
		t.Errorf("expected 2 properties in User schema, got %d", len(userSchema.Properties))
	}
}

func TestOpenAPICommandWithOverrides(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package test

// swagger:meta
type Meta struct{}

// swagger:route GET /test test getTest
type GetTestRequest struct{}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Set output file with overrides
	outputFile := filepath.Join(tmpDir, "openapi.json")
	openapiOutput = outputFile
	openapiFormat = "json"
	openapiTitle = "Custom Title"
	openapiVer = "2.0.0"

	// Change to temp directory so relative paths work
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	// Run command with relative path
	if err := runOpenAPI(nil, []string{"test.go"}); err != nil {
		t.Fatalf("runOpenAPI failed: %v", err)
	}

	// Read and parse output
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	var openapi spec.OpenAPI
	if err := json.Unmarshal(data, &openapi); err != nil {
		t.Fatalf("failed to parse OpenAPI JSON: %v", err)
	}

	// Verify overrides
	if openapi.Info.Title != "Custom Title" {
		t.Errorf("expected title 'Custom Title', got %q", openapi.Info.Title)
	}

	if openapi.Info.Version != "2.0.0" {
		t.Errorf("expected version '2.0.0', got %q", openapi.Info.Version)
	}
}
