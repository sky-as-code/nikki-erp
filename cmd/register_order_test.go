package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/loader"
	apptraitconstants "github.com/sky-as-code/nikki-erp/modules/apptrait/constants"
	coreconstants "github.com/sky-as-code/nikki-erp/modules/core/constants"
)


// Core owns the base schemas every other module builds on, so it must be registered and
// initialized before them. That ordering must come from the dependency graph alone:
// buildDependencyGraph makes every other module implicitly depend on core, and core
// depends on apptrait. Forcing core to the front afterwards would visit it twice, and
// RegisterModels is deliberately not idempotent.
func TestTopologicalSortPutsApptraitThenCoreFirst(t *testing.T) {
	order := realModuleOrder(t)

	require.GreaterOrEqual(t, len(order), 2)
	assert.Equal(t, apptraitconstants.AppTraitModuleName, order[0])
	assert.Equal(t, coreconstants.CoreModuleName, order[1])
}

func TestTopologicalSortVisitsEachModuleOnce(t *testing.T) {
	order := realModuleOrder(t)

	seen := make(map[string]int, len(order))
	for _, name := range order {
		seen[name]++
	}
	for name, count := range seen {
		assert.Equal(t, 1, count, "module %q must be visited exactly once", name)
	}
	assert.Len(t, order, len(seen), "order must not contain duplicates")
}

// Every loaded module must appear, so that dropping the prepend cannot silently skip one.
func TestTopologicalSortCoversEveryLoadedModule(t *testing.T) {
	mods, err := loader.StaticModuleLoader{}.LoadModules()
	require.NoError(t, err)

	order := realModuleOrder(t)
	for _, mod := range mods {
		assert.Contains(t, order, mod.Name())
	}
}

// realModuleOrder reproduces the sequence registerModelInOrder and initializeInOrder use.
func realModuleOrder(t *testing.T) []string {
	t.Helper()

	mods, err := loader.StaticModuleLoader{}.LoadModules()
	require.NoError(t, err)

	app := &Application{modules: mods}
	moduleMap := app.buildModuleMap()
	depGraph, err := app.buildDependencyGraph(moduleMap)
	require.NoError(t, err)
	require.NoError(t, app.validateDependencies(depGraph))

	order, err := topologicalSort(depGraph)
	require.NoError(t, err)

	return order
}
