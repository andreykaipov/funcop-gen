// This file has been automatically generated. Don't edit it.

package animal

import (
	"bytes"
	"encoding/gob"
	jen "github.com/dave/jennifer/jen"
	"time"
)

type TestOption func(*Test)

func NewTest(opts ...TestOption) *Test {
	o := &Test{
		Name:      "bobby",
		Statement: jen.Id("lol"),
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

func WithEmbedThis(x EmbedThis) TestOption {
	return func(o *Test) {
		o.EmbedThis = x
	}
}

func WithHi(x chan Bounds) TestOption {
	return func(o *Test) {
		o.Hi = x
	}
}

func WithName(x string) TestOption {
	return func(o *Test) {
		o.Name = x
	}
}

func WithProfiles(x []map[string]interface{}) TestOption {
	return func(o *Test) {
		o.Profiles = x
	}
}

func WithStatement(x *jen.Statement) TestOption {
	return func(o *Test) {
		o.Statement = x
	}
}

func WithA(x bytes.Buffer) TestOption {
	return func(o *Test) {
		o.a = x
	}
}

func WithB(x map[time.Time]*time.Time) TestOption {
	return func(o *Test) {
		o.b = x
	}
}

func WithBounds(x *Bounds) TestOption {
	return func(o *Test) {
		o.bounds = x
	}
}

func WithC(x *gob.Encoder) TestOption {
	return func(o *Test) {
		o.c = x
	}
}
