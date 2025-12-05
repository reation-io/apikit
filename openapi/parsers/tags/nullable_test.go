package tags

import (
	"go/ast"
	"testing"

	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/spec"
)

func TestNullableParser_Parse(t *testing.T) {
	tests := []struct {
		name           string
		comments       []string
		expectedResult bool
		expectNoResult bool
	}{
		{
			name:           "nullable true",
			comments:       []string{"// nullable: true"},
			expectedResult: true,
		},
		{
			name:           "nullable false",
			comments:       []string{"// nullable: false"},
			expectedResult: false,
		},
		{
			name:           "nullable yes",
			comments:       []string{"// nullable: yes"},
			expectedResult: true,
		},
		{
			name:           "no nullable directive",
			comments:       []string{"// Description: Some field"},
			expectNoResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &NullableParser{}

			var commentList []*ast.Comment
			for _, c := range tt.comments {
				commentList = append(commentList, &ast.Comment{Text: c})
			}
			comments := &ast.CommentGroup{List: commentList}

			result, err := parser.Parse(comments, parsers.ContextField)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expectNoResult {
				if result != nil {
					t.Errorf("expected no result, got %v", result)
				}
				return
			}

			nullable, ok := result.(bool)
			if !ok {
				t.Fatalf("expected bool result, got %T", result)
			}

			if nullable != tt.expectedResult {
				t.Errorf("expected %v, got %v", tt.expectedResult, nullable)
			}
		})
	}
}

func TestNullableParser_Apply(t *testing.T) {
	parser := &NullableParser{}

	schema := &spec.Schema{}

	err := parser.Apply(schema, true, parsers.ContextField)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !schema.Nullable {
		t.Error("expected schema to be nullable")
	}

	// Test applying false
	err = parser.Apply(schema, false, parsers.ContextField)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if schema.Nullable {
		t.Error("expected schema to not be nullable")
	}
}
