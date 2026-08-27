package services

import "sync/atomic"

// DeferredWaker is the waker the domain services hold, standing in for an engine that does not
// exist yet when they are built.
//
// The ordering problem it solves is real and not incidental: the domain services are constructed
// during Init, while the engine cannot start until OnAppStarted, because from its first tick it
// writes execution rows and must not do that against a half-built container. Injecting the
// engine directly would be a cycle - the engine reads jobs through the same services - so the
// indirection is what breaks it rather than a convenience.
//
// Before the engine attaches, a wake is dropped rather than queued. That is correct and not a
// gap: a wake carries no information beyond "look again", and the engine's first tick looks at
// everything anyway. Queueing them would only make the engine repeat its own first tick.
type DeferredWaker struct {
	engine atomic.Pointer[Engine]
}

func NewDeferredWaker() *DeferredWaker {
	return &DeferredWaker{}
}

// Attach names the engine that will receive wakes from now on. It is called once, from
// OnAppStarted, after the engine is built and started.
func (this *DeferredWaker) Attach(engine *Engine) {
	this.engine.Store(engine)
}

// Detach stops delivery during shutdown, so that a request still in flight cannot wake an engine
// that has already drained and would have nothing to run the work with.
func (this *DeferredWaker) Detach() {
	this.engine.Store(nil)
}

// Wake implements EngineWaker.
func (this *DeferredWaker) Wake(reason WakeReason) {
	if engine := this.engine.Load(); engine != nil {
		engine.Wake(reason)
	}
}

var _ EngineWaker = (*DeferredWaker)(nil)
