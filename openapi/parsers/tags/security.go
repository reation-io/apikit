package tags

import (
	"fmt"
	"go/ast"
	"regexp"
	"strings"

	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/spec"
)

// SecurityParser parses the Security directive for routes
// Format:
// Security:
// - bearer
// - api_key
// - oauth:
//   - read
//   - write
type SecurityParser struct {
	parsers.BaseParser
}

func init() {
	parsers.GlobalRegistry().Register("swagger:route", &SecurityParser{
		BaseParser: parsers.NewBaseParser(
			"security",
			parsers.ParserTypeMultiLine,
			[]parsers.ParseContext{parsers.ContextRoute},
			nil,
		),
	})
}

// Pattern matches security lines like "- bearer" or "- oauth:"
var securityLinePattern = regexp.MustCompile(`^\s*-\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*:?\s*$`)

// Pattern matches scope lines like "- read" (after trimming)
var scopeLinePattern = regexp.MustCompile(`^-\s+([a-zA-Z_][a-zA-Z0-9_:]*)\s*$`)

// Matches checks if the comment contains Security directive
func (p *SecurityParser) Matches(comment string, ctx parsers.ParseContext) bool {
	return ctx == parsers.ContextRoute && strings.Contains(comment, "Security:")
}

// Parse extracts security requirements from multi-line Security: section
func (p *SecurityParser) Parse(comments *ast.CommentGroup, ctx parsers.ParseContext) (any, error) {
	if ctx != parsers.ContextRoute {
		return nil, nil
	}

	text := comments.Text()

	// Extract the Security section
	section := extractSection(text, "Security:")
	if section == "" {
		return nil, nil
	}

	// Parse security requirements
	requirements := parseSecurityRequirements(section)

	return requirements, nil
}

// Apply applies the parsed security requirements to the operation
func (p *SecurityParser) Apply(target any, value any, ctx parsers.ParseContext) error {
	if ctx != parsers.ContextRoute {
		return nil
	}

	operation, ok := target.(*spec.Operation)
	if !ok {
		return &parsers.ErrInvalidTarget{
			ParserName:   "security",
			Context:      ctx,
			ExpectedType: "*spec.Operation",
			ActualType:   fmt.Sprintf("%T", target),
		}
	}

	requirements, ok := value.([]spec.SecurityRequirement)
	if !ok {
		// If value is nil, nothing to apply
		if value == nil {
			return nil
		}
		return &parsers.ErrInvalidValue{
			ParserName:   "security",
			ExpectedType: "[]spec.SecurityRequirement",
			ActualType:   fmt.Sprintf("%T", value),
		}
	}

	// Initialize security if needed
	if operation.Security == nil {
		operation.Security = []spec.SecurityRequirement{}
	}

	// Apply security requirements
	operation.Security = append(operation.Security, requirements...)

	return nil
}

// parseSecurityRequirements parses the security section and returns requirements
func parseSecurityRequirements(section string) []spec.SecurityRequirement {
	var requirements []spec.SecurityRequirement
	lines := strings.Split(section, "\n")

	var currentScheme string
	var currentScopes []string
	var hasScopes bool

	for _, line := range lines {
		// Don't trim yet - we need to check indentation
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Count leading spaces to determine if this is a scope (indented more)
		leadingSpaces := len(line) - len(strings.TrimLeft(line, " \t"))

		// Security schemes start with "- " at the beginning (after section extraction)
		// Scopes are indented with additional spaces before "- "
		isIndented := leadingSpaces > 0

		// Check if this is a security scheme line (not indented or minimally indented)
		if !isIndented || !hasScopes {
			matches := securityLinePattern.FindStringSubmatch(trimmed)
			if len(matches) == 2 {
				// Save previous scheme if exists
				if currentScheme != "" {
					req := spec.SecurityRequirement{
						currentScheme: currentScopes,
					}
					requirements = append(requirements, req)
				}

				// Start new scheme
				currentScheme = strings.TrimSpace(matches[1])
				currentScopes = []string{}
				hasScopes = strings.HasSuffix(trimmed, ":")

				// If no scopes expected, add immediately and reset
				if !hasScopes {
					req := spec.SecurityRequirement{
						currentScheme: []string{},
					}
					requirements = append(requirements, req)
					currentScheme = ""
				}
				continue
			}
		}

		// Check if this is a scope line (indented and we're expecting scopes)
		if currentScheme != "" && hasScopes && isIndented {
			scopeMatches := scopeLinePattern.FindStringSubmatch(trimmed)
			if len(scopeMatches) == 2 {
				scope := strings.TrimSpace(scopeMatches[1])
				currentScopes = append(currentScopes, scope)
			}
		}
	}

	// Add any remaining scheme
	if currentScheme != "" {
		req := spec.SecurityRequirement{
			currentScheme: currentScopes,
		}
		requirements = append(requirements, req)
	}

	return requirements
}

// SupportsContext returns true if the parser supports the given context
func (p *SecurityParser) SupportsContext(context parsers.ParseContext) bool {
	return context == parsers.ContextRoute
}

// Name returns the parser name
func (p *SecurityParser) Name() string {
	return "security"
}
