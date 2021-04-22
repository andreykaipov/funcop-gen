package animal

//go:generate go run github.com/andreykaipov/funcopgen -type=Animal -prefix=With -factory -unexported

type Animal struct {
	Surname string `default:"n/a"`
	Color   string `default:"red"`
	cute    bool
}
