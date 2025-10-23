package extract

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLocateFunctionAtLine tests finding a function declaration at a specific line
func TestLocateFunctionAtLine(t *testing.T) {
	// Given: Example 1 with Add function at line 7
	root := filepath.Join("..", "..", "examples", "ex1")
	file := filepath.Join(root, "pkg", "math", "add.go")
	line := 7 // func Add(a, b int) int {

	// When: We locate the symbol at that line
	locator := NewLocator()
	symbol, err := locator.LocateSymbol(root, file, line, 1)

	// Then: Should find the Add function
	require.NoError(t, err)
	assert.NotNil(t, symbol)
	assert.Equal(t, "Add", symbol.Name)
	assert.Equal(t, "func", symbol.Kind)
	assert.Equal(t, "example.com/ex1/pkg/math", symbol.Package)
	assert.True(t, symbol.Exported)
	assert.Contains(t, symbol.Code, "func Add(a, b int) int")
}

// TestLocateUnexportedFunction tests finding an unexported function
func TestLocateUnexportedFunction(t *testing.T) {
	// Given: Example 1 with validateInputs function at line 4 of util.go
	root := filepath.Join("..", "..", "examples", "ex1")
	file := filepath.Join(root, "pkg", "math", "util.go")
	line := 4 // func validateInputs(a, b int) bool {

	// When: We locate the symbol at that line
	locator := NewLocator()
	symbol, err := locator.LocateSymbol(root, file, line, 1)

	// Then: Should find the validateInputs function
	require.NoError(t, err)
	assert.NotNil(t, symbol)
	assert.Equal(t, "validateInputs", symbol.Name)
	assert.Equal(t, "func", symbol.Kind)
	assert.False(t, symbol.Exported)
}

// TestLocateSymbolNotFound tests behavior when no symbol exists at line
func TestLocateSymbolNotFound(t *testing.T) {
	// Given: A blank line or comment in the file
	root := filepath.Join("..", "..", "examples", "ex1")
	file := filepath.Join(root, "pkg", "math", "add.go")
	line := 4 // import "fmt" - not a symbol

	// When: We try to locate symbol at that line
	locator := NewLocator()
	symbol, err := locator.LocateSymbol(root, file, line, 1)

	// Then: Should return an error
	assert.Error(t, err)
	assert.Nil(t, symbol)
	// Error message can be either "no symbol found" or "no extractable symbol found"
	assert.Contains(t, err.Error(), "no")
}

// TestLocateInvalidFile tests behavior with invalid file path
func TestLocateInvalidFile(t *testing.T) {
	// Given: A non-existent file
	root := filepath.Join("..", "..", "examples", "ex1")
	file := filepath.Join(root, "pkg", "math", "nonexistent.go")
	line := 10

	// When: We try to locate symbol
	locator := NewLocator()
	symbol, err := locator.LocateSymbol(root, file, line, 1)

	// Then: Should return an error
	assert.Error(t, err)
	assert.Nil(t, symbol)
}

// TestLocateInvalidRoot tests behavior with invalid module root
func TestLocateInvalidRoot(t *testing.T) {
	// Given: A non-existent root directory
	root := "/nonexistent/path"
	file := "pkg/math/add.go"
	line := 7

	// When: We try to locate symbol
	locator := NewLocator()
	symbol, err := locator.LocateSymbol(root, file, line, 1)

	// Then: Should return an error
	assert.Error(t, err)
	assert.Nil(t, symbol)
}
