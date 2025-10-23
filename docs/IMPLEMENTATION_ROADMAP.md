# Implementation Roadmap - Go Scope Extractor

**Goal**: Extract functional context around Go symbols for comprehension, not compilation.

**Timeline**: 4-6 weeks with comprehensive TDD approach

---

## Phase 1: Core Extraction (Weeks 1-2)

### Week 1: Foundation

#### Day 1-2: Project Setup
- [x] Spec approved ✅
- [ ] Create project structure
  ```
  go-scope/
  ├── cmd/go-scope/main.go
  ├── internal/extract/
  ├── pkg/cli/
  ├── examples/
  ├── tests/
  └── docs/
  ```
- [ ] Initialize go.mod
  ```bash
  go mod init github.com/you/go-scope
  go get golang.org/x/tools/go/packages
  go get golang.org/x/tools/go/ast/astutil
  ```
- [ ] Set up testing framework
  ```bash
  go get github.com/stretchr/testify
  go get github.com/cucumber/godog/cmd/godog
  ```
- [ ] Create basic CLI structure with cobra/flag
- [ ] Write first test (RED)

**Deliverable**: Project builds, tests run (even if failing)

#### Day 3-5: Symbol Location (TDD)

**Test First**:
```go
// internal/extract/locator_test.go
func TestLocateFunctionAtLine(t *testing.T) {
    // Given: A simple Go file with a function
    // When: I locate symbol at the function's line
    // Then: Returns the function's types.Object
}

func TestLocateMethodAtLine(t *testing.T) { ... }
func TestLocateTypeAtLine(t *testing.T) { ... }
func TestLocateVarAtLine(t *testing.T) { ... }
func TestLocateConstAtLine(t *testing.T) { ... }
func TestLocateInterfaceAtLine(t *testing.T) { ... }

func TestLocateSymbolNotFound(t *testing.T) {
    // Given: Line with no symbol (comment, blank)
    // When: I try to locate symbol
    // Then: Returns error with nearby suggestions
}
```

**Implementation**:
```go
// internal/extract/locator.go
package extract

type Locator struct {
    pkg *packages.Package
    fset *token.FileSet
}

func (l *Locator) LocateSymbol(file string, line, col int) (*Symbol, error)
```

**Tests Green**: All locator tests pass

**Deliverable**: Can find any symbol type at a line/column

#### Day 6-7: Package Loading (TDD)

**Test First**:
```go
// internal/extract/loader_test.go
func TestLoadSinglePackage(t *testing.T) {
    // Given: Valid Go module
    // When: Load package containing file
    // Then: Returns package with types info
}

func TestLoadMultiPackage(t *testing.T) { ... }
func TestLoadInvalidRoot(t *testing.T) { ... }
func TestLoadNotGoModule(t *testing.T) { ... }
```

**Implementation**:
```go
// internal/extract/loader.go
package extract

func LoadPackage(root, file string) (*packages.Package, error)
func LoadModule(root string) ([]*packages.Package, error)
```

**Tests Green**: All loader tests pass

**Deliverable**: Robust package loading with error handling

### Week 2: Dependency Collection

#### Day 8-10: Depth-Limited Traversal (TDD)

**Test First**:
```go
// internal/extract/collector_test.go
func TestCollectDepth0(t *testing.T) {
    // Given: Function that calls helpers
    // When: Collect with depth 0
    // Then: Returns only target function
}

func TestCollectDepth1(t *testing.T) {
    // Given: Function that calls helpers
    // When: Collect with depth 1
    // Then: Returns target + helpers
}

func TestCollectDepth2(t *testing.T) {
    // Given: Function -> Helper -> SubHelper
    // When: Collect with depth 2
    // Then: Returns all three
}

func TestCollectExternalReference(t *testing.T) {
    // Given: Function using stdlib
    // When: Collect with depth 1
    // Then: Returns target, marks stdlib as external
}

func TestCollectCircularDependency(t *testing.T) {
    // Given: A <-> B circular reference
    // When: Collect with depth 2
    // Then: Includes both, no infinite loop
}

func TestCollectGenericFunction(t *testing.T) {
    // Given: Generic function with constraints
    // When: Collect with depth 1
    // Then: Includes type parameters and constraints
}
```

