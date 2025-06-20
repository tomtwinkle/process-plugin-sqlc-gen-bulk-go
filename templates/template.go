package templates

import (
	"fmt"
	"reflect"
	"strings"
)

// extractFieldValues takes a slice of a structure and an ordered list of field names to extract,
// extracts field values from all structures in the specified order and returns them as a flat []any slice.
func extractFieldValues[T any](args []T, paramFieldNames []string) ([]any, error) {
	values := make([]any, 0, len(args)*len(paramFieldNames))
	if len(paramFieldNames) == 0 {
		return values, nil
	}

	for i, arg := range args {
		rv := reflect.ValueOf(arg)
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}

		if rv.Kind() != reflect.Struct {
			return nil, fmt.Errorf("args[%d] (type %T) is not a struct or pointer to struct", i, arg)
		}

		for _, fieldName := range paramFieldNames {
			field := rv.FieldByName(fieldName)
			if !field.IsValid() {
				return nil, fmt.Errorf("args[%d] (type %T): field '%s' not found", i, arg, fieldName)
			}
			if !field.CanInterface() {
				return nil, fmt.Errorf("args[%d] (type %T): field '%s' cannot be interfaced (not exported or unaddressable)",
					i, arg, fieldName)
			}
			values = append(values, field.Interface())
		}
	}
	return values, nil
}

// buildBulkInsertQuery builds a SQL query string for bulk inserts.
// originalQuery: the original INSERT statement (e.g., "INSERT INTO users (id, name) VALUES ($1, $2)")
// numArgs: number of rows of data to insert
// numParamsPerArg: number of parameters per row (number of columns)
func buildBulkInsertQuery(originalQuery string, numArgs int, numParamsPerArg int) (string, error) {
	if numArgs == 0 {
		return "", fmt.Errorf("number of arguments (rows) for bulk insert cannot be zero")
	}
	if numParamsPerArg == 0 {
		return "", fmt.Errorf("number of parameters per argument (columns) for bulk insert cannot be zero")
	}

	// Extract the "INSERT INTO table (col1, col2) VALUES " part from the original query
	// First, remove the trailing semicolon, if any
	trimmedQuery := strings.TrimSpace(originalQuery)
	if strings.HasSuffix(trimmedQuery, ";") {
		trimmedQuery = trimmedQuery[:len(trimmedQuery)-1]
	}

	// search "VALUES" (case insensitive)
	valuesUpperIndex := strings.LastIndex(strings.ToUpper(trimmedQuery), "VALUES")
	if valuesUpperIndex == -1 {
		return "", fmt.Errorf("invalid query format: VALUES clause not found in original query: %s", originalQuery)
	}

	// Prefix the query up to "VALUES".
	// (e.g., "INSERT INTO users (id, name)")
	// Add "VALUES" to this
	queryPrefixStr := strings.TrimSpace(trimmedQuery[:valuesUpperIndex]) + " VALUES "

	// Find the suffix of the query (e.g., "ON DUPLICATE KEY UPDATE ...")
	// The suffix starts after the original placeholder, e.g., after `VALUES (?, ?)`
	var querySuffixStr string
	// Start searching after the "VALUES" keyword
	searchStartIndex := valuesUpperIndex + len("VALUES")
	// Find the first closing parenthesis ')' after the "VALUES" clause
	firstClosingParenIndex := strings.Index(trimmedQuery[searchStartIndex:], ")")

	if firstClosingParenIndex != -1 {
		// The suffix is the part of the string after the closing parenthesis.
		// We add 1 to skip the ')' character itself.
		suffixStartIndex := searchStartIndex + firstClosingParenIndex + 1
		if suffixStartIndex < len(trimmedQuery) {
			querySuffixStr = " " + strings.TrimSpace(trimmedQuery[suffixStartIndex:])
		}
	}

	var queryBuilder strings.Builder
	queryBuilder.WriteString(queryPrefixStr)

	valueStrings := make([]string, numArgs)
	for i := 0; i < numArgs; i++ {
		placeholders := make([]string, numParamsPerArg)
		for j := 0; j < numParamsPerArg; j++ {
			placeholders[j] = "?"
		}
		valueStrings[i] = fmt.Sprintf("(%s)", strings.Join(placeholders, ","))
	}
	queryBuilder.WriteString(strings.Join(valueStrings, ","))

	// Append the suffix if it exists.
	if querySuffixStr != "" {
		queryBuilder.WriteString(querySuffixStr)
	}

	return strings.TrimSpace(queryBuilder.String()), nil
}
