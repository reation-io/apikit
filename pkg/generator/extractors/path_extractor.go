package extractors

import (
	"fmt"
	"reflect"

	"github.com/reation-io/apikit/pkg/generator/parser"
)

func init() {
	Register(&PathExtractor{})
}

// PathExtractor extracts parameters from URL path
type PathExtractor struct{}

func (e *PathExtractor) Name() string {
	return "path"
}

func (e *PathExtractor) Priority() int {
	return 10 // Extract path params first
}

func (e *PathExtractor) CanExtract(field *parser.Field) bool {
	// Check if field has path tag
	if field.StructTag != "" {
		tag := reflect.StructTag(field.StructTag)
		if _, ok := tag.Lookup("path"); ok {
			return true
		}
	}
	// Check if field is marked with // in:path comment
	return field.InComment == "path"
}

func (e *PathExtractor) GenerateCode(field *parser.Field, structName string) (string, []string) {
	paramName := GetParameterName(field, "path")
	fieldName := field.Name
	typeName := GetBaseType(field)

	varName := fmt.Sprintf(`r.PathValue("%s")`, paramName)

	// Use the public helper to generate code based on type
	return GenerateCodeByType(varName, fieldName, typeName, field)
}
