package animal

//go:generate go run github.com/andreykaipov/funcopgen -type=Animal -prefix=With -factory -unexported

type Animal struct {
	Surname string
	Color   string
	cute    bool
}