**Implementation**:
```go
// internal/extract/collector.go
package extract

type Collector struct {
    pkg     *packages.Package
    visited map[string]bool
    depth   int
}

func (c *Collector) Collect(target *Symbol, maxDepth int) ([]Reference, error)
```

**Algorithm**:
```
BFS with depth tracking:
1. Queue: [(target, 0)]
2. While queue not empty:
   - Pop (symbol, depth)
   - If visited, skip
   - Mark visited
   - Add to result
   - If depth >= maxDepth, skip children
   - Find references in symbol's code
   - Enqueue references with depth+1
3. Return result
```

**Tests Green**: All collector tests pass

**Deliverable**: Depth-limited dependency collection works perfectly

#### Day 11-12: Basic Markdown Output (TDD)

**Test First**:
```go
// internal/extract/format/markdown_test.go
func TestFormatSimpleFunction(t *testing.T) {
    // Given: Extract with one function
    // When: Format as markdown
    // Then: Valid markdown with sections
}

func TestFormatWithDependencies(t *testing.T) {
    // Given: Extract with dependencies
    // When: Format as markdown
    // Then: Target section + Dependencies section
}

func TestFormatAnnotations(t *testing.T) {
    // Given: Code with references
    // When: Format with annotations
    // Then: Has file:line comments
}
```

**Implementation**:
```go
// internal/extract/format/markdown.go
package format

func ToMarkdown(extract *Extract, opts Options) (string, error)
```

**Tests Green**: Markdown output is valid and readable

**Deliverable**: Can extract and output readable markdown

#### Day 13-14: CLI Integration & Example 1

**CLI**:
```go
// cmd/go-scope/main.go
package main

func main() {
    // Parse flags
    // Load package
    // Locate symbol
    // Collect dependencies
    // Format output
    // Print or write to file
}
```

**Example 1**: Create `examples/ex1`
```
examples/ex1/
├── go.mod
├── pkg/math/
│   ├── add.go      // Add function calls helper
│   └── util.go     // helper function
└── README.md
```

**Manual Test**:
```bash
go run ./cmd/go-scope \
  --root examples/ex1 \
  --file pkg/math/add.go \
  --line 6 \
  --depth 1

# Should show:
# - Add function
# - helper function (depth 1)
# - External refs (fmt)
```

**Deliverable**: Working MVP! Can extract from Example 1.

---

## Phase 2: Review Features (Week 3)

### Day 15-16: Caller Analysis (TDD)

**Test First**:
```go
// internal/extract/analysis_test.go
func TestFindCallers_SamePackage(t *testing.T) {
    // Given: Function called by another in same package
    // When: Find callers
    // Then: Returns caller location
}

func TestFindCallers_CrossPackage(t *testing.T) { ... }
func TestFindCallers_None(t *testing.T) { ... }
func TestFindCallers_Multiple(t *testing.T) { ... }
func TestFindCallers_Tests(t *testing.T) {
    // Given: Function with test
    // When: Find callers with show_tests=true
    // Then: Includes test function
}
```

**Implementation**:
```go
// internal/extract/analysis.go
package extract

type Analyzer struct {
    pkgs []*packages.Package
}

func (a *Analyzer) FindCallers(target *Symbol, includeTests bool) ([]Caller, error)
```

**Algorithm**:
```
1. Build index: symbol -> locations where used
2. For each package in module:
   - Walk all call expressions
   - Resolve callee to types.Object
   - If matches target, record caller
3. Return list of caller locations
```

**Tests Green**: Caller discovery works

**Deliverable**: `--show-callers` flag works

### Day 17-18: External Stubs & HTML Output

**External Stubs Test**:
```go
func TestStubExternalFunction(t *testing.T) {
    // Given: Call to external package function
    // When: Collect with stub_external=true
    // Then: Shows signature, not implementation
}
```

**HTML Test**:
```go
// internal/extract/format/html_test.go
func TestHTMLOutput(t *testing.T) {
    // Given: Extract
    // When: Format as HTML
    // Then: Valid HTML5 with syntax highlighting
}

func TestHTMLHyperlinks(t *testing.T) {
    // Given: Extract with annotations
    // When: Format as HTML
    // Then: file:line become clickable links
}
```

