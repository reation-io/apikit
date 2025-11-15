package tags

import (
	"go/ast"
	"testing"

	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/spec"
)

func TestSecurityParser(t *testing.T) {
	tests := []struct {
		name     string
		comments string
		want     []spec.SecurityRequirement
		wantErr  bool
	}{
		{
			name: "single security scheme",
			comments: `swagger:route GET /users user listUsers
Security:
- bearer`,
			want: []spec.SecurityRequirement{
				{"bearer": []string{}},
			},
		},
		{
			name: "multiple security schemes",
			comments: `swagger:route POST /users user createUser
Security:
- bearer
- api_key`,
			want: []spec.SecurityRequirement{
				{"bearer": []string{}},
				{"api_key": []string{}},
			},
		},
		{
			name: "oauth with scopes",
			comments: `swagger:route POST /users user createUser
Security:
- oauth:
  - read
  - write`,
			want: []spec.SecurityRequirement{
				{"oauth": []string{"read", "write"}},
			},
		},
		{
			name: "mixed security schemes",
			comments: `swagger:route POST /users user createUser
Security:
- bearer
- oauth:
  - read
  - write
- api_key`,
			want: []spec.SecurityRequirement{
				{"bearer": []string{}},
				{"oauth": []string{"read", "write"}},
				{"api_key": []string{}},
			},
		},
		{
			name: "with other directives",
			comments: `swagger:route GET /api/v2/agent/workspaces 'Agent Workspaces [Agent]' listWorkspaces

summary: List authenticated agent workspaces

Security:
- bearer

Responses:
- 200: WorkspaceAssignmentListResponse`,
			want: []spec.SecurityRequirement{
				{"bearer": []string{}},
			},
		},
		{
			name: "no security section",
			comments: `swagger:route GET /users user listUsers
summary: List users`,
			want: []spec.SecurityRequirement{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create comment group
			commentGroup := &ast.CommentGroup{
				List: []*ast.Comment{},
			}
			for _, line := range splitLines(tt.comments) {
				commentGroup.List = append(commentGroup.List, &ast.Comment{
					Text: "// " + line,
				})
			}

			// Create operation
			operation := &spec.Operation{}

			// Parse
			parser := &SecurityParser{
				BaseParser: parsers.NewBaseParser(
					"security",
					parsers.ParserTypeMultiLine,
					[]parsers.ParseContext{parsers.ContextRoute},
					nil,
				),
			}

			value, err := parser.Parse(commentGroup, parsers.ContextRoute)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Apply to operation
			if value != nil {
				err = parser.Apply(operation, value, parsers.ContextRoute)
				if err != nil {
					t.Errorf("Apply() error = %v", err)
					return
				}
			}

			// Verify security requirements
			if len(operation.Security) != len(tt.want) {
				t.Errorf("Expected %d security requirements, got %d", len(tt.want), len(operation.Security))
				return
			}

			for i, wantReq := range tt.want {
				if i >= len(operation.Security) {
					t.Errorf("Missing security requirement at index %d", i)
					continue
				}

				gotReq := operation.Security[i]

				// Check that both have the same keys
				if len(gotReq) != len(wantReq) {
					t.Errorf("Requirement %d: expected %d schemes, got %d", i, len(wantReq), len(gotReq))
					continue
				}

				// Check each scheme
				for scheme, wantScopes := range wantReq {
					gotScopes, ok := gotReq[scheme]
					if !ok {
						t.Errorf("Requirement %d: missing scheme %q", i, scheme)
						continue
					}

					// Check scopes
					if len(gotScopes) != len(wantScopes) {
						t.Errorf("Requirement %d, scheme %q: expected %d scopes, got %d", i, scheme, len(wantScopes), len(gotScopes))
						continue
					}

					for j, wantScope := range wantScopes {
						if j >= len(gotScopes) {
							t.Errorf("Requirement %d, scheme %q: missing scope at index %d", i, scheme, j)
							continue
						}

						if gotScopes[j] != wantScope {
							t.Errorf("Requirement %d, scheme %q, scope %d: expected %q, got %q", i, scheme, j, wantScope, gotScopes[j])
						}
					}
				}
			}
		})
	}
}
