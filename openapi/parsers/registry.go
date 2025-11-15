package parsers

import (
	"fmt"
	"go/ast"
	"sync"
)

// ParserRegistry maintains a registry of all available parsers
type ParserRegistry struct {
	mu      sync.RWMutex
	parsers map[string][]TagParser // key: directive (swagger:meta, swagger:route, etc.)
}

var globalRegistry = &ParserRegistry{
	parsers: make(map[string][]TagParser),
}

// GlobalRegistry returns the global parser registry
func GlobalRegistry() *ParserRegistry {
	return globalRegistry
}

// Register registers a parser for a specific directive in the global registry
func Register(directive string, parser TagParser) {
	globalRegistry.Register(directive, parser)
}

// Register registers a parser for a specific directive
func (r *ParserRegistry) Register(directive string, parser TagParser) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.parsers[directive] = append(r.parsers[directive], parser)
}

// GetParsers returns all parsers for a directive
func (r *ParserRegistry) GetParsers(directive string) []TagParser {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.parsers[directive]
}

// Parse executes all parsers for a directive with a specific context
func (r *ParserRegistry) Parse(
	directive string,
	comments *ast.CommentGroup,
	target any,
	ctx ParseContext,
) error {
	if comments == nil {
		return nil
	}

	parsers := r.GetParsers(directive)
	commentText := comments.Text()

	for _, parser := range parsers {
		// Check if the parser supports this context and matches
		if !parser.Matches(commentText, ctx) {
			continue
		}

		// Parse the value
		value, err := parser.Parse(comments, ctx)
		if err != nil {
			return &ErrParseFailure{
				ParserName: parser.Name(),
				Context:    ctx,
				Cause:      err,
			}
		}

		// Apply the value to the target
		if err := parser.Apply(target, value, ctx); err != nil {
			// Ignore invalid target errors - this allows calling Parse with different targets
			// and only the parsers that match the target type will apply
			if _, ok := err.(*ErrInvalidTarget); ok {
				continue
			}
			return fmt.Errorf("applying %s failed: %w", parser.Name(), err)
		}
	}

	return nil
}

// ParseAll executes all registered parsers for a directive
// Useful when you don't know which context to use
func (r *ParserRegistry) ParseAll(
	directive string,
	comments *ast.CommentGroup,
	targets map[ParseContext]any,
) error {
	if comments == nil {
		return nil
	}

	for ctx, target := range targets {
		if err := r.Parse(directive, comments, target, ctx); err != nil {
			return err
		}
	}

	return nil
}

// ListParsers returns a list of all registered parsers
func (r *ParserRegistry) ListParsers() map[string][]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string][]string)
	for directive, parsers := range r.parsers {
		names := make([]string, len(parsers))
		for i, p := range parsers {
			names[i] = p.Name()
		}
		result[directive] = names
	}

	return result
}

// Clear clears all registered parsers (useful for testing)
func (r *ParserRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.parsers = make(map[string][]TagParser)
}
