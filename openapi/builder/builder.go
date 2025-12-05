package builder

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	constants "github.com/reation-io/apikit/openapi"
	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/scanner"
	"github.com/reation-io/apikit/openapi/spec"
	"github.com/reation-io/apikit/openapi/validator"

	// Import all parsers to trigger auto-registration
	_ "github.com/reation-io/apikit/openapi/parsers/tags"
)

// Builder builds an OpenAPI specification from Go source files
type Builder struct {
	spec             *spec.OpenAPI
	document         *spec.OpenAPI // alias for spec (used by embedded.go)
	fset             *token.FileSet
	config           *BuilderConfig
	files            map[string]*ast.File // Cached files from scanner
	enumRegistry     *spec.EnumRegistry   // Registry of discovered enums
	operations       map[string]*spec.Operation
	pendingParams    map[string][]ParameterGroup
	modelRegistry    map[string]*ast.StructType // Map of model name to struct type
	parsedModels     map[string]bool            // Set of parsed models
	processingModels map[string]bool            // Set of models currently being parsed (cycle detection)
	typeRegistry     map[string]*ast.StructType // Map of ALL struct types (for untagged embedding)
}

// ParameterGroup holds parameters and body to be attached to an operation
type ParameterGroup struct {
	Parameters  []*spec.Parameter
	RequestBody *spec.RequestBody
}

// NewBuilder creates a new OpenAPI builder
// For more advanced configuration, use NewBuilderWithOptions
func NewBuilder(patterns ...string) *Builder {
	var opts []Option
	if len(patterns) > 0 {
		opts = append(opts, WithPattern(patterns[0]))
	} else {
		opts = append(opts, WithPattern("./..."))
	}
	return NewBuilderWithOptions(opts...)
}

// NewBuilderWithOptions creates a new OpenAPI builder with functional options
// Example:
//
//	builder := NewBuilderWithOptions(
//	    WithPattern("./..."),
//	    WithDir("."),
//	    WithIgnorePaths("vendor/**"),
//	    WithValidation(true),
//	)
func NewBuilderWithOptions(opts ...Option) *Builder {
	config := &BuilderConfig{
		ScannerConfig: &scanner.Config{},
	}

	for _, opt := range opts {
		opt(config)
	}

	// Ensure default pattern if not set
	if config.ScannerConfig.Pattern == "" {
		config.ScannerConfig.Pattern = "./..."
	}

	openAPISpec := &spec.OpenAPI{
		OpenAPI: "3.0.3",
		Info: &spec.Info{
			Title:   "API",
			Version: "1.0.0",
		},
		Paths: &spec.Paths{
			PathItems: make(map[string]*spec.PathItem),
		},
	}
	return &Builder{
		spec:             openAPISpec,
		document:         openAPISpec,
		fset:             token.NewFileSet(),
		config:           config,
		files:            make(map[string]*ast.File),
		enumRegistry:     spec.NewEnumRegistry(),
		operations:       make(map[string]*spec.Operation),
		pendingParams:    make(map[string][]ParameterGroup),
		modelRegistry:    make(map[string]*ast.StructType),
		parsedModels:     make(map[string]bool),
		processingModels: make(map[string]bool),
		typeRegistry:     make(map[string]*ast.StructType),
	}
}

// Build scans files and builds the OpenAPI specification
func (b *Builder) Build() (*spec.OpenAPI, error) {
	s := scanner.NewWithConfig(b.config.ScannerConfig)

	files, fset, err := s.ScanFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to scan packages: %w", err)
	}

	b.fset = fset
	b.files = files

	// Pass 1: Discovery (Enums and Models)
	for _, file := range files {
		// Look for swagger:enum comments
		if err := b.parseEnums(file); err != nil {
			return nil, fmt.Errorf("failed to parse enums: %w", err)
		}
		// Register models
		if err := b.registerModels(file); err != nil {
			return nil, fmt.Errorf("failed to register models: %w", err)
		}
	}

	// Pass 2: Resolution (Parse Models with dependency handling)
	for name := range b.modelRegistry {
		if err := b.ensureModelParsed(name); err != nil {
			return nil, fmt.Errorf("failed to parse model %s: %w", name, err)
		}
	}

	// Pass 3: Routes and others
	for filePath, file := range files {
		// Look for swagger:parameters comments
		if err := b.parseParameters(file); err != nil {
			return nil, fmt.Errorf("failed to parse parameters in %s: %w", filePath, err)
		}

		// Look for swagger:meta comments
		if err := b.parseMeta(file); err != nil {
			return nil, fmt.Errorf("failed to parse meta in %s: %w", filePath, err)
		}

		// Look for swagger:route comments
		if err := b.parseRoutes(file); err != nil {
			return nil, fmt.Errorf("failed to parse routes in %s: %w", filePath, err)
		}
	}

	// Resolve pending parameters
	b.resolvePendingParameters()

	// Validate if enabled
	if b.config.Validation {
		if err := b.validate(); err != nil {
			return nil, err
		}
	}

	return b.spec, nil
}

