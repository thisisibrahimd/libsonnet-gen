package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sebdah/goldie/v2"
	"github.com/thisisibrahimd/libsonnet-gen/pkg/format"
	"github.com/thisisibrahimd/libsonnet-gen/pkg/model"
	"github.com/thisisibrahimd/libsonnet-gen/pkg/render"
	"github.com/thisisibrahimd/libsonnet-gen/pkg/swagger"
)

type k8sTestCase struct {
	version string
	swagger string
	prefix  string
}

var k8sTestCases = []k8sTestCase{
	{
		version: "1.35",
		swagger: "k8s-swagger-1.35.json",
		prefix:  "^io\\.k8s\\.(api|(kube-aggregator|apimachinery)\\.pkg\\.apis)\\..*",
	},
	// Add more versions here:
	// {
	// 	version: "1.34",
	// 	swagger: "k8s-swagger-1.34.json",
	// 	prefix:  "^io\\.k8s\\.(api|(kube-aggregator|apimachinery)\\.pkg\\.apis)\\..*",
	// },
}

func TestK8sSwagger_FullPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test: requires k8s swagger spec (~4MB)")
	}

	swaggerDir := filepath.Join("..", "swagger", "testdata")

	for _, tc := range k8sTestCases {
		t.Run(tc.version, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(swaggerDir, tc.swagger))
			if err != nil {
				t.Fatalf("failed to read swagger spec %s: %v", tc.swagger, err)
			}

			loader := &swagger.SwaggerLoader{}
			defs, err := loader.Load(data)
			if err != nil {
				t.Fatalf("failed to load swagger definitions: %v", err)
			}

			groups := model.Load(&defs, tc.prefix)
			if len(groups) == 0 {
				t.Fatal("expected non-empty groups from k8s swagger spec")
			}

			goldenDir := filepath.Join("..", "swagger", "testdata", "golden-k8s-"+tc.version)
			g := goldie.New(t,
				goldie.WithFixtureDir(goldenDir),
				goldie.WithNameSuffix(""),
				goldie.WithTestNameForDir(false),
				goldie.WithSubTestNameForDir(false),
			)

			index := render.Index(groups, "k8s", "", tc.version, "k8s jsonnet library")
			g.Assert(t, "gen.libsonnet", []byte(formatDebug(t, "gen.libsonnet", index.String())))

			main := render.Main(nil)
			g.Assert(t, "main.libsonnet", []byte(formatDebug(t, "main.libsonnet", main.String())))

			for name, group := range groups {
				generated := render.Group(name, group)
				for fn, obj := range generated {
					g.Assert(t, filepath.Join("_gen", name, fn), []byte(formatDebug(t, filepath.Join("_gen", name, fn), obj.String())))
				}
			}
		})
	}
}

// formatDebug formats content and, on failure, writes the raw output to a temp
// file for inspection. Test helper only — not exposed to end users.
func formatDebug(t *testing.T, path string, content string) string {
	t.Helper()
	formatted, err := format.Format(path, content)
	if err != nil {
		debugDir := filepath.Join("..", "swagger", "testdata", "debug")
		os.MkdirAll(debugDir, 0o755)
		debugPath := filepath.Join(debugDir, filepath.Base(path))
		os.WriteFile(debugPath, []byte(content), 0o644)
		t.Fatalf("%s: %v\nraw content dumped to: %s", path, err, debugPath)
	}
	return formatted
}
