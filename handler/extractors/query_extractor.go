package extractors

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/reation-io/apikit/core/definition"
	"github.com/reation-io/apikit/handler/parser"
)

func init() {
	Register(&QueryExtractor{})
}

// QueryExtractor extracts parameters from URL query string
type QueryExtractor struct{}

func (e *QueryExtractor) Name() string {
	return parser.SourceQuery
}

func (e *QueryExtractor) Priority() int {
	return 20 // Extract query params after path
}

func (e *QueryExtractor) CanExtract(field *definition.Field) bool {
	// check existing tags in field.Tags
	if field.Tags != "" {
		tag := reflect.StructTag(field.Tags)
		if _, ok := tag.Lookup(parser.TagQuery); ok {
			return true
		}
	}
	// check metadata for "in"
	return field.Metadata["in"] == parser.SourceQuery
}

func (e *QueryExtractor) GenerateCode(field *definition.Field, structName string) (string, []string) {
	paramName := GetParameterName(field, parser.TagQuery)
	fieldName := field.Name
	typeName := GetBaseType(field)

	// For slices, get all values using []
	// Example: ?tags=go&tags=api&tags=http → []string{"go", "api", "http"}
	// We check for slice by checking if GoType starts with []
	if strings.HasPrefix(field.Type.GoType, "[]") {
		// Slice type logic
		sliceType := field.Type.GoType[2:] // remove []
		varName := fmt.Sprintf(`r.URL.Query()["%s"]`, paramName)
		return GenerateSliceCodeByType(varName, fieldName, sliceType, field)
	}

	// For single values, use .Get()
	varName := fmt.Sprintf(`r.URL.Query().Get("%s")`, paramName)

	// Use the public helper to generate code based on type
	return GenerateCodeByType(varName, fieldName, typeName, field)
}
