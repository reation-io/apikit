package extractors

import (
	"fmt"

	"github.com/reation-io/apikit/core/definition"
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

func (e *ResponseExtractor) CanExtract(field *definition.Field) bool {
	return field.Type.GoType == "http.ResponseWriter"
}

func (e *ResponseExtractor) GenerateCode(field *definition.Field, structName string) (string, []string) {
	// Assign http.ResponseWriter to the payload field
	return fmt.Sprintf("payload.%s = w", field.Name), nil
}
