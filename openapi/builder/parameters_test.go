package builder

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuilder_Parameters(t *testing.T) {
	tmpDir := t.TempDir()

	// Create go.mod
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.20\n"), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	// Create a test file with swagger:parameters detached
	testFile := filepath.Join(tmpDir, "params.go")
	content := `package main

// swagger:route GET /test test testOp
// Responses:
//   200: description:OK
type TestRoute struct {}

// swagger:parameters testOp
type TestParams struct {
	// in: query
	Limit int ` + "`json:\"limit\"`" + `
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	builder := NewBuilderWithOptions(
		WithDir(tmpDir),
		WithPattern("."),
	)
	openapi, err := builder.Build()
	if err != nil {
		t.Fatalf("failed to build spec: %v", err)
	}

	// Verify parameter attached to operation
	pathItem := openapi.Paths.PathItems["/test"]
	if pathItem == nil {
		t.Fatal("expected /test path")
	}
	op := pathItem.Get
	if op == nil {
		t.Fatal("expected GET operation")
	}

	if len(op.Parameters) != 1 {
		t.Fatalf("expected 1 parameter, got %d", len(op.Parameters))
	}

	param := op.Parameters[0]
	if param.Name != "limit" {
		t.Errorf("expected param name 'limit', got %q", param.Name)
	}
	if param.In != "query" {
		t.Errorf("expected param in 'query', got %q", param.In)
	}
	if param.Schema.Type != "integer" {
		t.Errorf("expected param type 'integer', got %q", param.Schema.Type)
	}
}
