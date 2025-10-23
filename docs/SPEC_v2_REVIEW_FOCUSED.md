# Go Scoped Extractor - Specification v2.0 (Review-Focused)

**Version**: 2.0.0
**Date**: 2025-10-23
**Status**: Draft - For Approval
**Replaces**: Original "Compilation-Focused" Specification

---

## 1. Goal

Given `(repoRoot, filePath, lineNumber)` in a Go project, **extract the symbol at that location along with configurable context depth for code review**. Output includes the target code, referenced symbols with annotations, caller locations, and formatted presentation **optimized for human readability and understanding**.

**Non-Goal**: The extracted code does **not** need to compile. It's for review/inspection purposes.

---

## 2. Deliverables

1. **CLI Tool**: `go-scope` binary
2. **Library**: `internal/extract` with pure functions
3. **Examples**: Three small Go projects under `/examples/`
4. **Tests**: Unit tests, integration tests, BDD scenarios
5. **Documentation**: README, usage guide, API docs

---

## 3. Inputs and Outputs

### 3.1 Input

**Required**:
- `--root` - Path to Go module root
- `--file` - Path to source file (relative to root or absolute)
- `--line` - 1-based line number

**Optional**:
- `--column` - 1-based column number (default: 1)
- `--depth N` - Dependency traversal depth (default: 1, 0 = target only)
- `--format markdown|html|json` - Output format (default: markdown)
- `--stub-external` - Show signatures for external dependencies (default: true)
- `--show-callers` - Include reverse dependencies (default: false)
- `--show-tests` - Include test functions (default: false)
- `--context-lines N` - Extra lines before/after target (default: 0)
- `--annotate` - Add inline reference comments with file:line (default: true)
- `--metrics` - Include complexity metrics (default: false)
- `--git-blame` - Show git history for symbol (default: false)
- `--output PATH` - Write to file instead of stdout

**Example**:
```bash
# Basic extraction
go-scope --root . --file pkg/service/user.go --line 128

# Comprehensive review
go-scope --root . --file pkg/service/user.go --line 128 \
  --depth 2 \
  --format html \
  --show-callers \
  --metrics \
  --git-blame \
  --output review.html
```

### 3.2 Output

#### Markdown Format (Default)

Structured document with clear sections:

```markdown
# Code Extract: SymbolName

**File**: pkg/service/user.go:128
**Package**: github.com/example/app/pkg/service
**Kind**: Method
**Extracted**: 2025-10-23 14:32:00

## Metrics (if --metrics)
- Lines of Code: 45
- Cyclomatic Complexity: 8
- Direct Dependencies: 4
- External Packages: 3

---

## Target Symbol

[source code with file:line annotations]

---

## Direct Dependencies (depth 1)

### SymbolName1
**File**: path/to/file.go:line
[code or stub]

### SymbolName2
**File**: path/to/file.go:line
[code or stub]

---

## External References
- package/path.Symbol
- another/package.AnotherSymbol

---

## Called By (if --show-callers)
- file.go:line - functionName
- test.go:line - TestFunction

---

## Recent Changes (if --git-blame)
- 2025-10-20 by user - "commit message"

---

## Dependency Graph
[text tree or mermaid diagram]
```

#### HTML Format

Same content structure with:
- Syntax highlighting (chroma or highlight.js)
- Clickable file:line links
- Collapsible sections
- Copy buttons for code blocks
- Dark/light theme toggle
- Search within extract
- Self-contained single file (inline CSS/JS)

#### JSON Format

Structured machine-readable output:

```json
{
  "target": {
    "file": "pkg/service/user.go",
    "line": 128,
    "column": 1,
    "package": "github.com/example/app/pkg/service",
    "symbol": "CreateUser",
    "kind": "method",
    "receiver": "UserService",
    "code": "func (s *UserService) CreateUser...",
    "doc": "CreateUser creates a new user account"
  },
  "options": {
    "depth": 2,
    "stub_external": true,
    "show_callers": true,
    "metrics": true
  },
  "dependencies": [
    {
      "symbol": "validateEmail",
      "package": "github.com/example/app/pkg/service",
      "file": "pkg/service/validation.go",
      "line": 45,
      "kind": "function",
      "depth": 1,
      "external": false,
      "stub": false,
      "code": "func validateEmail(email string) error...",
      "referenced_by": "target",
      "reference_type": "direct-call"
    },
    {
      "symbol": "DB.Create",
      "package": "database/sql",
      "file": "",
      "line": 0,
      "kind": "method",
      "depth": 1,
      "external": true,
      "stub": true,
      "signature": "Create(value interface{}) error",
      "referenced_by": "target",
      "reference_type": "method-call"
    }
  ],
  "external_references": [
    {"package": "database/sql", "symbols": ["DB.Create"]},
    {"package": "fmt", "symbols": ["Errorf"]},
    {"package": "time", "symbols": ["Now"]}
  ],
  "callers": [
    {
      "file": "cmd/api/handlers.go",
      "line": 67,
      "function": "handleCreateUser",
      "context": "user, err := service.CreateUser(ctx, req.Email, req.Name)"
    }
  ],
  "metrics": {
    "lines_of_code": 45,
    "cyclomatic_complexity": 8,
    "dependency_count": 4,
    "external_packages": ["database/sql", "fmt", "time"]
  },
  "git_history": [
    {
      "commit": "abc123",
      "author": "alice@example.com",
      "date": "2025-10-20T10:30:00Z",
      "message": "Add email validation"
    }
  ],
  "graph": {
    "format": "mermaid",
    "content": "graph TD\n  CreateUser --> validateEmail\n  ..."
  },
  "metadata": {
    "extracted_at": "2025-10-23T14:32:00Z",
    "go_version": "1.22.0",
    "module": "github.com/example/app",
    "total_symbols": 8,
    "total_lines": 156
  }
}
```

---

## 4. Definition of "Scoped Code"

### 4.1 Target Symbol Identification

Start from the position `(file, line, column)`:

- **Inside function/method body** → Extract that function/method
- **On type declaration** → Extract that type definition
- **On interface method** → Extract interface definition
- **On var/const declaration** → Extract that declaration
- **On package-level function** → Extract that function
- **On struct field** → Extract containing struct
- **On method receiver** → Extract that method

### 4.2 Dependencies to Include

**Based on `--depth` setting**:

