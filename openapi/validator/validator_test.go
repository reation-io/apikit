package validator

import (
	"testing"

	"github.com/reation-io/apikit/openapi/spec"
)

func TestValidator_BasicStructure(t *testing.T) {
	t.Run("valid spec", func(t *testing.T) {
		s := createValidSpec()
		v := New(s)
		result := v.Validate()

		if result.HasErrors() {
			t.Errorf("expected no errors, got %v", result.Errors)
		}
	})

	t.Run("missing openapi version", func(t *testing.T) {
		s := createValidSpec()
		s.OpenAPI = ""

		v := New(s)
		result := v.Validate()

		if !result.HasErrors() {
			t.Error("expected errors for missing OpenAPI version")
		}
	})

	t.Run("invalid openapi version", func(t *testing.T) {
		s := createValidSpec()
		s.OpenAPI = "2.0"

		v := New(s)
		result := v.Validate()

		if !result.HasErrors() {
			t.Error("expected errors for invalid OpenAPI version")
		}
	})

	t.Run("missing info", func(t *testing.T) {
		s := createValidSpec()
		s.Info = nil

		v := New(s)
		result := v.Validate()

		if !result.HasErrors() {
			t.Error("expected errors for missing info")
		}
	})
}

func TestValidator_Info(t *testing.T) {
	t.Run("missing title", func(t *testing.T) {
		s := createValidSpec()
		s.Info.Title = ""

		v := New(s)
		result := v.Validate()

		if !result.HasErrors() {
			t.Error("expected errors for missing title")
		}
	})

	t.Run("missing version", func(t *testing.T) {
		s := createValidSpec()
		s.Info.Version = ""

		v := New(s)
		result := v.Validate()

		if !result.HasErrors() {
			t.Error("expected errors for missing version")
		}
	})

	t.Run("invalid contact email", func(t *testing.T) {
		s := createValidSpec()
		s.Info.Contact = &spec.Contact{
			Email: "invalid-email",
		}

		v := New(s)
		result := v.Validate()

		if !result.HasWarnings() {
			t.Error("expected warnings for invalid email")
		}
	})

	t.Run("valid contact", func(t *testing.T) {
		s := createValidSpec()
		s.Info.Contact = &spec.Contact{
			Email: "test@example.com",
			URL:   "https://example.com",
		}

		v := New(s)
		result := v.Validate()

		// Should not have contact-related warnings
		for _, w := range result.Warnings {
			if w.Path == "info.contact.email" || w.Path == "info.contact.url" {
				t.Errorf("unexpected warning: %v", w)
			}
		}
	})
}

func TestValidator_Schema(t *testing.T) {
	t.Run("invalid type", func(t *testing.T) {
		s := createValidSpec()
		s.Components = &spec.Components{
			Schemas: map[string]*spec.Schema{
				"Test": {
					Type: "invalid",
				},
			},
		}

		v := New(s)
		result := v.Validate()

		if !result.HasErrors() {
			t.Error("expected errors for invalid type")
		}
	})

	t.Run("array without items", func(t *testing.T) {
		s := createValidSpec()
		s.Components = &spec.Components{
			Schemas: map[string]*spec.Schema{
				"Test": {
					Type: "array",
					// Items is missing
				},
			},
		}

		v := New(s)
		result := v.Validate()

		if !result.HasErrors() {
			t.Error("expected errors for array without items")
		}
	})

	t.Run("invalid min/max", func(t *testing.T) {
		s := createValidSpec()
		min := 100.0
		max := 10.0
		s.Components = &spec.Components{
			Schemas: map[string]*spec.Schema{
				"Test": {
					Type:    "integer",
					Minimum: &min,
					Maximum: &max,
				},
			},
		}

		v := New(s)
		result := v.Validate()

		if !result.HasErrors() {
			t.Error("expected errors for min > max")
		}
	})

	t.Run("invalid regex pattern", func(t *testing.T) {
		s := createValidSpec()
		s.Components = &spec.Components{
			Schemas: map[string]*spec.Schema{
				"Test": {
					Type:    "string",
					Pattern: "[invalid",
				},
			},
		}

		v := New(s)
		result := v.Validate()

		if !result.HasErrors() {
			t.Error("expected errors for invalid regex")
		}
	})
}

