package main

import (
	"testing"

	"gotest.tools/v3/assert"
)

func Test_snakeToPascalCase(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		input string
		want  string
	}{
		"empty": {
			input: "",
			want:  "",
		},
		"simple": {
			// "name" -> "Name"
			input: "name",
			want:  "Name",
		},
		"standard snake case": {
			// "user_name" -> "UserName"
			input: "user_name",
			want:  "UserName",
		},
		"common initialism (id)": {
			// "id" -> "ID"
			input: "user_id",
			want:  "UserID",
		},
		"common initialism (id) at start": {
			input: "id_value",
			want:  "IDValue",
		},
		"common initialism (url)": {
			// "url" -> "URL"
			input: "profile_url",
			want:  "ProfileURL",
		},
		"common initialism (json)": {
			// "json" -> "JSON"
			input: "json_data",
			want:  "JSONData",
		},
		"double underscore": {
			// "user__id" -> "UserID"
			input: "user__id",
			want:  "UserID",
		},
		"leading underscore": {
			// "_user_name" -> "UserName"
			input: "_user_name",
			want:  "UserName",
		},
		"trailing underscore": {
			// "user_name_" -> "UserName"
			input: "user_name_",
			want:  "UserName",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := snakeToPascalCase(tt.input)
			assert.Equal(t, got, tt.want)
		})
	}
}
