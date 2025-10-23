# Go Scoped Extractor - Specification Analysis

**Date**: 2025-10-23
**Status**: Specification Review - Critical Issues Identified
**Priority**: High - Foundational design mismatch

## Executive Summary

The original specification focuses on **extracting compilable code**, but the actual requirement is for **code review/inspection**. This fundamental mismatch leads to unnecessary complexity and missing critical features. This document analyzes the issues and provides recommendations.

---

## Critical Finding

> **"The extracted code doesn't have to work, it's for reviewing"** - User Clarification

This single statement invalidates approximately 40% of the original specification's requirements and dramatically simplifies the tool's scope.

---

## Issues by Category

### 🔴 Category 1: Unnecessary Compilation Requirements

**Problem**: The spec is heavily focused on making extracted code compilable.

**Specific Issues**:

1. **Goal Statement (Section 1)**
   - ❌ "smallest compilable code closure"
   - ❌ "buildable with go build"
   - ✅ Should be: "extract symbol with context for human review"

2. **Output Requirements (Section 3)**
   - ❌ "buildable with go build when combined with go.mod"
   - ❌ "Generate minimal go.mod for workspace extract"
   - ❌ "Preserve package names for multi-file compilation"
   - ❌ Format validation with go/format
   - ✅ Should be: Readable output with clear structure

3. **Dependency Traversal (Section 5)**
   - ❌ "Include init() if var inits require it"
   - ❌ "Topological sort declarations by dependency"
   - ❌ "Kahn's algorithm with SCC collapse"
   - ❌ Import rewriting and synthetic package creation
   - ✅ Should be: Simple depth-limited traversal with annotations

4. **Edge Cases (Section 6)**
   - ❌ "Ensure topological sort respects cycles"
   - ❌ "init() in multiple files: include only those required"
   - ❌ "cgo unless CGO_ENABLED=1"
   - ✅ Should be: Mark but don't resolve edge cases

5. **Testing (Sections 11-12)**
   - ❌ "output file compiles"
   - ❌ "single-file compile with go tool compile"
   - ❌ "multi-file compile as module directory"
   - ✅ Should be: Verify correct symbols extracted

**Impact**:
- Adds ~2000+ lines of unnecessary code
- Increases complexity by ~300%
- Delays delivery by ~4-6 weeks
- Creates maintenance burden

**Recommendation**: Remove all compilation-related logic

---

### 🟡 Category 2: Missing Review-Critical Features

**Problem**: The spec lacks features essential for code review.

**Missing Features**:

1. **Caller Analysis**
   - Not specified: "What calls this function?"
   - Essential for: Understanding usage patterns
   - Implementation: Reverse dependency lookup

2. **Syntax Highlighting**
   - Not specified: Color-coded output
   - Essential for: Readability
   - Implementation: HTML/Markdown with syntax highlighting

3. **Depth Control**
   - Mentioned but not detailed: How deep to traverse dependencies
   - Essential for: Balancing context vs noise
   - Implementation: `--depth` flag with sensible defaults

4. **External Reference Stubs**
   - Not specified: Show signatures of external dependencies
   - Essential for: Understanding interfaces without pulling full implementations
   - Implementation: Type signature extraction

5. **Hyperlinked Output**
   - Not specified: Jump to definition links
   - Essential for: Interactive review
   - Implementation: HTML mode with file:line links

6. **Git Integration**
   - Not specified: Show blame, recent changes
   - Essential for: Historical context
   - Implementation: libgit2 bindings

7. **Metrics**
   - Not specified: Complexity, size, dependency count
   - Essential for: Quick assessment
   - Implementation: Cyclomatic complexity, LOC counter

**Impact**:
- Tool less useful for actual review tasks
- Users must augment with other tools
- Reduced competitive advantage

**Recommendation**: Add these as primary features

---

### 🟠 Category 3: Over-Engineering

**Problem**: Solving problems that don't exist for review use case.

**Over-Engineered Areas**:

1. **Two Output Modes**
   - `single-file`: Merge everything into one compilable file
   - `multi-file`: Create minimal directory tree
   - Reality: Just need readable formatted output
   - Simplification: One mode with clear section separators

2. **Import Rewriting**
   - Complex alias preservation
   - Import path manipulation
   - Package name conflicts
   - Reality: Just show import paths as-is
   - Simplification: Include import statements verbatim

3. **Build Tag Handling**
   - Complex combinatorics
   - Platform-specific includes
   - Vendor mode
   - Reality: Extract what's visible in default build
   - Simplification: Document build context, don't try to resolve

