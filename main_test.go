package main_test

import (
	"errors"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
	"testing"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	sqlcpluginbulkgo "github.com/tomtwinkle/process-plugin-sqlc-gen-bulk-go"
	"gotest.tools/v3/assert"
)

func TestGenerate(t *testing.T) {
	t.Parallel()

	type Args struct {
		req *plugin.GenerateRequest
	}
	type Expected struct {
		fileCount int
		err       error
	}

	tests := map[string]struct {
		arrange func(*testing.T) (Args, Expected)
	}{
		"valid:Normal INSERT Query": {
			arrange: func(t *testing.T) (Args, Expected) {
				req := &plugin.GenerateRequest{
					SqlcVersion:   "1.0.0",
					PluginOptions: []byte(`{"package": "sqlc"}`),
					Queries: []*plugin.Query{
						{
							Name: "InsertUser",
							Text: "INSERT INTO users (id, name) VALUES (?, ?)",
							Params: []*plugin.Parameter{
								{Column: &plugin.Column{Name: "id"}},
								{Column: &plugin.Column{Name: "name"}},
							},
						},
					},
				}
				return Args{req: req}, Expected{fileCount: 1, err: nil}
			},
		},
		"valid:Extra suffix INSERT Query": {
			arrange: func(t *testing.T) (Args, Expected) {
				req := &plugin.GenerateRequest{
					SqlcVersion:   "1.0.0",
					PluginOptions: []byte(`{"package": "sqlc"}`),
					Queries: []*plugin.Query{
						{
							Name: "InsertUser",
							Text: "INSERT INTO users (id, name) VALUES (?, ?) ON DUPLICATE KEY UPDATE id = id",
							Params: []*plugin.Parameter{
								{Column: &plugin.Column{Name: "id"}},
								{Column: &plugin.Column{Name: "name"}},
							},
						},
					},
				}
				return Args{req: req}, Expected{fileCount: 1, err: nil}
			},
		},
		"valid:No INSERT Queries": {
			arrange: func(t *testing.T) (Args, Expected) {
				req := &plugin.GenerateRequest{
					SqlcVersion:   "1.0.0",
					PluginOptions: []byte(`{"package": "sqlc"}`),
					Queries: []*plugin.Query{
						{
							Name: "SelectUser",
							Text: "SELECT * FROM users WHERE id = ?",
							Params: []*plugin.Parameter{
								{Column: &plugin.Column{Name: "id"}},
							},
						},
						{
							Name:   "InsertUserNoParams",
							Text:   "INSERT INTO users DEFAULT VALUES",
							Params: []*plugin.Parameter{}, // No parameters
						},
					},
				}
				return Args{req: req}, Expected{fileCount: 0, err: nil}
			},
		},
		"invalid:Options validation error": {
			arrange: func(t *testing.T) (Args, Expected) {
				req := &plugin.GenerateRequest{
					SqlcVersion:   "1.0.0",
					PluginOptions: []byte(`{}`), // "package" is required
					Queries:       []*plugin.Query{},
				}
				return Args{req: req}, Expected{err: errors.New(`"package" is required`)}
			},
		},
		"invalid:Options parse error": {
			arrange: func(t *testing.T) (Args, Expected) {
				req := &plugin.GenerateRequest{
					SqlcVersion:   "1.0.0",
					PluginOptions: []byte(`{"package": "sqlc"`), // Missing closing brace
					Queries:       []*plugin.Query{},
				}
				return Args{req: req}, Expected{err: errors.New("unexpected end of JSON input")}
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			args, want := tc.arrange(t)

			got, err := sqlcpluginbulkgo.Generate(t.Context(), args.req)
			if want.err != nil {
				assert.ErrorContains(t, err, want.err.Error())
				return
			}
			assert.NilError(t, err)
			assert.Equal(t, len(got.Files), want.fileCount, "Generated file count mismatch")

			// To perform type checking with assertGeneratedCodeIsValid,
			// we prepare a minimal mock of the code sqlc-gen-go is generate.
			const mockBaseGo = `
package sqlc

import (
	"context"
	"database/sql"
)

type DBTX interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

type Queries struct {
	db DBTX
}

const insertUser = "INSERT INTO users (id, name) VALUES (?, ?)"

type InsertUserParams struct {
	ID   any
	Name any
}
`
			// Combine mock files and generated files into slices
			mockFile := &plugin.File{
				Name:     "mock_base.go",
				Contents: []byte(mockBaseGo),
			}
			allFiles := append(got.Files, mockFile)

			assertGeneratedCodeIsValid(t, allFiles)
		})
	}
}

func assertGeneratedCodeIsValid(t *testing.T, files []*plugin.File) {
	t.Helper()

	fset := token.NewFileSet()
	var parsedFiles []*ast.File

	for _, file := range files {
		if !strings.HasSuffix(file.Name, ".go") {
			continue
		}

		node, err := parser.ParseFile(fset, file.Name, file.Contents, parser.ParseComments)
		if err != nil {
			t.Fatalf("Failed to parse file %s: %v", file.Name, err)
		}
		parsedFiles = append(parsedFiles, node)
	}
	if len(parsedFiles) == 0 {
		return
	}
	conf := types.Config{
		Importer: importer.Default(),
		Error: func(err error) {
			t.Fatalf("Failed to type check generated code: %v", err)
		},
	}
	pkgPath := parsedFiles[0].Name.Name
	_, err := conf.Check(pkgPath, fset, parsedFiles, nil)
	if err != nil {
		t.Fatalf("Type checking failed for generated code: %v", err)
	}
}
