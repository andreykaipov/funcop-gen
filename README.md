# funcopgen

funcopgen generates [functional
options](https://github.com/tmrts/go-patterns/blob/master/idiom/functional-options.md)
for your Go structs.

## usage

Better illustrated through example, here's an `Animal` struct:

```go
package animal

type Animal struct {
	Surname string `default:"n/a"`
	Color   string `default:"red"`
	cute    bool
}
```

Add a `go:generate` directive anywhere inside of the `animal` package as
follows.

```go
//go:generate go run github.com/andreykaipov/funcopgen -type=Animal -factory
```

Install the tool and generate!

```console
go install -mod=readonly github.com/andreykaipov/funcopgen@latest
go generate ./...
```

Enjoy the new file `zz_generated.animal_funcop.go` in your package:

```go
// This file has been automatically generated. Don't edit it.

package animal

type Option func(*Animal)

func NewAnimal(opts ...Option) *Animal {
	o := &Animal{
		Color:   "red",
		Surname: "n/a",
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

func Color(x string) Option {
	return func(o *Animal) {
		o.Color = x
	}
}

func Surname(x string) Option {
	return func(o *Animal) {
		o.Surname = x
	}
}
```

Now we can instantiate our animals as follows:

```go
db := NewAnimal(
	Surname("ducky"),
	Color("blue"),
)
```

Fun!

### extras

The generated code can be tweaked by passing extra flags:

```console
Usage of funcopgen:
  -factory
        If present, add a factory function for your type, e.g. NewAnimal(opt ...Option)
  -prefix string
        Prefix to attach to functional options, e.g. WithColor, WithName, etc.
  -type string
        Comma-delimited list of type names
  -unexported
        If present, functional options are also generated for unexported fields.
  -unique-option
        If present, prepends the type to the Option type, e.g. AnimalOption.
        Handy if generating for several structs within the same package.
```

See [examples/test.go](./examples/test.go) for an example of all of these flags.

## faq

### I vendor my dependencies. How can I vendor this tool?

You might want to read through [this Go
thread](https://github.com/golang/go/issues/25922) and check out [this
StackOverflow
answer](https://stackoverflow.com/questions/52428230/how-do-go-modules-work-with-installable-commands/54028731#54028731)
for suggestions on how others have accomplished vendoring development
dependencies.

The TLDR of it is to create a tools package with the following contents:

```go
// +build tools

package tools

import (
	_ "github.com/andreykaipov/funcopgen"
)
```

After a `go mod tidy` and a `go mod vendor`, the above `go:generate` directive
will use the vendored tool. Depending on your Go version, you might need to
explicitly tell Go to use the vendored dependencies, i.e. `go generate
-mod=vendor ./...`.

### How do I integrate it into my development lifecycle?

Code generation shouldn't happen often, but it's easy enough to integrate this
into our build. Just `go generate` before a `go build`. For example, if we're
using make as our build tool:

```Makefile
generate:
    go generate ./...

build: generate
    go build ./...
```
