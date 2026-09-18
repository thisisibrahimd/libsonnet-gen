package docsonnet

import (
	"strings"
	"testing"

	j "github.com/thisisibrahimd/libsonnet-gen/pkg/builder"
	"github.com/sebdah/goldie/v2"
)

// assertRender verifies that the rendered Type matches the golden file.
// Spaces in name are replaced with underscores for valid filenames.
func assertRender(t *testing.T, name string, o j.Type) {
	t.Helper()
	g := goldie.New(t,
		goldie.WithFixtureDir("./testdata"),
		goldie.WithNameSuffix(".golden.libsonnet"),
		goldie.WithTestNameForDir(false),
		goldie.WithSubTestNameForDir(false),
	)
	g.Assert(t, strings.ReplaceAll(name, " ", "_"), []byte(j.Doc{Root: o}.String()))
}

func TestImport(t *testing.T) {
	tests := []struct {
		name string
		fn   func() j.Type
	}{
		{
			name: "testimport-default",
			fn:   Import,
		},
		{
			name: "testimport-custom path",
			fn: func() j.Type {
				return j.Import("util", "doc-util/main.libsonnet")
			},
		},
		{
			name: "testimport-relative path",
			fn: func() j.Type {
				return j.Import("util", "../doc-util/main.libsonnet")
			},
		},
		{
			name: "testimport-github path",
			fn: func() j.Type {
				return j.Import("util", "github.com/example/doc-util/main.libsonnet")
			},
		},
		{
			name: "testimport-importstr",
			fn: func() j.Type {
				return j.ImportStr("raw", "doc-util/README.md")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn()
			assertRender(t, tt.name, got)
		})
	}
}

func TestArgs(t *testing.T) {
	tests := []struct {
		name        string
		input       []string
		wantLen     int
		wantSubstrs []string
	}{
		{
			name:        "single pair",
			input:       []string{"name", "String"},
			wantLen:     1,
			wantSubstrs: []string{`name="name"`, `type=d.T.String`},
		},
		{
			name:        "two pairs",
			input:       []string{"host", "String", "port", "Int"},
			wantLen:     2,
			wantSubstrs: []string{`name="host"`, `type=d.T.String`, `name="port"`, `type=d.T.Int`},
		},
		{
			name:        "empty",
			input:       []string{},
			wantLen:     0,
			wantSubstrs: []string{},
		},
		{
			name:        "three pairs",
			input:       []string{"name", "String", "replicas", "Int", "enabled", "Bool"},
			wantLen:     3,
			wantSubstrs: []string{`name="name"`, `type=d.T.String`, `name="replicas"`, `type=d.T.Int`, `name="enabled"`, `type=d.T.Bool`},
		},
		{
			name:        "four pairs",
			input:       []string{"host", "String", "port", "Int", "path", "String", "tls", "Bool"},
			wantLen:     4,
			wantSubstrs: []string{`name="host"`, `type=d.T.String`, `name="port"`, `type=d.T.Int`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := Args(tt.input...)
			if len(args) != tt.wantLen {
				t.Fatalf("Args(%v) returned %d types, want %d", tt.input, len(args), tt.wantLen)
			}
			for _, want := range tt.wantSubstrs {
				found := false
				for _, a := range args {
					if strings.Contains(a.String(), want) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Args() missing substring %q", want)
				}
			}
		})
	}
}

func TestArgs_Panics(t *testing.T) {
	tests := []struct {
		name    string
		fn      func()
		wantMsg string
	}{
		{
			name:    "single arg (odd)",
			fn:      func() { Args("name") },
			wantMsg: "Args expects even number of arguments",
		},
		{
			name:    "three args (odd)",
			fn:      func() { Args("name", "String", "extra") },
			wantMsg: "Args expects even number of arguments",
		},
		{
			name:    "five args (odd)",
			fn:      func() { Args("a", "1", "b", "2", "c") },
			wantMsg: "Args expects even number of arguments",
		},
		{
			name:    "seven args (odd)",
			fn:      func() { Args("a", "1", "b", "2", "c", "3", "d") },
			wantMsg: "Args expects even number of arguments",
		},
		{
			name:    "nine args (odd)",
			fn:      func() { Args("a", "1", "b", "2", "c", "3", "d", "4", "e") },
			wantMsg: "Args expects even number of arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil {
					t.Fatal("expected panic, but did not panic")
				}
				if tt.wantMsg != "" {
					msg := r.(string)
					if msg != tt.wantMsg && !strings.Contains(msg, tt.wantMsg) {
						t.Errorf("panic message = %q, want to contain %q", msg, tt.wantMsg)
					}
				}
			}()
			tt.fn()
		})
	}
}

func TestFunc(t *testing.T) {
	tests := []struct {
		name   string
		fnName string
		help   string
		args   []string
	}{
		{
			name:   "single arg",
			fnName: "new",
			help:   "Create a new widget",
			args:   []string{"kind", "String"},
		},
		{
			name:   "multi arg",
			fnName: "deploy",
			help:   "Deploy a service",
			args:   []string{"name", "String", "replicas", "Int"},
		},
		{
			name:   "no args",
			fnName: "init",
			help:   "Initialize the module",
			args:   []string{},
		},
		{
			name:   "bool arg",
			fnName: "enable",
			help:   "Enable a feature",
			args:   []string{"feature", "String", "enabled", "Bool"},
		},
		{
			name:   "many args",
			fnName: "configure",
			help:   "Configure a service",
			args:   []string{"name", "String", "port", "Int", "tls", "Bool"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Func(tt.fnName, tt.help, Args(tt.args...))
			assertRender(t, tt.name, got)
		})
	}
}

func TestObj(t *testing.T) {
	tests := []struct {
		name string
		obj  string
		help string
	}{
		{
			name: "testobj-simple",
			obj:  "Widget",
			help: "A widget configuration",
		},
		{
			name: "testobj-multi word",
			obj:  "MyService",
			help: "Service definition",
		},
		{
			name: "testobj-empty help",
			obj:  "Config",
			help: "",
		},
		{
			name: "testobj-api object",
			obj:  "Deployment",
			help: "Kubernetes Deployment resource",
		},
		{
			name: "testobj-nested object",
			obj:  "Spec",
			help: "Spec contains the desired state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Obj(tt.obj, tt.help)
			assertRender(t, tt.name, got)
		})
	}
}

func TestPkg(t *testing.T) {
	tests := []struct {
		name string
		pkg  string
		url  string
		help string
	}{
		{
			name: "testpkg-default",
			pkg:  "mylib",
			url:  "https://example.com/lib",
			help: "My library docs",
		},
		{
			name: "testpkg-empty help",
			pkg:  "lib",
			url:  "github.com/foo/lib",
			help: "",
		},
		{
			name: "testpkg-k8s package",
			pkg:  "k8s",
			url:  "https://github.com/kubernetes/k8s-lib",
			help: "Kubernetes library",
		},
		{
			name: "testpkg-versioned package",
			pkg:  "mylib.v2",
			url:  "https://example.com/v2",
			help: "Version 2 of my library",
		},
		{
			name: "testpkg-github package",
			pkg:  "utils",
			url:  "github.com/example/utils",
			help: "Common utilities",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Pkg(tt.pkg, tt.url, tt.help)
			assertRender(t, tt.name, got)
		})
	}
}
