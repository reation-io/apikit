package builder

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuilder_EmbeddedEdgeCases(t *testing.T) {
	// Test 1: Embedding a struct WITHOUT swagger:model
	t.Run("embeds non-model struct", func(t *testing.T) {
		tmpDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.20\n"), 0644); err != nil {
			t.Fatalf("failed to write go.mod: %v", err)
		}

		content := `package main

// BaseFields is NOT a swagger:model
type BaseFields struct {
ID        string ` + "`json:\"id\"`" + `
CreatedAt string ` + "`json:\"created_at\"`" + `
}

// swagger:model
type User struct {
BaseFields
Name string ` + "`json:\"name\"`" + `
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

		t.Logf("User properties: %v", user.Properties)

		if _, ok := user.Properties["name"]; !ok {
			t.Error("User should have 'name' property")
		}
		if _, ok := user.Properties["id"]; !ok {
			t.Error("User should have inherited 'id' from BaseFields")
		}
		if _, ok := user.Properties["created_at"]; !ok {
			t.Error("User should have inherited 'created_at' from BaseFields")
		}
	})

	// Test 2: Circular embedding (A embeds B, B embeds A)
	t.Run("handles circular embedding", func(t *testing.T) {
		tmpDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.20\n"), 0644); err != nil {
			t.Fatalf("failed to write go.mod: %v", err)
		}

		content := `package main

// swagger:model
type ModelA struct {
*ModelB
FieldA string ` + "`json:\"field_a\"`" + `
}

// swagger:model
type ModelB struct {
*ModelA
FieldB string ` + "`json:\"field_b\"`" + `
}
`
		if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write main.go: %v", err)
		}

		builder := NewBuilderWithOptions(WithDir(tmpDir), WithPattern("."))
		spec, err := builder.Build()

		// Should NOT panic or hang
		if err != nil {
			t.Logf("Build returned error (acceptable): %v", err)
		}

		if spec != nil && spec.Components != nil {
			t.Logf("ModelA properties: %v", spec.Components.Schemas["ModelA"])
			t.Logf("ModelB properties: %v", spec.Components.Schemas["ModelB"])
		}
	})

	// Test 3: Deep nesting (A embeds B, B embeds C)
	t.Run("resolves deep nesting", func(t *testing.T) {
		tmpDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.20\n"), 0644); err != nil {
			t.Fatalf("failed to write go.mod: %v", err)
		}

		// Files in reverse order to test dependency resolution
		fileC := `package main
// swagger:model
type ModelC struct {
FieldC string ` + "`json:\"field_c\"`" + `
}
`
		fileB := `package main
// swagger:model
type ModelB struct {
ModelC
FieldB string ` + "`json:\"field_b\"`" + `
}
`
		fileA := `package main
// swagger:model
type ModelA struct {
ModelB
FieldA string ` + "`json:\"field_a\"`" + `
}
`
		// Write in alphabetical order (A first, C last)
		if err := os.WriteFile(filepath.Join(tmpDir, "a.go"), []byte(fileA), 0644); err != nil {
			t.Fatalf("failed to write a.go: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "b.go"), []byte(fileB), 0644); err != nil {
			t.Fatalf("failed to write b.go: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "c.go"), []byte(fileC), 0644); err != nil {
			t.Fatalf("failed to write c.go: %v", err)
		}

		builder := NewBuilderWithOptions(WithDir(tmpDir), WithPattern("."))
		spec, err := builder.Build()
		if err != nil {
			t.Fatalf("failed to build: %v", err)
		}

		modelA := spec.Components.Schemas["ModelA"]
		if modelA == nil {
			t.Fatal("ModelA not found")
		}

		t.Logf("ModelA properties: %v", modelA.Properties)

		// ModelA should have all three fields
		if _, ok := modelA.Properties["field_a"]; !ok {
			t.Error("ModelA should have field_a")
		}
		if _, ok := modelA.Properties["field_b"]; !ok {
			t.Error("ModelA should have inherited field_b from ModelB")
		}
		if _, ok := modelA.Properties["field_c"]; !ok {
			t.Error("ModelA should have inherited field_c from ModelC (via ModelB)")
		}
	})

	// Test 4: Field override (child overrides parent field)
	t.Run("child field overrides embedded field", func(t *testing.T) {
		tmpDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.20\n"), 0644); err != nil {
			t.Fatalf("failed to write go.mod: %v", err)
		}

		content := `package main

// swagger:model
type Base struct {
// Description: Base ID
ID string ` + "`json:\"id\"`" + `
}

// swagger:model
type Child struct {
Base
// Description: Child's own ID (overrides Base)
ID int ` + "`json:\"id\"`" + `
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

		child := spec.Components.Schemas["Child"]
		if child == nil {
			t.Fatal("Child schema not found")
		}

		idProp := child.Properties["id"]
		if idProp == nil {
			t.Fatal("Child should have 'id' property")
		}

		// Should be integer (from Child), not string (from Base)
		t.Logf("Child.id type: %s", idProp.Type)
		if idProp.Type != "integer" {
			t.Errorf("Child.id should be integer (child override), got %s", idProp.Type)
		}
	})
}
