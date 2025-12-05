package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseFile_Handler(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "handler.go")

	content := `package test

import "net/http"

// HandleRequest handles a request
// apikit:handler GET /api/v1/test
func HandleRequest(w http.ResponseWriter, r *http.Request) {
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	p := New()
	def, err := p.ParseFile(testFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(def.Operations) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(def.Operations))
	}

	op := def.Operations[0]
	if op.Method != "GET" {
		t.Errorf("expected method GET, got %q", op.Method)
	}
	if op.Path != "/api/v1/test" {
		t.Errorf("expected path /api/v1/test, got %q", op.Path)
	}
	if op.ID != "HandleRequest" {
		t.Errorf("expected ID HandleRequest, got %q", op.ID)
	}
}

func TestParseFile_StructRoute(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "route.go")

	content := `package test

// CreateUserRequest creates a user
// swagger:route POST /users user createUser
// Summary: Create User
type CreateUserRequest struct {
	Name string
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	p := New()
	def, err := p.ParseFile(testFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(def.Operations) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(def.Operations))
	}

	op := def.Operations[0]
	if op.Method != "POST" {
		t.Errorf("expected method POST, got %q", op.Method)
	}
	if op.Path != "/users" {
		t.Errorf("expected path /users, got %q", op.Path)
	}
	if op.ID != "createUser" {
		t.Errorf("expected ID createUser, got %q", op.ID)
	}
	if len(op.Tags) != 1 || op.Tags[0] != "user" {
		t.Errorf("expected tag user, got %v", op.Tags)
	}
	if op.RequestType != "CreateUserRequest" {
		t.Errorf("expected request type CreateUserRequest, got %q", op.RequestType)
	}
	if !strings.Contains(op.Description, "Summary: Create User") {
		t.Errorf("expected description to contain summary, got %q", op.Description)
	}
}

func TestParseFile_Model(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "model.go")

	content := `package test

// User represents a user
// swagger:model
type User struct {
	ID   int    ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	p := New()
	def, err := p.ParseFile(testFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(def.Types) != 1 {
		t.Fatalf("expected 1 type, got %d", len(def.Types))
	}

	typ, ok := def.Types["User"]
	if !ok {
		t.Fatal("expected User type")
	}

	if typ.Name != "User" {
		t.Errorf("expected name User, got %q", typ.Name)
	}
	if len(typ.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(typ.Fields))
	}
}