- **Depth 0**: Target symbol only
- **Depth 1**: Target + directly referenced symbols
- **Depth 2**: Target + direct refs + their refs
- **Depth N**: Transitive closure to N levels

**What counts as a reference**:

1. **Function/method calls**: `foo()`, `obj.Method()`
2. **Type references**: Variable declarations, parameters, return types, struct fields
3. **Selector expressions**: `pkg.Symbol`, `obj.field`
4. **Composite literals**: `Type{...}` requires Type definition
5. **Type assertions/switches**: `x.(Type)` requires Type
6. **Embedded types**: Struct embedding, interface embedding
7. **Generic instantiations**: `Func[T]()` requires T's constraint definitions
8. **Const/var initializers**: RHS expressions' symbols

**Resolution rules**:

- **Same package**: Include complete definition
- **Same module, exported**: Include complete definition
- **Same module, unexported**: Include if accessible, otherwise note as unavailable
- **External module**: Stub signature only (if `--stub-external`)
- **Standard library**: Note reference, don't include implementation

### 4.3 What's Excluded

- Unused functions/methods in same file
- Unused methods on included types (unless called)
- Test files (unless `--show-tests`)
- Code beyond `--depth` limit
- Generated code comments (e.g., `// Code generated by...`) - preserved but not traversed
- Unexported symbols from external packages (noted but not included)

---

## 5. Algorithm

### 5.1 Load Module

```go
cfg := &packages.Config{
    Mode: packages.NeedName |
          packages.NeedFiles |
          packages.NeedCompiledGoFiles |
          packages.NeedSyntax |
          packages.NeedTypes |
          packages.NeedTypesInfo |
          packages.NeedDeps |
          packages.NeedModule,
    Dir: rootPath,
    Env: os.Environ(),
}
pkgs, err := packages.Load(cfg, "./...")
```

### 5.2 Locate Target Symbol

1. **Find file's package** from loaded packages
2. **Parse position** using `token.FileSet`
3. **Find enclosing node** using `astutil.PathEnclosingInterval(file, pos, pos)`
4. **Identify symbol**:
   - For `*ast.FuncDecl`: Function or method
   - For `*ast.TypeSpec`: Type definition
   - For `*ast.ValueSpec`: Var or const
   - For `*ast.Ident` in other contexts: Resolve via `types.Info.Uses[ident]`
5. **Get `types.Object`** for the symbol

### 5.3 Collect Dependencies (Depth-Limited BFS)

```
Initialize:
  queue = [(targetSymbol, 0)]  // (symbol, depth)
  visited = {}
  result = []

While queue not empty:
  (sym, depth) = queue.pop()

  If sym in visited:
    continue

  visited.add(sym)
  result.append((sym, depth))

  If depth >= maxDepth:
    continue  // Don't traverse deeper

  // Find all references in sym's code
  refs = findReferences(sym)

  For each ref in refs:
    refSym = resolveSymbol(ref)

    If refSym is external:
      If stub_external:
        result.append((refSym, depth+1, isStub=true))
    Else:
      queue.append((refSym, depth+1))

Return result
```

**Reference Finding**:
- Walk AST of symbol's declaration
- For each `*ast.Ident`, check `types.Info.Uses[ident]`
- For each `*ast.SelectorExpr`, check `types.Info.Selections[sel]`
- Track source location for annotations

### 5.4 Find Callers (if --show-callers)

```
Build reverse index:
  For each package in module:
    For each file in package:
      Walk AST:
        For each call expression:
          Resolve callee to types.Object
          If callee == targetSymbol:
            Record (file, line, enclosingFunction)

Return list of caller locations
```

### 5.5 Compute Metrics (if --metrics)

**Cyclomatic Complexity**:
```
complexity = 1
For each branch (if, for, switch, case, &&, ||, ?:):
  complexity += 1
```

**Lines of Code**:
- Physical: Count newlines in source
- Logical: Count statements (excluding braces, comments)

**Dependencies**:
- Direct: Count depth-1 symbols
- Transitive: Count all included symbols
- External: Count unique external packages

### 5.6 Git Integration (if --git-blame)

```bash
# Last author per line
git blame -L <start>,<end> <file>

# Recent changes
git log -p --follow -n 5 -- <file>

# Uncommitted changes
git diff <file>
```

### 5.7 Format Output

**Markdown**:
1. Generate header with metadata
2. For each symbol, format as fenced code block with annotations
3. Group by dependency level
4. Add external references section
5. Add callers section
6. Add git history section
7. Generate dependency graph (text tree or mermaid)

**HTML**:
1. Same structure as markdown
2. Apply syntax highlighting (chroma)
3. Convert file:line annotations to `<a>` tags
4. Add interactive JavaScript for collapsing/searching
5. Inline CSS for styling
6. Add copy buttons to code blocks

**JSON**:
1. Build structured object per schema in section 3.2
2. Marshal with pretty-printing

---

## 6. Edge Cases

### 6.1 Multiple Symbols at Same Position

Example: Multiple methods on same line in interface

```go
type Writer interface {
    Write([]byte) (int, error); Close() error  // Two methods on one line
}
```

**Resolution**: Use column to disambiguate, or choose first if column not specified

### 6.2 Generic Code

```go
func Map[T, U any](slice []T, f func(T) U) []U { ... }
```

**Handling**:
- Include type parameters in symbol definition
- Include constraint definitions (e.g., `any`, custom constraints)
- Show instantiation examples in callers if available

### 6.3 Interface Methods

```go
type Service interface {
    DoThing() error
}
```

If extracting interface method:
- Include interface definition
- With `--show-callers`: Find concrete implementations and their usage
- Don't attempt whole-program implementation search by default

### 6.4 Unexported Cross-Package Symbols

```go
// pkg/internal/helper.go
func doStuff() { ... }  // unexported

// pkg/service/user.go (different package)
func CreateUser() {
    internal.doStuff()  // Not accessible!
}
```

**Handling**:
- Note in output: "Reference to unexported symbol: pkg/internal.doStuff (unavailable)"
- Don't include code
- Suggest reviewing that package separately

### 6.5 Circular Dependencies

```go
// a.go
func Foo() { Bar() }

// b.go
func Bar() { Foo() }
```

**Handling**:
- Mark both as included (no need for topological sort)
- Note cycle in dependency graph: `Foo <--> Bar`
- Include both if within depth limit

### 6.6 Test Symbols

```go
func TestCreateUser(t *testing.T) { ... }
```

