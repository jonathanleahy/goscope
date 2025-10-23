package extract

import (
	"go/ast"
	"go/token"
	gotypes "go/types"

	"github.com/extract-scope-go/go-scope/internal/types"
	"golang.org/x/tools/go/packages"
)

// InterfaceAnalyzer finds interface-to-implementation mappings
type InterfaceAnalyzer struct {
	pkgs []*packages.Package
	fset *token.FileSet
}

// NewInterfaceAnalyzer creates a new interface analyzer
func NewInterfaceAnalyzer(pkgs []*packages.Package, fset *token.FileSet) *InterfaceAnalyzer {
	return &InterfaceAnalyzer{
		pkgs: pkgs,
		fset: fset,
	}
}

// AnalyzeInterfaces finds all interface-implementation relationships
func (ia *InterfaceAnalyzer) AnalyzeInterfaces(symbols []types.Symbol) []types.InterfaceMapping {
	var mappings []types.InterfaceMapping

	// Find all interfaces in the symbols
	interfaces := ia.findInterfaces(symbols)

	// Also discover interfaces implemented by extracted structs
	discoveredInterfaces := ia.discoverInterfacesFromStructs(symbols)
	interfaces = append(interfaces, discoveredInterfaces...)

	// Deduplicate interfaces
	seenInterfaces := make(map[string]bool)
	uniqueInterfaces := []types.Symbol{}
	for _, iface := range interfaces {
		key := iface.Package + "." + iface.Name
		if !seenInterfaces[key] {
			seenInterfaces[key] = true
			uniqueInterfaces = append(uniqueInterfaces, iface)
		}
	}

	for _, iface := range uniqueInterfaces {
		mapping := types.InterfaceMapping{
			Interface:       iface,
			Implementations: []types.Symbol{},
		}

		// Find implementations
		impls := ia.findImplementations(iface, symbols)
		mapping.Implementations = impls

		// Find constructor
		constructor := ia.findConstructor(iface, symbols)
		if constructor != nil {
			mapping.Constructor = constructor
			mapping.DIFramework = ia.detectConstructorFramework(constructor)
		}

		if len(impls) > 0 || constructor != nil {
			mappings = append(mappings, mapping)
		}
	}

	return mappings
}

// findInterfaces extracts all interface symbols
func (ia *InterfaceAnalyzer) findInterfaces(symbols []types.Symbol) []types.Symbol {
	var interfaces []types.Symbol
	for _, sym := range symbols {
		if sym.Kind == "interface" {
			interfaces = append(interfaces, sym)
		}
	}
	return interfaces
}

// discoverInterfacesFromStructs finds interfaces implemented by structs
func (ia *InterfaceAnalyzer) discoverInterfacesFromStructs(symbols []types.Symbol) []types.Symbol {
	var discovered []types.Symbol

	for _, sym := range symbols {
		if sym.Kind != "struct" {
			continue
		}

		// Find the struct type object
		structObj := ia.findObjectBySymbol(sym)
		if structObj == nil {
			continue
		}

		structType := structObj.Type()

		// Get the Go types package for this struct
		goPkg := structObj.Pkg()
		if goPkg == nil {
			continue
		}

		// Find the packages.Package for this Go types package
		var pkgForAnalysis *packages.Package
		for _, p := range ia.pkgs {
			if p.Types != nil && p.Types.Path() == goPkg.Path() {
				pkgForAnalysis = p
				break
			}
		}

		if pkgForAnalysis == nil {
			continue
		}

		// Look through all definitions in the package scope
		scope := goPkg.Scope()
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)

			// Check if it's an interface
			if typeName, ok := obj.(*gotypes.TypeName); ok {
				if ifaceType, ok := typeName.Type().Underlying().(*gotypes.Interface); ok {
					// Check if struct implements this interface
					if gotypes.Implements(structType, ifaceType) ||
					   gotypes.Implements(gotypes.NewPointer(structType), ifaceType) {

						// Extract the interface symbol
						ifaceSym := ia.extractInterfaceSymbol(goPkg, typeName, ifaceType, pkgForAnalysis)
						if ifaceSym != nil {
							discovered = append(discovered, *ifaceSym)
						}
					}
				}
			}
		}
	}

	return discovered
}