// Validate validates the built spec and returns validation errors
func (b *Builder) Validate() *validator.ValidationResult {
	v := validator.New(b.spec)
	return v.Validate()
}

// validate runs validation and returns an error if there are critical issues
func (b *Builder) validate() error {
	result := b.Validate()
	if result.HasErrors() {
		// Return first error as the main error
		return fmt.Errorf("validation failed: %s", result.Errors[0].Error())
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
		if !hasDirective(genDecl.Doc, constants.DirectiveMeta) {
			continue
		}

		// Parse meta tags into Info (ignoring invalid target errors)
		if err := parsers.GlobalRegistry().Parse(constants.DirectiveMeta, genDecl.Doc, b.spec.Info, parsers.ContextMeta); err != nil {
			if !isInvalidTargetError(err) {
				return err
			}
		}

		// Parse meta tags that target OpenAPI root (Consumes, Produces, SecuritySchemes, Servers)
		// Ignore invalid target errors since some parsers target Info, not OpenAPI
		if err := parsers.GlobalRegistry().Parse(constants.DirectiveMeta, genDecl.Doc, b.spec, parsers.ContextMeta); err != nil {
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
		if !hasDirective(genDecl.Doc, constants.DirectiveRoute) {
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
		if err := parsers.GlobalRegistry().Parse(constants.DirectiveRoute, genDecl.Doc, operation, parsers.ContextRoute); err != nil {
			if !isInvalidTargetError(err) {
				return err
			}
		}

		// Find and process the struct type to extract parameters and request body
		for _, s := range genDecl.Specs {
			typeSpec, ok := s.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			// Parse parameters and request body from struct fields
			b.parseRequestStruct(structType, operation)
		}

		// Add operation to path
		if b.spec.Paths.PathItems[routeInfo.Path] == nil {
			b.spec.Paths.PathItems[routeInfo.Path] = &spec.PathItem{}
		}

		pathItem := b.spec.Paths.PathItems[routeInfo.Path]
		switch strings.ToUpper(routeInfo.Method) {
		case constants.MethodGET:
			pathItem.Get = operation
		case constants.MethodPOST:
			pathItem.Post = operation
		case constants.MethodPUT:
			pathItem.Put = operation
		case constants.MethodDELETE:
			pathItem.Delete = operation
		case constants.MethodPATCH:
			pathItem.Patch = operation
		case constants.MethodOPTIONS:
			pathItem.Options = operation
		case constants.MethodHEAD:
			pathItem.Head = operation
		}

		// Register operation
		if operation.OperationID != "" {
			b.operations[operation.OperationID] = operation
		}
	}

	return nil
}

// registerModels finds swagger:model directives and registers them (Pass 1)
func (b *Builder) registerModels(file *ast.File) error {
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

			// Initialize Components if needed
			if b.spec.Components == nil {
				b.spec.Components = &spec.Components{}
			}
			if b.spec.Components.Schemas == nil {
				b.spec.Components.Schemas = make(map[string]*spec.Schema)
			}

			// Register the model name first (with placeholder)
			b.spec.Components.Schemas[typeSpec.Name.Name] = &spec.Schema{Type: "object"}

			// Register for second pass
			if structType, ok := typeSpec.Type.(*ast.StructType); ok {
				b.modelRegistry[typeSpec.Name.Name] = structType
			}
		}
	}

	// Collect ALL struct types for potential embedding
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, s := range genDecl.Specs {
			typeSpec, ok := s.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if structType, ok := typeSpec.Type.(*ast.StructType); ok {
				b.typeRegistry[typeSpec.Name.Name] = structType
			}
		}
	}
	return nil
}

