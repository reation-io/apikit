package tags

import (
	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/parsers/base"
	"github.com/reation-io/apikit/openapi/spec"
)

// NewFormatParser creates a Format parser for field comments
// Common formats: date-time, email, hostname, ipv4, ipv6, uri, uuid, etc.
func NewFormatParser() parsers.TagParser {
	return base.NewSingleLineParser(
		"Format",
		parsers.RxFormat,
		[]parsers.ParseContext{
			parsers.ContextField,
		},
		parsers.SetterMap{
			parsers.ContextField: func(target any, value any) error {
				schema, ok := target.(*spec.Schema)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "Format",
						Context:      parsers.ContextField,
						ExpectedType: "*spec.Schema",
						ActualType:   getTypeName(target),
					}
				}
				format, ok := value.(string)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "Format",
						ExpectedType: "string",
						ActualType:   getTypeName(value),
					}
				}
				schema.Format = format
				return nil
			},
		},
	)
}

func init() {
	parsers.Register("swagger:model", NewFormatParser())
}

