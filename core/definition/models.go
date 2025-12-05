package definition

import (
	"go/ast"
	"go/token"
)

// Definition represents the complete API definition
type Definition struct {
	// Name of the API or Service
	Name string
	// Package name where the definition was found
	Package string
	// Operations defined in the API
	Operations []*Operation
	// Types (Schemas) defined in the API
	Types map[string]*Type
	// Metadata contains generic metadata about the API (e.g. from swagger:meta)
	Metadata map[string]any
}

// Operation represents an API operation (e.g., GET /users)
type Operation struct {
	ID          string   // Unique identifier for the operation
	Method      string   // HTTP method (e.g., GET, POST)
	Path        string   // URL path
	Tags        []string // Tags for grouping operations
	Description string   // Description of the operation
	// Name of the request struct (e.g., "CreateUserRequest")
	RequestType string
	// Return type of the handler function
	ReturnType  string
	Parameters  []*Parameter
	RequestBody *RequestBody
	Responses   map[string]*Response
	Pos         token.Position // Position in the source file
	// Doc contains raw documentation comments (used for further parsing)
	Doc *ast.CommentGroup
}

// Parameter represents an operation parameter
type Parameter struct {
	Name        string
	In          string // Location of the parameter (e.g., query, path, header, cookie)
	Description string
	Required    bool
	Type        *Type // Type of the parameter
}

// RequestBody represents an operation request body
type RequestBody struct {
	Description string
	Content     map[string]*MediaType // Media types for the request body
}

// Response represents an operation response
type Response struct {
	Description string
	Content     map[string]*MediaType
}

// MediaType represents a media type for request or response body
type MediaType struct {
	Schema *Type // Schema of the media type
}

// Type represents a schema or data type
type Type struct {
	Name        string // Name of the type (e.g., User, string, int)
	Kind        string // Kind of type (e.g., struct, primitive, array, map)
	GoType      string // Original Go type string (e.g., "*User", "string", "[]int")
	Description string // Description of the type

	// For primitives
	Format string // date-time, uuid, etc.

	// For structs
	Fields []*Field

	// For slices/arrays
	ElementType *Type // For array types, the type of elements

	// For maps
	KeyType   *Type // For map types, the type of keys
	ValueType *Type // For map types, the type of values

	// For references
	RefName string
	// Doc contains raw documentation comments
	Doc *ast.CommentGroup
	// For enums
	EnumValues []any // List of allowed values
}

// Field represents a field in a struct
type Field struct {
	Name        string
	JSONName    string
	Type        *Type
	Description string
	Required    bool
	Tags        string         // Raw struct tags
	Metadata    map[string]any // Extra metadata (e.g. from comments)
	// Doc contains raw documentation comments
	Doc *ast.CommentGroup
	// Comment contains inline comments
	Comment *ast.CommentGroup
}
