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

func init() {
	parsers.Register("swagger:model", NewMinimumParser())
	parsers.Register("swagger:model", NewMaximumParser())
	parsers.Register("swagger:model", NewMinLengthParser())
	parsers.Register("swagger:model", NewMaxLengthParser())
	parsers.Register("swagger:model", NewPatternParser())
}
