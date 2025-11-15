package tags

import (
	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/parsers/base"
	"github.com/reation-io/apikit/openapi/spec"
)

// NewSummaryParser creates a Summary parser for swagger:route
func NewSummaryParser() parsers.TagParser {
	return base.NewSingleLineParser(
		"Summary",
		parsers.RxSummary,
		[]parsers.ParseContext{
			parsers.ContextRoute,
		},
		parsers.SetterMap{
			parsers.ContextRoute: func(target any, value any) error {
				operation, ok := target.(*spec.Operation)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "Summary",
						Context:      parsers.ContextRoute,
						ExpectedType: "*spec.Operation",
						ActualType:   getTypeName(target),
					}
				}
				summary, ok := value.(string)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "Summary",
						ExpectedType: "string",
						ActualType:   getTypeName(value),
					}
				}
				operation.Summary = summary
				return nil
			},
		},
	)
}

func init() {
	parsers.Register("swagger:route", NewSummaryParser())
}

