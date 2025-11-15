package tags

import (
	"go/ast"
	"testing"

	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/spec"
)

func TestResponsesParser(t *testing.T) {
	tests := []struct {
		name     string
		comments string
		want     map[string]string // statusCode -> schema ref
		wantErr  bool
	}{
		{
			name: "single response",
			comments: `swagger:route GET /users user listUsers
Responses:
- 200: UserListResponse`,
			want: map[string]string{
				"200": "#/components/schemas/UserListResponse",
			},
		},
		{
			name: "multiple responses",
			comments: `swagger:route POST /users user createUser
Responses:
- 200: User
- 400: ErrorResponse
- 401: UnauthorizedResponse`,
			want: map[string]string{
				"200": "#/components/schemas/User",
				"400": "#/components/schemas/ErrorResponse",
				"401": "#/components/schemas/UnauthorizedResponse",
			},
		},
		{
			name: "with default response",
			comments: `swagger:route GET /users user listUsers
Responses:
- 200: UserListResponse
- default: ErrorResponse`,
			want: map[string]string{
				"200":     "#/components/schemas/UserListResponse",
				"default": "#/components/schemas/ErrorResponse",
			},
		},
		{
			name: "with other directives",
			comments: `swagger:route GET /api/v2/agent/workspaces 'Agent Workspaces [Agent]' listWorkspaces

List all Workspaces assigned to the authenticated Agent.

summary: List authenticated agent workspaces
description: Authenticated agents can retrieve all workspace assignments linked to their account.

Security:
- bearer

Responses:
- 200: WorkspaceAssignmentListResponse
- 401: UnauthorizedResponse`,
			want: map[string]string{
				"200": "#/components/schemas/WorkspaceAssignmentListResponse",
				"401": "#/components/schemas/UnauthorizedResponse",
			},
		},
		{
			name: "no responses section",
			comments: `swagger:route GET /users user listUsers
summary: List users`,
			want: map[string]string{},
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
			operation := &spec.Operation{
				Responses: &spec.Responses{
					StatusCodeResponses: make(map[string]*spec.Response),
				},
			}

			// Parse
			parser := &ResponsesParser{
				BaseParser: parsers.NewBaseParser(
					"responses",
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

			// Verify responses
			for statusCode, expectedRef := range tt.want {
				var response *spec.Response
				if statusCode == "default" {
					response = operation.Responses.Default
				} else {
					response = operation.Responses.StatusCodeResponses[statusCode]
				}

				if response == nil {
					t.Errorf("Expected response for status %s, got nil", statusCode)
					continue
				}

				if response.Content == nil || response.Content["application/json"] == nil {
					t.Errorf("Expected JSON content for status %s", statusCode)
					continue
				}

				schema := response.Content["application/json"].Schema
				if schema == nil {
					t.Errorf("Expected schema for status %s", statusCode)
					continue
				}

				if schema.Ref != expectedRef {
					t.Errorf("Status %s: expected ref %q, got %q", statusCode, expectedRef, schema.Ref)
				}
			}

			// Verify no extra responses
			totalResponses := len(operation.Responses.StatusCodeResponses)
			if operation.Responses.Default != nil {
				totalResponses++
			}

			if totalResponses != len(tt.want) {
				t.Errorf("Expected %d responses, got %d", len(tt.want), totalResponses)
			}
		})
	}
}

func splitLines(s string) []string {
	lines := []string{}
	current := ""
	for _, c := range s {
		if c == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}
