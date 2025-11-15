package tags

import (
	"regexp"

	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/parsers/base"
	"github.com/reation-io/apikit/openapi/spec"
)

var (
	// rxDescription matches "Description:" followed by content until next directive or end
	// Stops at lines starting with capital letter followed by colon (e.g., "Security:", "Responses:")
	rxDescription = regexp.MustCompile(`(?ims)[Dd]escription\s*:\s*(.*?)(?:^[A-Z][a-zA-Z]*:\s*$|\z)`)
)

// NewDescriptionParser creates a reusable Description parser
// This parser works in multiple contexts: meta, route, and field
func NewDescriptionParser() parsers.TagParser {
	return base.NewMultiLineParser(
		"Description",
		rxDescription,
		[]parsers.ParseContext{
			parsers.ContextMeta,
			parsers.ContextRoute,
			parsers.ContextField,
		},
		parsers.SetterMap{
			// For swagger:meta - sets Info.Description
			parsers.ContextMeta: func(target any, value any) error {
				info, ok := target.(*spec.Info)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "Description",
						Context:      parsers.ContextMeta,
						ExpectedType: "*spec.Info",
						ActualType:   getTypeName(target),
					}
				}
				desc, ok := value.(string)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "Description",
						ExpectedType: "string",
						ActualType:   getTypeName(value),
					}
				}
				info.Description = desc
				return nil
			},

			// For swagger:route - sets Operation.Description
			parsers.ContextRoute: func(target any, value any) error {
				operation, ok := target.(*spec.Operation)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "Description",
						Context:      parsers.ContextRoute,
						ExpectedType: "*spec.Operation",
						ActualType:   getTypeName(target),
					}
				}
				desc, ok := value.(string)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "Description",
						ExpectedType: "string",
						ActualType:   getTypeName(value),
					}
				}
				operation.Description = desc
				return nil
			},

			// For field comments - sets Schema.Description
			parsers.ContextField: func(target any, value any) error {
				schema, ok := target.(*spec.Schema)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "Description",
						Context:      parsers.ContextField,
						ExpectedType: "*spec.Schema",
						ActualType:   getTypeName(target),
					}
				}
				desc, ok := value.(string)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "Description",
						ExpectedType: "string",
						ActualType:   getTypeName(value),
					}
				}
				schema.Description = desc
				return nil
			},
		},
		false, // dropEmpty = false - preserve blank lines for paragraph separation
	)
}

// Auto-register the parser on package init
func init() {
	parser := NewDescriptionParser()
	parsers.Register("swagger:meta", parser)
	parsers.Register("swagger:route", parser)
	parsers.Register("swagger:model", parser)
}
