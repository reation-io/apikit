package spec

// Schema represents a JSON Schema (OpenAPI 3.0)
type Schema struct {
	// Core schema properties
	Type        string `json:"type,omitempty" yaml:"type,omitempty"`
	Format      string `json:"format,omitempty" yaml:"format,omitempty"`
	Title       string `json:"title,omitempty" yaml:"title,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Default     any    `json:"default,omitempty" yaml:"default,omitempty"`
	Example     any    `json:"example,omitempty" yaml:"example,omitempty"`

	// Validation properties
	MultipleOf       *float64 `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`
	Maximum          *float64 `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	ExclusiveMaximum bool     `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
	Minimum          *float64 `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	ExclusiveMinimum bool     `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`
	MaxLength        *int64   `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	MinLength        *int64   `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	Pattern          string   `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	MaxItems         *int64   `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	MinItems         *int64   `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	UniqueItems      bool     `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`
	MaxProperties    *int64   `json:"maxProperties,omitempty" yaml:"maxProperties,omitempty"`
	MinProperties    *int64   `json:"minProperties,omitempty" yaml:"minProperties,omitempty"`
	Required         []string `json:"required,omitempty" yaml:"required,omitempty"`
	Enum             []any    `json:"enum,omitempty" yaml:"enum,omitempty"`

	// Object properties
	Properties           map[string]*Schema `json:"properties,omitempty" yaml:"properties,omitempty"`
	AdditionalProperties any                `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`

	// Array properties
	Items *Schema `json:"items,omitempty" yaml:"items,omitempty"`

	// Composition
	AllOf []*Schema `json:"allOf,omitempty" yaml:"allOf,omitempty"`
	OneOf []*Schema `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	AnyOf []*Schema `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
	Not   *Schema   `json:"not,omitempty" yaml:"not,omitempty"`

	// Reference
	Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"`

	// Metadata
	Nullable   bool `json:"nullable,omitempty" yaml:"nullable,omitempty"`
	ReadOnly   bool `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	WriteOnly  bool `json:"writeOnly,omitempty" yaml:"writeOnly,omitempty"`
	Deprecated bool `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	XML        *XML `json:"xml,omitempty" yaml:"xml,omitempty"`
}

// XML represents XML metadata
type XML struct {
	Name      string `json:"name,omitempty" yaml:"name,omitempty"`
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Prefix    string `json:"prefix,omitempty" yaml:"prefix,omitempty"`
	Attribute bool   `json:"attribute,omitempty" yaml:"attribute,omitempty"`
	Wrapped   bool   `json:"wrapped,omitempty" yaml:"wrapped,omitempty"`
}
