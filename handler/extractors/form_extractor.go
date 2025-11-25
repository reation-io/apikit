package extractors

import (
	"fmt"
	"reflect"
	"strings"

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

func (e *FormExtractor) CanExtract(field *parser.Field) bool {
	// Skip special fields
	if field.IsRequest || field.IsResponseWriter || field.IsRawBody {
		return false
	}

	// Check if field has form tag
	if field.StructTag != "" {
		tag := reflect.StructTag(field.StructTag)
		if _, ok := tag.Lookup(parser.TagForm); ok {
			return true
		}
	}

	// Check if field is marked with // in:form comment
	return field.InComment == parser.SourceForm
}

func (e *FormExtractor) GenerateCode(field *parser.Field, structName string) (string, []string) {
	var imports []string
	var code string

	// Handle nested structs marked with // in:form
	// These should have their fields extracted individually
	if field.NestedStruct != nil && field.InComment == parser.SourceForm {
		return e.generateNestedStructCode(field, &imports)
	}

	// Get the form field name from tag or use field name
	formName := GetParameterName(field, parser.TagForm)

	// Handle file uploads
	if field.IsFile {
		code = e.generateFileCode(field, formName, field.Name)
		imports = append(imports, "mime/multipart")
		return code, imports
	}

	// Handle regular form fields
	if field.IsSlice {
		// For slices, use r.Form[key] which returns []string
		code = fmt.Sprintf(`if vals := r.Form["%s"]; len(vals) > 0 {
		payload.%s = vals
	}`, formName, field.Name)
	} else {
		// For single values, use r.FormValue
		varName := fmt.Sprintf(`r.FormValue("%s")`, formName)

		// Generate type-specific parsing code
		typeCode, typeImports := GenerateCodeByType(varName, field.Name, field.Type, field)
		code = typeCode
		imports = append(imports, typeImports...)
	}

	return code, imports
}

// generateNestedStructCode generates extraction code for nested struct fields
func (e *FormExtractor) generateNestedStructCode(field *parser.Field, imports *[]string) (string, []string) {
	var lines []string

	// Iterate through nested struct fields
	for _, nestedField := range field.NestedStruct.Fields {
		// Skip special fields
		if nestedField.IsRequest || nestedField.IsResponseWriter || nestedField.IsRawBody {
			continue
		}

		// Check if nested field should be extracted as form field
		hasFormTag := false
		if nestedField.StructTag != "" {
			tag := reflect.StructTag(nestedField.StructTag)
			if _, ok := tag.Lookup(parser.TagForm); ok {
				hasFormTag = true
			}
		}

		// Extract if it has form tag or // in:form comment
		if hasFormTag || nestedField.InComment == parser.SourceForm {
			formName := GetParameterName(&nestedField, parser.TagForm)
			fieldPath := fmt.Sprintf("%s.%s", field.Name, nestedField.Name)

			// Handle file uploads
			if nestedField.IsFile {
				code := e.generateFileCode(&nestedField, formName, fieldPath)
				lines = append(lines, code)
				*imports = append(*imports, "mime/multipart")
				continue
			}

			// Handle regular form fields
			if nestedField.IsSlice {
				code := fmt.Sprintf(`if vals := r.Form["%s"]; len(vals) > 0 {
			payload.%s = vals
		}`, formName, fieldPath)
				lines = append(lines, code)
			} else {
				varName := fmt.Sprintf(`r.FormValue("%s")`, formName)
				typeCode, typeImports := GenerateCodeByType(varName, fieldPath, nestedField.Type, &nestedField)
				lines = append(lines, typeCode)
				*imports = append(*imports, typeImports...)
			}
		}
	}

	return strings.Join(lines, "\n\t"), *imports
}

func (e *FormExtractor) generateFileCode(field *parser.Field, formName string, fieldPath string) string {
	if fieldPath == "" {
		fieldPath = field.Name
	}

	if field.IsSlice {
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
