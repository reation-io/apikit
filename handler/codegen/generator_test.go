package codegen

import (
	"strings"
	"testing"

	"github.com/reation-io/apikit/core/definition"
)

func TestNew(t *testing.T) {
	gen, err := New()

	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if gen == nil {
		t.Fatal("expected generator to be created")
	}

	if gen.tmpl == nil {
		t.Error("expected template to be initialized")
	}
}

func TestGenerate_NoHandlers(t *testing.T) {
	gen, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	result := &definition.Definition{
		Operations: []*definition.Operation{},
		Types:      make(map[string]*definition.Type),
		Package:    "test",
	}

	_, err = gen.Generate(result)
	if err == nil {
		t.Error("expected error for no handlers")
	}
}

func TestGenerate_SimpleHandler(t *testing.T) {
	gen, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Create a simple handler with a request struct
	reqType := &definition.Type{
		Name: "CreateUserRequest",
		Kind: "struct",
		Fields: []*definition.Field{
			{
				Name: "Name",
				Type: &definition.Type{GoType: "string", Kind: "primitive"},
				Tags: `json:"name" validate:"required"`,
			},
			{
				Name: "Email",
				Type: &definition.Type{GoType: "string", Kind: "primitive"},
				Tags: `json:"email" validate:"required,email"`,
			},
		},
	}

	handler := &definition.Operation{
		ID:          "CreateUser",
		RequestType: "CreateUserRequest",
		ReturnType:  "CreateUserResponse",
	}

	result := &definition.Definition{
		Operations: []*definition.Operation{handler},
		Types: map[string]*definition.Type{
			"CreateUserRequest": reqType,
		},
		Package: "test",
	}

	code, err := gen.Generate(result)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	codeStr := string(code)

	// Check that generated code contains expected elements
	expectedElements := []string{
		"package test",
		"CreateUser",
		"CreateUserRequest",
		"CreateUserResponse",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(codeStr, expected) {
			t.Errorf("expected generated code to contain %q, got:\n%s", expected, codeStr)
		}
	}
}

func TestGenerate_WithPathParameter(t *testing.T) {
	gen, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	reqType := &definition.Type{
		Name: "GetUserRequest",
		Kind: "struct",
		Fields: []*definition.Field{
			{
				Name: "UserID",
				Type: &definition.Type{GoType: "string", Kind: "primitive"},
				Tags: `path:"userId"`,
			},
		},
	}

	handler := &definition.Operation{
		ID:          "GetUser",
		RequestType: "GetUserRequest",
		ReturnType:  "GetUserResponse",
	}

	result := &definition.Definition{
		Operations: []*definition.Operation{handler},
		Types: map[string]*definition.Type{
			"GetUserRequest": reqType,
		},
		Package: "test",
	}

	code, err := gen.Generate(result)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	codeStr := string(code)

	// Should contain path extraction code
	if !strings.Contains(codeStr, "PathValue") {
		t.Error("expected generated code to contain PathValue for path parameter")
	}
}

func TestGenerate_UsesHandleResponse(t *testing.T) {
	gen, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	reqType := &definition.Type{
		Name: "CreateUserRequest",
		Kind: "struct",
		Fields: []*definition.Field{
			{
				Name: "Name",
				Type: &definition.Type{GoType: "string", Kind: "primitive"},
				Tags: `json:"name"`,
			},
		},
	}

	handler := &definition.Operation{
		ID:          "CreateUser",
		RequestType: "CreateUserRequest",
		ReturnType:  "CreateUserResponse",
	}

	result := &definition.Definition{
		Operations: []*definition.Operation{handler},
		Types: map[string]*definition.Type{
			"CreateUserRequest": reqType,
		},
		Package: "test",
	}

	code, err := gen.Generate(result)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	codeStr := string(code)

	// Should use HandleResponse instead of separate HandleError and WriteJSON
	if !strings.Contains(codeStr, "apikit.HandleResponse(w, response, err)") {
		t.Error("expected generated code to use apikit.HandleResponse")
	}

	// Should NOT contain the old pattern
	if strings.Contains(codeStr, "apikit.WriteJSON(w, response)") {
		t.Error("expected generated code to NOT use apikit.WriteJSON directly")
	}

	// Should NOT contain the old error handling pattern
	oldPattern := "if err != nil {\n\t\tapikit.HandleError(w, err)\n\t\treturn\n\t}\n\n\t// Write response"
	if strings.Contains(codeStr, oldPattern) {
		t.Error("expected generated code to NOT use old error handling pattern")
	}
}
