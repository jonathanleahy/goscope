package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/extract-scope-go/go-scope/internal/extract"
	"github.com/extract-scope-go/go-scope/internal/types"
)

func main() {
	// Define flags
	var (
		file    = flag.String("file", "", "Source file to extract from (required)")
		line    = flag.Int("line", 0, "Line number of target symbol (required)")
		col     = flag.Int("col", 1, "Column number (default: 1)")
		depth   = flag.Int("depth", 1, "Dependency depth (0=target only, 1=direct deps, etc)")
		format  = flag.String("format", "markdown", "Output format: markdown, json, html")
		output  = flag.String("output", "", "Output file (default: stdout)")
		verbose = flag.Bool("verbose", false, "Show verbose output")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Extract Go code with dependencies for review and understanding.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Extract function at line 42 with direct dependencies\n")
		fmt.Fprintf(os.Stderr, "  %s -file=pkg/math/add.go -line=42\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Extract with depth 2 (dependencies of dependencies)\n")
		fmt.Fprintf(os.Stderr, "  %s -file=pkg/math/add.go -line=42 -depth=2\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Save output to file\n")
		fmt.Fprintf(os.Stderr, "  %s -file=pkg/math/add.go -line=42 -output=extract.md\n\n", os.Args[0])
	}

	flag.Parse()

	// Validate required flags
	if *file == "" || *line == 0 {
		flag.Usage()
		os.Exit(1)
	}

	// Get current working directory as root
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Make file path absolute if relative
	absFile := *file
	if !filepath.IsAbs(*file) {
		absFile = filepath.Join(root, *file)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "Root: %s\n", root)
		fmt.Fprintf(os.Stderr, "File: %s\n", absFile)
		fmt.Fprintf(os.Stderr, "Line: %d, Column: %d\n", *line, *col)
		fmt.Fprintf(os.Stderr, "Depth: %d\n", *depth)
		fmt.Fprintf(os.Stderr, "Format: %s\n", *format)
	}

	// Create target and options
	target := types.Target{
		Root:   root,
		File:   absFile,
		Line:   *line,
		Column: *col,
	}

	opts := types.Options{
		Depth:  *depth,
		Format: *format,
	}

	// Extract and format
	ctx := context.Background()
	result, err := extract.ExtractAndFormat(ctx, target, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Write output
	var writer *os.File
	if *output == "" {
		writer = os.Stdout
	} else {
		writer, err = os.Create(*output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to create output file: %v\n", err)
			os.Exit(1)
		}
		defer writer.Close()
	}

	fmt.Fprint(writer, result.Rendered)

	if *verbose && *output != "" {
		fmt.Fprintf(os.Stderr, "Output written to: %s\n", *output)
		fmt.Fprintf(os.Stderr, "Total symbols: %d\n", result.Metadata.TotalSymbols)
	}
}
