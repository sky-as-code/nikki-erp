package deps_inject

import (
	"errors"

	"go.uber.org/dig"
)

// See dig's documentation for usage: https://pkg.go.dev/go.uber.org/dig

var container *dig.Container = dig.New()

func Container() *dig.Container {
	return container
}

func Register(constructors ...any) error {
	var err error
	err = nil
	for _, constructor := range constructors {
		if err = container.Provide(constructor); err != nil {
			err = errors.Join(err)
		}
	}
	return err
}

// RegisterNamed registers a constructor whose result is retrievable only by the given name.
// Consumers must declare a dig.In param struct field tagged `name:"<name>"`.
func RegisterNamed(name string, constructor any) error {
	return container.Provide(constructor, dig.Name(name))
}

func Invoke(function any) error {
	return container.Invoke(function)
}
