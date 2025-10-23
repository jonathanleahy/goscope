package di

import (
	"go/ast"
	"go/token"
	gotypes "go/types"
	"strings"

	"github.com/extract-scope-go/go-scope/internal/types"
	"golang.org/x/tools/go/packages"
)

// Detector identifies dependency injection patterns
type Detector struct {
	pkgs []*packages.Package
	fset *token.FileSet
}

// NewDetector creates a new DI detector
func NewDetector(pkgs []*packages.Package, fset *token.FileSet) *Detector {
	return &Detector{
		pkgs: pkgs,
		fset: fset,
	}
}

// DetectFramework identifies which DI framework is in use
func (d *Detector) DetectFramework() string {
	if d.hasWire() {
		return "wire"
	}
	if d.hasFx() {
		return "fx"
	}
	if d.hasManualDI() {
		return "manual"
	}
	return "none"
}

// AnalyzeDIBindings extracts all DI bindings
func (d *Detector) AnalyzeDIBindings(symbols []types.Symbol) []types.DIBinding {
	framework := d.DetectFramework()

	switch framework {
	case "wire":
		return d.analyzeWireBindings(symbols)
	case "fx":
		return d.analyzeFxBindings(symbols)
	case "manual":
		return d.analyzeManualBindings(symbols)
	default:
		return []types.DIBinding{}
	}
}

// hasWire checks for Google Wire
func (d *Detector) hasWire() bool {
	for _, pkg := range d.pkgs {
		for _, astFile := range pkg.Syntax {
			// Check for wireinject build tag
			for _, comment := range astFile.Comments {
				for _, c := range comment.List {
					text := c.Text
					if strings.Contains(text, "wireinject") ||
						strings.Contains(text, "+build wireinject") {
						return true
					}
				}
			}

			// Check for wire imports
			for _, imp := range astFile.Imports {
				if imp.Path.Value == `"github.com/google/wire"` {
					return true
				}
			}
		}
	}
	return false
}

// hasFx checks for Uber Fx
func (d *Detector) hasFx() bool {
	for _, pkg := range d.pkgs {
		for _, astFile := range pkg.Syntax {
			for _, imp := range astFile.Imports {
				if imp.Path.Value == `"go.uber.org/fx"` {
					return true
				}
			}
		}
	}
	return false
}

// hasManualDI checks for manual DI patterns (constructors that take dependencies)
func (d *Detector) hasManualDI() bool {
	// If we have constructor functions (New*) that take parameters, it's manual DI
	for _, pkg := range d.pkgs {
		scope := pkg.Types.Scope()
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)
			if fn, ok := obj.(*gotypes.Func); ok {
				if strings.HasPrefix(fn.Name(), "New") {
					sig := fn.Type().(*gotypes.Signature)
					if sig.Params().Len() > 0 {
						return true
					}
				}
			}
		}
	}
	return false
}

// analyzeWireBindings parses Wire configuration
func (d *Detector) analyzeWireBindings(symbols []types.Symbol) []types.DIBinding {
	var bindings []types.DIBinding

	for _, pkg := range d.pkgs {
		for _, astFile := range pkg.Syntax {
			// Look for wire.NewSet calls
			ast.Inspect(astFile, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				// Check if this is wire.NewSet
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				pkgIdent, ok := sel.X.(*ast.Ident)
				if !ok || pkgIdent.Name != "wire" || sel.Sel.Name != "NewSet" {
					return true
				}

				// Extract providers from NewSet arguments
				for _, arg := range call.Args {
					if ident, ok := arg.(*ast.Ident); ok {
						// This is a provider function
						binding := d.createBindingFromProvider(pkg, ident.Name, symbols, "wire")
						if binding != nil {
							bindings = append(bindings, *binding)
						}
					}
				}

				return true
			})
		}
	}

	return bindings
}

