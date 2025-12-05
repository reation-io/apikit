package tags

import (
	"strconv"

	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/parsers/base"
	"github.com/reation-io/apikit/openapi/spec"
)

// NewMinimumParser creates a Minimum parser for field comments
func NewMinimumParser() parsers.TagParser {
	return base.NewSingleLineParser(
		"Minimum",
		parsers.RxMinimum,
		[]parsers.ParseContext{parsers.ContextField},
		parsers.SetterMap{
			parsers.ContextField: func(target any, value any) error {
				schema, ok := target.(*spec.Schema)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "Minimum",
						Context:      parsers.ContextField,
						ExpectedType: "*spec.Schema",
						ActualType:   getTypeName(target),
					}
				}
				minStr, ok := value.(string)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "Minimum",
						ExpectedType: "string",
						ActualType:   getTypeName(value),
					}
				}
				min, err := strconv.ParseFloat(minStr, 64)
				if err != nil {
					return &parsers.ErrParseFailure{
						ParserName: "Minimum",
						Context:    parsers.ContextField,
						Cause:      err,
					}
				}
				schema.Minimum = &min
				return nil
			},
		},
	)
}

// NewMaximumParser creates a Maximum parser for field comments
func NewMaximumParser() parsers.TagParser {
	return base.NewSingleLineParser(
		"Maximum",
		parsers.RxMaximum,
		[]parsers.ParseContext{parsers.ContextField},
		parsers.SetterMap{
			parsers.ContextField: func(target any, value any) error {
				schema, ok := target.(*spec.Schema)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "Maximum",
						Context:      parsers.ContextField,
						ExpectedType: "*spec.Schema",
						ActualType:   getTypeName(target),
					}
				}
				maxStr, ok := value.(string)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "Maximum",
						ExpectedType: "string",
						ActualType:   getTypeName(value),
					}
				}
				max, err := strconv.ParseFloat(maxStr, 64)
				if err != nil {
					return &parsers.ErrParseFailure{
						ParserName: "Maximum",
						Context:    parsers.ContextField,
						Cause:      err,
					}
				}
				schema.Maximum = &max
				return nil
			},
		},
	)
}

// NewMinLengthParser creates a MinLength parser for field comments
func NewMinLengthParser() parsers.TagParser {
	return base.NewSingleLineParser(
		"MinLength",
		parsers.RxMinLength,
		[]parsers.ParseContext{parsers.ContextField},
		parsers.SetterMap{
			parsers.ContextField: func(target any, value any) error {
				schema, ok := target.(*spec.Schema)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "MinLength",
						Context:      parsers.ContextField,
						ExpectedType: "*spec.Schema",
						ActualType:   getTypeName(target),
					}
				}
				minLenStr, ok := value.(string)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "MinLength",
						ExpectedType: "string",
						ActualType:   getTypeName(value),
					}
				}
				minLen, err := strconv.ParseInt(minLenStr, 10, 64)
				if err != nil {
					return &parsers.ErrParseFailure{
						ParserName: "MinLength",
						Context:    parsers.ContextField,
						Cause:      err,
					}
				}
				schema.MinLength = &minLen
				return nil
			},
		},
	)
}

// NewMaxLengthParser creates a MaxLength parser for field comments
func NewMaxLengthParser() parsers.TagParser {
	return base.NewSingleLineParser(
		"MaxLength",
		parsers.RxMaxLength,
		[]parsers.ParseContext{parsers.ContextField},
		parsers.SetterMap{
			parsers.ContextField: func(target any, value any) error {
				schema, ok := target.(*spec.Schema)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "MaxLength",
						Context:      parsers.ContextField,
						ExpectedType: "*spec.Schema",
						ActualType:   getTypeName(target),
					}
				}
				maxLenStr, ok := value.(string)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "MaxLength",
						ExpectedType: "string",
						ActualType:   getTypeName(value),
					}
				}
				maxLen, err := strconv.ParseInt(maxLenStr, 10, 64)
				if err != nil {
					return &parsers.ErrParseFailure{
						ParserName: "MaxLength",
						Context:    parsers.ContextField,
						Cause:      err,
					}
				}
				schema.MaxLength = &maxLen
				return nil
			},
		},
	)
}

// NewPatternParser creates a Pattern parser for field comments
func NewPatternParser() parsers.TagParser {
	return base.NewSingleLineParser(
		"Pattern",
		parsers.RxPattern,
		[]parsers.ParseContext{parsers.ContextField},
		parsers.SetterMap{
			parsers.ContextField: func(target any, value any) error {
				schema, ok := target.(*spec.Schema)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "Pattern",
						Context:      parsers.ContextField,
						ExpectedType: "*spec.Schema",
						ActualType:   getTypeName(target),
					}
				}
				pattern, ok := value.(string)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "Pattern",
						ExpectedType: "string",
						ActualType:   getTypeName(value),
					}
				}
				schema.Pattern = pattern
				return nil
			},
		},
	)
}

