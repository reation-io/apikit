package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/reation-io/apikit/pkg/generator/codegen"
	_ "github.com/reation-io/apikit/pkg/generator/extractors"
	"github.com/reation-io/apikit/pkg/generator/parser"
	"github.com/spf13/cobra"
)

var (
	sourceFile string
	outputFile string
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate HTTP handler wrappers",
	Long: `Generate HTTP handler wrappers from annotated Go functions.

When called from //go:generate, it automatically detects the source file
from the GOFILE environment variable.

You can also specify a source file explicitly using the --file flag.

Examples:
  # From go:generate (automatic)
  //go:generate apikit generate

  # Explicit file
  apikit generate --file handlers.go

  # With verbose output
  apikit generate --verbose

  # Dry run (show output without writing)
  apikit generate --dry-run`,
	RunE: runGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVarP(&sourceFile, "file", "f", "", "source file to process (defaults to GOFILE env var)")
	generateCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file (defaults to <source>_apikit.go)")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Determine source file
	if sourceFile == "" {
		// When called from go:generate, GOFILE env var contains the source file
		sourceFile = os.Getenv("GOFILE")
		if sourceFile == "" {
			return fmt.Errorf("no source file specified\n" +
				"Use --file flag or call from //go:generate directive:\n" +
				"  //go:generate apikit generate")
		}
	}

	// Get current directory (where go generate was called)
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	sourceFilePath := filepath.Join(cwd, sourceFile)

	// Check if source file exists
	if _, err := os.Stat(sourceFilePath); os.IsNotExist(err) {
		return fmt.Errorf("source file not found: %s", sourceFilePath)
	}

	if verbose {
		log.Printf("Processing file: %s", sourceFilePath)
	}

	// Run generation
	if err := generate(sourceFilePath); err != nil {
		return err
	}

	if verbose {
		log.Println("Generation completed successfully")
	}

	return nil
}

func generate(sourceFilePath string) error {
	// Create parser
	p := parser.New()

	// Parse the source file
	if verbose {
		log.Printf("Parsing %s...", sourceFilePath)
	}

	result, err := p.ParseFile(sourceFilePath)
	if err != nil {
		return fmt.Errorf("parsing file: %w", err)
	}

	// Print warnings if any
	if len(result.Warnings) > 0 && verbose {
		for _, warning := range result.Warnings {
			log.Printf("Warning: %s", warning)
		}
	}

	// Check if any handlers were found
	if len(result.Handlers) == 0 {
		if verbose {
			log.Println("No handlers found with //apikit:handler comment")
		}
		return nil
	}

	if verbose {
		log.Printf("Found %d handler(s):", len(result.Handlers))
		for _, h := range result.Handlers {
			log.Printf("  - %s", h.Name)
			if h.HasResponseWriter {
				log.Printf("    → with http.ResponseWriter")
			}
			if h.HasRequest {
				log.Printf("    → with *http.Request")
			}
		}
	}

	// Create generator
	gen, err := codegen.New()
	if err != nil {
		return fmt.Errorf("creating generator: %w", err)
	}

	// Generate code
	if verbose {
		log.Println("Generating wrapper code...")
	}

	code, err := gen.Generate(result)
	if err != nil {
		return fmt.Errorf("generating code: %w", err)
	}

	// Determine output file name
	output := outputFile
	if output == "" {
		output = strings.TrimSuffix(sourceFilePath, ".go") + "_apikit.go"
	}

	if dryRun {
		fmt.Printf("Would write to %s:\n", output)
		fmt.Println(string(code))
		return nil
	}

	// Write output file
	if verbose {
		log.Printf("Writing %s...", output)
	}

	if err := os.WriteFile(output, code, 0644); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}

	if verbose {
		log.Printf("Successfully generated %s", output)
	}

	return nil
}
