package base

import (
	"go/ast"
	"regexp"
	"strings"

	"github.com/reation-io/apikit/openapi/parsers"
)

// MultiLineParser is a reusable parser for multi-line tags
// Example:
// Description:
//
//	This is a multi-line
//	description
type MultiLineParser struct {
	parsers.BaseParser
	pattern   *regexp.Regexp
	dropEmpty bool // If true, removes empty lines
}

// NewMultiLineParser creates a new multi-line parser
func NewMultiLineParser(
	name string,
	pattern *regexp.Regexp,
	contexts []parsers.ParseContext,
	setters parsers.SetterMap,
	dropEmpty bool,
) *MultiLineParser {
	return &MultiLineParser{
		BaseParser: parsers.NewBaseParser(name, parsers.ParserTypeMultiLine, contexts, setters),
		pattern:    pattern,
		dropEmpty:  dropEmpty,
	}
}

// Matches checks if the comment matches the pattern
func (p *MultiLineParser) Matches(comment string, ctx parsers.ParseContext) bool {
	// Check if the context is supported
	if !p.SupportsContext(ctx) {
		return false
	}

	return p.pattern.MatchString(comment)
}

// Parse extracts the value from the multi-line comment
func (p *MultiLineParser) Parse(comments *ast.CommentGroup, ctx parsers.ParseContext) (any, error) {
	text := comments.Text()
	matches := p.pattern.FindStringSubmatch(text)

	if len(matches) < 2 {
		return "", nil
	}

	// Extract multi-line content
	content := matches[1]

	// Clean lines
	lines := strings.Split(content, "\n")
	cleaned := make([]string, 0, len(lines))

	for _, line := range lines {
		// Remove leading whitespace (indentation)
		trimmed := strings.TrimSpace(line)

		// If dropEmpty is true, skip empty lines
		if p.dropEmpty && trimmed == "" {
			continue
		}

		cleaned = append(cleaned, trimmed)
	}

	// Join lines
	result := strings.Join(cleaned, "\n")

	// Remove trailing newline if present
	result = strings.TrimSuffix(result, "\n")

	return result, nil
}

// Apply applies the value to the target using the context's setter
func (p *MultiLineParser) Apply(target any, value any, ctx parsers.ParseContext) error {
	return p.ApplyWithSetter(target, value, ctx)
}
