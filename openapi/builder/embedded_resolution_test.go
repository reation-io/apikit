package builder

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuilder_EmbeddedResolution(t *testing.T) {
	// Tests that embedded fields are resolved correctly regardless of file order
	t.Run("resolves embedded dependency when dependency is parsed last", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create go.mod
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.20\n"), 0644); err != nil {
			t.Fatalf("failed to write go.mod: %v", err)
		}

		// File A: Embeds B
		fileA := filepath.Join(tmpDir, "a.go")
		contentA := `package main
// swagger:model
type ModelA struct {
	ModelB
	FieldA string ` + "`json:\"field_a\"`" + `
}
`
		if err := os.WriteFile(fileA, []byte(contentA), 0644); err != nil {
			t.Fatalf("failed to write fileA: %v", err)
		}

		// File B: Defines B
		fileB := filepath.Join(tmpDir, "b.go")
		contentB := `package main
// swagger:model
type ModelB struct {
	FieldB string ` + "`json:\"field_b\"`" + `
}
`
		if err := os.WriteFile(fileB, []byte(contentB), 0644); err != nil {
			t.Fatalf("failed to write fileB: %v", err)
		}

		// Build
		builder := NewBuilderWithOptions(
			WithDir(tmpDir),
			WithPattern("."),
		)
		spec, err := builder.Build()
		if err != nil {
			t.Fatalf("failed to build: %v", err)
		}

		schemaA := spec.Components.Schemas["ModelA"]
		if schemaA == nil {
			t.Fatal("ModelA not found")
		}

		// Check if FieldB (from embedded ModelB) is present in ModelA
		if _, ok := schemaA.Properties["field_b"]; !ok {
			t.Error("ModelA should have inherited field_b from ModelB")
		}
		if _, ok := schemaA.Properties["field_a"]; !ok {
			t.Error("ModelA should have field_a")
		}
	})
}
