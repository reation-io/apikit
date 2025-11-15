package builder

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/spec"

	// Import all parsers to trigger auto-registration
	_ "github.com/reation-io/apikit/openapi/parsers/tags"
)

// Builder builds an OpenAPI specification from Go source files
type Builder struct {
	spec     *spec.OpenAPI
	fset     *token.FileSet
	patterns []string // File patterns to scan
}

// NewBuilder creates a new OpenAPI builder
func NewBuilder(patterns ...string) *Builder {
	if len(patterns) == 0 {
		patterns = []string{"**/*.go"}
	}

	return &Builder{
		spec: &spec.OpenAPI{
			OpenAPI: "3.0.3",
			Info: &spec.Info{
				Title:   "API",
				Version: "1.0.0",
			},
			Paths: &spec.Paths{
				PathItems: make(map[string]*spec.PathItem),
			},
		},
		fset:     token.NewFileSet(),
		patterns: patterns,
	}
}

// Build scans files and builds the OpenAPI specification
func (b *Builder) Build() (*spec.OpenAPI, error) {
	// Find all Go files matching patterns
	files, err := b.findFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to find files: %w", err)
	}

	// Parse each file
	for _, file := range files {
		if err := b.parseFile(file); err != nil {
			return nil, fmt.Errorf("failed to parse file %s: %w", file, err)
		}
	}

	return b.spec, nil
}

// findFiles finds all Go files matching the patterns
func (b *Builder) findFiles() ([]string, error) {
	var files []string
	for _, pattern := range b.patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
		files = append(files, matches...)
	}
	return files, nil
}

// parseFile parses a single Go file and extracts OpenAPI information
func (b *Builder) parseFile(filename string) error {
	// Parse the file
	file, err := parser.ParseFile(b.fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	// Look for swagger:meta comments
	if err := b.parseMeta(file); err != nil {
		return fmt.Errorf("failed to parse meta: %w", err)
	}

	// Look for swagger:route comments
	if err := b.parseRoutes(file); err != nil {
		return fmt.Errorf("failed to parse routes: %w", err)
	}

	// Look for swagger:model comments
	if err := b.parseModels(file); err != nil {
		return fmt.Errorf("failed to parse models: %w", err)
	}

	return nil
}

// parseMeta parses swagger:meta comments
func (b *Builder) parseMeta(file *ast.File) error {
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Doc == nil {
			continue
		}

		// Check if this is a swagger:meta comment
		if !hasDirective(genDecl.Doc, "swagger:meta") {
			continue
		}

		// Parse meta tags into Info (ignoring invalid target errors)
		if err := parsers.GlobalRegistry().Parse("swagger:meta", genDecl.Doc, b.spec.Info, parsers.ContextMeta); err != nil {
			if !isInvalidTargetError(err) {
				return err
			}
		}

		// Parse meta tags that target OpenAPI root (Consumes, Produces, SecuritySchemes, Servers)
		// Ignore invalid target errors since some parsers target Info, not OpenAPI
		if err := parsers.GlobalRegistry().Parse("swagger:meta", genDecl.Doc, b.spec, parsers.ContextMeta); err != nil {
			if !isInvalidTargetError(err) {
				return err
			}
		}
	}

	return nil
}