4. **Go.mod Generation**
   - Minimal module synthesis
   - Require/replace pruning
   - Workspace configuration
   - Reality: Not needed for review
   - Simplification: Remove entirely

5. **Unexported Symbol Resolution**
   - Complex cross-package unexported handling
   - Forcing multi-file mode
   - Reality: Just note "unexported from pkg X" and move on
   - Simplification: Stub or skip with annotation

6. **File Count Limits**
   - `--max-files` guardrail
   - Reality: Not a concern without compilation
   - Simplification: Remove

**Impact**:
- Architecture complexity: 7 internal packages → 4 packages
- Code volume: ~5000 lines → ~2000 lines
- Cognitive load: High → Medium
- Test complexity: 150+ test cases → 60+ test cases

**Recommendation**: Radically simplify

---

### 🟢 Category 4: What Works Well

**Keep These Aspects**:

1. ✅ **Use of `golang.org/x/tools/go/packages`**
   - Correct choice for Go code analysis
   - Handles module resolution well

2. ✅ **Symbol Location Logic (Section 5.2)**
   - PathEnclosingInterval approach
   - Handling func/method/type/var/const/interface
   - Sound design

3. ✅ **Testing Strategy Foundation**
   - TDD/BDD approach
   - Unit + Integration + BDD mix
   - Golden file testing
   - Just needs refocused test cases

4. ✅ **Architecture Separation**
   - loader, locator, walker separate concerns
   - Good separation of duties
   - Just remove unnecessary assemblers

5. ✅ **Example-Driven Development**
   - Multiple example projects
   - Feature files referencing specific lines
   - Clear acceptance criteria

6. ✅ **Manifest/Metadata**
   - JSON output option
   - Structured information
   - Just needs schema update for review focus

7. ✅ **Documentation Structure**
   - Clear sections
   - Detailed specifications
   - Just needs content update

---

## Quantitative Analysis

### Complexity Metrics

| Aspect | Original Spec | Review-Focused | Reduction |
|--------|--------------|----------------|-----------|
| Core features | 15 | 8 | 47% |
| CLI flags | 12 | 9 | 25% |
| Output modes | 2 complex | 1 simple | 50% |
| Internal packages | 7 | 4 | 43% |
| Edge cases | 25+ | 12 | 52% |
| Test scenarios | 40+ | 20 | 50% |
| Estimated LOC | 5000-6000 | 2000-2500 | 58% |
| Dev time (weeks) | 10-12 | 4-6 | 50% |

### Feature Priority (Review Use Case)

| Priority | Feature | Original Spec | Recommendation |
|----------|---------|---------------|----------------|
| P0 | Symbol extraction | ✅ Covered | Keep |
| P0 | Reference traversal | ✅ Covered | Simplify |
| P0 | Readable output | ⚠️ Partial | Enhance |
| P0 | Syntax highlighting | ❌ Missing | Add |
| P1 | Depth control | ⚠️ Mentioned | Specify |
| P1 | Caller analysis | ❌ Missing | Add |
| P1 | External stubs | ❌ Missing | Add |
| P1 | HTML output | ❌ Missing | Add |
| P2 | Git integration | ❌ Missing | Add |
| P2 | Metrics | ❌ Missing | Add |
| P3 | Compilation | ✅ Over-specified | Remove |
| P3 | Import rewriting | ✅ Over-specified | Remove |
| P3 | Multi-file mode | ✅ Over-specified | Remove |

---

## Detailed Recommendations

### 1. Rewrite Goal Statement

**Current** (Section 1):
> "Given (repoRoot, filePath, lineNumber) in a Go project, extract the smallest compilable code closure that implements the symbol at that source location, plus all directly required declarations across files and packages."

**Revised**:
> "Given (repoRoot, filePath, lineNumber) in a Go project, extract the symbol at that location along with configurable context depth for code review. Output includes the target code, referenced symbols with annotations, caller locations, and formatted presentation optimized for human readability and understanding."

### 2. Simplify CLI Interface

**Remove**:
```bash
--mode single-file|multi-file
--build-tags
--vendor
--max-files
--skip-generated  # Less relevant for review
```

**Add**:
```bash
--depth N              # Dependency traversal depth (default: 1)
--format markdown|html|json  # Output format (default: markdown)
--stub-external        # Show signatures for external deps
--show-callers         # Include reverse dependencies
--show-tests           # Include tests for this symbol
--context-lines N      # Extra lines around target
--annotate            # Add inline reference comments
--highlight-changes    # Show git blame / recent changes
--metrics             # Include complexity metrics
```

