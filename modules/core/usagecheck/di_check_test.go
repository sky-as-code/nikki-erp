package usagecheck

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

// The dispatcher is constructed by the container, so its dependencies must be things the
// container actually holds by the time core initializes: the bus (registered by cqrs's own
// InitSubModule) and the dependant registry (registered by cmd before any module's Init).
// A mismatch here surfaces at boot as a dig error, not at compile time.
func TestInitSubModule_DispatcherResolves(t *testing.T) {
	require.NoError(t, deps.Register(func() cqrs.CqrsBus { return newFakeBus() }))
	registry, err := modules.NewModuleDependantRegistry(nil)
	require.NoError(t, err)
	require.NoError(t, deps.Register(func() *modules.ModuleDependantRegistry { return registry }))
	require.NoError(t, deps.Register(func() logging.LoggerService { return nil }))

	require.NoError(t, InitSubModule())

	err = deps.Invoke(func(dispatcher *Dispatcher) {
		assert.NotNil(t, dispatcher)
	})
	assert.NoError(t, err)
}
