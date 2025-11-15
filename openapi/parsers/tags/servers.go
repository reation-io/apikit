package tags

import (
	"encoding/json"

	"github.com/reation-io/apikit/openapi/parsers"
	"github.com/reation-io/apikit/openapi/parsers/base"
	"github.com/reation-io/apikit/openapi/spec"
)

// NewServersParser creates a Servers parser for swagger:meta
// Parses YAML content like:
// Servers:
//   - url: https://api.example.com/v1
//     description: Production server
//   - url: https://staging-api.example.com/v1
//     description: Staging server
func NewServersParser() parsers.TagParser {
	return base.NewYAMLParser(
		"Servers",
		parsers.RxServers,
		[]parsers.ParseContext{
			parsers.ContextMeta,
		},
		parsers.SetterMap{
			parsers.ContextMeta: func(target any, value any) error {
				openapi, ok := target.(*spec.OpenAPI)
				if !ok {
					return &parsers.ErrInvalidTarget{
						ParserName:   "Servers",
						Context:      parsers.ContextMeta,
						ExpectedType: "*spec.OpenAPI",
						ActualType:   getTypeName(target),
					}
				}

				// Value is json.RawMessage from YAMLParser
				rawMsg, ok := value.(json.RawMessage)
				if !ok {
					return &parsers.ErrInvalidValue{
						ParserName:   "Servers",
						ExpectedType: "json.RawMessage",
						ActualType:   getTypeName(value),
					}
				}

				// Parse into array of Server
				var servers []*spec.Server
				if err := json.Unmarshal(rawMsg, &servers); err != nil {
					return &parsers.ErrParseFailure{
						ParserName: "Servers",
						Context:    parsers.ContextMeta,
						Cause:      err,
					}
				}

				// Set servers
				openapi.Servers = servers

				return nil
			},
		},
	)
}

func init() {
	parsers.Register("swagger:meta", NewServersParser())
}

