package parser

import (
	"go/ast"
	"go/token"
	"reflect"
	"strings"

	coreast "github.com/reation-io/apikit/core/ast"
	"github.com/reation-io/apikit/core/definition"
)

// Parser converts generic AST to API definition
type Parser struct {
	fset *token.FileSet
	def  *definition.Definition
}

// New creates a new parser
func New() *Parser {
	return &Parser{
		fset: token.NewFileSet(),
		def: &definition.Definition{
			Operations: make([]*definition.Operation, 0),
			Types:      make(map[string]*definition.Type),
		},
	}
}

// ParseFile parses a single Go source file
func (p *Parser) ParseFile(path string) (*definition.Definition, error) {
	// Use core/ast parser to get generic AST
	astParser := coreast.New()
	result, err := astParser.Parse(path)
	if err != nil {
		return nil, err
	}

	// Convert to definition
	return p.Parse(result), nil
}

// Parse converts a generic AST result into an API definition
func (p *Parser) Parse(generic *coreast.ParseResult) *definition.Definition {

	p.def.Name = generic.Package
	p.def.Package = generic.Package

	// 1. Convert all structs to Types
	for _, s := range generic.Structs {
		p.parseType(s)

		// Check if struct defines a route
		if route := extractSwaggerRoute(s.Doc); route != nil {
			op := &definition.Operation{
				ID:          route.OperationID,
				Method:      route.Method,
				Path:        route.Path,
				Tags:        []string{route.Tag},
				RequestType: s.Name,
				Description: getDocText(s.Doc),
				Pos:         s.Pos,
				Doc:         s.Doc,
			}
			p.def.Operations = append(p.def.Operations, op)
		}
	}

	// 2. Convert other types (Enums, Aliases)
	for _, t := range generic.Types {
		// Skip if already parsed (e.g. it was a struct)
		if _, exists := p.def.Types[t.Name]; exists {
			continue
		}
		p.parseTypeDecl(t)
	}

	// 3. Extract Enum Values from Constants
	p.extractEnumValues(generic)

	// 2. Find and convert handlers to Operations
	for _, fn := range generic.Functions {
		if hasDirective(fn.Doc, "apikit:handler") {
			op, err := p.parseOperation(fn)
			if err != nil {
				// TODO: handle error or accumulate warnings
				continue
			}
			if op != nil {
				p.def.Operations = append(p.def.Operations, op)
			}
		}
	}

	return p.def
}

// parseTypeExpr converts an AST type expression to definition.Type
func (p *Parser) parseTypeExpr(expr ast.Expr) *definition.Type {
	switch e := expr.(type) {
	case *ast.StructType:
		t := &definition.Type{
			Kind:   "struct",
			GoType: p.typeToString(expr),
			Fields: make([]*definition.Field, 0),
		}
		// Parse fields
		if e.Fields != nil {
			for _, f := range e.Fields.List {
				// Handle embedded fields and named fields
				names := make([]string, 0)
				if len(f.Names) > 0 {
					for _, name := range f.Names {
						names = append(names, name.Name)
					}
				} else {
					// Embedded field
					names = append(names, p.typeName(p.typeToString(f.Type)))
				}

				for _, name := range names {
					field := &definition.Field{
						Name:     name,
						Tags:     "",
						Metadata: make(map[string]string),
						Doc:      f.Doc,
						Comment:  f.Comment,
					}
					if f.Tag != nil {
						field.Tags = strings.Trim(f.Tag.Value, "`")
					}

					// Recursive parsing
					field.Type = p.parseTypeExpr(f.Type)

					// Parse comments for metadata
					inLoc, inName := parseInLocation(field)
					if inLoc != "" {
						field.Metadata["in"] = inLoc
						if inName != "" {
							field.Metadata["in_name"] = inName
						}
					}

					t.Fields = append(t.Fields, field)
				}
			}
		}
		return t

	case *ast.StarExpr:
		t := p.parseTypeExpr(e.X)
		// For pointer types, we want to update the GoType but potentially preserve Kind="struct"
		// if the underlying type is a struct.
		t.GoType = "*" + t.GoType
		return t

	default:
		return &definition.Type{
			GoType: p.typeToString(expr),
			Kind:   "primitive",
		}
	}
}

// Helper to get simple type name from string
func (p *Parser) typeName(t string) string {
	t = strings.TrimPrefix(t, "*")
	parts := strings.Split(t, ".")
	return parts[len(parts)-1]
}

