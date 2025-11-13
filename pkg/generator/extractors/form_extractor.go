package extractors

import (
	"fmt"
	"reflect"

	"github.com/reation-io/apikit/pkg/generator/parser"
)

func init() {
	Register(&FormExtractor{})
}

// FormExtractor extracts parameters from multipart form-data
type FormExtractor struct{}

func (e *FormExtractor) Name() string {
	return "form"
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
		if _, ok := tag.Lookup("form"); ok {
			return true
		}
	}

	// Check if field is marked with // in:form comment
	return field.InComment == "form"
}

func (e *FormExtractor) GenerateCode(field *parser.Field, structName string) (string, []string) {
	var imports []string
	var code string

	// Get the form field name from tag or use field name
	formName := GetParameterName(field, "form")

	// Handle file uploads
	if field.IsFile {
		code = e.generateFileCode(field, formName)
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

func (e *FormExtractor) generateFileCode(field *parser.Field, formName string) string {
	if field.IsSlice {
		// Multiple files: []*multipart.FileHeader
		return fmt.Sprintf(`if form := r.MultipartForm; form != nil {
		if files := form.File["%s"]; len(files) > 0 {
			payload.%s = files
		}
	}`, formName, field.Name)
	}

	// Single file: *multipart.FileHeader
	return fmt.Sprintf(`if _, header, err := r.FormFile("%s"); err == nil {
		payload.%s = header
	} else if err != http.ErrMissingFile {
		return fmt.Errorf("reading file '%s': %%w", err)
	}`, formName, field.Name, formName)
}
