package deps_inject

import (
	"errors"

	"go.uber.org/dig"
)

var container *dig.Container = dig.New()

func Container() *dig.Container {
	return container
}

func Provide(constructors ...any) error {
	var err error
	err = nil
	for _, constructor := range constructors {
		if err = container.Provide(constructor); err != nil {
			err = errors.Join(err)
		}
	}
	return err
}

func Invoke(function any) error {
	return container.Invoke(function)
}
