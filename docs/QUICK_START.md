# Go Scope Extractor - Quick Reference

## What It Does

**Extract the functional context around any Go symbol to understand what's happening.**

Given a line in a Go file, extract:
- ✅ The code at that location (function, method, type, etc.)
- ✅ What it references (functions it calls, types it uses)
- ✅ Where it's used (what calls it)
- ✅ How complex it is (metrics)
- ✅ Who wrote it and when (git history)

**Output**: Clean, readable document optimized for human understanding.

---

## Core Concept

```
You: "What's happening at user.go:128?"

go-scope:
  [Shows you CreateUser method]
  [Shows validateEmail helper it calls]
  [Shows User type it creates]
  [Shows what calls CreateUser]
  [Notes external packages used]
```

**NOT**: A compilable program
**IS**: A readable code review document

---

## Quick Examples

### Example 1: Basic Extraction
```bash
# "What does this function do?"
go-scope --root . --file pkg/service/user.go --line 128
```

**Output**:
```markdown
# CreateUser

## Target
func (s *UserService) CreateUser(...) (*User, error) {
    // Full implementation
}

## Uses (depth 1)
- validateEmail (pkg/service/validation.go:45)
- User (pkg/model/user.go:12)
- fmt.Errorf (stdlib)
- database/sql.DB.Create (external)

## Called By
- handleCreateUser (cmd/api/handlers.go:67)
- TestCreateUser (pkg/service/user_test.go:34)
```

### Example 2: Deeper Context
```bash
# "Show me more context - what does validateEmail do?"
go-scope --root . --file pkg/service/user.go --line 128 --depth 2
```

Now includes:
- CreateUser (target)
- validateEmail **full implementation** (depth 1)
- User type **full definition** (depth 1)
- strings.Contains (depth 2, marked as stdlib)

### Example 3: Complete Understanding
```bash
# "Full context please"
go-scope --root . --file pkg/service/user.go --line 128 \
  --depth 2 \
  --show-callers \
  --metrics \
  --format html \
  --output review.html
```

Gets you:
- Target code + 2 levels of dependencies
- All callers (who uses this?)
- Complexity metrics (how complex is this?)
- Beautiful HTML with syntax highlighting
- Clickable file:line links

---

## Key Flags

### Essential
- `--root PATH` - Project root (required)
- `--file PATH` - File to extract from (required)
- `--line N` - Line number (required)

### Control Context
- `--depth N` - How deep to follow dependencies
  - `0` = Just show me the target
  - `1` = Target + what it directly uses (default)
  - `2+` = Keep going deeper

### What to Include
- `--show-callers` - Where is this used?
- `--show-tests` - Include test functions
- `--metrics` - How complex is this?
- `--git-blame` - Who wrote this and when?

### Output Format
- `--format markdown` - Text (default)
- `--format html` - Pretty, with syntax highlighting
- `--format json` - For tools/scripts

---

## Depth Explained

Think of depth as "how many hops away am I willing to look":

### Depth 0: "Just this function"
```go
func CreateUser(...) {
    validateEmail(email)  // ← Not included, just shows it's called
    ...
}
```

### Depth 1: "This + what it uses"
```go
func CreateUser(...) {
    validateEmail(email)  // ← INCLUDED BELOW
    ...
}

// Depth 1 dependency:
func validateEmail(email string) error {
    strings.Contains(...)  // ← Not included, just noted
}
```

### Depth 2: "This + dependencies + their dependencies"
```go
func CreateUser(...) {
    validateEmail(email)
    ...
}

func validateEmail(email string) error {
    strings.Contains(...)  // ← Now we'd show this too (if local)
}
```

**Rule of thumb**:
- Depth 1 for most cases
- Depth 2 when you're confused about helpers
- Depth 0 when you just want to see one thing

---

## Output Structure

Every extraction has these sections:

### 1. Target Symbol
The code you asked for, with annotations:
```go
// File: pkg/service/user.go:128-145
func (s *UserService) CreateUser(...) {
    if err := validateEmail(email); err != nil {  // → validation.go:45
        ...
    }
}
```

