package tags

import (
	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/parsers/base"
	"github.com/reation-io/apikit/openapi/spec"
)

// NewTitleParser creates a Title parser
// Works in: meta (Info.Title), route (Operation.Summary if Summary not set)
func NewTitleParser() parsers.TagParser {
	return base.NewSingleLineParser(
		"Title",
		parsers.RxTitle,
		[]parsers.ParseContext{
			parsers.ContextMeta,
			parsers.ContextRoute,
		},
		parsers.SetterMap{
			parsers.ContextMeta: func(target any, value any) error {
				info, ok := target.(*spec.Info)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "Title",
						Context:      parsers.ContextMeta,
						ExpectedType: "*spec.Info",
						ActualType:   getTypeName(target),
					}
				}
				title, ok := value.(string)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "Title",
						ExpectedType: "string",
						ActualType:   getTypeName(value),
					}
				}
				info.Title = title
				return nil
			},
			parsers.ContextRoute: func(target any, value any) error {
				operation, ok := target.(*spec.Operation)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "Title",
						Context:      parsers.ContextRoute,
						ExpectedType: "*spec.Operation",
						ActualType:   getTypeName(target),
					}
				}
				title, ok := value.(string)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "Title",
						ExpectedType: "string",
						ActualType:   getTypeName(value),
					}
				}
				// Use title as summary if summary is not set
				if operation.Summary == "" {
					operation.Summary = title
				}
				return nil
			},
		},
	)
}

func init() {
	parsers.Register("swagger:meta", NewTitleParser())
	parsers.Register("swagger:route", NewTitleParser())
}

