package tags

import (
	"fmt"
	"go/ast"
	"regexp"
	"strings"

	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/spec"
)

// ResponsesParser parses the Responses directive for routes
// Format:
// Responses:
// - 200: SuccessResponse
// - 400: ErrorResponse
// - 404: NotFoundResponse
type ResponsesParser struct {
	parsers.BaseParser
}

func init() {
	parsers.GlobalRegistry().Register("swagger:route", &ResponsesParser{
		BaseParser: parsers.NewBaseParser(
			"responses",
			parsers.ParserTypeMultiLine,
			[]parsers.ParseContext{parsers.ContextRoute},
			nil,
		),
	})
}

// Pattern matches response lines like "- 200: ResponseType" or "- default: ErrorResponse"
var responseLinePattern = regexp.MustCompile(`^\s*-\s*(\d{3}|default)\s*:\s*(.+)$`)

// Pattern to extract Responses section
var responsesPattern = regexp.MustCompile(`(?ms)^Responses:\s*$(.*?)(?:^[A-Z][a-zA-Z]*:\s*$|\z)`)

// Matches checks if the comment contains Responses directive
func (p *ResponsesParser) Matches(comment string, ctx parsers.ParseContext) bool {
	return ctx == parsers.ContextRoute && strings.Contains(comment, "Responses:")
}

// Parse extracts responses from multi-line Responses: section
func (p *ResponsesParser) Parse(comments *ast.CommentGroup, ctx parsers.ParseContext) (any, error) {
	if ctx != parsers.ContextRoute {
		return nil, nil
	}

	text := comments.Text()

	// Extract the Responses section
	section := extractSection(text, "Responses:")
	if section == "" {
		return nil, nil
	}

	// Parse each response line and return as map
	responses := make(map[string]*spec.Response)
	var defaultResponse *spec.Response

	lines := strings.Split(section, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse response line
		response := parseResponseLine(line)
		if response == nil {
			continue
		}

		// Store response
		if response.StatusCode == "default" {
			defaultResponse = response.Response
		} else {
			responses[response.StatusCode] = response.Response
		}
	}

	return &ParsedResponses{
		StatusCodeResponses: responses,
		Default:             defaultResponse,
	}, nil
}

// Apply applies the parsed responses to the operation
func (p *ResponsesParser) Apply(target any, value any, ctx parsers.ParseContext) error {
	if ctx != parsers.ContextRoute {
		return nil
	}

	operation, ok := target.(*spec.Operation)
	if !ok {
		return &parsers.ErrInvalidTarget{
			ParserName:   "responses",
			Context:      ctx,
			ExpectedType: "*spec.Operation",
			ActualType:   fmt.Sprintf("%T", target),
		}
	}

	parsedResponses, ok := value.(*ParsedResponses)
	if !ok {
		// If value is nil, nothing to apply
		if value == nil {
			return nil
		}
		return &parsers.ErrInvalidValue{
			ParserName:   "responses",
			ExpectedType: "*ParsedResponses",
			ActualType:   fmt.Sprintf("%T", value),
		}
	}

	// Initialize responses if needed
	if operation.Responses == nil {
		operation.Responses = &spec.Responses{
			StatusCodeResponses: make(map[string]*spec.Response),
		}
	}

	// Apply responses
	for statusCode, response := range parsedResponses.StatusCodeResponses {
		operation.Responses.StatusCodeResponses[statusCode] = response
	}

	if parsedResponses.Default != nil {
		operation.Responses.Default = parsedResponses.Default
	}

	return nil
}

// ParsedResponses holds the parsed response data
type ParsedResponses struct {
	StatusCodeResponses map[string]*spec.Response
	Default             *spec.Response
}

// ParsedResponse represents a parsed response line
type ParsedResponse struct {
	StatusCode string
	Response   *spec.Response
}

// parseResponseLine parses a single response line
// Format: "- 200: ResponseType" or "- default: ErrorResponse"
func parseResponseLine(line string) *ParsedResponse {
	matches := responseLinePattern.FindStringSubmatch(line)
	if len(matches) != 3 {
		return nil
	}

	statusCode := strings.TrimSpace(matches[1])
	responseType := strings.TrimSpace(matches[2])

	if statusCode == "" || responseType == "" {
		return nil
	}

	// Create response with schema reference
	response := &spec.Response{
		Description: getDefaultDescription(statusCode),
		Content:     make(map[string]*spec.MediaType),
	}

	// Add JSON content with schema reference
	response.Content["application/json"] = &spec.MediaType{
		Schema: &spec.Schema{
			Ref: fmt.Sprintf("#/components/schemas/%s", responseType),
		},
	}

	return &ParsedResponse{
		StatusCode: statusCode,
		Response:   response,
	}
}

// getDefaultDescription returns a default description for common status codes
func getDefaultDescription(statusCode string) string {
	descriptions := map[string]string{
		"200":     "OK",
		"201":     "Created",
		"202":     "Accepted",
		"204":     "No Content",
		"400":     "Bad Request",
		"401":     "Unauthorized",
		"403":     "Forbidden",
		"404":     "Not Found",
		"409":     "Conflict",
		"422":     "Unprocessable Entity",
		"500":     "Internal Server Error",
		"502":     "Bad Gateway",
		"503":     "Service Unavailable",
		"default": "Error",
	}

	if desc, ok := descriptions[statusCode]; ok {
		return desc
	}

	return "Response"
}

// SupportsContext returns true if the parser supports the given context
func (p *ResponsesParser) SupportsContext(context parsers.ParseContext) bool {
	return context == parsers.ContextRoute
}

// Name returns the parser name
func (p *ResponsesParser) Name() string {
	return "responses"
}

// extractSection extracts a multi-line section from comments
// Format:
// SectionName:
// content line 1
// content line 2
// NextSection:
func extractSection(text, sectionName string) string {
	// Pattern to match section until next directive or end
	pattern := fmt.Sprintf(`(?ms)^%s\s*$(.*?)(?:^[A-Z][a-zA-Z]*:\s*$|\z)`, regexp.QuoteMeta(sectionName))
	re := regexp.MustCompile(pattern)

	matches := re.FindStringSubmatch(text)
	if len(matches) < 2 {
		return ""
	}

	return strings.TrimSpace(matches[1])
}
