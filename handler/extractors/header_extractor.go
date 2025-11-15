package extractors

import (
	"fmt"
	"reflect"

	"github.com/reation-io/apikit/handler/parser"
)

func init() {
	Register(&HeaderExtractor{})
}

// HeaderExtractor extracts parameters from HTTP headers
type HeaderExtractor struct{}

func (e *HeaderExtractor) Name() string {
	return "header"
}

func (e *HeaderExtractor) Priority() int {
	return 30 // Extract headers after query params
}

func (e *HeaderExtractor) CanExtract(field *parser.Field) bool {
	// Check if field has header tag
	if field.StructTag != "" {
		tag := reflect.StructTag(field.StructTag)
		if _, ok := tag.Lookup("header"); ok {
			return true
		}
	}
	// Check if field is marked with // in:header comment
	return field.InComment == "header"
}

func (e *HeaderExtractor) GenerateCode(field *parser.Field, structName string) (string, []string) {
	headerName := GetParameterName(field, "header")
	fieldName := field.Name
	typeName := GetBaseType(field)

	// For slices, get all header values
	// Example: X-Tags: go, X-Tags: api, X-Tags: http → []string{"go", "api", "http"}
	if field.IsSlice {
		varName := fmt.Sprintf(`r.Header["%s"]`, headerName)
		return GenerateSliceCodeByType(varName, fieldName, field.SliceType, field)
	}

	// For single values, use .Get()
	varName := fmt.Sprintf(`r.Header.Get("%s")`, headerName)

	// Use the public helper to generate code based on type
	return GenerateCodeByType(varName, fieldName, typeName, field)
}
