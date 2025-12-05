package builder

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuilder_SwaggerIgnoreLocations(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.20\n"), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	content := `package main

// swagger:model
type User struct {
	// This field should be included
	Name string ` + "`json:\"name\"`" + `

	// swagger:ignore
	// This field has ignore in Doc (above)
	SecretDoc string ` + "`json:\"secret_doc\"`" + `

	SecretComment string ` + "`json:\"secret_comment\"`" + ` // swagger:ignore

	// Normal field
	Email string ` + "`json:\"email\"`" + `
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write main.go: %v", err)
	}

	builder := NewBuilderWithOptions(WithDir(tmpDir), WithPattern("."))
	spec, err := builder.Build()
	if err != nil {
		t.Fatalf("failed to build: %v", err)
	}

	user := spec.Components.Schemas["User"]
	if user == nil {
		t.Fatal("User schema not found")
	}

	t.Run("includes normal fields", func(t *testing.T) {
		if _, ok := user.Properties["name"]; !ok {
			t.Error("should include 'name' field")
		}
		if _, ok := user.Properties["email"]; !ok {
			t.Error("should include 'email' field")
		}
	})

	t.Run("ignores field with swagger:ignore in Doc", func(t *testing.T) {
		if _, ok := user.Properties["secret_doc"]; ok {
			t.Error("should NOT include 'secret_doc' - has swagger:ignore in Doc")
		}
	})

	t.Run("ignores field with swagger:ignore in Comment (inline)", func(t *testing.T) {
		if _, ok := user.Properties["secret_comment"]; ok {
			t.Error("should NOT include 'secret_comment' - has swagger:ignore as inline Comment")
		}
	})
}

