package spec

// RouteInfo stores information about a parsed route during scanning
type RouteInfo struct {
	// Method is the HTTP method (GET, POST, PUT, etc.)
	Method string

	// Path is the URL path
	Path string

	// Tags are the operation tags
	Tags []string

	// OperationID is the unique operation identifier
	OperationID string

	// Summary is a short description
	Summary string

	// Description is a detailed description
	Description string

	// Deprecated indicates if the operation is deprecated
	Deprecated bool

	// Responses contains operation responses
	Responses []ResponseInfo

	// Security contains security requirement names
	Security []string

	// Consumes contains accepted content types
	Consumes []string

	// Produces contains response content types
	Produces []string

	// IgnoredParameters contains parameter names to ignore
	IgnoredParameters []string

	// FilePath is the source file where the route was defined
	FilePath string

	// Specs contains spec names for multi-spec support
	Specs []string
}

// ResponseInfo stores information about a route response
type ResponseInfo struct {
	// StatusCode is the HTTP status code
	StatusCode string

	// Type is the response type name
	Type string

	// Description is the response description
	Description string

	// IsArray indicates if the response is an array
	IsArray bool

	// IsMap indicates if the response is a map
	IsMap bool

	// MapKeyType is the key type for map responses
	MapKeyType string
}

// NewRouteInfo creates a new RouteInfo with initialized slices
func NewRouteInfo() *RouteInfo {
	return &RouteInfo{
		Tags:              make([]string, 0),
		Responses:         make([]ResponseInfo, 0),
		Security:          make([]string, 0),
		Consumes:          make([]string, 0),
		Produces:          make([]string, 0),
		IgnoredParameters: make([]string, 0),
		Specs:             make([]string, 0),
	}
}

// HasIgnoredParameter checks if a parameter name is in the ignored list
func (r *RouteInfo) HasIgnoredParameter(paramName string) bool {
	for _, ignored := range r.IgnoredParameters {
		if ignored == paramName {
			return true
		}
	}
	return false
}

