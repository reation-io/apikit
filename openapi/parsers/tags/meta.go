package tags

import (
	"encoding/json"
	"strings"

	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/parsers/base"
	"github.com/reation-io/apikit/openapi/spec"
)

// NewTermsOfServiceParser creates a TermsOfService parser for swagger:meta
func NewTermsOfServiceParser() parsers.TagParser {
	return base.NewSingleLineParser(
		"TermsOfService",
		parsers.RxTermsOfService,
		[]parsers.ParseContext{parsers.ContextMeta},
		parsers.SetterMap{
			parsers.ContextMeta: func(target any, value any) error {
				info, ok := target.(*spec.Info)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "TermsOfService",
						Context:      parsers.ContextMeta,
						ExpectedType: "*spec.Info",
						ActualType:   getTypeName(target),
					}
				}
				if str, ok := value.(string); ok {
					info.TermsOfService = str
				}
				return nil
			},
		},
	)
}

// NewContactParser creates a Contact parser for swagger:meta
// Parses YAML content like:
// Contact:
//
//	name: API Support
//	url: https://www.example.com/support
//	email: support@example.com
func NewContactParser() parsers.TagParser {
	return base.NewYAMLParser(
		"Contact",
		parsers.RxContact,
		[]parsers.ParseContext{parsers.ContextMeta},
		parsers.SetterMap{
			parsers.ContextMeta: func(target any, value any) error {
				info, ok := target.(*spec.Info)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "Contact",
						Context:      parsers.ContextMeta,
						ExpectedType: "*spec.Info",
						ActualType:   getTypeName(target),
					}
				}

				rawMsg, ok := value.(json.RawMessage)
				if !ok {
					// Try as string (simple email format)
					if str, ok := value.(string); ok {
						info.Contact = &spec.Contact{Email: strings.TrimSpace(str)}
						return nil
					}
					return nil
				}

				var contact spec.Contact
				if err := json.Unmarshal(rawMsg, &contact); err != nil {
					// If JSON parse fails, try simple email
					info.Contact = &spec.Contact{Email: strings.TrimSpace(string(rawMsg))}
					return nil
				}

				info.Contact = &contact
				return nil
			},
		},
	)
}

// NewLicenseParser creates a License parser for swagger:meta
// Parses YAML content like:
// License:
//
//	name: Apache 2.0
//	url: https://www.apache.org/licenses/LICENSE-2.0.html
func NewLicenseParser() parsers.TagParser {
	return base.NewYAMLParser(
		"License",
		parsers.RxLicense,
		[]parsers.ParseContext{parsers.ContextMeta},
		parsers.SetterMap{
			parsers.ContextMeta: func(target any, value any) error {
				info, ok := target.(*spec.Info)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "License",
						Context:      parsers.ContextMeta,
						ExpectedType: "*spec.Info",
						ActualType:   getTypeName(target),
					}
				}

				rawMsg, ok := value.(json.RawMessage)
				if !ok {
					// Try as string (simple name format)
					if str, ok := value.(string); ok {
						info.License = &spec.License{Name: strings.TrimSpace(str)}
						return nil
					}
					return nil
				}

				var license spec.License
				if err := json.Unmarshal(rawMsg, &license); err != nil {
					// If JSON parse fails, try simple name
					info.License = &spec.License{Name: strings.TrimSpace(string(rawMsg))}
					return nil
				}

				info.License = &license
				return nil
			},
		},
	)
}

// NewExternalDocsParser creates an ExternalDocs parser for swagger:meta
// Parses YAML content like:
// ExternalDocs:
//
//	description: Find out more about Swagger
//	url: https://swagger.io
func NewExternalDocsParser() parsers.TagParser {
	return base.NewYAMLParser(
		"ExternalDocs",
		parsers.RxExternalDocs,
		[]parsers.ParseContext{parsers.ContextMeta},
		parsers.SetterMap{
			parsers.ContextMeta: func(target any, value any) error {
				openapi, ok := target.(*spec.OpenAPI)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "ExternalDocs",
						Context:      parsers.ContextMeta,
						ExpectedType: "*spec.OpenAPI",
						ActualType:   getTypeName(target),
					}
				}

				rawMsg, ok := value.(json.RawMessage)
				if !ok {
					return nil
				}

				var extDocs spec.ExternalDocs
				if err := json.Unmarshal(rawMsg, &extDocs); err != nil {
					return nil
				}

				openapi.ExternalDocs = &extDocs
				return nil
			},
		},
	)
}

// NewGlobalTagsParser creates a GlobalTags parser for swagger:meta
// Parses YAML content like:
// GlobalTags:
//   - name: pet
//     description: Everything about your Pets
//     externalDocs:
//     description: Find out more
//     url: https://swagger.io
func NewGlobalTagsParser() parsers.TagParser {
	return base.NewYAMLParser(
		"GlobalTags",
		parsers.RxGlobalTags,
		[]parsers.ParseContext{parsers.ContextMeta},
		parsers.SetterMap{
			parsers.ContextMeta: func(target any, value any) error {
				openapi, ok := target.(*spec.OpenAPI)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "GlobalTags",
						Context:      parsers.ContextMeta,
						ExpectedType: "*spec.OpenAPI",
						ActualType:   getTypeName(target),
					}
				}

				rawMsg, ok := value.(json.RawMessage)
				if !ok {
					return nil
				}

				var tags []*spec.Tag
				if err := json.Unmarshal(rawMsg, &tags); err != nil {
					return nil
				}

				openapi.Tags = tags
				return nil
			},
		},
	)
}

func init() {
	parsers.Register("swagger:meta", NewTermsOfServiceParser())
	parsers.Register("swagger:meta", NewContactParser())
	parsers.Register("swagger:meta", NewLicenseParser())
	parsers.Register("swagger:meta", NewExternalDocsParser())
	parsers.Register("swagger:meta", NewGlobalTagsParser())
}