// typeToString maps AST expr to string
func (p *Parser) typeToString(expr ast.Expr) string {
	// Re-using implementation concept from core/ast or simplistic version here
	// Since we don't have access to core/ast internal helpers easily, we implement a basic one
	// or rely on types.ExprString (if we imported go/types or go/format/Node)
	// But we can just use a switch like in core/ast
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return "*" + p.typeToString(e.X)
	case *ast.SelectorExpr:
		return p.typeToString(e.X) + "." + e.Sel.Name
	case *ast.ArrayType:
		if e.Len == nil {
			return "[]" + p.typeToString(e.Elt)
		}
		return "[" + p.exprToString(e.Len) + "]" + p.typeToString(e.Elt)
	case *ast.MapType:
		return "map[" + p.typeToString(e.Key) + "]" + p.typeToString(e.Value)
	case *ast.InterfaceType:
		return "any"
	case *ast.Ellipsis:
		return "..." + p.typeToString(e.Elt) // Variadic handled as slice usually
	case *ast.StructType:
		// We shouldn't really convert full struct to string here recursively if it's large,
		// but GoType usually expects string repr.
		// For anonymous struct "struct { A int }"
		return "struct { ... }" // Simplification, or formatted
	default:
		return ""
	}
}

func (p *Parser) exprToString(expr ast.Expr) string {
	if lit, ok := expr.(*ast.BasicLit); ok {
		return lit.Value
	}
	return ""
}

func (p *Parser) parseType(s *coreast.Struct) *definition.Type {
	if existing, ok := p.def.Types[s.Name]; ok {
		return existing
	}

	t := &definition.Type{
		Name:   s.Name,
		Kind:   "struct",
		GoType: s.Name,
		Fields: make([]*definition.Field, 0),
		Doc:    s.Doc,
	}

	for _, f := range s.Fields {
		field := &definition.Field{
			Name:     f.Name,
			Required: !f.IsPointer,
			Tags:     f.Tag,
			Metadata: make(map[string]string),
			Doc:      f.Doc,
			Comment:  f.Comment,
		}

		// Parse JSON tag for name
		if val, ok := getTag(f.Tag, "json"); ok {
			parts := strings.Split(val, ",")
			field.JSONName = parts[0]
		}

		// Resolve type using recursive helper
		field.Type = p.parseTypeExpr(f.ASTType)

		// Parse metadata for main struct fields too
		inLoc, inName := parseInLocation(field)
		if inLoc != "" {
			field.Metadata["in"] = inLoc
			if inName != "" {
				field.Metadata["in_name"] = inName
			}
		}

		t.Fields = append(t.Fields, field)
	}

	p.def.Types[s.Name] = t
	return t
}
func (p *Parser) parseOperation(fn *coreast.Function) (*definition.Operation, error) {
	op := &definition.Operation{
		ID:          fn.Name,
		Description: getDocText(fn.Doc),
		Pos:         fn.Pos,
		Doc:         fn.Doc,
	}

	// Capture return type
	if len(fn.Results) > 0 {
		var returns []string
		for _, res := range fn.Results {
			returns = append(returns, res.Type)
		}
		if len(returns) > 1 {
			op.ReturnType = "(" + strings.Join(returns, ", ") + ")"
		} else {
			op.ReturnType = returns[0]
		}
	}

	// Extract Method and Path from comments (e.g., // @GET /api/users)
	// Or rely on tags inside the struct as apikit currently does?
	// The current apikit seems to rely on the struct fields tags `in:"query"`, etc.
	// But where is the HTTP method/path defined?
	// Looking at apikit/handler/parser/parser.go, it doesn't seem to extract Method/Path for the handler itself?
	// Wait, `handler/extractor` just extracts values.
	// Ah, the Method/Path might be defined in the routing layer, often manual in `main.go` or similar.
	// OR maybe it's missing from the current `handler` model because it wasn't needed for *wrapper* generation?
	// But `openapi` needs it.

	// Check for apikit:handler comment
	// apikit:handler METHOD PATH
	if route := extractAPIKitHandler(fn.Doc); route != nil {
		op.Method = route.Method
		op.Path = route.Path
	}

	// Check if there is a swagger:route comment
	// swagger:route POST /users users createUser
	if route := extractSwaggerRoute(fn.Doc); route != nil {
		op.Method = route.Method
		op.Path = route.Path
		op.Tags = []string{route.Tag}
	}

	// Helper to find the request struct
	// Based on handler signature: func(ctx, RequestStruct) ...
	if len(fn.Params) >= 2 {
		reqTypeStr := fn.Params[1].Type
		// Clean pointer/package
		reqTypeName := cleanTypeName(reqTypeStr)

		op.RequestType = reqTypeName

		if reqType, ok := p.def.Types[reqTypeName]; ok {
			// Extract parameters from the request struct
			for _, field := range reqType.Fields {
				// Parse "in" location from comments or tags
				inLoc, inName := parseInLocation(field)

				if inLoc != "" {
					field.Metadata["in"] = inLoc
					if inName != "" {
						field.Metadata["in_name"] = inName
					}

					param := &definition.Parameter{
						Name:        field.Name,
						In:          inLoc,
						Description: field.Description,
						Required:    field.Required,
						Type:        field.Type,
					}

					// Override name if specified (e.g. // in:query id)
					if inName != "" {
						param.Name = inName
					}

					// If no specific name in comment, check json tag
					if param.Name == field.Name && field.JSONName != "" {
						param.Name = field.JSONName
					}

					op.Parameters = append(op.Parameters, param)
				} else if field.JSONName != "-" {
					// Assume it's part of the Request Body if not in path/query/header
					if op.RequestBody == nil {
						op.RequestBody = &definition.RequestBody{
							Content: make(map[string]*definition.MediaType),
						}
						// TODO: Build the schema for the body
						// For now we just mark existence
					}
				}
			}
		}
	}

	// Parse return type
	if len(fn.Results) > 0 {
		op.ReturnType = fn.Results[0].Type
	}

	return op, nil
}

