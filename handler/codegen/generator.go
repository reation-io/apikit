// Package codegen generates Go code from parsed handler information using extractors
package codegen

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"reflect"
	"slices"
	"strings"
	"text/template"

	"github.com/reation-io/apikit/core/definition"
	"github.com/reation-io/apikit/handler/extractors"
	"golang.org/x/tools/imports"
)

//go:embed templates/handler.tmpl
var handlerTemplate string

// Generator generates wrapper code for handlers using the extractor system
type Generator struct {
	tmpl *template.Template
}

// New creates a new code generator
func New() (*Generator, error) {
	tmpl, err := template.New("handler").Funcs(templateFuncs()).Parse(handlerTemplate)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	return &Generator{
		tmpl: tmpl,
	}, nil
}

// TemplateData holds data for template execution
type TemplateData struct {
	PackageName string
	Imports     []string
	Handlers    []HandlerData
}

// HandlerData holds data for a single handler
type HandlerData struct {
	Name              string
	WrapperName       string
	ParseFuncName     string
	ParamType         string
	ReturnType        string
	HasExtractionCode bool
	ExtractionCode    string
	HasBody           bool
	BodyFieldName     string
	HasRawBody        bool
	RawBodyFieldName  string
	HasValidation     bool
	HasResponseWriter bool
	HasRequest        bool
	HasMultipartForm  bool
	MaxMemory         int64 // Max memory for multipart form parsing (default 32MB)
}

// Generate creates wrapper code for the given API definition
func (g *Generator) Generate(def *definition.Definition) ([]byte, error) {
	if len(def.Operations) == 0 {
		return nil, fmt.Errorf("no handlers found")
	}

	// Prepare template data using extractors
	data := g.prepareTemplateData(def)

	// Execute template
	var buf bytes.Buffer
	if err := g.tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}

	// Format with goimports (handles imports and formatting)
	formatted, err := imports.Process("", buf.Bytes(), nil)
	if err != nil {
		// Fallback to basic formatting
		formatted, err = format.Source(buf.Bytes())
		if err != nil {
			// Return nil with error - unformatted code indicates a serious issue
			// The caller should not use malformed code
			return nil, fmt.Errorf("formatting code: %w", err)
		}
	}

	return formatted, nil
}

func (g *Generator) prepareTemplateData(def *definition.Definition) *TemplateData {
	data := &TemplateData{
		PackageName: def.Package,
		Imports:     []string{},
		Handlers:    []HandlerData{},
	}

	importsMap := make(map[string]bool)

	// Always add apikit import since we use it for error handling
	importsMap["github.com/reation-io/apikit"] = true

	for _, op := range def.Operations {
		hd := g.prepareHandlerData(op, def, importsMap)
		data.Handlers = append(data.Handlers, hd)
	}

	// Convert imports map to slice and sort alphabetically for deterministic output
	for imp := range importsMap {
		data.Imports = append(data.Imports, imp)
	}
	slices.Sort(data.Imports)

	return data
}

func (g *Generator) prepareHandlerData(op *definition.Operation, def *definition.Definition, importsMap map[string]bool) HandlerData {
	hd := HandlerData{
		Name:          op.ID,
		WrapperName:   toCamelCasePrivate(op.ID) + "APIKit",
		ParseFuncName: "parse" + capitalize(op.ID) + "Request",
		ParamType:     op.RequestType,
		ReturnType:    op.ReturnType,
	}

	// If no request struct, we are done
	if op.RequestType == "" {
		return hd
	}

	reqType, ok := def.Types[op.RequestType]
	if !ok {
		return hd
	}

	// Use extractors to generate code for each field
	extractionCode := g.generateExtractionCode(reqType, op.ID, importsMap)

	hd.HasExtractionCode = extractionCode != ""
	hd.ExtractionCode = extractionCode

	// Check for body fields
	hd.HasBody = g.hasBodyFields(reqType)
	if hd.HasBody {
		bodyField := g.findBodyField(reqType)
		if bodyField != "" {
			hd.BodyFieldName = bodyField
		}
	}

	// Check for RawBody
	rawBodyField := g.findRawBodyField(reqType)
	if rawBodyField != "" {
		hd.HasRawBody = true
		hd.RawBodyFieldName = rawBodyField
	}

	// Check validation
	hd.HasValidation = g.hasValidationTags(reqType)
	if hd.HasValidation {
		importsMap["github.com/reation-io/apikit/validator"] = true
	}

	// Check multipart
	hd.HasMultipartForm = g.hasMultipartFormFields(reqType)
	if hd.HasMultipartForm {
		hd.MaxMemory = 32 << 20 // 32MB default
	}

	// Check for special fields in the struct (Request/ResponseWriter)
	// Some patterns put them in the struct.
	for _, f := range reqType.Fields {
		if f.Type.GoType == "*http.Request" {
			hd.HasRequest = true
		}
		if f.Type.GoType == "http.ResponseWriter" {
			hd.HasResponseWriter = true
		}
	}

	return hd
}

