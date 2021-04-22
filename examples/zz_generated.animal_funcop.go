// This file has been automatically generated. Don't edit it.

package animal

import ()

type Option func(*Animal)

func NewAnimal(opts ...Option) *Animal {
	o := &Animal{}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

func WithColor(x string) Option {
	return func(o *Animal) {
		o.Color = x
	}
}

func WithSurname(x string) Option {
	return func(o *Animal) {
		o.Surname = x
	}
}

func WithCute(x bool) Option {
	return func(o *Animal) {
		o.cute = x
	}
}
