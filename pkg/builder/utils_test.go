package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sebdah/goldie/v2"
	"github.com/thisisibrahimd/libsonnet-gen/pkg/format"
)

// assertRender verifies that the rendered Type matches the golden file.
// Golden files are stored in ./testdata/<name>.golden.libsonnet.
// Spaces in name are replaced with underscores for valid filenames.
// The rendered output is also validated as syntactically valid Jsonnet.
func assertRender(t *testing.T, name string, o Type) {
	t.Helper()
	output := Doc{Root: o}.String()
	formatDebug(t, name, output)

	g := goldie.New(t,
		goldie.WithFixtureDir("./testdata"),
		goldie.WithNameSuffix(".golden.libsonnet"),
		goldie.WithTestNameForDir(false),
		goldie.WithSubTestNameForDir(false),
	)
	g.Assert(t, strings.ReplaceAll(name, " ", "_"), []byte(output))
}

// assertValidJsonnet verifies that output is syntactically valid Jsonnet
// by running it through the go-jsonnet formatter, which parses before formatting.
func assertValidJsonnet(t *testing.T, name string, output string) {
	t.Helper()
	formatDebug(t, name, output)
}

// formatDebug formats content and, on failure, writes the raw output to a temp
// file for inspection. Test helper only — not exposed to end users.
func formatDebug(t *testing.T, path string, content string) {
	t.Helper()
	_, err := format.Format(path, content)
	if err != nil {
		debugPath := filepath.Join(t.TempDir(), filepath.Base(path)+".libsonnet")
		os.WriteFile(debugPath, []byte(content), 0o644)
		t.Fatalf("%s: %v\nraw content dumped to: %s", path, err, debugPath)
	}
}

// assertPanics verifies that fn panics and optionally checks the panic message.
func assertPanics(t *testing.T, name string, fn func(), wantMsg string) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic, but did not panic")
			}
			if wantMsg != "" {
				msg := fmt.Sprintf("%v", r)
				if !strings.Contains(msg, wantMsg) {
					t.Errorf("panic message = %q, want to contain %q", msg, wantMsg)
				}
			}
		}()
		fn()
	})
}

// assertRenderBytes verifies that output matches the golden file.
// Use when rendering is not via Doc{Root: Type}, e.g. Doc with Locals.
// Spaces in name are replaced with underscores for valid filenames.
// Output is validated as syntactically valid Jsonnet.
func assertRenderBytes(t *testing.T, name string, output string) {
	t.Helper()
	assertValidJsonnet(t, name, output)

	g := goldie.New(t,
		goldie.WithFixtureDir("./testdata"),
		goldie.WithNameSuffix(".golden.libsonnet"),
		goldie.WithTestNameForDir(false),
		goldie.WithSubTestNameForDir(false),
	)
	g.Assert(t, strings.ReplaceAll(name, " ", "_"), []byte(output))
}
