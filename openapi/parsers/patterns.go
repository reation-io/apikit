package parsers

import (
	"fmt"
	"regexp"
	"sync"
)

// Centralized regex patterns for all parsers
// Inspired by go-swagger's regexprs.go

// Common pattern suffixes
const (
	// SectionEndPattern matches the end of a multi-line section
	// It matches either the start of a new directive or end of string
	SectionEndPattern = `(?:^[A-Z][a-zA-Z]*\s*:|\z)`

	// MapTypePattern matches Go map types: map[KeyType]ValueType
	MapTypePattern = `^map\[([^\]]+)\](.+)$`

	// MarkdownLinkPattern matches markdown links: [text](url)
	MarkdownLinkPattern = `\[([^\]]+)\]\(([^)]+)\)`

	// URLPattern matches HTTP/HTTPS URLs
	URLPattern = `(https?://\S+)`

	// EmailPattern matches email in angle brackets: <email@example.com>
	EmailPattern = `<([^>]+)>`
)

// Compiled regex cache for performance
var (
	regexCache      = make(map[string]*regexp.Regexp)
	regexCacheMutex sync.RWMutex
)

// getCompiledRegex returns a cached compiled regex for the given pattern
func getCompiledRegex(pattern string) *regexp.Regexp {
	regexCacheMutex.RLock()
	if rx, ok := regexCache[pattern]; ok {
		regexCacheMutex.RUnlock()
		return rx
	}
	regexCacheMutex.RUnlock()

	regexCacheMutex.Lock()
	defer regexCacheMutex.Unlock()

	// Double-check after acquiring write lock
	if rx, ok := regexCache[pattern]; ok {
		return rx
	}

	rx := regexp.MustCompile(pattern)
	regexCache[pattern] = rx
	return rx
}

// GetSectionRegex returns a cached compiled regex for extracting multi-line sections.
// The directive should be like "Description:" or "Responses:".
// The pattern captures everything from after the directive until the next directive or end of string.
func GetSectionRegex(directive string) *regexp.Regexp {
	pattern := fmt.Sprintf(`(?ms)^%s\s*$(.*?)%s`, regexp.QuoteMeta(directive), SectionEndPattern)
	return getCompiledRegex(pattern)
}

// GetSingleLineRegex returns a cached compiled regex for extracting single-line values.
// The directive should be like "Title:" or "Version:".
// The pattern captures the value on the same line as the directive.
func GetSingleLineRegex(directive string) *regexp.Regexp {
	pattern := fmt.Sprintf(`(?i)%s\s*([^\n]+)`, regexp.QuoteMeta(directive))
	return getCompiledRegex(pattern)
}

// GetMapTypeRegex returns a cached compiled regex for matching Go map types.
// Pattern matches: map[KeyType]ValueType
func GetMapTypeRegex() *regexp.Regexp {
	return getCompiledRegex(MapTypePattern)
}

// GetMarkdownLinkRegex returns a cached compiled regex for matching markdown links.
// Pattern matches: [text](url)
func GetMarkdownLinkRegex() *regexp.Regexp {
	return getCompiledRegex(MarkdownLinkPattern)
}

// GetURLRegex returns a cached compiled regex for matching URLs.
// Pattern matches: http:// or https:// URLs
func GetURLRegex() *regexp.Regexp {
	return getCompiledRegex(URLPattern)
}

// GetEmailRegex returns a cached compiled regex for matching emails in angle brackets.
// Pattern matches: <email@example.com>
func GetEmailRegex() *regexp.Regexp {
	return getCompiledRegex(EmailPattern)
}