**Implementation**:
```go
// internal/extract/format/html.go
package format

func ToHTML(extract *Extract, opts Options) (string, error)
```

Use: `github.com/alecthomas/chroma` for syntax highlighting

**Tests Green**: HTML output works

**Deliverable**: Beautiful HTML output with highlighting

### Day 19-20: JSON Output & Example 2

**JSON Test**:
```go
// internal/extract/format/json_test.go
func TestJSONOutput(t *testing.T) {
    // Given: Extract
    // When: Format as JSON
    // Then: Valid JSON matching schema
}

func TestJSONRoundTrip(t *testing.T) {
    // Given: Extract
    // When: Format to JSON then parse
    // Then: Data preserved
}
```

**Implementation**:
```go
// internal/extract/format/json.go
package format

func ToJSON(extract *Extract, opts Options) (string, error)
```

**Example 2**: Create `examples/ex2`
```
examples/ex2/
├── go.mod
├── cmd/api/
│   └── main.go         // HTTP handler
├── pkg/
│   ├── service/
│   │   ├── user.go     // UserService
│   │   └── validation.go
│   └── model/
│       └── user.go     // User type
└── README.md
```

**Manual Tests**:
```bash
# Extract with callers
go run ./cmd/go-scope \
  --root examples/ex2 \
  --file pkg/service/user.go \
  --line 15 \
  --show-callers

# HTML output
go run ./cmd/go-scope \
  --root examples/ex2 \
  --file pkg/service/user.go \
  --line 15 \
  --depth 2 \
  --format html \
  --output review.html
```

**Deliverable**: All output formats work. Example 2 complete.

### Day 21: Integration Tests

**Integration Tests**:
```go
// tests/integration/extract_test.go
func TestExtractEx1_AddFunction(t *testing.T) {
    // Real extraction from ex1
    result := extractSymbol(t, "examples/ex1", "pkg/math/add.go", 6, 1)

    assert.Contains(result.Target.Name, "Add")
    assert.Len(result.References, 1) // helper
    assert.Contains(result.External, "fmt")
}

func TestExtractEx2_UserService(t *testing.T) {
    // Real extraction from ex2
    result := extractSymbol(t, "examples/ex2", "pkg/service/user.go", 15, 1)

    assert.Contains(result.Target.Name, "CreateUser")
    assert.Len(result.Callers, 2) // handler + test
}
```

**CLI Integration**:
```go
// tests/integration/cli_test.go
func TestCLI_BasicExtract(t *testing.T) {
    stdout, stderr, code := runCLI(
        "--root", "examples/ex1",
        "--file", "pkg/math/add.go",
        "--line", "6",
    )

    assert.Equal(0, code)
    assert.Contains(stdout, "func Add")
}
```

**Deliverable**: Integration tests ensure end-to-end works

---

## Phase 3: Advanced Features (Week 4)

### Day 22-23: Metrics Computation (TDD)

**Test First**:
```go
// internal/extract/analysis_test.go (continued)
func TestCyclomaticComplexity_Simple(t *testing.T) {
    // Given: Function with no branches
    // When: Compute complexity
    // Then: Returns 1
}

func TestCyclomaticComplexity_OneBranch(t *testing.T) {
    // Given: Function with one if
    // When: Compute complexity
    // Then: Returns 2
}

func TestCyclomaticComplexity_Nested(t *testing.T) {
    // Given: Nested if/for/switch
    // When: Compute complexity
    // Then: Returns sum of all branches + 1
}

func TestLinesOfCode(t *testing.T) {
    // Given: Function
    // When: Count lines
    // Then: Returns physical and logical lines
}

func TestDependencyCount(t *testing.T) {
    // Given: Extract with dependencies
    // When: Count deps
    // Then: Returns direct + transitive counts
}
```

**Implementation**:
```go
// internal/extract/analysis.go (extended)
func (a *Analyzer) ComputeMetrics(target *Symbol, refs []Reference) (*Metrics, error)
func (a *Analyzer) cyclomaticComplexity(node ast.Node) int
func (a *Analyzer) linesOfCode(node ast.Node) (physical, logical int)
```

