package extract

import (
	"path/filepath"
	"testing"
)

// TestDebugLocator helps debug the locator functionality
func TestDebugLocator(t *testing.T) {
	root := filepath.Join("..", "..", "examples", "ex1")
	file := filepath.Join(root, "pkg", "math", "add.go")

	t.Logf("Root: %s", root)
	t.Logf("File: %s", file)

	locator := NewLocator()

	// Try to load packages
	err := locator.loadPackages(root, file)
	if err != nil {
		t.Fatalf("Failed to load packages: %v", err)
	}

	t.Logf("Loaded %d packages", len(locator.pkgs))
	for i, pkg := range locator.pkgs {
		t.Logf("  Package %d: %s (path: %s)", i, pkg.Name, pkg.PkgPath)
		t.Logf("    Files: %d", len(pkg.CompiledGoFiles))
		for j, f := range pkg.CompiledGoFiles {
			t.Logf("      File %d: %s", j, f)
		}
	}

	// Try to find the file
	pkg, astFile := locator.findFileInPackages(file)
	if pkg == nil || astFile == nil {
		t.Fatal("Could not find file in packages")
	}

	t.Logf("Found file in package: %s", pkg.PkgPath)

	// Try to find position at line 7
	pos := locator.findPosition(astFile, 7, 1)
	if !pos.IsValid() {
		t.Fatal("Could not find valid position")
	}

	t.Logf("Position for line 7: %v", locator.fset.Position(pos))

	// Try to find enclosing node
	path, exact := locator.findEnclosingNode(astFile, pos)
	t.Logf("Path length: %d", len(path))
	if exact != nil {
		t.Logf("Exact node: %T at %v", exact, locator.fset.Position(exact.Pos()))
	}

	for i, node := range path {
		pos := locator.fset.Position(node.Pos())
		t.Logf("  Path[%d]: %T at line %d", i, node, pos.Line)
	}
}
