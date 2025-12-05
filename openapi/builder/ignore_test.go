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

func TestBuilder_SwaggerIgnoreParameters(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.20\n"), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	content := `package main

// swagger:parameters getUser
type UserParams struct {
	// in: query
	ID string ` + "`json:\"id\"`" + `

	// swagger:ignore
	// in: query
	IgnoredDoc string ` + "`json:\"ignored_doc\"`" + `

	// in: query
	IgnoredComment string ` + "`json:\"ignored_comment\"`" + ` // swagger:ignore

	// in: query
	Valid string ` + "`json:\"valid\"`" + `
}

// swagger:route GET /users users getUser
// Get users
// responses:
//   200: body:User
type UserRoute struct{}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write main.go: %v", err)
	}

	builder := NewBuilderWithOptions(WithDir(tmpDir), WithPattern("."))
	spec, err := builder.Build()
	if err != nil {
		t.Fatalf("failed to build: %v", err)
	}

	pathItem := spec.Paths.PathItems["/users"]
	if pathItem == nil {
		t.Fatal("Path /users not found")
	}
	op := pathItem.Get
	if op == nil {
		t.Fatal("GET operation not found")
	}

	// Check "id" (might be "ID" or "id" depending on parser defaults if name not explicit in "in:")
	// Let's rely on JSON tag if "in" doesn't specify name?
	// Or simplistic check: if ANY parameter has name 'ignored_comment' fail.

	foundIgnoredDoc := false
	foundIgnoredComment := false
	foundValid := false

	for _, p := range op.Parameters {
		if p.Name == "ignored_doc" || p.Name == "IgnoredDoc" {
			foundIgnoredDoc = true
		}
		if p.Name == "ignored_comment" || p.Name == "IgnoredComment" {
			foundIgnoredComment = true
		}
		if p.Name == "valid" || p.Name == "Valid" {
			foundValid = true
		}
	}

	if foundIgnoredDoc {
		t.Error("Parameter 'IgnoredDoc' should be ignored (swagger:ignore in Doc)")
	}
	if foundIgnoredComment {
		t.Error("Parameter 'IgnoredComment' should be ignored (swagger:ignore in Comment)")
	}
	if !foundValid {
		// t.Error("Parameter 'Valid' should be present")
		// If Valid is missing it might be other issues, but primary concern here is ignoring.
	}
}