// Helper functions

func hasDirective(group *ast.CommentGroup, directive string) bool {
	if group == nil {
		return false
	}
	for _, c := range group.List {
		if strings.Contains(c.Text, directive) {
			return true
		}
	}
	return false
}

func getDocText(group *ast.CommentGroup) string {
	if group == nil {
		return ""
	}
	var lines []string
	for _, c := range group.List {
		lines = append(lines, strings.TrimPrefix(c.Text, "// "))
	}
	return strings.Join(lines, "\n")
}

func getTag(tagStr string, key string) (string, bool) {
	if tagStr == "" {
		return "", false
	}
	return reflect.StructTag(tagStr).Lookup(key)
}

func cleanTypeName(t string) string {
	t = strings.TrimPrefix(t, "*")
	parts := strings.Split(t, ".")
	return parts[len(parts)-1]
}

type routeInfo struct {
	Method      string
	Path        string
	Tag         string
	OperationID string
}

func extractAPIKitHandler(group *ast.CommentGroup) *routeInfo {
	if group == nil {
		return nil
	}
	for _, c := range group.List {
		text := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
		if strings.HasPrefix(text, "apikit:handler") {
			parts := strings.Fields(text)
			// apikit:handler METHOD PATH
			if len(parts) >= 3 {
				return &routeInfo{
					Method: parts[1],
					Path:   parts[2],
				}
			}
		}
	}
	return nil
}

func extractSwaggerRoute(group *ast.CommentGroup) *routeInfo {
	if group == nil {
		return nil
	}
	for _, c := range group.List {
		text := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
		if strings.HasPrefix(text, "swagger:route") {
			parts := strings.Fields(text)
			if len(parts) >= 4 {
				info := &routeInfo{
					Method: parts[1],
					Path:   parts[2],
					Tag:    parts[3],
				}
				if len(parts) >= 5 {
					info.OperationID = parts[4]
				}
				return info
			}
		}
	}
	return nil
}

func parseInLocation(field *definition.Field) (string, string) {
	// 1. Check tags: `in:"query"`
	if val, ok := getTag(field.Tags, "in"); ok {
		return val, ""
	}

	// 2. Check tags: `query:"name"`
	if val, ok := getTag(field.Tags, "query"); ok {
		return "query", val
	}
	if val, ok := getTag(field.Tags, "path"); ok {
		return "path", val
	}
	if val, ok := getTag(field.Tags, "header"); ok {
		return "header", val
	}
	if val, ok := getTag(field.Tags, "cookie"); ok {
		return "cookie", val
	}

	// 3. Check tags: `json:"name"` (only if we determine it's NOT body? No, that's just name)

	// 4. Parse comments // in:query
	if field.Doc != nil {
		for _, c := range field.Doc.List {
			text := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
			if strings.HasPrefix(text, "in:") {
				parts := strings.Fields(text)
				if len(parts) >= 1 {
					// format: in:location [name]
					loc := strings.TrimPrefix(parts[0], "in:")
					name := ""
					if len(parts) > 1 {
						name = parts[1]
					}
					return loc, name
				}
			}
		}
	}

	return "", ""
}

func (p *Parser) parseTypeDecl(td *coreast.TypeDecl) {
	// Simple parsing for non-struct types
	typeName := p.typeToString(td.TypeSpec.Type)
	t := &definition.Type{
		Name:        td.Name,
		Kind:        "primitive", // Default, might be alias
		GoType:      typeName,
		Description: getDocText(td.Doc),
		Doc:         td.Doc,
	}

	// Check for swagger:enum
	if hasDirective(td.Doc, "swagger:enum") {
		t.Kind = "enum"
	}

	p.def.Types[td.Name] = t
}

func (p *Parser) extractEnumValues(generic *coreast.ParseResult) {
	// Map constants to their types
	for _, c := range generic.Constants {
		if c.Type == "" {
			continue
		}

		// Check if the type is a registered enum
		if t, ok := p.def.Types[c.Type]; ok {
			// Add value to enum
			t.EnumValues = append(t.EnumValues, c.Value)
		}
	}
}
