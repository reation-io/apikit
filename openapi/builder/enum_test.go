package builder

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuilder_Enum(t *testing.T) {
	t.Run("basic enum parsing", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a test file with swagger:enum
		testFile := filepath.Join(tmpDir, "status.go")
		content := `package main

// swagger:enum UserStatus
// Status of a user account
// example: active
type UserStatus string

const (
	StatusActive   UserStatus = "active"
	StatusInactive UserStatus = "inactive"
	StatusPending  UserStatus = "pending"
)
`
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		// Build the spec
		builder := NewBuilder(filepath.Join(tmpDir, "*.go"))
		_, err := builder.Build()
		if err != nil {
			t.Fatalf("failed to build spec: %v", err)
		}

		// Verify enum was registered
		enumInfo := builder.enumRegistry.GetByTypeName("UserStatus")
		if enumInfo == nil {
			t.Fatal("expected UserStatus enum to be registered")
		}

		if enumInfo.Name != "UserStatus" {
			t.Errorf("expected name 'UserStatus', got %q", enumInfo.Name)
		}

		if enumInfo.BaseType != "string" {
			t.Errorf("expected base type 'string', got %q", enumInfo.BaseType)
		}

		if len(enumInfo.Values) != 3 {
			t.Errorf("expected 3 values, got %d", len(enumInfo.Values))
		}

		// Check specific values
		if enumInfo.Values["StatusActive"] != "active" {
			t.Errorf("expected StatusActive='active', got %v", enumInfo.Values["StatusActive"])
		}
	})

	t.Run("enum with custom name", func(t *testing.T) {
		tmpDir := t.TempDir()

		testFile := filepath.Join(tmpDir, "role.go")
		content := `package main

// swagger:enum Role
// User role enumeration
type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)
`
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		builder := NewBuilder(filepath.Join(tmpDir, "*.go"))
		_, err := builder.Build()
		if err != nil {
			t.Fatalf("failed to build spec: %v", err)
		}

		// Should be registered under custom name "Role"
		enumInfo := builder.enumRegistry.Get("Role")
		if enumInfo == nil {
			t.Fatal("expected Role enum to be registered")
		}

		// Should also be accessible by type name
		enumInfo2 := builder.enumRegistry.GetByTypeName("UserRole")
		if enumInfo2 == nil {
			t.Fatal("expected enum to be accessible by type name")
		}
	})

	t.Run("enum field in model", func(t *testing.T) {
		tmpDir := t.TempDir()

		testFile := filepath.Join(tmpDir, "models.go")
		content := `package main

// swagger:enum
type UserStatus string

const (
	StatusActive UserStatus = "active"
	StatusInactive UserStatus = "inactive"
)

// swagger:model
type User struct {
	Name   string ` + "`json:\"name\"`" + `
	Status UserStatus ` + "`json:\"status\"`" + `
}
`
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		builder := NewBuilder(filepath.Join(tmpDir, "*.go"))
		spec, err := builder.Build()
		if err != nil {
			t.Fatalf("failed to build spec: %v", err)
		}

		// Check that User model has status field with enum
		userSchema := spec.Components.Schemas["User"]
		if userSchema == nil {
			t.Fatal("expected User schema to exist")
		}

		statusField := userSchema.Properties["status"]
		if statusField == nil {
			t.Fatal("expected status field to exist")
		}

		// The field should have inline enum values
		if statusField.Type != "string" {
			t.Errorf("expected type 'string', got %q", statusField.Type)
		}

		if len(statusField.Enum) != 2 {
			t.Errorf("expected 2 enum values, got %d: %v", len(statusField.Enum), statusField.Enum)
		}
	})

	t.Run("integer enum", func(t *testing.T) {
		tmpDir := t.TempDir()

		testFile := filepath.Join(tmpDir, "priority.go")
		content := `package main

// swagger:enum
type Priority int

const (
	PriorityLow    Priority = 1
	PriorityMedium Priority = 2
	PriorityHigh   Priority = 3
)
`
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		builder := NewBuilder(filepath.Join(tmpDir, "*.go"))
		_, err := builder.Build()
		if err != nil {
			t.Fatalf("failed to build spec: %v", err)
		}

		enumInfo := builder.enumRegistry.GetByTypeName("Priority")
		if enumInfo == nil {
			t.Fatal("expected Priority enum to be registered")
		}

		if enumInfo.BaseType != "integer" {
			t.Errorf("expected base type 'integer', got %q", enumInfo.BaseType)
		}
	})
}