**Handling**:
- Exclude by default
- Include with `--show-tests`
- When showing callers for production code, include test callers if `--show-tests`

---

## 7. Architecture

```
go-scope/
├── cmd/
│   └── go-scope/
│       └── main.go                 // CLI entry point
├── internal/
│   └── extract/
│       ├── loader.go               // packages.Load wrapper
│       ├── locator.go              // position → symbol resolution
│       ├── collector.go            // depth-limited dependency collection
│       ├── analysis.go             // caller analysis, metrics computation
│       ├── git.go                  // git integration (blame, log)
│       └── format/
│           ├── markdown.go         // markdown formatter
│           ├── html.go             // HTML with syntax highlighting
│           └── json.go             // structured JSON output
├── pkg/
│   └── cli/
│       ├── flags.go                // flag parsing
│       ├── help.go                 // help text
│       └── version.go              // version info
├── examples/
│   ├── ex1/                        // single package example
│   ├── ex2/                        // multi-package HTTP service
│   └── ex3/                        // generics and interfaces
├── tests/
│   ├── unit/                       // unit tests
│   ├── integration/                // integration tests
│   └── features/                   // BDD scenarios (godog)
└── docs/
    ├── README.md
    ├── USAGE.md
    ├── API.md
    └── DESIGN.md
```

---

## 8. Public API (Library)

```go
package extract

import (
    "context"
    "time"
)

// Target specifies what to extract
type Target struct {
    Root   string  // Module root path
    File   string  // Source file path (relative or absolute)
    Line   int     // 1-based line number
    Column int     // 1-based column (default: 1)
}

// Options configures extraction behavior
type Options struct {
    Depth          int      // Dependency depth (default: 1, 0 = target only)
    Format         string   // "markdown", "html", "json" (default: "markdown")
    StubExternal   bool     // Show signatures for external deps (default: true)
    ShowCallers    bool     // Include reverse dependencies (default: false)
    ShowTests      bool     // Include test functions (default: false)
    ContextLines   int      // Extra lines around target (default: 0)
    Annotate       bool     // Add inline reference comments (default: true)
    IncludeMetrics bool     // Compute complexity metrics (default: false)
    GitBlame       bool     // Include git history (default: false)
}

// Symbol represents a Go symbol (function, type, var, etc.)
type Symbol struct {
    Package  string   // Full package path
    Name     string   // Symbol name
    Kind     string   // "func", "method", "type", "var", "const", "interface"
    Receiver string   // For methods: receiver type
    File     string   // Source file path
    Line     int      // Start line
    EndLine  int      // End line
    Code     string   // Source code
    Doc      string   // Documentation comment
    Exported bool     // Whether symbol is exported
}

// Reference represents a dependency
type Reference struct {
    Symbol       Symbol   // Referenced symbol
    Reason       string   // "direct-call", "type-reference", "field-access", etc.
    Depth        int      // 0 = target, 1 = direct dep, etc.
    External     bool     // True if from different module
    Stub         bool     // True if only signature included
    Signature    string   // For stubs: type signature
    ReferencedBy string   // Which symbol references this
}

// Caller represents a reverse dependency
type Caller struct {
    File     string   // Source file
    Line     int      // Call site line
    Function string   // Containing function name
    Context  string   // Code snippet around call
}

// Metrics represents code complexity metrics
type Metrics struct {
    LinesOfCode         int
    LogicalLines        int
    CyclomaticComplexity int
    DependencyCount     int
    DirectDeps          int
    TransitiveDeps      int
    ExternalPackages    []string
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
    Target     Symbol        // The requested symbol
    References []Reference   // Included dependencies
    External   []string      // External package references (pkg.Symbol format)
    Callers    []Caller      // What calls this symbol
    Metrics    *Metrics      // Optional metrics
    GitHistory []GitBlame    // Optional git history
    Graph      string        // Dependency graph (mermaid or text format)
}

// Result is the final output
type Result struct {
    Extract  Extract   // Structured extract
    Rendered string    // Formatted output (markdown/HTML)
    Metadata Metadata  // Extraction metadata
}

// Metadata contains extraction information
type Metadata struct {
    ExtractedAt time.Time
    GoVersion   string
    Module      string
    TotalSymbols int
    TotalLines   int
    Options     Options
}

// ExtractSymbol is the main entry point
func ExtractSymbol(ctx context.Context, target Target, opts Options) (*Result, error)

// Convenience functions
func ExtractToFile(ctx context.Context, target Target, opts Options, outputPath string) error
func ExtractToMarkdown(ctx context.Context, target Target, depth int) (string, error)
func ExtractToHTML(ctx context.Context, target Target, depth int) (string, error)
func ExtractToJSON(ctx context.Context, target Target, depth int) (string, error)
```

---

## 9. CLI Interface

```bash
go-scope [flags]

Required Flags:
  --root PATH           Module root path
  --file PATH           Source file path
  --line N              Line number (1-based)

Optional Flags:
  --column N            Column number (1-based, default: 1)
  --depth N             Dependency depth (default: 1, 0 = target only)
  --format FORMAT       Output format: markdown, html, json (default: markdown)
  --stub-external       Show signatures for external deps (default: true)
  --show-callers        Include reverse dependencies (default: false)
  --show-tests          Include test functions (default: false)
  --context-lines N     Extra lines around target (default: 0)
  --annotate            Add inline file:line comments (default: true)
  --metrics             Include complexity metrics (default: false)
  --git-blame           Show git history (default: false)
  --output PATH         Write to file instead of stdout
  --help                Show help
  --version             Show version

Examples:

  # Basic extraction (depth 1, markdown to stdout)
  go-scope --root . --file pkg/service/user.go --line 128

  # Deep extraction with callers
  go-scope --root . --file pkg/service/user.go --line 128 \
    --depth 3 --show-callers

  # HTML output with all features
  go-scope --root . --file pkg/api/handler.go --line 45 \
    --format html \
    --show-callers \
    --metrics \
    --git-blame \
    --output review.html

  # JSON for tool integration
  go-scope --root . --file pkg/model/user.go --line 12 \
    --format json \
    --depth 0

  # Target only (no dependencies)
  go-scope --root . --file pkg/util/math.go --line 7 --depth 0

Exit Codes:
  0 - Success
  1 - Error (invalid input, file not found, etc.)
  2 - Symbol not found at position
  3 - Package load error
```

