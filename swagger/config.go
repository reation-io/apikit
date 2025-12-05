package swagger

import "io/fs"

// Config holds the configuration for the Swagger handler.
type Config struct {
	// specs maps spec names to their content (JSON or YAML).
	specs map[string][]byte

	// basePath is the base path for API endpoints (default: "/openapi").
	basePath string

	// uiPath is the path where Swagger UI is served (default: "/swagger/").
	uiPath string

	// uiFS is a custom filesystem for Swagger UI assets.
	// If nil, the embedded default UI is used.
	uiFS fs.FS
}

// defaultConfig returns the default configuration.
func defaultConfig() *Config {
	return &Config{
		specs:    make(map[string][]byte),
		basePath: "/openapi",
		uiPath:   "/swagger/",
	}
}

// Option is a function that configures a Handler.
type Option func(*Config)

// WithSpec adds an OpenAPI specification with the given name.
// The spec can be JSON or YAML content.
func WithSpec(name string, spec []byte) Option {
	return func(c *Config) {
		if c.specs == nil {
			c.specs = make(map[string][]byte)
		}
		c.specs[name] = spec
	}
}

// WithSpecs sets multiple OpenAPI specifications at once.
func WithSpecs(specs map[string][]byte) Option {
	return func(c *Config) {
		c.specs = specs
	}
}

// WithBasePath sets the base path for API endpoints.
// Default is "/openapi".
func WithBasePath(path string) Option {
	return func(c *Config) {
		c.basePath = path
	}
}

// WithUIPath sets the path where Swagger UI is served.
// Default is "/swagger/".
func WithUIPath(path string) Option {
	return func(c *Config) {
		c.uiPath = path
	}
}

// WithUIFS sets a custom filesystem for Swagger UI assets.
// This can be used to serve a custom version of Swagger UI.
func WithUIFS(fsys fs.FS) Option {
	return func(c *Config) {
		c.uiFS = fsys
	}
}