**Algorithm (Cyclomatic Complexity)**:
```
complexity = 1
Walk AST:
  If node is:
    - IfStmt: +1
    - ForStmt: +1
    - RangeStmt: +1
    - SwitchStmt: +1
    - CaseClause: +1 (per case)
    - BinaryExpr with &&, ||: +1
Return complexity
```

**Tests Green**: Metrics accurate

**Deliverable**: `--metrics` flag works

### Day 24-25: Git Integration (TDD)

**Test First**:
```go
// internal/extract/git_test.go
func TestGitBlame(t *testing.T) {
    // Given: Git repo with commits
    // When: Blame target symbol
    // Then: Returns author/date per line
}

func TestGitLog(t *testing.T) {
    // Given: Git repo
    // When: Get log for file
    // Then: Returns recent commits
}

func TestGitNotAvailable(t *testing.T) {
    // Given: Not a git repo
    // When: Try git operations
    // Then: Gracefully skips, no error
}
```

**Implementation**:
```go
// internal/extract/git.go
package extract

func GitBlame(file string, startLine, endLine int) ([]GitBlame, error)
func GitLog(file string, limit int) ([]GitBlame, error)
```

Use: `exec.Command("git", "blame", ...)` or `github.com/go-git/go-git`

**Tests Green**: Git integration works

**Deliverable**: `--git-blame` flag works

### Day 26-27: Interactive HTML & Example 3

**Interactive HTML Features**:
- Collapsible sections
- Search box
- Copy buttons
- Dark/light theme toggle

**Example 3**: Create `examples/ex3`
```
examples/ex3/
├── go.mod
├── pkg/
│   ├── iter/
│   │   ├── map.go      // Generic Map[T, U]
│   │   └── filter.go   // Generic Filter[T]
│   └── io/
│       └── writer.go   // Interface + implementation
└── README.md
```

**Manual Test**:
```bash
go run ./cmd/go-scope \
  --root examples/ex3 \
  --file pkg/iter/map.go \
  --line 4 \
  --depth 1 \
  --format html \
  --output generics.html
```

**Deliverable**: All features work. All examples complete.

### Day 28: BDD Tests (Godog)

**Setup Godog**:
```bash
go install github.com/cucumber/godog/cmd/godog@latest
```

**Features** (`tests/features/`):

1. `extract_function.feature`
```gherkin
Feature: Extract function for review

  Scenario: Basic function extraction
    Given example module "ex1"
    When I extract "pkg/math/add.go" line 6 with depth 1
    Then output includes function "Add"
    And output includes function "validateInputs"
    And output marks "fmt.Println" as external
    And output does not include function "Sub"
```

2. `depth_control.feature`
```gherkin
Feature: Control dependency depth

  Scenario: Depth 0
    Given example module "ex1"
    When I extract "pkg/math/add.go" line 6 with depth 0
    Then output includes only function "Add"
```

3. `callers.feature`
```gherkin
Feature: Find callers

  Scenario: Show callers
    Given example module "ex2"
    When I extract "pkg/service/user.go" line 15 with --show-callers
    Then callers section includes "cmd/api/main.go"
```

**Step Definitions**:
```go
// tests/features/steps.go
func (s *Suite) iExtract(file string, line int, depth int) error {
    // Run extraction
}

func (s *Suite) outputIncludesFunction(name string) error {
    // Assert function in output
}

// etc.
```

**Run BDD**:
```bash
cd tests/features
godog
```

**Deliverable**: Comprehensive BDD coverage

---

## Phase 4: Polish (Weeks 5-6)

### Week 5: Testing & Documentation

#### Day 29-30: Unit Test Coverage

**Goal**: >80% coverage on core packages

```bash
go test -cover ./internal/extract/...
go test -cover ./internal/extract/format/...
go test -cover ./pkg/cli/...
```

**Focus Areas**:
- Edge cases in locator
- Error paths in loader
- Depth limits in collector
- Format edge cases

**Use**:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Deliverable**: High test coverage

#### Day 31-32: Golden File Tests

