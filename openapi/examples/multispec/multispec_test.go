package multispec

import (
	"os"
	"path/filepath"
	"testing"

	coreast "github.com/reation-io/apikit/core/ast"
	"github.com/reation-io/apikit/openapi/builder"
	"gopkg.in/yaml.v3"
)

func TestMultiSpecGeneration(t *testing.T) {
	// Get the path to the handlers file
	handlersPath, err := filepath.Abs("handlers.go")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Parse the file
	parser := coreast.NewCachedParser()
	result, err := parser.Parse(handlersPath)
	if err != nil {
		t.Fatalf("Failed to parse handlers.go: %v", err)
	}

	// Extract multiple specs
	specs, err := builder.ExtractMultipleFromGeneric([]*coreast.ParseResult{result})
	if err != nil {
		t.Fatalf("Failed to extract specs: %v", err)
	}

	// Verify we have the expected specs
	expectedSpecs := []string{"admin", "mobile", "public", "default"}
	for _, specName := range expectedSpecs {
		if _, ok := specs[specName]; !ok {
			t.Errorf("Expected spec %q not found", specName)
		}
	}

	// Test admin spec
	t.Run("AdminSpec", func(t *testing.T) {
		adminSpec, ok := specs["admin"]
		if !ok {
			t.Fatal("Admin spec not found")
		}

		// Should have 2 path items: /admin/users, /users/{id}
		expectedPaths := 2
		if len(adminSpec.Paths.PathItems) != expectedPaths {
			t.Errorf("Expected %d path items in admin spec, got %d", expectedPaths, len(adminSpec.Paths.PathItems))
		}

		// Verify /admin/users exists
		if adminSpec.Paths.PathItems["/admin/users"] == nil {
			t.Error("Expected /admin/users path in admin spec")
		} else {
			if adminSpec.Paths.PathItems["/admin/users"].Get == nil {
				t.Error("Expected GET /admin/users in admin spec")
			}
			if adminSpec.Paths.PathItems["/admin/users"].Post == nil {
				t.Error("Expected POST /admin/users in admin spec")
			}
		}

		// Verify /users/{id} exists (shared route)
		if adminSpec.Paths.PathItems["/users/{id}"] == nil {
			t.Error("Expected /users/{id} path in admin spec")
		}

		// Verify title
		if adminSpec.Info.Title != "Admin API" {
			t.Errorf("Expected title 'Admin API', got %q", adminSpec.Info.Title)
		}
	})

	// Test mobile spec
	t.Run("MobileSpec", func(t *testing.T) {
		mobileSpec, ok := specs["mobile"]
		if !ok {
			t.Fatal("Mobile spec not found")
		}

		// Should have 2 path items: /mobile/profile, /users/{id}
		expectedPaths := 2
		if len(mobileSpec.Paths.PathItems) != expectedPaths {
			t.Errorf("Expected %d path items in mobile spec, got %d", expectedPaths, len(mobileSpec.Paths.PathItems))
		}

		// Verify /mobile/profile exists
		if mobileSpec.Paths.PathItems["/mobile/profile"] == nil {
			t.Error("Expected /mobile/profile path in mobile spec")
		} else {
			if mobileSpec.Paths.PathItems["/mobile/profile"].Get == nil {
				t.Error("Expected GET /mobile/profile in mobile spec")
			}
			if mobileSpec.Paths.PathItems["/mobile/profile"].Put == nil {
				t.Error("Expected PUT /mobile/profile in mobile spec")
			}
		}

		// Verify /users/{id} exists (shared route)
		if mobileSpec.Paths.PathItems["/users/{id}"] == nil {
			t.Error("Expected /users/{id} path in mobile spec")
		}

		// Verify title
		if mobileSpec.Info.Title != "Mobile API" {
			t.Errorf("Expected title 'Mobile API', got %q", mobileSpec.Info.Title)
		}
	})

	// Test public spec
	t.Run("PublicSpec", func(t *testing.T) {
		publicSpec, ok := specs["public"]
		if !ok {
			t.Fatal("Public spec not found")
		}

		// Should have 2 routes: /users/{id} (GET), /public/info (GET)
		expectedPaths := 2
		if len(publicSpec.Paths.PathItems) != expectedPaths {
			t.Errorf("Expected %d paths in public spec, got %d", expectedPaths, len(publicSpec.Paths.PathItems))
		}

		// Verify /users/{id} exists (shared route)
		if publicSpec.Paths.PathItems["/users/{id}"] == nil {
			t.Error("Expected /users/{id} path in public spec")
		}

		// Verify /public/info exists
		if publicSpec.Paths.PathItems["/public/info"] == nil {
			t.Error("Expected /public/info path in public spec")
		}

		// Verify title
		if publicSpec.Info.Title != "Public API" {
			t.Errorf("Expected title 'Public API', got %q", publicSpec.Info.Title)
		}
	})

	// Test default spec
	t.Run("DefaultSpec", func(t *testing.T) {
		defaultSpec, ok := specs["default"]
		if !ok {
			t.Fatal("Default spec not found")
		}

		// Should have 1 route: /health (GET) - routes without Spec: tag
		expectedPaths := 1
		if len(defaultSpec.Paths.PathItems) != expectedPaths {
			t.Errorf("Expected %d path in default spec, got %d", expectedPaths, len(defaultSpec.Paths.PathItems))
		}

		// Verify /health exists
		if defaultSpec.Paths.PathItems["/health"] == nil {
			t.Error("Expected /health path in default spec")
		}
	})

	// Test that all specs have the shared models
	t.Run("SharedModels", func(t *testing.T) {
		expectedModels := []string{"User", "ErrorResponse", "HealthResponse"}

		for specName, spec := range specs {
			if spec.Components == nil || spec.Components.Schemas == nil {
				t.Errorf("Spec %q has no schemas", specName)
				continue
			}

			for _, modelName := range expectedModels {
				if _, ok := spec.Components.Schemas[modelName]; !ok {
					t.Errorf("Spec %q missing model %q", specName, modelName)
				}
			}
		}
	})
}

func TestMultiSpecGeneration_YAMLOutput(t *testing.T) {
	// Get the path to the handlers file
	handlersPath, err := filepath.Abs("handlers.go")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Parse the file
	parser := coreast.NewCachedParser()
	result, err := parser.Parse(handlersPath)
	if err != nil {
		t.Fatalf("Failed to parse handlers.go: %v", err)
	}

	// Extract multiple specs
	specs, err := builder.ExtractMultipleFromGeneric([]*coreast.ParseResult{result})
	if err != nil {
		t.Fatalf("Failed to extract specs: %v", err)
	}

	// Write each spec to YAML
	for specName, spec := range specs {
		// Skip empty specs
		if len(spec.Paths.PathItems) == 0 {
			continue
		}

		filename := specName + ".yml"
		output, err := yaml.Marshal(spec)
		if err != nil {
			t.Fatalf("Failed to marshal %s to YAML: %v", specName, err)
		}

		if err := os.WriteFile(filename, output, 0644); err != nil {
			t.Fatalf("Failed to write %s: %v", filename, err)
		}

		t.Logf("Generated %s", filename)

		// Clean up
		defer os.Remove(filename)
	}
}
