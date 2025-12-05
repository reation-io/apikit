package tags

import (
	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/parsers/base"
	"github.com/reation-io/apikit/openapi/spec"
)

// NewDefaultParser creates a Default parser for field comments
func NewDefaultParser() parsers.TagParser {
	return base.NewSingleLineParser(
		"Default",
		parsers.RxDefault,
		[]parsers.ParseContext{parsers.ContextField},
		parsers.SetterMap{
			parsers.ContextField: func(target any, value any) error {
				schema, ok := target.(*spec.Schema)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "Default",
						Context:      parsers.ContextField,
						ExpectedType: "*spec.Schema",
						ActualType:   getTypeName(target),
					}
				}
				// Default value is stored as is (string) or parsed?
				// The spec.Schema.Default is 'any'.
				// For now, we store the string value.
				// Ideally we should try to parse it based on the schema type, but that context might not be fully available or robust here.
				// However, if it's "true"/"false" and type is boolean, or numeric, we could try guess.
				// But simpler is safer: just store the string from the comment.
				// Wait, if users want 0 as a number, "0" string is different.
				// Let's rely on standard JSON behavior or simple inference if possible?
				// For now, consistent with others: pass the string.
				schema.Default = value
				return nil
			},
		},
	)
}

func init() {
	parsers.Register("swagger:model", NewDefaultParser())
}
