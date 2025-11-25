package openapi

// OpenAPI Directives
const (
	// DirectiveMeta defines API metadata (title, version, description, etc.)
	DirectiveMeta = "swagger:meta"

	// DirectiveRoute defines API endpoints (paths and operations)
	DirectiveRoute = "swagger:route"

	// DirectiveModel defines data models (schemas)
	DirectiveModel = "swagger:model"
)

// Meta-level Tags (swagger:meta)
const (
	// API Information
	TagTitle          = "Title"
	TagVersion        = "Version"
	TagDescription    = "Description"
	TagTermsOfService = "TermsOfService"
	TagContact        = "Contact"
	TagLicense        = "License"

	// Server Configuration (Swagger 2.0)
	TagHost     = "Host"
	TagBasePath = "BasePath"
	TagSchemes  = "Schemes"

	// Server Configuration (OpenAPI 3.0)
	TagServers = "Servers"

	// Global Content Types
	TagConsumes = "Consumes"
	TagProduces = "Produces"

	// Security
	TagSecuritySchemes = "SecuritySchemes"
)

// Route-level Tags (swagger:route)
const (
	// Operation Identification
	TagOperationID = "OperationID"
	TagSummary     = "Summary"
	TagTags        = "Tags"

	// Operation Behavior
	TagDeprecated = "Deprecated"

	// Request/Response
	TagResponses  = "Responses"
	TagParameters = "Parameters"
	TagSecurity   = "Security"

	// Multi-Spec Support
	TagSpec = "Spec"
)

// Model-level Tags (swagger:model)
const (
	// Schema Properties
	TagExample   = "Example"
	TagDefault   = "Default"
	TagEnum      = "Enum"
	TagFormat    = "Format"
	TagMinimum   = "Minimum"
	TagMaximum   = "Maximum"
	TagMinLength = "MinLength"
	TagMaxLength = "MaxLength"
	TagPattern   = "Pattern"
	TagRequired  = "Required"
	TagReadOnly  = "ReadOnly"
	TagWriteOnly = "WriteOnly"

	// Extensions
	TagExtensions = "Extensions"
)

// Content Types
const (
	ContentTypeJSON           = "application/json"
	ContentTypeXML            = "application/xml"
	ContentTypeFormURLEncoded = "application/x-www-form-urlencoded"
	ContentTypeMultipartForm  = "multipart/form-data"
	ContentTypeTextPlain      = "text/plain"
	ContentTypeTextHTML       = "text/html"
	ContentTypeOctetStream    = "application/octet-stream"
)

// HTTP Methods
const (
	MethodGET     = "GET"
	MethodPOST    = "POST"
	MethodPUT     = "PUT"
	MethodPATCH   = "PATCH"
	MethodDELETE  = "DELETE"
	MethodHEAD    = "HEAD"
	MethodOPTIONS = "OPTIONS"
)

// Response Status Codes
const (
	StatusOK                  = "200"
	StatusCreated             = "201"
	StatusAccepted            = "202"
	StatusNoContent           = "204"
	StatusBadRequest          = "400"
	StatusUnauthorized        = "401"
	StatusForbidden           = "403"
	StatusNotFound            = "404"
	StatusConflict            = "409"
	StatusUnprocessableEntity = "422"
	StatusInternalServerError = "500"
	StatusDefault             = "default"
)

// Security Scheme Types
const (
	SecurityTypeAPIKey = "apiKey"
	SecurityTypeHTTP   = "http"
	SecurityTypeOAuth2 = "oauth2"
	SecurityTypeOpenID = "openIdConnect"
)

// OpenAPI Versions
const (
	Version303 = "3.0.3"
	Version310 = "3.1.0"
)

// Parameter Locations (in:xxx)
const (
	ParamLocationPath   = "path"
	ParamLocationQuery  = "query"
	ParamLocationHeader = "header"
	ParamLocationCookie = "cookie"
)

// Schema Formats
const (
	FormatInt32    = "int32"
	FormatInt64    = "int64"
	FormatFloat    = "float"
	FormatDouble   = "double"
	FormatByte     = "byte"
	FormatBinary   = "binary"
	FormatDate     = "date"
	FormatDateTime = "date-time"
	FormatPassword = "password"
	FormatEmail    = "email"
	FormatUUID     = "uuid"
	FormatURI      = "uri"
)

