package tags

import (
	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/parsers/base"
	"github.com/reation-io/apikit/openapi/spec"
)

// NewVersionParser creates a Version parser for swagger:meta
func NewVersionParser() parsers.TagParser {
	return base.NewSingleLineParser(
		"Version",
		parsers.RxVersion,
		[]parsers.ParseContext{
			parsers.ContextMeta,
		},
		parsers.SetterMap{
			parsers.ContextMeta: func(target any, value any) error {
				info, ok := target.(*spec.Info)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "Version",
						Context:      parsers.ContextMeta,
						ExpectedType: "*spec.Info",
						ActualType:   getTypeName(target),
					}
				}
				version, ok := value.(string)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "Version",
						ExpectedType: "string",
						ActualType:   getTypeName(value),
					}
				}
				info.Version = version
				return nil
			},
		},
	)
}

func init() {
	parsers.Register("swagger:meta", NewVersionParser())
}

