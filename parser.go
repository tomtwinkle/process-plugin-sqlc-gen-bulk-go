package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

// parseGoCode parses a Go source file and extracts the code of a specific function by its name.
// It returns the function code as a byte slice or an error if the function is not found.
func parseGoCode(sourceFile string, targetFuncName string) ([]byte, error) {
	srcBytes, err := templates.ReadFile(sourceFile)
	if err != nil {
		return nil, err
	}

	// Parsing files and building ASTs with go/parser
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, sourceFile, srcBytes, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var funcCode []byte
	// Scanning AST top-level declarations
	ast.Inspect(node, func(n ast.Node) bool {
		// Find function declarations (FuncDecl)
		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			return true // If it is not a function declaration, continue the search.
		}

		// Check if it matches the desired function name
		if fn.Name.Name == targetFuncName {
			// Get the start and end positions of the function
			start := fset.Position(fn.Pos()).Offset
			end := fset.Position(fn.End()).Offset

			// Cut out the function part as a string from the original source code
			funcCode = srcBytes[start:end]
			return false // Finish the search because the desired function has been found.
		}
		return true
	})

	if len(funcCode) == 0 {
		return nil, fmt.Errorf("function '%s' not found", targetFuncName)
	}
	return funcCode, nil
}
