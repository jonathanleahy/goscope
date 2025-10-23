package extract

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	gotypes "go/types"

	"github.com/extract-scope-go/go-scope/internal/extract/di"
	"github.com/extract-scope-go/go-scope/internal/types"
	"golang.org/x/tools/go/packages"
)

// Collector gathers dependencies for a target symbol using depth-limited BFS
type Collector struct {
	pkgs     []*packages.Package
	fset     *token.FileSet
	visited  visitedSet
	maxDepth int
}

// NewCollector creates a new dependency collector
func NewCollector(pkgs []*packages.Package, fset *token.FileSet, maxDepth int) *Collector {
	return &Collector{
		pkgs:     pkgs,
		fset:     fset,
		visited:  make(visitedSet),
		maxDepth: maxDepth,
	}
}

// Collect gathers dependencies starting from target symbol
func (c *Collector) Collect(target *types.Symbol) ([]types.Reference, []string, error) {
	if c.maxDepth == 0 {
		// Depth 0: return only target, no dependencies
		return []types.Reference{}, []string{}, nil
	}

	// Find the target's AST node and package
	targetPkg, targetNode, err := c.findSymbolNode(target)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find target symbol: %w", err)
	}

	// BFS traversal
	queue := []objectInfo{{
		obj:   c.getObjectFromNode(targetPkg, targetNode),
		depth: 0,
	}}

	var references []types.Reference
	var external []string
	externalSet := make(map[string]bool)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.obj == nil {
			continue
		}

		// Check if already visited
		key := c.makeKey(current.obj)
		if c.visited[key] {
			continue
		}
		c.visited[key] = true

		// Don't go deeper than maxDepth
		if current.depth >= c.maxDepth {
			continue
		}

		// Find references in this object's code
		refs, exts := c.findReferences(current.obj, current.depth+1)

		// Add to results
		for _, ref := range refs {
			references = append(references, ref)

			// Add referenced objects to queue if not external
			if !ref.External && ref.Symbol.Name != "" {
				// Try to find and queue this reference
				refPkg, refNode, err := c.findSymbolByName(ref.Symbol.Package, ref.Symbol.Name)
				if err == nil && refNode != nil {
					refObj := c.getObjectFromNode(refPkg, refNode)
					if refObj != nil {
						queue = append(queue, objectInfo{
							obj:   refObj,
							depth: ref.Depth,
						})
					}
				}
			}
		}

		// Collect external references
		for _, ext := range exts {
			if !externalSet[ext] {
				external = append(external, ext)
				externalSet[ext] = true
			}
		}
	}

	return references, external, nil
}

// findSymbolNode finds the AST node for a symbol
func (c *Collector) findSymbolNode(sym *types.Symbol) (*packages.Package, ast.Node, error) {
	// Find the package
	var pkg *packages.Package
	for _, p := range c.pkgs {
		if p.PkgPath == sym.Package {
			pkg = p
			break
		}
	}

	if pkg == nil {
		return nil, nil, fmt.Errorf("package not found: %s", sym.Package)
	}

	// Find the file
	var astFile *ast.File
	for i, file := range pkg.CompiledGoFiles {
		if file == sym.File {
			if i < len(pkg.Syntax) {
				astFile = pkg.Syntax[i]
				break
			}
		}
	}

	if astFile == nil {
		return nil, nil, fmt.Errorf("file not found: %s", sym.File)
	}

	// Find the node by name and line (simpler approach)
	var foundNode ast.Node

	ast.Inspect(astFile, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		// Check for function declarations
		if fn, ok := n.(*ast.FuncDecl); ok {
			if fn.Name.Name == sym.Name && c.fset.Position(fn.Pos()).Line == sym.Line {
				foundNode = fn
				return false
			}
		}

		// Check for type specs
		if ts, ok := n.(*ast.TypeSpec); ok {
			if ts.Name.Name == sym.Name {
				foundNode = ts
				return false
			}
		}

		return true
	})

	if foundNode == nil {
		return nil, nil, fmt.Errorf("node not found for symbol: %s", sym.Name)
	}

	return pkg, foundNode, nil
}

// findSymbolByName finds a symbol by package and name
func (c *Collector) findSymbolByName(pkgPath, name string) (*packages.Package, ast.Node, error) {
	// Find the package
	var pkg *packages.Package
	for _, p := range c.pkgs {
		if p.PkgPath == pkgPath {
			pkg = p
			break
		}
	}

	if pkg == nil {
		return nil, nil, fmt.Errorf("package not found: %s", pkgPath)
	}

	// Search all files in package
	for _, astFile := range pkg.Syntax {
		var foundNode ast.Node
		ast.Inspect(astFile, func(n ast.Node) bool {
			if n == nil {
				return false
			}

			// Check for function declarations
			if fn, ok := n.(*ast.FuncDecl); ok {
				if fn.Name.Name == name {
					foundNode = fn
					return false
				}
			}

			// Check for type specs
			if ts, ok := n.(*ast.TypeSpec); ok {
				if ts.Name.Name == name {
					foundNode = ts
					return false
				}
			}

			return true
		})

		if foundNode != nil {
			return pkg, foundNode, nil
		}
	}

	return nil, nil, fmt.Errorf("symbol not found: %s.%s", pkgPath, name)
}

