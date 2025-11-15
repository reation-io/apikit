package tags

import (
	"encoding/json"
	"strconv"

	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/parsers/base"
	"github.com/reation-io/apikit/openapi/spec"
)

// NewExampleParser creates an Example parser for field comments
func NewExampleParser() parsers.TagParser {
	return base.NewSingleLineParser(
		"Example",
		parsers.RxExample,
		[]parsers.ParseContext{
			parsers.ContextField,
		},
		parsers.SetterMap{
			parsers.ContextField: func(target any, value any) error {
				schema, ok := target.(*spec.Schema)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "Example",
						Context:      parsers.ContextField,
						ExpectedType: "*spec.Schema",
						ActualType:   getTypeName(target),
					}
				}
				exampleStr, ok := value.(string)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "Example",
						ExpectedType: "string",
						ActualType:   getTypeName(value),
					}
				}

				// Try to parse as JSON first
				var jsonValue any
				if err := json.Unmarshal([]byte(exampleStr), &jsonValue); err == nil {
					schema.Example = jsonValue
					return nil
				}

				// Try to parse as number
				if num, err := strconv.ParseFloat(exampleStr, 64); err == nil {
					schema.Example = num
					return nil
				}

				// Try to parse as boolean
				if b, err := strconv.ParseBool(exampleStr); err == nil {
					schema.Example = b
					return nil
				}

				// Use as string
				schema.Example = exampleStr
				return nil
			},
		},
	)
}

func init() {
	parsers.Register("swagger:model", NewExampleParser())
}

