package extractors

import (
	"reflect"
	"strings"

	"github.com/reation-io/apikit/core/definition"
	"github.com/reation-io/apikit/handler/parser"
)

func init() {
	Register(&BodyExtractor{})
}

// BodyExtractor extracts parameters from JSON body
type BodyExtractor struct{}

func (e *BodyExtractor) Name() string {
	return parser.SourceBody
}

func (e *BodyExtractor) Priority() int {
	return 40 // Extract body last
}

func (e *BodyExtractor) CanExtract(field *definition.Field) bool {
	// Skip special fields - checked via type or name?
	// core/definition doesn't explicitly flag IsRequest/IsResponseWriter yet
	// But those are usually separate arguments in the handler function, not part of the request struct...
	// Wait, in apikit (original), standard http fields are SOMETIMES in the request struct?
	// Looking at `handler/parser/adapter.go`, `extractSpecialFields`.
	// If `core/parser` doesn't filter them out from the "Type", they are present.
	// We can check the Type name.

	if field.Type.GoType == "*http.Request" || field.Type.GoType == "http.ResponseWriter" {
		return false
	}
	// Check for RawBody (usually []byte with "body" tag or name)
	if field.Type.GoType == "[]byte" && (field.Name == "Body" || strings.Contains(strings.ToLower(field.Name), "rawbody")) {
		return false
	}

	// Body extraction is handled at the struct level, not field level
	// This extractor is used to detect if we need body parsing
	// core/parser marks Metadata["in"] = "body" ? No, body is default if json tag exists.

	// Check if field has json tag
	if field.Tags != "" {
		tag := reflect.StructTag(field.Tags)
		if val, ok := tag.Lookup("json"); ok && val != "-" {
			return true
		}
	}

	return false
}

func (e *BodyExtractor) GenerateCode(field *definition.Field, structName string) (string, []string) {
	// Body parsing is done at the struct level via json.Unmarshal
	// Individual fields don't need extraction code
	return "", nil
}
