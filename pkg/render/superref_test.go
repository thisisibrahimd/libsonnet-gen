package render

import (
	"testing"

	j "github.com/thisisibrahimd/libsonnet-gen/pkg/builder"
	"github.com/stretchr/testify/assert"
)

func TestSuperRef(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// Reserved keywords must use bracket notation
		{"reserved-function", "function", "super['function']"},
		{"reserved-self", "self", "super['self']"},
		{"reserved-super", "super", "super['super']"},
		{"reserved-local", "local", "super['local']"},
		{"reserved-for", "for", "super['for']"},
		{"reserved-if", "if", "super['if']"},
		{"reserved-else", "else", "super['else']"},
		{"reserved-true", "true", "super['true']"},
		{"reserved-false", "false", "super['false']"},
		{"reserved-null", "null", "super['null']"},
		{"reserved-import", "import", "super['import']"},
		{"reserved-importstr", "importstr", "super['importstr']"},
		{"reserved-tailstrict", "tailstrict", "super['tailstrict']"},
		{"reserved-in", "in", "super['in']"},
		{"reserved-assert", "assert", "super['assert']"},
		{"reserved-error", "error", "super['error']"},
		{"reserved-then", "then", "super['then']"},
		// Special characters must use bracket notation
		{"hyphenated", "api-version", "super['api-version']"},
		{"dotted", "api.version", "super['api.version']"},
		{"slashed", "path/to", "super['path/to']"},
		{"hash-prefix", "#comment", "super['#comment']"},
		// Normal identifiers use dot notation
		{"normal-identifier", "name", "super.name"},
		{"normal-camelCase", "dnsNames", "super.dnsNames"},
		{"normal-underscore", "my_field", "super.my_field"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := j.SuperRef(tt.input).String()
			assert.Equal(t, tt.want, got)
		})
	}
}
