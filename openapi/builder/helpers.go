package builder

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/reation-io/apikit/openapi/parsers"
)

// hasDirective checks if comments contain a specific directive
func hasDirective(comments *ast.CommentGroup, directive string) bool {
	if comments == nil {
		return false
	}
	text := comments.Text()
	return strings.Contains(text, directive)
}

// isInvalidTargetError checks if an error is an invalid target error
func isInvalidTargetError(err error) bool {
	_, ok := err.(*parsers.ErrInvalidTarget)
	return ok
}

// routeInfo contains parsed route information
type routeInfo struct {
	Method      string
	Path        string
	Tag         string
	OperationID string
}

// parseRouteLine parses the swagger:route line
// Format: swagger:route METHOD PATH TAG OPERATION_ID
// TAG can be quoted with single or double quotes if it contains spaces
func parseRouteLine(comments *ast.CommentGroup) (*routeInfo, error) {
	if comments == nil {
		return nil, fmt.Errorf("no comments provided")
	}

	for _, comment := range comments.List {
		text := strings.TrimSpace(comment.Text)
		text = strings.TrimPrefix(text, "//")
		text = strings.TrimPrefix(text, "/*")
		text = strings.TrimSuffix(text, "*/")
		text = strings.TrimSpace(text)

		if !strings.HasPrefix(text, "swagger:route") {
			continue
		}

		// Remove "swagger:route" prefix
		text = strings.TrimPrefix(text, "swagger:route")
		text = strings.TrimSpace(text)

		// Parse with quote awareness
		parts := parseQuotedFields(text)
		if len(parts) < 4 {
			return nil, fmt.Errorf("invalid swagger:route format, expected: swagger:route METHOD PATH TAG OPERATION_ID")
		}

		return &routeInfo{
			Method:      parts[0],
			Path:        parts[1],
			Tag:         parts[2],
			OperationID: parts[3],
		}, nil
	}

	return nil, fmt.Errorf("no swagger:route directive found")
}

// parseQuotedFields parses a string into fields, respecting quoted strings
// Example: "GET /path 'My Tag' opId" -> ["GET", "/path", "My Tag", "opId"]
func parseQuotedFields(s string) []string {
	var fields []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for i, r := range s {
		switch {
		case (r == '\'' || r == '"') && !inQuote:
			// Start of quoted string
			inQuote = true
			quoteChar = r
		case r == quoteChar && inQuote:
			// End of quoted string
			inQuote = false
			quoteChar = 0
			// Add the accumulated field
			if current.Len() > 0 {
				fields = append(fields, current.String())
				current.Reset()
			}
		case (r == ' ' || r == '\t') && !inQuote:
			// Whitespace outside quotes - field separator
			if current.Len() > 0 {
				fields = append(fields, current.String())
				current.Reset()
			}
		default:
			// Regular character
			current.WriteRune(r)
		}

		// Handle end of string
		if i == len(s)-1 && current.Len() > 0 {
			fields = append(fields, current.String())
		}
	}

	return fields
}