// extractInterfaceSymbol creates a Symbol for an interface
func (ia *InterfaceAnalyzer) extractInterfaceSymbol(pkg *gotypes.Package, typeName *gotypes.TypeName, ifaceType *gotypes.Interface, pkgForAnalysis *packages.Package) *types.Symbol {
	// Find the AST node for this interface
	pos := ia.fset.Position(typeName.Pos())

	// Use the provided package for analysis
	for _, astFile := range pkgForAnalysis.Syntax {
		var foundSpec *ast.TypeSpec
		var foundDoc *ast.CommentGroup

		ast.Inspect(astFile, func(n ast.Node) bool {
			if genDecl, ok := n.(*ast.GenDecl); ok {
				for _, spec := range genDecl.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						if typeSpec.Name.Name == typeName.Name() {
							if _, isIface := typeSpec.Type.(*ast.InterfaceType); isIface {
								foundSpec = typeSpec
								foundDoc = genDecl.Doc
								return false
							}
						}
					}
				}
			}
			return true
		})

		if foundSpec != nil {
			// Extract code from actual source
			code := ia.extractCode(foundSpec.Pos(), foundSpec.End())

			doc := ""
			if foundDoc != nil {
				doc = foundDoc.Text()
			}

			return &types.Symbol{
				Package:  pkg.Path(),
				Name:     typeName.Name(),
				Kind:     "interface",
				Exported: typeName.Exported(),
				File:     pos.Filename,
				Line:     pos.Line,
				EndLine:  ia.fset.Position(foundSpec.End()).Line,
				Code:     code,
				Doc:      doc,
			}
		}
	}

	return nil
}

// extractCode extracts source code from position range
func (ia *InterfaceAnalyzer) extractCode(start, end token.Pos) string {
	tf := ia.fset.File(start)
	if tf == nil {
		return ""
	}

	_ = tf.Offset(start) // startOffset
	_ = tf.Offset(end)   // endOffset

	// Placeholder for now - in a full implementation, you'd read from the actual file
	return "(interface definition)"

	// Proper implementation would be:
	// import "io/ioutil"
	// content, err := ioutil.ReadFile(tf.Name())
	// if err == nil {
	//     return string(content[startOffset:endOffset])
	// }
}

// findImplementations finds all structs that implement an interface
func (ia *InterfaceAnalyzer) findImplementations(iface types.Symbol, symbols []types.Symbol) []types.Symbol {
	var implementations []types.Symbol

	// Get the interface type
	ifaceObj := ia.findObjectBySymbol(iface)
	if ifaceObj == nil {
		return implementations
	}

	ifaceType, ok := ifaceObj.Type().Underlying().(*gotypes.Interface)
	if !ok {
		return implementations
	}

	// Check each struct to see if it implements the interface
	for _, sym := range symbols {
		if sym.Kind == "struct" {
			if ia.implementsInterface(sym, ifaceType) {
				implementations = append(implementations, sym)
			}
		}
	}

	return implementations
}

// implementsInterface checks if a struct implements an interface
func (ia *InterfaceAnalyzer) implementsInterface(structSym types.Symbol, iface *gotypes.Interface) bool {
	structObj := ia.findObjectBySymbol(structSym)
	if structObj == nil {
		return false
	}

	structType := structObj.Type()

	// Check if the struct type implements the interface
	// We need to check both T and *T
	return gotypes.Implements(structType, iface) ||
		gotypes.Implements(gotypes.NewPointer(structType), iface)
}

// findConstructor finds a constructor function that returns the interface
func (ia *InterfaceAnalyzer) findConstructor(iface types.Symbol, symbols []types.Symbol) *types.Symbol {
	// Common constructor patterns: New<InterfaceName>, New<Name>Service, etc.
	constructorPatterns := []string{
		"New" + iface.Name,
		"Create" + iface.Name,
		"Make" + iface.Name,
	}

	for _, sym := range symbols {
		if sym.Kind != "func" {
			continue
		}

		// Check if name matches pattern
		matches := false
		for _, pattern := range constructorPatterns {
			if sym.Name == pattern {
				matches = true
				break
			}
		}

		if !matches {
			continue
		}

		// Check if function returns the interface type
		funcObj := ia.findObjectBySymbol(sym)
		if funcObj == nil {
			continue
		}

		funcType, ok := funcObj.Type().(*gotypes.Signature)
		if !ok {
			continue
		}

		// Check return type
		results := funcType.Results()
		if results.Len() == 0 {
			continue
		}

		// Get first return type
		firstReturn := results.At(0).Type()

		// Check if it's a named type matching the interface
		if named, ok := firstReturn.(*gotypes.Named); ok {
			if named.Obj().Name() == iface.Name && named.Obj().Pkg().Path() == iface.Package {
				// Found constructor!
				symCopy := sym
				symCopy.InterfaceType = iface.Name

				// Try to find what concrete type it instantiates
				impl := ia.findConstructorImplementation(&symCopy)
				if impl != "" {
					symCopy.Implementation = impl
				}

				return &symCopy
			}
		}
	}

	return nil
}

