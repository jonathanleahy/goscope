# Go Scope Extractor

A tool to extract Go code with its dependencies for review and understanding. Unlike traditional code navigation tools that focus on compilation, this tool is designed specifically for **code review and comprehension**.

## Key Features

- **Review-Focused**: Extracts readable code with functional context, not necessarily compilable code
- **Depth-Limited Extraction**: Control how deep to traverse dependencies (0=target only, 1=direct deps, 2=transitive deps, etc.)
- **Multiple Output Formats**: Markdown (default), JSON, HTML
- **Smart Dependency Resolution**: Uses BFS traversal to gather dependencies systematically
- **Documentation Included**: Preserves documentation comments
- **External Reference Tracking**: Lists external packages used

## Installation

```bash
go build -o bin/go-scope ./cmd/go-scope
```

## Usage

```bash
# Extract a function with direct dependencies
go-scope -file=pkg/math/add.go -line=42 -depth=1

# Extract target only (no dependencies)
go-scope -file=pkg/math/add.go -line=42 -depth=0

# Extract with dependencies of dependencies (depth 2)
go-scope -file=pkg/math/add.go -line=42 -depth=2

# Save output to file
go-scope -file=pkg/math/add.go -line=42 -output=extract.md

# Verbose mode
go-scope -file=pkg/math/add.go -line=42 -verbose
```

## Command Line Options

```
  -file string
        Source file to extract from (required)
  -line int
        Line number of target symbol (required)
  -col int
        Column number (default: 1)
  -depth int
        Dependency depth (0=target only, 1=direct deps, etc) (default: 1)
  -format string
        Output format: markdown, json, html (default: "markdown")
  -output string
        Output file (default: stdout)
  -verbose
        Show verbose output
```

## Example

Given this code:

```go
// pkg/math/add.go
func Add(a, b int) int {
    if !validateInputs(a, b) {
        fmt.Println("invalid inputs")
        return 0
    }
    return a + b
}

// pkg/math/util.go
func validateInputs(a, b int) bool {
    return a >= 0 && b >= 0
}
```

Running:
```bash
cd examples/ex1
../../bin/go-scope -file=pkg/math/add.go -line=7 -depth=1
```

Produces a markdown extract with:
- The `Add` function code and documentation
- The `validateInputs` dependency code
- External references to `fmt.Println`
- Metadata and location information

## Architecture

The tool follows a clean architecture with three main phases:

1. **Symbol Location** (`internal/extract/locator.go`)
   - Loads Go packages using `golang.org/x/tools/go/packages`
   - Finds target symbol at specified file:line:column
   - Extracts symbol information (name, kind, code, documentation)

2. **Dependency Collection** (`internal/extract/collector.go`)
   - Performs depth-limited BFS traversal
   - Collects internal and external dependencies
   - Deduplicates symbols to avoid repetition
   - Tracks depth level for each dependency

3. **Formatting** (`internal/extract/format/`)
   - Markdown formatter with syntax highlighting
   - Groups dependencies by depth
   - Includes metadata and statistics
   - Future: JSON and HTML formatters

## Development

Built using Test-Driven Development (TDD) with comprehensive test coverage:

```bash
# Run tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run specific package tests
go test ./internal/extract -v
go test ./internal/extract/format -v
```

Test coverage:
- Symbol locator: 75.3%
- Dependency collector: 75.3%
- Markdown formatter: 78.8%

## Project Structure

```
.
├── cmd/
│   └── go-scope/          # CLI entry point
│       └── main.go
├── internal/
│   ├── extract/           # Core extraction logic
│   │   ├── locator.go     # Symbol location
│   │   ├── collector.go   # Dependency collection
│   │   ├── api.go         # Public API
│   │   ├── helpers.go     # Internal types
│   │   └── format/        # Output formatters
│   │       └── markdown.go
│   └── types/             # Shared type definitions
│       └── types.go
├── examples/
│   └── ex1/               # Example Go project for testing
│       └── pkg/math/
│           ├── add.go
│           └── util.go
├── docs/                  # Documentation
│   ├── SPEC_v2_REVIEW_FOCUSED.md
│   ├── QUICK_START.md
│   └── ...
└── bin/                   # Built binaries
    └── go-scope
```

## Design Principles

1. **Review-Focused, Not Compilation-Focused**
   - Goal is readability and understanding, not compilation
   - Extracts code in context with dependencies
   - Preserves documentation and structure

2. **Depth-Limited Traversal**
   - User controls how much context to include
   - BFS ensures systematic dependency gathering
   - Deduplication prevents repetition

3. **Clean Architecture**
   - Separation of concerns (locate, collect, format)
   - No circular dependencies between packages
   - Shared types in separate package

4. **Test-Driven Development**
   - Tests written before implementation
   - High test coverage (75-78%)
   - RED-GREEN-REFACTOR cycle

## Future Enhancements (Phase 2)

- [ ] Interactive web visualizer (see `docs/PHASE_2_VISUALIZER.md`)
- [ ] JSON output format
- [ ] HTML output format
- [ ] Caller analysis (reverse dependencies)
- [ ] Complexity metrics (cyclomatic complexity)
- [ ] Git blame integration
- [ ] Test function inclusion
- [ ] Context lines around code

## Documentation

See `docs/` directory for detailed documentation:
- `SPEC_v2_REVIEW_FOCUSED.md` - Complete technical specification
- `QUICK_START.md` - Quick start guide with examples
- `IMPLEMENTATION_ROADMAP.md` - Development roadmap
- `PHASE_2_VISUALIZER.md` - Future visualizer plans

## License

MIT

## Contributing

This project was built using Test-Driven Development. When contributing:
1. Write tests first (RED)
2. Implement to make tests pass (GREEN)
3. Refactor for clarity (REFACTOR)
4. Ensure test coverage remains high

## Status

✅ **Phase 1 Complete** - Core extraction functionality working with CLI

Current capabilities:
- ✅ Symbol location at file:line:column
- ✅ Depth-limited BFS dependency collection
- ✅ Markdown formatting
- ✅ CLI tool
- ✅ External reference tracking
- ✅ Documentation preservation
- ✅ High test coverage (75-78%)

Next: Phase 2 - Interactive visualizer and additional output formats
