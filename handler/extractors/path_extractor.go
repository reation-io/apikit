package extractors

import (
	"fmt"
	"reflect"

	"github.com/reation-io/apikit/core/definition"
	"github.com/reation-io/apikit/handler/parser"
)

func init() {
	Register(&PathExtractor{})
}

// PathExtractor extracts parameters from URL path
type PathExtractor struct{}

func (e *PathExtractor) Name() string {
	return parser.SourcePath
}

func (e *PathExtractor) Priority() int {
	return 10 // Extract path params first
}

func (e *PathExtractor) CanExtract(field *definition.Field) bool {
	// Check if field has path tag
	if field.Tags != "" {
		tag := reflect.StructTag(field.Tags)
		if _, ok := tag.Lookup(parser.TagPath); ok {
			return true
		}
	}
	// check metadata
	return field.Metadata["in"] == parser.SourcePath
}

func (e *PathExtractor) GenerateCode(field *definition.Field, structName string) (string, []string) {
	paramName := GetParameterName(field, parser.TagPath)
	fieldName := field.Name
	typeName := GetBaseType(field)

	varName := fmt.Sprintf(`r.PathValue("%s")`, paramName)

	// Use the public helper to generate code based on type
	return GenerateCodeByType(varName, fieldName, typeName, field)
}
