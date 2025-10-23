package format

import (
	"strings"
	"testing"

	"github.com/extract-scope-go/go-scope/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFormatSimpleFunction tests formatting a single function
func TestFormatSimpleFunction(t *testing.T) {
	// Given: An extract with just a target symbol
	ext := types.Extract{
		Target: types.Symbol{
			Package:  "example.com/test/pkg",
			Name:     "Add",
			Kind:     "func",
			File:     "/path/to/add.go",
			Line:     10,
			EndLine:  15,
			Code:     "func Add(a, b int) int {\n\treturn a + b\n}",
			Doc:      "Add returns the sum of two integers",
			Exported: true,
		},
		References: []types.Reference{},
		External:   []string{},
	}

	// When: We format as markdown
	result, err := ToMarkdown(ext, types.Options{})

	// Then: Should have valid markdown with header and code block
	require.NoError(t, err)
	assert.Contains(t, result, "# Code Extract: Add")
	assert.Contains(t, result, "**File**: /path/to/add.go:10")
	assert.Contains(t, result, "**Kind**: func")
	assert.Contains(t, result, "## Target Symbol")
	assert.Contains(t, result, "```go")
	assert.Contains(t, result, "func Add(a, b int) int")
	assert.Contains(t, result, "Add returns the sum of two integers")
}

// TestFormatWithDependencies tests formatting with dependencies
func TestFormatWithDependencies(t *testing.T) {
	// Given: An extract with target + dependencies
	ext := types.Extract{
		Target: types.Symbol{
			Name: "Main",
			Kind: "func",
			File: "/path/to/main.go",
			Line: 5,
			Code: "func Main() { Helper() }",
		},
		References: []types.Reference{
			{
				Symbol: types.Symbol{
					Name:    "Helper",
					Kind:    "func",
					File:    "/path/to/helper.go",
					Line:    10,
					Code:    "func Helper() {}",
					Package: "example.com/test",
				},
				Depth:    1,
				External: false,
			},
		},
		External: []string{},
	}

	// When: We format as markdown
	result, err := ToMarkdown(ext, types.Options{Depth: 1})

	// Then: Should have target section and dependencies section
	require.NoError(t, err)
	assert.Contains(t, result, "## Target Symbol")
	assert.Contains(t, result, "## Dependencies")
	assert.Contains(t, result, "Helper")
	assert.Contains(t, result, "helper.go:10") // Uses basename
}

// TestFormatWithExternal tests formatting with external references
func TestFormatWithExternal(t *testing.T) {
	// Given: An extract with external references
	ext := types.Extract{
		Target: types.Symbol{
			Name: "PrintHello",
			Code: "func PrintHello() { fmt.Println() }",
		},
		External: []string{"fmt.Println", "os.Exit"},
	}

	// When: We format as markdown
	result, err := ToMarkdown(ext, types.Options{})

	// Then: Should have external references section
	require.NoError(t, err)
	assert.Contains(t, result, "## External References")
	assert.Contains(t, result, "fmt.Println")
	assert.Contains(t, result, "os.Exit")
}

// TestFormatWithCallers tests formatting with caller information
func TestFormatWithCallers(t *testing.T) {
	// Given: An extract with callers
	ext := types.Extract{
		Target: types.Symbol{
			Name: "DoWork",
		},
		Callers: []types.Caller{
			{
				File:     "/path/to/main.go",
				Line:     25,
				Function: "main",
				Context:  "DoWork()",
			},
			{
				File:     "/path/to/test.go",
				Line:     10,
				Function: "TestDoWork",
				Context:  "DoWork()",
			},
		},
	}

	// When: We format as markdown with show-callers
	result, err := ToMarkdown(ext, types.Options{ShowCallers: true})

	// Then: Should have callers section
	require.NoError(t, err)
	assert.Contains(t, result, "## Called By")
	assert.Contains(t, result, "main.go:25") // Uses basename
	assert.Contains(t, result, "main")
	assert.Contains(t, result, "test.go:10") // Uses basename
	assert.Contains(t, result, "TestDoWork")
}

// TestFormatWithMetrics tests formatting with complexity metrics
func TestFormatWithMetrics(t *testing.T) {
	// Given: An extract with metrics
	ext := types.Extract{
		Target: types.Symbol{
			Name: "ComplexFunc",
		},
		Metrics: &types.Metrics{
			LinesOfCode:          45,
			CyclomaticComplexity: 8,
			DependencyCount:      4,
			ExternalPackages:     []string{"fmt", "os"},
		},
	}

	// When: We format with metrics enabled
	result, err := ToMarkdown(ext, types.Options{IncludeMetrics: true})

	// Then: Should have metrics section
	require.NoError(t, err)
	assert.Contains(t, result, "## Metrics")
	assert.Contains(t, result, "Lines of Code: 45")
	assert.Contains(t, result, "Cyclomatic Complexity: 8")
	assert.Contains(t, result, "Dependencies: 4")
}

// TestFormatMarkdownSyntax tests that output is valid markdown
func TestFormatMarkdownSyntax(t *testing.T) {
	// Given: A simple extract
	ext := types.Extract{
		Target: types.Symbol{
			Name: "Test",
			Code: "func Test() {}",
		},
	}

	// When: We format as markdown
	result, err := ToMarkdown(ext, types.Options{})

	// Then: Should have proper markdown structure
	require.NoError(t, err)

	// Should start with # heading
	lines := strings.Split(result, "\n")
	assert.True(t, strings.HasPrefix(lines[0], "#"), "Should start with markdown heading")

	// Should have code fence
	assert.Contains(t, result, "```go")
	assert.Contains(t, result, "```")

	// Should have markdown bold syntax
	assert.Contains(t, result, "**")
}

// TestFormatEmptyExtract tests handling of minimal extract
func TestFormatEmptyExtract(t *testing.T) {
	// Given: An extract with minimal info
	ext := types.Extract{
		Target: types.Symbol{
			Name: "Minimal",
		},
	}

	// When: We format as markdown
	result, err := ToMarkdown(ext, types.Options{})

	// Then: Should not error, should have basic structure
	require.NoError(t, err)
	assert.Contains(t, result, "Minimal")
	assert.NotEmpty(t, result)
}

// TestFormatDepthGrouping tests that dependencies are grouped by depth
func TestFormatDepthGrouping(t *testing.T) {
	// Given: Dependencies at different depths
	ext := types.Extract{
		Target: types.Symbol{
			Name: "Target",
		},
		References: []types.Reference{
			{
				Symbol: types.Symbol{Name: "Dep1"},
				Depth:  1,
			},
			{
				Symbol: types.Symbol{Name: "Dep2"},
				Depth:  1,
			},
			{
				Symbol: types.Symbol{Name: "SubDep1"},
				Depth:  2,
			},
		},
	}

	// When: We format with depth 2
	result, err := ToMarkdown(ext, types.Options{Depth: 2})

	// Then: Should group by depth
	require.NoError(t, err)

	// Check that depth 1 and depth 2 are clearly separated
	depth1Pos := strings.Index(result, "Dep1")
	depth2Pos := strings.Index(result, "SubDep1")

	assert.True(t, depth1Pos < depth2Pos, "Depth 1 deps should come before depth 2")
}
