package parsers

import (
	"go/ast"
)

// ParseContext defines the context where parsing is happening
type ParseContext string

const (
	ContextMeta      ParseContext = "meta"      // swagger:meta
	ContextRoute     ParseContext = "route"     // swagger:route
	ContextModel     ParseContext = "model"     // swagger:model
	ContextField     ParseContext = "field"     // Field comments
	ContextParameter ParseContext = "parameter" // swagger:parameters
)

// TagParser is the base interface for all tag/directive parsers
type TagParser interface {
	// Name returns the tag name
	Name() string

	// Contexts returns the contexts where this parser is applicable
	Contexts() []ParseContext

	// Matches checks if this parser can handle the comment
	Matches(comment string, ctx ParseContext) bool

	// Parse extracts the value from the comment
	Parse(comments *ast.CommentGroup, ctx ParseContext) (any, error)

	// Apply applies the value to the target according to the context
	Apply(target any, value any, ctx ParseContext) error
}

// SetterFunc is a function that applies a value to a target
// Inspired by go-swagger but with support for multiple contexts
type SetterFunc func(target any, value any) error

// SetterMap maps contexts to setters
// Allows a parser to have different behaviors depending on the context
type SetterMap map[ParseContext]SetterFunc

// ParserType indicates the type of parser
type ParserType int

const (
	ParserTypeSingleLine ParserType = iota
	ParserTypeMultiLine
	ParserTypeYAML
	ParserTypeJSON
	ParserTypeList
)

// BaseParser contains common functionality for all parsers
type BaseParser struct {
	name       string
	parserType ParserType
	contexts   []ParseContext
	setters    SetterMap
}

// NewBaseParser creates a new BaseParser
func NewBaseParser(name string, parserType ParserType, contexts []ParseContext, setters SetterMap) BaseParser {
	return BaseParser{
		name:       name,
		parserType: parserType,
		contexts:   contexts,
		setters:    setters,
	}
}

// Name returns the parser name
func (p *BaseParser) Name() string {
	return p.name
}

// Contexts returns the supported contexts
func (p *BaseParser) Contexts() []ParseContext {
	return p.contexts
}

// SupportsContext checks if the parser supports a context
func (p *BaseParser) SupportsContext(ctx ParseContext) bool {
	for _, c := range p.contexts {
		if c == ctx {
			return true
		}
	}
	return false
}

// GetSetter returns the setter for a specific context
func (p *BaseParser) GetSetter(ctx ParseContext) (SetterFunc, bool) {
	setter, ok := p.setters[ctx]
	return setter, ok
}

// ApplyWithSetter applies a value using the context's setter
func (p *BaseParser) ApplyWithSetter(target any, value any, ctx ParseContext) error {
	setter, ok := p.GetSetter(ctx)
	if !ok {
		return &ErrNoSetterForContext{
			ParserName: p.name,
			Context:    ctx,
		}
	}
	return setter(target, value)
}
