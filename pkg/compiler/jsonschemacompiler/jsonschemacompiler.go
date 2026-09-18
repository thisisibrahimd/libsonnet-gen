package jsonschemacompiler

import (
	"fmt"

	"github.com/thisisibrahimd/libsonnet-gen/pkg/builder"
	"github.com/thisisibrahimd/libsonnet-gen/pkg/render"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stoewer/go-strcase"
)

func genAdditiveObjWrapper(in builder.Type, path []string) builder.Type {
	if len(path) == 0 {
		return in
	}

	return builder.Merge(builder.Object(path[0], genAdditiveObjWrapper(in, path[1:])))
}

func CompileLibsonnet(s *jsonschema.Schema, name string, curPath []string) builder.Type {
	// if the current schema is just a ref to another def in another location, replace the schema in that ref with our current schema.
	if s.Ref != nil {
		s = s.Ref
	}

	if len(curPath) > 50 {
		panic("Currently 50 levels deep into properties. There is most likely a acylic loop")
	}

	// loop through the properties in this current schema
	// limitations: we do not support special properties
	propTypes := []builder.Type{}
	for pName, p := range s.Properties {
		newPath := make([]string, len(curPath))
		copy(newPath, curPath)
		newPath = append(newPath, pName)

		if p.Ref != nil {
			p = p.Ref
		}

		var pT string
		if p.Types != nil {
			pT = p.Types.ToStrings()[0]
		} else {
			if p.Enum != nil {
				pT = "string"
			} else if len(p.AnyOf) > 0 {
				pT = "string"
			}
		}

		switch pT {
		case "array":
			funcName := fmt.Sprintf("with%s", strcase.UpperCamelCase(pName))
			propType := builder.Func(
				funcName,
				builder.Args(
					builder.Required(builder.SafeString(strcase.LowerCamelCase(pName), "")),
				),
				builder.Object(
					strcase.LowerCamelCase(pName),
					genAdditiveObjWrapper(
						builder.IfThenElse(
							pName,
							builder.Call("", "std.isArray", builder.Args(builder.Ref("v", builder.SafeIdentifier(strcase.LowerCamelCase(pName))))),
							builder.Ref("", builder.SafeIdentifier(strcase.LowerCamelCase(pName))),
							builder.List("", builder.Ref("", builder.SafeIdentifier(strcase.LowerCamelCase(pName)))),
						),
						curPath,
					),
				),
			)
			propTypes = append(propTypes, propType)

			// create mixin function for ararys
			propTypeMixin := builder.Func(
				fmt.Sprintf("%sMixin", funcName),
				builder.Args(
					builder.Required(builder.SafeString(strcase.LowerCamelCase(pName), "")),
				),
				builder.Object(
					strcase.LowerCamelCase(pName),
					genAdditiveObjWrapper(
						builder.Merge(
							builder.IfThenElse(
								pName,
								builder.Call("", "std.isArray", builder.Args(builder.Ref("v", builder.SafeIdentifier(strcase.LowerCamelCase(pName))))),
								builder.Ref("", builder.SafeIdentifier(strcase.LowerCamelCase(pName))),
								builder.List("", builder.Ref("", builder.SafeIdentifier(strcase.LowerCamelCase(pName)))),
							),
						),
						curPath,
					),
				),
			)
			propTypes = append(propTypes, propTypeMixin)

			// nested helpers for array item properties
			if p.Items != nil {
				if itemSchema, ok := p.Items.(*jsonschema.Schema); ok && len(itemSchema.Properties) > 0 {
					nestedObj := CompileLibsonnet(itemSchema, strcase.LowerCamelCase(pName), newPath)
					propTypes = append(propTypes, nestedObj)
				}
			}

			// mapXyz(f) - for all arrays
			mapPath := append(append([]string{}, curPath...), pName)
			mapInner := builder.Call(pName, "std.map", builder.Args(
				builder.CallArgFrom(builder.Ref("", "f")),
				builder.CallArgFrom(builder.SuperRef(pName)),
			))
			mapWrapped := genAdditiveObjWrapper(mapInner, mapPath[:len(mapPath)-1])
			if len(mapPath) > 1 {
				mapWrapped = builder.Merge(mapWrapped)
			}
			mapFunc := builder.Func(
				fmt.Sprintf("map%s", strcase.UpperCamelCase(pName)),
				builder.Args(builder.Required(builder.Ref("f", ""))),
				builder.ConciseObject("", mapWrapped),
			)
			propTypes = append(propTypes, mapFunc)

			// mapXyzByName(name, transformFunc) - only if items have "name" field
			hasName := false
			if itemSchema, ok := p.Items.(*jsonschema.Schema); ok {
				_, hasName = itemSchema.Properties["name"]
			}
			if hasName {
				comprehension := builder.ArrayComprehension(pName, "c", "name",
					builder.SuperRef(pName),
					builder.Ref("", "name"),
					builder.Call("", "transformFunc", builder.Args(builder.CallArgFrom(builder.Ref("", "c")))),
				)
				mapByNameWrapped := genAdditiveObjWrapper(comprehension, mapPath[:len(mapPath)-1])
				if len(mapPath) > 1 {
					mapByNameWrapped = builder.Merge(mapByNameWrapped)
				}
				mapByNameFunc := builder.Func(
					fmt.Sprintf("map%sByName", strcase.UpperCamelCase(pName)),
					builder.Args(
						builder.Required(builder.String("name", "")),
						builder.Required(builder.Ref("transformFunc", "")),
					),
					builder.ConciseObject("", mapByNameWrapped),
				)
				propTypes = append(propTypes, mapByNameFunc)
			}

		case "object":
			propType := CompileLibsonnet(p, strcase.LowerCamelCase(pName), newPath)
			propTypes = append(propTypes, propType)
		default:
			funcName := fmt.Sprintf("with%s", strcase.UpperCamelCase(pName))
			propType := builder.Func(
				funcName,
				builder.Args(
					builder.Required(builder.SafeString(strcase.LowerCamelCase(pName), "")),
				),
				builder.Object(
					strcase.LowerCamelCase(pName),
					genAdditiveObjWrapper(
						builder.Ref(pName, builder.SafeIdentifier(strcase.LowerCamelCase(pName))),
						curPath,
					),
				),
			)
			propTypes = append(propTypes, propType)
		}
	}

	render.SortFields(propTypes)

	return builder.Object(name, propTypes...)
}
