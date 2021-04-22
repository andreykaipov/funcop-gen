package animal

//go:generate go run github.com/andreykaipov/funcopgen -type=Animal -factory

type Animal struct {
	Surname string `default:"n/a"`
	Color   string `default:"red"`
	cute    bool
}
