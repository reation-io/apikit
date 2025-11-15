package tags

import (
	"regexp"
	"strings"

	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/parsers/base"
	"github.com/reation-io/apikit/openapi/spec"
)

var (
	// rxSpec matches "Spec:" followed by space/comma-separated spec names
	rxSpec = regexp.MustCompile(`(?i)Spec\s*:\s*(.+)`)

	// rxSpecName validates spec names (alphanumeric, hyphens, underscores)
	rxSpecName = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// NewSpecParser creates a Spec parser
// Works in: meta (defines which specs the meta applies to), route (defines which specs the route belongs to)
// Parses space or comma-separated spec names: "Spec: admin mobile public" or "Spec: admin, mobile, public"
// Spec names are normalized to lowercase
func NewSpecParser() parsers.TagParser {
	return base.NewSingleLineParser(
		"Spec",
		rxSpec,
		[]parsers.ParseContext{
			parsers.ContextMeta,
			parsers.ContextRoute,
		},
		parsers.SetterMap{
			parsers.ContextMeta: func(target any, value any) error {
				info, ok := target.(*spec.Info)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "Spec",
						Context:      parsers.ContextMeta,
						ExpectedType: "*spec.Info",
						ActualType:   getTypeName(target),
					}
				}
				specsStr, ok := value.(string)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "Spec",
						ExpectedType: "string",
						ActualType:   getTypeName(value),
					}
				}

				specs := parseSpecNames(specsStr)
				if len(specs) == 0 {
					return nil
				}

				// Store in Info.Extensions
				if info.Extensions == nil {
					info.Extensions = make(map[string]any)
				}
				info.Extensions["x-specs"] = specs
				return nil
			},
			parsers.ContextRoute: func(target any, value any) error {
				operation, ok := target.(*spec.Operation)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "Spec",
						Context:      parsers.ContextRoute,
						ExpectedType: "*spec.Operation",
						ActualType:   getTypeName(target),
					}
				}
				specsStr, ok := value.(string)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "Spec",
						ExpectedType: "string",
						ActualType:   getTypeName(value),
					}
				}

				specs := parseSpecNames(specsStr)
				if len(specs) == 0 {
					return nil
				}

				// Store in Operation.Extensions
				if operation.Extensions == nil {
					operation.Extensions = make(map[string]any)
				}
				operation.Extensions["x-specs"] = specs
				return nil
			},
		},
	)
}

// parseSpecNames parses a space or comma-separated list of spec names
// Normalizes to lowercase and validates each name
func parseSpecNames(input string) []string {
	// Split by both spaces and commas
	input = strings.ReplaceAll(input, ",", " ")
	parts := strings.Fields(input)

	var specs []string
	for _, part := range parts {
		// Normalize to lowercase
		normalized := strings.ToLower(strings.TrimSpace(part))

		// Validate spec name
		if !rxSpecName.MatchString(normalized) {
			// Skip invalid names silently (could log warning in future)
			continue
		}

		specs = append(specs, normalized)
	}

	return specs
}

func init() {
	parsers.Register("swagger:meta", NewSpecParser())
	parsers.Register("swagger:route", NewSpecParser())
}
