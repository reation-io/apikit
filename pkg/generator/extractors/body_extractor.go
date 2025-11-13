package extractors

import (
	"reflect"

	"github.com/reation-io/apikit/pkg/generator/parser"
)

func init() {
	Register(&BodyExtractor{})
}

// BodyExtractor extracts parameters from JSON body
type BodyExtractor struct{}

func (e *BodyExtractor) Name() string {
	return "body"
}

func (e *BodyExtractor) Priority() int {
	return 40 // Extract body last
}

func (e *BodyExtractor) CanExtract(field *parser.Field) bool {
	// Skip special fields - they have their own extractors
	if field.IsRequest || field.IsResponseWriter || field.IsRawBody {
		return false
	}

	// Body extraction is handled at the struct level, not field level
	// This extractor is used to detect if we need body parsing
	if field.IsBody {
		return true
	}

	// Check if field has json tag
	if field.StructTag != "" {
		tag := reflect.StructTag(field.StructTag)
		if _, ok := tag.Lookup("json"); ok {
			return true
		}
	}

	return false
}

func (e *BodyExtractor) GenerateCode(field *parser.Field, structName string) (string, []string) {
	// Body parsing is done at the struct level via json.Unmarshal
	// Individual fields don't need extraction code
	return "", nil
}
