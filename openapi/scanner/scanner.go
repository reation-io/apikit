package scanner

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Config holds configuration options for the Scanner
type Config struct {
	// Pattern is the package pattern to scan (e.g., "./...", "./api/...")
	Pattern string
	// Dir is the base directory to scan from
	Dir string
	// IgnorePaths contains path patterns to exclude during scanning
	IgnorePaths []string
}

// Option is a function that configures the Scanner
type Option func(*Config)

// WithPattern sets the package pattern to scan
func WithPattern(pattern string) Option {
	return func(c *Config) {
		c.Pattern = pattern
	}
}

// WithDir sets the base directory to scan from
func WithDir(dir string) Option {
	return func(c *Config) {
		c.Dir = dir
	}
}

// WithIgnorePaths sets path patterns to ignore during scanning
func WithIgnorePaths(paths ...string) Option {
	return func(c *Config) {
		c.IgnorePaths = append(c.IgnorePaths, paths...)
	}
}

// ScanResult holds the results of scanning packages
type ScanResult struct {
	// Fset is the file set used during parsing
	Fset *token.FileSet
	// Files maps file paths to their parsed AST
	Files map[string]*ast.File
	// Packages contains the loaded package information
	Packages []*packages.Package
}

// Scanner scans Go packages using golang.org/x/tools/go/packages
type Scanner struct {
	config *Config
}

// New creates a new Scanner with the given options
func New(opts ...Option) *Scanner {
	config := &Config{
		Pattern:     "./...",
		Dir:         ".",
		IgnorePaths: []string{},
	}

	for _, opt := range opts {
		opt(config)
	}

	return &Scanner{
		config: config,
	}
}

// NewWithConfig creates a new Scanner with explicit configuration
func NewWithConfig(config *Config) *Scanner {
	if config.Pattern == "" {
		config.Pattern = "./..."
	}
	if config.Dir == "" {
		config.Dir = "."
	}

	return &Scanner{
		config: config,
	}
}

// Scan scans all packages matching the configured pattern
func (s *Scanner) Scan() (*ScanResult, error) {
	fset := token.NewFileSet()

	cfg := &packages.Config{
		Mode:  packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedTypes | packages.NeedSyntax,
		Dir:   s.config.Dir,
		Fset:  fset,
		Tests: false,
	}

	pkgs, err := packages.Load(cfg, s.config.Pattern)
	if err != nil {
		return nil, &ScanError{
			Message: "failed to load packages",
			Pattern: s.config.Pattern,
			Cause:   err,
		}
	}

	result := &ScanResult{
		Fset:     fset,
		Files:    make(map[string]*ast.File),
		Packages: make([]*packages.Package, 0),
	}

	for _, pkg := range pkgs {
		// Check for package errors
		if packages.PrintErrors([]*packages.Package{pkg}) > 0 {
			continue // Skip packages with errors
		}

		// Check if package should be ignored
		if s.shouldIgnorePath(pkg.PkgPath) {
			continue
		}

		result.Packages = append(result.Packages, pkg)

		// Process each file in the package
		for i, file := range pkg.Syntax {
			if i >= len(pkg.GoFiles) {
				continue
			}

			filePath := pkg.GoFiles[i]

			// Check if file should be ignored
			if s.shouldIgnorePath(filePath) {
				continue
			}

			result.Files[filePath] = file
		}
	}

	return result, nil
}

// ScanFiles returns just the file map (for backward compatibility with existing builder)
func (s *Scanner) ScanFiles() (map[string]*ast.File, *token.FileSet, error) {
	result, err := s.Scan()
	if err != nil {
		return nil, nil, err
	}

	return result.Files, result.Fset, nil
}

// shouldIgnorePath checks if a path matches any ignore pattern
func (s *Scanner) shouldIgnorePath(path string) bool {
	for _, pattern := range s.config.IgnorePaths {
		// Try exact match first
		if strings.Contains(path, strings.TrimSuffix(pattern, "/**")) {
			return true
		}

		// Try glob match
		matched, err := filepath.Match(pattern, path)
		if err == nil && matched {
			return true
		}

		// Check if path starts with the pattern (directory prefix)
		cleanPattern := strings.TrimSuffix(strings.TrimSuffix(pattern, "/**"), "/*")
		if strings.HasPrefix(path, cleanPattern) {
			return true
		}
	}

	return false
}

// Config returns the scanner configuration
func (s *Scanner) Config() *Config {
	return s.config
}
