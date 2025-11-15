package tags

import (
	"strings"

	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/parsers/base"
	"github.com/reation-io/apikit/openapi/spec"
)

// NewTagsParser creates a Tags parser for swagger:route
// Parses comma-separated tags: "Tags: user, admin, api"
func NewTagsParser() parsers.TagParser {
	return base.NewSingleLineParser(
		"Tags",
		parsers.RxTags,
		[]parsers.ParseContext{
			parsers.ContextRoute,
		},
		parsers.SetterMap{
			parsers.ContextRoute: func(target any, value any) error {
				operation, ok := target.(*spec.Operation)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "Tags",
						Context:      parsers.ContextRoute,
						ExpectedType: "*spec.Operation",
						ActualType:   getTypeName(target),
					}
				}
				tagsStr, ok := value.(string)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "Tags",
						ExpectedType: "string",
						ActualType:   getTypeName(value),
					}
				}

				// Parse comma-separated tags
				tagsList := strings.Split(tagsStr, ",")
				tags := make([]string, 0, len(tagsList))
				for _, tag := range tagsList {
					trimmed := strings.TrimSpace(tag)
					if trimmed != "" {
						tags = append(tags, trimmed)
					}
				}

				operation.Tags = tags
				return nil
			},
		},
	)
}

func init() {
	parsers.Register("swagger:route", NewTagsParser())
}