// ensureModelParsed ensures a model is fully parsed, handling recursion (Pass 2)
func (b *Builder) ensureModelParsed(name string) error {
	// Check if already parsed
	if b.parsedModels[name] {
		return nil
	}

	// Check if we are already parsing it (cycle detection)
	if b.processingModels[name] {
		return nil
	}

	// Get struct type
	structType, ok := b.modelRegistry[name]
	if !ok {
		// Model not found in registry (should not happen if registered correctly)
		return nil
	}

	// Mark as processing
	b.processingModels[name] = true
	defer func() {
		b.processingModels[name] = false
	}()

	// Parse schema
	schema := b.parseStruct(structType)

	// Update schema in components (in-place update relative to the map)
	b.spec.Components.Schemas[name] = schema

	// Mark as parsed
	b.parsedModels[name] = true

	return nil
}

// parseStruct parses a struct type into a schema
func (b *Builder) parseStruct(structType *ast.StructType) *spec.Schema {
	schema := &spec.Schema{
		Type:       "object",
		Properties: make(map[string]*spec.Schema),
	}

	var requiredFields []string

	for _, field := range structType.Fields.List {
		// Check for swagger:ignore directive
		if hasSwaggerIgnore(field.Doc) || hasSwaggerIgnore(field.Comment) {
			continue
		}

		// Handle embedded fields (no name)
		if len(field.Names) == 0 {
			embeddedSchema := b.resolveEmbeddedSchema(field.Type)
			if embeddedSchema != nil && embeddedSchema.Properties != nil {
				// Merge embedded schema properties
				for propName, propSchema := range embeddedSchema.Properties {
					if _, exists := schema.Properties[propName]; !exists {
						schema.Properties[propName] = cloneSchema(propSchema)
					}
				}
				// Merge embedded required fields
				for _, req := range embeddedSchema.Required {
					if !containsString(requiredFields, req) {
						requiredFields = append(requiredFields, req)
					}
				}
			}
			continue
		}

		// Create field schema
		fieldSchema := b.parseFieldType(field.Type)

		// Check for nullable (pointer types)
		if _, ok := field.Type.(*ast.StarExpr); ok {
			fieldSchema.Nullable = true
		}

		// Check if field is required from comments or struct tags
		isRequired := false
		hasOmitempty := false

		// Check struct tags for omitempty
		if field.Tag != nil {
			tagStr := strings.Trim(field.Tag.Value, "`")
			if strings.Contains(tagStr, "omitempty") || strings.Contains(tagStr, "omitzero") {
				hasOmitempty = true
			}
		}

		if field.Doc != nil {
			for _, comment := range field.Doc.List {
				text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
				lower := strings.ToLower(text)
				if strings.HasPrefix(lower, "required:") {
					val := strings.TrimSpace(text[9:])
					isRequired = val == "true" || val == "yes"
				}
			}

			// Parse field tags (Description, Example, Format, etc.)
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

		// Track required fields (required explicitly OR non-pointer without omitempty)
		if isRequired || (!hasOmitempty && !fieldSchema.Nullable) {
			requiredFields = append(requiredFields, jsonName)
		}
	}

	// Set required array if there are required fields
	if len(requiredFields) > 0 {
		schema.Required = requiredFields
	}

	return schema
}

// resolveEmbeddedSchema resolves an embedded type to its schema
func (b *Builder) resolveEmbeddedSchema(expr ast.Expr) *spec.Schema {
	// Handle pointer types
	if starExpr, ok := expr.(*ast.StarExpr); ok {
		return b.resolveEmbeddedSchema(starExpr.X)
	}

	// Get the type name
	var typeName string
	switch t := expr.(type) {
	case *ast.Ident:
		typeName = t.Name
	case *ast.SelectorExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			typeName = ident.Name + "." + t.Sel.Name
		}
	}

	if typeName == "" {
		return nil
	}

	// Look up the schema in components
	if b.spec.Components != nil && b.spec.Components.Schemas != nil {
		if schema, ok := b.spec.Components.Schemas[typeName]; ok {
			// Trigger parsing if needed
			if !b.parsedModels[typeName] && b.modelRegistry[typeName] != nil {
				_ = b.ensureModelParsed(typeName)
				return b.spec.Components.Schemas[typeName]
			}
			return schema
		}
	}

	// Fallback: Check type registry for untagged structs
	// These are NOT standalone models, so we parse them on-the-fly to get their properties
	if structType, ok := b.typeRegistry[typeName]; ok {
		// Avoid cycles for untagged structs too (primitive check)
		// Since we parse a NEW schema every time, there's no caching.
		// FIXME: Deep nesting might be inefficient but correct for "Yellow->Green" fix.
		return b.parseStruct(structType)
	}

	return nil
}

