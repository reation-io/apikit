package extractors

import (
	"strings"
	"testing"

	"github.com/reation-io/apikit/handler/parser"
)

func TestRegister(t *testing.T) {
	// Save original extractors
	originalExtractors := globalRegistry.extractors
	defer func() {
		globalRegistry.extractors = originalExtractors
	}()

	// Reset registry
	globalRegistry.extractors = []Extractor{}

	// Create mock extractor
	mockExtractor := &mockExtractor{name: "test", priority: 50}
	Register(mockExtractor)

	extractors := GetExtractors()
	if len(extractors) != 1 {
		t.Errorf("expected 1 extractor, got %d", len(extractors))
	}

	if extractors[0].Name() != "test" {
		t.Errorf("expected extractor name 'test', got %q", extractors[0].Name())
	}
}

func TestRegister_Priority(t *testing.T) {
	// Save original extractors
	originalExtractors := globalRegistry.extractors
	defer func() {
		globalRegistry.extractors = originalExtractors
	}()

	// Reset registry
	globalRegistry.extractors = []Extractor{}

	// Register extractors with different priorities
	Register(&mockExtractor{name: "low", priority: 100})
	Register(&mockExtractor{name: "high", priority: 10})
	Register(&mockExtractor{name: "medium", priority: 50})

	extractors := GetExtractors()
	if len(extractors) != 3 {
		t.Fatalf("expected 3 extractors, got %d", len(extractors))
	}

	// Should be sorted by priority
	if extractors[0].Name() != "high" {
		t.Errorf("expected first extractor to be 'high', got %q", extractors[0].Name())
	}
	if extractors[1].Name() != "medium" {
		t.Errorf("expected second extractor to be 'medium', got %q", extractors[1].Name())
	}
	if extractors[2].Name() != "low" {
		t.Errorf("expected third extractor to be 'low', got %q", extractors[2].Name())
	}
}

func TestGetExtractor(t *testing.T) {
	// Save original extractors
	originalExtractors := globalRegistry.extractors
	defer func() {
		globalRegistry.extractors = originalExtractors
	}()

	// Reset and register mock extractor
	globalRegistry.extractors = []Extractor{}
	mockExt := &mockExtractor{name: "test", canExtract: true}
	Register(mockExt)

	field := &parser.Field{Name: "TestField"}
	extractor := GetExtractor(field)

	if extractor == nil {
		t.Fatal("expected to find extractor")
	}
	if extractor.Name() != "test" {
		t.Errorf("expected extractor name 'test', got %q", extractor.Name())
	}
}

func TestGetExtractor_NotFound(t *testing.T) {
	// Save original extractors
	originalExtractors := globalRegistry.extractors
	defer func() {
		globalRegistry.extractors = originalExtractors
	}()

	// Reset and register mock extractor that can't extract
	globalRegistry.extractors = []Extractor{}
	mockExt := &mockExtractor{name: "test", canExtract: false}
	Register(mockExt)

	field := &parser.Field{Name: "TestField"}
	extractor := GetExtractor(field)

	if extractor != nil {
		t.Error("expected not to find extractor")
	}
}

