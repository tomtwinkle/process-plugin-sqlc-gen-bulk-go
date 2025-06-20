package main

import (
	"context"
	"fmt"

	"github.com/sqlc-dev/plugin-sdk-go/codegen"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

const (
	generateFileName = "bulk.sql.go"

	sourceTemplateFuncPath = "templates/template.go"
	sourceTemplateFunc1    = "extractFieldValues"
	sourceTemplateFunc2    = "buildBulkInsertQuery"
)

func main() {
	codegen.Run(Generate)
}

func Generate(ctx context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	opts, err := ParseOptions(req)
	if err != nil {
		return nil, err
	}
	if err := ValidateOptions(opts); err != nil {
		return nil, err
	}

	bulkInserts := buildBulkInsert(req, opts)

	if len(bulkInserts) == 0 {
		// Returns an empty response if nothing is generated
		return &plugin.GenerateResponse{}, nil
	}

	// Return the response with the generated code
	return generate(ctx, req, opts, bulkInserts)
}

func generate(
	ctx context.Context, req *plugin.GenerateRequest, opts *Options, structs BulkInserts,
) (*plugin.GenerateResponse, error) {
	extractFieldValuesFn, err := parseGoCode(sourceTemplateFuncPath, sourceTemplateFunc1)
	if err != nil {
		return nil, fmt.Errorf("failed to parse function %s: %w", sourceTemplateFunc1, err)
	}
	buildBulkInsertQueryFn, err := parseGoCode(sourceTemplateFuncPath, sourceTemplateFunc2)
	if err != nil {
		return nil, fmt.Errorf("failed to parse function %s: %w", sourceTemplateFunc2, err)
	}

	tmpl := struct {
		Package       string
		SqlcVersion   string
		BulkInsert    []BulkInsert
		ExtractFnName string
		ExtractFn     string
		BuildFnName   string
		BuildFn       string
	}{
		Package:       opts.Package,
		SqlcVersion:   req.GetSqlcVersion(),
		BulkInsert:    structs,
		ExtractFnName: sourceTemplateFunc1,
		ExtractFn:     string(extractFieldValuesFn),
		BuildFnName:   sourceTemplateFunc2,
		BuildFn:       string(buildBulkInsertQueryFn),
	}

	code, err := executeTemplate(ctx, "bulkInsertFile", tmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}
	return &plugin.GenerateResponse{
		Files: []*plugin.File{
			{
				Name:     generateFileName,
				Contents: code,
			},
		},
	}, nil
}
