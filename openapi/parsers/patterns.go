package parsers

import "regexp"

// Centralized regex patterns for all parsers
// Inspired by go-swagger's regexprs.go

var (
	// Meta-level patterns (swagger:meta) - single line patterns
	RxVersion        = regexp.MustCompile(`(?i)Version\s*:\s*([^\n]+)`)
	RxTitle          = regexp.MustCompile(`(?i)Title\s*:\s*([^\n]+)`)
	RxDescription    = regexp.MustCompile(`(?is)Description\s*:\s*(.*)`) // Multi-line
	RxTermsOfService = regexp.MustCompile(`(?i)TermsOfService\s*:\s*([^\n]+)`)
	RxContact        = regexp.MustCompile(`(?i)Contact\s*:\s*([^\n]+)`)
	RxLicense        = regexp.MustCompile(`(?i)License\s*:\s*([^\n]+)`)
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
	RxExample   = regexp.MustCompile(`(?i)Example\s*:\s*([^\n]+)`)
	RxDefault   = regexp.MustCompile(`(?i)Default\s*:\s*([^\n]+)`)
	RxEnum      = regexp.MustCompile(`(?i)Enum\s*:\s*([^\n]+)`)
	RxFormat    = regexp.MustCompile(`(?i)Format\s*:\s*([^\n]+)`)
	RxMinimum   = regexp.MustCompile(`(?i)Minimum\s*:\s*([^\n]+)`)
	RxMaximum   = regexp.MustCompile(`(?i)Maximum\s*:\s*([^\n]+)`)
	RxMinLength = regexp.MustCompile(`(?i)MinLength\s*:\s*([^\n]+)`)
	RxMaxLength = regexp.MustCompile(`(?i)MaxLength\s*:\s*([^\n]+)`)
	RxPattern   = regexp.MustCompile(`(?i)Pattern\s*:\s*([^\n]+)`)
	RxRequired  = regexp.MustCompile(`(?i)Required\s*:\s*(true|false|yes|no)`)
	RxReadOnly  = regexp.MustCompile(`(?i)ReadOnly\s*:\s*(true|false|yes|no)`)
	RxWriteOnly = regexp.MustCompile(`(?i)WriteOnly\s*:\s*(true|false|yes|no)`)

	// Extension patterns
	RxExtensions = regexp.MustCompile(`(?is)Extensions\s*:\s*\n((?:.*\n?)*)`)
)
