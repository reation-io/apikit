package tags

import (
	"fmt"
	"go/ast"
	"regexp"
	"strings"

	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/spec"
)

// NullableParser parses nullable directives from model field comments
// Format:
// swagger:model User
//
//	type User struct {
//	    // nullable: true
//	    MiddleName *string `json:"middle_name"`
//	}
type NullableParser struct {
	parsers.BaseParser
}

func init() {
	parsers.GlobalRegistry().Register("swagger:model", &NullableParser{
		BaseParser: parsers.NewBaseParser(
			"nullable",
			parsers.ParserTypeMultiLine,
			[]parsers.ParseContext{parsers.ContextModel, parsers.ContextField},
			nil,
		),
	})
}

// Pattern to match nullable directive
var nullablePattern = regexp.MustCompile(`(?mi)^nullable:\s*(true|false|yes|no)$`)

// Matches checks if the comment contains nullable directive
func (p *NullableParser) Matches(comment string, ctx parsers.ParseContext) bool {
	if ctx != parsers.ContextModel && ctx != parsers.ContextField {
		return false
	}
	return strings.Contains(strings.ToLower(comment), "nullable:")
}

// Parse extracts nullable value from the comment
func (p *NullableParser) Parse(comments *ast.CommentGroup, ctx parsers.ParseContext) (any, error) {
	if ctx != parsers.ContextModel && ctx != parsers.ContextField {
		return nil, nil
	}

	text := comments.Text()
	matches := nullablePattern.FindStringSubmatch(text)
	if len(matches) < 2 {
		return nil, nil
	}

	value := strings.ToLower(strings.TrimSpace(matches[1]))
	return value == "true" || value == "yes", nil
}

// Apply applies the parsed nullable value to the Schema
func (p *NullableParser) Apply(target any, value any, ctx parsers.ParseContext) error {
	if ctx != parsers.ContextModel && ctx != parsers.ContextField {
		return nil
	}

	schema, ok := target.(*spec.Schema)
	if !ok {
		return &parsers.ErrInvalidTarget{
			ParserName:   "nullable",
			Context:      ctx,
			ExpectedType: "*spec.Schema",
			ActualType:   fmt.Sprintf("%T", target),
		}
	}

	nullable, ok := value.(bool)
	if !ok {
		return nil
	}

	schema.Nullable = nullable
	return nil
}