// findConstructorImplementation analyzes constructor body to find concrete type
func (ia *InterfaceAnalyzer) findConstructorImplementation(constructor *types.Symbol) string {
	// Find the function AST node
	for _, pkg := range ia.pkgs {
		if pkg.PkgPath != constructor.Package {
			continue
		}

		for _, astFile := range pkg.Syntax {
			ast.Inspect(astFile, func(n ast.Node) bool {
				fn, ok := n.(*ast.FuncDecl)
				if !ok || fn.Name.Name != constructor.Name {
					return true
				}

				// Look for return statements
				ast.Inspect(fn.Body, func(ret ast.Node) bool {
					retStmt, ok := ret.(*ast.ReturnStmt)
					if !ok {
						return true
					}

					for _, expr := range retStmt.Results {
						// Look for &TypeName{...} pattern
						if unary, ok := expr.(*ast.UnaryExpr); ok && unary.Op == token.AND {
							if comp, ok := unary.X.(*ast.CompositeLit); ok {
								if ident, ok := comp.Type.(*ast.Ident); ok {
									constructor.Implementation = ident.Name
									return false
								}
								if sel, ok := comp.Type.(*ast.SelectorExpr); ok {
									constructor.Implementation = sel.Sel.Name
									return false
								}
							}
						}

						// Look for TypeName{...} pattern
						if comp, ok := expr.(*ast.CompositeLit); ok {
							if ident, ok := comp.Type.(*ast.Ident); ok {
								constructor.Implementation = ident.Name
								return false
							}
						}
					}
					return true
				})

				return false
			})
		}
	}

	return constructor.Implementation
}

// detectConstructorFramework detects if constructor uses a DI framework
func (ia *InterfaceAnalyzer) detectConstructorFramework(constructor *types.Symbol) string {
	// Check for Wire annotations in the file
	for _, pkg := range ia.pkgs {
		if pkg.PkgPath != constructor.Package {
			continue
		}

		for _, astFile := range pkg.Syntax {
			// Check for //go:build wireinject
			for _, comment := range astFile.Comments {
				for _, c := range comment.List {
					if c.Text == "//go:build wireinject" || c.Text == "// +build wireinject" {
						return "wire"
					}
				}
			}

			// Check for fx.Provide calls
			ast.Inspect(astFile, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					if ident, ok := sel.X.(*ast.Ident); ok {
						if ident.Name == "fx" && sel.Sel.Name == "Provide" {
							return false // Found fx
						}
					}
				}
				return true
			})
		}
	}

	return "manual"
}

// findObjectBySymbol finds the gotypes.Object for a Symbol
func (ia *InterfaceAnalyzer) findObjectBySymbol(sym types.Symbol) gotypes.Object {
	for _, pkg := range ia.pkgs {
		if pkg.PkgPath != sym.Package {
			continue
		}

		for _, astFile := range pkg.Syntax {
			var foundObj gotypes.Object

			ast.Inspect(astFile, func(n ast.Node) bool {
				if foundObj != nil {
					return false
				}

				switch node := n.(type) {
				case *ast.TypeSpec:
					if node.Name.Name == sym.Name {
						foundObj = pkg.TypesInfo.Defs[node.Name]
						return false
					}
				case *ast.FuncDecl:
					if node.Name.Name == sym.Name {
						foundObj = pkg.TypesInfo.Defs[node.Name]
						return false
					}
				}
				return true
			})

			if foundObj != nil {
				return foundObj
			}
		}
	}

	return nil
}

// ExtractInterfaceReferences creates Reference entries for interface relationships
func ExtractInterfaceReferences(mappings []types.InterfaceMapping, depth int) []types.Reference {
	var refs []types.Reference

	for _, mapping := range mappings {
		// Add references for each implementation
		for _, impl := range mapping.Implementations {
			ref := types.Reference{
				Symbol:       impl,
				Reason:       "implements-interface",
				Depth:        depth,
				External:     false,
				ReferencedBy: mapping.Interface.Name,
			}
			refs = append(refs, ref)

			// Add reverse reference from implementation to interface
			ifaceRef := types.Reference{
				Symbol:       mapping.Interface,
				Reason:       "interface-contract",
				Depth:        depth,
				External:     false,
				ReferencedBy: impl.Name,
			}
			refs = append(refs, ifaceRef)
		}

		// Add reference for constructor if present
		if mapping.Constructor != nil {
			constructorRef := types.Reference{
				Symbol:       *mapping.Constructor,
				Reason:       "returns-interface",
				Depth:        depth,
				External:     false,
				ReferencedBy: mapping.Interface.Name,
			}
			refs = append(refs, constructorRef)
		}
	}

	return refs
}
