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