// parseRoutes parses swagger:route comments
func (b *Builder) parseRoutes(file *ast.File) error {
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Doc == nil {
			continue
		}

		// Check if this is a swagger:route comment
		if !hasDirective(genDecl.Doc, "swagger:route") {
			continue
		}

		// Parse the route line: swagger:route METHOD PATH TAG OPERATION_ID
		routeInfo, err := parseRouteLine(genDecl.Doc)
		if err != nil {
			return err
		}

		// Create operation
		operation := &spec.Operation{
			OperationID: routeInfo.OperationID,
			Tags:        []string{routeInfo.Tag},
			Responses: &spec.Responses{
				StatusCodeResponses: make(map[string]*spec.Response),
			},
		}

		// Parse operation tags
		if err := parsers.GlobalRegistry().Parse("swagger:route", genDecl.Doc, operation, parsers.ContextRoute); err != nil {
			if !isInvalidTargetError(err) {
				return err
			}
		}

		// Add operation to path
		if b.spec.Paths.PathItems[routeInfo.Path] == nil {
			b.spec.Paths.PathItems[routeInfo.Path] = &spec.PathItem{}
		}

		pathItem := b.spec.Paths.PathItems[routeInfo.Path]
		switch strings.ToUpper(routeInfo.Method) {
		case "GET":
			pathItem.Get = operation
		case "POST":
			pathItem.Post = operation
		case "PUT":
			pathItem.Put = operation
		case "DELETE":
			pathItem.Delete = operation
		case "PATCH":
			pathItem.Patch = operation
		case "OPTIONS":
			pathItem.Options = operation
		case "HEAD":
			pathItem.Head = operation
		}
	}

	return nil
}

// parseModels parses swagger:model comments
func (b *Builder) parseModels(file *ast.File) error {
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Doc == nil {
			continue
		}

		// Check if this is a swagger:model comment
		if !hasDirective(genDecl.Doc, "swagger:model") {
			continue
		}

		// Find the type spec
		for _, s := range genDecl.Specs {
			typeSpec, ok := s.(*ast.TypeSpec)
			if !ok {
				continue
			}

			// Parse struct type
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			// Create schema
			schema := b.parseStruct(structType)

			// Initialize Components if needed
			if b.spec.Components == nil {
				b.spec.Components = &spec.Components{}
			}
			if b.spec.Components.Schemas == nil {
				b.spec.Components.Schemas = make(map[string]*spec.Schema)
			}

			// Add schema to components
			b.spec.Components.Schemas[typeSpec.Name.Name] = schema
		}
	}

	return nil
}

// parseStruct parses a struct type into a schema
func (b *Builder) parseStruct(structType *ast.StructType) *spec.Schema {
	schema := &spec.Schema{
		Type:       "object",
		Properties: make(map[string]*spec.Schema),
	}

	for _, field := range structType.Fields.List {
		// Skip fields without names (embedded structs)
		if len(field.Names) == 0 {
			continue
		}

		// Create field schema
		fieldSchema := b.parseFieldType(field.Type)

		// Parse field tags (Description, Example, Format, etc.)
		if field.Doc != nil {
			if err := parsers.GlobalRegistry().Parse("swagger:model", field.Doc, fieldSchema, parsers.ContextField); err != nil {
				// Ignore errors for now
				_ = err
			}
		}

		// Get JSON tag name
		jsonName := b.getJSONName(field)
		if jsonName == "" || jsonName == "-" {
			continue
		}

		schema.Properties[jsonName] = fieldSchema
	}

	return schema
}

// parseFieldType parses a field type into a schema type
func (b *Builder) parseFieldType(expr ast.Expr) *spec.Schema {
	schema := &spec.Schema{}

	switch t := expr.(type) {
	case *ast.Ident:
		// Basic types
		schema.Type = goTypeToJSONType(t.Name)
	case *ast.ArrayType:
		schema.Type = "array"
		schema.Items = b.parseFieldType(t.Elt)
	case *ast.StarExpr:
		// Pointer type
		return b.parseFieldType(t.X)
	case *ast.SelectorExpr:
		// External type (e.g., time.Time)
		if ident, ok := t.X.(*ast.Ident); ok {
			if ident.Name == "time" && t.Sel.Name == "Time" {
				schema.Type = "string"
				schema.Format = "date-time"
			}
		}
	}

	return schema
}

// goTypeToJSONType converts Go types to JSON Schema types
func goTypeToJSONType(goType string) string {
	switch goType {
	case "string":
		return "string"
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		return "integer"
	case "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	default:
		return "object"
	}
}

