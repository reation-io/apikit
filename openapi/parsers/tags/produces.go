package tags

import (
	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/parsers/base"
	"github.com/reation-io/apikit/openapi/spec"
)

// NewProducesParser creates a Produces parser
// Works in: meta (global), route (operation-specific)
// Parses comma-separated MIME types: "Produces: application/json, application/xml"
func NewProducesParser() parsers.TagParser {
	return base.NewSingleLineParser(
		"Produces",
		parsers.RxProduces,
		[]parsers.ParseContext{
			parsers.ContextMeta,
			parsers.ContextRoute,
		},
		parsers.SetterMap{
			parsers.ContextMeta: func(target any, value any) error {
				openapi, ok := target.(*spec.OpenAPI)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "Produces",
						Context:      parsers.ContextMeta,
						ExpectedType: "*spec.OpenAPI",
						ActualType:   getTypeName(target),
					}
				}
				producesStr, ok := value.(string)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "Produces",
						ExpectedType: "string",
						ActualType:   getTypeName(value),
					}
				}

				// Parse comma-separated MIME types
				mimeTypes := parseMimeTypes(producesStr)
				
				// Store in extensions for now (OpenAPI 3.0 doesn't have global produces)
				// This can be used to set default response content types
				if openapi.Extensions == nil {
					openapi.Extensions = make(map[string]any)
				}
				openapi.Extensions["x-produces"] = mimeTypes
				return nil
			},
			parsers.ContextRoute: func(target any, value any) error {
				operation, ok := target.(*spec.Operation)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "Produces",
						Context:      parsers.ContextRoute,
						ExpectedType: "*spec.Operation",
						ActualType:   getTypeName(target),
					}
				}
				producesStr, ok := value.(string)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "Produces",
						ExpectedType: "string",
						ActualType:   getTypeName(value),
					}
				}

				// Parse comma-separated MIME types
				mimeTypes := parseMimeTypes(producesStr)

				// Store in extensions (will be used when creating responses)
				if operation.Extensions == nil {
					operation.Extensions = make(map[string]any)
				}
				operation.Extensions["x-produces"] = mimeTypes

				return nil
			},
		},
	)
}

func init() {
	parsers.Register("swagger:meta", NewProducesParser())
	parsers.Register("swagger:route", NewProducesParser())
}

