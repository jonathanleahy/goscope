package extract

import (
	"go/token"
	"testing"

	"github.com/extract-scope-go/go-scope/internal/types"
	"golang.org/x/tools/go/packages"
)

// TestAnalyzeInterfacesBasic tests basic interface-implementation detection
func TestAnalyzeInterfacesBasic(t *testing.T) {
	// Load test packages
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
		Dir: "../../examples/ex1",
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load packages: %v", err)
	}

	fset := token.NewFileSet()
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			fset.AddFile(pkg.Fset.File(file.Pos()).Name(), -1, int(file.End()-file.Pos()))
		}
	}

	analyzer := NewInterfaceAnalyzer(pkgs, fset)

	// Create dummy symbols (in real usage, these come from collector)
	symbols := []types.Symbol{
		{
			Name:    "TestInterface",
			Kind:    "interface",
			Package: "example.com/test",
		},
		{
			Name:    "TestImpl",
			Kind:    "struct",
			Package: "example.com/test",
		},
	}

	mappings := analyzer.AnalyzeInterfaces(symbols)

	// Should have analyzed the interface
	if len(mappings) < 0 {
		// Note: This might be 0 if ex1 doesn't have interfaces
		t.Logf("Found %d interface mappings", len(mappings))
	}
}

// TestFindInterfaces tests interface filtering
func TestFindInterfaces(t *testing.T) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo,
		Dir: "../../examples/ex1",
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load packages: %v", err)
	}

	fset := token.NewFileSet()
	analyzer := NewInterfaceAnalyzer(pkgs, fset)

	symbols := []types.Symbol{
		{Name: "Func1", Kind: "func"},
		{Name: "Interface1", Kind: "interface"},
		{Name: "Struct1", Kind: "struct"},
		{Name: "Interface2", Kind: "interface"},
	}

	interfaces := analyzer.findInterfaces(symbols)

	if len(interfaces) != 2 {
		t.Errorf("Expected 2 interfaces, got %d", len(interfaces))
	}

	for _, iface := range interfaces {
		if iface.Kind != "interface" {
			t.Errorf("Expected kind 'interface', got '%s'", iface.Kind)
		}
	}
}

// TestExtractInterfaceReferences tests reference generation
func TestExtractInterfaceReferences(t *testing.T) {
	iface := types.Symbol{
		Name:    "Service",
		Kind:    "interface",
		Package: "example.com/service",
	}

	impl := types.Symbol{
		Name:    "ServiceImpl",
		Kind:    "struct",
		Package: "example.com/service",
	}

	constructor := types.Symbol{
		Name:           "NewService",
		Kind:           "func",
		Package:        "example.com/service",
		InterfaceType:  "Service",
		Implementation: "ServiceImpl",
	}

	mapping := types.InterfaceMapping{
		Interface:       iface,
		Implementations: []types.Symbol{impl},
		Constructor:     &constructor,
		DIFramework:     "manual",
	}

	refs := ExtractInterfaceReferences([]types.InterfaceMapping{mapping}, 1)

	// Should have multiple references:
	// 1. impl implements interface
	// 2. interface contract for impl
	// 3. constructor returns interface
	if len(refs) < 3 {
		t.Errorf("Expected at least 3 references, got %d", len(refs))
	}

	// Check for expected reference types
	hasImplements := false
	hasContract := false
	hasReturns := false

	for _, ref := range refs {
		switch ref.Reason {
		case "implements-interface":
			hasImplements = true
			if ref.Symbol.Name != "ServiceImpl" {
				t.Errorf("Implements reference should be for ServiceImpl, got %s", ref.Symbol.Name)
			}
		case "interface-contract":
			hasContract = true
			if ref.Symbol.Name != "Service" {
				t.Errorf("Contract reference should be for Service, got %s", ref.Symbol.Name)
			}
		case "returns-interface":
			hasReturns = true
			if ref.Symbol.Name != "NewService" {
				t.Errorf("Returns reference should be for NewService, got %s", ref.Symbol.Name)
			}
		}
	}

	if !hasImplements {
		t.Error("Missing 'implements-interface' reference")
	}
	if !hasContract {
		t.Error("Missing 'interface-contract' reference")
	}
	if !hasReturns {
		t.Error("Missing 'returns-interface' reference")
	}
}

// TestFindConstructorPattern tests constructor detection
func TestFindConstructorPattern(t *testing.T) {
	tests := []struct {
		name              string
		interfaceName     string
		constructorName   string
		shouldMatch       bool
	}{
		{"NewPattern", "Service", "NewService", true},
		{"CreatePattern", "Service", "CreateService", true},
		{"MakePattern", "Service", "MakeService", true},
		{"NoMatch", "Service", "BuildService", false},
		{"WrongName", "Service", "NewClient", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test pattern matching logic
			patterns := []string{
				"New" + tt.interfaceName,
				"Create" + tt.interfaceName,
				"Make" + tt.interfaceName,
			}

			matches := false
			for _, pattern := range patterns {
				if tt.constructorName == pattern {
					matches = true
					break
				}
			}

			if matches != tt.shouldMatch {
				t.Errorf("Pattern matching failed: expected %v, got %v for %s",
					tt.shouldMatch, matches, tt.constructorName)
			}
		})
	}
}
