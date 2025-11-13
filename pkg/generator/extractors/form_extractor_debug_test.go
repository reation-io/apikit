package extractors

import (
	"testing"

	"github.com/reation-io/apikit/pkg/generator/parser"
)

func TestFormExtractor_GenerateCode_FileField_Debug(t *testing.T) {
	e := &FormExtractor{}
	field := &parser.Field{
		Name:      "Image",
		Type:      "*multipart.FileHeader",
		StructTag: `form:"image"`,
		IsFile:    true,
	}

	code, imports := e.GenerateCode(field, "UploadRequest")

	t.Logf("Generated code:\n%s", code)
	t.Logf("Imports: %v", imports)

	if len(imports) == 0 {
		t.Error("expected imports to be returned")
	}

	hasMultipartImport := false
	for _, imp := range imports {
		if imp == "mime/multipart" {
			hasMultipartImport = true
			break
		}
	}

	if !hasMultipartImport {
		t.Errorf("expected mime/multipart import, got: %v", imports)
	}
}

