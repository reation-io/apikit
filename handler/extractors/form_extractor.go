package extractors

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/reation-io/apikit/core/definition"
	"github.com/reation-io/apikit/handler/parser"
)

func init() {
	Register(&FormExtractor{})
}

// FormExtractor extracts parameters from multipart form-data
type FormExtractor struct{}

func (e *FormExtractor) Name() string {
	return parser.SourceForm
}

func (e *FormExtractor) Priority() int {
	return 15 // Extract form fields after path but before query
}

func (e *FormExtractor) CanExtract(field *definition.Field) bool {
	// Skip special fields (Request/ResponseWriter)
	// definition doesn't explicit flag, check type
	if field.Type.GoType == "*http.Request" || field.Type.GoType == "http.ResponseWriter" {
		return false
	}
	// Skip RawBody
	if field.Type.GoType == "[]byte" && (field.Name == "Body" || strings.Contains(strings.ToLower(field.Name), "rawbody")) {
		return false
	}

	// Check if field has form tag
	if field.Tags != "" {
		tag := reflect.StructTag(field.Tags)
		if _, ok := tag.Lookup(parser.TagForm); ok {
			return true
		}
	}

	// Check metadata
	return field.Metadata["in"] == parser.SourceForm
}

func (e *FormExtractor) GenerateCode(field *definition.Field, structName string) (string, []string) {
	var imports []string
	var code string

	// Handle nested structs marked with // in:form
	// These should have their fields extracted individually
	// In definition model, if Type.Kind is struct, it's nested
	if field.Type.Kind == "struct" && field.Metadata["in"] == parser.SourceForm {
		return e.generateNestedStructCode(field, &imports)
	}

	// Get the form field name from tag or use field name
	formName := GetParameterName(field, parser.TagForm)

	// Handle file uploads (check type name since we lack IsFile flag)
	// Usually *multipart.FileHeader or []*multipart.FileHeader
	if strings.Contains(field.Type.GoType, "multipart.FileHeader") {
		code = e.generateFileCode(field, formName, field.Name)
		imports = append(imports, "mime/multipart")
		return code, imports
	}

	// Handle regular form fields
	if strings.HasPrefix(field.Type.GoType, "[]") {
		// For slices, use r.Form[key] which returns []string
		code = fmt.Sprintf(`if vals := r.Form["%s"]; len(vals) > 0 {
		payload.%s = vals
	}`, formName, field.Name)
	} else {
		// For single values, use r.FormValue
		varName := fmt.Sprintf(`r.FormValue("%s")`, formName)

		// Generate type-specific parsing code
		typeName := GetBaseType(field)
		typeCode, typeImports := GenerateCodeByType(varName, field.Name, typeName, field)
		code = typeCode
		imports = append(imports, typeImports...)
	}

	return code, imports
}

// generateNestedStructCode generates extraction code for nested struct fields
func (e *FormExtractor) generateNestedStructCode(field *definition.Field, imports *[]string) (string, []string) {
	var lines []string

	if field.Type == nil {
		return "", *imports
	}

	// Iterate through nested struct fields
	for _, nestedField := range field.Type.Fields {
		// Skip special fields (simplified check)
		if nestedField.Type.GoType == "*http.Request" {
			continue
		}

		// Check if nested field should be extracted as form field
		hasFormTag := false
		if nestedField.Tags != "" {
			tag := reflect.StructTag(nestedField.Tags)
			if _, ok := tag.Lookup(parser.TagForm); ok {
				hasFormTag = true
			}
		}

		// Extract if it has form tag or metadata["in"]
		if hasFormTag || nestedField.Metadata["in"] == parser.SourceForm {
			formName := GetParameterName(nestedField, parser.TagForm)
			fieldPath := fmt.Sprintf("%s.%s", field.Name, nestedField.Name)

			// Handle file uploads
			if strings.Contains(nestedField.Type.GoType, "multipart.FileHeader") {
				code := e.generateFileCode(nestedField, formName, fieldPath)
				lines = append(lines, code)
				*imports = append(*imports, "mime/multipart")
				continue
			}

			// Handle regular form fields
			if strings.HasPrefix(nestedField.Type.GoType, "[]") {
				code := fmt.Sprintf(`if vals := r.Form["%s"]; len(vals) > 0 {
			payload.%s = vals
		}`, formName, fieldPath)
				lines = append(lines, code)
			} else {
				varName := fmt.Sprintf(`r.FormValue("%s")`, formName)
				typeName := GetBaseType(nestedField)
				typeCode, typeImports := GenerateCodeByType(varName, fieldPath, typeName, nestedField)
				lines = append(lines, typeCode)
				*imports = append(*imports, typeImports...)
			}
		}
	}

	return strings.Join(lines, "\n\t"), *imports
}

func (e *FormExtractor) generateFileCode(field *definition.Field, formName string, fieldPath string) string {
	if fieldPath == "" {
		fieldPath = field.Name
	}

	if strings.HasPrefix(field.Type.GoType, "[]") {
		// Multiple files: []*multipart.FileHeader
		return fmt.Sprintf(`if form := r.MultipartForm; form != nil {
		if files := form.File["%s"]; len(files) > 0 {
			payload.%s = files
		}
	}`, formName, fieldPath)
	}

	// Single file: *multipart.FileHeader
	return fmt.Sprintf(`if _, header, err := r.FormFile("%s"); err == nil {
		payload.%s = header
	} else if err != http.ErrMissingFile {
		return fmt.Errorf("reading file '%s': %%w", err)
	}`, formName, fieldPath, formName)
}
