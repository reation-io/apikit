package tags

import (
	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/parsers/base"
	"github.com/reation-io/apikit/openapi/spec"
)

// NewReadOnlyParser creates a ReadOnly parser for field comments
func NewReadOnlyParser() parsers.TagParser {
	return base.NewSingleLineParser(
		"ReadOnly",
		parsers.RxReadOnly,
		[]parsers.ParseContext{parsers.ContextField},
		parsers.SetterMap{
			parsers.ContextField: func(target any, value any) error {
				schema, ok := target.(*spec.Schema)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "ReadOnly",
						Context:      parsers.ContextField,
						ExpectedType: "*spec.Schema",
						ActualType:   getTypeName(target),
					}
				}
				valStr, ok := value.(string)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "ReadOnly",
						ExpectedType: "string",
						ActualType:   getTypeName(value),
					}
				}
				schema.ReadOnly = valStr == "true" || valStr == "yes"
				return nil
			},
		},
	)
}

// NewWriteOnlyParser creates a WriteOnly parser for field comments
func NewWriteOnlyParser() parsers.TagParser {
	return base.NewSingleLineParser(
		"WriteOnly",
		parsers.RxWriteOnly,
		[]parsers.ParseContext{parsers.ContextField},
		parsers.SetterMap{
			parsers.ContextField: func(target any, value any) error {
				schema, ok := target.(*spec.Schema)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "WriteOnly",
						Context:      parsers.ContextField,
						ExpectedType: "*spec.Schema",
						ActualType:   getTypeName(target),
					}
				}
				valStr, ok := value.(string)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "WriteOnly",
						ExpectedType: "string",
						ActualType:   getTypeName(value),
					}
				}
				schema.WriteOnly = valStr == "true" || valStr == "yes"
				return nil
			},
		},
	)
}

func init() {
	parsers.Register("swagger:model", NewReadOnlyParser())
	parsers.Register("swagger:model", NewWriteOnlyParser())
}
