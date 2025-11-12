package extractors

import (
	"fmt"
	"reflect"

	"github.com/reation-io/apikit/pkg/generator/parser"
)

func init() {
	Register(&QueryExtractor{})
}

// QueryExtractor extracts parameters from URL query string
type QueryExtractor struct{}

func (e *QueryExtractor) Name() string {
	return "query"
}

func (e *QueryExtractor) Priority() int {
	return 20 // Extract query params after path
}

func (e *QueryExtractor) CanExtract(field *parser.Field) bool {
	// Check if field has query tag
	if field.StructTag != "" {
		tag := reflect.StructTag(field.StructTag)
		if _, ok := tag.Lookup("query"); ok {
			return true
		}
	}
	// Check if field is marked with // in:query comment
	return field.InComment == "query"
}

func (e *QueryExtractor) GenerateCode(field *parser.Field, structName string) (string, []string) {
	paramName := GetParameterName(field, "query")
	fieldName := field.Name
	typeName := GetBaseType(field)

	// For slices, get all values using []
	// Example: ?tags=go&tags=api&tags=http → []string{"go", "api", "http"}
	if field.IsSlice {
		varName := fmt.Sprintf(`r.URL.Query()["%s"]`, paramName)
		return GenerateSliceCodeByType(varName, fieldName, field.SliceType, field)
	}

	// For single values, use .Get()
	varName := fmt.Sprintf(`r.URL.Query().Get("%s")`, paramName)

	// Use the public helper to generate code based on type
	return GenerateCodeByType(varName, fieldName, typeName, field)
}
