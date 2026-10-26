package deps_inject

import (
	"errors"

	"go.uber.org/dig"
)

// See dig's documentation for usage: https://pkg.go.dev/go.uber.org/dig

var container *dig.Container = dig.New()

// containerMu serializes every access to the container: dig is not safe for concurrent use, and
// several paths resolve from it at request time (lazy org-lifecycle builds, engine lookups,
// computed-field function calls). It is re-entrant because constructors and InitSubModule
// functions run inside Invoke and register or invoke in turn.
var containerMu reentrantMutex

// Container exposes the raw container. Calls on it bypass the lock, so use it only during the
// single-threaded boot.
func Container() *dig.Container {
	return container
}

func Register(constructors ...any) error {
	containerMu.Lock()
	defer containerMu.Unlock()
	var err error
	for _, constructor := range constructors {
		err = errors.Join(err, container.Provide(constructor))
	}
	return err
}

// RegisterNamed registers a constructor whose result is retrievable only by the given name.
// Consumers must declare a dig.In param struct field tagged `name:"<name>"`.
func RegisterNamed(name string, constructor any) error {
	containerMu.Lock()
	defer containerMu.Unlock()
	return container.Provide(constructor, dig.Name(name))
}

// Invoke holds the container lock for the whole call, so function must only resolve and return:
// a function that blocks (starting a server, waiting on a channel) stalls every other goroutine
// that resolves from the container, which is every request. Resolve inside, run outside.
func Invoke(function any) error {
	containerMu.Lock()
	defer containerMu.Unlock()
	return container.Invoke(function)
}
