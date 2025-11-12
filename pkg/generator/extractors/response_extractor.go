package extractors

import (
	"github.com/reation-io/apikit/pkg/generator/parser"
)

func init() {
	Register(&ResponseExtractor{})
}

// ResponseExtractor handles http.ResponseWriter parameter
type ResponseExtractor struct{}

func (e *ResponseExtractor) Name() string {
	return "response"
}

func (e *ResponseExtractor) Priority() int {
	return 60 // Extract after all other params
}

func (e *ResponseExtractor) CanExtract(field *parser.Field) bool {
	return field.IsResponseWriter
}

func (e *ResponseExtractor) GenerateCode(field *parser.Field, structName string) (string, []string) {
	// http.ResponseWriter is passed directly to the handler, no extraction needed
	return "", nil
}
