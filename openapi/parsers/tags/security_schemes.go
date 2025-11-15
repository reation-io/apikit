package tags

import (
	"encoding/json"

	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/parsers/base"
	"github.com/reation-io/apikit/openapi/spec"
)

// NewSecuritySchemesParser creates a SecuritySchemes parser for swagger:meta
// Parses YAML content like:
// SecuritySchemes:
//   bearer:
//     type: http
//     scheme: bearer
//     bearerFormat: JWT
func NewSecuritySchemesParser() parsers.TagParser {
	return base.NewYAMLParser(
		"SecuritySchemes",
		parsers.RxSecuritySchemes,
		[]parsers.ParseContext{
			parsers.ContextMeta,
		},
		parsers.SetterMap{
			parsers.ContextMeta: func(target any, value any) error {
				openapi, ok := target.(*spec.OpenAPI)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "SecuritySchemes",
						Context:      parsers.ContextMeta,
						ExpectedType: "*spec.OpenAPI",
						ActualType:   getTypeName(target),
					}
				}

				// Value is json.RawMessage from YAMLParser
				rawMsg, ok := value.(json.RawMessage)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "SecuritySchemes",
						ExpectedType: "json.RawMessage",
						ActualType:   getTypeName(value),
					}
				}

				// Parse into map of SecurityScheme
				var schemes map[string]*spec.SecurityScheme
				if err := json.Unmarshal(rawMsg, &schemes); err != nil {
					return &parsers.ErrParseFailure{
						ParserName: "SecuritySchemes",
						Context:    parsers.ContextMeta,
						Cause:      err,
					}
				}

				// Initialize Components if needed
				if openapi.Components == nil {
					openapi.Components = &spec.Components{}
				}
				if openapi.Components.SecuritySchemes == nil {
					openapi.Components.SecuritySchemes = make(map[string]*spec.SecurityScheme)
				}

				// Merge security schemes
				for name, scheme := range schemes {
					openapi.Components.SecuritySchemes[name] = scheme
				}

				return nil
			},
		},
	)
}

func init() {
	parsers.Register("swagger:meta", NewSecuritySchemesParser())
}

