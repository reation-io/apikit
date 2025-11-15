package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	coreast "github.com/reation-io/apikit/core/ast"
	"github.com/reation-io/apikit/openapi/builder"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	openapiOutput    string
	openapiFormat    string
	openapiTitle     string
	openapiVer       string
	openapiMultiSpec bool   // Enable multi-spec mode
	openapiOutputDir string // Output directory for multi-spec mode
)

// openapiCmd represents the openapi command
var openapiCmd = &cobra.Command{
	Use:     "openapi [files...]",
	Aliases: []string{"swagger"},
	Short:   "Generate OpenAPI specification from Go source files",
	Long: `Generate OpenAPI 3.0 specification from Go source files with swagger annotations.

The command scans Go files for swagger:meta, swagger:route, and swagger:model
directives and generates a complete OpenAPI specification.

Supported directives:
  • swagger:meta    - API metadata (title, version, description, etc.)
  • swagger:route   - API endpoints (paths and operations)
  • swagger:model   - Data models (schemas)

Examples:
  # Generate from all Go files in current directory
  apikit openapi *.go

  # Generate from specific files
  apikit openapi handlers.go models.go

  # Generate with custom output file
  apikit openapi --output openapi.json *.go

  # Generate YAML output
  apikit openapi --format yaml --output openapi.yaml *.go

  # Override API metadata
  apikit openapi --title "My API" --version "2.0.0" *.go`,
	RunE: runOpenAPI,
}

func init() {
	rootCmd.AddCommand(openapiCmd)

	openapiCmd.Flags().StringVarP(&openapiOutput, "output", "o", "openapi.json", "output file path (single-spec mode)")
	openapiCmd.Flags().StringVarP(&openapiFormat, "format", "f", "json", "output format (json or yaml)")
	openapiCmd.Flags().StringVar(&openapiTitle, "title", "", "override API title")
	openapiCmd.Flags().StringVar(&openapiVer, "version", "", "override API version")
	openapiCmd.Flags().BoolVar(&openapiMultiSpec, "multi-spec", false, "generate multiple spec files based on Spec: tags")
	openapiCmd.Flags().StringVar(&openapiOutputDir, "output-dir", ".", "output directory for multi-spec mode")
}

func runOpenAPI(cmd *cobra.Command, args []string) error {
	// Validate format
	if openapiFormat != "json" && openapiFormat != "yaml" {
		return fmt.Errorf("invalid format %q, must be 'json' or 'yaml'", openapiFormat)
	}

	// Collect source files
	var sourceFiles []string

	if len(args) > 0 {
		// Use provided arguments
		sourceFiles = args
	} else {
		// Default to all Go files in current directory
		matches, err := filepath.Glob("*.go")
		if err != nil {
			return fmt.Errorf("failed to find Go files: %w", err)
		}
		sourceFiles = matches
	}

	if len(sourceFiles) == 0 {
		return fmt.Errorf("no Go files found\nUsage: apikit openapi [files...]")
	}

	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	// Resolve all source files
	var resolvedFiles []string
	for _, file := range sourceFiles {
		filePath := filepath.Join(cwd, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("source file not found: %s", filePath)
		}
		resolvedFiles = append(resolvedFiles, filePath)
	}

	if verbose {
		log.Printf("Processing %d file(s)...", len(resolvedFiles))
	}

	// Parse all files with generic parser
	genericParser := coreast.NewCachedParser()
	var parseResults []*coreast.ParseResult

	for i, sourceFilePath := range resolvedFiles {
		if verbose {
			log.Printf("[%d/%d] Parsing %s", i+1, len(resolvedFiles), sourceFilePath)
		}

		result, err := genericParser.Parse(sourceFilePath)
		if err != nil {
			return fmt.Errorf("parsing %s: %w", sourceFilePath, err)
		}

		parseResults = append(parseResults, result)
	}

	// Extract OpenAPI specification(s)
	if openapiMultiSpec {
		// Multi-spec mode
		if verbose {
			log.Println("Extracting multiple OpenAPI specifications...")
		}

		specs, err := builder.ExtractMultipleFromGeneric(parseResults)
		if err != nil {
			return fmt.Errorf("extracting OpenAPI specs: %w", err)
		}

		// Override metadata if provided
		if openapiTitle != "" || openapiVer != "" {
			for _, spec := range specs {
				if openapiTitle != "" {
					spec.Info.Title = openapiTitle
				}
				if openapiVer != "" {
					spec.Info.Version = openapiVer
				}
			}
		}

		// Write each spec to its own file
		for specName, spec := range specs {
			// Skip empty specs (no routes)
			if len(spec.Paths.PathItems) == 0 {
				if verbose {
					log.Printf("Skipping empty spec: %s", specName)
				}
				continue
			}

			// Determine output filename
			var ext string
			if openapiFormat == "yaml" {
				ext = ".yml"
			} else {
				ext = ".json"
			}
			filename := filepath.Join(openapiOutputDir, specName+ext)

			// Marshal to requested format
			var output []byte
			if openapiFormat == "yaml" {
				output, err = yaml.Marshal(spec)
				if err != nil {
					return fmt.Errorf("marshaling %s to YAML: %w", specName, err)
				}
			} else {
				output, err = json.MarshalIndent(spec, "", "  ")
				if err != nil {
					return fmt.Errorf("marshaling %s to JSON: %w", specName, err)
				}
			}

			// Write output
			if err := os.WriteFile(filename, output, 0644); err != nil {
				return fmt.Errorf("writing %s: %w", filename, err)
			}

			fmt.Printf("✓ Generated %s specification: %s\n", specName, filename)
			if verbose {
				log.Printf("  Format: %s", openapiFormat)
				log.Printf("  Title: %s", spec.Info.Title)
				log.Printf("  Version: %s", spec.Info.Version)
				log.Printf("  Paths: %d", len(spec.Paths.PathItems))
				if spec.Components != nil && spec.Components.Schemas != nil {
					log.Printf("  Schemas: %d", len(spec.Components.Schemas))
				}
			}
		}
	} else {
		// Single-spec mode (default, backward compatible)
		if verbose {
			log.Println("Extracting OpenAPI specification...")
		}

		spec, err := builder.ExtractFromGeneric(parseResults)
		if err != nil {
			return fmt.Errorf("extracting OpenAPI spec: %w", err)
		}

		// Override metadata if provided
		if openapiTitle != "" {
			spec.Info.Title = openapiTitle
		}
		if openapiVer != "" {
			spec.Info.Version = openapiVer
		}

		// Marshal to requested format
		var output []byte
		if openapiFormat == "yaml" {
			output, err = yaml.Marshal(spec)
			if err != nil {
				return fmt.Errorf("marshaling to YAML: %w", err)
			}
		} else {
			output, err = json.MarshalIndent(spec, "", "  ")
			if err != nil {
				return fmt.Errorf("marshaling to JSON: %w", err)
			}
		}

		// Write output
		if err := os.WriteFile(openapiOutput, output, 0644); err != nil {
			return fmt.Errorf("writing output file: %w", err)
		}

		fmt.Printf("✓ Generated OpenAPI specification: %s\n", openapiOutput)
		if verbose {
			log.Printf("  Format: %s", openapiFormat)
			log.Printf("  Title: %s", spec.Info.Title)
			log.Printf("  Version: %s", spec.Info.Version)
			log.Printf("  Paths: %d", len(spec.Paths.PathItems))
			if spec.Components != nil && spec.Components.Schemas != nil {
				log.Printf("  Schemas: %d", len(spec.Components.Schemas))
			}
		}
	}

	return nil
}
