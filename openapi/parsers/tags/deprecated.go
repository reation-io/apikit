package tags

import (
	"strings"

	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/parsers/base"
	"github.com/reation-io/apikit/openapi/spec"
)

// NewDeprecatedParser creates a Deprecated parser
// Works in: route (Operation.Deprecated), field (Schema.Deprecated)
func NewDeprecatedParser() parsers.TagParser {
	return base.NewSingleLineParser(
		"Deprecated",
		parsers.RxDeprecated,
		[]parsers.ParseContext{
			parsers.ContextRoute,
			parsers.ContextField,
		},
		parsers.SetterMap{
			parsers.ContextRoute: func(target any, value any) error {
				operation, ok := target.(*spec.Operation)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "Deprecated",
						Context:      parsers.ContextRoute,
						ExpectedType: "*spec.Operation",
						ActualType:   getTypeName(target),
					}
				}
				deprecatedStr, ok := value.(string)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "Deprecated",
						ExpectedType: "string",
						ActualType:   getTypeName(value),
					}
				}
				operation.Deprecated = parseBool(deprecatedStr)
				return nil
			},
			parsers.ContextField: func(target any, value any) error {
				schema, ok := target.(*spec.Schema)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "Deprecated",
						Context:      parsers.ContextField,
						ExpectedType: "*spec.Schema",
						ActualType:   getTypeName(target),
					}
				}
				deprecatedStr, ok := value.(string)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "Deprecated",
						ExpectedType: "string",
						ActualType:   getTypeName(value),
					}
				}
				schema.Deprecated = parseBool(deprecatedStr)
				return nil
			},
		},
	)
}

// parseBool parses a boolean string (true, false, yes, no)
func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "yes"
}

func init() {
	parsers.Register("swagger:route", NewDeprecatedParser())
	parsers.Register("swagger:model", NewDeprecatedParser())
}

