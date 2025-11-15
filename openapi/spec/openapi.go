package spec

import "gopkg.in/yaml.v3"

// OpenAPI representa la estructura raíz de una especificación OpenAPI 3.0
type OpenAPI struct {
	OpenAPI      string                `json:"openapi" yaml:"openapi"`
	Info         *Info                 `json:"info" yaml:"info"`
	Servers      []*Server             `json:"servers,omitempty" yaml:"servers,omitempty"`
	Paths        *Paths                `json:"paths" yaml:"paths"`
	Components   *Components           `json:"components,omitempty" yaml:"components,omitempty"`
	Security     []SecurityRequirement `json:"security,omitempty" yaml:"security,omitempty"`
	Tags         []*Tag                `json:"tags,omitempty" yaml:"tags,omitempty"`
	ExternalDocs *ExternalDocs         `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	Extensions   map[string]any        `json:"-" yaml:"-"` // Extensions for custom properties
}

// Info contiene metadata sobre la API
type Info struct {
	Title          string   `json:"title" yaml:"title"`
	Description    string   `json:"description,omitempty" yaml:"description,omitempty"`
	TermsOfService string   `json:"termsOfService,omitempty" yaml:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty" yaml:"contact,omitempty"`
	License        *License `json:"license,omitempty" yaml:"license,omitempty"`
	Version        string   `json:"version" yaml:"version"`
}

// Contact contiene información de contacto
type Contact struct {
	Name  string `json:"name,omitempty" yaml:"name,omitempty"`
	URL   string `json:"url,omitempty" yaml:"url,omitempty"`
	Email string `json:"email,omitempty" yaml:"email,omitempty"`
}

// License contiene información de licencia
type License struct {
	Name string `json:"name" yaml:"name"`
	URL  string `json:"url,omitempty" yaml:"url,omitempty"`
}

// Server representa un servidor
type Server struct {
	URL         string                     `json:"url" yaml:"url"`
	Description string                     `json:"description,omitempty" yaml:"description,omitempty"`
	Variables   map[string]*ServerVariable `json:"variables,omitempty" yaml:"variables,omitempty"`
}

// ServerVariable representa una variable de servidor
type ServerVariable struct {
	Enum        []string `json:"enum,omitempty" yaml:"enum,omitempty"`
	Default     string   `json:"default" yaml:"default"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
}

// Paths contiene todos los paths disponibles
type Paths struct {
	PathItems map[string]*PathItem `json:"-" yaml:"-"`
}

// MarshalJSON implementa json.Marshaler
func (p *Paths) MarshalJSON() ([]byte, error) {
	return marshalMap(p.PathItems)
}

// MarshalYAML implementa yaml.Marshaler
func (p *Paths) MarshalYAML() (any, error) {
	return p.PathItems, nil
}

// UnmarshalJSON implementa json.Unmarshaler
func (p *Paths) UnmarshalJSON(data []byte) error {
	p.PathItems = make(map[string]*PathItem)
	return unmarshalMap(data, &p.PathItems)
}

// UnmarshalYAML implementa yaml.Unmarshaler
func (p *Paths) UnmarshalYAML(node *yaml.Node) error {
	p.PathItems = make(map[string]*PathItem)
	return node.Decode(&p.PathItems)
}

// PathItem describe las operaciones disponibles en un path
type PathItem struct {
	Ref         string       `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Summary     string       `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string       `json:"description,omitempty" yaml:"description,omitempty"`
	Get         *Operation   `json:"get,omitempty" yaml:"get,omitempty"`
	Put         *Operation   `json:"put,omitempty" yaml:"put,omitempty"`
	Post        *Operation   `json:"post,omitempty" yaml:"post,omitempty"`
	Delete      *Operation   `json:"delete,omitempty" yaml:"delete,omitempty"`
	Options     *Operation   `json:"options,omitempty" yaml:"options,omitempty"`
	Head        *Operation   `json:"head,omitempty" yaml:"head,omitempty"`
	Patch       *Operation   `json:"patch,omitempty" yaml:"patch,omitempty"`
	Trace       *Operation   `json:"trace,omitempty" yaml:"trace,omitempty"`
	Servers     []*Server    `json:"servers,omitempty" yaml:"servers,omitempty"`
	Parameters  []*Parameter `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

// Tag agrupa operaciones
type Tag struct {
	Name         string        `json:"name" yaml:"name"`
	Description  string        `json:"description,omitempty" yaml:"description,omitempty"`
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
}

// ExternalDocs referencia documentación externa
type ExternalDocs struct {
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	URL         string `json:"url" yaml:"url"`
}

// SecurityRequirement lista los esquemas de seguridad requeridos
type SecurityRequirement map[string][]string