**Setup**:
```
tests/golden/
├── ex1_add_depth0.md
├── ex1_add_depth1.md
├── ex1_add_depth1.html
├── ex1_add_depth1.json
├── ex2_user_depth1.md
├── ex2_user_depth2.md
└── ...
```

**Test**:
```go
// tests/golden/golden_test.go
func TestGoldenFiles(t *testing.T) {
    tests := []struct{
        name   string
        target Target
        opts   Options
        golden string
    }{
        {
            name: "ex1_add_depth0",
            target: Target{Root: "../../examples/ex1", File: "pkg/math/add.go", Line: 6},
            opts: Options{Depth: 0, Format: "markdown"},
            golden: "ex1_add_depth0.md",
        },
        // ...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := extractSymbol(t, tt.target, tt.opts)

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

**Run**:
```bash
# Generate golden files
go test ./tests/golden/... -update

# Verify against golden files
go test ./tests/golden/...
```

**Deliverable**: Regression tests via golden files

#### Day 33-35: Documentation

**README.md**:
- Project description
- Quick start
- Installation
- Examples
- Features
- Links to detailed docs

**USAGE.md**:
- Flag reference
- Use cases
- Examples
- Tips & tricks

**API.md**:
- Go package docs
- Type reference
- Function signatures
- Library usage examples

**DESIGN.md**:
- Architecture overview
- Algorithm details
- Design decisions
- Trade-offs

**FAQ.md**:
- Common questions
- Troubleshooting
- Limitations

**Deliverable**: Comprehensive documentation

### Week 6: Optimization & Release

#### Day 36-37: Performance Optimization

**Benchmark**:
```go
// internal/extract/benchmark_test.go
func BenchmarkExtractDepth1(b *testing.B) {
    for i := 0; i < b.N; i++ {
        extractSymbol(...)
    }
}

func BenchmarkExtractDepth3(b *testing.B) { ... }
func BenchmarkCallerAnalysis(b *testing.B) { ... }
```

**Profile**:
```bash
go test -bench=. -cpuprofile=cpu.prof ./internal/extract/
go tool pprof cpu.prof
```

**Optimize**:
- Cache package loads
- Parallel reference collection
- Optimize AST walks
- Reduce allocations

**Target**:
- Depth 1: < 1 second
- Depth 2: < 3 seconds
- Caller analysis: < 2 seconds

**Deliverable**: Fast extraction

#### Day 38-39: Error Messages & UX

**Improve Errors**:
```go
// Before:
return fmt.Errorf("symbol not found")

// After:
return fmt.Errorf("symbol not found at %s:%d\n\nThis line may be a comment or blank line.\n\nNearby symbols:\n  Line %d: %s\n  Line %d: %s\n\nTry one of these lines instead.",
    file, line, nearbySymbols...)
```

**Add Suggestions**:
- No symbol found → suggest nearby lines
- Depth too high → suggest reducing depth
- Large output → suggest HTML format
- Missing callers → remind about --show-callers

**Polish CLI**:
- Colorize output (in terminal mode)
- Progress indicators for slow operations
- Better help text
- Examples in --help

**Deliverable**: Great UX

#### Day 40-42: Final Testing & Bug Fixes

**Test Matrix**:
- [ ] All examples work
- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] All BDD scenarios pass
- [ ] All golden files match
- [ ] Benchmarks meet targets
- [ ] No known bugs

**Test on Real Projects**:
```bash
# Test on popular Go projects
git clone https://github.com/gin-gonic/gin
go-scope --root gin --file gin.go --line 50

