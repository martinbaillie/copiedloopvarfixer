package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/ast/astutil"
)

const (
	fmtRemove  = "Removing %s:%d, "
	fmtAliased = "Aliased names (manual intervention required) %s:%d, "
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <dir_to_walk>\n", os.Args[0])
		return
	}

	if err := processDir(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func processDir(root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".go" {
			return err
		}

		return processFile(path)
	})
}

func processFile(path string) error {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	toRemove := make(map[ast.Stmt]bool)

	ast.Inspect(file, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.ForStmt, *ast.RangeStmt:
			ast.Inspect(n.(ast.Stmt), func(n ast.Node) bool {
				if assign, ok := n.(*ast.AssignStmt); ok && assign.Tok == token.DEFINE {
					for _, rhsExpr := range assign.Rhs {
						if id, ok := rhsExpr.(*ast.Ident); ok && id.Obj != nil && id.Obj.Kind == ast.Var {
							for _, lhsExpr := range assign.Lhs {
								if lhsIdent, ok := lhsExpr.(*ast.Ident); ok {
									if lhsIdent.Name == id.Name {
										toRemove[assign] = true
									}

									pos := fset.Position(assign.Pos())
									if toRemove[assign] {
										fmt.Fprintf(os.Stderr, fmtRemove, pos.Filename, pos.Line)
									} else {
										fmt.Fprintf(os.Stderr, fmtAliased, pos.Filename, pos.Line)
									}
									printer.Fprint(os.Stderr, fset, assign)
									fmt.Println()
								}
							}
						}
					}
				}

				return true
			})
		}

		return true
	})

	fileChanged := false
	astutil.Apply(file, func(c *astutil.Cursor) bool {
		if stmt, ok := c.Node().(ast.Stmt); ok && toRemove[stmt] {
			c.Delete()
			fileChanged = true
		}

		return true
	}, nil)

	if fileChanged {
		fileOut, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0o666)
		if err != nil {
			return err
		}
		defer fileOut.Close()

		if err := format.Node(fileOut, fset, file); err != nil {
			return err
		}
	}

	return nil
}
