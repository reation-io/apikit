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

	"github.com/reation-io/apikit/pkg/generator/extractors"
	"github.com/reation-io/apikit/pkg/generator/parser"
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

// Generate creates wrapper code for the given handlers
func (g *Generator) Generate(result *parser.ParseResult) ([]byte, error) {
	if len(result.Handlers) == 0 {
		return nil, fmt.Errorf("no handlers found")
	}

	// Prepare template data using extractors
	data := g.prepareTemplateData(result)

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

func (g *Generator) prepareTemplateData(result *parser.ParseResult) *TemplateData {
	data := &TemplateData{
		PackageName: result.Source.Package,
		Imports:     []string{},
		Handlers:    []HandlerData{},
	}

	importsMap := make(map[string]bool)

	// Always add apikit import since we use it for error handling
	importsMap["github.com/reation-io/apikit"] = true

	for _, handler := range result.Handlers {
		hd := g.prepareHandlerData(&handler, importsMap)
		data.Handlers = append(data.Handlers, hd)
	}

	// Convert imports map to slice and sort alphabetically for deterministic output
	for imp := range importsMap {
		data.Imports = append(data.Imports, imp)
	}
	slices.Sort(data.Imports)

	return data
}

func (g *Generator) prepareHandlerData(handler *parser.Handler, importsMap map[string]bool) HandlerData {
	hd := HandlerData{
		Name:              handler.Name,
		WrapperName:       toCamelCasePrivate(handler.Name) + "APIKit",
		ParseFuncName:     "parse" + capitalize(handler.Name) + "Request",
		ParamType:         handler.ParamType,
		ReturnType:        handler.ReturnType,
		HasResponseWriter: handler.HasResponseWriter,
		HasRequest:        handler.HasRequest,
	}

	if handler.Struct == nil {
		return hd
	}

	// Use extractors to generate code for each field
	extractionCode := g.generateExtractionCode(handler.Struct, importsMap)

	hd.HasExtractionCode = extractionCode != ""
	hd.ExtractionCode = extractionCode

	// Check if we need body parsing and find the body field name
	hd.HasBody = g.hasBodyFields(handler.Struct)
	if hd.HasBody {
		bodyField := g.findBodyField(handler.Struct)
		if bodyField != "" {
			hd.BodyFieldName = bodyField
		}
	}

	// Check if there's a RawBody field
	rawBodyField := g.findRawBodyField(handler.Struct)
	if rawBodyField != "" {
		hd.HasRawBody = true
		hd.RawBodyFieldName = rawBodyField
	}

	// Check if validation is needed
	hd.HasValidation = g.hasValidationTags(handler.Struct)
	if hd.HasValidation {
		// Add validator import
		importsMap["github.com/reation-io/apikit/validator"] = true
	}

	// Check if multipart form parsing is needed
	hd.HasMultipartForm = g.hasMultipartFormFields(handler.Struct)
	if hd.HasMultipartForm {
		hd.MaxMemory = 32 << 20 // 32MB default
	}

	return hd
}

func (g *Generator) generateExtractionCode(s *parser.Struct, importsMap map[string]bool) string {
	var lines []string

	// Get all registered extractors (already sorted by priority)
	allExtractors := extractors.GetExtractors()

	// Process each field
	for _, field := range s.Fields {
		// Handle embedded structs - expand their fields
		if field.IsEmbedded {
			if field.NestedStruct != nil {
				nestedCode := g.generateExtractionCode(field.NestedStruct, importsMap)
				if nestedCode != "" {
					lines = append(lines, nestedCode)
				}
			}
			continue
		}

		// Skip RawBody field (handled separately in template)
		if field.IsRawBody {
			continue
		}

		// Find the appropriate extractor for this field
		for _, ext := range allExtractors {
			if ext.CanExtract(&field) {
				code, imports := ext.GenerateCode(&field, s.Name)
				if code != "" {
					// Add imports
					for _, imp := range imports {
						importsMap[imp] = true
					}

					// Add code (extractors are already sorted by priority)
					lines = append(lines, code)
				}
				break // Only use the first matching extractor
			}
		}
	}

	return strings.Join(lines, "\n\t")
}

func (g *Generator) hasBodyFields(s *parser.Struct) bool {
	for _, field := range s.Fields {
		// Check embedded structs recursively
		if field.IsEmbedded && field.NestedStruct != nil {
			if g.hasBodyFields(field.NestedStruct) {
				return true
			}
		}

		// Field is a body field if:
		// 1. It has IsBody = true (from "in: body" comment), OR
		// 2. It has json:"body" tag
		if field.IsBody {
			return true
		}

		if field.StructTag != "" {
			tag := reflect.StructTag(field.StructTag)
			if jsonTag, ok := tag.Lookup("json"); ok && jsonTag == "body" {
				return true
			}
		}
	}
	return false
}

// findBodyField searches for a body field in the struct
// Returns the field name if found, empty string otherwise
func (g *Generator) findBodyField(s *parser.Struct) string {
	for _, field := range s.Fields {
		// Check embedded structs recursively
		if field.IsEmbedded && field.NestedStruct != nil {
			if bodyField := g.findBodyField(field.NestedStruct); bodyField != "" {
				return bodyField
			}
		}

		// Check if this is a body field
		if field.IsBody {
			return field.Name
		}

		// Check if field has json:"body" tag
		if field.StructTag != "" {
			tag := reflect.StructTag(field.StructTag)
			if jsonTag, ok := tag.Lookup("json"); ok && jsonTag == "body" {
				return field.Name
			}
		}
	}
	return ""
}

// findRawBodyField searches for a RawBody field ([]byte) in the struct
// Returns the field name if found, empty string otherwise
func (g *Generator) findRawBodyField(s *parser.Struct) string {
	for _, field := range s.Fields {
		// Check embedded structs recursively
		if field.IsEmbedded && field.NestedStruct != nil {
			if rawBodyField := g.findRawBodyField(field.NestedStruct); rawBodyField != "" {
				return rawBodyField
			}
		}

		// Check if this is a RawBody field
		// More flexible detection: any field with type []byte that contains "body" (case-insensitive)
		if field.IsRawBody {
			return field.Name
		}
	}
	return ""
}

// hasValidationTags checks if the struct has any validation tags
// Returns true if any field has a validate tag
func (g *Generator) hasValidationTags(s *parser.Struct) bool {
	for _, field := range s.Fields {
		// Check embedded structs recursively
		if field.IsEmbedded && field.NestedStruct != nil {
			if g.hasValidationTags(field.NestedStruct) {
				return true
			}
		}

		// Check if this field has a validate tag
		if field.StructTag != "" {
			tag := reflect.StructTag(field.StructTag)
			if _, ok := tag.Lookup("validate"); ok {
				return true
			}
		}
	}
	return false
}

// hasMultipartFormFields checks if the struct has any multipart form fields
// Returns true if any field has a form tag or is a file upload field
func (g *Generator) hasMultipartFormFields(s *parser.Struct) bool {
	for _, field := range s.Fields {
		// Check embedded structs recursively
		if field.IsEmbedded && field.NestedStruct != nil {
			if g.hasMultipartFormFields(field.NestedStruct) {
				return true
			}
		}

		// Check if this is a file field
		if field.IsFile {
			return true
		}

		// Check if this field has a form tag
		if field.StructTag != "" {
			tag := reflect.StructTag(field.StructTag)
			if _, ok := tag.Lookup("form"); ok {
				return true
			}
		}

		// Check if field is marked with // in:form comment
		if field.InComment == "form" {
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
// Example: "GetUser" -> "getUser", "SearchUsers" -> "searchUsers"
func toCamelCasePrivate(s string) string {
	if s == "" {
		return s
	}
	// Convert first character to lowercase
	runes := []rune(s)
	runes[0] = []rune(strings.ToLower(string(runes[0])))[0]
	return string(runes)
}

// capitalize converts the first letter to uppercase (PascalCase)
// Example: "listTransactions" -> "ListTransactions", "getUser" -> "GetUser"
func capitalize(s string) string {
	if s == "" {
		return s
	}
	// Convert first character to uppercase
	runes := []rune(s)
	runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
	return string(runes)
}
