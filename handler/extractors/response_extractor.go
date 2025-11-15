package extractors

import (
	"fmt"

	"github.com/reation-io/apikit/handler/parser"
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
	// Assign http.ResponseWriter to the payload field
	return fmt.Sprintf("payload.%s = w", field.Name), nil
}