// NewMultipleOfParser creates a MultipleOf parser for field comments
func NewMultipleOfParser() parsers.TagParser {
	return base.NewSingleLineParser(
		"MultipleOf",
		parsers.RxMultipleOf,
		[]parsers.ParseContext{parsers.ContextField},
		parsers.SetterMap{
			parsers.ContextField: func(target any, value any) error {
				schema, ok := target.(*spec.Schema)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "MultipleOf",
						Context:      parsers.ContextField,
						ExpectedType: "*spec.Schema",
						ActualType:   getTypeName(target),
					}
				}
				multStr, ok := value.(string)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "MultipleOf",
						ExpectedType: "string",
						ActualType:   getTypeName(value),
					}
				}
				mult, err := strconv.ParseFloat(multStr, 64)
				if err != nil {
					return &parsers.ErrParseFailure{
						ParserName: "MultipleOf",
						Context:    parsers.ContextField,
						Cause:      err,
					}
				}
				schema.MultipleOf = &mult
				return nil
			},
		},
	)
}

// NewMinItemsParser creates a MinItems parser for field comments
func NewMinItemsParser() parsers.TagParser {
	return base.NewSingleLineParser(
		"MinItems",
		parsers.RxMinItems,
		[]parsers.ParseContext{parsers.ContextField},
		parsers.SetterMap{
			parsers.ContextField: func(target any, value any) error {
				schema, ok := target.(*spec.Schema)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "MinItems",
						Context:      parsers.ContextField,
						ExpectedType: "*spec.Schema",
						ActualType:   getTypeName(target),
					}
				}
				minItemsStr, ok := value.(string)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "MinItems",
						ExpectedType: "string",
						ActualType:   getTypeName(value),
					}
				}
				minItems, err := strconv.ParseInt(minItemsStr, 10, 64)
				if err != nil {
					return &parsers.ErrParseFailure{
						ParserName: "MinItems",
						Context:    parsers.ContextField,
						Cause:      err,
					}
				}
				schema.MinItems = &minItems
				return nil
			},
		},
	)
}

// NewMaxItemsParser creates a MaxItems parser for field comments
func NewMaxItemsParser() parsers.TagParser {
	return base.NewSingleLineParser(
		"MaxItems",
		parsers.RxMaxItems,
		[]parsers.ParseContext{parsers.ContextField},
		parsers.SetterMap{
			parsers.ContextField: func(target any, value any) error {
				schema, ok := target.(*spec.Schema)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "MaxItems",
						Context:      parsers.ContextField,
						ExpectedType: "*spec.Schema",
						ActualType:   getTypeName(target),
					}
				}
				maxItemsStr, ok := value.(string)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "MaxItems",
						ExpectedType: "string",
						ActualType:   getTypeName(value),
					}
				}
				maxItems, err := strconv.ParseInt(maxItemsStr, 10, 64)
				if err != nil {
					return &parsers.ErrParseFailure{
						ParserName: "MaxItems",
						Context:    parsers.ContextField,
						Cause:      err,
					}
				}
				schema.MaxItems = &maxItems
				return nil
			},
		},
	)
}

// NewUniqueItemsParser creates a UniqueItems parser for field comments
func NewUniqueItemsParser() parsers.TagParser {
	return base.NewSingleLineParser(
		"UniqueItems",
		parsers.RxUniqueItems,
		[]parsers.ParseContext{parsers.ContextField},
		parsers.SetterMap{
			parsers.ContextField: func(target any, value any) error {
				schema, ok := target.(*spec.Schema)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "UniqueItems",
						Context:      parsers.ContextField,
						ExpectedType: "*spec.Schema",
						ActualType:   getTypeName(target),
					}
				}
				valStr, ok := value.(string)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "UniqueItems",
						ExpectedType: "string",
						ActualType:   getTypeName(value),
					}
				}
				schema.UniqueItems = valStr == "true" || valStr == "yes"
				return nil
			},
		},
	)
}

func init() {
	parsers.Register("swagger:model", NewMinimumParser())
	parsers.Register("swagger:model", NewMaximumParser())
	parsers.Register("swagger:model", NewMinLengthParser())
	parsers.Register("swagger:model", NewMaxLengthParser())
	parsers.Register("swagger:model", NewMinItemsParser())
	parsers.Register("swagger:model", NewMaxItemsParser())
	parsers.Register("swagger:model", NewUniqueItemsParser())
	parsers.Register("swagger:model", NewMultipleOfParser())
	parsers.Register("swagger:model", NewPatternParser())
}
