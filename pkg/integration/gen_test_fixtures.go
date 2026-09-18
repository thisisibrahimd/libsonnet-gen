//go:build ignore

package main

import (
	"fmt"
	"os"

	j "github.com/thisisibrahimd/libsonnet-gen/pkg/builder"
	"github.com/thisisibrahimd/libsonnet-gen/pkg/model"
	"github.com/thisisibrahimd/libsonnet-gen/pkg/render"
	"github.com/thisisibrahimd/libsonnet-gen/pkg/swagger"
)

const dMock = `local d = {
  fn(help='', args=[]): null,
  arg(name='', type=''): null,
  T: {array: 'array', object: 'object', string: 'string', 'function': 'function'},
};

`

func main() {
	cases := []struct {
		name    string
		modName string
		mod     model.Modifier
	}{
		{
			name:    "simple_array",
			modName: "withFinalizers",
			mod: model.Modifier{
				Arg:    model.Parameter{Key: "finalizers"},
				Target: "metadata.finalizers",
				Type:   swagger.TypeArray,
			},
		},
		{
			name:    "nested_path",
			modName: "withContainers",
			mod: model.Modifier{
				Arg:    model.Parameter{Key: "containers"},
				Target: "spec.containers",
				Type:   swagger.TypeArray,
			},
		},
		{
			name:    "by_name",
			modName: "withContainers",
			mod: model.Modifier{
				Arg:         model.Parameter{Key: "containers"},
				Target:      "spec.containers",
				Type:        swagger.TypeArray,
				ItemHasName: true,
			},
		},
	}

	for _, tc := range cases {
		types := render.Modifier(tc.modName, tc.mod)
		genObj := j.Object(tc.modName, types...)
		out := dMock + genObj.String()
		path := fmt.Sprintf("pkg/integration/testdata/%s.modifier.libsonnet", tc.name)
		os.WriteFile(path, []byte(out), 0644)
		fmt.Println(path)
	}
}
