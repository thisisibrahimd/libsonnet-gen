// Package docsonnet provides helpers for generating docsonnet documentation.
//
// Docsonnet is a documentation format for Jsonnet libraries. This package
// provides builders to create package, function, and object documentation
// that can be merged into generated Jsonnet output.
package docsonnet

import (
	"fmt"
	"strings"

	j "github.com/thisisibrahimd/libsonnet-gen/pkg/builder"
)

const dPkg = "doc-util/main.libsonnet"

// Import returns a local binding for the docsonnet package.
//
// Use this to add the docsonnet import to a generated document.
func Import() j.Type {
	return j.Local(j.Import("d", dPkg))
}

// Args converts name-type pairs into docsonnet argument definitions.
//
// Each pair becomes a call to d.arg(name=..., type=d.T...).
// Panics if the number of arguments is odd.
func Args(s ...string) []j.Type {
	if len(s)%2 != 0 {
		panic("Args expects even number of arguments (pairs)")
	}

	args := make([]j.Type, 0, len(s)/2)
	for i := range s {
		if i%2 != 0 {
			continue
		}

		typeRef := dTRef(s[i+1])
		args = append(args, j.Call("", "d.arg", []j.Type{
			j.String("name", s[i]),
			j.Ref("type", typeRef),
		}))
	}

	return args
}

// dTRef returns a safe reference to d.T.<type>, using bracket notation
// when the type is a reserved Jsonnet keyword.
func dTRef(typeName string) string {
	if j.IsReservedKeyword(typeName) || strings.ContainsAny(typeName, "-./") || strings.HasPrefix(typeName, "#") {
		return fmt.Sprintf("d.T['%s']", typeName)
	}
	return "d.T." + typeName
}

// Func creates a docsonnet function documentation entry.
//
// The name is prefixed with "#" to create a hidden field in the output.
func Func(name, help string, args []j.Type) j.Type {
	return j.Hidden(j.Call("#"+name, "d.fn", j.Args(
		j.String("help", help),
		j.List("args", args...),
	)))
}

// Obj creates a docsonnet object documentation entry.
//
// The name is prefixed with "#" to create a hidden field in the output.
func Obj(name, help string) j.Type {
	return j.Hidden(j.Call("#"+name, "d.obj", j.Args(
		j.String("help", help),
	)))
}

// Pkg creates a docsonnet package documentation entry.
func Pkg(name, url, help string) j.Type {
	return j.Hidden(j.Call("#", "d.pkg", j.Args(
		j.String("name", name),
		j.String("url", url),
		j.String("help", help),
	)))
}
