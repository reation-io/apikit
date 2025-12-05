package extractors

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/reation-io/apikit/core/definition"
	"github.com/reation-io/apikit/handler/parser"
)

func init() {
	Register(&HeaderExtractor{})
}

// HeaderExtractor extracts parameters from HTTP headers
type HeaderExtractor struct{}

func (e *HeaderExtractor) Name() string {
	return parser.SourceHeader
}

func (e *HeaderExtractor) Priority() int {
	return 30 // Extract headers after query params
}

func (e *HeaderExtractor) CanExtract(field *definition.Field) bool {
	// check existing tags
	if field.Tags != "" {
		tag := reflect.StructTag(field.Tags)
		if _, ok := tag.Lookup(parser.TagHeader); ok {
			return true
		}
	}
	// check metadata
	return field.Metadata["in"] == parser.SourceHeader
}

func (e *HeaderExtractor) GenerateCode(field *definition.Field, structName string) (string, []string) {
	headerName := GetParameterName(field, parser.TagHeader)
	fieldName := field.Name
	typeName := GetBaseType(field)

	// For slices, get all header values
	// Example: X-Tags: go, X-Tags: api, X-Tags: http → []string{"go", "api", "http"}
	if strings.HasPrefix(field.Type.GoType, "[]") {
		sliceType := field.Type.GoType[2:]
		varName := fmt.Sprintf(`r.Header["%s"]`, headerName)
		return GenerateSliceCodeByType(varName, fieldName, sliceType, field)
	}

	// For single values, use .Get()
	varName := fmt.Sprintf(`r.Header.Get("%s")`, headerName)

	// Use the public helper to generate code based on type
	return GenerateCodeByType(varName, fieldName, typeName, field)
}
