package spec

// Operation describe una operación en un path
type Operation struct {
	Tags         []string              `json:"tags,omitempty" yaml:"tags,omitempty"`
	Summary      string                `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description  string                `json:"description,omitempty" yaml:"description,omitempty"`
	ExternalDocs *ExternalDocs         `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	OperationID  string                `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Parameters   []*Parameter          `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody  *RequestBody          `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses    *Responses            `json:"responses" yaml:"responses"`
	Callbacks    map[string]*Callback  `json:"callbacks,omitempty" yaml:"callbacks,omitempty"`
	Deprecated   bool                  `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	Security     []SecurityRequirement `json:"security,omitempty" yaml:"security,omitempty"`
	Servers      []*Server             `json:"servers,omitempty" yaml:"servers,omitempty"`
	Extensions   map[string]any        `json:"-" yaml:"-"` // Extensions for custom properties
}

// Parameter describe un parámetro de operación
type Parameter struct {
	Name            string              `json:"name" yaml:"name"`
	In              string              `json:"in" yaml:"in"` // query, header, path, cookie
	Description     string              `json:"description,omitempty" yaml:"description,omitempty"`
	Required        bool                `json:"required,omitempty" yaml:"required,omitempty"`
	Deprecated      bool                `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	AllowEmptyValue bool                `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
	Schema          *Schema             `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example         any                 `json:"example,omitempty" yaml:"example,omitempty"`
	Examples        map[string]*Example `json:"examples,omitempty" yaml:"examples,omitempty"`
}

// RequestBody describe el cuerpo de una petición
type RequestBody struct {
	Description string                `json:"description,omitempty" yaml:"description,omitempty"`
	Content     map[string]*MediaType `json:"content" yaml:"content"`
	Required    bool                  `json:"required,omitempty" yaml:"required,omitempty"`
}

// MediaType describe un media type
type MediaType struct {
	Schema   *Schema              `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example  any                  `json:"example,omitempty" yaml:"example,omitempty"`
	Examples map[string]*Example  `json:"examples,omitempty" yaml:"examples,omitempty"`
	Encoding map[string]*Encoding `json:"encoding,omitempty" yaml:"encoding,omitempty"`
}

// Encoding describe la codificación de una propiedad
type Encoding struct {
	ContentType   string             `json:"contentType,omitempty" yaml:"contentType,omitempty"`
	Headers       map[string]*Header `json:"headers,omitempty" yaml:"headers,omitempty"`
	Style         string             `json:"style,omitempty" yaml:"style,omitempty"`
	Explode       bool               `json:"explode,omitempty" yaml:"explode,omitempty"`
	AllowReserved bool               `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
}

// Header describe un header
type Header struct {
	Description     string              `json:"description,omitempty" yaml:"description,omitempty"`
	Required        bool                `json:"required,omitempty" yaml:"required,omitempty"`
	Deprecated      bool                `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	AllowEmptyValue bool                `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
	Schema          *Schema             `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example         any                 `json:"example,omitempty" yaml:"example,omitempty"`
	Examples        map[string]*Example `json:"examples,omitempty" yaml:"examples,omitempty"`
}

// Example representa un ejemplo
type Example struct {
	Summary       string `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description   string `json:"description,omitempty" yaml:"description,omitempty"`
	Value         any    `json:"value,omitempty" yaml:"value,omitempty"`
	ExternalValue string `json:"externalValue,omitempty" yaml:"externalValue,omitempty"`
}

// Responses contiene las respuestas de una operación
type Responses struct {
	Default             *Response            `json:"default,omitempty" yaml:"default,omitempty"`
	StatusCodeResponses map[string]*Response `json:"-" yaml:"-"`
}

// MarshalJSON implementa json.Marshaler
func (r *Responses) MarshalJSON() ([]byte, error) {
	m := make(map[string]*Response)
	for k, v := range r.StatusCodeResponses {
		m[k] = v
	}
	if r.Default != nil {
		m["default"] = r.Default
	}
	return marshalMap(m)
}

// MarshalYAML implementa yaml.Marshaler
func (r *Responses) MarshalYAML() (any, error) {
	m := make(map[string]*Response)
	for k, v := range r.StatusCodeResponses {
		m[k] = v
	}
	if r.Default != nil {
		m["default"] = r.Default
	}
	return m, nil
}

// Response describe una respuesta
type Response struct {
	Description string                `json:"description" yaml:"description"`
	Headers     map[string]*Header    `json:"headers,omitempty" yaml:"headers,omitempty"`
	Content     map[string]*MediaType `json:"content,omitempty" yaml:"content,omitempty"`
	Links       map[string]*Link      `json:"links,omitempty" yaml:"links,omitempty"`
}

// Link representa un link
type Link struct {
	OperationRef string         `json:"operationRef,omitempty" yaml:"operationRef,omitempty"`
	OperationID  string         `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Parameters   map[string]any `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody  any            `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Description  string         `json:"description,omitempty" yaml:"description,omitempty"`
	Server       *Server        `json:"server,omitempty" yaml:"server,omitempty"`
}

// Callback representa un callback
type Callback map[string]*PathItem
