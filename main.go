package main

import (
	"flag"
	"fmt"
	"go/ast"
	"os"
	"sort"
	"strings"
	"unicode"

	. "github.com/dave/jennifer/jen"
	"github.com/fatih/structtag"
	"golang.org/x/tools/go/packages"
)

// StructFieldMap is a map of a struct's field names to data about the field.
type StructFieldMap map[string]*FieldData

type FieldData struct {
	// Tags is a map representing a field's tags, e.g. `default:"hello"`
	Tags *structtag.Tags

	// Type is the string representation of a type, e.g. "string" or
	// "[]*myQualified.StructType"
	Type string
}

func firstRune(str string) (r rune) {
	for _, r = range str {
		return
	}
	return
}

func structFieldsToMap(s *ast.StructType) StructFieldMap {
	out := StructFieldMap{}

	for _, f := range s.Fields.List {
		typeName := findFieldType(f)

		data := &FieldData{
			Type: typeName,
			Tags: findFieldTags(f),
		}

		// anonymous field, i.e. an embedded field
		if f.Names == nil {
			out[typeName] = data
			continue
		}

		for _, n := range f.Names {
			out[n.Name] = data
		}
	}

	return out
}

func findFieldTags(f *ast.Field) *structtag.Tags {
	if f.Tag == nil {
		return &structtag.Tags{}
	}

	// get rid of the backticks
	trimmed := f.Tag.Value[1 : len(f.Tag.Value)-1]

	tag, err := structtag.Parse(trimmed)
	if err != nil {
		panic(err)
	}

	return tag
}

func findFieldType(field *ast.Field) string {
	var f func(e interface{}) string

	f = func(e interface{}) (typeName string) {
		switch typ := e.(type) {
		case *ast.Ident:
			typeName = typ.Name
		case *ast.StarExpr:
			typeName = fmt.Sprintf("*%s", f(typ.X))
		case *ast.SelectorExpr:
			typeName = fmt.Sprintf("%s.%s", f(typ.X), typ.Sel.Name)
			selectorExprs[typeName] = nil
		case *ast.MapType:
			typeName = fmt.Sprintf("map[%s]%s", f(typ.Key), f(typ.Value))
		case *ast.ArrayType:
			typeName = fmt.Sprintf("[]%s", f(typ.Elt))
		default:
			panic(fmt.Errorf("unhandled case for expression %#v", typ))
		}

		return typeName
	}

	return f(field.Type)
}

var (
	selectorExprs = map[string]interface{}{}
	typeNames     = flag.String("type", "", "comma-delimited list of type names")
	prefix        = flag.String("prefix", "", "prefix to attach to functional options")
	factory       = flag.Bool("factory", false, "if present, add a factory function for your type, e.g. NewX")
	unexported    = flag.Bool("unexported", false, "if present, functional options are also generated for unexported fields")
)

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
	// should only ever be one package.
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

	// Compares the package imports to import only those that have a prefix
	// with any of our found selector expressions.
	setImports := func(g *Group) {
		for _, p := range pkg.Imports {
			for e := range selectorExprs {
				if strings.HasPrefix(e, p.Name) {
					// Jennifer doesn't have a nice func for
					// manual imports. See
					// https://github.com/dave/jennifer/issues/20.
					g.Add(Id(p.Name).Lit(p.PkgPath), Line())
					break
				}
			}
		}
	}

	for _, t := range types {
		fields, ok := structs[t]
		if !ok {
			fmt.Fprintf(os.Stderr, "Unknown type %q in %q in package %q\n", t, pkg.Name, pkg.PkgPath)
			os.Exit(1)
		}

		// Sort the fields so we can traverse the map in a deterministic
		// order as we want the generated code to be the same between
		// subsequent runs.
		keys := make([]string, 0, len(fields))
		for k := range fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		f := NewFile(pkg.Name)

		f.HeaderComment("This file has been automatically generated. Don't edit it.")

		f.Add(Id("import").CustomFunc(Options{Open: "(", Close: ")"}, setImports))

		f.Add(Type().Id("Option").Func().Params(Op("*").Id(t)), Line())

		setDefaults := func(d Dict) {
			for _, field := range keys {
				tags := fields[field].Tags

				if tag, _ := tags.Get("default"); tag != nil {
					switch fields[field].Type {
					case "string":
						d[Id(field)] = Lit(tag.Name)
					default:
						d[Id(field)] = Id(tag.Name)
					}
				}
			}
		}

		if *factory {
			f.Add(
				Func().Id("New"+t).Params(Id("opts").Op("...").Id("Option")).Op("*").Id(t).Block(
					Id("o").Op(":=").Op("&").Id(t).Values(DictFunc(setDefaults)),
					Line(),
					For(Id("_, opt").Op(":=").Range().Id("opts")).Block(
						Id("opt").Call(Id("o")),
					),
					Line(),
					Return(Id("o")),
				),
				Line(),
			)
		}

		for _, field := range keys {
			typeName := Id(fields[field].Type)

			titledField := field
			if unicode.IsLower(firstRune(field)) {
				if !*unexported {
					continue
				}
				titledField = strings.Title(field)
			}

			f.Add(
				Func().Id(*prefix+titledField).Params(Id("x").Add(typeName)).Id("Option").Block(
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

		if err := f.Save(outFile); err != nil {
			panic(err)
		}
	}
}
