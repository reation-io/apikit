package extractors

import (
	"fmt"
	"reflect"

	"github.com/reation-io/apikit/parser"
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
	return 35 // Extract cookies after headers but before body
}

func (e *CookieExtractor) CanExtract(field *parser.Field) bool {
	// Check if field has cookie tag
	if field.StructTag != "" {
		tag := reflect.StructTag(field.StructTag)
		if _, ok := tag.Lookup(parser.TagCookie); ok {
			return true
		}
	}
	// Check if field is marked with // in:cookie comment
	return field.InComment == parser.SourceCookie
}

func (e *CookieExtractor) GenerateCode(field *parser.Field, structName string) (string, []string) {
	cookieName := GetParameterName(field, parser.TagCookie)
	fieldName := field.Name
	typeName := GetBaseType(field)

	varName := fmt.Sprintf(`apikit.GetCookie(r, "%s")`, cookieName)

	// Use the public helper to generate code based on type
	return GenerateCodeByType(varName, fieldName, typeName, field)
}
