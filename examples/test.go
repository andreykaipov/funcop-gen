package animal

import (
	"bytes"
	"encoding/gob"
	time "time"

	"github.com/dave/jennifer/jen"
)

//go:generate go run github.com/andreykaipov/funcopgen -type=Test -prefix=With -factory -unexported -unique-option

type Test struct {
	Name           string `default:"bobby"`
	a              bytes.Buffer
	b              map[time.Time]*time.Time
	c              *gob.Encoder
	*jen.Statement `default:"jen.Id(\"lol\")"`
}