func TestValidator_Operation(t *testing.T) {
	t.Run("missing responses", func(t *testing.T) {
		s := createValidSpec()
		s.Paths.PathItems["/test"] = &spec.PathItem{
			Get: &spec.Operation{
				Responses: nil,
			},
		}

		v := New(s)
		result := v.Validate()

		if !result.HasErrors() {
			t.Error("expected errors for missing responses")
		}
	})

	t.Run("empty responses", func(t *testing.T) {
		s := createValidSpec()
		s.Paths.PathItems["/test"] = &spec.PathItem{
			Get: &spec.Operation{
				Responses: &spec.Responses{},
			},
		}

		v := New(s)
		result := v.Validate()

		if !result.HasErrors() {
			t.Error("expected errors for empty responses")
		}
	})

	t.Run("path parameter not required", func(t *testing.T) {
		s := createValidSpec()
		s.Paths.PathItems["/users/{id}"] = &spec.PathItem{
			Get: &spec.Operation{
				Parameters: []*spec.Parameter{
					{
						Name:     "id",
						In:       "path",
						Required: false, // Should be true
					},
				},
				Responses: &spec.Responses{
					StatusCodeResponses: map[string]*spec.Response{
						"200": {Description: "OK"},
					},
				},
			},
		}

		v := New(s)
		result := v.Validate()

		if !result.HasErrors() {
			t.Error("expected errors for path parameter not required")
		}
	})

	t.Run("invalid parameter location", func(t *testing.T) {
		s := createValidSpec()
		s.Paths.PathItems["/test"] = &spec.PathItem{
			Get: &spec.Operation{
				Parameters: []*spec.Parameter{
					{
						Name: "param",
						In:   "invalid",
					},
				},
				Responses: &spec.Responses{
					StatusCodeResponses: map[string]*spec.Response{
						"200": {Description: "OK"},
					},
				},
			},
		}

		v := New(s)
		result := v.Validate()

		if !result.HasErrors() {
			t.Error("expected errors for invalid parameter location")
		}
	})
}

func TestValidator_SecurityScheme(t *testing.T) {
	t.Run("invalid security type", func(t *testing.T) {
		s := createValidSpec()
		s.Components = &spec.Components{
			SecuritySchemes: map[string]*spec.SecurityScheme{
				"test": {
					Type: "invalid",
				},
			},
		}

		v := New(s)
		result := v.Validate()

		if !result.HasErrors() {
			t.Error("expected errors for invalid security type")
		}
	})

	t.Run("apiKey missing name", func(t *testing.T) {
		s := createValidSpec()
		s.Components = &spec.Components{
			SecuritySchemes: map[string]*spec.SecurityScheme{
				"test": {
					Type: "apiKey",
					In:   "header",
					// Name is missing
				},
			},
		}

		v := New(s)
		result := v.Validate()

		if !result.HasErrors() {
			t.Error("expected errors for apiKey missing name")
		}
	})

	t.Run("http missing scheme", func(t *testing.T) {
		s := createValidSpec()
		s.Components = &spec.Components{
			SecuritySchemes: map[string]*spec.SecurityScheme{
				"test": {
					Type: "http",
					// Scheme is missing
				},
			},
		}

		v := New(s)
		result := v.Validate()

		if !result.HasErrors() {
			t.Error("expected errors for http missing scheme")
		}
	})
}

func TestValidationResult(t *testing.T) {
	t.Run("empty result is valid", func(t *testing.T) {
		result := &ValidationResult{}

		if !result.IsValid() {
			t.Error("empty result should be valid")
		}
		if result.HasErrors() {
			t.Error("empty result should have no errors")
		}
		if result.HasWarnings() {
			t.Error("empty result should have no warnings")
		}
	})

	t.Run("result with errors is invalid", func(t *testing.T) {
		result := &ValidationResult{
			Errors: []ValidationError{{Message: "test"}},
		}

		if result.IsValid() {
			t.Error("result with errors should be invalid")
		}
		if !result.HasErrors() {
			t.Error("result should have errors")
		}
	})

	t.Run("result with only warnings is valid", func(t *testing.T) {
		result := &ValidationResult{
			Warnings: []ValidationError{{Message: "test"}},
		}

		if !result.IsValid() {
			t.Error("result with only warnings should be valid")
		}
		if !result.HasWarnings() {
			t.Error("result should have warnings")
		}
	})
}

// Helper functions

func createValidSpec() *spec.OpenAPI {
	return &spec.OpenAPI{
		OpenAPI: "3.0.3",
		Info: &spec.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: &spec.Paths{
			PathItems: map[string]*spec.PathItem{
				"/health": {
					Get: &spec.Operation{
						Responses: &spec.Responses{
							StatusCodeResponses: map[string]*spec.Response{
								"200": {Description: "OK"},
							},
						},
					},
				},
			},
		},
	}
}
