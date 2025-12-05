package tags

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/spec"
)

func TestCompositionParser_OneOf(t *testing.T) {
	tests := []struct {
		name           string
		comment        string
		expectedCount  int
		expectedNames  []string
		shouldMatch    bool
	}{
		{
			name: "comma separated",
			comment: `// swagger:model PaymentMethod
// oneOf: CreditCard, BankTransfer, Crypto`,
			expectedCount: 3,
			expectedNames: []string{"CreditCard", "BankTransfer", "Crypto"},
			shouldMatch:   true,
		},
		{
			name: "space separated",
			comment: `// swagger:model PaymentMethod
// oneOf: CreditCard BankTransfer`,
			expectedCount: 2,
			expectedNames: []string{"CreditCard", "BankTransfer"},
			shouldMatch:   true,
		},
		{
			name: "no oneOf",
			comment: `// swagger:model SimpleModel
// Description only`,
			expectedCount: 0,
			shouldMatch:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comments := parseTestComment(t, tt.comment)

			parser := &CompositionParser{
				BaseParser: parsers.NewBaseParser(
					"oneOf",
					parsers.ParserTypeMultiLine,
					[]parsers.ParseContext{parsers.ContextModel},
					nil,
				),
				compositionType: "oneOf",
			}

			// Check Matches
			if got := parser.Matches(comments.Text(), parsers.ContextModel); got != tt.shouldMatch {
				t.Errorf("Matches() = %v, want %v", got, tt.shouldMatch)
			}

			if !tt.shouldMatch {
				return
			}

			// Parse
			value, err := parser.Parse(comments, parsers.ContextModel)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			composition, ok := value.(*ParsedComposition)
			if !ok {
				t.Fatalf("expected *ParsedComposition, got %T", value)
			}

			if len(composition.Schemas) != tt.expectedCount {
				t.Errorf("expected %d schemas, got %d", tt.expectedCount, len(composition.Schemas))
			}

			// Apply to schema
			schema := &spec.Schema{}
			if err := parser.Apply(schema, value, parsers.ContextModel); err != nil {
				t.Fatalf("Apply() error = %v", err)
			}

			if len(schema.OneOf) != tt.expectedCount {
				t.Errorf("expected %d OneOf schemas, got %d", tt.expectedCount, len(schema.OneOf))
			}

			// Verify references
			for i, expectedName := range tt.expectedNames {
				expectedRef := "#/components/schemas/" + expectedName
				if schema.OneOf[i].Ref != expectedRef {
					t.Errorf("expected ref %q, got %q", expectedRef, schema.OneOf[i].Ref)
				}
			}
		})
	}
}

func TestCompositionParser_AllOf(t *testing.T) {
	comment := `// swagger:model ExtendedUser
// allOf: BaseUser, AdminFields`

	comments := parseTestComment(t, comment)

	parser := &CompositionParser{
		BaseParser: parsers.NewBaseParser(
			"allOf",
			parsers.ParserTypeMultiLine,
			[]parsers.ParseContext{parsers.ContextModel},
			nil,
		),
		compositionType: "allOf",
	}

	if !parser.Matches(comments.Text(), parsers.ContextModel) {
		t.Fatal("expected to match allOf")
	}

	value, err := parser.Parse(comments, parsers.ContextModel)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	schema := &spec.Schema{}
	if err := parser.Apply(schema, value, parsers.ContextModel); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	if len(schema.AllOf) != 2 {
		t.Errorf("expected 2 AllOf schemas, got %d", len(schema.AllOf))
	}
}

func TestCompositionParser_AnyOf(t *testing.T) {
	comment := `// swagger:model FlexibleType
// anyOf: TypeA, TypeB, TypeC`

	comments := parseTestComment(t, comment)

	parser := &CompositionParser{
		BaseParser: parsers.NewBaseParser(
			"anyOf",
			parsers.ParserTypeMultiLine,
			[]parsers.ParseContext{parsers.ContextModel},
			nil,
		),
		compositionType: "anyOf",
	}

	if !parser.Matches(comments.Text(), parsers.ContextModel) {
		t.Fatal("expected to match anyOf")
	}

	value, err := parser.Parse(comments, parsers.ContextModel)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	schema := &spec.Schema{}
	if err := parser.Apply(schema, value, parsers.ContextModel); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	if len(schema.AnyOf) != 3 {
		t.Errorf("expected 3 AnyOf schemas, got %d", len(schema.AnyOf))
	}
}

func parseTestComment(t *testing.T, comment string) *ast.CommentGroup {
	t.Helper()

	src := comment + "\ntype Test struct{}"
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", "package test\n\n"+src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse comment: %v", err)
	}

	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Doc != nil {
			return genDecl.Doc
		}
	}

	t.Fatal("no comments found")
	return nil
}
