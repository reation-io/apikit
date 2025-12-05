package builder

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuilder_FieldValidations(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.20\n"), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	content := `package main

// swagger:model
type Product struct {
	// Description: Product name
	// Minimum: 0
	// Maximum: 100
	Price float64 ` + "`json:\"price\"`" + `

	// MinLength: 3
	// MaxLength: 50
	// Pattern: ^[a-zA-Z]+$
	Name string ` + "`json:\"name\"`" + `

	// Format: email
	Email string ` + "`json:\"email\"`" + `

	// Example: SKU-12345
	SKU string ` + "`json:\"sku\"`" + `

	// Default: 1
	Quantity int ` + "`json:\"quantity\"`" + `

	// ReadOnly: true
	ID string ` + "`json:\"id\"`" + `

	// WriteOnly: true  
	Password string ` + "`json:\"password\"`" + `

	// MinItems: 1
	// MaxItems: 10
	Tags []string ` + "`json:\"tags\"`" + `

	// UniqueItems: true
	Categories []string ` + "`json:\"categories\"`" + `

	// MultipleOf: 0.01
	Discount float64 ` + "`json:\"discount\"`" + `
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

	product := spec.Components.Schemas["Product"]
	if product == nil {
		t.Fatal("Product schema not found")
	}

	tests := []struct {
		field    string
		check    string
		validate func() bool
	}{
		{"price", "Minimum", func() bool { return product.Properties["price"].Minimum != nil && *product.Properties["price"].Minimum == 0 }},
		{"price", "Maximum", func() bool { return product.Properties["price"].Maximum != nil && *product.Properties["price"].Maximum == 100 }},
		{"name", "MinLength", func() bool { return product.Properties["name"].MinLength != nil && *product.Properties["name"].MinLength == 3 }},
		{"name", "MaxLength", func() bool { return product.Properties["name"].MaxLength != nil && *product.Properties["name"].MaxLength == 50 }},
		{"name", "Pattern", func() bool { return product.Properties["name"].Pattern == "^[a-zA-Z]+$" }},
		{"email", "Format", func() bool { return product.Properties["email"].Format == "email" }},
		{"sku", "Example", func() bool { return product.Properties["sku"].Example == "SKU-12345" }},
		{"quantity", "Default", func() bool { return product.Properties["quantity"].Default != nil }},
		{"id", "ReadOnly", func() bool { return product.Properties["id"].ReadOnly }},
		{"password", "WriteOnly", func() bool { return product.Properties["password"].WriteOnly }},
		{"tags", "MinItems", func() bool { return product.Properties["tags"].MinItems != nil && *product.Properties["tags"].MinItems == 1 }},
		{"tags", "MaxItems", func() bool { return product.Properties["tags"].MaxItems != nil && *product.Properties["tags"].MaxItems == 10 }},
		{"categories", "UniqueItems", func() bool { return product.Properties["categories"].UniqueItems }},
		{"discount", "MultipleOf", func() bool { return product.Properties["discount"].MultipleOf != nil && *product.Properties["discount"].MultipleOf == 0.01 }},
	}

	for _, tt := range tests {
		t.Run(tt.field+"_"+tt.check, func(t *testing.T) {
			prop := product.Properties[tt.field]
			if prop == nil {
				t.Fatalf("Property %s not found", tt.field)
			}
			if !tt.validate() {
				t.Errorf("%s.%s validation failed", tt.field, tt.check)
			}
		})
	}
}

