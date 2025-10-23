package extract

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	gotypes "go/types"
	"os"
	"path/filepath"

	"github.com/extract-scope-go/go-scope/internal/types"
	"golang.org/x/tools/go/packages"
)

// Locator finds symbols at specific positions in Go source files
type Locator struct {
	fset *token.FileSet
	pkgs []*packages.Package
}

// NewLocator creates a new Locator instance
func NewLocator() *Locator {
	return &Locator{
		fset: token.NewFileSet(),
	}
}

// LocateSymbol finds the symbol at the specified file and line
func (l *Locator) LocateSymbol(root, file string, line, col int) (*types.Symbol, error) {
	// Load packages containing the file
	if err := l.loadPackages(root, file); err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}

	// Find the package containing this file
	pkg, astFile := l.findFileInPackages(file)
	if pkg == nil || astFile == nil {
		return nil, fmt.Errorf("file not found in loaded packages: %s", file)
	}

	// Find the position in the file
	pos := l.findPosition(astFile, line, col)
	if !pos.IsValid() {
		return nil, fmt.Errorf("invalid position: line %d, column %d", line, col)
	}

	// Find the enclosing node at this position
	path, _ := l.findEnclosingNode(astFile, pos)
	if len(path) == 0 {
		return nil, fmt.Errorf("no symbol found at %s:%d:%d", file, line, col)
	}

	// Extract symbol information
	symbol, err := l.extractSymbol(pkg, astFile, path, pos)
	if err != nil {
		return nil, fmt.Errorf("failed to extract symbol: %w", err)
	}

	return symbol, nil
}

// loadPackages loads all packages in the module
func (l *Locator) loadPackages(root, file string) error {
	// Make file path absolute if it's relative
	absFile := file
	if !filepath.IsAbs(file) {
		absFile = filepath.Join(root, file)
	}

	// Check if file exists
	if _, err := os.Stat(absFile); err != nil {
		return fmt.Errorf("file does not exist: %s", absFile)
	}

	// Configure package loading
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedModule,
		Dir:   root,
		Fset:  l.fset,
		Tests: false,
	}

	// Load packages
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return fmt.Errorf("package load error: %w", err)
	}

	// Check for errors in loaded packages
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			// Note: We still proceed even with some errors
			// as they might not affect our target file
		}
	}

	l.pkgs = pkgs
	return nil
}

// findFileInPackages finds which package contains the given file
func (l *Locator) findFileInPackages(file string) (*packages.Package, *ast.File) {
	absFile, _ := filepath.Abs(file)

	for _, pkg := range l.pkgs {
		for i, pkgFile := range pkg.CompiledGoFiles {
			absPkgFile, _ := filepath.Abs(pkgFile)
			if absPkgFile == absFile {
				if i < len(pkg.Syntax) {
					return pkg, pkg.Syntax[i]
				}
			}
		}
	}

	return nil, nil
}

// findPosition converts line/column to token.Pos
func (l *Locator) findPosition(file *ast.File, line, col int) token.Pos {
	tf := l.fset.File(file.Pos())
	if tf == nil {
		return token.NoPos
	}

	// Convert line to position
	if line < 1 || line > tf.LineCount() {
		return token.NoPos
	}

	linePos := tf.LineStart(line)
	if col > 1 {
		linePos += token.Pos(col - 1)
	}

	return linePos
}

// findEnclosingNode finds the smallest node enclosing the position
// Returns the path from root to the innermost node, and the exact node
func (l *Locator) findEnclosingNode(file *ast.File, pos token.Pos) ([]ast.Node, ast.Node) {
	path := []ast.Node{}
	found := false

	// Walk the AST and collect all nodes that contain the position
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		// Check if position is within this node
		if n.Pos() <= pos && pos <= n.End() {
			path = append(path, n)
			found = true
			return true // Continue to children
		}

		// If position is before this node, no point continuing this branch
		if pos < n.Pos() {
			return false
		}

		// If position is after this node, continue to siblings
		return pos >= n.Pos()
	})

	if !found || len(path) == 0 {
		return nil, nil
	}

	// Return the innermost node as exact
	exact := path[len(path)-1]
	return path, exact
}

