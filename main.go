package main

import (
	"flag"
	"fmt"
	"go/ast"
	"os"
	"sort"
	"strings"

	. "github.com/dave/jennifer/jen"
	"golang.org/x/tools/go/packages"
)

// StructFieldMap is a map of a struct's field names to the type of the field.
type StructFieldMap map[string]string

func structFieldsToMap(s *ast.StructType) StructFieldMap {
	out := map[string]string{}

	for _, f := range s.Fields.List {
		typeName := findExprType(f.Type)

		// anonymous field, i.e. an embedded field
		if f.Names == nil {
			out[typeName] = typeName
			continue
		}

		for _, n := range f.Names {
			out[n.Name] = typeName
		}
	}

	return out
}

func findExprType(e interface{}) (typeName string) {
	switch typ := e.(type) {
	case *ast.Ident:
		typeName = typ.Name
	case *ast.StarExpr:
		typeName = fmt.Sprintf("*%s", findExprType(typ.X))
	case *ast.SelectorExpr:
		typeName = fmt.Sprintf("%s.%s", findExprType(typ.X), typ.Sel.Name)
	case *ast.MapType:
		typeName = fmt.Sprintf("map[%s]%s", findExprType(typ.Key), findExprType(typ.Value))
	case *ast.ArrayType:
		typeName = fmt.Sprintf("[]%s", findExprType(typ.Elt))
	default:
		panic(fmt.Errorf("unhandled case for expression %#v", typ))
	}

	return typeName
}

var (
	typeNames = flag.String("type", "", "comma-delimited list of type names")
	prefix    = flag.String("prefix", "", "prefix to attach to functional options")
)

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of funcopgen:\n")
	fmt.Fprintf(os.Stderr, "\tfuncopgen -type T\n")
}

func main() {
	flag.Parse()

	if len(*typeNames) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	types := strings.Split(*typeNames, ",")

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedSyntax | packages.NeedImports,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "load: %v\n", err)
		os.Exit(1)
	}

	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}

	// Since this tool is meant to be used from a go:generate comment, there
	// should only ever be on package.
	if len(pkgs) != 1 {
		fmt.Fprintf(os.Stderr, "expected only one package!")
		os.Exit(1)
	}

	pkg := pkgs[0]

	// Find structs
	structs := map[string]StructFieldMap{}

	for _, file := range pkg.Syntax {
		for objName, obj := range file.Scope.Objects {

			switch obj.Kind {
			case ast.Typ:
				switch typ := obj.Decl.(*ast.TypeSpec).Type.(type) {
				case *ast.StructType:
					structs[objName] = structFieldsToMap(typ)
				case *ast.Ident:
					continue
				default:
					continue
				}
			case ast.Con:
				continue
			default:
				continue
			}

		}
	}

	// This is a silly way of importing the same imports from the current
	// file to the generated one, as Jennifer doesn't allow manual imports
	// without any qualifiers. See
	// https://github.com/dave/jennifer/issues/20.
	imports := Empty()
	for _, p := range pkg.Imports {
		imports.Add(Id(p.Name).Lit(p.PkgPath), Line())
	}

	for _, t := range types {
		fields, ok := structs[t]
		if !ok {
			fmt.Fprintf(os.Stderr, "Unknown type %q in %q in package %q\n", t, pkg.Name, pkg.PkgPath)
			os.Exit(1)
		}

		f := NewFile(pkg.Name)

		f.HeaderComment("This file has been automatically generated. Don't edit it.")

		f.Add(Id("import").Parens(imports))

		f.Add(Type().Id("Option").Func().Params(Op("*").Id(t)), Line())

		f.Add(
			Func().Id("New"+t).Params(Id("opts").Op("...").Id("Option")).Op("*").Id(t).Block(
				Id("o").Op(":=").Op("&").Id(t).Values(),
				Line(),
				For(Id("_, opt").Op(":=").Range().Id("opts")).Block(
					Id("opt").Call(Id("o")),
				),
				Line(),
				Return(Id("o")),
			),
			Line(),
		)

		// Sort the fields so we can traverse the map in a deterministic
		// order as we want the generated code to be the same between
		// subsequent runs.
		keys := make([]string, 0, len(fields))
		for k := range fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, field := range keys {
			typeName := Id(fields[field])

			f.Add(
				Func().Id(*prefix+field).Params(Id("x").Add(typeName)).Id("Option").Block(
					Return(
						Func().Params(Id("o").Op("*").Id(t)).Block(
							Id("o").Dot(field).Op("=").Id("x"),
						),
					),
				),
				Line(),
			)
		}

		outFile := fmt.Sprintf("zz_generated.%s_funcop.go", strings.ToLower(t))
		f.Save(outFile)
	}
}
