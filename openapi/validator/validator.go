package validator

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/reation-io/apikit/openapi/spec"
)

// Level represents the severity of a validation error
type Level int

const (
	// LevelError indicates a critical validation error
	LevelError Level = iota
	// LevelWarning indicates a non-critical validation issue
	LevelWarning
)

func (l Level) String() string {
	switch l {
	case LevelError:
		return "error"
	case LevelWarning:
		return "warning"
	default:
		return "unknown"
	}
}

// ValidationError represents a validation error with context
type ValidationError struct {
	Path    string // e.g., "paths./users.get.responses.200"
	Message string
	Level   Level
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Level, e.Path, e.Message)
}

// ValidationResult holds the complete validation results
type ValidationResult struct {
	Errors   []ValidationError
	Warnings []ValidationError
}

// HasErrors returns true if there are any errors
func (r *ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// HasWarnings returns true if there are any warnings
func (r *ValidationResult) HasWarnings() bool {
	return len(r.Warnings) > 0
}

// IsValid returns true if there are no errors
func (r *ValidationResult) IsValid() bool {
	return !r.HasErrors()
}

// Validator validates OpenAPI specifications
type Validator struct {
	spec         *spec.OpenAPI
	result       *ValidationResult
	operationIDs map[string]string // map[id]path
}

// New creates a new validator for the given spec
func New(s *spec.OpenAPI) *Validator {
	return &Validator{
		spec: s,
		result: &ValidationResult{
			Errors:   []ValidationError{},
			Warnings: []ValidationError{},
		},
		operationIDs: make(map[string]string),
	}
}

// Validate performs comprehensive validation of the OpenAPI spec
func (v *Validator) Validate() *ValidationResult {
	v.result = &ValidationResult{
		Errors:   []ValidationError{},
		Warnings: []ValidationError{},
	}
	v.operationIDs = make(map[string]string)

	// Validate basic structure
	v.validateBasicStructure()

	// Validate info section
	v.validateInfo()

	// Validate components
	if v.spec.Components != nil {
		v.validateComponents()
	}

	// Validate paths
	if v.spec.Paths != nil {
		v.validatePaths()
	}

	// Validate security schemes
	v.validateSecuritySchemes()

	return v.result
}

// validateBasicStructure validates the basic structure of the OpenAPI spec
func (v *Validator) validateBasicStructure() {
	if v.spec.OpenAPI == "" {
		v.addError("", "OpenAPI version is required")
	} else if !strings.HasPrefix(v.spec.OpenAPI, "3.0") && !strings.HasPrefix(v.spec.OpenAPI, "3.1") {
		v.addError("", fmt.Sprintf("Unsupported OpenAPI version: %s", v.spec.OpenAPI))
	}

	if v.spec.Info == nil {
		v.addError("", "Info object is required")
	}

	if v.spec.Paths == nil || len(v.spec.Paths.PathItems) == 0 {
		v.addWarning("", "No paths defined in OpenAPI spec")
	}
}

// validateInfo validates the info section
func (v *Validator) validateInfo() {
	if v.spec.Info == nil {
		return
	}

	if v.spec.Info.Title == "" {
		v.addError("info", "Title is required")
	}

	if v.spec.Info.Version == "" {
		v.addError("info", "Version is required")
	}

	// Validate contact
	if v.spec.Info.Contact != nil {
		v.validateContact(v.spec.Info.Contact)
	}

	// Validate license
	if v.spec.Info.License != nil {
		v.validateLicense(v.spec.Info.License)
	}
}

// validateContact validates contact information
func (v *Validator) validateContact(contact *spec.Contact) {
	if contact.Email != "" && !isValidEmail(contact.Email) {
		v.addWarning("info.contact.email", fmt.Sprintf("Invalid email format: %s", contact.Email))
	}

	if contact.URL != "" && !isValidURL(contact.URL) {
		v.addWarning("info.contact.url", fmt.Sprintf("Invalid URL format: %s", contact.URL))
	}
}

// validateLicense validates license information
func (v *Validator) validateLicense(license *spec.License) {
	if license.Name == "" {
		v.addError("info.license", "License name is required")
	}

	if license.URL != "" && !isValidURL(license.URL) {
		v.addWarning("info.license.url", fmt.Sprintf("Invalid URL format: %s", license.URL))
	}
}

// validateComponents validates the components section
func (v *Validator) validateComponents() {
	// Validate schemas
	for name, schema := range v.spec.Components.Schemas {
		v.validateSchema(fmt.Sprintf("components.schemas.%s", name), schema)
	}

	// Validate security schemes
	for name, scheme := range v.spec.Components.SecuritySchemes {
		v.validateSecurityScheme(fmt.Sprintf("components.securitySchemes.%s", name), scheme)
	}
}

// validateSchema validates a schema object
func (v *Validator) validateSchema(path string, schema *spec.Schema) {
	if schema == nil {
		v.addError(path, "Schema cannot be null")
		return
	}

	// Skip reference schemas
	if schema.Ref != "" {
		return
	}

	// Validate type
	validTypes := []string{"string", "number", "integer", "boolean", "array", "object", ""}
	if schema.Type != "" && !contains(validTypes, schema.Type) {
		v.addError(path+".type", fmt.Sprintf("Invalid type: %s", schema.Type))
	}

	// Validate array items
	if schema.Type == "array" && schema.Items == nil {
		v.addError(path, "Array schema must have items definition")
	}

	// Validate numeric constraints
	if schema.Type == "number" || schema.Type == "integer" {
		if schema.Minimum != nil && schema.Maximum != nil && *schema.Minimum > *schema.Maximum {
			v.addError(path, "Minimum value cannot be greater than maximum value")
		}
	}

	// Validate string constraints
	if schema.Type == "string" {
		if schema.MinLength != nil && schema.MaxLength != nil && *schema.MinLength > *schema.MaxLength {
			v.addError(path, "MinLength cannot be greater than maxLength")
		}

		if schema.Pattern != "" {
			if _, err := regexp.Compile(schema.Pattern); err != nil {
				v.addError(path+".pattern", fmt.Sprintf("Invalid regex pattern: %s", err))
			}
		}
	}

	// Validate properties for object type
	if schema.Type == "object" || len(schema.Properties) > 0 {
		for propName, propSchema := range schema.Properties {
			v.validateSchema(fmt.Sprintf("%s.properties.%s", path, propName), propSchema)
		}
	}

	// Validate composition
	v.validateComposition(path, schema)
}

// validateComposition validates schema composition (oneOf, allOf, anyOf)
func (v *Validator) validateComposition(path string, schema *spec.Schema) {
	compositionCount := 0

	if len(schema.OneOf) > 0 {
		compositionCount++
		for i, subSchema := range schema.OneOf {
			v.validateSchema(fmt.Sprintf("%s.oneOf[%d]", path, i), subSchema)
		}
	}
	if len(schema.AllOf) > 0 {
		compositionCount++
		for i, subSchema := range schema.AllOf {
			v.validateSchema(fmt.Sprintf("%s.allOf[%d]", path, i), subSchema)
		}
	}
	if len(schema.AnyOf) > 0 {
		compositionCount++
		for i, subSchema := range schema.AnyOf {
			v.validateSchema(fmt.Sprintf("%s.anyOf[%d]", path, i), subSchema)
		}
	}

	if compositionCount > 1 {
		v.addWarning(path, "Schema has multiple composition keywords (oneOf, allOf, anyOf)")
	}
}

// validatePaths validates all paths in the spec
func (v *Validator) validatePaths() {
	for path, pathItem := range v.spec.Paths.PathItems {
		v.validatePath(path, pathItem)
	}
}

// validatePath validates a single path item
func (v *Validator) validatePath(path string, pathItem *spec.PathItem) {
	// Validate path pattern
	if !isValidPathPattern(path) {
		v.addWarning(fmt.Sprintf("paths.%s", path), "Invalid path pattern")
	}

	// Validate operations
	operations := map[string]*spec.Operation{
		"get":     pathItem.Get,
		"post":    pathItem.Post,
		"put":     pathItem.Put,
		"delete":  pathItem.Delete,
		"patch":   pathItem.Patch,
		"head":    pathItem.Head,
		"options": pathItem.Options,
	}

	for method, operation := range operations {
		if operation != nil {
			v.validateOperation(fmt.Sprintf("paths.%s.%s", path, method), operation)
		}
	}
}

// validateOperation validates an operation
func (v *Validator) validateOperation(path string, operation *spec.Operation) {
	// Check for duplicate OperationID
	if operation.OperationID != "" {
		if existingPath, ok := v.operationIDs[operation.OperationID]; ok {
			v.addError(path+".operationId", fmt.Sprintf("Duplicate operationId '%s' found at %s", operation.OperationID, existingPath))
		} else {
			v.operationIDs[operation.OperationID] = path
		}
	}

	// Validate parameters
	for i, param := range operation.Parameters {
		v.validateParameter(fmt.Sprintf("%s.parameters[%d]", path, i), param)
	}

	// Validate path parameters consistency
	v.validatePathParametersConsistency(path, operation)

	// Validate request body
	if operation.RequestBody != nil {
		v.validateRequestBody(fmt.Sprintf("%s.requestBody", path), operation.RequestBody)
	}

	// Validate responses
	if operation.Responses == nil {
		v.addError(path, "Operation must have responses")
	} else if operation.Responses.Default == nil && len(operation.Responses.StatusCodeResponses) == 0 {
		v.addError(path, "Operation must have at least one response")
	} else {
		v.validateResponses(fmt.Sprintf("%s.responses", path), operation.Responses)
	}
}

// validateParameter validates a parameter
func (v *Validator) validateParameter(path string, param *spec.Parameter) {
	if param.Name == "" {
		v.addError(path, "Parameter name is required")
	}

	validLocations := []string{"query", "header", "path", "cookie"}
	if !contains(validLocations, param.In) {
		v.addError(path+".in", fmt.Sprintf("Invalid parameter location: %s", param.In))
	}

	if param.In == "path" && !param.Required {
		v.addError(path, "Path parameters must be required")
	}

	if param.Schema != nil {
		v.validateSchema(path+".schema", param.Schema)
	}
}

// validateRequestBody validates a request body
func (v *Validator) validateRequestBody(path string, body *spec.RequestBody) {
	if len(body.Content) == 0 {
		v.addError(path, "Request body must have content")
	}

	for mediaType, content := range body.Content {
		if content.Schema != nil {
			v.validateSchema(fmt.Sprintf("%s.content.%s.schema", path, mediaType), content.Schema)
		}
	}
}

// validateResponses validates response definitions
func (v *Validator) validateResponses(path string, responses *spec.Responses) {
	// Validate default response if present
	if responses.Default != nil {
		v.validateResponse(fmt.Sprintf("%s.default", path), responses.Default)
	}

	// Validate status code responses
	for statusCode, response := range responses.StatusCodeResponses {
		v.validateResponse(fmt.Sprintf("%s.%s", path, statusCode), response)
	}
}

// validateResponse validates a single response
func (v *Validator) validateResponse(path string, response *spec.Response) {
	if response.Description == "" {
		v.addError(path, "Response description is required")
	}

	for mediaType, content := range response.Content {
		if content.Schema != nil {
			v.validateSchema(fmt.Sprintf("%s.content.%s.schema", path, mediaType), content.Schema)
		}
	}
}

// validateSecurityScheme validates a security scheme
func (v *Validator) validateSecurityScheme(path string, scheme *spec.SecurityScheme) {
	validTypes := []string{"apiKey", "http", "oauth2", "openIdConnect"}
	if !contains(validTypes, scheme.Type) {
		v.addError(path+".type", fmt.Sprintf("Invalid security scheme type: %s", scheme.Type))
	}

	switch scheme.Type {
	case "apiKey":
		if scheme.Name == "" {
			v.addError(path, "API key name is required")
		}
		validLocations := []string{"query", "header", "cookie"}
		if !contains(validLocations, scheme.In) {
			v.addError(path+".in", fmt.Sprintf("Invalid API key location: %s", scheme.In))
		}
	case "http":
		if scheme.Scheme == "" {
			v.addError(path, "HTTP scheme is required")
		}
	case "oauth2":
		if scheme.Flows == nil {
			v.addError(path, "OAuth2 flows are required")
		}
	}
}

// validateSecuritySchemes validates that referenced security schemes exist
func (v *Validator) validateSecuritySchemes() {
	if v.spec.Components == nil || v.spec.Components.SecuritySchemes == nil {
		return
	}

	// Check operation security references
	if v.spec.Paths != nil {
		for pathKey, pathItem := range v.spec.Paths.PathItems {
			operations := map[string]*spec.Operation{
				"get":    pathItem.Get,
				"post":   pathItem.Post,
				"put":    pathItem.Put,
				"delete": pathItem.Delete,
				"patch":  pathItem.Patch,
			}

			for method, operation := range operations {
				if operation != nil && len(operation.Security) > 0 {
					path := fmt.Sprintf("paths.%s.%s.security", pathKey, method)
					v.validateSecurityRequirements(path, operation.Security)
				}
			}
		}
	}
}

// validateSecurityRequirements validates security requirements reference existing schemes
func (v *Validator) validateSecurityRequirements(path string, requirements []spec.SecurityRequirement) {
	if v.spec.Components == nil || v.spec.Components.SecuritySchemes == nil {
		if len(requirements) > 0 {
			v.addError(path, "Security requirements defined but no security schemes in components")
		}
		return
	}

	for i, req := range requirements {
		for schemeName := range req {
			if _, exists := v.spec.Components.SecuritySchemes[schemeName]; !exists {
				v.addError(fmt.Sprintf("%s[%d]", path, i),
					fmt.Sprintf("Security scheme '%s' not found in components", schemeName))
			}
		}
	}
}

// validatePathParametersConsistency validates that path parameters match the path template
func (v *Validator) validatePathParametersConsistency(contextPath string, operation *spec.Operation) {
	// Extract path from context (e.g., paths./users/{id}.get -> /users/{id})
	// This is a bit hacky, relying on the contextPath format constructed in validatePath
	parts := strings.Split(contextPath, ".")
	if len(parts) < 2 {
		return
	}
	// The path part might contain dots, so we need to be careful.
	// The format is paths.<path>.<method> or similar.
	// But validateOperation receives `paths./users/{id}.get`.
	// Let's assume the path string is passed down or stored.
	// Actually, contextPath is just for error reporting strings.
	// We need the ACTUAL URL path template to validate.
	// The generic validator doesn't seem to pass it down easily.
	// Let's reconstruct it or change signature if needed.
	// Wait! v.operationIDs stores path! But incomplete.

	// Better approach: pass path template to validateOperation.
	// Since I can't easily change the signature everywhere in one go without breaking,
	// I'll parse it from contextPath which looks like "paths./foo/bar.get"
	// Find the substring between "paths." and result of last dot

	start := strings.Index(contextPath, "paths.")
	if start == -1 {
		return
	}
	start += 6 // len("paths.")

	end := strings.LastIndex(contextPath, ".")
	if end == -1 || end <= start {
		return
	}

	pathTemplate := contextPath[start:end]

	// Find parameters defined in path
	requiredParams := make(map[string]bool)
	matches := regexp.MustCompile(`\{([^}]+)\}`).FindAllStringSubmatch(pathTemplate, -1)
	for _, match := range matches {
		if len(match) > 1 {
			requiredParams[match[1]] = true
		}
	}

	// Check defined parameters
	definedParams := make(map[string]bool)
	for _, param := range operation.Parameters {
		if param.In == "path" {
			definedParams[param.Name] = true
			if !requiredParams[param.Name] {
				v.addError(contextPath, fmt.Sprintf("Parameter '%s' is in location 'path' but not in path template '%s'", param.Name, pathTemplate))
			}
		}
	}

	// Check missing parameters
	for req := range requiredParams {
		if !definedParams[req] {
			v.addError(contextPath, fmt.Sprintf("Path parameter '%s' is missing from operation parameters", req))
		}
	}
}

// Helper methods

func (v *Validator) addError(path string, message string) {
	v.result.Errors = append(v.result.Errors, ValidationError{
		Path:    path,
		Message: message,
		Level:   LevelError,
	})
}

func (v *Validator) addWarning(path string, message string) {
	v.result.Warnings = append(v.result.Warnings, ValidationError{
		Path:    path,
		Message: message,
		Level:   LevelWarning,
	})
}

// Utility functions

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func isValidURL(urlStr string) bool {
	parsed, err := url.Parse(urlStr)
	return err == nil && parsed.Scheme != "" && parsed.Host != ""
}

func isValidPathPattern(path string) bool {
	// Basic validation: path should start with /
	return strings.HasPrefix(path, "/")
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