### 2. Dependencies
Code this uses (controlled by --depth):
```markdown
### validateEmail
File: pkg/service/validation.go:45
[Full code or signature]
```

### 3. External References
Packages you import:
```markdown
- database/sql: DB.Create
- fmt: Errorf
```

### 4. Called By (if --show-callers)
Where this is used:
```markdown
- cmd/api/handlers.go:67 - handleCreateUser
- pkg/service/user_test.go:34 - TestCreateUser
```

### 5. Metrics (if --metrics)
```markdown
- Lines of Code: 17
- Cyclomatic Complexity: 3
- Dependencies: 4
```

### 6. Git History (if --git-blame)
```markdown
- 2025-10-20 by alice - "Add validation"
- 2025-10-15 by bob - "Initial implementation"
```

---

## Common Use Cases

### 1. "I found a bug, what does this code do?"
```bash
go-scope --root . --file buggy.go --line 45 --depth 1
```
Shows the buggy function + what it calls.

### 2. "Where is this function used?"
```bash
go-scope --root . --file service.go --line 100 --show-callers
```
Shows the function + all its callers.

### 3. "How complex is this code?"
```bash
go-scope --root . --file complex.go --line 200 --metrics
```
Shows complexity metrics.

### 4. "I need to review this PR"
```bash
go-scope --root . --file changed.go --line 50 \
  --depth 2 \
  --show-callers \
  --metrics \
  --git-blame \
  --format html \
  --output pr-review.html
```
Complete review document with all context.

### 5. "Share this code with my team"
```bash
go-scope --root . --file interesting.go --line 75 \
  --format html \
  --output share.html
# Email share.html - it's self-contained!
```

### 6. "Understand unfamiliar codebase"
```bash
# Start with main
go-scope --root . --file cmd/main.go --line 10 --show-callers

# Follow interesting functions
go-scope --root . --file pkg/handler.go --line 45 --depth 2
```

---

## Tips & Tricks

### Start Shallow, Go Deeper
```bash
# First pass - quick look
go-scope --root . --file foo.go --line 10 --depth 0

# Still confused? Go deeper
go-scope --root . --file foo.go --line 10 --depth 1

# Need more? Keep going
go-scope --root . --file foo.go --line 10 --depth 2
```

### Use HTML for Big Extracts
```bash
# Markdown is great for small extracts
go-scope --root . --file foo.go --line 10

# HTML is better when there's lots of code
go-scope --root . --file foo.go --line 10 --depth 3 --format html
```

### Combine with grep
```bash
# Find interesting function
grep -rn "ProcessPayment" .

# Extract it
go-scope --root . --file pkg/payment.go --line 234 --depth 2
```

### Save for Later
```bash
# Generate dated review files
go-scope --root . --file service.go --line 50 \
  --format html \
  --output "review-$(date +%Y%m%d).html"
```

---

## What You Get vs Don't Get

### ✅ You Get
- **Readable code** formatted nicely
- **Context** - what calls what
- **Annotations** - file:line markers
- **Metrics** - complexity, size
- **History** - who wrote it
- **Structure** - organized sections

### ❌ You Don't Get
- Compilable program
- Complete application
- All files in the project
- Build artifacts
- Running code
- Test execution results

---

## Understanding the Output

### Annotations
```go
func Foo() {
    Bar()  // → utils.go:45
}
```
The `→ utils.go:45` tells you where `Bar` is defined.

### Depth Markers
```markdown
## Direct Dependencies (depth 1)
### Bar
...

## Transitive Dependencies (depth 2)
### Baz (used by Bar)
...
```

### External vs Local
```markdown
## Dependencies
- validateEmail (local - full code shown)

## External References
- fmt.Errorf (stdlib - just signature)
```

Local = same module, you'll see full code
External = other module/stdlib, you'll see signature or note

---

## Performance

