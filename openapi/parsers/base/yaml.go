package base

import (
	"encoding/json"
	"go/ast"
	"regexp"
	"strings"

	"github.com/reation-io/apikit/openapi/parsers"
	"gopkg.in/yaml.v3"
)

// YAMLParser is a reusable parser for YAML content
// Example:
// SecuritySchemes:
//
//	bearer:
//	  type: http
//	  scheme: bearer
type YAMLParser struct {
	parsers.BaseParser
	pattern *regexp.Regexp
}

// NewYAMLParser creates a new YAML parser
func NewYAMLParser(
	name string,
	pattern *regexp.Regexp,
	contexts []parsers.ParseContext,
	setters parsers.SetterMap,
) *YAMLParser {
	return &YAMLParser{
		BaseParser: parsers.NewBaseParser(name, parsers.ParserTypeYAML, contexts, setters),
		pattern:    pattern,
	}
}

// Matches checks if the comment matches the pattern
func (p *YAMLParser) Matches(comment string, ctx parsers.ParseContext) bool {
	// Check if the context is supported
	if !p.SupportsContext(ctx) {
		return false
	}

	return p.pattern.MatchString(comment)
}

// Parse extracts and parses the YAML content
func (p *YAMLParser) Parse(comments *ast.CommentGroup, ctx parsers.ParseContext) (any, error) {
	text := comments.Text()
	matches := p.pattern.FindStringSubmatch(text)

	if len(matches) < 2 {
		return nil, nil
	}

	// Extract YAML content
	yamlContent := matches[1]

	// Clean comment indentation while preserving relative indentation
	lines := strings.Split(yamlContent, "\n")
	cleaned := make([]string, 0, len(lines))

	// Find minimum indentation (excluding empty lines)
	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " \t"))
		if minIndent < 0 || indent < minIndent {
			minIndent = indent
		}
	}
	if minIndent < 0 {
		minIndent = 0
	}

	for _, line := range lines {
		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Remove common indentation
		if len(line) >= minIndent {
			line = line[minIndent:]
		}

		// Remove comment prefix if still present (after comments.Text())
		if strings.HasPrefix(strings.TrimSpace(line), "//") {
			// Strip leading whitespace, //, then recalculate indent
			trimmed := strings.TrimSpace(line)
			trimmed = strings.TrimPrefix(trimmed, "//")
			// Count original leading spaces after removing //
			if len(trimmed) > 0 && trimmed[0] == ' ' {
				trimmed = trimmed[1:] // Remove single space after //
			}
			line = trimmed
		}

		if strings.TrimSpace(line) != "" {
			cleaned = append(cleaned, line)
		}
	}

	yamlText := strings.Join(cleaned, "\n")

	// Parse YAML
	var yamlValue any
	if err := yaml.Unmarshal([]byte(yamlText), &yamlValue); err != nil {
		return nil, &parsers.ErrParseFailure{
			ParserName: p.Name(),
			Context:    ctx,
			Cause:      err,
		}
	}

	// Convert to JSON to facilitate later unmarshaling
	jsonBytes, err := json.Marshal(yamlValue)
	if err != nil {
		return nil, &parsers.ErrParseFailure{
			ParserName: p.Name(),
			Context:    ctx,
			Cause:      err,
		}
	}

	return json.RawMessage(jsonBytes), nil
}

// Apply applies the value to the target using the context's setter
func (p *YAMLParser) Apply(target any, value any, ctx parsers.ParseContext) error {
	return p.ApplyWithSetter(target, value, ctx)
}