**Example Usage**:
```bash
# Basic extraction
go-scope --root . --file pkg/service/user.go --line 128

# Deep review with context
go-scope --root . --file pkg/service/user.go --line 128 \
  --depth 2 \
  --format html \
  --show-callers \
  --metrics

# Quick signature view
go-scope --root . --file pkg/api/handler.go --line 45 \
  --depth 0 \
  --context-lines 5
```

### 3. Redesign Output Format

**Current**: Single compilable .go file or directory tree

**Revised**: Structured document with sections

**Markdown Output**:
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
// CreateUser creates a new user account with validation
func (s *UserService) CreateUser(ctx context.Context, email, name string) (*User, error) {
    if err := validateEmail(email); err != nil {  // → pkg/service/validation.go:45
        return nil, fmt.Errorf("invalid email: %w", err)
    }

    user := &User{
        Email: email,
        Name:  name,
        CreatedAt: time.Now(),
    }

    if err := s.db.Create(user); err != nil {  // → external: database/sql
        return nil, fmt.Errorf("failed to create user: %w", err)
    }

    return user, nil
}
```

---

## Direct Dependencies (depth 1)

### validateEmail
**File**: pkg/service/validation.go:45
**Kind**: Function

```go
// File: pkg/service/validation.go:45-52
func validateEmail(email string) error {
    if !strings.Contains(email, "@") {  // → stdlib: strings
        return errors.New("invalid email format")
    }
    return nil
}
```

### User
**File**: pkg/model/user.go:12
**Kind**: Type

```go
// File: pkg/model/user.go:12-18
type User struct {
    ID        int64
    Email     string
    Name      string
    CreatedAt time.Time
}
```

---

## External References

- **database/sql**: `DB.Create`
- **fmt**: `Errorf`
- **time**: `Now`
- **errors**: `New`
- **strings**: `Contains`

---

## Called By

1. **cmd/api/handlers.go:67** - `handleCreateUser`
2. **cmd/cli/user.go:123** - `createUserCommand`
3. **pkg/service/user_test.go:34** - `TestCreateUser`

---

## Recent Changes (Git)

- **2025-10-20** by alice@example.com - "Add email validation"
- **2025-10-15** by bob@example.com - "Initial implementation"

---

## Dependency Graph

```
CreateUser (target)
├── validateEmail (pkg/service)
│   ├── strings.Contains (stdlib)
│   └── errors.New (stdlib)
├── User (pkg/model)
│   └── time.Time (stdlib)
├── UserService.db (interface)
│   └── [external implementation]
└── fmt.Errorf (stdlib)
```
```

**HTML Output**: Same content with:
- Syntax highlighting
- Hyperlinks to file:line
- Collapsible sections
- Copy buttons
- Dark/light theme toggle

### 4. Simplify Architecture

**Current** (7 packages):
```
internal/extract/
  loader.go
  locator.go
  graph.go
  walker.go
  assemble_single.go  ❌ Remove
  assemble_multi.go   ❌ Remove
  manifest.go
  rewrite.go          ❌ Remove
  guard.go            ❌ Remove (limits not needed)
```

**Revised** (4 packages):
```
internal/extract/
  loader.go          // packages.Load wrapper
  locator.go         // position → symbol
  collector.go       // NEW: depth-limited dependency collection
  analysis.go        // NEW: caller analysis, metrics
  format/
    markdown.go      // NEW: markdown formatter
    html.go          // NEW: HTML formatter with syntax highlighting
    json.go          // manifest/structured output
```

### 5. Update API

**Current**:
```go
type Target struct {
  Root, File string
  Line, Col  int
  BuildTags  []string
  IncludeTests bool
}

type Result struct {
  Mode    string
  Files   []EmittedFile  // Compilable files
  Manifest Manifest
}

func Extract(ctx context.Context, t Target, opts Options) (Result, error)
```

