//go:build !dynamicmods
// +build !dynamicmods

package loader

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
