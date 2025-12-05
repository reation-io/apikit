package extractors

import (
	"fmt"
	"reflect"

	"github.com/reation-io/apikit/core/definition"
	"github.com/reation-io/apikit/handler/parser"
)

func init() {
	Register(&CookieExtractor{})
}

// CookieExtractor extracts parameters from HTTP cookies
type CookieExtractor struct{}

func (e *CookieExtractor) Name() string {
	return parser.SourceCookie
}

func (e *CookieExtractor) Priority() int {
	return 35 // Extract cookies after headers
}

func (e *CookieExtractor) CanExtract(field *definition.Field) bool {
	// check existing tags in field.Tags
	if field.Tags != "" {
		tag := reflect.StructTag(field.Tags)
		if _, ok := tag.Lookup(parser.TagCookie); ok {
			return true
		}
	}
	// check metadata for "in"
	return field.Metadata["in"] == parser.SourceCookie
}

func (e *CookieExtractor) GenerateCode(field *definition.Field, structName string) (string, []string) {
	cookieName := GetParameterName(field, parser.TagCookie)
	fieldName := field.Name
	typeName := GetBaseType(field)

	varName := fmt.Sprintf(`func() string { if c, err := r.Cookie("%s"); err == nil { return c.Value } else { return "" } }()`, cookieName)

	return GenerateCodeByType(varName, fieldName, typeName, field)
}