git clone https://github.com/gorilla/mux
go-scope --root mux --file mux.go --line 100
```

**Fix Issues**:
- Edge cases found in real projects
- Performance bottlenecks
- Output formatting issues

**Deliverable**: Production-ready tool

---

## Release Checklist

### Pre-Release

- [ ] All tests pass (unit, integration, BDD)
- [ ] Coverage >80%
- [ ] Documentation complete
- [ ] Examples work
- [ ] Performance targets met
- [ ] No known bugs
- [ ] README polished
- [ ] CHANGELOG written
- [ ] License file added (MIT)

### Release v1.0.0

- [ ] Tag release: `git tag v1.0.0`
- [ ] Push tag: `git push origin v1.0.0`
- [ ] Create GitHub release with notes
- [ ] Publish to GitHub
- [ ] Announce on relevant channels

### Post-Release

- [ ] Monitor for issues
- [ ] Respond to feedback
- [ ] Plan v1.1 features based on usage

---

## Success Metrics

### Functionality
- ✅ Can extract any Go symbol type
- ✅ Depth control works (0, 1, 2, 3+)
- ✅ Caller analysis finds all callers
- ✅ All output formats work (markdown, HTML, JSON)
- ✅ Metrics are accurate
- ✅ Git integration works

### Quality
- ✅ >80% test coverage
- ✅ All BDD scenarios pass
- ✅ Golden files validate output
- ✅ No regressions

### Performance
- ✅ Depth 1: <1s
- ✅ Depth 2: <3s
- ✅ Depth 3: <10s
- ✅ No memory leaks

### Documentation
- ✅ README is clear
- ✅ Examples work
- ✅ API docs complete
- ✅ Usage guide helpful

### UX
- ✅ CLI is intuitive
- ✅ Error messages helpful
- ✅ Output is readable
- ✅ HTML is beautiful

---

## Daily Workflow

### Morning (TDD Cycle)
1. Write failing test (RED)
2. Write minimal code to pass (GREEN)
3. Refactor for clarity (REFACTOR)
4. Commit: `git commit -m "feat: implement X with tests"`

### Afternoon
5. Write more tests for edge cases
6. Ensure all tests still pass
7. Update documentation if needed
8. Commit: `git commit -m "test: add edge case tests for X"`

### Before End of Day
9. Run full test suite
10. Check coverage
11. Update roadmap status
12. Plan next day

---

## Risk Mitigation

### Risk: Package Loading Fails
**Mitigation**: Comprehensive error handling, fallback to best-effort

### Risk: Complex Code Slows Extraction
**Mitigation**: Depth limits, timeout guards, caching

### Risk: Output Too Large
**Mitigation**: Warnings, suggest reducing depth or HTML format

### Risk: Scope Creep
**Mitigation**: Stick to roadmap, defer non-essential features to v1.1

---

## Post-v1.0 Ideas (Not in Scope)

- VS Code extension
- Web service / HTTP API
- Diff mode (compare versions)
- Data flow analysis
- Impact analysis ("what breaks if I change this?")
- GitHub Action for PR comments
- Shared extracts (upload & share)
- Interactive web UI

These can be v1.1, v1.2, etc. after v1.0 proves the core concept.

---

## Timeline Summary

| Week | Phase | Focus | Deliverable |
|------|-------|-------|-------------|
| 1 | Core (Part 1) | Locator, Loader | Can find symbols |
| 2 | Core (Part 2) | Collector, Basic Output | Working MVP |
| 3 | Review Features | Callers, HTML, JSON | Feature complete |
| 4 | Advanced | Metrics, Git, BDD | All features |
| 5 | Polish (Part 1) | Tests, Docs | High quality |
| 6 | Polish (Part 2) | Perf, UX, Release | v1.0.0 |

**Total**: 6 weeks, production-ready tool with comprehensive TDD

---

## Getting Started (Implementation Begins)

### Step 1: Create Project Structure
```bash
mkdir -p cmd/go-scope
mkdir -p internal/extract/format
mkdir -p pkg/cli
mkdir -p examples/{ex1,ex2,ex3}
mkdir -p tests/{unit,integration,features,golden}
mkdir -p docs
```

### Step 2: Initialize Module
```bash
go mod init github.com/you/go-scope
go get golang.org/x/tools/go/packages
go get golang.org/x/tools/go/ast/astutil
go get github.com/stretchr/testify
```

### Step 3: Write First Test
```go
// internal/extract/locator_test.go
package extract

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestLocateFunctionAtLine(t *testing.T) {
    t.Skip("TODO: implement")
}
```

### Step 4: Make Test Fail (RED)
```bash
go test ./internal/extract/
# Test skipped - good, we're ready to implement
```

### Step 5: Begin TDD Cycle
Remove `t.Skip()`, watch test fail, implement until green.

---

**Ready to begin! Start with Phase 1, Day 1. Let's build this! 🚀**
