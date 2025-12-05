package builder

import (
	"github.com/reation-io/apikit/openapi/scanner"
)

// BuilderConfig holds configuration options for the Builder
type BuilderConfig struct {
	// Scanner configuration
	ScannerConfig *scanner.Config

	// Validation enables automatic validation of the generated spec
	Validation bool
}

// Option is a function that configures the Builder
type Option func(*BuilderConfig)

// WithPattern sets the package pattern to scan (e.g., "./...", "./api/...")
// This enables the new go/packages scanner
func WithPattern(pattern string) Option {
	return func(c *BuilderConfig) {
		if c.ScannerConfig == nil {
			c.ScannerConfig = &scanner.Config{}
		}
		c.ScannerConfig.Pattern = pattern
	}
}

// WithDir sets the base directory to scan from
// This enables the new go/packages scanner
func WithDir(dir string) Option {
	return func(c *BuilderConfig) {
		if c.ScannerConfig == nil {
			c.ScannerConfig = &scanner.Config{}
		}
		c.ScannerConfig.Dir = dir
	}
}

// WithIgnorePaths sets path patterns to ignore during scanning
func WithIgnorePaths(paths ...string) Option {
	return func(c *BuilderConfig) {
		if c.ScannerConfig == nil {
			c.ScannerConfig = &scanner.Config{}
		}
		c.ScannerConfig.IgnorePaths = append(c.ScannerConfig.IgnorePaths, paths...)
	}
}

// WithValidation enables automatic validation of the generated spec
func WithValidation(enabled bool) Option {
	return func(c *BuilderConfig) {
		c.Validation = enabled
	}
}

// WithScannerConfig sets the full scanner configuration
func WithScannerConfig(config *scanner.Config) Option {
	return func(c *BuilderConfig) {
		c.ScannerConfig = config
	}
}
