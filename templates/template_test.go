package templates

import (
	"errors"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestExtractFieldValues(t *testing.T) {
	t.Parallel()
	type Args struct {
		args            []any
		paramFieldNames []string
	}
	type Expected struct {
		fieldValues []any
		err         error
	}

	tests := map[string]struct {
		arrange func(*testing.T) (Args, Expected)
	}{
		"empty slice": {
			arrange: func(t *testing.T) (Args, Expected) {
				return Args{
						args:            []any{},
						paramFieldNames: []string{},
					}, Expected{
						fieldValues: []any{},
						err:         nil,
					}
			},
		},
		"same field names": {
			arrange: func(t *testing.T) (Args, Expected) {
				type TestStruct struct {
					Field1 string
					Field2 int
					Field3 bool
					Field4 float64
					Field5 time.Time
				}
				timeNow := time.Now()

				return Args{
						args: []any{
							TestStruct{"value1", 42, true, 3.14, timeNow},
							TestStruct{"value2", 84, false, 2.71, timeNow},
						},
						paramFieldNames: []string{"Field1", "Field2", "Field3", "Field4", "Field5"},
					}, Expected{
						fieldValues: []any{
							"value1", 42, true, 3.14, timeNow,
							"value2", 84, false, 2.71, timeNow,
						},
						err: nil,
					}
			},
		},
		"pointer slice": {
			arrange: func(t *testing.T) (Args, Expected) {
				type TestStruct struct {
					Field1 string
					Field2 int
				}
				return Args{
						args: []any{
							&TestStruct{"value1", 42},
							&TestStruct{"value2", 84},
						},
						paramFieldNames: []string{"Field1", "Field2"},
					}, Expected{
						fieldValues: []any{
							"value1", 42,
							"value2", 84,
						},
						err: nil,
					}
			},
		},
		"error: field not found": {
			arrange: func(t *testing.T) (Args, Expected) {
				type TestStruct struct {
					Field1 string
				}
				return Args{
						args:            []any{TestStruct{"value1"}},
						paramFieldNames: []string{"NonExistentField"},
					}, Expected{
						fieldValues: nil,
						err:         errors.New("field 'NonExistentField' not found"),
					}
			},
		},
		"error: not a struct slice": {
			arrange: func(t *testing.T) (Args, Expected) {
				return Args{
						args:            []any{"string1", "string2"},
						paramFieldNames: []string{"Field1"},
					}, Expected{
						fieldValues: nil,
						err:         errors.New("is not a struct or pointer to struct"),
					}
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			arg, expected := tc.arrange(t)
			result, err := extractFieldValues(arg.args, arg.paramFieldNames)
			if expected.err != nil {
				assert.ErrorContains(t, err, expected.err.Error())
				return
			}
			assert.NilError(t, err)
			assert.DeepEqual(t, result, expected.fieldValues)
		})
	}
}

func TestBuildBulkInsertQuery(t *testing.T) {
	t.Parallel()
	type Args struct {
		originalQuery   string
		numArgs         int
		numParamsPerArg int
	}
	type Expected struct {
		query string
		err   error
	}

	tests := map[string]struct {
		arrange func(*testing.T) (Args, Expected)
	}{
		"valid:Standard": {
			arrange: func(t *testing.T) (Args, Expected) {
				return Args{
						originalQuery:   "INSERT INTO users (id, name) VALUES (?, ?);",
						numArgs:         3,
						numParamsPerArg: 2,
					}, Expected{
						query: "INSERT INTO users (id, name) VALUES (?,?),(?,?),(?,?)",
						err:   nil,
					}
			},
		},
		"valid:upsert": {
			arrange: func(t *testing.T) (Args, Expected) {
				return Args{
						originalQuery:   "INSERT INTO users (id, name) VALUES (?, ?) ON DUPLICATE KEY UPDATE id = VALUES(id), name = VALUES(name);",
						numArgs:         3,
						numParamsPerArg: 2,
					}, Expected{
						query: "INSERT INTO users (id, name) VALUES (?,?),(?,?),(?,?) ON DUPLICATE KEY UPDATE id = VALUES(id), name = VALUES(name)",
						err:   nil,
					}
			},
		},
		"valid:upsert (ON DUPLICATE KEY UPDATE)": {
			arrange: func(t *testing.T) (Args, Expected) {
				return Args{
						originalQuery:   "INSERT INTO users (id, name) VALUES (?, ?) ON DUPLICATE KEY UPDATE id = VALUES(id), name = VALUES(name);",
						numArgs:         2,
						numParamsPerArg: 2,
					}, Expected{
						query: "INSERT INTO users (id, name) VALUES (?,?),(?,?) ON DUPLICATE KEY UPDATE id = VALUES(id), name = VALUES(name)",
						err:   nil,
					}
			},
		},
		"valid:upsert (ON DUPLICATE KEY UPDATE) case-insensitive": {
			arrange: func(t *testing.T) (Args, Expected) {
				return Args{
						originalQuery:   "insert into users (id, name) values (?, ?) on duplicate key update id = values(id), name = values(name);",
						numArgs:         2,
						numParamsPerArg: 2,
					}, Expected{
						query: "insert into users (id, name) VALUES (?,?),(?,?) on duplicate key update id = values(id), name = values(name)",
						err:   nil,
					}
			},
		},
		"valid:upsert (ON CONFLICT)": {
			arrange: func(t *testing.T) (Args, Expected) {
				return Args{
						originalQuery:   "INSERT INTO users (id, name) VALUES (?, ?) ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name;",
						numArgs:         2,
						numParamsPerArg: 2,
					}, Expected{
						query: "INSERT INTO users (id, name) VALUES (?,?),(?,?) ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name",
						err:   nil,
					}
			},
		},
		"valid:upsert (ON CONFLICT) case-insensitive": {
			arrange: func(t *testing.T) (Args, Expected) {
				return Args{
						originalQuery:   "insert into users (id, name) values (?, ?) on conflict (id) do nothing;",
						numArgs:         2,
						numParamsPerArg: 2,
					}, Expected{
						query: "insert into users (id, name) VALUES (?,?),(?,?) on conflict (id) do nothing",
						err:   nil,
					}
			},
		},
		"valid:RETURNING clause": {
			arrange: func(t *testing.T) (Args, Expected) {
				return Args{
						originalQuery:   "INSERT INTO users (id, name) VALUES (?, ?) RETURNING id;",
						numArgs:         2,
						numParamsPerArg: 2,
					}, Expected{
						query: "INSERT INTO users (id, name) VALUES (?,?),(?,?) RETURNING id",
						err:   nil,
					}
			},
		},
		"valid:RETURNING clause with ON CONFLICT": {
			arrange: func(t *testing.T) (Args, Expected) {
				return Args{
						originalQuery:   "INSERT INTO users (id, name) VALUES (?, ?) ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name RETURNING id;",
						numArgs:         2,
						numParamsPerArg: 2,
					}, Expected{
						query: "INSERT INTO users (id, name) VALUES (?,?),(?,?) ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name RETURNING id",
						err:   nil,
					}
			},
		},
		"valid:Squeeze spaces": {
			arrange: func(t *testing.T) (Args, Expected) {
				return Args{
						originalQuery:   "INSERT INTO users(id, name)VALUES(?,?);",
						numArgs:         3,
						numParamsPerArg: 2,
					}, Expected{
						query: "INSERT INTO users(id, name) VALUES (?,?),(?,?),(?,?)",
						err:   nil,
					}
			},
		},
		"valid:Extra line break": {
			arrange: func(t *testing.T) (Args, Expected) {
				return Args{
						originalQuery:   "INSERT\nINTO\nusers\n(id,\nname\n)\nVALUES\n(\n?,\n?\n);",
						numArgs:         3,
						numParamsPerArg: 2,
					}, Expected{
						query: "INSERT\nINTO\nusers\n(id,\nname\n) VALUES (?,?),(?,?),(?,?)",
						err:   nil,
					}
			},
		},
		"valid:Extra spaces": {
			arrange: func(t *testing.T) (Args, Expected) {
				return Args{
						originalQuery:   "INSERT INTO users   (   id    ,   name   )         VALUES       (   ?   ,  ?   )    ;",
						numArgs:         3,
						numParamsPerArg: 2,
					}, Expected{
						query: "INSERT INTO users   (   id    ,   name   ) VALUES (?,?),(?,?),(?,?)",
						err:   nil,
					}
			},
		},
		"valid:Extra suffix": {
			arrange: func(t *testing.T) (Args, Expected) {
				return Args{
						originalQuery:   "INSERT INTO users (id, name) VALUES (?, ?) ON DUPLICATE KEY UPDATE id = id;",
						numArgs:         3,
						numParamsPerArg: 2,
					}, Expected{
						query: "INSERT INTO users (id, name) VALUES (?,?),(?,?),(?,?) ON DUPLICATE KEY UPDATE id = id",
						err:   nil,
					}
			},
		},
		"error: numArgs is zero": {
			arrange: func(t *testing.T) (Args, Expected) {
				return Args{
						originalQuery:   "INSERT INTO users (id, name) VALUES (?, ?);",
						numArgs:         0, // 0
						numParamsPerArg: 2,
					}, Expected{
						query: "",
						err:   errors.New("number of arguments (rows) for bulk insert cannot be zero"),
					}
			},
		},
		"error: numParamsPerArg is zero": {
			arrange: func(t *testing.T) (Args, Expected) {
				return Args{
						originalQuery:   "INSERT INTO users (id, name) VALUES (?, ?);",
						numArgs:         3,
						numParamsPerArg: 0, // 0
					}, Expected{
						query: "",
						err:   errors.New("number of parameters per argument (columns) for bulk insert cannot be zero"),
					}
			},
		},
		"error: VALUES clause not found": {
			arrange: func(t *testing.T) (Args, Expected) {
				return Args{
						originalQuery:   "INSERT INTO users (id, name) SET id = ?, name = ?;", // VALUES がない
						numArgs:         3,
						numParamsPerArg: 2,
					}, Expected{
						query: "",
						err:   errors.New("VALUES clause not found"),
					}
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			arg, expected := tc.arrange(t)
			result, err := buildBulkInsertQuery(arg.originalQuery, arg.numArgs, arg.numParamsPerArg)
			if expected.err != nil {
				assert.ErrorContains(t, err, expected.err.Error())
				return
			}
			assert.NilError(t, err)
			assert.Equal(t, result, expected.query)
		})
	}
}
