package services

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

// InitDomainServices registers the constructors the container resolves the domain layer from.
//
// The engine itself is not registered here. It is built and started in OnAppStarted, because it
// writes execution rows from the moment it starts and doing that against a half-built container
// would fail on the first tick.
func InitDomainServices() error {
	return stdErr.Join(
		// Configuration is read once here rather than per tick. That is what lets a running
		// execution keep the snapshot it was created with: a value re-read on every tick could
		// lengthen or shorten a retry chain already under way.
		deps.Register(LoadSchedulerConfig),
		deps.Register(NewDeferredWaker),
		// The waker is registered under its interface as well as its concrete type: the domain
		// services depend on the interface, while OnAppStarted needs the concrete one to attach
		// the engine to it. Both must resolve to the same instance, which is what this does.
		deps.Register(func(waker *DeferredWaker) EngineWaker { return waker }),
		deps.Register(NewJobDomainServiceImpl),
		deps.Register(NewExecutionDomainServiceImpl),
	)
}
