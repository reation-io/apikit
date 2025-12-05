package builder

import (
	"testing"

	"github.com/reation-io/apikit/openapi/spec"
)

func TestCleanupUnreferencedSchemas(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(*Builder)
		expectedSchemas []string
		removedSchemas  []string
	}{
		{
			name: "keeps referenced schemas",
			setup: func(b *Builder) {
				b.spec.Components = &spec.Components{
					Schemas: map[string]*spec.Schema{
						"User": {Type: "object"},
						"Pet":  {Type: "object"},
					},
				}
				// Add a route that references User
				b.spec.Paths.PathItems["/users"] = &spec.PathItem{
					Get: &spec.Operation{
						Responses: &spec.Responses{
							StatusCodeResponses: map[string]*spec.Response{
								"200": {
									Content: map[string]*spec.MediaType{
										"application/json": {
											Schema: &spec.Schema{
												Ref: "#/components/schemas/User",
											},
										},
									},
								},
							},
						},
					},
				}
			},
			expectedSchemas: []string{"User"},
			removedSchemas:  []string{"Pet"},
		},
		{
			name: "keeps transitively referenced schemas",
			setup: func(b *Builder) {
				b.spec.Components = &spec.Components{
					Schemas: map[string]*spec.Schema{
						"User": {
							Type: "object",
							Properties: map[string]*spec.Schema{
								"address": {Ref: "#/components/schemas/Address"},
							},
						},
						"Address": {Type: "object"},
						"Unused":  {Type: "object"},
					},
				}
				// Add a route that references User
				b.spec.Paths.PathItems["/users"] = &spec.PathItem{
					Get: &spec.Operation{
						Responses: &spec.Responses{
							StatusCodeResponses: map[string]*spec.Response{
								"200": {
									Content: map[string]*spec.MediaType{
										"application/json": {
											Schema: &spec.Schema{
												Ref: "#/components/schemas/User",
											},
										},
									},
								},
							},
						},
					},
				}
			},
			expectedSchemas: []string{"User", "Address"},
			removedSchemas:  []string{"Unused"},
		},
		{
			name: "handles oneOf/allOf/anyOf references",
			setup: func(b *Builder) {
				b.spec.Components = &spec.Components{
					Schemas: map[string]*spec.Schema{
						"Payment": {
							OneOf: []*spec.Schema{
								{Ref: "#/components/schemas/CreditCard"},
								{Ref: "#/components/schemas/BankTransfer"},
							},
						},
						"CreditCard":   {Type: "object"},
						"BankTransfer": {Type: "object"},
						"Unused":       {Type: "object"},
					},
				}
				b.spec.Paths.PathItems["/payments"] = &spec.PathItem{
					Post: &spec.Operation{
						RequestBody: &spec.RequestBody{
							Content: map[string]*spec.MediaType{
								"application/json": {
									Schema: &spec.Schema{
										Ref: "#/components/schemas/Payment",
									},
								},
							},
						},
						Responses: &spec.Responses{
							StatusCodeResponses: map[string]*spec.Response{
								"200": {Description: "OK"},
							},
						},
					},
				}
			},
			expectedSchemas: []string{"Payment", "CreditCard", "BankTransfer"},
			removedSchemas:  []string{"Unused"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder()
			tt.setup(b)

			b.CleanupUnreferencedSchemas()

			// Check expected schemas still exist
			for _, schemaName := range tt.expectedSchemas {
				if _, ok := b.spec.Components.Schemas[schemaName]; !ok {
					t.Errorf("expected schema %s to exist after cleanup", schemaName)
				}
			}

			// Check removed schemas are gone
			for _, schemaName := range tt.removedSchemas {
				if _, ok := b.spec.Components.Schemas[schemaName]; ok {
					t.Errorf("expected schema %s to be removed after cleanup", schemaName)
				}
			}
		})
	}
}

