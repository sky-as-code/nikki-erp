//go:build !dynamicmods
// +build !dynamicmods

package loader

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/modules"
)

// TestEveryDependencyIsLoaded guards the rule cmd.buildDependencyGraph enforces at boot: every name
// a module lists in Deps() must resolve to a module that is actually registered, or the application
// refuses to start with "module 'X' requires 'Y' but it's not loaded".
//
// The failure is invisible to the compiler, which is why it needs a test. Deps() returns strings, so
// a module can name a dependency that is commented out of getStaticModules() and the whole tree
// still builds — it only dies when someone runs the server.
func TestEveryDependencyIsLoaded(t *testing.T) {
	loaded := make(map[string]bool)
	mods := (StaticModuleLoader{}).getStaticModules()
	for _, mod := range mods {
		loaded[mod.Name()] = true
	}

	for _, mod := range mods {
		for _, dep := range mod.Deps() {
			assert.True(t, loaded[dep],
				"module %q requires %q, but %q is not in getStaticModules()", mod.Name(), dep, dep)
		}
	}
}

// TestModuleNamesAreUnique guards against two modules answering the same Name().
//
// buildModuleMap keys by Name(), so a collision means one module silently replaces the other and
// disappears from the application with no error anywhere.
func TestModuleNamesAreUnique(t *testing.T) {
	seen := make(map[string]int)
	for _, mod := range (StaticModuleLoader{}).getStaticModules() {
		seen[mod.Name()]++
	}

	for name, count := range seen {
		assert.Equal(t, 1, count, "module name %q is registered %d times", name, count)
	}
}

// TestEveryDependantIsLoadedOrKnownAbsent covers Dependants() the way this binary can.
//
// nikkierp does not load every module in the workspace, so a declared dependant legitimately may
// be missing: inventory names "vendingmachine", which ships only in coremart. The registry skips
// such a name rather than failing startup, which is what lets one hard-coded list serve both
// binaries — at the cost of a misspelling looking exactly like a legitimate absence
// (docs/problems/inventory/001).
//
// This test pins that cost to an explicit allow-list. A new absent name has to be added here
// deliberately, so it cannot arrive by accident; coremart's TestCoreMartEveryDependantIsLoaded
// catches the typo case, since every module is loaded there.
func TestEveryDependantIsLoadedOrKnownAbsent(t *testing.T) {
	knownAbsent := map[string]bool{
		// Coremart-only: the vending machine references Inventory's product variants, but the
		// nikkierp binary has no vending machine to ask.
		"vendingmachine": true,
	}

	loaded := make(map[string]bool)
	mods := (StaticModuleLoader{}).getStaticModules()
	for _, mod := range mods {
		loaded[mod.Name()] = true
	}

	for _, mod := range mods {
		modWithDependants, ok := mod.(modules.InCodeModuleDependants)
		if !ok {
			continue
		}
		for _, dependant := range modWithDependants.Dependants() {
			assert.True(t, loaded[dependant] || knownAbsent[dependant],
				"module %q declares dependant %q, which is neither loaded nor a known "+
					"coremart-only module; add it to getStaticModules() or to knownAbsent",
				mod.Name(), dependant)
		}
	}
}

// TestDependantListsAreWellFormed mirrors what the registry rejects at boot — a module naming
// itself, or naming the same dependant twice — so a bad list fails in CI rather than at startup.
func TestDependantListsAreWellFormed(t *testing.T) {
	_, err := modules.NewModuleDependantRegistry((StaticModuleLoader{}).getStaticModules())

	assert.NoError(t, err)
}