// extractSymbol extracts Symbol information from AST path
func (l *Locator) extractSymbol(pkg *packages.Package, file *ast.File, path []ast.Node, pos token.Pos) (*types.Symbol, error) {
	// Look for function, type, var, const declarations
	for i := len(path) - 1; i >= 0; i-- {
		node := path[i]

		switch n := node.(type) {
		case *ast.FuncDecl:
			return l.extractFromFuncDecl(pkg, file, n)

		case *ast.GenDecl:
			// Handle type, var, const declarations
			for _, spec := range n.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if l.containsPos(s, pos) {
						return l.extractFromTypeSpec(pkg, file, s, n.Doc)
					}
				case *ast.ValueSpec:
					if l.containsPos(s, pos) {
						return l.extractFromValueSpec(pkg, file, s, n.Doc, n.Tok)
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("no extractable symbol found at position")
}

// containsPos checks if a node contains the position
func (l *Locator) containsPos(node ast.Node, pos token.Pos) bool {
	return node.Pos() <= pos && pos <= node.End()
}

// extractFromFuncDecl extracts symbol from function declaration
func (l *Locator) extractFromFuncDecl(pkg *packages.Package, file *ast.File, decl *ast.FuncDecl) (*types.Symbol, error) {
	symbol := &types.Symbol{
		Package:  pkg.PkgPath,
		Name:     decl.Name.Name,
		Kind:     "func",
		Exported: ast.IsExported(decl.Name.Name),
		File:     l.fset.Position(decl.Pos()).Filename,
		Line:     l.fset.Position(decl.Pos()).Line,
		EndLine:  l.fset.Position(decl.End()).Line,
		Column:   l.fset.Position(decl.Pos()).Column,
	}

	// Check if this is a method (has receiver)
	if decl.Recv != nil && len(decl.Recv.List) > 0 {
		symbol.Kind = "method"
		if recvType := decl.Recv.List[0].Type; recvType != nil {
			symbol.Receiver = l.exprToString(recvType)
		}
	}

	// Extract code
	symbol.Code = l.extractCode(decl.Pos(), decl.End())

	// Extract documentation
	if decl.Doc != nil {
		symbol.Doc = decl.Doc.Text()
	}

	return symbol, nil
}

// extractFromTypeSpec extracts symbol from type declaration
func (l *Locator) extractFromTypeSpec(pkg *packages.Package, file *ast.File, spec *ast.TypeSpec, doc *ast.CommentGroup) (*types.Symbol, error) {
	kind := "type"

	// Determine more specific type kind
	switch spec.Type.(type) {
	case *ast.InterfaceType:
		kind = "interface"
	case *ast.StructType:
		kind = "struct"
	}

	symbol := &types.Symbol{
		Package:  pkg.PkgPath,
		Name:     spec.Name.Name,
		Kind:     kind,
		Exported: ast.IsExported(spec.Name.Name),
		File:     l.fset.Position(spec.Pos()).Filename,
		Line:     l.fset.Position(spec.Pos()).Line,
		EndLine:  l.fset.Position(spec.End()).Line,
		Column:   l.fset.Position(spec.Pos()).Column,
	}

	// Extract code (including type keyword)
	// Need to get parent GenDecl for full code
	symbol.Code = l.extractCode(spec.Pos(), spec.End())

	// Extract documentation
	if doc != nil {
		symbol.Doc = doc.Text()
	}

	return symbol, nil
}

// extractFromValueSpec extracts symbol from var/const declaration
func (l *Locator) extractFromValueSpec(pkg *packages.Package, file *ast.File, spec *ast.ValueSpec, doc *ast.CommentGroup, tok token.Token) (*types.Symbol, error) {
	if len(spec.Names) == 0 {
		return nil, fmt.Errorf("value spec has no names")
	}

	// Use first name if multiple
	name := spec.Names[0].Name

	kind := "var"
	if tok == token.CONST {
		kind = "const"
	}

	symbol := &types.Symbol{
		Package:  pkg.PkgPath,
		Name:     name,
		Kind:     kind,
		Exported: ast.IsExported(name),
		File:     l.fset.Position(spec.Pos()).Filename,
		Line:     l.fset.Position(spec.Pos()).Line,
		EndLine:  l.fset.Position(spec.End()).Line,
		Column:   l.fset.Position(spec.Pos()).Column,
	}

	// Extract code
	symbol.Code = l.extractCode(spec.Pos(), spec.End())

	// Extract documentation
	if doc != nil {
		symbol.Doc = doc.Text()
	}

	return symbol, nil
}

// extractCode extracts source code between two positions
func (l *Locator) extractCode(start, end token.Pos) string {
	startPos := l.fset.Position(start)

	// Read the file
	content, err := os.ReadFile(startPos.Filename)
	if err != nil {
		return ""
	}

	// Parse the file to get proper offsets
	tf := l.fset.File(start)
	if tf == nil {
		return ""
	}

	startOffset := tf.Offset(start)
	endOffset := tf.Offset(end)

	if startOffset < 0 || endOffset > len(content) || startOffset > endOffset {
		return ""
	}

	return string(content[startOffset:endOffset])
}

// exprToString converts an expression to string (for receiver types)
func (l *Locator) exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return "*" + l.exprToString(e.X)
	case *ast.SelectorExpr:
		return l.exprToString(e.X) + "." + e.Sel.Name
	default:
		return fmt.Sprintf("%T", expr)
	}
}

// Helper: parse a single file without full package loading (for quick tests)
func parseFile(filename string) (*token.FileSet, *ast.File, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}
	return fset, file, nil
}

// Helper: check if object is exported
func isExported(name string) bool {
	return token.IsExported(name)
}

// Helper: get object position
func getObjectPosition(fset *token.FileSet, obj gotypes.Object) token.Position {
	return fset.Position(obj.Pos())
}
