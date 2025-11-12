package extractors

import (
	"github.com/reation-io/apikit/pkg/generator/parser"
)

func init() {
	Register(&RequestExtractor{})
}

// RequestExtractor handles *http.Request parameter
type RequestExtractor struct{}

func (e *RequestExtractor) Name() string {
	return "request"
}

func (e *RequestExtractor) Priority() int {
	return 50 // Extract after all other params
}

func (e *RequestExtractor) CanExtract(field *parser.Field) bool {
	return field.IsRequest
}

func (e *RequestExtractor) GenerateCode(field *parser.Field, structName string) (string, []string) {
	// *http.Request is passed directly to the handler, no extraction needed
	return "", nil
}
