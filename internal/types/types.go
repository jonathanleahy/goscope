package types

import (
	"time"
)

// Target specifies what to extract
type Target struct {
	Root   string // Module root path
	File   string // Source file path (relative or absolute)
	Line   int    // 1-based line number
	Column int    // 1-based column (default: 1)
}

// Options configures extraction behavior
type Options struct {
	Depth          int    // Dependency depth (default: 1, 0 = target only)
	Format         string // "markdown", "html", "json" (default: "markdown")
	StubExternal   bool   // Show signatures for external deps (default: true)
	ShowCallers    bool   // Include reverse dependencies (default: false)
	ShowTests      bool   // Include test functions (default: false)
	ContextLines   int    // Extra lines around target (default: 0)
	Annotate       bool   // Add inline reference comments (default: true)
	IncludeMetrics bool   // Compute complexity metrics (default: false)
	GitBlame       bool   // Include git history (default: false)
}

// Symbol represents a Go symbol (function, type, var, etc.)
type Symbol struct {
	Package  string // Full package path
	Name     string // Symbol name
	Kind     string // "func", "method", "type", "var", "const", "interface"
	Receiver string // For methods: receiver type
	File     string // Source file path
	Line     int    // Start line
	EndLine  int    // End line
	Column   int    // Start column
	Code     string // Source code
	Doc      string // Documentation comment
	Exported bool   // Whether symbol is exported
}

// Reference represents a dependency
type Reference struct {
	Symbol       Symbol // Referenced symbol
	Reason       string // "direct-call", "type-reference", "field-access", etc.
	Depth        int    // 0 = target, 1 = direct dep, etc.
	External     bool   // True if from different module
	Stub         bool   // True if only signature included
	Signature    string // For stubs: type signature
	ReferencedBy string // Which symbol references this
}

// Caller represents a reverse dependency
type Caller struct {
	File     string // Source file
	Line     int    // Call site line
	Function string // Containing function name
	Context  string // Code snippet around call
}

// Metrics represents code complexity metrics
type Metrics struct {
	LinesOfCode          int
	LogicalLines         int
	CyclomaticComplexity int
	DependencyCount      int
	DirectDeps           int
	TransitiveDeps       int
	ExternalPackages     []string
}

// GitBlame represents git history for a line/symbol
type GitBlame struct {
	Commit  string    // Commit hash
	Author  string    // Author email
	Date    time.Time // Commit date
	Message string    // Commit message
}

// Extract represents the extraction result
type Extract struct {
	Target     Symbol      // The requested symbol
	References []Reference // Included dependencies
	External   []string    // External package references (pkg.Symbol format)
	Callers    []Caller    // What calls this symbol
	Metrics    *Metrics    // Optional metrics
	GitHistory []GitBlame  // Optional git history
	Graph      string      // Dependency graph (mermaid or text format)
}

// Result is the final output
type Result struct {
	Extract  Extract  // Structured extract
	Rendered string   // Formatted output (markdown/HTML)
	Metadata Metadata // Extraction metadata
}

// Metadata contains extraction information
type Metadata struct {
	ExtractedAt  time.Time
	GoVersion    string
	Module       string
	ModuleVersion string
	TotalSymbols int
	TotalLines   int
	Options      Options
}

