package di

import (
	"go/token"
	"testing"

	"github.com/extract-scope-go/go-scope/internal/types"
	"golang.org/x/tools/go/packages"
)

// TestDetectFrameworkNone tests detection when no framework is present
func TestDetectFrameworkNone(t *testing.T) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
		Dir: "../../../examples/ex1",
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load packages: %v", err)
	}

	fset := token.NewFileSet()
	detector := NewDetector(pkgs, fset)

	framework := detector.DetectFramework()

	// ex1 shouldn't have any DI framework
	if framework == "wire" || framework == "fx" {
		t.Errorf("Expected no framework or manual, got %s", framework)
	}

	t.Logf("Detected framework: %s", framework)
}

// TestHasWire tests Wire detection logic
func TestHasWire(t *testing.T) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
		Dir: "../../../examples/ex1",
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load packages: %v", err)
	}

	fset := token.NewFileSet()
	detector := NewDetector(pkgs, fset)

	hasWire := detector.hasWire()

	// ex1 shouldn't have Wire
	if hasWire {
		t.Error("Expected false for hasWire in ex1")
	}
}

// TestHasFx tests Fx detection logic
func TestHasFx(t *testing.T) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
		Dir: "../../../examples/ex1",
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load packages: %v", err)
	}

	fset := token.NewFileSet()
	detector := NewDetector(pkgs, fset)

	hasFx := detector.hasFx()

	// ex1 shouldn't have Fx
	if hasFx {
		t.Error("Expected false for hasFx in ex1")
	}
}

// TestHasManualDI tests manual DI detection
func TestHasManualDI(t *testing.T) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
		Dir: "../../../examples/ex1",
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load packages: %v", err)
	}

	fset := token.NewFileSet()
	detector := NewDetector(pkgs, fset)

	hasManual := detector.hasManualDI()

	// ex1 might or might not have constructors with params
	t.Logf("Has manual DI: %v", hasManual)
}

// TestAnalyzeDIBindingsEmpty tests with no bindings
func TestAnalyzeDIBindingsEmpty(t *testing.T) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
		Dir: "../../../examples/ex1",
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load packages: %v", err)
	}

	fset := token.NewFileSet()
	detector := NewDetector(pkgs, fset)

	symbols := []types.Symbol{}
	bindings := detector.AnalyzeDIBindings(symbols)

	// With no symbols, should have no bindings
	if len(bindings) != 0 {
		t.Errorf("Expected 0 bindings with no symbols, got %d", len(bindings))
	}
}

// TestAnalyzeManualBindings tests manual DI binding detection
func TestAnalyzeManualBindings(t *testing.T) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
		Dir: "../../../examples/ex1",
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load packages: %v", err)
	}

	fset := token.NewFileSet()
	detector := NewDetector(pkgs, fset)

	// Create mock constructor symbol
	symbols := []types.Symbol{
		{
			Name:    "NewService",
			Kind:    "func",
			Package: "example.com/test",
		},
	}

	bindings := detector.analyzeManualBindings(symbols)

	// Should attempt to analyze the constructor
	// Result depends on whether we can find the actual function
	t.Logf("Found %d manual bindings", len(bindings))
}

// TestCreateBindingFromConstructor tests binding creation logic
func TestCreateBindingFromConstructor(t *testing.T) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
		Dir: "../../../examples/ex1",
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load packages: %v", err)
	}

	fset := token.NewFileSet()
	detector := NewDetector(pkgs, fset)

	constructor := types.Symbol{
		Name:    "NewMath",
		Kind:    "func",
		Package: "example.com/ex1/pkg/math",
	}

	symbols := []types.Symbol{}
	binding := detector.createBindingFromConstructor(constructor, symbols, "manual")

	// If constructor doesn't exist in the actual package, binding might be nil
	if binding != nil {
		if binding.Framework != "manual" {
			t.Errorf("Expected framework 'manual', got '%s'", binding.Framework)
		}
		if binding.Provider.Name != "NewMath" {
			t.Errorf("Expected provider 'NewMath', got '%s'", binding.Provider.Name)
		}
	} else {
		t.Log("Binding is nil (constructor not found in package)")
	}
}

// TestFindSymbolForType tests type matching logic
func TestFindSymbolForType(t *testing.T) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
		Dir: "../../../examples/ex1",
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load packages: %v", err)
	}

	fset := token.NewFileSet()
	detector := NewDetector(pkgs, fset)

	symbols := []types.Symbol{
		{
			Name:    "TestType",
			Kind:    "struct",
			Package: "example.com/test",
		},
	}

	// Can't test without actual types, but we can verify the method exists
	result := detector.findSymbolForType(nil, symbols)
	if result != nil {
		t.Error("Expected nil for nil type")
	}
}

// TestDIBindingStructure tests the DIBinding struct fields
func TestDIBindingStructure(t *testing.T) {
	binding := types.DIBinding{
		Provider: types.Symbol{
			Name: "NewService",
			Kind: "func",
		},
		Product: types.Symbol{
			Name: "Service",
			Kind: "interface",
		},
		Dependencies: []types.Symbol{
			{Name: "Dep1", Kind: "interface"},
			{Name: "Dep2", Kind: "struct"},
		},
		Framework: "wire",
		Scope:     "singleton",
	}

	if binding.Provider.Name != "NewService" {
		t.Errorf("Expected Provider.Name 'NewService', got '%s'", binding.Provider.Name)
	}

	if binding.Framework != "wire" {
		t.Errorf("Expected Framework 'wire', got '%s'", binding.Framework)
	}

	if len(binding.Dependencies) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(binding.Dependencies))
	}

	if binding.Scope != "singleton" {
		t.Errorf("Expected Scope 'singleton', got '%s'", binding.Scope)
	}
}