// parseFieldType parses a field type into a schema type with format
func (b *Builder) parseFieldType(expr ast.Expr) *spec.Schema {
	schema := &spec.Schema{}

	switch t := expr.(type) {
	case *ast.Ident:
		// Check if this type is an enum
		if enumInfo := b.enumRegistry.GetByTypeName(t.Name); enumInfo != nil {
			return b.createEnumSchema(enumInfo)
		}

		// Check if this is a known model type (use $ref)
		if b.isModelType(t.Name) {
			return &spec.Schema{
				Ref: "#/components/schemas/" + t.Name,
			}
		}

		// Basic types with format
		mapping := MapGoTypeToOpenAPI(t.Name)
		schema.Type = mapping.Type
		schema.Format = mapping.Format

	case *ast.ArrayType:
		schema.Type = "array"
		schema.Items = b.parseFieldType(t.Elt)

	case *ast.StarExpr:
		// Pointer type
		return b.parseFieldType(t.X)

	case *ast.SelectorExpr:
		// Check if this qualified type is an enum (e.g., pkg.EnumType)
		if enumInfo := b.enumRegistry.GetByTypeName(t.Sel.Name); enumInfo != nil {
			return b.createEnumSchema(enumInfo)
		}

		// Check if this is a known model type (use $ref)
		if b.isModelType(t.Sel.Name) {
			return &spec.Schema{
				Ref: "#/components/schemas/" + t.Sel.Name,
			}
		}

		// Check for registered type processors (e.g. time.Time, uuid.UUID)
		tempInfo := spec.NewFieldInfo()
		if ProcessTypeWithProcessors(tempInfo, expr) {
			schema.Type = tempInfo.Type
			if format, ok := tempInfo.Validations["format"]; ok {
				schema.Format = format
			}
			return schema
		}

	case *ast.MapType:
		schema.Type = "object"
		if t.Value != nil {
			schema.AdditionalProperties = b.parseFieldType(t.Value)
		}

	case *ast.StructType:
		// Inline struct - parse it directly
		schema = b.parseStruct(t)
	}

	return schema
}

// isModelType checks if a type name is a registered model
func (b *Builder) isModelType(typeName string) bool {
	if b.spec.Components == nil || b.spec.Components.Schemas == nil {
		return false
	}
	_, exists := b.spec.Components.Schemas[typeName]
	return exists
}

