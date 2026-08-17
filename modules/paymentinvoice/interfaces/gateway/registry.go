package gateway

import (
	"sync"

	"go.bryk.io/pkg/errors"
)

// Registry holds the gateway implementations this build ships, keyed by adapter code.
//
// It is the seam between data and code: a payment-method row names an adapter, and this is what
// turns that name into something that can take a payment. A row naming an adapter that is absent
// — not built, or not enabled by configuration — fails here, when the payment is attempted, with
// a message naming the adapter. That is deliberate: the alternative is refusing the row at write
// time, which would make configuration depend on build flags.
type Registry struct {
	mutex    sync.RWMutex
	adapters map[string]PaymentGateway
}

func NewRegistry() *Registry {
	return &Registry{adapters: map[string]PaymentGateway{}}
}

// Register adds an adapter. It is called during Init, before any request is served.
func (this *Registry) Register(adapter PaymentGateway) error {
	if adapter == nil {
		return errors.New("gateway registry: cannot register a nil adapter")
	}
	code := adapter.AdapterCode()
	if code == "" {
		return errors.New("gateway registry: adapter code must not be empty")
	}

	this.mutex.Lock()
	defer this.mutex.Unlock()

	if _, exists := this.adapters[code]; exists {
		// Two adapters answering to one code would make which of them runs depend on
		// registration order, so this is refused rather than resolved.
		return errors.Errorf("gateway registry: adapter '%s' is already registered", code)
	}
	this.adapters[code] = adapter
	return nil
}

// Get returns the adapter for a code, or false when this build has none enabled.
func (this *Registry) Get(adapterCode string) (PaymentGateway, bool) {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	adapter, exists := this.adapters[adapterCode]
	return adapter, exists
}

// Codes lists the registered adapter codes. Useful for diagnostics and for telling an
// administrator which adapter_code values a payment-method row may name.
func (this *Registry) Codes() []string {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	codes := make([]string, 0, len(this.adapters))
	for code := range this.adapters {
		codes = append(codes, code)
	}
	return codes
}