---

## 10. Examples

### Example 1: Single Package (`ex1`)

```
examples/ex1/
├── go.mod
├── pkg/
│   └── math/
│       ├── add.go          // Add, Sub functions
│       └── util.go         // helper function used by Add
└── README.md
```

**add.go**:
```go
package math

import "fmt"

// Add returns the sum of two integers
func Add(a, b int) int {
    if !validateInputs(a, b) {  // → util.go
        fmt.Println("invalid inputs")
        return 0
    }
    return a + b
}

// Sub returns the difference
func Sub(a, b int) int {
    return a - b
}
```

**util.go**:
```go
package math

func validateInputs(a, b int) bool {
    return a >= 0 && b >= 0
}
```

**Test Command**:
```bash
go-scope --root examples/ex1 --file pkg/math/add.go --line 6 --depth 1
```

**Expected Output**:
- Include `Add` function
- Include `validateInputs` helper (depth 1)
- Mark `fmt.Println` as external
- Exclude `Sub` (unused)

### Example 2: Multi-Package HTTP Service (`ex2`)

```
examples/ex2/
├── go.mod
├── cmd/
│   └── api/
│       └── main.go         // HTTP handlers
├── pkg/
│   ├── service/
│   │   ├── user.go         // UserService with CreateUser method
│   │   └── validation.go   // validateEmail helper
│   ├── model/
│   │   └── user.go         // User type
│   └── repo/
│       └── unused.go       // Not referenced
├── internal/
│   └── token/
│       └── jwt.go          // Internal helper
└── README.md
```

**Test Commands**:

```bash
# Extract handler at depth 0 (just the handler)
go-scope --root examples/ex2 --file cmd/api/main.go --line 30 --depth 0

# Extract handler at depth 2 (include service layer)
go-scope --root examples/ex2 --file cmd/api/main.go --line 30 --depth 2

# Extract with callers
go-scope --root examples/ex2 --file pkg/service/user.go --line 15 --show-callers
```

**Expected Outputs**:
- Depth 0: Only `handleCreateUser` code
- Depth 2: Handler + `UserService.CreateUser` + `validateEmail` + `User` type
- Callers: Should find `cmd/api/main.go` routes calling the handler

### Example 3: Generics and Interfaces (`ex3`)

```
examples/ex3/
├── go.mod
├── pkg/
│   ├── iter/
│   │   ├── map.go          // Generic Map[T, U] function
│   │   └── filter.go       // Generic Filter[T] function
│   └── io/
│       └── writer.go       // Writer interface + FileWriter impl
└── README.md
```

**map.go**:
```go
package iter

// Map applies function f to each element
func Map[T any, U any](slice []T, f func(T) U) []U {
    result := make([]U, len(slice))
    for i, v := range slice {
        result[i] = f(v)
    }
    return result
}
```

**Test Command**:
```bash
go-scope --root examples/ex3 --file pkg/iter/map.go --line 4 --depth 1
```

**Expected Output**:
- Include `Map[T, U]` function with type parameters
- Include `any` constraint (stdlib, noted as external)
- Show usage examples from callers if `--show-callers`

---

## 11. Testing Strategy

### 11.1 Unit Tests

**Package**: `internal/extract`

**Test Coverage**:

1. **loader_test.go**
   - Load valid module
   - Load multi-package module
   - Error: invalid root path
   - Error: not a Go module

2. **locator_test.go**
   - Locate function at line
   - Locate method at line
   - Locate type at line
   - Locate var/const at line
   - Locate interface at line
   - Locate with column disambiguation
   - Error: position outside file bounds
   - Error: position has no symbol (e.g., comment line)

3. **collector_test.go**
   - Collect depth 0 (target only)
   - Collect depth 1 (direct dependencies)
   - Collect depth 2+ (transitive)
   - Handle external references (stdlib, third-party)
   - Handle unexported cross-package references
   - Handle circular dependencies
   - Handle generic type parameters
   - Handle interface method references

4. **analysis_test.go**
   - Find callers within same package
   - Find callers across packages
   - Find test callers
   - Compute cyclomatic complexity (simple, nested, multiple branches)
   - Count lines of code
   - Count dependencies

5. **git_test.go**
   - Parse git blame output
   - Parse git log output
   - Handle non-git repositories gracefully
   - Handle files not in git

