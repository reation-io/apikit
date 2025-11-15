package petstore

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/reation-io/apikit/openapi/builder"
	"gopkg.in/yaml.v3"
)

func TestPetstoreGeneration(t *testing.T) {
	// Get the current directory
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Build OpenAPI spec from petstore.go
	pattern := filepath.Join(dir, "petstore.go")
	b := builder.NewBuilder(pattern)
	spec, err := b.Build()
	if err != nil {
		t.Fatalf("Failed to build OpenAPI spec: %v", err)
	}

	// Marshal to YAML
	yamlData, err := yaml.Marshal(spec)
	if err != nil {
		t.Fatalf("Failed to marshal to YAML: %v", err)
	}

	// Write generated spec to file
	generatedPath := filepath.Join(dir, "generated.yml")
	if err := os.WriteFile(generatedPath, yamlData, 0644); err != nil {
		t.Fatalf("Failed to write generated spec: %v", err)
	}

	t.Logf("✓ Generated OpenAPI spec: %s", generatedPath)

	// Verify basic structure
	if spec.OpenAPI == "" {
		t.Error("OpenAPI version is empty")
	}
	if spec.Info == nil {
		t.Error("Info is nil")
	}
	if spec.Paths == nil || spec.Paths.PathItems == nil {
		t.Error("Paths is nil")
	}

	// Count routes
	routeCount := 0
	for _, pathItem := range spec.Paths.PathItems {
		if pathItem.Get != nil {
			routeCount++
		}
		if pathItem.Post != nil {
			routeCount++
		}
		if pathItem.Put != nil {
			routeCount++
		}
		if pathItem.Delete != nil {
			routeCount++
		}
	}

	t.Logf("✓ Generated %d routes", routeCount)

	// Verify we have the expected routes
	expectedPaths := []string{
		"/pet",
		"/pet/findByStatus",
		"/pet/findByTags",
		"/pet/{petId}",
		"/pet/{petId}/uploadImage",
		"/store/inventory",
		"/store/order",
		"/store/order/{orderId}",
	}

	for _, path := range expectedPaths {
		if _, exists := spec.Paths.PathItems[path]; !exists {
			t.Errorf("Expected path %s not found", path)
		}
	}

	// Verify models
	if spec.Components == nil || spec.Components.Schemas == nil {
		t.Error("Components.Schemas is nil")
	} else {
		expectedModels := []string{
			"Pet",
			"Category",
			"Tag",
			"Order",
			"User",
			"ApiResponse",
			"Error",
		}

		for _, model := range expectedModels {
			if _, exists := spec.Components.Schemas[model]; !exists {
				t.Errorf("Expected model %s not found", model)
			}
		}

		t.Logf("✓ Generated %d models", len(spec.Components.Schemas))
	}

	// Verify security and responses are present in routes
	petRoute := spec.Paths.PathItems["/pet"]
	if petRoute != nil && petRoute.Put != nil {
		if len(petRoute.Put.Security) == 0 {
			t.Error("PUT /pet should have security requirements")
		}
		responseCount := len(petRoute.Put.Responses.StatusCodeResponses)
		if petRoute.Put.Responses.Default != nil {
			responseCount++
		}
		if responseCount == 0 {
			t.Error("PUT /pet should have responses")
		}
		t.Logf("✓ PUT /pet has %d security schemes and %d responses",
			len(petRoute.Put.Security), responseCount)
	}
}

func TestPetstoreJSONOutput(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	pattern := filepath.Join(dir, "petstore.go")
	b := builder.NewBuilder(pattern)
	spec, err := b.Build()
	if err != nil {
		t.Fatalf("Failed to build OpenAPI spec: %v", err)
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal to JSON: %v", err)
	}

	// Write JSON output
	jsonPath := filepath.Join(dir, "generated.json")
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		t.Fatalf("Failed to write JSON spec: %v", err)
	}

	t.Logf("✓ Generated JSON spec: %s", jsonPath)
}
