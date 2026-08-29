package sales

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sky-as-code/nikki-erp/modules"
)

// A DynamicModule assignment that stops compiling is the point of the declared type: under the
// wider InCodeModule the method set is found by a type assertion, so a module that lost
// RegisterModels would still build and silently register no schemas.
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
// fall through to its moduleName+"_" default. A module whose prefix stops matching its name emits
// an empty migration rather than an error.
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

// TestEveryConsumedModuleIsDeclaredAsADependency guards an init-ordering bug that only shows up at
// boot: infra/external binds each consumed module's port eagerly at Init, so a module whose service
// is not registered yet panics the whole application at start-up.
func TestEveryConsumedModuleIsDeclaredAsADependency(t *testing.T) {
	declared := map[string]bool{}
	for _, dep := range (&SalesModule{}).Deps() {
		declared[dep] = true
	}

	// The module behind each port Sales binds in infra/external.
	for port, module := range map[string]string{
		"PaymentMethodExtService":        "paymentinvoice",
		"TaxCalculationExtService":       "accounting",
		"SettingsRegistrationExtService": "settings",
		"EffectiveSettingsExtService":    "settings",
		"UomUsageProbe":                  "essential",
	} {
		if !declared[module] {
			t.Errorf("Sales binds %s but does not declare %q in Deps(); the loader may start "+
				"Sales before that module registers its service, and Init will panic",
				port, module)
		}
	}
}

// TestOwnAppServicesAreNotBoundAsExternalPorts guards an init-ordering cycle that only shows up at
// boot: infra/external binds ports eagerly and runs first in Init, while Sales' own application
// services register several steps later, so resolving one of them in InitExternal fails the boot
// while every unit test still passes.
func TestOwnAppServicesAreNotBoundAsExternalPorts(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("infra", "external", "index.go"))
	if err != nil {
		t.Fatalf("infra/external/index.go must be readable: %v", err)
	}

	// Sales' own interfaces packages. A port bound in infra/external must belong to ANOTHER module.
	for _, own := range []string{
		"modules/sales/interfaces/channel",
	} {
		if strings.Contains(string(source), own) {
			t.Errorf("infra/external binds %q, which is one of Sales' own application services; "+
				"it is registered after InitExternal runs, so the container cannot resolve it "+
				"there and Init will fail at boot", own)
		}
	}
}