6. **format/*_test.go**
   - Generate valid markdown
   - Generate valid HTML (parse with html parser)
   - Generate valid JSON (unmarshal)
   - Syntax highlighting works
   - File:line annotations correct
   - Hyperlinks in HTML mode

**Test Utilities**:
```go
// testdata/fixtures/
//   simple/      - single file, single function
//   nested/      - function calling function
//   cross_pkg/   - cross-package references
//   generics/    - generic code
//   interfaces/  - interface definitions

func loadTestPackage(t *testing.T, fixtureName string) *packages.Package
func assertSymbolFound(t *testing.T, result *Extract, symbolName string)
func assertSymbolNotFound(t *testing.T, result *Extract, symbolName string)
func assertDepth(t *testing.T, ref Reference, expectedDepth int)
```

### 11.2 Integration Tests

**Package**: `tests/integration`

**Test Scenarios**:

1. **Extract from real-world projects**
   - Clone small public Go projects
   - Extract known symbols
   - Verify output structure and content

2. **Format validation**
   - Markdown output parses correctly
   - HTML output is valid HTML5
   - JSON output matches schema

3. **End-to-end CLI**
   - Run CLI binary with various flag combinations
   - Parse output
   - Verify exit codes

4. **Performance**
   - Extract from large files (1000+ lines)
   - Extract with deep dependencies (depth 5+)
   - Measure time, ensure < 5 seconds for typical cases

**Test Helper**:
```go
func runCLI(t *testing.T, args ...string) (stdout, stderr string, exitCode int)
func parseMarkdownOutput(output string) (*ParsedExtract, error)
func parseHTMLOutput(output string) (*ParsedExtract, error)
func parseJSONOutput(output string) (*Result, error)
```

### 11.3 BDD Tests (Godog)

**Package**: `tests/features`

**Feature Files**:

**`extract_function.feature`**:
```gherkin
Feature: Extract function for review

  Scenario: Basic function extraction
    Given example module "ex1"
    When I extract "pkg/math/add.go" line 6 with depth 1
    Then the output includes function "Add"
    And the output includes function "validateInputs"
    And the output marks "fmt.Println" as external reference
    And the output does not include function "Sub"
    And the output is valid markdown

  Scenario: Target only (depth 0)
    Given example module "ex1"
    When I extract "pkg/math/add.go" line 6 with depth 0
    Then the output includes only function "Add"
    And the output shows "validateInputs" call as stub or external reference
```

**`depth_control.feature`**:
```gherkin
Feature: Control dependency depth

  Scenario: Shallow extraction
    Given example module "ex2"
    When I extract "cmd/api/main.go" line 30 with depth 0
    Then the output includes only "handleCreateUser"
    And the output shows "UserService.CreateUser" as stub

  Scenario: Medium depth
    Given example module "ex2"
    When I extract "cmd/api/main.go" line 30 with depth 1
    Then the output includes "handleCreateUser"
    And the output includes "UserService.CreateUser" signature

  Scenario: Deep extraction
    Given example module "ex2"
    When I extract "cmd/api/main.go" line 30 with depth 3
    Then the output includes "handleCreateUser"
    And the output includes "UserService.CreateUser" implementation
    And the output includes "validateEmail" helper
    And the output includes "User" type definition
```

**`callers.feature`**:
```gherkin
Feature: Show callers (reverse dependencies)

  Scenario: Find callers
    Given example module "ex2"
    When I extract "pkg/service/user.go" line 15 with --show-callers
    Then the output includes "CreateUser" method
    And the callers section includes "cmd/api/main.go:30"
    And the callers section includes "cmd/api/main.go:55"

  Scenario: Include test callers
    Given example module "ex2"
    When I extract "pkg/service/user.go" line 15 with --show-callers and --show-tests
    Then the callers section includes "pkg/service/user_test.go:20"
```

**`formats.feature`**:
```gherkin
Feature: Output formats

  Scenario: Markdown output
    Given example module "ex1"
    When I extract "pkg/math/add.go" line 6 with format "markdown"
    Then the output is valid markdown
    And the output has section "Target Symbol"
    And the output has section "Direct Dependencies"

  Scenario: HTML output
    Given example module "ex1"
    When I extract "pkg/math/add.go" line 6 with format "html"
    Then the output is valid HTML5
    And the output has syntax highlighting
    And the output has clickable file:line links

  Scenario: JSON output
    Given example module "ex1"
    When I extract "pkg/math/add.go" line 6 with format "json"
    Then the output is valid JSON
    And the JSON has field "target.symbol" = "Add"
    And the JSON has array "dependencies" with 1+ items
```

**`generics.feature`**:
```gherkin
Feature: Generic code extraction

  Scenario: Extract generic function
    Given example module "ex3"
    When I extract "pkg/iter/map.go" line 4
    Then the output includes generic function "Map[T any, U any]"
    And the output includes type parameter definitions
    And the output includes constraint "any"
```

**`edge_cases.feature`**:
```gherkin
Feature: Handle edge cases

  Scenario: Unexported cross-package symbol
    Given a module with unexported cross-package reference
    When I extract the referencing function
    Then the output notes "Reference to unexported symbol"
    And the output does not include the unexported symbol's code

  Scenario: Circular dependency
    Given a module with circular function calls
    When I extract one of the functions
    Then the output includes both functions
    And the dependency graph shows the cycle

  Scenario: Interface method
    Given example module "ex3"
    When I extract "pkg/io/writer.go" line 5 (interface method)
    Then the output includes the interface definition
    And the output does not include implementations (unless --show-callers)
```

**BDD Step Implementations**:
```go
// tests/features/steps.go

func (s *Suite) iExtractWithDepth(file string, line int, depth int) error {
    result, err := extract.ExtractSymbol(context.Background(),
        extract.Target{Root: s.moduleRoot, File: file, Line: line},
        extract.Options{Depth: depth})
    s.lastResult = result
    return err
}

func (s *Suite) theOutputIncludesFunction(name string) error {
    if !strings.Contains(s.lastResult.Rendered, "func "+name) {
        return fmt.Errorf("output does not include function %s", name)
    }
    return nil
}

func (s *Suite) theOutputIsValidMarkdown() error {
    // Parse markdown and check structure
    return validateMarkdown(s.lastResult.Rendered)
}

// etc.
```

### 11.4 Golden File Tests

**Package**: `tests/golden`

For each example, store expected output:

```
tests/golden/
├── ex1_add_depth0.md
├── ex1_add_depth1.md
├── ex1_add_depth1.html
├── ex1_add_depth1.json
├── ex2_handler_depth0.md
├── ex2_handler_depth2.md
└── ...
```

**Test**:
```go
func TestGoldenFiles(t *testing.T) {
    tests := []struct{
        name   string
        target extract.Target
        opts   extract.Options
        golden string
    }{
        {
            name: "ex1_add_depth0",
            target: extract.Target{Root: "../../examples/ex1", File: "pkg/math/add.go", Line: 6},
            opts: extract.Options{Depth: 0, Format: "markdown"},
            golden: "ex1_add_depth0.md",
        },
        // ...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := extract.ExtractSymbol(context.Background(), tt.target, tt.opts)
            require.NoError(t, err)

            goldenPath := filepath.Join("golden", tt.golden)
            if *update {
                os.WriteFile(goldenPath, []byte(result.Rendered), 0644)
            }

            expected, _ := os.ReadFile(goldenPath)
            assert.Equal(t, string(expected), result.Rendered)
        })
    }
}
```

Run with `-update` flag to regenerate golden files after intentional changes.

---

## 12. Performance

### 12.1 Targets

- **Typical extraction** (depth 1, < 10 symbols): < 1 second
- **Medium extraction** (depth 2, < 50 symbols): < 3 seconds
- **Large extraction** (depth 3+, < 200 symbols): < 10 seconds

### 12.2 Optimizations

1. **Lazy loading**: Don't load all packages upfront
2. **Caching**: Cache `packages.Load` results keyed by go.mod hash
3. **Parallel traversal**: Collect dependencies concurrently
4. **Early termination**: Stop at depth limit immediately
5. **Symbol deduplication**: Use map for visited symbols

### 12.3 Guardrails

- **Max symbols**: Warn if > 500 symbols included
- **Max output size**: Warn if > 5 MB
- **Timeout**: Context with 60s timeout for caller analysis

---

## 13. Error Handling

### 13.1 Error Categories

1. **Input Validation**
   - `ErrInvalidRoot`: Root path doesn't exist or isn't a Go module
   - `ErrInvalidFile`: File path invalid or not in module
   - `ErrInvalidPosition`: Line/column out of bounds

2. **Symbol Resolution**
   - `ErrSymbolNotFound`: No symbol at specified position
   - `ErrAmbiguousSymbol`: Multiple symbols at position (need column)

3. **Package Loading**
   - `ErrPackageLoadFailed`: Compilation errors in package
   - `ErrModuleNotFound`: go.mod missing

4. **External Tools**
   - `ErrGitNotAvailable`: Git commands failed (graceful degradation)
   - `ErrBuildFailed`: Package has build errors (load anyway, mark errors)

### 13.2 Error Messages

**Good Error Messages**:
```
Error: Symbol not found at position

Could not resolve a symbol at pkg/service/user.go:128:1

This line may be:
- A comment
- A blank line
- Inside a complex expression (try specifying --column)

Nearby symbols:
- Line 125: func (s *UserService) CreateUser
- Line 135: func (s *UserService) UpdateUser
```

```
Error: Unexported symbol required from other package

Found reference to unexported symbol:
  pkg/internal/helper.doStuff

This symbol cannot be accessed across package boundaries.

Suggestions:
- Review pkg/internal/helper separately
- Export the symbol if appropriate
- Refactor to avoid the cross-package unexported dependency
```

---

## 14. Limitations

### 14.1 Known Limitations

1. **No whole-program analysis**
   - Cannot find all implementations of an interface
   - Cannot track dynamic dispatch
   - Cannot follow reflection calls

2. **Build context**
   - Extracts for default GOOS/GOARCH
   - No build tag variations
   - No cgo support (graceful skip)

3. **Depth limits**
   - Deep transitive closures may miss context
   - User must choose appropriate depth

4. **External packages**
   - Only signatures for external deps
   - Cannot show implementation details
   - Stdlib usage just noted

5. **Dynamic behavior**
   - Cannot show runtime values
   - Cannot trace execution paths
   - No data flow analysis

### 14.2 Non-Goals

- ❌ Making extracted code compilable
- ❌ Full program slicing
- ❌ Whole-program analysis
- ❌ Build tag combinatorics
- ❌ Cross-module vendoring
- ❌ Code generation

---

## 15. Documentation Deliverables

### 15.1 README.md

```markdown
# go-scope: Extract Go Code for Review

Extract Go symbols with context for code review.

## Quick Start

```bash
# Install
go install github.com/you/go-scope/cmd/go-scope@latest

# Basic usage
go-scope --root . --file pkg/service/user.go --line 128

# With features
go-scope --root . --file pkg/service/user.go --line 128 \
  --depth 2 \
  --show-callers \
  --format html \
  --output review.html
```

## Features
- 🎯 Extract functions, methods, types, vars, consts
- 📊 Depth-controlled dependency traversal
- 🔍 Find callers (reverse dependencies)
- 📈 Complexity metrics
- 🎨 Syntax-highlighted HTML output
- 🌳 Dependency graph visualization
- 📝 Markdown, HTML, or JSON output
- 🔗 Git blame integration

## Use Cases
- Code review preparation
- Understanding unfamiliar code
- Documenting implementation details
- Sharing code snippets with context
- Analyzing dependencies

## Documentation
- [Usage Guide](docs/USAGE.md)
- [API Reference](docs/API.md)
- [Examples](examples/)

## License
MIT
```

### 15.2 USAGE.md

Comprehensive guide covering:
- Installation
- Basic usage patterns
- Flag reference
- Output format details
- Example workflows
- Tips and tricks
- Troubleshooting

### 15.3 API.md

Library API documentation:
- Go package docs
- Type reference
- Function signatures
- Usage examples
- Integration guide

### 15.4 DESIGN.md

Technical design doc:
- Architecture overview
- Algorithm details
- Design decisions
- Trade-offs
- Performance considerations
- Future enhancements

---

## 16. Milestones

### Phase 1: Core Extraction (Weeks 1-2)
- ✅ Project setup, go.mod, basic structure
- ✅ Implement `loader.go` (packages.Load wrapper)
- ✅ Implement `locator.go` (position → symbol)
- ✅ Implement `collector.go` (depth-limited BFS)
- ✅ Basic markdown output
- ✅ CLI with essential flags
- ✅ Unit tests for core logic
- ✅ Example ex1 working

**Deliverable**: MVP that can extract a function with depth 1

### Phase 2: Review Features (Week 3)
- ✅ Implement caller analysis (`analysis.go`)
- ✅ Implement external stub generation
- ✅ File:line annotations
- ✅ HTML output with syntax highlighting
- ✅ JSON output
- ✅ Examples ex2 and ex3
- ✅ Integration tests

**Deliverable**: Feature-complete extraction tool

### Phase 3: Advanced Features (Week 4)
- ✅ Metrics computation (cyclomatic complexity, LOC)
- ✅ Git integration (blame, log)
- ✅ Dependency graph generation
- ✅ Interactive HTML features (collapsible sections, search)
- ✅ Performance optimization
- ✅ BDD tests (godog)

**Deliverable**: Production-ready tool with all features

### Phase 4: Polish (Weeks 5-6)
- ✅ Comprehensive documentation
- ✅ Golden file tests
- ✅ Error message improvements
- ✅ CLI polish (help text, examples)
- ✅ Performance benchmarks
- ✅ Release preparation

**Deliverable**: v1.0.0 release

---

## 17. Acceptance Criteria

### Must Have (P0)

- ✅ Extract function/method at specified line
- ✅ Extract type definition at specified line
- ✅ Depth 0, 1, 2 traversal works correctly
- ✅ External references marked clearly
- ✅ Markdown output is readable and well-structured
- ✅ HTML output has syntax highlighting
- ✅ JSON output is valid and complete
- ✅ CLI accepts all documented flags
- ✅ Unit tests cover core logic (> 80% coverage)
- ✅ Integration tests validate end-to-end
- ✅ All three examples work as documented
- ✅ README with quick start exists

### Should Have (P1)

- ✅ Caller analysis finds all callers within module
- ✅ Complexity metrics are accurate
- ✅ Git blame shows recent changes
- ✅ Dependency graph is generated
- ✅ HTML output has interactive features
- ✅ Error messages are helpful
- ✅ Performance meets targets (< 3s typical)
- ✅ BDD tests cover main scenarios
- ✅ API documentation exists

### Nice to Have (P2)

- ⬜ Export to Gist/Pastebin
- ⬜ Terminal hyperlinks (OSC 8)
- ⬜ Watch mode (re-extract on file change)
- ⬜ VS Code extension
- ⬜ GitHub Action
- ⬜ Web service (HTTP API)

---

## 18. Future Enhancements (Post-v1.0)

### 18.1 Advanced Analysis

- **Data flow tracking**: Show how data flows through functions
- **Call graph visualization**: Full module call graph
- **Impact analysis**: "What breaks if I change this?"
- **Type hierarchy**: Show interface implementations

### 18.2 Integration

- **IDE plugins**: VS Code, GoLand extensions
- **CI/CD**: GitHub Action, GitLab CI integration
- **Code review tools**: GitHub PR comments, GitLab MR integration
- **Documentation generation**: Auto-generate docs from extracts

### 18.3 Collaboration

- **Shared extracts**: Upload and share via web service
- **Annotations**: Add comments/notes to extracted code
- **Discussions**: Inline discussion threads
- **Versions**: Compare extracts over time

---

## 19. Comparison to Original Spec

| Aspect | Original (Compilation) | Revised (Review) |
|--------|----------------------|------------------|
| **Primary Goal** | Compilable code | Readable code |
| **Output** | .go files or directory tree | Markdown/HTML/JSON |
| **Complexity** | High | Medium |
| **Core Features** | 15 | 8 |
| **Import Handling** | Rewrite, merge | Preserve, annotate |
| **Topology** | Sort, resolve cycles | Natural order |
| **Testing Focus** | "Compiles?" | "Correct symbols?" |
| **User Value** | Low (can't compile anyway) | High (reviews easier) |
| **Dev Time** | 10-12 weeks | 4-6 weeks |
| **LOC Estimate** | 5000-6000 | 2000-2500 |

**Key Removals**:
- ❌ Single-file vs multi-file modes
- ❌ Import rewriting
- ❌ Topological sorting
- ❌ go.mod generation
- ❌ Build tag handling
- ❌ Vendor mode
- ❌ Init() inclusion logic
- ❌ File count limits

**Key Additions**:
- ✅ Depth control
- ✅ Caller analysis
- ✅ Syntax highlighting
- ✅ External stubs
- ✅ HTML output
- ✅ Metrics
- ✅ Git integration
- ✅ Dependency graphs

---

## Appendix A: Example Outputs

### A.1 Markdown Output Example

```markdown
# Code Extract: CreateUser

**File**: pkg/service/user.go:128
**Package**: github.com/example/app/pkg/service
**Kind**: Method (UserService)
**Extracted**: 2025-10-23 14:32:00

## Metrics
- Lines of Code: 45
- Cyclomatic Complexity: 8
- Direct Dependencies: 4
- External Packages: 3

---

## Target Symbol

```go
// File: pkg/service/user.go:128-172
// CreateUser creates a new user account with email validation
func (s *UserService) CreateUser(ctx context.Context, email, name string) (*User, error) {
    // Validate email format
    if err := validateEmail(email); err != nil {  // → pkg/service/validation.go:45
        return nil, fmt.Errorf("invalid email: %w", err)
    }

    // Create user model
    user := &User{  // → pkg/model/user.go:12
        Email:     email,
        Name:      name,
        CreatedAt: time.Now(),  // → stdlib: time
    }

    // Persist to database
    if err := s.db.Create(user); err != nil {  // → external: database/sql.DB.Create
        return nil, fmt.Errorf("failed to create user: %w", err)
    }

    return user, nil
}
```

---

## Direct Dependencies (depth 1)

### validateEmail
**File**: pkg/service/validation.go:45-52
**Kind**: Function
**Package**: github.com/example/app/pkg/service

```go
// validateEmail checks if email format is valid
func validateEmail(email string) error {
    if !strings.Contains(email, "@") {  // → stdlib: strings
        return errors.New("invalid email format")  // → stdlib: errors
    }

    // Additional validation...
    if len(email) < 3 {
        return errors.New("email too short")
    }

    return nil
}
```

### User
**File**: pkg/model/user.go:12-18
**Kind**: Type (Struct)
**Package**: github.com/example/app/pkg/model

```go
// User represents a user account
type User struct {
    ID        int64
    Email     string
    Name      string
    CreatedAt time.Time  // → stdlib: time
}
```

---

## External References

- **database/sql**: `DB.Create(value interface{}) error`
- **fmt**: `Errorf(format string, a ...any) error`
- **time**: `Now() Time`, `Time`
- **errors**: `New(text string) error`
- **strings**: `Contains(s, substr string) bool`

---

## Called By

1. **cmd/api/handlers.go:67** - `handleCreateUser`
   ```go
   user, err := service.CreateUser(ctx, req.Email, req.Name)
   ```

2. **cmd/cli/user.go:123** - `createUserCommand`
   ```go
   user, err := svc.CreateUser(context.Background(), email, name)
   ```

3. **pkg/service/user_test.go:34** - `TestCreateUser`
   ```go
   user, err := service.CreateUser(ctx, "test@example.com", "Test User")
   ```

---

## Recent Changes (Git)

- **2025-10-20 10:30** by alice@example.com
  ```
  Add email validation to CreateUser

  - Call validateEmail helper
  - Return error if validation fails
  ```

- **2025-10-15 14:22** by bob@example.com
  ```
  Initial implementation of CreateUser
  ```

---

## Dependency Graph

```
CreateUser (target) [pkg/service]
├── validateEmail [pkg/service]
│   ├── strings.Contains [stdlib]
│   └── errors.New [stdlib]
├── User [pkg/model]
│   └── time.Time [stdlib]
├── DB.Create [external: database/sql]
├── fmt.Errorf [stdlib]
└── time.Now [stdlib]
```

---

*Generated by go-scope v1.0.0*
```

### A.2 JSON Output Example

```json
{
  "target": {
    "file": "pkg/service/user.go",
    "line": 128,
    "column": 1,
    "package": "github.com/example/app/pkg/service",
    "symbol": "CreateUser",
    "kind": "method",
    "receiver": "UserService",
    "exported": true,
    "code": "func (s *UserService) CreateUser(ctx context.Context, email, name string) (*User, error) {\n    if err := validateEmail(email); err != nil {\n        return nil, fmt.Errorf(\"invalid email: %w\", err)\n    }\n\n    user := &User{\n        Email:     email,\n        Name:      name,\n        CreatedAt: time.Now(),\n    }\n\n    if err := s.db.Create(user); err != nil {\n        return nil, fmt.Errorf(\"failed to create user: %w\", err)\n    }\n\n    return user, nil\n}",
    "doc": "CreateUser creates a new user account with email validation",
    "start_line": 128,
    "end_line": 145
  },
  "options": {
    "depth": 2,
    "format": "json",
    "stub_external": true,
    "show_callers": true,
    "show_tests": true,
    "annotate": true,
    "include_metrics": true,
    "git_blame": true
  },
  "dependencies": [
    {
      "symbol": {
        "package": "github.com/example/app/pkg/service",
        "name": "validateEmail",
        "kind": "function",
        "file": "pkg/service/validation.go",
        "line": 45,
        "end_line": 52,
        "code": "func validateEmail(email string) error {...}",
        "doc": "validateEmail checks if email format is valid",
        "exported": false
      },
      "reason": "direct-call",
      "depth": 1,
      "external": false,
      "stub": false,
      "referenced_by": "CreateUser"
    },
    {
      "symbol": {
        "package": "github.com/example/app/pkg/model",
        "name": "User",
        "kind": "type",
        "file": "pkg/model/user.go",
        "line": 12,
        "end_line": 18,
        "code": "type User struct {...}",
        "doc": "User represents a user account",
        "exported": true
      },
      "reason": "type-reference",
      "depth": 1,
      "external": false,
      "stub": false,
      "referenced_by": "CreateUser"
    },
    {
      "symbol": {
        "package": "database/sql",
        "name": "DB.Create",
        "kind": "method",
        "exported": true
      },
      "reason": "method-call",
      "depth": 1,
      "external": true,
      "stub": true,
      "signature": "Create(value interface{}) error",
      "referenced_by": "CreateUser"
    }
  ],
  "external_references": [
    {"package": "database/sql", "symbols": ["DB.Create"]},
    {"package": "fmt", "symbols": ["Errorf"]},
    {"package": "time", "symbols": ["Now", "Time"]},
    {"package": "errors", "symbols": ["New"]},
    {"package": "strings", "symbols": ["Contains"]}
  ],
  "callers": [
    {
      "file": "cmd/api/handlers.go",
      "line": 67,
      "function": "handleCreateUser",
      "context": "user, err := service.CreateUser(ctx, req.Email, req.Name)"
    },
    {
      "file": "cmd/cli/user.go",
      "line": 123,
      "function": "createUserCommand",
      "context": "user, err := svc.CreateUser(context.Background(), email, name)"
    },
    {
      "file": "pkg/service/user_test.go",
      "line": 34,
      "function": "TestCreateUser",
      "context": "user, err := service.CreateUser(ctx, \"test@example.com\", \"Test User\")"
    }
  ],
  "metrics": {
    "lines_of_code": 17,
    "logical_lines": 12,
    "cyclomatic_complexity": 3,
    "dependency_count": 4,
    "direct_deps": 3,
    "transitive_deps": 1,
    "external_packages": ["database/sql", "fmt", "time", "errors", "strings"]
  },
  "git_history": [
    {
      "commit": "abc123def456",
      "author": "alice@example.com",
      "date": "2025-10-20T10:30:00Z",
      "message": "Add email validation to CreateUser"
    },
    {
      "commit": "789ghi012jkl",
      "author": "bob@example.com",
      "date": "2025-10-15T14:22:00Z",
      "message": "Initial implementation of CreateUser"
    }
  ],
  "graph": {
    "format": "mermaid",
    "content": "graph TD\n  A[CreateUser] --> B[validateEmail]\n  B --> C[strings.Contains]\n  B --> D[errors.New]\n  A --> E[User]\n  E --> F[time.Time]\n  A --> G[DB.Create]\n  A --> H[fmt.Errorf]\n  A --> I[time.Now]"
  },
  "metadata": {
    "extracted_at": "2025-10-23T14:32:00Z",
    "go_version": "1.22.0",
    "module": "github.com/example/app",
    "module_version": "v0.1.0",
    "total_symbols": 8,
    "total_lines": 156,
    "tool_version": "1.0.0"
  }
}
```

---

## Appendix B: FAQ

**Q: Why doesn't the extracted code compile?**
A: This tool is designed for code review, not compilation. It focuses on readability and context over compilability.

**Q: How do I know what depth to use?**
A: Start with depth 1. If you need more context, increase to 2 or 3. Depth 0 shows just the target symbol.

**Q: Can I extract from third-party packages?**
A: Yes, as long as they're in your module's dependencies. External package code will be stubbed by default.

**Q: What if the symbol I want spans multiple lines?**
A: Just specify any line within the symbol's definition. The tool will extract the entire symbol.

**Q: How are callers found?**
A: The tool searches all packages in the module for references to the target symbol using Go's type information.

**Q: Can I use this in CI/CD?**
A: Yes! Use JSON output for machine-readable results, or generate HTML reports for code review.

**Q: What about code with build tags?**
A: The tool uses the default build context. Code with platform-specific build tags may not be fully visible.

**Q: Does this work with cgo?**
A: Partially. Go code is extracted, but C code and cgo interop details are not included.

---

## Appendix C: Glossary

- **Target Symbol**: The Go identifier (function, type, etc.) specified by file:line position
- **Depth**: How many levels of dependencies to include (0 = target only, 1 = direct deps, etc.)
- **External Reference**: A symbol from a different module or the standard library
- **Stub**: A signature-only representation of an external symbol
- **Caller**: A location where the target symbol is referenced/used
- **Annotation**: An inline comment showing where a reference is defined (file:line)
- **Cyclomatic Complexity**: A measure of code complexity based on the number of branching paths

---

## Appendix D: Related Tools

- **`go doc`**: Shows documentation for symbols (no dependencies)
- **`gopls definition`**: Jump to definition (IDE feature, single symbol)
- **`go-callvis`**: Visualizes call graphs (whole-program)
- **`gource`**: Visualizes git history (not code-focused)
- **`staticcheck`**: Lints code (different purpose)
- **`guru`**: Whole-program analysis (deprecated)

**go-scope differentiator**: Focused extraction for review with controlled depth and rich formatting.

---

**End of Specification v2.0**
