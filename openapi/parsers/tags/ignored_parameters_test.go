package tags

import (
	"go/ast"
	"testing"

	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/spec"
)

func TestIgnoredParametersParser_Parse(t *testing.T) {
	tests := []struct {
		name           string
		comments       []string
		expectedParams []string
		expectNoResult bool
	}{
		{
			name: "comma separated",
			comments: []string{
				"// swagger:route GET /users users listUsers",
				"// IgnoredParameters: token, session",
			},
			expectedParams: []string{"token", "session"},
		},
		{
			name: "space separated",
			comments: []string{
				"// swagger:route GET /users users listUsers",
				"// IgnoredParameters: token session requestId",
			},
			expectedParams: []string{"token", "session", "requestId"},
		},
		{
			name: "single parameter",
			comments: []string{
				"// swagger:route GET /users users listUsers",
				"// IgnoredParameters: internalId",
			},
			expectedParams: []string{"internalId"},
		},
		{
			name: "no ignored parameters",
			comments: []string{
				"// swagger:route GET /users users listUsers",
				"// Summary: List all users",
			},
			expectNoResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &IgnoredParametersParser{}

			var commentList []*ast.Comment
			for _, c := range tt.comments {
				commentList = append(commentList, &ast.Comment{Text: c})
			}
			comments := &ast.CommentGroup{List: commentList}

			result, err := parser.Parse(comments, parsers.ContextRoute)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expectNoResult {
				if result != nil {
					t.Errorf("expected no result, got %v", result)
				}
				return
			}

			params, ok := result.([]string)
			if !ok {
				t.Fatalf("expected []string result, got %T", result)
			}

			if len(params) != len(tt.expectedParams) {
				t.Errorf("expected %d params, got %d: %v", len(tt.expectedParams), len(params), params)
			}

			for i, expected := range tt.expectedParams {
				if i >= len(params) || params[i] != expected {
					t.Errorf("expected param %d to be %q", i, expected)
				}
			}
		})
	}
}

func TestIgnoredParametersParser_Apply(t *testing.T) {
	parser := &IgnoredParametersParser{}

	routeInfo := spec.NewRouteInfo()
	params := []string{"token", "session"}

	err := parser.Apply(routeInfo, params, parsers.ContextRoute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(routeInfo.IgnoredParameters) != 2 {
		t.Errorf("expected 2 ignored params, got %d", len(routeInfo.IgnoredParameters))
	}

	if !routeInfo.HasIgnoredParameter("token") {
		t.Error("expected 'token' to be in ignored parameters")
	}
	if !routeInfo.HasIgnoredParameter("session") {
		t.Error("expected 'session' to be in ignored parameters")
	}
}
