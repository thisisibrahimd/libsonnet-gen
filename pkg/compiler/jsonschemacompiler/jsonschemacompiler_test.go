package jsonschemacompiler_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/sebdah/goldie/v2"
	"github.com/thisisibrahimd/libsonnet-gen/pkg/compiler/jsonschemacompiler"
	"github.com/thisisibrahimd/libsonnet-gen/pkg/format"
)

func TestCompileLibsonnet(t *testing.T) {
	// create jsonschema compiler
	comp := jsonschema.NewCompiler()
	l, _ := jsonschemacompiler.NewLoader(false, "")
	comp.UseLoader(l)

	tests := []struct {
		name       string // description of this test case
		schema     *jsonschema.Schema
		schemaName string
		curPath    []string
	}{
		{
			name:   "golangci-schema",
			schema: comp.MustCompile("testdata/golangci-schema.json"),
		},
		{
			name:   "kind-schema",
			schema: comp.MustCompile("testdata/kind-schema.json"),
		},
		{
			name:   "identity-schema",
			schema: comp.MustCompile("testdata/identity-schema.json"),
		},
		{
			name:   "kratos-schema",
			schema: comp.MustCompile("testdata/kratos-schema.json"),
		},
		{
			name:   "simple-schema",
			schema: comp.MustCompile("testdata/simple-schema.json"),
		},
		{
			name:   "tslint-schema",
			schema: comp.MustCompile("testdata/tslint-schema.json"),
		},
		{
			name:   "map-test-schema",
			schema: comp.MustCompile("testdata/map-test-schema.json"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// compile jsonschema into libsonnet
			got := jsonschemacompiler.CompileLibsonnet(tt.schema, "schema", []string{})

			// format the generated libsonnet file
			formattedLibsonnet := formatDebug(t, tt.name, got.String())

			// compare
			verifyLibsonnetFileMatch(t, tt.name, []byte(formattedLibsonnet))
		})
	}
}

// verifyLibsonnetFileMatch verify that the generated libsonnet file matches the golden libsonnet file in our golden directory
func verifyLibsonnetFileMatch(t *testing.T, name string, v []byte) {
	g := goldie.New(
		t,
		goldie.WithFixtureDir("./testdata"),
		goldie.WithNameSuffix(".golden.libsonnet"),
		goldie.WithTestNameForDir(false),
		goldie.WithSubTestNameForDir(false),
	)

	g.Assert(t, strings.ReplaceAll(name, " ", "_"), v)
}

// formatDebug formats content and, on failure, writes the raw output to a temp
// file for inspection. Test helper only — not exposed to end users.
func formatDebug(t *testing.T, path string, content string) string {
	t.Helper()
	formatted, err := format.Format(path, content)
	if err != nil {
		debugPath := filepath.Join(t.TempDir(), filepath.Base(path)+".libsonnet")
		os.WriteFile(debugPath, []byte(content), 0o644)
		t.Fatalf("%s: %v\nraw content dumped to: %s", path, err, debugPath)
	}
	return formatted
}