var (
	// Meta-level patterns (swagger:meta) - single line patterns
	RxVersion        = regexp.MustCompile(`(?i)Version\s*:\s*([^\n]+)`)
	RxTitle          = regexp.MustCompile(`(?i)Title\s*:\s*([^\n]+)`)
	RxDescription    = regexp.MustCompile(`(?ms)^Description\s*:\s*$(.*?)(?:^[A-Z][a-zA-Z]*\s*:|\z)`) // Multi-line, case-sensitive, starts at beginning of line
	RxTermsOfService = regexp.MustCompile(`(?i)TermsOfService\s*:\s*([^\n]+)`)
	RxContact        = regexp.MustCompile(`(?is)Contact\s*:\s*\n((?:.*\n?)*)`)
	RxLicense        = regexp.MustCompile(`(?is)License\s*:\s*\n((?:.*\n?)*)`)
	RxHost           = regexp.MustCompile(`(?i)Host\s*:\s*([^\n]+)`)
	RxBasePath       = regexp.MustCompile(`(?i)BasePath\s*:\s*([^\n]+)`)
	RxSchemes        = regexp.MustCompile(`(?i)Schemes\s*:\s*([^\n]+)`)
	RxConsumes       = regexp.MustCompile(`(?i)Consumes\s*:\s*([^\n]+)`)
	RxProduces       = regexp.MustCompile(`(?i)Produces\s*:\s*([^\n]+)`)

	// Server patterns (OpenAPI 3.0)
	RxServers = regexp.MustCompile(`(?is)Servers\s*:\s*\n((?:.*\n?)*)`)

	// Security patterns
	RxSecurity        = regexp.MustCompile(`(?i)Security\s*:\s*([^\n]+)`)
	RxSecuritySchemes = regexp.MustCompile(`(?is)SecuritySchemes\s*:\s*\n((?:.*\n?)*)`)

	// Operation patterns (swagger:route)
	RxOperationID = regexp.MustCompile(`(?i)OperationID\s*:\s*([^\n]+)`)
	RxSummary     = regexp.MustCompile(`(?i)Summary\s*:\s*([^\n]+)`)
	RxTags        = regexp.MustCompile(`(?i)Tags\s*:\s*([^\n]+)`)
	RxDeprecated  = regexp.MustCompile(`(?i)Deprecated\s*:\s*(true|false|yes|no)`)
	RxResponses   = regexp.MustCompile(`(?is)Responses\s*:\s*\n((?:.*\n?)*)`)
	RxParameters  = regexp.MustCompile(`(?is)Parameters\s*:\s*\n((?:.*\n?)*)`)

	// Field patterns - all single line
	RxExample     = regexp.MustCompile(`(?i)Example\s*:\s*([^\n]+)`)
	RxDefault     = regexp.MustCompile(`(?i)Default\s*:\s*([^\n]+)`)
	RxEnum        = regexp.MustCompile(`(?i)Enum\s*:\s*([^\n]+)`)
	RxFormat      = regexp.MustCompile(`(?i)Format\s*:\s*([^\n]+)`)
	RxMinimum     = regexp.MustCompile(`(?i)Minimum\s*:\s*([^\n]+)`)
	RxMaximum     = regexp.MustCompile(`(?i)Maximum\s*:\s*([^\n]+)`)
	RxMinLength   = regexp.MustCompile(`(?i)MinLength\s*:\s*([^\n]+)`)
	RxMaxLength   = regexp.MustCompile(`(?i)MaxLength\s*:\s*([^\n]+)`)
	RxMinItems    = regexp.MustCompile(`(?i)MinItems\s*:\s*([^\n]+)`)
	RxMaxItems    = regexp.MustCompile(`(?i)MaxItems\s*:\s*([^\n]+)`)
	RxUniqueItems = regexp.MustCompile(`(?i)UniqueItems\s*:\s*(true|false|yes|no)`)
	RxMultipleOf  = regexp.MustCompile(`(?i)MultipleOf\s*:\s*([^\n]+)`)
	RxPattern     = regexp.MustCompile(`(?i)Pattern\s*:\s*([^\n]+)`)
	RxRequired    = regexp.MustCompile(`(?i)Required\s*:\s*(true|false|yes|no)`)
	RxReadOnly    = regexp.MustCompile(`(?i)ReadOnly\s*:\s*(true|false|yes|no)`)
	RxWriteOnly   = regexp.MustCompile(`(?i)WriteOnly\s*:\s*(true|false|yes|no)`)

	// External docs patterns
	RxExternalDocs = regexp.MustCompile(`(?is)ExternalDocs\s*:\s*\n((?:.*\n?)*)`)

	// Global Tags patterns (for swagger:meta)
	RxGlobalTags = regexp.MustCompile(`(?is)GlobalTags\s*:\s*\n((?:.*\n?)*)`)

	// Extension patterns
	RxExtensions = regexp.MustCompile(`(?is)Extensions\s*:\s*\n((?:.*\n?)*)`)
)