// getJSONName extracts the JSON name from struct tags
func (b *Builder) getJSONName(field *ast.Field) string {
	if field.Tag == nil {
		return ""
	}

	tag := field.Tag.Value
	tag = strings.Trim(tag, "`")

	// Parse json tag
	for _, part := range strings.Fields(tag) {
		if strings.HasPrefix(part, "json:") {
			jsonTag := strings.TrimPrefix(part, "json:")
			jsonTag = strings.Trim(jsonTag, `"`)
			parts := strings.Split(jsonTag, ",")
			if len(parts) > 0 {
				return parts[0]
			}
		}
	}

	return ""
}

// BuildMultiple scans files and builds multiple OpenAPI specifications based on Spec: tags
// Returns a map of spec name to OpenAPI spec
func (b *Builder) BuildMultiple() (map[string]*spec.OpenAPI, error) {
	// First, build the complete spec with all routes
	if _, err := b.Build(); err != nil {
		return nil, err
	}

	// Distribute routes into multiple specs
	return b.distributeRoutes(), nil
}

// distributeRoutes distributes routes from the main spec into multiple specs based on x-specs extension
func (b *Builder) distributeRoutes() map[string]*spec.OpenAPI {
	specs := make(map[string]*spec.OpenAPI)

	// Initialize default spec
	specs["default"] = b.createEmptySpec()

	// Iterate through all paths and operations
	for path, pathItem := range b.spec.Paths.PathItems {
		// Check each HTTP method
		operations := map[string]*spec.Operation{
			"GET":     pathItem.Get,
			"POST":    pathItem.Post,
			"PUT":     pathItem.Put,
			"DELETE":  pathItem.Delete,
			"PATCH":   pathItem.Patch,
			"OPTIONS": pathItem.Options,
			"HEAD":    pathItem.Head,
		}

		for method, operation := range operations {
			if operation == nil {
				continue
			}

			// Get spec names from operation extensions
			specNames := getSpecNamesFromOperation(operation)

			if len(specNames) == 0 {
				// No spec tag → goes to default
				b.addOperationToSpec(specs["default"], path, method, operation)
			} else {
				// Add to each specified spec
				for _, specName := range specNames {
					if specs[specName] == nil {
						specs[specName] = b.createEmptySpec()
					}
					b.addOperationToSpec(specs[specName], path, method, operation)
				}
			}
		}
	}

	// Copy models to all specs (models are shared)
	for specName, targetSpec := range specs {
		if b.spec.Components != nil && b.spec.Components.Schemas != nil {
			if targetSpec.Components == nil {
				targetSpec.Components = &spec.Components{}
			}
			if targetSpec.Components.Schemas == nil {
				targetSpec.Components.Schemas = make(map[string]*spec.Schema)
			}
			// Copy all schemas
			for schemaName, schema := range b.spec.Components.Schemas {
				targetSpec.Components.Schemas[schemaName] = schema
			}
		}

		// Copy security schemes to all specs
		if b.spec.Components != nil && b.spec.Components.SecuritySchemes != nil {
			if targetSpec.Components == nil {
				targetSpec.Components = &spec.Components{}
			}
			if targetSpec.Components.SecuritySchemes == nil {
				targetSpec.Components.SecuritySchemes = make(map[string]*spec.SecurityScheme)
			}
			for schemeName, scheme := range b.spec.Components.SecuritySchemes {
				targetSpec.Components.SecuritySchemes[schemeName] = scheme
			}
		}

		// Apply meta-level spec filtering if Info has x-specs
		if b.spec.Info != nil && b.spec.Info.Extensions != nil {
			if metaSpecs, ok := b.spec.Info.Extensions["x-specs"].([]string); ok {
				// Check if this spec should get the meta info
				if !containsString(metaSpecs, specName) && specName != "default" {
					// This spec shouldn't use this meta, create default meta
					targetSpec.Info = &spec.Info{
						Title:   "API",
						Version: "1.0.0",
					}
				}
			}
		}
	}

	return specs
}

