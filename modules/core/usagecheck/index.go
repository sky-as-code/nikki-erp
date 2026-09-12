package usagecheck

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

// InitSubModule registers the dispatcher every owning module resolves to run its pre-delete
// check. There is one dispatcher for the application: it is stateless apart from the bus and
// the dependant registry, and it reads the owning module's name from each call rather than
// holding one.
//
// The per-module CheckerRegistry is NOT registered here. Each dependant module owns its own and
// registers it in its own Init, because a single shared registry would let one module's
// checkers answer under another module's name.
func InitSubModule() error {
	return deps.Register(NewDispatcher)
}
