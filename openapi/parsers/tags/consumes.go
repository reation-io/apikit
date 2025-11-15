package tags

import (
	"strings"

	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/parsers/base"
	"github.com/reation-io/apikit/openapi/spec"
)

// NewConsumesParser creates a Consumes parser
// Works in: meta (global), route (operation-specific)
// Parses comma-separated MIME types: "Consumes: application/json, application/xml"
func NewConsumesParser() parsers.TagParser {
	return base.NewSingleLineParser(
		"Consumes",
		parsers.RxConsumes,
		[]parsers.ParseContext{
			parsers.ContextMeta,
			parsers.ContextRoute,
		},
		parsers.SetterMap{
			parsers.ContextMeta: func(target any, value any) error {
				openapi, ok := target.(*spec.OpenAPI)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "Consumes",
						Context:      parsers.ContextMeta,
						ExpectedType: "*spec.OpenAPI",
						ActualType:   getTypeName(target),
					}
				}
				consumesStr, ok := value.(string)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "Consumes",
						ExpectedType: "string",
						ActualType:   getTypeName(value),
					}
				}

				// Parse comma-separated MIME types
				mimeTypes := parseMimeTypes(consumesStr)
				
				// Store in extensions for now (OpenAPI 3.0 doesn't have global consumes)
				// This can be used to set default requestBody content types
				if openapi.Extensions == nil {
					openapi.Extensions = make(map[string]any)
				}
				openapi.Extensions["x-consumes"] = mimeTypes
				return nil
			},
			parsers.ContextRoute: func(target any, value any) error {
				operation, ok := target.(*spec.Operation)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "Consumes",
						Context:      parsers.ContextRoute,
						ExpectedType: "*spec.Operation",
						ActualType:   getTypeName(target),
					}
				}
				consumesStr, ok := value.(string)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "Consumes",
						ExpectedType: "string",
						ActualType:   getTypeName(value),
					}
				}

				// Parse comma-separated MIME types
				mimeTypes := parseMimeTypes(consumesStr)

				// Create RequestBody if it doesn't exist
				if operation.RequestBody == nil {
					operation.RequestBody = &spec.RequestBody{
						Content: make(map[string]*spec.MediaType),
					}
				}

				// Add content types to RequestBody
				for _, mimeType := range mimeTypes {
					if operation.RequestBody.Content[mimeType] == nil {
						operation.RequestBody.Content[mimeType] = &spec.MediaType{}
					}
				}

				return nil
			},
		},
	)
}

// parseMimeTypes parses comma-separated MIME types
func parseMimeTypes(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func init() {
	parsers.Register("swagger:meta", NewConsumesParser())
	parsers.Register("swagger:route", NewConsumesParser())
}