// createEnumSchema creates a schema for an enum type with inline values
func (b *Builder) createEnumSchema(enumInfo *spec.EnumInfo) *spec.Schema {
	schema := &spec.Schema{
		Type:        enumInfo.BaseType,
		Description: enumInfo.Description,
	}

	// Add enum values
	if len(enumInfo.Values) > 0 {
		schema.Enum = enumInfo.GetEnumValues()
	}

	// Add example if available
	if enumInfo.Example != nil {
		schema.Example = enumInfo.Example
	}

	return schema
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

	// Handle both []string and []any (from JSON unmarshaling)
	switch v := specs.(type) {
	case []string:
		return v
	case []any:
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

// cloneSchema creates a deep copy of a schema
func cloneSchema(s *spec.Schema) *spec.Schema {
	if s == nil {
		return nil
	}

	cloned := *s // Shallow copy struct

	// Clone slices
	if s.Required != nil {
		cloned.Required = make([]string, len(s.Required))
		copy(cloned.Required, s.Required)
	}
	if s.Enum != nil {
		cloned.Enum = make([]any, len(s.Enum))
		copy(cloned.Enum, s.Enum)
	}

	// Clone maps (and recursive schemas)
	if s.Properties != nil {
		cloned.Properties = make(map[string]*spec.Schema, len(s.Properties))
		for k, v := range s.Properties {
			cloned.Properties[k] = cloneSchema(v)
		}
	}

	// Clone nested schemas
	if s.Items != nil {
		cloned.Items = cloneSchema(s.Items)
	}
	if s.AllOf != nil {
		cloned.AllOf = make([]*spec.Schema, len(s.AllOf))
		for i, v := range s.AllOf {
			cloned.AllOf[i] = cloneSchema(v)
		}
	}
	if s.OneOf != nil {
		cloned.OneOf = make([]*spec.Schema, len(s.OneOf))
		for i, v := range s.OneOf {
			cloned.OneOf[i] = cloneSchema(v)
		}
	}
	if s.AnyOf != nil {
		cloned.AnyOf = make([]*spec.Schema, len(s.AnyOf))
		for i, v := range s.AnyOf {
			cloned.AnyOf[i] = cloneSchema(v)
		}
	}
	if s.Not != nil {
		cloned.Not = cloneSchema(s.Not)
	}

	// Clone AdditionalProperties if it's a schema
	if s.AdditionalProperties != nil {
		if schema, ok := s.AdditionalProperties.(*spec.Schema); ok {
			cloned.AdditionalProperties = cloneSchema(schema)
		}
	}

	// Clone pointers to primitives
	if s.MultipleOf != nil {
		v := *s.MultipleOf
		cloned.MultipleOf = &v
	}
	if s.Maximum != nil {
		v := *s.Maximum
		cloned.Maximum = &v
	}
	if s.Minimum != nil {
		v := *s.Minimum
		cloned.Minimum = &v
	}
	if s.MaxLength != nil {
		v := *s.MaxLength
		cloned.MaxLength = &v
	}
	if s.MinLength != nil {
		v := *s.MinLength
		cloned.MinLength = &v
	}
	if s.MaxItems != nil {
		v := *s.MaxItems
		cloned.MaxItems = &v
	}
	if s.MinItems != nil {
		v := *s.MinItems
		cloned.MinItems = &v
	}
	if s.MaxProperties != nil {
		v := *s.MaxProperties
		cloned.MaxProperties = &v
	}
	if s.MinProperties != nil {
		v := *s.MinProperties
		cloned.MinProperties = &v
	}
	if s.XML != nil {
		v := *s.XML
		cloned.XML = &v
	}

	return &cloned
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

// parseParameters parses swagger:parameters comments
func (b *Builder) parseParameters(file *ast.File) error {
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Doc == nil {
			continue
		}

		// Check for swagger:parameters
		if !hasDirective(genDecl.Doc, constants.DirectiveParameters) {
			continue
		}

		// Extract Operation IDs
		var opIDs []string
		for _, comment := range genDecl.Doc.List {
			text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
			if strings.HasPrefix(text, constants.DirectiveParameters) {
				parts := strings.Fields(text)
				if len(parts) > 1 {
					opIDs = append(opIDs, parts[1:]...)
				}
			}
		}

		if len(opIDs) == 0 {
			continue
		}

		// Find and process the struct type
		for _, s := range genDecl.Specs {
			typeSpec, ok := s.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			params, body := b.extractParameters(structType, nil)

			group := ParameterGroup{
				Parameters:  params,
				RequestBody: body,
			}

			for _, opID := range opIDs {
				b.pendingParams[opID] = append(b.pendingParams[opID], group)
			}
		}
	}
	return nil
}

// resolvePendingParameters attaches pending parameters to operations
func (b *Builder) resolvePendingParameters() {
	for opID, groups := range b.pendingParams {
		operation, ok := b.operations[opID]
		if !ok {
			// Operation not found (maybe defined in another package not scanned, or typo)
			continue
		}

		for _, group := range groups {
			// Append parameters
			if len(group.Parameters) > 0 {
				operation.Parameters = append(operation.Parameters, group.Parameters...)
			}

			// Merge request body
			if group.RequestBody != nil {
				if operation.RequestBody == nil {
					operation.RequestBody = group.RequestBody
				} else {
					// Merge content
					if operation.RequestBody.Description == "" {
						operation.RequestBody.Description = group.RequestBody.Description
					}
					if operation.RequestBody.Content == nil {
						operation.RequestBody.Content = make(map[string]*spec.MediaType)
					}
					for k, v := range group.RequestBody.Content {
						operation.RequestBody.Content[k] = v
					}
				}
			}
		}
	}
}