// getObjectFromNode gets gotypes.Object from AST node
func (c *Collector) getObjectFromNode(pkg *packages.Package, node ast.Node) gotypes.Object {
	switch n := node.(type) {
	case *ast.FuncDecl:
		if n.Name != nil {
			return pkg.TypesInfo.Defs[n.Name]
		}
	case *ast.TypeSpec:
		if n.Name != nil {
			return pkg.TypesInfo.Defs[n.Name]
		}
	case *ast.ValueSpec:
		if len(n.Names) > 0 {
			return pkg.TypesInfo.Defs[n.Names[0]]
		}
	}
	return nil
}

// findReferences finds all references in an object's code
func (c *Collector) findReferences(obj gotypes.Object, depth int) ([]types.Reference, []string) {
	if obj == nil {
		return nil, nil
	}

	// Find the AST node for this object
	var node ast.Node
	var pkg *packages.Package

	// Search all packages for this object
	for _, p := range c.pkgs {
		for _, file := range p.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				if n == nil {
					return false
				}

				// Check if this node defines our object
				switch n := n.(type) {
				case *ast.FuncDecl:
					if n.Name != nil && p.TypesInfo.Defs[n.Name] == obj {
						node = n
						pkg = p
						return false
					}
				case *ast.TypeSpec:
					if n.Name != nil && p.TypesInfo.Defs[n.Name] == obj {
						node = n
						pkg = p
						return false
					}
				}
				return true
			})

			if node != nil {
				break
			}
		}
		if node != nil {
			break
		}
	}

	if node == nil || pkg == nil {
		return nil, nil
	}

	// Walk the node and find all identifiers
	var references []types.Reference
	var external []string
	externalSet := make(map[string]bool)

	ast.Inspect(node, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		// Check for identifiers (function calls, variable uses)
		if ident, ok := n.(*ast.Ident); ok {
			if usedObj := pkg.TypesInfo.Uses[ident]; usedObj != nil {
				ref, ext := c.makeReference(usedObj, depth, obj.Name())
				if ref != nil {
					// Check if already in references
					found := false
					for _, existing := range references {
						if existing.Symbol.Package == ref.Symbol.Package &&
							existing.Symbol.Name == ref.Symbol.Name {
							found = true
							break
						}
					}
					if !found {
						references = append(references, *ref)
					}
				}
				if ext != "" && !externalSet[ext] {
					external = append(external, ext)
					externalSet[ext] = true
				}
			}
		}

		// Check for selector expressions (pkg.Function calls)
		if sel, ok := n.(*ast.SelectorExpr); ok {
			if selObj := pkg.TypesInfo.Uses[sel.Sel]; selObj != nil {
				ref, ext := c.makeReference(selObj, depth, obj.Name())
				if ref != nil {
					found := false
					for _, existing := range references {
						if existing.Symbol.Package == ref.Symbol.Package &&
							existing.Symbol.Name == ref.Symbol.Name {
							found = true
							break
						}
					}
					if !found {
						references = append(references, *ref)
					}
				}
				if ext != "" && !externalSet[ext] {
					external = append(external, ext)
					externalSet[ext] = true
				}
			}
		}

		return true
	})

	return references, external
}

// makeReference creates a Reference from gotypes.Object
func (c *Collector) makeReference(obj gotypes.Object, depth int, referencedBy string) (*types.Reference, string) {
	if obj == nil {
		return nil, ""
	}

	// Get package path
	pkgPath := ""
	if obj.Pkg() != nil {
		pkgPath = obj.Pkg().Path()
	}

	// Check if external (different module or stdlib)
	isExternal := c.isExternal(pkgPath)

	// Don't include builtin types/functions
	if pkgPath == "" {
		return nil, ""
	}

	// Create symbol
	sym := types.Symbol{
		Package:  pkgPath,
		Name:     obj.Name(),
		Exported: obj.Exported(),
	}

	// Determine kind
	switch obj.(type) {
	case *gotypes.Func:
		sym.Kind = "func"
	case *gotypes.TypeName:
		sym.Kind = "type"
	case *gotypes.Var:
		sym.Kind = "var"
	case *gotypes.Const:
		sym.Kind = "const"
	default:
		sym.Kind = "unknown"
	}

	// If not external, try to get full info
	if !isExternal {
		pos := c.fset.Position(obj.Pos())
		sym.File = pos.Filename
		sym.Line = pos.Line
		sym.Column = pos.Column

		// Try to get code
		if fullSym, err := c.getFullSymbol(obj); err == nil && fullSym != nil {
			sym = *fullSym
		}
	}

	ref := &types.Reference{
		Symbol:       sym,
		Depth:        depth,
		External:     isExternal,
		Stub:         isExternal,
		ReferencedBy: referencedBy,
		Reason:       "direct-call", // Simplified for now
	}

	// Create external reference string
	var externalRef string
	if isExternal {
		externalRef = fmt.Sprintf("%s.%s", pkgPath, obj.Name())
	}

	return ref, externalRef
}