**Revised**:
```go
type Target struct {
  Root   string  // Module root
  File   string  // Relative or absolute path
  Line   int     // 1-based line number
  Column int     // 1-based column (default: 1)
}

type Options struct {
  Depth          int      // Dependency depth (default: 1, 0 = target only)
  Format         string   // "markdown", "html", "json"
  StubExternal   bool     // Show signatures for external deps
  ShowCallers    bool     // Include reverse dependencies
  ShowTests      bool     // Include test functions
  ContextLines   int      // Extra lines around target
  Annotate       bool     // Add inline reference comments
  IncludeMetrics bool     // Compute complexity metrics
  GitBlame       bool     // Include git history
}

type Symbol struct {
  Package   string  // Full package path
  Name      string  // Symbol name
  Kind      string  // "func", "method", "type", "var", "const", "interface"
  File      string  // Source file path
  Line      int     // Definition start line
  EndLine   int     // Definition end line
  Code      string  // Source code
  Doc       string  // Documentation comment
}

type Reference struct {
  Symbol   Symbol   // Referenced symbol
  Reason   string   // "direct-call", "type-reference", "field-access", etc.
  Depth    int      // 0 = target, 1 = direct dep, etc.
  External bool     // True if from different module
  Stub     bool     // True if only signature included
}

type Caller struct {
  File     string
  Line     int
  Function string
  Context  string  // Surrounding code snippet
}

type Metrics struct {
  LinesOfCode         int
  CyclomaticComplexity int
  DependencyCount     int
  ExternalPackages    []string
}

type Extract struct {
  Target     Symbol       // The requested symbol
  References []Reference  // Included dependencies
  External   []string     // External package references (pkg.Symbol)
  Callers    []Caller     // What calls this symbol
  Metrics    *Metrics     // Optional metrics
  GitBlame   []GitBlame   // Optional git history
  Graph      string       // Dependency graph (mermaid or dot format)
}

type Result struct {
  Extract  Extract
  Rendered string  // Formatted output (markdown/html)
  Metadata Metadata
}

func ExtractSymbol(ctx context.Context, target Target, opts Options) (*Result, error)
```

### 6. Revise Test Strategy

**Remove** (Compilation-focused):
- ❌ "output file compiles"
- ❌ "go tool compile" tests
- ❌ "multi-file compile as module directory"
- ❌ Build tag validation
- ❌ go.mod generation tests

**Add** (Review-focused):
- ✅ Symbol extraction accuracy (correct func/method/type found)
- ✅ Depth limiting (stop at specified depth)
- ✅ Reference annotation (file:line markers correct)
- ✅ Caller discovery (all callers found)
- ✅ Format validation (markdown/HTML valid)
- ✅ Metrics accuracy (complexity calculations)
- ✅ External stub generation (correct signatures)

**BDD Scenarios** (Revised):

```gherkin
Feature: Extract function for review

  Scenario: Basic function extraction
    Given example module "ex1"
    When I extract "pkg/math/add.go" line 7 with depth 1
    Then output includes function "Add" definition
    And output includes helper "validateInputs" from same package
    And output marks "fmt.Println" as external reference
    And output excludes unused function "Sub"
    And output is valid markdown

  Scenario: Depth control
    Given example module "ex2"
    When I extract "cmd/api/handler.go" line 30 with depth 0
    Then output includes only "handleCreateUser" definition
    And output shows "UserService.CreateUser" call as stub

    When I extract same location with depth 2
    Then output includes "handleCreateUser" definition
    And output includes "UserService.CreateUser" implementation
    And output includes "validateEmail" helper
    And output stubs external "database/sql.DB.Create"

  Scenario: Caller analysis
    Given example module "ex2"
    When I extract "pkg/service/user.go" line 128 with --show-callers
    Then output includes "CreateUser" method
    And callers section includes "cmd/api/handlers.go:67"
    And callers section includes "cmd/cli/user.go:123"
    And callers section includes test "pkg/service/user_test.go:34"

  Scenario: Format variations
    Given example module "ex1"
    When I extract "pkg/math/add.go" line 7 with format "html"
    Then output is valid HTML
    And output has syntax highlighting
    And output has clickable file:line links

    When I extract same location with format "json"
    Then output is valid JSON
    And JSON includes target symbol details
    And JSON includes references array
    And JSON includes metrics object

  Scenario: Generic code
    Given example module "ex3"
    When I extract "pkg/iter/map.go" line 10
    Then output includes generic function "Map[T,U]"
    And output includes type constraint definitions
    And output compiles when type parameters provided

  Scenario: External stubs
    Given example module "ex2"
    When I extract "pkg/service/user.go" line 128 with --stub-external
    Then output includes local implementation details
    And output shows "db.Create" signature: "Create(value interface{}) error"
    And output does not include database implementation

  Scenario: Metrics inclusion
    Given example module "ex2"
    When I extract "pkg/service/user.go" line 128 with --metrics
    Then output includes lines of code count
    And output includes cyclomatic complexity
    And output includes dependency count
    And output includes list of external packages
```

### 7. Simplify Edge Cases

**Current** (Section 6): 14 detailed edge cases

**Revised**: 6 essential review cases

