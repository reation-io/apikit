package builder

import (
	"go/ast"
	"go/token"
	"strings"

	constants "github.com/reation-io/apikit/openapi"
	"github.com/reation-io/apikit/openapi/spec"
)

// parseEnums parses swagger:enum comments and their associated constants
func (b *Builder) parseEnums(file *ast.File) error {
	// First pass: find all enum type declarations
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		if genDecl.Doc == nil {
			continue
		}

		// Check if this is a swagger:enum comment
		if !hasDirective(genDecl.Doc, constants.DirectiveEnum) {
			continue
		}

		// Process each type spec in this declaration
		for _, s := range genDecl.Specs {
			typeSpec, ok := s.(*ast.TypeSpec)
			if !ok {
				continue
			}

			enumInfo := b.parseEnumDecl(genDecl.Doc, typeSpec)
			if enumInfo != nil {
				b.enumRegistry.Register(enumInfo)
			}
		}
	}

	// Second pass: find constants for registered enums
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			continue
		}

		for _, s := range genDecl.Specs {
			valueSpec, ok := s.(*ast.ValueSpec)
			if !ok || valueSpec.Type == nil || len(valueSpec.Values) == 0 {
				continue
			}

			// Get the type name
			typeIdent, ok := valueSpec.Type.(*ast.Ident)
			if !ok {
				continue
			}

			// Check if this type is a registered enum
			enumInfo := b.enumRegistry.GetByTypeName(typeIdent.Name)
			if enumInfo == nil {
				continue
			}

			// Extract values from this constant declaration
			for i, name := range valueSpec.Names {
				if i >= len(valueSpec.Values) {
					break
				}

				value := extractConstValue(valueSpec.Values[i])
				if value != nil {
					b.enumRegistry.AddValue(enumInfo.Name, name.Name, value)
				}
			}
		}
	}

	return nil
}

// parseEnumDecl parses a single enum type declaration
func (b *Builder) parseEnumDecl(doc *ast.CommentGroup, typeSpec *ast.TypeSpec) *spec.EnumInfo {
	typeName := typeSpec.Name.Name

	// Get enum name from directive (swagger:enum EnumName) or use type name
	enumName := extractDirectiveValue(doc, constants.DirectiveEnum)
	if enumName == "" {
		enumName = typeName
	}

	// Get base type
	baseType := getEnumBaseType(typeSpec.Type)

	// Get description (everything except the directive and example lines)
	description := extractEnumDescription(doc)

	// Get example value
	example := extractEnumTagValue(doc, "example:")

	return &spec.EnumInfo{
		Name:        enumName,
		TypeName:    typeName,
		BaseType:    baseType,
		Values:      make(map[string]any),
		Description: description,
		Example:     example,
	}
}

// getEnumBaseType returns the base type of an enum type declaration
func getEnumBaseType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		// Map Go types to JSON Schema types
		mapping := MapGoTypeToOpenAPI(t.Name)
		if mapping.Type != "object" {
			return mapping.Type
		}
		// Default to string if assumed enum type is object
		return "string"
	case *ast.SelectorExpr:
		// Handle qualified types
		return "string"
	}
	return "string"
}

// extractConstValue extracts a value from a constant expression
func extractConstValue(expr ast.Expr) any {
	switch v := expr.(type) {
	case *ast.BasicLit:
		switch v.Kind {
		case token.INT:
			return v.Value
		case token.FLOAT:
			return v.Value
		case token.STRING:
			// Remove quotes
			return strings.Trim(v.Value, `"'`)
		}
	case *ast.Ident:
		// For identifiers like iota, skip them
		if v.Name == "iota" {
			return nil
		}
		return v.Name
	}
	return nil
}

// extractEnumDescription extracts description from enum comments
// excluding the directive and example lines
func extractEnumDescription(doc *ast.CommentGroup) string {
	if doc == nil {
		return ""
	}

	var lines []string
	for _, comment := range doc.List {
		text := strings.TrimPrefix(comment.Text, "//")
		text = strings.TrimPrefix(text, "/*")
		text = strings.TrimSuffix(text, "*/")
		text = strings.TrimSpace(text)

		// Skip directive and tag lines
		if strings.HasPrefix(text, "swagger:") ||
			strings.HasPrefix(strings.ToLower(text), "example:") {
			continue
		}

		if text != "" {
			lines = append(lines, text)
		}
	}

	return strings.Join(lines, "\n")
}

// extractEnumTagValue extracts a tag value from enum comments (e.g., "example: value")
func extractEnumTagValue(doc *ast.CommentGroup, tag string) any {
	if doc == nil {
		return nil
	}

	tagLower := strings.ToLower(tag)
	for _, comment := range doc.List {
		text := strings.TrimPrefix(comment.Text, "//")
		text = strings.TrimPrefix(text, "/*")
		text = strings.TrimSuffix(text, "*/")
		text = strings.TrimSpace(text)

		if strings.HasPrefix(strings.ToLower(text), tagLower) {
			value := strings.TrimSpace(text[len(tag):])
			return strings.Trim(value, `"'`)
		}
	}

	return nil
}

// extractDirectiveValue extracts the value after a swagger directive
// e.g., "swagger:enum UserStatus" returns "UserStatus"
func extractDirectiveValue(doc *ast.CommentGroup, directive string) string {
	if doc == nil {
		return ""
	}

	for _, comment := range doc.List {
		text := strings.TrimPrefix(comment.Text, "//")
		text = strings.TrimPrefix(text, "/*")
		text = strings.TrimSuffix(text, "*/")
		text = strings.TrimSpace(text)

		if strings.HasPrefix(text, directive) {
			value := strings.TrimSpace(text[len(directive):])
			// Take only the first word
			if idx := strings.IndexAny(value, " \t\n"); idx > 0 {
				value = value[:idx]
			}
			return value
		}
	}

	return ""
}