// isExternal checks if a package is external to the current module
func (c *Collector) isExternal(pkgPath string) bool {
	if pkgPath == "" {
		return true // Builtin
	}

	// Check if package is in our loaded packages
	for _, pkg := range c.pkgs {
		if pkg.PkgPath == pkgPath {
			return false // Found in our module
		}
	}

	return true // Not in our module = external
}

// getFullSymbol gets complete symbol information
func (c *Collector) getFullSymbol(obj gotypes.Object) (*types.Symbol, error) {
	pos := c.fset.Position(obj.Pos())

	// Find the package containing this object
	var pkg *packages.Package
	for _, p := range c.pkgs {
		if obj.Pkg() != nil && p.PkgPath == obj.Pkg().Path() {
			pkg = p
			break
		}
	}

	if pkg == nil {
		return nil, fmt.Errorf("package not found for object")
	}

	// Use locator to get full details
	locator := &Locator{
		fset: c.fset,
		pkgs: c.pkgs,
	}

	// Find the file
	var astFile *ast.File
	for i, file := range pkg.CompiledGoFiles {
		if file == pos.Filename {
			if i < len(pkg.Syntax) {
				astFile = pkg.Syntax[i]
				break
			}
		}
	}

	if astFile == nil {
		return nil, fmt.Errorf("file not found")
	}

	// Find and extract the symbol
	tf := c.fset.File(astFile.Pos())
	if tf == nil {
		return nil, fmt.Errorf("file not found in fileset")
	}
	filePos := tf.LineStart(pos.Line)
	if pos.Column > 1 {
		filePos += token.Pos(pos.Column - 1)
	}
	path, _ := locator.findEnclosingNode(astFile, filePos)

	if len(path) == 0 {
		return nil, fmt.Errorf("no path found")
	}

	symbol, err := locator.extractSymbol(pkg, astFile, path, filePos)
	if err != nil {
		return nil, err
	}

	return symbol, nil
}

// makeKey creates a unique key for an object
func (c *Collector) makeKey(obj gotypes.Object) symbolKey {
	pkgPath := ""
	if obj.Pkg() != nil {
		pkgPath = obj.Pkg().Path()
	}
	return symbolKey{
		pkg:  pkgPath,
		name: obj.Name(),
		pos:  obj.Pos(),
	}
}

// ExtractSymbol is the main entry point for symbol extraction
func ExtractSymbol(ctx context.Context, target types.Target, opts types.Options) (*types.Result, error) {
	// Set defaults
	if opts.Depth < 0 {
		opts.Depth = 1
	}
	if opts.Format == "" {
		opts.Format = "markdown"
	}

	// Step 1: Locate the target symbol
	locator := NewLocator()
	symbol, err := locator.LocateSymbol(target.Root, target.File, target.Line, target.Column)
	if err != nil {
		return nil, fmt.Errorf("failed to locate symbol: %w", err)
	}

	// Step 2: Collect dependencies
	collector := NewCollector(locator.pkgs, locator.fset, opts.Depth)
	references, external, err := collector.Collect(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to collect dependencies: %w", err)
	}

	// Step 3: Analyze interfaces and implementations
	allSymbols := []types.Symbol{*symbol}
	for _, ref := range references {
		allSymbols = append(allSymbols, ref.Symbol)
	}

	interfaceAnalyzer := NewInterfaceAnalyzer(locator.pkgs, locator.fset)
	interfaceMappings := interfaceAnalyzer.AnalyzeInterfaces(allSymbols)

	// Add interface relationship references
	interfaceRefs := ExtractInterfaceReferences(interfaceMappings, opts.Depth+1)
	references = append(references, interfaceRefs...)

	// Step 4: Detect DI framework and analyze bindings
	diDetector := di.NewDetector(locator.pkgs, locator.fset)
	detectedFramework := diDetector.DetectFramework()
	diBindings := diDetector.AnalyzeDIBindings(allSymbols)

	// Step 5: Build extract
	extract := types.Extract{
		Target:              *symbol,
		References:          references,
		External:            external,
		InterfaceMappings:   interfaceMappings,
		DIBindings:          diBindings,
		DetectedDIFramework: detectedFramework,
	}

	// Step 4: Format output (will be done by API layer to avoid circular imports)
	rendered := "" // Empty for now, API layer will format

	// Step 5: Build result
	result := &types.Result{
		Extract:  extract,
		Rendered: rendered,
		Metadata: types.Metadata{
			Options:      opts,
			TotalSymbols: len(references) + 1, // +1 for target
		},
	}

	return result, nil
}