1. **Multiple decls with same name**: Choose the one at specified position
2. **Generated files**: Include by default (no skip logic needed)
3. **Interface methods**: Include interface definition; note implementations if --show-callers
4. **Generics**: Include type parameters and constraints as part of definition
5. **Unexported cross-package**: Note as "unexported from package X" and skip or stub
6. **Test symbols**: Include only with --show-tests

**Remove**:
- ❌ Cycles and topological sort
- ❌ init() inclusion logic
- ❌ cgo handling
- ❌ Build tag combinatorics
- ❌ Multi-module vendoring
- ❌ Unsafe and reflect special cases (just include if referenced)

### 8. Add New Features Section

**Section 20: Review-Optimized Features**

1. **Syntax Highlighting**
   - HTML: Chroma or highlight.js
   - Markdown: Fenced code blocks with language tags
   - Terminal: Color output with fatih/color

2. **Hyperlinked References**
   - HTML mode: `<a href="file://...">symbol</a>`
   - Markdown: Can't link to local files, use annotations instead
   - Terminal: OSC 8 hyperlinks for modern terminals

3. **Caller Discovery**
   - Use `types.Info.Uses` to build reverse index
   - Search all packages in module for references
   - Group by file, show context snippet

4. **Git Integration**
   - `git blame` for last author/date per line
   - `git log -p` for recent changes to symbol
   - `git diff` to highlight uncommitted changes

5. **Metrics**
   - Cyclomatic complexity: Count branches
   - Lines of Code: Physical vs logical
   - Dependency count: Direct + transitive
   - Maintainability index: Composite score

6. **Output Formats**
   - Markdown: GitHub-flavored, portable
   - HTML: Self-contained single file with inline CSS/JS
   - JSON: Structured for tool integration
   - Terminal: ANSI colored for piping
   - Diff: Compare two versions side-by-side

7. **Interactive Features** (HTML)
   - Collapsible sections
   - Search within extract
   - Copy code blocks
   - Toggle between "compact" and "detailed" view
   - Dark/light theme

8. **Export Options**
   - Save to file
   - Copy to clipboard
   - GitHub Gist upload
   - Pastebin upload
   - Share link generation

---

## Migration Strategy

### Phase 1: Core Extraction (Week 1-2)
- ✅ Implement loader.go (packages.Load)
- ✅ Implement locator.go (position → symbol)
- ✅ Implement collector.go (depth-limited dependency gathering)
- ✅ Basic markdown output
- ✅ CLI with essential flags (--root, --file, --line, --depth)

### Phase 2: Review Features (Week 3)
- ✅ Caller analysis
- ✅ External stub generation
- ✅ Annotation generation (file:line markers)
- ✅ HTML output with syntax highlighting

### Phase 3: Advanced Features (Week 4)
- ✅ Metrics calculation
- ✅ Git integration
- ✅ JSON manifest
- ✅ Interactive HTML features

### Phase 4: Polish (Week 5-6)
- ✅ Comprehensive tests
- ✅ Examples
- ✅ Documentation
- ✅ Performance optimization

---

## Risk Assessment

### Risks Removed by Simplification

| Risk | Original | Review-Focused |
|------|----------|----------------|
| Compilation failures | High | None |
| Import conflicts | Medium | None |
| Build tag issues | Medium | Low |
| Cross-platform problems | Medium | Low |
| Init order bugs | Medium | None |
| Circular dependencies | Medium | Low |
| go.mod generation | High | None |

### New Risks

1. **Incomplete Context**: Depth limit may exclude needed context
   - Mitigation: Make depth configurable, provide graph view

2. **Large Output**: Deep traversal + callers could be huge
   - Mitigation: Add output size warnings, summary mode

3. **Caller Analysis Performance**: Reverse index might be slow
   - Mitigation: Cache results, limit search scope

---

## Conclusion

The original specification is well-structured but fundamentally misaligned with the actual requirement. By pivoting from "compilable extraction" to "review-optimized extraction", we can:

- **Reduce complexity by ~60%**
- **Deliver 50% faster**
- **Add critical missing features** (callers, syntax highlighting, metrics)
- **Improve usability dramatically**

**Recommendation**: Adopt the review-focused specification detailed in `SPEC_v2_REVIEW_FOCUSED.md`.

---

## Next Steps

1. ✅ Review and approve this analysis
2. ⏳ Write revised specification (`SPEC_v2_REVIEW_FOCUSED.md`)
3. ⏳ Create detailed changelog (`CHANGELOG_v1_to_v2.md`)
4. ⏳ Update example projects
5. ⏳ Begin implementation with Phase 1
