package tags

import (
	"go/ast"
	"regexp"
	"strings"

	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/spec"
)

// IgnoredParametersParser parses IgnoredParameters directives from route comments
// Format:
// swagger:route GET /path tags operationId
// IgnoredParameters: paramName1 paramName2
// or
// IgnoredParameters: paramName1, paramName2
type IgnoredParametersParser struct {
	parsers.BaseParser
}

func init() {
	parsers.GlobalRegistry().Register("swagger:route", &IgnoredParametersParser{
		BaseParser: parsers.NewBaseParser(
			"IgnoredParameters",
			parsers.ParserTypeMultiLine,
			[]parsers.ParseContext{parsers.ContextRoute},
			nil,
		),
	})
}

// Pattern to match IgnoredParameters directive
var ignoredParametersPattern = regexp.MustCompile(`(?mi)^IgnoredParameters:\s*(.+)$`)

// Matches checks if the comment contains IgnoredParameters directive
func (p *IgnoredParametersParser) Matches(comment string, ctx parsers.ParseContext) bool {
	if ctx != parsers.ContextRoute {
		return false
	}
	return strings.Contains(strings.ToLower(comment), "ignoredparameters:")
}

// Parse extracts ignored parameter names from the comment
func (p *IgnoredParametersParser) Parse(comments *ast.CommentGroup, ctx parsers.ParseContext) (any, error) {
	if ctx != parsers.ContextRoute {
		return nil, nil
	}

	text := comments.Text()
	matches := ignoredParametersPattern.FindStringSubmatch(text)
	if len(matches) < 2 {
		return nil, nil
	}

	// Parse the parameter names (comma or space separated)
	paramsStr := strings.TrimSpace(matches[1])
	return parseParameterList(paramsStr), nil
}

// Apply applies the parsed ignored parameters to the RouteInfo
func (p *IgnoredParametersParser) Apply(target any, value any, ctx parsers.ParseContext) error {
	if ctx != parsers.ContextRoute {
		return nil
	}

	routeInfo, ok := target.(*spec.RouteInfo)
	if !ok {
		return nil
	}

	params, ok := value.([]string)
	if !ok || len(params) == 0 {
		return nil
	}

	routeInfo.IgnoredParameters = append(routeInfo.IgnoredParameters, params...)
	return nil
}

// parseParameterList parses a comma or space separated list of parameter names
func parseParameterList(input string) []string {
	var result []string

	// First, try comma separation
	if strings.Contains(input, ",") {
		parts := strings.Split(input, ",")
		for _, part := range parts {
			name := strings.TrimSpace(part)
			if name != "" {
				result = append(result, name)
			}
		}
		return result
	}

	// Otherwise, use space separation
	parts := strings.Fields(input)
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name != "" {
			result = append(result, name)
		}
	}

	return result
}

