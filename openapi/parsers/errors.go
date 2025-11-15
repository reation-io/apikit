package parsers

import "fmt"

// ErrNoSetterForContext indicates that there is no setter for a context
type ErrNoSetterForContext struct {
	ParserName string
	Context    ParseContext
}

func (e *ErrNoSetterForContext) Error() string {
	return fmt.Sprintf("parser %s has no setter for context %s", e.ParserName, e.Context)
}

// ErrInvalidTarget indicates that the target is not of the expected type
type ErrInvalidTarget struct {
	ParserName   string
	Context      ParseContext
	ExpectedType string
	ActualType   string
}

func (e *ErrInvalidTarget) Error() string {
	return fmt.Sprintf("parser %s in context %s expects target type %s, got %s",
		e.ParserName, e.Context, e.ExpectedType, e.ActualType)
}

// ErrInvalidValue indicates that the value is not of the expected type
type ErrInvalidValue struct {
	ParserName   string
	ExpectedType string
	ActualType   string
}

func (e *ErrInvalidValue) Error() string {
	return fmt.Sprintf("parser %s expects value type %s, got %s",
		e.ParserName, e.ExpectedType, e.ActualType)
}

// ErrParseFailure indicates that parsing failed
type ErrParseFailure struct {
	ParserName string
	Context    ParseContext
	Cause      error
}

func (e *ErrParseFailure) Error() string {
	return fmt.Sprintf("parser %s failed in context %s: %v",
		e.ParserName, e.Context, e.Cause)
}

func (e *ErrParseFailure) Unwrap() error {
	return e.Cause
}