// createEmptySpec creates a new empty OpenAPI spec with default values
func (b *Builder) createEmptySpec() *spec.OpenAPI {
	newSpec := &spec.OpenAPI{
		OpenAPI: "3.0.3",
		Info:    &spec.Info{},
		Paths: &spec.Paths{
			PathItems: make(map[string]*spec.PathItem),
		},
	}

	// Copy Info from main spec
	if b.spec.Info != nil {
		newSpec.Info.Title = b.spec.Info.Title
		newSpec.Info.Version = b.spec.Info.Version
		newSpec.Info.Description = b.spec.Info.Description
		newSpec.Info.TermsOfService = b.spec.Info.TermsOfService
		newSpec.Info.Contact = b.spec.Info.Contact
		newSpec.Info.License = b.spec.Info.License
	}

	// Copy Servers from main spec
	if b.spec.Servers != nil {
		newSpec.Servers = make([]*spec.Server, len(b.spec.Servers))
		copy(newSpec.Servers, b.spec.Servers)
	}

	return newSpec
}

// addOperationToSpec adds an operation to a spec at the given path and method
func (b *Builder) addOperationToSpec(targetSpec *spec.OpenAPI, path, method string, operation *spec.Operation) {
	// Ensure path exists
	if targetSpec.Paths.PathItems[path] == nil {
		targetSpec.Paths.PathItems[path] = &spec.PathItem{}
	}

	pathItem := targetSpec.Paths.PathItems[path]

	// Clone the operation to avoid sharing references
	clonedOp := cloneOperation(operation)

	// Add operation to the appropriate method
	switch strings.ToUpper(method) {
	case "GET":
		pathItem.Get = clonedOp
	case "POST":
		pathItem.Post = clonedOp
	case "PUT":
		pathItem.Put = clonedOp
	case "DELETE":
		pathItem.Delete = clonedOp
	case "PATCH":
		pathItem.Patch = clonedOp
	case "OPTIONS":
		pathItem.Options = clonedOp
	case "HEAD":
		pathItem.Head = clonedOp
	}
}

// getSpecNamesFromOperation extracts spec names from operation's x-specs extension
func getSpecNamesFromOperation(operation *spec.Operation) []string {
	if operation.Extensions == nil {
		return nil
	}

	specs, ok := operation.Extensions["x-specs"]
	if !ok {
		return nil
	}

	// Handle both []string and []interface{} (from JSON unmarshaling)
	switch v := specs.(type) {
	case []string:
		return v
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	default:
		return nil
	}
}

// cloneOperation creates a deep copy of an operation
func cloneOperation(op *spec.Operation) *spec.Operation {
	if op == nil {
		return nil
	}

	cloned := &spec.Operation{
		Tags:        make([]string, len(op.Tags)),
		Summary:     op.Summary,
		Description: op.Description,
		OperationID: op.OperationID,
		Deprecated:  op.Deprecated,
	}

	copy(cloned.Tags, op.Tags)

	// Clone parameters
	if op.Parameters != nil {
		cloned.Parameters = make([]*spec.Parameter, len(op.Parameters))
		copy(cloned.Parameters, op.Parameters)
	}

	// Clone request body
	cloned.RequestBody = op.RequestBody

	// Clone responses
	if op.Responses != nil {
		cloned.Responses = &spec.Responses{
			StatusCodeResponses: make(map[string]*spec.Response),
			Default:             op.Responses.Default,
		}
		for code, resp := range op.Responses.StatusCodeResponses {
			cloned.Responses.StatusCodeResponses[code] = resp
		}
	}

	// Clone security
	if op.Security != nil {
		cloned.Security = make([]spec.SecurityRequirement, len(op.Security))
		copy(cloned.Security, op.Security)
	}

	// Clone servers
	if op.Servers != nil {
		cloned.Servers = make([]*spec.Server, len(op.Servers))
		copy(cloned.Servers, op.Servers)
	}

	// Don't copy Extensions (we don't want x-specs in the output)
	// cloned.Extensions = op.Extensions

	return cloned
}

// containsString checks if a string slice contains a specific string
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
