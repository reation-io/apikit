package tags

import (
	"fmt"
	"go/ast"
	"regexp"
	"strings"

	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/spec"
)

// CompositionParser parses composition directives (oneOf, allOf, anyOf) for models
// Format:
// swagger:model PaymentMethod
// oneOf: CreditCard, BankTransfer, Crypto
// or
// allOf: BaseModel, ExtendedFields
// or
// anyOf: TypeA, TypeB
type CompositionParser struct {
	parsers.BaseParser
	compositionType string // "oneOf", "allOf", or "anyOf"
}

func init() {
	// Register parsers for each composition type
	parsers.GlobalRegistry().Register("swagger:model", &CompositionParser{
		BaseParser: parsers.NewBaseParser(
			"oneOf",
			parsers.ParserTypeMultiLine,
			[]parsers.ParseContext{parsers.ContextModel},
			nil,
		),
		compositionType: "oneOf",
	})

	parsers.GlobalRegistry().Register("swagger:model", &CompositionParser{
		BaseParser: parsers.NewBaseParser(
			"allOf",
			parsers.ParserTypeMultiLine,
			[]parsers.ParseContext{parsers.ContextModel},
			nil,
		),
		compositionType: "allOf",
	})

	parsers.GlobalRegistry().Register("swagger:model", &CompositionParser{
		BaseParser: parsers.NewBaseParser(
			"anyOf",
			parsers.ParserTypeMultiLine,
			[]parsers.ParseContext{parsers.ContextModel},
			nil,
		),
		compositionType: "anyOf",
	})
}

// Pattern to extract composition schemas
var compositionPatterns = map[string]*regexp.Regexp{
	"oneOf": regexp.MustCompile(`(?mi)^oneOf:\s*(.+)$`),
	"allOf": regexp.MustCompile(`(?mi)^allOf:\s*(.+)$`),
	"anyOf": regexp.MustCompile(`(?mi)^anyOf:\s*(.+)$`),
}

// Matches checks if the comment contains the composition directive
func (p *CompositionParser) Matches(comment string, ctx parsers.ParseContext) bool {
	if ctx != parsers.ContextModel {
		return false
	}

	directive := p.compositionType + ":"
	return strings.Contains(strings.ToLower(comment), strings.ToLower(directive))
}

// Parse extracts composition schemas from the comment
func (p *CompositionParser) Parse(comments *ast.CommentGroup, ctx parsers.ParseContext) (any, error) {
	if ctx != parsers.ContextModel {
		return nil, nil
	}

	text := comments.Text()
	pattern := compositionPatterns[p.compositionType]
	if pattern == nil {
		return nil, nil
	}

	matches := pattern.FindStringSubmatch(text)
	if len(matches) < 2 {
		return nil, nil
	}

	// Parse the comma-separated list of schema names
	schemasStr := strings.TrimSpace(matches[1])
	schemaNames := parseSchemaList(schemasStr)

	if len(schemaNames) == 0 {
		return nil, nil
	}

	// Create schema references
	schemas := make([]*spec.Schema, 0, len(schemaNames))
	for _, name := range schemaNames {
		schemas = append(schemas, &spec.Schema{
			Ref: fmt.Sprintf("#/components/schemas/%s", name),
		})
	}

	return &ParsedComposition{
		Type:    p.compositionType,
		Schemas: schemas,
	}, nil
}

// Apply applies the parsed composition to the schema
func (p *CompositionParser) Apply(target any, value any, ctx parsers.ParseContext) error {
	if ctx != parsers.ContextModel {
		return nil
	}

	schema, ok := target.(*spec.Schema)
	if !ok {
		return &parsers.ErrInvalidTarget{
			ParserName:   p.compositionType,
			Context:      ctx,
			ExpectedType: "*spec.Schema",
			ActualType:   fmt.Sprintf("%T", target),
		}
	}

	composition, ok := value.(*ParsedComposition)
	if !ok {
		if value == nil {
			return nil
		}
		return &parsers.ErrInvalidValue{
			ParserName:   p.compositionType,
			ExpectedType: "*ParsedComposition",
			ActualType:   fmt.Sprintf("%T", value),
		}
	}

	// Apply composition based on type
	switch composition.Type {
	case "oneOf":
		schema.OneOf = composition.Schemas
	case "allOf":
		schema.AllOf = composition.Schemas
	case "anyOf":
		schema.AnyOf = composition.Schemas
	}

	return nil
}

// Name returns the parser name
func (p *CompositionParser) Name() string {
	return p.compositionType
}

// ParsedComposition holds the parsed composition data
type ParsedComposition struct {
	Type    string        // "oneOf", "allOf", or "anyOf"
	Schemas []*spec.Schema // Schema references
}

// parseSchemaList parses a comma or space separated list of schema names
func parseSchemaList(input string) []string {
	// Support both comma and space separation
	var result []string

	// First, try comma separation
	if strings.Contains(input, ",") {
		parts := strings.Split(input, ",")
		for _, part := range parts {
			name := strings.TrimSpace(part)
			if name != "" {
				result = append(result, name)
			}
		}
		return result
	}

	// Otherwise, use space separation
	parts := strings.Fields(input)
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name != "" {
			result = append(result, name)
		}
	}

	return result
}
