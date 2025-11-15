package extractors

import (
	"fmt"

	"github.com/reation-io/apikit/handler/parser"
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
	// Assign *http.Request to the payload field
	return fmt.Sprintf("payload.%s = r", field.Name), nil
}
