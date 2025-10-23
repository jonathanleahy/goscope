package extract

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/extract-scope-go/go-scope/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCollectDepth0 tests collecting only the target symbol (no dependencies)
func TestCollectDepth0(t *testing.T) {
	// Given: Example 1 with Add function that calls validateInputs
	root := filepath.Join("..", "..", "examples", "ex1")
	file := filepath.Join(root, "pkg", "math", "add.go")
	line := 7 // func Add(a, b int) int

	// When: We collect dependencies with depth 0
	result, err := ExtractAndFormat(context.Background(), types.Target{
		Root:   root,
		File:   file,
		Line:   line,
		Column: 1,
	}, types.Options{
		Depth: 0,
	})

	// Then: Should include only the target (Add function)
	require.NoError(t, err)
	assert.Equal(t, "Add", result.Extract.Target.Name)
	assert.Equal(t, "func", result.Extract.Target.Kind)

	// Should have NO dependencies (depth 0)
	assert.Len(t, result.Extract.References, 0)

	// External references should be noted but not included
	// Add calls validateInputs (internal) and fmt.Println (external)
	// At depth 0, we don't include either
}

// TestCollectDepth1 tests collecting target + direct dependencies
func TestCollectDepth1(t *testing.T) {
	// Given: Example 1 with Add function
	root := filepath.Join("..", "..", "examples", "ex1")
	file := filepath.Join(root, "pkg", "math", "add.go")
	line := 7 // func Add(a, b int) int

	// When: We collect dependencies with depth 1
	result, err := ExtractAndFormat(context.Background(), types.Target{
		Root:   root,
		File:   file,
		Line:   line,
		Column: 1,
	}, types.Options{
		Depth: 1,
	})

	// Then: Should include target + direct dependencies
	require.NoError(t, err)
	assert.Equal(t, "Add", result.Extract.Target.Name)

	// Should include validateInputs (depth 1, same package)
	assert.GreaterOrEqual(t, len(result.Extract.References), 1)

	foundValidate := false
	for _, ref := range result.Extract.References {
		if ref.Symbol.Name == "validateInputs" {
			foundValidate = true
			assert.Equal(t, 1, ref.Depth, "validateInputs should be at depth 1")
			assert.False(t, ref.External, "validateInputs is in same module")
			assert.False(t, ref.Stub, "validateInputs should have full code")
		}
	}
	assert.True(t, foundValidate, "Should have found validateInputs dependency")

	// External references like fmt.Println should be noted
	assert.Contains(t, result.Extract.External, "fmt.Println")
}

// TestCollectDepth2 tests transitive dependencies
func TestCollectDepth2(t *testing.T) {
	// Given: Example 1 with Add -> validateInputs -> (no further deps)
	root := filepath.Join("..", "..", "examples", "ex1")
	file := filepath.Join(root, "pkg", "math", "add.go")
	line := 7

	// When: We collect with depth 2
	result, err := ExtractAndFormat(context.Background(), types.Target{
		Root:   root,
		File:   file,
		Line:   line,
		Column: 1,
	}, types.Options{
		Depth: 2,
	})

	// Then: Should include target + deps + their deps
	require.NoError(t, err)
	assert.Equal(t, "Add", result.Extract.Target.Name)

	// validateInputs has no further local dependencies, so depth 2 = same as depth 1
	// But we should still traverse and mark them correctly
	for _, ref := range result.Extract.References {
		assert.LessOrEqual(t, ref.Depth, 2, "No dependency should exceed depth 2")
	}
}

// TestCollectExternalReference tests handling of external packages
func TestCollectExternalReference(t *testing.T) {
	// Given: Add function uses fmt.Println (stdlib)
	root := filepath.Join("..", "..", "examples", "ex1")
	file := filepath.Join(root, "pkg", "math", "add.go")
	line := 7

	// When: We collect with depth 1 and stub_external = true
	result, err := ExtractAndFormat(context.Background(), types.Target{
		Root:   root,
		File:   file,
		Line:   line,
		Column: 1,
	}, types.Options{
		Depth:        1,
		StubExternal: true,
	})

	// Then: External references should be noted, not fully included
	require.NoError(t, err)

	// fmt.Println should be in external list
	assert.Contains(t, result.Extract.External, "fmt.Println")

	// Should not have full fmt.Println code (it's external)
	for _, ref := range result.Extract.References {
		if ref.Symbol.Name == "Println" {
			assert.True(t, ref.External, "fmt.Println is external")
			assert.True(t, ref.Stub, "External should be stubbed")
		}
	}
}

// TestCollectCircularDependency tests handling of circular references
func TestCollectCircularDependency(t *testing.T) {
	// Note: Our example doesn't have circular deps, but we should handle them
	// This test will be added when we create an example with circular deps
	t.Skip("TODO: Create example with circular dependency")
}

// TestCollectDeduplication tests that symbols aren't collected multiple times
func TestCollectDeduplication(t *testing.T) {
	// Given: A function that calls the same helper multiple times
	// (Add calls validateInputs once, but this tests general dedup logic)
	root := filepath.Join("..", "..", "examples", "ex1")
	file := filepath.Join(root, "pkg", "math", "add.go")
	line := 7

	// When: We collect dependencies
	result, err := ExtractAndFormat(context.Background(), types.Target{
		Root:   root,
		File:   file,
		Line:   line,
		Column: 1,
	}, types.Options{
		Depth: 1,
	})

	// Then: Each symbol should appear only once
	require.NoError(t, err)

	symbolNames := make(map[string]int)
	for _, ref := range result.Extract.References {
		symbolNames[ref.Symbol.Name]++
	}

	for name, count := range symbolNames {
		assert.Equal(t, 1, count, "Symbol %s should appear exactly once, but appears %d times", name, count)
	}
}

// TestCollectUnexportedSymbol tests including unexported symbols from same package
func TestCollectUnexportedSymbol(t *testing.T) {
	// Given: Add function calls unexported validateInputs
	root := filepath.Join("..", "..", "examples", "ex1")
	file := filepath.Join(root, "pkg", "math", "add.go")
	line := 7

	// When: We collect with depth 1
	result, err := ExtractAndFormat(context.Background(), types.Target{
		Root:   root,
		File:   file,
		Line:   line,
		Column: 1,
	}, types.Options{
		Depth: 1,
	})

	// Then: Should include unexported validateInputs (same package)
	require.NoError(t, err)

	found := false
	for _, ref := range result.Extract.References {
		if ref.Symbol.Name == "validateInputs" {
			found = true
			assert.False(t, ref.Symbol.Exported, "validateInputs is unexported")
			assert.False(t, ref.External, "validateInputs is same package")
		}
	}
	assert.True(t, found, "Should include unexported symbol from same package")
}
