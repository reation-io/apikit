package codegen

import (
	"testing"

	"github.com/reation-io/apikit/pkg/generator/extractors"
	"github.com/reation-io/apikit/pkg/generator/parser"
)

func TestExtractors_FormExtractorRegistered(t *testing.T) {
	allExtractors := extractors.GetExtractors()

	var formExtractor extractors.Extractor
	for _, ext := range allExtractors {
		if ext.Name() == "form" {
			formExtractor = ext
			break
		}
	}

	if formExtractor == nil {
		t.Fatal("FormExtractor not registered")
	}

	t.Logf("FormExtractor found with priority: %d", formExtractor.Priority())

	// Test if it can extract a form field
	field := parser.Field{
		Name:      "Image",
		Type:      "*multipart.FileHeader",
		StructTag: `form:"image"`,
		IsFile:    true,
	}

	if !formExtractor.CanExtract(&field) {
		t.Error("FormExtractor should be able to extract file field")
	}

	code, imports := formExtractor.GenerateCode(&field, "TestStruct")
	t.Logf("Generated code:\n%s", code)
	t.Logf("Imports: %v", imports)

	if len(imports) == 0 {
		t.Error("expected imports to be returned")
	}
}