Typical extractions are **fast**:
- Depth 0-1: < 1 second
- Depth 2: < 3 seconds
- Depth 3+: < 10 seconds

If it's slow:
1. Reduce depth
2. Large file? That's OK, but might take a moment
3. First run loads packages (subsequent runs are cached)

---

## Troubleshooting

### "Symbol not found at position"
**Problem**: That line has no extractable symbol (comment, blank line, etc.)

**Solution**: Try nearby lines. The tool will suggest nearby symbols.

### "Unexported symbol from other package"
**Problem**: Can't access unexported symbols across packages.

**Solution**: Extract from that package directly, or the symbol will be noted but not included.

### "Package load failed"
**Problem**: Project has compilation errors.

**Solution**: Fix compilation errors first, or the tool can't analyze types.

### Output is too large
**Problem**: Depth too high, too many dependencies.

**Solution**: Reduce `--depth`, or use `--format html` for better navigation.

### Output is too small
**Problem**: Depth too low, missing context.

**Solution**: Increase `--depth`, add `--show-callers`.

---

## Philosophy

> **"Show me what I need to understand what's happening, nothing more."**

This tool is about **comprehension**, not compilation:
- No build systems
- No import paths to fix
- No vendor directories
- No compilation errors

Just: **"Here's the code, here's what it uses, here's what uses it."**

---

## Real-World Example

**Scenario**: Debugging a bug in user creation.

**Step 1**: Find the code
```bash
grep -rn "CreateUser" .
# Found: pkg/service/user.go:128
```

**Step 2**: Extract with context
```bash
go-scope --root . --file pkg/service/user.go --line 128 --depth 1 --show-callers
```

**Output shows**:
- `CreateUser` method (target)
- `validateEmail` helper it calls (depth 1)
- `User` type it creates (depth 1)
- `handleCreateUser` API handler that calls it (caller)
- `TestCreateUser` test (caller)

**Step 3**: Spot the issue
```go
func validateEmail(email string) error {
    if !strings.Contains(email, "@") {  // ← BUG: Too simple!
        return errors.New("invalid email")
    }
}
```

**Step 4**: Share with team
```bash
go-scope --root . --file pkg/service/user.go --line 128 \
  --depth 1 \
  --show-callers \
  --format html \
  --output bug-analysis.html
```

Email `bug-analysis.html` with your analysis. Team sees full context immediately.

---

## Next Steps

1. **Install** (once implemented):
   ```bash
   go install github.com/you/go-scope/cmd/go-scope@latest
   ```

2. **Try basic extraction**:
   ```bash
   cd your-go-project
   go-scope --root . --file path/to/file.go --line 100
   ```

3. **Experiment with depth**:
   ```bash
   # Start shallow
   go-scope --root . --file path/to/file.go --line 100 --depth 0

   # Add context as needed
   go-scope --root . --file path/to/file.go --line 100 --depth 1
   go-scope --root . --file path/to/file.go --line 100 --depth 2
   ```

4. **Add features**:
   ```bash
   go-scope --root . --file path/to/file.go --line 100 \
     --depth 1 \
     --show-callers \
     --metrics
   ```

5. **Generate review**:
   ```bash
   go-scope --root . --file path/to/file.go --line 100 \
     --depth 2 \
     --show-callers \
     --metrics \
     --git-blame \
     --format html \
     --output review.html
   ```

---

## Cheat Sheet

```bash
# Basic
go-scope --root . --file FILE --line N

# With context
go-scope --root . --file FILE --line N --depth 2

# With callers
go-scope --root . --file FILE --line N --show-callers

# Full analysis
go-scope --root . --file FILE --line N \
  --depth 2 --show-callers --metrics --git-blame

# Pretty HTML
go-scope --root . --file FILE --line N \
  --format html --output out.html

# JSON for scripts
go-scope --root . --file FILE --line N --format json
```

---

**Remember**: This tool is about **understanding code**, not **running code**. You're extracting the functional context to comprehend what's happening at a specific location.
