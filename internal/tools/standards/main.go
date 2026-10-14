package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxDepth        = 3
	goExtension     = ".go"
	testSuffix      = "_test.go"
	directivePrefix = "//go:"
	exitFailure     = 1
	defaultRoot     = "."
)

var skippedDirs = map[string]bool{".git": true, "testdata": true, "vendor": true}

var forbiddenCalls = map[string]map[string]bool{
	"fmt": {"Print": true, "Printf": true, "Println": true},
	"log": {"Print": true, "Printf": true, "Println": true, "Fatal": true, "Fatalf": true, "Fatalln": true, "Panic": true, "Panicf": true, "Panicln": true},
}

var forbiddenBuiltins = map[string]bool{"print": true, "println": true, "panic": true}

type violation struct {
	pos     token.Position
	message string
}

type checker struct {
	fset       *token.FileSet
	violations []violation
}

func main() {
	roots := os.Args[1:]
	if len(roots) == 0 {
		roots = []string{defaultRoot}
	}
	c := &checker{fset: token.NewFileSet()}
	for _, root := range roots {
		c.walk(strings.TrimSuffix(root, "/..."))
	}
	for _, v := range c.violations {
		fmt.Fprintf(os.Stderr, "%s: %s\n", v.pos, v.message)
	}
	if len(c.violations) > 0 {
		os.Exit(exitFailure)
	}
}

func (c *checker) walk(root string) {
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && skippedDirs[d.Name()] {
			return filepath.SkipDir
		}
		if d.IsDir() || filepath.Ext(path) != goExtension {
			return nil
		}
		return c.checkFile(path)
	})
	if err != nil {
		c.report(token.Position{Filename: root}, err.Error())
	}
}

func (c *checker) checkFile(path string) error {
	file, err := parser.ParseFile(c.fset, path, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	production := !strings.HasSuffix(path, testSuffix)
	c.checkComments(file)
	ast.Inspect(file, func(n ast.Node) bool {
		c.checkNode(n, production)
		return true
	})
	for _, decl := range file.Decls {
		c.checkDepth(decl)
	}
	return nil
}

func (c *checker) report(pos token.Position, message string) {
	c.violations = append(c.violations, violation{pos: pos, message: message})
}

func (c *checker) checkNode(n ast.Node, production bool) {
	if stmt, ok := n.(*ast.IfStmt); ok && stmt.Else != nil {
		c.report(c.fset.Position(stmt.Else.Pos()), "else proibido: use retorno antecipado")
	}
	if _, ok := n.(*ast.SwitchStmt); ok {
		c.report(c.fset.Position(n.Pos()), "switch proibido: use retorno antecipado ou tabela")
	}
	if _, ok := n.(*ast.TypeSwitchStmt); ok {
		c.report(c.fset.Position(n.Pos()), "type switch proibido: use errors.As ou asserção com retorno antecipado")
	}
	if call, ok := n.(*ast.CallExpr); ok && production {
		c.checkCall(call)
	}
}

func (c *checker) checkCall(call *ast.CallExpr) {
	if ident, ok := call.Fun.(*ast.Ident); ok && forbiddenBuiltins[ident.Name] {
		c.report(c.fset.Position(call.Pos()), ident.Name+" proibido em código de biblioteca")
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}
	pkg, ok := selector.X.(*ast.Ident)
	if ok && forbiddenCalls[pkg.Name][selector.Sel.Name] {
		c.report(c.fset.Position(call.Pos()), pkg.Name+"."+selector.Sel.Name+" proibido: sem saída em console")
	}
}

func (c *checker) checkComments(file *ast.File) {
	allowed := allowedComments(file)
	for _, group := range file.Comments {
		if allowed[group] || isDirective(group) {
			continue
		}
		c.report(c.fset.Position(group.Pos()), "comentário proibido: só doc de identificador exportado")
	}
}

func isDirective(group *ast.CommentGroup) bool {
	for _, comment := range group.List {
		if !strings.HasPrefix(comment.Text, directivePrefix) {
			return false
		}
	}
	return true
}

func allowedComments(file *ast.File) map[*ast.CommentGroup]bool {
	allowed := map[*ast.CommentGroup]bool{file.Doc: true}
	ast.Inspect(file, func(n ast.Node) bool {
		for _, group := range exportedDocs(n) {
			allowed[group] = true
		}
		return true
	})
	return allowed
}

func exportedDocs(n ast.Node) []*ast.CommentGroup {
	if decl, ok := n.(*ast.FuncDecl); ok && decl.Name.IsExported() {
		return []*ast.CommentGroup{decl.Doc}
	}
	if decl, ok := n.(*ast.GenDecl); ok && genDeclExported(decl) {
		return []*ast.CommentGroup{decl.Doc}
	}
	if spec, ok := n.(*ast.TypeSpec); ok && spec.Name.IsExported() {
		return []*ast.CommentGroup{spec.Doc}
	}
	if spec, ok := n.(*ast.ValueSpec); ok && anyExported(spec.Names) {
		return []*ast.CommentGroup{spec.Doc}
	}
	if field, ok := n.(*ast.Field); ok && anyExported(field.Names) {
		return []*ast.CommentGroup{field.Doc}
	}
	return nil
}

func genDeclExported(decl *ast.GenDecl) bool {
	for _, spec := range decl.Specs {
		if specExported(spec) {
			return true
		}
	}
	return false
}

func specExported(spec ast.Spec) bool {
	if typeSpec, ok := spec.(*ast.TypeSpec); ok {
		return typeSpec.Name.IsExported()
	}
	valueSpec, ok := spec.(*ast.ValueSpec)
	return ok && anyExported(valueSpec.Names)
}

func anyExported(names []*ast.Ident) bool {
	for _, name := range names {
		if name.IsExported() {
			return true
		}
	}
	return false
}

func (c *checker) checkDepth(decl ast.Decl) {
	fn, ok := decl.(*ast.FuncDecl)
	if !ok || fn.Body == nil {
		return
	}
	c.measure(fn.Body, 0)
}

func (c *checker) measure(node ast.Node, depth int) {
	ast.Inspect(node, func(n ast.Node) bool {
		if n == node {
			return true
		}
		if lit, ok := n.(*ast.FuncLit); ok {
			c.measure(lit.Body, 0)
			return false
		}
		if !nests(n) {
			return true
		}
		if depth+1 > maxDepth {
			c.report(c.fset.Position(n.Pos()), fmt.Sprintf("profundidade maior que %d", maxDepth))
		}
		c.measure(n, depth+1)
		return false
	})
}

func nests(n ast.Node) bool {
	_, isIf := n.(*ast.IfStmt)
	_, isFor := n.(*ast.ForStmt)
	_, isRange := n.(*ast.RangeStmt)
	_, isSelect := n.(*ast.SelectStmt)
	return isIf || isFor || isRange || isSelect
}
