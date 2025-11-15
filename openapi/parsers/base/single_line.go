package base

import (
	"go/ast"
	"regexp"
	"strings"

	"github.com/reation-io/apikit/openapi/parsers"
)

// SingleLineParser is a reusable parser for single-line tags
// Example: "Version: 1.0.0"
type SingleLineParser struct {
	parsers.BaseParser
	pattern *regexp.Regexp
}

// NewSingleLineParser creates a new single-line parser
func NewSingleLineParser(
	name string,
	pattern *regexp.Regexp,
	contexts []parsers.ParseContext,
	setters parsers.SetterMap,
) *SingleLineParser {
	return &SingleLineParser{
		BaseParser: parsers.NewBaseParser(name, parsers.ParserTypeSingleLine, contexts, setters),
		pattern:    pattern,
	}
}

// Matches checks if the comment matches the pattern
func (p *SingleLineParser) Matches(comment string, ctx parsers.ParseContext) bool {
	// Check if the context is supported
	if !p.SupportsContext(ctx) {
		return false
	}

	return p.pattern.MatchString(comment)
}

// Parse extracts the value from the comment
func (p *SingleLineParser) Parse(comments *ast.CommentGroup, ctx parsers.ParseContext) (any, error) {
	text := comments.Text()
	matches := p.pattern.FindStringSubmatch(text)

	if len(matches) < 2 {
		return "", nil
	}

	return strings.TrimSpace(matches[1]), nil
}

// Apply applies the value to the target using the context's setter
func (p *SingleLineParser) Apply(target any, value any, ctx parsers.ParseContext) error {
	return p.ApplyWithSetter(target, value, ctx)
}