func (g *Generator) generateExtractionCode(t *definition.Type, structName string, importsMap map[string]bool) string {
	var lines []string

	allExtractors := extractors.GetExtractors()

	for _, field := range t.Fields {
		// handle embedded fields if Type.Kind == "struct" and it has no name?
		// for now skip complex embedded logic unless explicit

		// Skip special/body fields handled separately
		if field.Name == "Body" && field.Type.GoType == "[]byte" {
			continue
		} // RawBody
		if field.Type.GoType == "*http.Request" || field.Type.GoType == "http.ResponseWriter" {
			continue
		}

		// Find extractor
		for _, ext := range allExtractors {
			if ext.CanExtract(field) {
				code, imports := ext.GenerateCode(field, structName)
				if code != "" {
					for _, imp := range imports {
						importsMap[imp] = true
					}
					lines = append(lines, code)
				}
				break
			}
		}
	}

	return strings.Join(lines, "\n\t")
}

func (g *Generator) hasBodyFields(t *definition.Type) bool {
	for _, field := range t.Fields {
		if field.Metadata["in"] == "body" {
			return true
		}
		if field.Tags != "" {
			tag := reflect.StructTag(field.Tags)
			if val, ok := tag.Lookup("json"); ok && val == "body" {
				return true
			}
		}

		// Also BodyExtractor knows. But we are here just checking flags.
		if field.Name == "Body" && field.Type.GoType == "[]byte" {
			return false
		} // RawBody is separate
	}
	return false
}

func (g *Generator) findBodyField(t *definition.Type) string {
	for _, field := range t.Fields {
		if field.Metadata["in"] == "body" {
			return field.Name
		}
		if field.Tags != "" {
			tag := reflect.StructTag(field.Tags)
			if val, ok := tag.Lookup("json"); ok && val == "body" {
				return field.Name
			}
		}
	}
	return ""
}

func (g *Generator) findRawBodyField(t *definition.Type) string {
	for _, field := range t.Fields {
		if field.Type.GoType == "[]byte" && (field.Name == "Body" || strings.Contains(strings.ToLower(field.Name), "rawbody")) {
			return field.Name
		}
	}
	return ""
}

func (g *Generator) hasValidationTags(t *definition.Type) bool {
	for _, field := range t.Fields {
		if field.Tags != "" {
			tag := reflect.StructTag(field.Tags)
			if _, ok := tag.Lookup("validate"); ok {
				return true
			}
		}
	}
	return false
}

func (g *Generator) hasMultipartFormFields(t *definition.Type) bool {
	for _, field := range t.Fields {
		if field.Metadata["in"] == "form" {
			return true
		}
		if field.Tags != "" {
			tag := reflect.StructTag(field.Tags)
			if _, ok := tag.Lookup("form"); ok {
				return true
			}
		}
		if strings.Contains(field.Type.GoType, "multipart.FileHeader") {
			return true
		}
	}
	return false
}

// Template helper functions
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"toLower": strings.ToLower,
		"toUpper": strings.ToUpper,
	}
}

// toCamelCasePrivate converts a string to camelCase with first letter lowercase
func toCamelCasePrivate(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	if len(runes) > 0 {
		runes[0] = []rune(strings.ToLower(string(runes[0])))[0]
	}
	return string(runes)
}

// capitalize converts the first letter to uppercase (PascalCase)
func capitalize(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	if len(runes) > 0 {
		runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
	}
	return string(runes)
}
