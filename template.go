package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"go/format"
	"strconv"
	"strings"
	"text/template"

	"github.com/sqlc-dev/plugin-sdk-go/sdk"
)

func executeTemplate(
	_ context.Context, templateName string, data any,
) ([]byte, error) {
	funcMap := template.FuncMap{
		"lowerTitle": sdk.LowerTitle,
		"comment":    sdk.DoubleSlashComment,
		"escape":     sdk.EscapeBacktick,
		"hasPrefix":  strings.HasPrefix,
		// Helper to output string slices as Go slice literals
		"stringSliceLiteral": func(slice []string) string {
			if len(slice) == 0 {
				return "[]string{}"
			}
			var sb strings.Builder
			sb.WriteString("[]string{")
			for i, s := range slice {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(fmt.Sprintf("\"%s\"", s))
			}
			sb.WriteString("}")
			return sb.String()
		},
		"quote": strconv.Quote,
	}
	tmpl := template.Must(
		template.New("table").
			Funcs(funcMap).
			ParseFS(
				templates,
				"templates/*.tmpl",
			),
	)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	if err := tmpl.ExecuteTemplate(w, templateName, data); err != nil {
		return nil, err
	}
	if err := w.Flush(); err != nil {
		return nil, fmt.Errorf("failed to flush writer: %w", err)
	}
	code, err := format.Source(b.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to format generated code: %w", err)
	}
	return code, nil
}