func TestGetDefaultTag(t *testing.T) {
	tests := []struct {
		name     string
		field    *parser.Field
		expected string
	}{
		{
			name:     "no tag",
			field:    &parser.Field{StructTag: ""},
			expected: "",
		},
		{
			name:     "with default tag",
			field:    &parser.Field{StructTag: `json:"name" default:"unknown"`},
			expected: "unknown",
		},
		{
			name:     "no default tag",
			field:    &parser.Field{StructTag: `json:"name"`},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDefaultTag(tt.field)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"UserID", "userID"},
		{"FirstName", "firstName"},
		{"APIKey", "aPIKey"},
		{"name", "name"},
		{"", ""},
		{"A", "a"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toCamelCase(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetParameterName(t *testing.T) {
	tests := []struct {
		name     string
		field    *parser.Field
		tagName  string
		expected string
	}{
		{
			name:     "from tag",
			field:    &parser.Field{Name: "UserID", StructTag: `path:"userId"`},
			tagName:  "path",
			expected: "userId",
		},
		{
			name:     "from comment",
			field:    &parser.Field{Name: "UserID", InCommentName: "user_id"},
			tagName:  "path",
			expected: "user_id",
		},
		{
			name:     "from field name",
			field:    &parser.Field{Name: "UserID"},
			tagName:  "path",
			expected: "userID",
		},
		{
			name:     "empty tag value falls back to comment",
			field:    &parser.Field{Name: "UserID", StructTag: `path:""`, InCommentName: "user_id"},
			tagName:  "path",
			expected: "user_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetParameterName(tt.field, tt.tagName)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGenerateIntParsing(t *testing.T) {
	code := GenerateIntParsing("value", "Age", "int64")

	if !strings.Contains(code, "strconv.ParseInt") {
		t.Error("expected ParseInt call")
	}
	if !strings.Contains(code, "payload.Age") {
		t.Error("expected field assignment")
	}
	if !strings.Contains(code, "int64(i)") {
		t.Error("expected type conversion")
	}
}

func TestGenerateUintParsing(t *testing.T) {
	code := GenerateUintParsing("value", "Count", "uint32")

	if !strings.Contains(code, "strconv.ParseUint") {
		t.Error("expected ParseUint call")
	}
	if !strings.Contains(code, "payload.Count") {
		t.Error("expected field assignment")
	}
	if !strings.Contains(code, "uint32(i)") {
		t.Error("expected type conversion")
	}
}

func TestGenerateFloatParsing(t *testing.T) {
	code := GenerateFloatParsing("value", "Price", "32")

	if !strings.Contains(code, "strconv.ParseFloat") {
		t.Error("expected ParseFloat call")
	}
	if !strings.Contains(code, "payload.Price") {
		t.Error("expected field assignment")
	}
	if !strings.Contains(code, "float32(f)") {
		t.Error("expected type conversion")
	}
}

func TestGenerateBoolParsing(t *testing.T) {
	code := GenerateBoolParsing("value", "Active")

	if !strings.Contains(code, "strconv.ParseBool") {
		t.Error("expected ParseBool call")
	}
	if !strings.Contains(code, "payload.Active") {
		t.Error("expected field assignment")
	}
}

func TestIsStringType(t *testing.T) {
	tests := []struct {
		typeName string
		expected bool
	}{
		{"string", true},
		{"int", false},
		{"bool", false},
	}

	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			result := IsStringType(tt.typeName)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsIntType(t *testing.T) {
	intTypes := []string{"int", "int8", "int16", "int32", "int64"}
	for _, typeName := range intTypes {
		t.Run(typeName, func(t *testing.T) {
			if !IsIntType(typeName) {
				t.Errorf("expected %s to be int type", typeName)
			}
		})
	}

	if IsIntType("uint") {
		t.Error("uint should not be int type")
	}
	if IsIntType("string") {
		t.Error("string should not be int type")
	}
}

func TestIsUintType(t *testing.T) {
	uintTypes := []string{"uint", "uint8", "uint16", "uint32", "uint64"}
	for _, typeName := range uintTypes {
		t.Run(typeName, func(t *testing.T) {
			if !IsUintType(typeName) {
				t.Errorf("expected %s to be uint type", typeName)
			}
		})
	}

	if IsUintType("int") {
		t.Error("int should not be uint type")
	}
}

func TestIsFloatType(t *testing.T) {
	if !IsFloatType("float32") {
		t.Error("expected float32 to be float type")
	}
	if !IsFloatType("float64") {
		t.Error("expected float64 to be float type")
	}
	if IsFloatType("int") {
		t.Error("int should not be float type")
	}
}

func TestIsBoolType(t *testing.T) {
	if !IsBoolType("bool") {
		t.Error("expected bool to be bool type")
	}
	if IsBoolType("string") {
		t.Error("string should not be bool type")
	}
}

// Mock extractor for testing
type mockExtractor struct {
	name       string
	priority   int
	canExtract bool
}

func (m *mockExtractor) Name() string                        { return m.name }
func (m *mockExtractor) Priority() int                       { return m.priority }
func (m *mockExtractor) CanExtract(field *parser.Field) bool { return m.canExtract }
func (m *mockExtractor) GenerateCode(field *parser.Field, structName string) (string, []string) {
	return "mock code", []string{}
}
