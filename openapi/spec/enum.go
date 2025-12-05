package spec

// EnumInfo stores information about a detected enum type
type EnumInfo struct {
	// Name is the name of the enum (from swagger:enum directive or type name)
	Name string `json:"-" yaml:"-"`
	// TypeName is the Go type name
	TypeName string `json:"-" yaml:"-"`
	// BaseType is the underlying type (string, int, etc.)
	BaseType string `json:"-" yaml:"-"`
	// Values maps constant names to their values
	Values map[string]any `json:"-" yaml:"-"`
	// Description is the enum description from comments
	Description string `json:"-" yaml:"-"`
	// Example is an example value for the enum
	Example any `json:"-" yaml:"-"`
	// FilePath is the source file where the enum was defined
	FilePath string `json:"-" yaml:"-"`
}

// EnumRegistry stores all discovered enums during scanning
type EnumRegistry struct {
	// Enums maps enum names to their info
	Enums map[string]*EnumInfo
	// TypeToEnum maps Go type names to enum names (for type aliases)
	TypeToEnum map[string]string
	// Sources maps enum names to their source file paths
	Sources map[string]string
}

// NewEnumRegistry creates a new enum registry
func NewEnumRegistry() *EnumRegistry {
	return &EnumRegistry{
		Enums:      make(map[string]*EnumInfo),
		TypeToEnum: make(map[string]string),
		Sources:    make(map[string]string),
	}
}

// Register adds an enum to the registry
func (r *EnumRegistry) Register(enum *EnumInfo) {
	if enum == nil || enum.Name == "" {
		return
	}

	r.Enums[enum.Name] = enum
	r.TypeToEnum[enum.TypeName] = enum.Name
	r.Sources[enum.Name] = enum.FilePath
}

// Get returns an enum by name
func (r *EnumRegistry) Get(name string) *EnumInfo {
	return r.Enums[name]
}

// GetByTypeName returns an enum by its Go type name
func (r *EnumRegistry) GetByTypeName(typeName string) *EnumInfo {
	if enumName, ok := r.TypeToEnum[typeName]; ok {
		return r.Enums[enumName]
	}
	return nil
}

// AddValue adds a value to an existing enum
func (r *EnumRegistry) AddValue(enumName, constName string, value any) {
	if enum, ok := r.Enums[enumName]; ok {
		if enum.Values == nil {
			enum.Values = make(map[string]any)
		}
		enum.Values[constName] = value

		// Set as example if not set
		if enum.Example == nil {
			enum.Example = value
		}
	}
}

// RegisterTypeAlias registers a type alias for an existing enum
func (r *EnumRegistry) RegisterTypeAlias(aliasName, originalTypeName string) {
	if enumName, ok := r.TypeToEnum[originalTypeName]; ok {
		r.TypeToEnum[aliasName] = enumName
	}
}

// GetEnumValues returns the values of an enum as a slice (sorted by constant name)
func (e *EnumInfo) GetEnumValues() []any {
	if e == nil || len(e.Values) == 0 {
		return nil
	}

	values := make([]any, 0, len(e.Values))
	for _, v := range e.Values {
		values = append(values, v)
	}
	return values
}