// analyzeFxBindings parses Uber Fx configuration
func (d *Detector) analyzeFxBindings(symbols []types.Symbol) []types.DIBinding {
	var bindings []types.DIBinding

	for _, pkg := range d.pkgs {
		for _, astFile := range pkg.Syntax {
			// Look for fx.Provide calls
			ast.Inspect(astFile, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				pkgIdent, ok := sel.X.(*ast.Ident)
				if !ok || pkgIdent.Name != "fx" || sel.Sel.Name != "Provide" {
					return true
				}

				// Extract providers from Provide arguments
				for _, arg := range call.Args {
					if ident, ok := arg.(*ast.Ident); ok {
						binding := d.createBindingFromProvider(pkg, ident.Name, symbols, "fx")
						if binding != nil {
							bindings = append(bindings, *binding)
						}
					}
				}

				return true
			})
		}
	}

	return bindings
}

// analyzeManualBindings infers DI from constructor patterns
func (d *Detector) analyzeManualBindings(symbols []types.Symbol) []types.DIBinding {
	var bindings []types.DIBinding

	for _, sym := range symbols {
		if sym.Kind == "func" && strings.HasPrefix(sym.Name, "New") {
			binding := d.createBindingFromConstructor(sym, symbols, "manual")
			if binding != nil {
				bindings = append(bindings, *binding)
			}
		}
	}

	return bindings
}

// createBindingFromProvider creates a DIBinding from a provider function
func (d *Detector) createBindingFromProvider(pkg *packages.Package, providerName string, symbols []types.Symbol, framework string) *types.DIBinding {
	// Find the provider function in symbols
	var provider *types.Symbol
	for i, sym := range symbols {
		if sym.Name == providerName && sym.Package == pkg.PkgPath {
			provider = &symbols[i]
			break
		}
	}

	if provider == nil {
		return nil
	}

	return d.createBindingFromConstructor(*provider, symbols, framework)
}

// createBindingFromConstructor analyzes a constructor to create a binding
func (d *Detector) createBindingFromConstructor(constructor types.Symbol, symbols []types.Symbol, framework string) *types.DIBinding {
	// Find the function object
	var funcObj *gotypes.Func
	for _, pkg := range d.pkgs {
		if pkg.PkgPath != constructor.Package {
			continue
		}

		obj := pkg.Types.Scope().Lookup(constructor.Name)
		if fn, ok := obj.(*gotypes.Func); ok {
			funcObj = fn
			break
		}
	}

	if funcObj == nil {
		return nil
	}

	sig := funcObj.Type().(*gotypes.Signature)

	binding := &types.DIBinding{
		Provider:     constructor,
		Dependencies: []types.Symbol{},
		Framework:    framework,
		Scope:        "singleton", // Default assumption
	}

	// Extract parameters (dependencies)
	params := sig.Params()
	for i := 0; i < params.Len(); i++ {
		param := params.At(i)
		paramSym := d.findSymbolForType(param.Type(), symbols)
		if paramSym != nil {
			binding.Dependencies = append(binding.Dependencies, *paramSym)
		}
	}

	// Extract return type (product)
	results := sig.Results()
	if results.Len() > 0 {
		returnType := results.At(0).Type()
		productSym := d.findSymbolForType(returnType, symbols)
		if productSym != nil {
			binding.Product = *productSym
		}
	}

	return binding
}

// findSymbolForType finds a Symbol that matches a gotypes.Type
func (d *Detector) findSymbolForType(typ gotypes.Type, symbols []types.Symbol) *types.Symbol {
	// Handle named types
	if named, ok := typ.(*gotypes.Named); ok {
		obj := named.Obj()
		for i, sym := range symbols {
			if sym.Name == obj.Name() && sym.Package == obj.Pkg().Path() {
				return &symbols[i]
			}
		}
	}

	// Handle pointer types
	if ptr, ok := typ.(*gotypes.Pointer); ok {
		return d.findSymbolForType(ptr.Elem(), symbols)
	}

	return nil
}
