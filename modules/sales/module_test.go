package sales

import (
	"testing"

	"github.com/sky-as-code/nikki-erp/modules"
)

// The module contract this package must keep, checked here rather than discovered at boot.

// A DynamicModule assignment that stops compiling is the point of the declared type: under the
// wider InCodeModule the method set is found by a type assertion instead, so a module that lost
// RegisterModels would still build, still load, and silently register no schemas at all.
var _ modules.DynamicModule = ModuleSingleton

func TestModuleIdentity(t *testing.T) {
	if got := ModuleSingleton.Name(); got != "sales" {
		t.Errorf("Name() = %q, want %q", got, "sales")
	}
	if ModuleSingleton.IsInternal() {
		t.Error("IsInternal() = true, want false: Sales serves external callers")
	}
	if got := ModuleSingleton.LabelKey(); got == "" {
		t.Error("LabelKey() is empty")
	}
}

// TestSchemaPrefixIsDerivable pins the assumption that lets cmd/application.go's schemaPrefixesOf
// fall through to its default.
//
// That function maps a module to the prefixes its schemas are named with, and returns
// moduleName+"_" unless the module has an explicit case. Sales names every schema "sales_", so it
// needs no case — but a module whose prefix stops matching its name emits an empty migration
// rather than an error, so the assumption is worth stating where a rename would break it.
func TestSchemaPrefixIsDerivable(t *testing.T) {
	const wantPrefix = "sales_"
	if got := ModuleSingleton.Name() + "_"; got != wantPrefix {
		t.Errorf("derived schema prefix = %q, want %q; "+
			"add a case to schemaPrefixesOf in cmd/application.go", got, wantPrefix)
	}
}

// TestDepsAreDeclaredOnce guards against a duplicate creeping into Deps during the many tasks that
// will extend this module.
func TestDepsAreDeclaredOnce(t *testing.T) {
	seen := map[string]bool{}
	for _, dep := range ModuleSingleton.Deps() {
		if seen[dep] {
			t.Errorf("dependency %q is declared twice", dep)
		}
		seen[dep] = true
	}
	if len(seen) == 0 {
		t.Error("Deps() is empty; Sales reads at least dynamicresource")
	}
}
