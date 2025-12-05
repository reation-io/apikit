package spec

// FieldInfo stores information about a struct field during parsing
// This is used internally by the builder to track field metadata
type FieldInfo struct {
	// Name is the field name (from json tag or Go field name)
	Name string

	// Type is the OpenAPI type (string, integer, object, etc.)
	Type string

	// Description from comments
	Description string

	// Default value
	Default string

	// Example value
	Example string

	// Required indicates if the field is required
	Required bool

	// Nullable indicates if the field can be null
	Nullable bool

	// Validations contains validation rules (min, max, pattern, etc.)
	Validations map[string]string

	// Enum contains enum type name reference
	Enum string

	// Tags contains struct tags (json, xml, etc.)
	Tags map[string]string

	// IsArray indicates if this is an array/slice type
	IsArray bool

	// IsPointer indicates if this is a pointer type
	IsPointer bool

	// IsMap indicates if this is a map type
	IsMap bool

	// MapKeyType is the key type for maps
	MapKeyType string

	// IsRequestBody indicates if this field is the request body
	IsRequestBody bool

	// IsInlineStruct indicates if this is an inline struct
	IsInlineStruct bool

	// InlineStruct contains the inline struct info
	InlineStruct *StructInfo

	// HasOmitempty indicates if the field has omitempty/omitzero tag
	HasOmitempty bool

	// ExplicitRequired indicates if the field is explicitly marked as required
	ExplicitRequired bool

	// ExplicitOptional indicates if the field is explicitly marked as optional
	ExplicitOptional bool
}

// StructInfo stores information about a struct during parsing
type StructInfo struct {
	// Name is the struct name (from swagger:model/parameters directive or Go type name)
	Name string

	// Fields contains all struct fields
	Fields []*FieldInfo

	// Description from comments
	Description string

	// IsParameter indicates this is a swagger:parameters struct
	IsParameter bool

	// IsModel indicates this is a swagger:model struct
	IsModel bool

	// FilePath is the source file where the struct was defined
	FilePath string

	// OneOf contains schema names for oneOf composition
	OneOf []string

	// AllOf contains schema names for allOf composition
	AllOf []string

	// AnyOf contains schema names for anyOf composition
	AnyOf []string
}

// NewFieldInfo creates a new FieldInfo with initialized maps
func NewFieldInfo() *FieldInfo {
	return &FieldInfo{
		Tags:        make(map[string]string),
		Validations: make(map[string]string),
	}
}

// NewStructInfo creates a new StructInfo with initialized slices
func NewStructInfo() *StructInfo {
	return &StructInfo{
		Fields: make([]*FieldInfo, 0),
	}
}

// Clone creates a deep copy of FieldInfo
func (f *FieldInfo) Clone() *FieldInfo {
	if f == nil {
		return nil
	}

	clone := &FieldInfo{
		Name:             f.Name,
		Type:             f.Type,
		Description:      f.Description,
		Default:          f.Default,
		Example:          f.Example,
		Required:         f.Required,
		Nullable:         f.Nullable,
		Enum:             f.Enum,
		IsArray:          f.IsArray,
		IsPointer:        f.IsPointer,
		IsMap:            f.IsMap,
		MapKeyType:       f.MapKeyType,
		IsRequestBody:    f.IsRequestBody,
		IsInlineStruct:   f.IsInlineStruct,
		HasOmitempty:     f.HasOmitempty,
		ExplicitRequired: f.ExplicitRequired,
		ExplicitOptional: f.ExplicitOptional,
		Tags:             make(map[string]string),
		Validations:      make(map[string]string),
	}

	for k, v := range f.Tags {
		clone.Tags[k] = v
	}
	for k, v := range f.Validations {
		clone.Validations[k] = v
	}

	return clone
}

