package extractors

import (
	"fmt"

	"github.com/reation-io/apikit/core/definition"
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

func (e *RequestExtractor) CanExtract(field *definition.Field) bool {
	return field.Type.GoType == "*http.Request"
}

func (e *RequestExtractor) GenerateCode(field *definition.Field, structName string) (string, []string) {
	// Assign *http.Request to the payload field
	return fmt.Sprintf("payload.%s = r", field.Name), nil
}
