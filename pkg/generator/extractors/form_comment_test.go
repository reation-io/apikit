package extractors

import (
	"strings"
	"testing"

	"github.com/reation-io/apikit/pkg/generator/parser"
)

func TestFormExtractor_CanExtract_WithComment(t *testing.T) {
	e := &FormExtractor{}

	tests := []struct {
		name     string
		field    *parser.Field
		expected bool
	}{
		{
			name: "field with in:form comment",
			field: &parser.Field{
				Name:      "Title",
				Type:      "string",
				InComment: "form",
			},
			expected: true,
		},
		{
			name: "field with in:form comment and name",
			field: &parser.Field{
				Name:          "Title",
				Type:          "string",
				InComment:     "form",
				InCommentName: "custom_title",
			},
			expected: true,
		},
		{
			name: "field with form tag",
			field: &parser.Field{
				Name:      "Title",
				Type:      "string",
				StructTag: `form:"title"`,
			},
			expected: true,
		},
		{
			name: "field without form tag or comment",
			field: &parser.Field{
				Name: "Title",
				Type: "string",
			},
			expected: false,
		},
		{
			name: "field with in:query comment (not form)",
			field: &parser.Field{
				Name:      "Title",
				Type:      "string",
				InComment: "query",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := e.CanExtract(tt.field)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFormExtractor_GenerateCode_WithCommentName(t *testing.T) {
	e := &FormExtractor{}

	tests := []struct {
		name              string
		field             *parser.Field
		expectedParamName string
	}{
		{
			name: "uses InCommentName when available",
			field: &parser.Field{
				Name:          "Title",
				Type:          "string",
				InComment:     "form",
				InCommentName: "custom_title",
			},
			expectedParamName: "custom_title",
		},
		{
			name: "uses tag value over InCommentName",
			field: &parser.Field{
				Name:          "Title",
				Type:          "string",
				StructTag:     `form:"tag_title"`,
				InComment:     "form",
				InCommentName: "comment_title",
			},
			expectedParamName: "tag_title",
		},
		{
			name: "uses field name when no tag or comment name",
			field: &parser.Field{
				Name:      "UserID",
				Type:      "string",
				InComment: "form",
			},
			expectedParamName: "userID", // camelCase conversion
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _ := e.GenerateCode(tt.field, "TestRequest")

			if !strings.Contains(code, tt.expectedParamName) {
				t.Errorf("expected code to contain parameter name %q, got:\n%s", tt.expectedParamName, code)
			}
		})
	}
}

func TestFormExtractor_GenerateCode_FileWithComment(t *testing.T) {
	e := &FormExtractor{}

	field := &parser.Field{
		Name:          "Avatar",
		Type:          "*multipart.FileHeader",
		InComment:     "form",
		InCommentName: "user_avatar",
		IsFile:        true,
	}

	code, imports := e.GenerateCode(field, "UploadRequest")

	if !strings.Contains(code, `r.FormFile("user_avatar")`) {
		t.Errorf("expected code to use custom name 'user_avatar', got:\n%s", code)
	}

	if !strings.Contains(code, "payload.Avatar") {
		t.Errorf("expected code to assign to payload.Avatar, got:\n%s", code)
	}

	hasMultipartImport := false
	for _, imp := range imports {
		if imp == "mime/multipart" {
			hasMultipartImport = true
			break
		}
	}

	if !hasMultipartImport {
		t.Error("expected mime/multipart import")
	}
}
